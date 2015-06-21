package main

import (
	"fmt"
	"github.com/freman/gobin/pastes"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/tbruyelle/hipchat-go/hipchat"
	"log"
	"net/http"
	"strings"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "new", make(map[string]interface{}))
}

func viewPasteHandler(w http.ResponseWriter, r *http.Request) {
	paste := loadPasteFromRequest(w, r)
	if paste == nil {
		return
	}

	data := make(map[string]interface{})
	data["Paste"] = paste

	if cookie, err := r.Cookie("gobin"); err == nil {
		setCookieDefaults(cookie)
		http.SetCookie(w, cookie)

		data["HaveCookie"] = true
		data["Cookie"] = cookie
		data["CookieMatch"] = cookie.Value == paste.Author
	}

	renderTemplate(w, "view", data)
}

func viewAttachmentHandler(w http.ResponseWriter, r *http.Request) {
	paste := loadPasteFromRequest(w, r)
	if paste == nil {
		return
	}

	if paste.Content == "binary" && strings.Contains(paste.Syntax, "/") {
		sendBinaryAttachment(paste, w)
		return
	}

	http.Error(w, "404 page not found", http.StatusNotFound)
}

func getPasteHandler(w http.ResponseWriter, r *http.Request) {
	paste := loadPasteFromRequest(w, r)
	if paste == nil {
		return
	}

	if paste.Content == "binary" && strings.Contains(paste.Syntax, "/") {
		sendBinaryAttachment(paste, w)
	} else {
		filename := "paste.txt"
		contentType := "text/plain"

		if strings.Contains(paste.Title, ".") {
			filename = paste.Title
		} else if paste.Syntax != "" {
			filename = "paste." + getExtension(paste.Syntax)
		}

		w.Header().Add("Content-Disposition", "attachment; filename="+filename)
		w.Header().Set("Content-Type", contentType)
		fmt.Fprintf(w, "%s", paste.Content)
	}
}

func newPasteHandler(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	id := r.URL.Path
	if len(id) > 4 {
		bin.Load(id)
		data["Paste"] = loadPasteFromRequest(w, r)
	}

	renderTemplate(w, "new", data)
}

func saveNewPasteHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1024 * 1024)
	if err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	var parentPaste, paste *pastes.Paste

	parent := r.FormValue("parent")
	if parent != "" {
		parentPaste = loadPaste(w, parent)
		paste = parentPaste.Modify()
	} else {
		paste = bin.New()
	}

	cookie := getOrGenerateCookie(r)
	http.SetCookie(w, cookie)

	paste.Author = cookie.Value
	paste.Title = r.FormValue("title")
	paste.Syntax = r.FormValue("syntax")
	//	paste.Expiration, err = strconv.Atoi(r.FormValue("expiration"))
	paste.Content = r.FormValue("content")
	err = paste.Save()
	if err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	newRecentPaste(paste)
	if parentPaste != nil {
		parentPaste.Save()
	}
	http.Redirect(w, r, "/p/"+paste.ID, http.StatusSeeOther)
}

func uploadNewPasteHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1024 * 1024)
	if err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	cookie := getOrGenerateCookie(r)
	http.SetCookie(w, cookie)

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	log.Println(fileHeader.Header["Content-Disposition"], fileHeader.Header["Content-Type"], err)

	// Content type as provided by the browser
	strContentType := fileHeader.Header.Get("Content-Type")
	contentType := strings.Split(strContentType, "/")

	http.SetCookie(w, cookie)
	paste := bin.New()
	paste.Author = cookie.Value

	paste.Title = fileHeader.Filename

	// Attempt to work with the browser content type
	if syntax, ok := guessByContentType(strContentType); ok {
		paste.Syntax = syntax
		paste.ContentFromReader(file)
	} else {
		// Still working with the browsers content type
		switch {
		case contentType[0] == "image":
			paste.Syntax = fileHeader.Header.Get("Content-Type")
			paste.Attach(file)
		case contentType[0] == "text":
			paste.ContentFromReader(file)
		// todo video/avi
		default:
			// Ok now we try to work with the content directly
			if syntax, ok := guessByContent(file); ok {
				paste.Syntax = syntax
				paste.ContentFromReader(file)
			} else {
				paste.Attach(file)
			}
		}
	}

	//	paste.Expiration, err = strconv.Atoi(r.FormValue("expiration"))
	err = paste.Save()
	if err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	newRecentPaste(paste)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"response\": \"ok\", \"target\": \"/p/"+paste.ID+"\"}")
}

func diffPasteHandler(w http.ResponseWriter, r *http.Request) {
	ids := strings.Split(r.URL.Path, "/")
	alpha := loadPaste(w, ids[0])
	if alpha == nil {
		return
	}
	beta := loadPaste(w, ids[1])
	if beta == nil {
		return
	}

	dmp := diffmatchpatch.New()
	diff := dmp.DiffCleanupSemantic(dmp.DiffMain(beta.Content, alpha.Content, true))

	data := make(map[string]interface{})
	data["Diff"] = diff
	data["Alpha"] = alpha
	data["Beta"] = beta

	renderTemplate(w, "diff", data)
}

func editPasteHandler(w http.ResponseWriter, r *http.Request) {
	paste := loadPasteFromRequest(w, r)
	if paste == nil {
		return
	}

	if !checkEditCookie(r, paste) {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
	}

	data := make(map[string]interface{})
	data["Paste"] = paste
	data["Which"] = "Edit"

	renderTemplate(w, "edit", data)
}

func saveEditPasteHandler(w http.ResponseWriter, r *http.Request) {
	paste := loadPasteFromRequest(w, r)
	if paste == nil {
		return
	}

	if !checkEditCookie(r, paste) {
		saveNewPasteHandler(w, r)
		return
	}

	err := r.ParseMultipartForm(1024 * 1024)
	if err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	paste.Title = r.FormValue("title")
	paste.Syntax = r.FormValue("syntax")
	//	paste.Expiration, err = strconv.Atoi(r.FormValue("expiration"))
	paste.Content = r.FormValue("content")

	err = paste.Save()
	if err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/p/"+paste.ID, http.StatusSeeOther)
}

func sharePasteHandler(w http.ResponseWriter, r *http.Request) {
	bits := strings.Split(r.URL.Path, "/")
	if len(bits) < 3 {
		http.Error(w, "400 Method Not Allowed", http.StatusBadRequest)
		return
	}

	platform := bits[0]
	room := bits[1]
	paste := loadPaste(w, bits[2])

	if paste == nil {
		return
	}

	switch (platform) {
	case "hipchat":
		if hipChat == nil {
			http.Error(w, "404 Not Found", http.StatusNotFound)
			return
		}
		if config.HipChat.Room != room {
			http.Error(w, "404 Not Found", http.StatusNotFound)
			return
		}

		//todo make sure room is permitted

		notification := &hipchat.NotificationRequest{
			Message: fmt.Sprintf("Paste: <a href=\"%s/p/%s\">%s</a>", config.BaseURL, paste.ID, paste.Title),
		}

		_ = notification

		if _, err := hipChat.Room.Notification(room, notification); err != nil {
			log.Printf("Unable to send notification to room %s: %s", room, err)
		}

		http.Redirect(w, r, "/p/"+paste.ID, http.StatusFound)
	}
}

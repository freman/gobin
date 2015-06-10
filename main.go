package main

import (
	"flag"
	"fmt"
	"github.com/freman/gobin/pastes"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/tbruyelle/hipchat-go/hipchat"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Handlers map[string]http.HandlerFunc

var (
	config Config
	bin              *pastes.Pastes
	templates        map[string]*template.Template
	pageDefaults     map[string]interface{}
	hipChat          *hipchat.Client
)

func init() {
	var flConfig string
	flag.StringVar(&flConfig, "config", "./config.toml", "Configuration file")
	flag.Parse()

	config = loadConfig(flConfig)

	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	templatesDir := "web/templates" //config.Templates.Path
	pages, err := filepath.Glob(filepath.Join(templatesDir, "pages", "*.tmpl"))
	if err != nil {
		log.Fatal(err)
	}

	includes, err := filepath.Glob(filepath.Join(templatesDir, "includes", "*.tmpl"))
	if err != nil {
		log.Fatal(err)
	}

	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"IsBinary": func(p *pastes.Paste) bool {
			return p.Content == "binary" && strings.Contains(p.Syntax, "/")
		},
		"IsImage": func(p *pastes.Paste) bool {
			return p.Content == "binary" && strings.HasPrefix(p.Syntax, "image/")
		},
		"DiffSpan": func(d diffmatchpatch.Diff) interface{} {
			if d.Type == diffmatchpatch.DiffEqual {
				return d.Text
			}
			class := "add"
			if d.Type == diffmatchpatch.DiffDelete {
				class = "remove"
			}
			span := "<span class=\"diff " + class + "\">"
			output := ""
			lineEndings := ""
			if strings.Contains(d.Text, "\r") {
				lineEndings += "\r"
			}
			if strings.Contains(d.Text, "\n") {
				lineEndings += "\n"
			}
			for n, v := range strings.Split(d.Text, lineEndings) {
				if n > 0 {
					output += lineEndings
				}
				output += span + template.HTMLEscapeString(v) + "</span>"
			}
			return template.HTML(output)
		},
	}

	for _, page := range pages {
		files := append(includes, page)
		templates[filepath.Base(page)] = template.Must(template.New("thetemplate").Funcs(funcMap).ParseFiles(files...))
	}

	if pageDefaults == nil {
		pageDefaults = make(map[string]interface{})
	}

	pageDefaults["Site"] = config.Site
	pageDefaults["GuessLanguages"] = template.JS(config.Site.GuessLanguages.String())

	pageDefaults["HaveCookie"] = false
	pageDefaults["CookieMatch"] = false
	pageDefaults["HipChat"] = nil

	// HTML and CSS is hard, fixme
	config.HipChat.ForceRoom = true

	if config.HipChat.Enabled && config.HipChat.ApiToken != "" {
		log.Print("Verifying HipChat configuration")
		hipChat = hipchat.NewClient(config.HipChat.ApiToken)
		if config.HipChat.ForceRoom == false && len(config.HipChat.PermittedRooms) == 0 {
			rooms, _, err := hipChat.Room.List()
			if err != nil {
				log.Fatalf("Unable to get list of HipChat rooms: %s", err)
			}

			for _, r := range rooms.Items {
				config.HipChat.PermittedRooms = append(config.HipChat.PermittedRooms, r.Name)
			}
		}

		if _, _, err := hipChat.Room.Get(config.HipChat.DefaultRoom); err != nil {
			log.Fatalf("Unable to find room %s: %s", config.HipChat.DefaultRoom, err)
		}

		pageDefaults["HipChat"] = config.HipChat
	}

	pageDefaults["ShowNotify"] = pageDefaults["HipChat"] != nil
}

func renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) error {
	tmpl, ok := templates[name+".tmpl"]
	if !ok {
		return fmt.Errorf("Template %s doesn't exist.", name)
	}

	for k, v := range pageDefaults {
		if _, e := data[k]; !e {
			data[k] = v
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func loadPasteFromRequest(w http.ResponseWriter, r *http.Request) *pastes.Paste {
	id := r.URL.Path
	return loadPaste(w, id)
}

func loadPaste(w http.ResponseWriter, id string) *pastes.Paste {
	if len(id) < 4 {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return nil
	}

	paste, err := bin.Load(id)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "404 page not found", http.StatusNotFound)
			return nil
		}
	}

	return paste
}

func setCookieDefaults(c *http.Cookie) {
	c.MaxAge = int(12 * time.Hour / time.Second)
	c.Expires = time.Now().Add(12 * time.Hour)
	c.Path = "/"
}

func getOrGenerateCookie(r *http.Request) *http.Cookie {
	cookie, err := r.Cookie("gobin")
	if err != nil {
		if err == http.ErrNoCookie {
			author := bin.GenerateID()
			cookie = &http.Cookie{
				Name:  "gobin",
				Value: author,
			}
		}
	}
	setCookieDefaults(cookie)
	return cookie
}

func sendBinaryAttachment(paste *pastes.Paste, w http.ResponseWriter) {
	attachment, err := paste.Attachment()
	if err != nil {
		log.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer attachment.Close()

	w.Header().Add("Content-Disposition", "attachment; filename="+paste.Title)
	w.Header().Set("Content-Type", paste.Syntax)
	io.Copy(w, attachment)
}

func checkEditCookie(r *http.Request, paste *pastes.Paste) bool {
	cookie, err := r.Cookie("gobin")
	if err != nil {
		return false
	}

	return cookie.Value == paste.Author
}

func handlerForMethod(handlers Handlers) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler, ok := handlers[r.Method]; ok {
			handler(w, r)
			return
		}
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
	})
}

func handlerForPrefix(mux *http.ServeMux, prefix string, handler http.HandlerFunc) {
	mux.Handle(prefix, http.StripPrefix(prefix, http.HandlerFunc(handler)))
}

func main() {
	bin = pastes.New(config.Path)
	loadRecentPastes()
	pageDefaults["Recent"] = recentPastes

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	handlerForPrefix(mux, "/", handlerForMethod(Handlers{"GET": indexHandler}))
	handlerForPrefix(mux, "/p/", handlerForMethod(Handlers{"GET": viewPasteHandler}))
	handlerForPrefix(mux, "/a/", handlerForMethod(Handlers{"GET": viewAttachmentHandler}))
	handlerForPrefix(mux, "/g/", handlerForMethod(Handlers{"GET": getPasteHandler}))
	handlerForPrefix(mux, "/n/", handlerForMethod(Handlers{"GET": newPasteHandler, "POST": saveNewPasteHandler}))
	handlerForPrefix(mux, "/f/", handlerForMethod(Handlers{"POST": uploadNewPasteHandler}))
	handlerForPrefix(mux, "/d/", handlerForMethod(Handlers{"GET": diffPasteHandler}))
	handlerForPrefix(mux, "/e/", handlerForMethod(Handlers{"GET": editPasteHandler, "POST": saveEditPasteHandler}))
	handlerForPrefix(mux, "/s/", handlerForMethod(Handlers{"GET": sharePasteHandler}))

	go func() {
		for {
			<-time.After(config.SaveInterval.Duration)
			saveRecentPastes()
		}
	}()

	http.ListenAndServe(config.Listen, mux)
}

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"io/ioutil"
	"github.com/freman/gobin/pastes"
	"path/filepath"
)

type Ring struct {
	head int
	buf []interface{}
	Empty bool
}

type Recent struct {
	ID string
	Title string
}

var (
	recentPastes *Ring
	recentDirty bool
)

func init() {
	recentPastes = &Ring{-1, make([]interface{}, flRecent), true}
}

func (r *Ring) Push(i interface{}) {
	r.head = (r.head + 1) % len(r.buf)
	r.buf[r.head] = i
}

func (r *Ring) Each(f func(i int, v interface{})) {
	if r.head == -1 {
		return
	}

	head := r.head
	length := len(r.buf)
	for i := 0; i < length; i ++ {
		pos := (head + i + 1) % length
		f(i, r.buf[pos])
	}
}

func (r *Ring) Items() []interface{} {
	result := make([]interface{}, len(r.buf))
	r.Each(func (i int, v interface{}) {
		result[i] = v
	})
	return result;
}

func (r *Recent) String() string {
	return fmt.Sprintf("%s:%s", r.ID, r.Title)
}

func loadRecentPastes() {
	content, err := ioutil.ReadFile(filepath.Join(flPath, "recent"))
	if err != nil {
		if (!os.IsNotExist(err)) {
			log.Println("Problem reading recent pastes: ", err)
		}
		return // we don't care
	}
	list := strings.Split(string(content), "\n");

	for _, v := range list {
		if v != "" {
			recentPastes.Empty = false
			keyval := strings.SplitN(v, ":", 2)
			recentPastes.Push(&Recent{keyval[0], keyval[1]})
		}
	}
}

func saveRecentPastes() {
	if !recentDirty {
		return
	}

	file, err := os.Create(filepath.Join(flPath, "recent"))
	if err != nil {
		log.Println("Can't save recent pastes: ", err)
		return
	}
	defer file.Close()

	recentPastes.Each(func(i int, v interface{}) {
		if v != nil {
			fmt.Fprintln(file, v)
		}
	})
}

func newRecentPaste(p *pastes.Paste) {
	recentDirty = true
	recentPastes.Empty = false
	recentPastes.Push(&Recent{p.ID, p.Title})
}

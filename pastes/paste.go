package pastes

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
)

type Paste struct {
	pastes     *Pastes        `json:"-"`
	saved      bool           `json:"-"`
	attachment multipart.File `json:"-"`
	ID         string         `json:"id"`
	Author     string         `json:"author"`
	Parent     string         `json:"parent,omitempty"`
	Children   Children       `json:"children,omitempty"`
	Title      string         `json:"title,omitempty"`
	Syntax     string         `json:"syntax,omitempty"`
	Expiration int64          `json:"ttl,omitempty"`
	Content    string         `json:"content,omitEmpty"`
}

func (p *Paste) Save() error {
	err := p.pastes.Save(p)
	if err == nil {
		p.saved = true
	}
	return err
}

func (p *Paste) Attach(f multipart.File) {
	p.Content = "binary"
	p.attachment = f
	if p.Syntax == "" {
		p.Syntax = "application/binary"
	}
}

func (p *Paste) Attachment() (io.ReadCloser, error) {
	return p.pastes.loadAttachment(p)
}

func (p *Paste) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Paste) LoadParent() (*Paste, error) {
	if p.Parent != "" {
		return p.pastes.Load(p.Parent)
	}
	return nil, nil
}

func (p *Paste) SetParent(parent *Paste) {
	if parent.Children == nil {
		parent.Children = make(Children)
	}
	parent.Children[p.ID] = true
	p.Parent = parent.ID
	if p.saved {
		p.Save()
	}
}

func (p *Paste) AddChild(child *Paste) {
	if p.Children == nil {
		p.Children = make(Children)
	}
	child.Parent = p.ID
	p.Children[child.ID] = true
	if p.saved {
		p.Save()
	}
}

func (p *Paste) Modify() *Paste {
	np := &Paste{
		pastes:     p.pastes,
		ID:         p.pastes.GenerateID(),
		Author:     p.Author,
		Title:      p.Title,
		Syntax:     p.Syntax,
		Expiration: p.Expiration,
		Content:    p.Content,
	}
	p.AddChild(np)
	if p.saved {
		p.Save()
	}
	return np
}

func (p *Paste) HasChildren() bool {
	return p.Children != nil && len(p.Children) > 0
}

func (p *Paste) ContentFromReader(file multipart.File) {
	buf := new(bytes.Buffer)
	read, _ := io.CopyN(buf, file, p.pastes.RenderMax+1)
	if read > p.pastes.RenderMax {
		// Yeh no, don't render this
		file.Seek(0, 0)
		p.Attach(file)
	} else {
		p.Content = buf.String()
	}
}

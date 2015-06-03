package pastes

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type Pastes struct {
	rand      *rand.Rand
	path      string
	RenderMax int64
}

func New(path string) *Pastes {
	return &Pastes{
		rand.New(rand.NewSource(time.Now().UnixNano())),
		path,
		512000,
	}
}

func (p *Pastes) filePath(id, ext string) string {
	return filepath.Join(p.path, string(id[0]), string(id[1]), string(id[2]), string(id[3]), id+"."+ext)
}

func (p *Pastes) GenerateID() string {
	buf := make([]byte, 20)
	binary.PutVarint(buf, time.Now().UnixNano())
	binary.PutVarint(buf[9:19], p.rand.Int63())
	return base64.URLEncoding.EncodeToString(buf[0:18])
}

func (p *Pastes) New() *Paste {
	return &Paste{ID: p.GenerateID(), pastes: p}
}

func (p *Pastes) Load(id string) (*Paste, error) {
	jsonFile := p.filePath(id, "json")
	input, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}

	result := Paste{}
	err = json.Unmarshal(input, &result)
	if err != nil {
		return nil, err
	}

	result.pastes = p

	return &result, nil
}

func (p *Pastes) loadAttachment(paste *Paste) (io.ReadCloser, error) {
	binFile := p.filePath(paste.ID, "bin")
	return os.Open(binFile)
}

func (p *Pastes) Save(paste *Paste) error {
	output, err := paste.ToJSON()
	jsonFile := p.filePath(paste.ID, "json")
	err = os.MkdirAll(filepath.Dir(jsonFile), 0700)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(jsonFile, output, 0600)
	if err != nil {
		return err
	}

	if paste.attachment != nil {
		binFile := p.filePath(paste.ID, "bin")
		if _, err = os.Stat(binFile); os.IsNotExist(err) {
			file, err := os.Create(binFile)
			if err != nil {
				os.Remove(binFile)
				os.Remove(jsonFile)
				return err
			}
			file.Chmod(0600)
			_, err = io.Copy(file, paste.attachment)
			if err != nil {
				os.Remove(binFile)
				os.Remove(jsonFile)
				return err
			}
			file.Close()
		}
	}

	return nil
}

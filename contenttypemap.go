package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var mWillRender map[string]string
var mContentMatch map[string][]*regexp.Regexp
var saveAs map[string]string

func init() {
	mWillRender = map[string]string{
		"text/javascript":  "javascript",
		"text/php":         "php",
		"text/markdown":    "markdown",
		"application/json": "json",
		"text/x-perl-script": "perl",
	}

	mContentMatch = map[string][]*regexp.Regexp{
		"bash": {regexp.MustCompile(`^#!\s*/bin/(?:b?a)?sh`)},
	}

	saveAs  = map[string]string{
		"javascript": "js",
		"markdown": "md",
		"perl": "pl",
		"bash": "sh",
	}
}

func guessByContentType(contentType string) (t string, ok bool) {
	t, ok = mWillRender[contentType]
	return
}

func guessByContent(r io.ReadSeeker) (syntax string, text bool) {
	buf := new(bytes.Buffer)
	io.CopyN(buf, r, 512)
	r.Seek(0, 0)
	detected := http.DetectContentType(buf.Bytes())
	log.Println("detected: " + detected)
	if strings.HasPrefix(detected, "text/plain;") {
		text = true
		match := buf.String()

	foundType:
		for t, exprs := range mContentMatch {
			for _, expr := range exprs {
				if expr.MatchString(match) {
					syntax = t
					break foundType
				}
			}
		}
	}
	return
}

func getExtension(syntax string) string {
	if syntax == "" {
		return "txt"
	}

	if ext, ok := saveAs[syntax]; ok {
		return ext
	}
	return syntax
}

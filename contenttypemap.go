package main

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var mWillRender map[string]string
var mContentMatch map[string][]*regexp.Regexp
var saveAs map[string]string

func init() {
	mWillRender = map[string]string{
		"application/json":   "json",
		"text/javascript":    "javascript",
		"text/html":          "html",
		"text/markdown":      "markdown",
		"text/php":           "php",
		"text/x-perl-script": "perl",
		"text/xml":           "xml",
	}

	mContentMatch = map[string][]*regexp.Regexp{
		"bash": {regexp.MustCompile(`^#!\s*/bin/(?:b?a)?sh`)},
	}

	saveAs = map[string]string{
		"javascript": "js",
		"markdown":   "md",
		"perl":       "pl",
		"bash":       "sh",
	}
}

func guessByContentType(contentType string) (t string, ok bool) {
	t, ok = mWillRender[contentType]
	return
}

func guessByContent(r io.ReadSeeker) (syntax string, ok bool) {
	buf := new(bytes.Buffer)
	io.CopyN(buf, r, 512)
	r.Seek(0, 0)
	detected := http.DetectContentType(buf.Bytes())
	contentType := detected
	if index := strings.Index(contentType, ";"); index >= 0 {
		contentType = contentType[:index]
	}

	switch contentType {
	case "text/plain":
		ok = true
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
	default:
		syntax, ok = guessByContentType(contentType)
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

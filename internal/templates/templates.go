package templates

import (
	"html/template"
	"log"
)

var Templates *template.Template

func Init() {
	var err error

	Templates, err = template.ParseGlob("templates/*.tmpl")
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}
}

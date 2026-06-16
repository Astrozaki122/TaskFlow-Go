package view

import (
	"html/template"
	"log"
)

var Templates *template.Template

func Init() {
	tmpl, err := template.ParseGlob("templates/*.tmpl")
	if err != nil {
		log.Fatalf("failed to load templates: %v", err)
	}

	Templates = tmpl
}

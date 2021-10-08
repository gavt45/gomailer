package main

import (
	"errors"
	"github.com/flosch/pongo2"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type TemplateEng struct {
	m map[string]*pongo2.Template
}

func (t *TemplateEng) loadTemplates(template_path string) {
	log.Println(template_path)
	fileInfo, err := ioutil.ReadDir(template_path)
	if err != nil {
		panic(err)
	}

	for _, file := range fileInfo {
		log.Println("File name: [", file.Name(), "]")
		var tpl = pongo2.Must(pongo2.FromFile(filepath.Join(template_path, file.Name())))
		t.m[strings.Replace(file.Name(), ".html", "", -1)] = tpl
	}
}

func (t *TemplateEng) execTemplate(template_name string, args map[string]string) (string, error) {
	if template, ok := t.m[template_name]; ok {
		return template.Execute(pongo2.Context{"args": args})
	} else {
		return "", errors.New("No such template!")
	}
}

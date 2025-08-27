package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

var (
	templates    *template.Template
	templateOnce sync.Once
)

var CommonFuncMap = template.FuncMap{
	"toLocalTime": func(t time.Time) time.Time {
		return t.Local()
	},
	"getStatusClass": func(lastSeen time.Time) string {
		if time.Since(lastSeen) < 5*time.Minute {
			return "online"
		}
		return "offline"
	},
}

func initTemplates() {
	templates = template.Must(template.New("").Funcs(CommonFuncMap).ParseGlob(filepath.Join("templates", "*.html")))
}

func GetTemplates() *template.Template {
	templateOnce.Do(initTemplates)
	return templates
}

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := GetTemplates().ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Template execution failed:", err)
	}
}

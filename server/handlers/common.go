package handlers

import (
	"html/template"
	"time"
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

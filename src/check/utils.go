package check

import (
	"github.com/samuel/go-gettext/gettext"
	"html/template"
	"log"
	"net/http"
)

func UpToDate(r *http.Request) bool {
	if r.URL.Query().Get("uptodate") == "0" {
		return false
	}
	return true
}

func Small(r *http.Request) bool {
	if len(r.URL.Query().Get("small")) > 0 {
		return true
	}
	return false
}

func Lang(r *http.Request) string {
	lang := r.URL.Query().Get("lang")
	if len(lang) == 0 {
		lang = "en_US"
	}
	return lang
}

func FuncMap(domain *gettext.Domain) template.FuncMap {
	return template.FuncMap{
		"UnEscaped": func(x string) interface{} {
			return template.HTML(x)
		},
		"UnEscapedURL": func(x string) interface{} {
			return template.URL(x)
		},
		"GetText": func(lang string, text string) string {
			return domain.GetText(lang, text)
		},
	}
}

var Layout *template.Template

func CompileTemplate(domain *gettext.Domain, templateName string) *template.Template {
	var err error
	if Layout == nil {
		Layout = template.New("")
		Layout = Layout.Funcs(FuncMap(domain))
		Layout, err = Layout.ParseFiles("public/base.html")
		if err != nil {
			log.Fatal(err)
		}
	}
	l, err := Layout.Clone()
	if err != nil {
		log.Fatal(err)
	}
	l, err = l.ParseFiles("public/" + templateName)
	if err != nil {
		log.Fatal(err)
	}
	return l
}

package main

import (
	"check"
	"fmt"
	"github.com/samuel/go-gettext/gettext"
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {

	// determine which port to run on
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8000"
	}

	// load i18n
	domain, err := gettext.NewDomain("check", "locale")
	if err != nil {
		log.Fatal(err)
	}

	// Load Tor exits and listen for SIGUSR2 to reload
	Exits := new(check.Exits)
	Exits.Run()

	// add template funcs
	Layout := template.New("")
	Layout = Layout.Funcs(template.FuncMap{
		"UnEscaped": func(x string) interface{} {
			return template.HTML(x)
		},
		"UnEscapedURL": func(x string) interface{} {
			return template.URL(x)
		},
		"GetText": func(lang string, text string) string {
			return domain.GetText(lang, text)
		},
	})

	// load layout
	Layout, err = Layout.ParseFiles(
		"public/index.html",
		"public/bulk.html",
		"public/torbutton.html",
	)
	if err != nil {
		log.Fatal(err)
	}

	// files
	files := http.FileServer(http.Dir("./public"))
	Phttp := http.NewServeMux()
	Phttp.Handle("/torcheck/", http.StripPrefix("/torcheck/", files))
	Phttp.Handle("/", files)

	// routes
	http.HandleFunc("/", check.RootHandler(Layout, Exits, Phttp))
	bulk := check.BulkHandler(Layout, Exits)
	http.HandleFunc("/torbulkexitlist", bulk)
	http.HandleFunc("/cgi-bin/TorBulkExitList.py", bulk)

	// start the server
	log.Printf("Listening on port: %s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))

}

package main

import (
	"fmt"
	"github.com/samuel/go-gettext/gettext"
	"log"
	"net/http"
	"os"
    "html/template"
)

var	Phttp = http.NewServeMux()

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

	// add template funcs
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

    // Load Tor Exits into exits and listen for SIGUSR2 to reload
    exits.Run()

	// routes
	http.HandleFunc("/", RootHandler)
	http.HandleFunc("/torbulkexitlist", BulkHandler)
	http.HandleFunc("/cgi-bin/TorBulkExitList.py", BulkHandler)

	// files
	files := http.FileServer(http.Dir("./public"))
	Phttp.Handle("/torcheck/", http.StripPrefix("/torcheck/", files))
	Phttp.Handle("/", files)

	// start the server
	log.Printf("Listening on port: %s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))

}

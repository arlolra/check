package main

import (
	"check"
	"fmt"
	"github.com/samuel/go-gettext/gettext"
	"log"
	"net/http"
	"os"
)

func main() {

	// write pid
	pid, err := os.Create("check.pid")
	if err != nil {
		log.Fatal(err)
	}
	if _, err = fmt.Fprintf(pid, "%d\n", os.Getpid()); err != nil {
		log.Fatal(err)
	}
	if err = pid.Close(); err != nil {
		log.Fatal(err)
	}

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

	// files
	files := http.FileServer(http.Dir("./public"))
	Phttp := http.NewServeMux()
	Phttp.Handle("/torcheck/", http.StripPrefix("/torcheck/", files))
	Phttp.Handle("/", files)

	// routes
	http.HandleFunc("/", check.RootHandler(check.CompileTemplate(domain, "index.html"), Exits, Phttp))
	bulk := check.BulkHandler(check.CompileTemplate(domain, "bulk.html"), Exits)
	http.HandleFunc("/torbulkexitlist", bulk)
	http.HandleFunc("/cgi-bin/TorBulkExitList.py", bulk)

	// start the server
	log.Printf("Listening on port: %s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))

}

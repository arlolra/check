package main

import (
	"flag"
	"fmt"
	"github.com/samuel/go-gettext/gettext"
	"log"
	"net/http"
	"os"
	"path"
)

func main() {

	// command line args
	logPath := flag.String("log", "", "path to log file; otherwise stdout")
	pidPath := flag.String("pid", "./", "path to create pid")
	port := flag.Int("port", 8000, "port to listen on")
	flag.Parse()

	// log to file
	if len(*logPath) > 0 {
		f, err := os.Create(*logPath)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(f)
	}

	// write pid
	pid, err := os.Create(path.Join(*pidPath, "check.pid"))
	if err != nil {
		log.Fatal(err)
	}
	if _, err = fmt.Fprintf(pid, "%d\n", os.Getpid()); err != nil {
		log.Fatal(err)
	}
	if err = pid.Close(); err != nil {
		log.Fatal(err)
	}

	// load i18n
	domain, err := gettext.NewDomain("check", "locale")
	if err != nil {
		log.Fatal(err)
	}

	// Load Tor exits and listen for SIGUSR2 to reload
	exits := new(Exits)
	exits.Run()

	// files
	files := http.FileServer(http.Dir("./public"))
	Phttp := http.NewServeMux()
	Phttp.Handle("/torcheck/", http.StripPrefix("/torcheck/", files))
	Phttp.Handle("/", files)

	// routes
	http.HandleFunc("/", RootHandler(CompileTemplate(domain, "index.html"), exits, domain, Phttp))
	bulk := BulkHandler(CompileTemplate(domain, "bulk.html"), exits, domain)
	http.HandleFunc("/torbulkexitlist", bulk)
	http.HandleFunc("/cgi-bin/TorBulkExitList.py", bulk)

	// start the server
	log.Printf("Listening on port: %d\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))

}

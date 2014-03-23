package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/samuel/go-gettext/gettext"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// page model
type Page struct {
	IsTor       bool
	NotUpToDate bool
	Small       bool
	Fingerprint string
	OnOff       string
	Lang        string
	IP          string
	Locales     map[string]string
}

func RootHandler(Layout *template.Template, Exits *Exits, domain *gettext.Domain, Phttp *http.ServeMux, Locales map[string]string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// serve public files
		if len(r.URL.Path) > 1 {
			Phttp.ServeHTTP(w, r)
			return
		}

		var (
			err         error
			isTor       bool
			fingerprint string
		)

		// get remote ip
		host, err := GetHost(r)

		// determine if we're in Tor
		if err != nil {
			isTor = false
		} else {
			fingerprint, isTor = Exits.IsTor(host)
		}

		// short circuit for torbutton
		if len(r.URL.Query().Get("TorButton")) > 0 {
			WriteHTMLBuf(w, r, Layout, domain, "torbutton.html", Page{IsTor: isTor})
			return
		}

		// string used for classes and such
		// in the template
		var onOff string
		if isTor {
			onOff = "on"
		} else {
			onOff = "off"
		}

		// instance of your page model
		p := Page{
			isTor,
			!UpToDate(r),
			Small(r),
			fingerprint,
			onOff,
			Lang(r),
			host,
			Locales,
		}

		// render the template
		WriteHTMLBuf(w, r, Layout, domain, "index.html", p)
	}

}

type IPResp struct {
	IsTor bool
}

func APIHandler(Exits *Exits) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		isTor := false
		if host, err := GetHost(r); err == nil {
			_, isTor = Exits.IsTor(host)
		}
		ip, _ := json.Marshal(IPResp{isTor})
		w.Write(ip)
	}
}

func BulkHandler(Layout *template.Template, Exits *Exits, domain *gettext.Domain) http.HandlerFunc {

	ApiPath := regexp.MustCompile("^/api/")

	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		ip := q.Get("ip")
		if net.ParseIP(ip) == nil {
			WriteHTMLBuf(w, r, Layout, domain, "bulk.html", Page{Lang: "en"})
			return
		}

		port, port_str := GetQS(q, "port", 80)
		n, n_str := GetQS(q, "n", 16)

		w.Header().Set("Last-Modified", Exits.UpdateTime.UTC().Format(http.TimeFormat))

		if q.Get("format") == "json" || ApiPath.MatchString(r.URL.Path) {
			w.Header().Set("Content-Type", "application/json")
			Exits.DumpJSON(w, n, ip, port)
		} else {
			str := fmt.Sprintf("# This is a list of all Tor exit nodes from the past %d hours that can contact %s on port %d #\n", n, ip, port)
			str += fmt.Sprintf("# You can update this list by visiting https://check.torproject.org/cgi-bin/TorBulkExitList.py?ip=%s%s%s #\n", ip, port_str, n_str)
			str += fmt.Sprintf("# This file was generated on %v #\n", Exits.UpdateTime.UTC().Format(time.UnixDate))
			fmt.Fprintf(w, str)
			Exits.Dump(w, n, ip, port)
		}

	}

}

func WriteHTMLBuf(w http.ResponseWriter, r *http.Request, Layout *template.Template, domain *gettext.Domain, tmp string, p Page) {
	buf := new(bytes.Buffer)

	// render template
	if err := Layout.ExecuteTemplate(buf, tmp, p); err != nil {
		log.Printf("Layout.ExecuteTemplate: %v", err)
		http.Error(w, domain.GetText(p.Lang, "Sorry, your query failed or an unexpected response was received."), http.StatusInternalServerError)
		return
	}

	// set some headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == "HEAD" {
		w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
		return
	}

	// write buf
	if _, err := io.Copy(w, buf); err != nil {
		log.Printf("io.Copy: %v", err)
	}
}

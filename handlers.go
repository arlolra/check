package main

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"time"
)

var Locales = GetLocaleList()

// page model
type Page struct {
	IsTor       bool
	UpToDate    bool
	NotSmall    bool
	Fingerprint string
	OnOff       string
	Lang        string
	IP          string
	Extra       string
	Locales     map[string]string
}

func RootHandler(Layout *template.Template, Exits *Exits, Phttp *http.ServeMux) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		// serve public files
		if len(r.URL.Path) > 1 {
			Phttp.ServeHTTP(w, r)
			return
		}

		// get remote ip
		host := r.Header.Get("X-Forwarded-For")
		var err error
		if len(host) == 0 {
			host, _, err = net.SplitHostPort(r.RemoteAddr)
		}

		var (
			isTor       bool
			fingerprint string
		)
		// determine if we're in Tor
		if err != nil {
			isTor = false
		} else {
			fingerprint, isTor = Exits.IsTor(host)
		}

		// short circuit for torbutton
		if len(r.URL.Query().Get("TorButton")) > 0 {
			if err := Layout.ExecuteTemplate(w, "torbutton.html", isTor); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
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

		small := Small(r)
		upToDate := UpToDate(r)

		// querystring params
		extra := ""
		if small {
			extra += "&small=1"
		}
		if !upToDate {
			extra += "&uptodate=0"
		}

		// instance of your page model
		p := Page{
			isTor,
			isTor && !upToDate,
			!small,
			fingerprint,
			onOff,
			Lang(r),
			host,
			extra,
			Locales,
		}

		// render the template
		if err := Layout.ExecuteTemplate(w, "index.html", p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	}

}

func BulkHandler(Layout *template.Template, Exits *Exits) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		ip := q.Get("ip")
		if net.ParseIP(ip) == nil {
			if err := Layout.ExecuteTemplate(w, "bulk.html", nil); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		port, port_str := GetQS(q, "port", 80)
		n, n_str := GetQS(q, "n", 16)

		str := fmt.Sprintf("# This is a list of all Tor exit nodes from the past %d hours that can contact %s on port %d #\n", n, ip, port)
		str += fmt.Sprintf("# You can update this list by visiting https://check.torproject.org/cgi-bin/TorBulkExitList.py?ip=%s%s%s #\n", ip, port_str, n_str)
		str += fmt.Sprintf("# This file was generated on %v #\n", Exits.UpdateTime.UTC().Format(time.UnixDate))
		fmt.Fprintf(w, str)

		Exits.Dump(w, n, ip, port)
	}

}

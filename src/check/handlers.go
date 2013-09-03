package check

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strconv"
	"time"
)

var (
	// locales map
	Locales = getLocaleList()
)

// page model
type Page struct {
	IsTor    bool
	UpToDate bool
	NotSmall bool
	OnOff    string
	Lang     string
	IP       string
	Extra    string
	Locales  map[string]string
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

		// determine if we're in Tor
		var isTor bool
		if err != nil {
			isTor = false
		} else {
			isTor = Exits.IsTor(host)
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

		ip := r.URL.Query().Get("ip")
		if net.ParseIP(ip) == nil {
			if err := Layout.ExecuteTemplate(w, "bulk.html", nil); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		port_str := r.URL.Query().Get("port")
		port, err := strconv.Atoi(port_str)
		port_str = "&port=" + port_str
		if err != nil {
			port = 80
			port_str = ""
		}

		str := fmt.Sprintf("# This is a list of all Tor exit nodes that can contact %s on Port %d #\n", ip, port)
		str += fmt.Sprintf("# You can update this list by visiting https://check.torproject.org/cgi-bin/TorBulkExitList.py?ip=%s%s #\n", ip, port_str)
		str += fmt.Sprintf("# This file was generated on %v #\n", Exits.UpdateTime.UTC().Format(time.UnixDate))
		str += Exits.Dump(ip, port)
		fmt.Fprintf(w, str)

	}

}

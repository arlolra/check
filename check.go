package main

import (
	"fmt"
	"github.com/samuel/go-gettext/gettext"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

// page model
type Page struct {
	IsTor   bool
	OnOff   string
	Lang    string
	IP      string
	InTor   string
	YourIP  string
	Locales map[string]string
}

// language domain
var domain *gettext.Domain

// layout template
var layout = template.New("")

// public file server
var phttp = http.NewServeMux()

// locales map
var locales = map[string]string{
	"ar":    "&#1593;&#1585;&#1576;&#1610;&#1577;&nbsp;(Arabiya)",
	"bms":   "Burmese",
	"cs":    "&#269;esky",
	"da":    "Dansk",
	"de":    "Deutsch",
	"el":    "&#917;&#955;&#955;&#951;&#957;&#953;&#954;&#940;&nbsp;(Ellinika)",
	"en_US": "English",
	"es":    "Espa&ntilde;ol",
	"et":    "Estonian",
	"fa_IR": "&#1601;&#1575;&#1585;&#1587;&#1740; (F&#257;rs&#299;)",
	"fr":    "Fran&ccedil;ais",
	"it_IT": "Italiano",
	"ja":    "&#26085;&#26412;&#35486;&nbsp;(Nihongo)",
	"nb":    "Norsk&nbsp;(Bokm&aring;l)",
	"nl":    "Nederlands",
	"pl":    "Polski",
	"pt":    "Portugu&ecirc;s",
	"pt_BR": "Portugu&ecirc;s do Brasil",
	"ro":    "Rom&acirc;n&#259;",
	"fi":    "Suomi",
	"ru":    "&#1056;&#1091;&#1089;&#1089;&#1082;&#1080;&#1081;&nbsp;(Russkij)",
	"th":    "Thai",
	"tr":    "T&uuml;rk&ccedil;e",
	"uk":    "&#1091;&#1082;&#1088;&#1072;&#1111;&#1085;&#1089;&#1100;&#1082;&#1072;&nbsp;(Ukrajins\"ka)",
	"vi":    "Vietnamese",
	"zh_CN": "&#20013;&#25991;(&#31616;)",
}

// return raw html
func unescaped(x string) interface{} {
	return template.HTML(x)
}

// rejigger this to not make dns queries
func IsTor(remoteAddr string) bool {
	if net.ParseIP(remoteAddr).To4() == nil {
		return false
	}
	ips := strings.Split(remoteAddr, ".")
	var ip string
	for i := len(ips) - 1; i >= 0; i-- {
		ip += ips[i] + "."
	}
	host := "80.38.229.70.31.ip-port.exitlist.torproject.org"
	addresses, err := net.LookupHost(ip + host)
	if err != nil {
		return false
	}
	inTor := true
	for _, val := range addresses {
		if val != "127.0.0.2" {
			inTor = false
			break
		}
	}
	return inTor
}

func RootHandler(w http.ResponseWriter, r *http.Request) {

	// serve public files
	if len(r.URL.Path) > 1 {
		phttp.ServeHTTP(w, r)
		return
	}

	// determine if we"re in Tor
	isTor := IsTor(r.RemoteAddr)

	// prepare Tor string
	var inTor string
	var onOff string
	if isTor {
		onOff = "on"
		inTor = "Congratulations. Your browser is configured to use Tor."
	} else {
		onOff = "off"
		inTor = "Sorry. You are not using Tor."
	}

	// determin which language to use. default to english
	lang := r.URL.Query().Get("lang")
	if len(lang) == 0 {
		lang = "en_US"
	}

	// instance of your page model
	p := Page{
		isTor,
		onOff,
		lang,
		r.RemoteAddr,
		domain.GetText(lang, inTor),
		domain.GetText(lang, "Your IP address appears to be: "),
		locales,
	}

	// render the template
	layout.ExecuteTemplate(w, "index.html", p)

}

func main() {

	// determine which port to run on
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "9000"
	}

	// load i18n
	var err error
	domain, err = gettext.NewDomain("check", "locale")
	if err != nil {
		log.Fatal(err)
	}

	// add template funcs
	layout = layout.Funcs(template.FuncMap{
		"unescaped": unescaped,
	})

	// load layout
	layout, err = layout.ParseFiles("public/index.html")
	if err != nil {
		log.Fatal(err)
	}

	// routes
	http.HandleFunc("/", RootHandler)
	phttp.Handle("/", http.FileServer(http.Dir("./public")))

	// start the server
	log.Printf("Listening on port: %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))

}

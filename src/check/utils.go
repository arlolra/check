package check

import (
	"encoding/json"
	"github.com/samuel/go-gettext/gettext"
	"html/template"
	"io"
	"io/ioutil"
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

type locale struct {
	Code string
	Name string
	// Unnecessary fields
	//PluralEquation string
	//NumPlurals     float64 `json:"nplurals"`
}

func getLocaleList() map[string]string {
	// TODO: This should be it's own translation file
	haveTranslatedNames := map[string]string{
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

	// for all folders in locale which match a locale from https://www.transifex.com/api/2/languages/
	// use the language name unless we have an override
	webLocales, err := fetchTranslationLocales()
	if err != nil {
		log.Printf("Failed to get up to date language list, using fallback.")
		return haveTranslatedNames
	}

	return getInstalledLocales(webLocales, haveTranslatedNames)
}

func fetchTranslationLocales() (map[string]locale, error) {
	resp, err := http.Get("https://www.transifex.com/api/2/languages/")
	if err != nil {
		return nil, err
	}

	webLocales := make(map[string]locale)
	// Parse the api response into a list of possible locales
	dec := json.NewDecoder(resp.Body)
	for {
		var webList []locale
		if err = dec.Decode(&webList); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		// The api returns an array, so we need to map it
		for _, l := range webList {
			webLocales[l.Code] = l
		}
	}

	return webLocales, nil
}

// Get a list of all languages installed in our locale folder with translations if available
func getInstalledLocales(webLocales map[string]locale, nameTranslations map[string]string) map[string]string {
	localFiles, err := ioutil.ReadDir("locale")
	if err != nil {
		log.Fatal("No locales found in './locale'. Try running 'make i18n'.")
	}

	locales := make(map[string]string, len(localFiles))
	for _, f := range localFiles {
		// TODO: Ensure a language has 100% of the template file
		// Currently this is what should be on the torcheck_completed
		// branch on the translations git should be, so we don't really
		// have to check it in theory...
		code := f.Name()

		// Only accept folders which have corresponding locale
		if !f.IsDir() || webLocales[code] == (locale{}) {
			continue
		}

		// If we have a translated name for a given locale, use it
		if transName := nameTranslations[code]; transName != "" {
			locales[code] = transName
		} else {
			locales[code] = webLocales[code].Name
		}
	}

	return locales
}

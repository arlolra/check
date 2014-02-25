package main

import (
	"encoding/json"
	"fmt"
	"github.com/samuel/go-gettext/gettext"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
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

func GetQS(q url.Values, param string, deflt int) (num int, str string) {
	str = q.Get(param)
	num, err := strconv.Atoi(str)
	if err != nil {
		num = deflt
		str = ""
	} else {
		str = fmt.Sprintf("&%s=%s", param, str)
	}
	return
}

var HaveManual = map[string]bool{
	"ar":    true,
	"zh_CN": true,
	"cs":    true,
	"nl":    true,
	"en":    true,
	"fa":    true,
	"fr":    true,
	"de":    true,
	"el":    true,
	"hu":    true,
	"it":    true,
	"ja":    true,
	"lv":    true,
	"nb":    true,
	"pl":    true,
	"pt_BR": true,
	"ru":    true,
	"es":    true,
	"sv":    true,
	"tr":    true,
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
		"Equal": func(one string, two string) bool {
			return one == two
		},
		"Not": func(b bool) bool {
			return !b
		},
		"And": func(a bool, b bool) bool {
			return a && b
		},
		"UserManual": func(lang string) string {
			if _, ok := HaveManual[lang]; !ok {
				lang = "en"
			}
			return lang
		},
	}
}

var Layout *template.Template

func CompileTemplate(base string, domain *gettext.Domain, templateName string) *template.Template {
	if Layout == nil {
		Layout = template.New("")
		Layout = Layout.Funcs(FuncMap(domain))
		Layout = template.Must(Layout.ParseFiles(
			path.Join(base, "public/base.html"),
			path.Join(base, "public/torbutton.html"),
		))
	}
	l, err := Layout.Clone()
	if err != nil {
		log.Fatal(err)
	}
	return template.Must(l.ParseFiles(path.Join(base, "public/", templateName)))
}

type locale struct {
	Code string
	Name string
}

func GetLocaleList(base string) map[string]string {
	// populated from https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
	haveTranslatedNames := map[string]string{
		"ar": "العربية",
		"bs": "Bosanski jezik",
		"ca": "Català",
		"cs": "čeština",
		"cy": "Cymraeg",
		"da": "Dansk",
		"de": "Deutsch",
		"el": "ελληνικά",
		"es": "Español",
		"et": "Eesti",
		"eu": "Euskara",
		"fa": "فارسی",
		"fi": "Suomi",
		"fr": "Français",
		"gl": "Galego",
		"he": "עברית",
		"hi": "हिन्दी, हिंदी",
		"hr": "Hrvatski jezik",
		"hu": "Magyar",
		"id": "Bahasa Indonesia",
		"it": "Italiano",
		"ja": "日本語",
		"km": "មែរ",
		"kn": "ಕನ್ನಡ",
		"ko": "한국어",
		"lv": "Latviešu valoda",
		"my": "ဗမာစာ",
		"nb": "Norsk bokmål",
		"nl": "Nederlands",
		"pa": "ਪੰਜਾਬੀ",
		"pl": "Język polski",
		"pt": "Português",
		"ru": "русский язык",
		"sk": "Slovenčina",
		"sl": "Slovenski jezik",
		"sr": "српски језик",
		"sv": "Svenska",
		"th": "ไทย",
		"tr": "Türkçe",
		"uk": "українська мова",
	}

	// for all folders in locale which match a locale from https://www.transifex.com/api/2/languages/
	// use the language name unless we have an override
	webLocales, err := FetchTranslationLocales(base)
	if err != nil {
		log.Printf("Failed to get up to date language list, using fallback.")
		return haveTranslatedNames
	}

	return GetInstalledLocales(base, webLocales, haveTranslatedNames)
}

func FetchTranslationLocales(base string) (map[string]locale, error) {
	file, err := os.Open(path.Join(base, "data/langs"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	webLocales := make(map[string]locale)
	// Parse the api response into a list of possible locales
	dec := json.NewDecoder(file)
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
func GetInstalledLocales(base string, webLocales map[string]locale, nameTranslations map[string]string) map[string]string {
	localFiles, err := ioutil.ReadDir(path.Join(base, "locale"))

	if err != nil {
		log.Print("No locales found in 'locale'. Try running 'make i18n'.")
		log.Fatal(err)
	}

	locales := make(map[string]string, len(localFiles))
	locales["en_US"] = "English"

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

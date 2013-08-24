package main

import (
	"net"
	"net/http"
)

func IsTor(remoteAddr string) bool {
	if net.ParseIP(remoteAddr).To4() == nil {
		return false
	}
	return exits.list[remoteAddr].CanExit(443)
}

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

// determine which language to use. default to english
func Lang(r *http.Request) string {
	lang := r.URL.Query().Get("lang")
	if len(lang) == 0 {
		lang = "en_US"
	}
	return lang
}

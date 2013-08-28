package check

import (
	"io"
	"log"
	"net/http"
)

var (
	HIDDEN_SERVICE_HOSTNAME_PATH = Env("HIDDEN_SERVICE_HOSTNAME_PATH", "./hidden_service/hostname")
	ONION_DOMAIN                 = loadHiddenServiceHostname()
)

func HiddenServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Handle asset requests (favicon etc)
	if assetHandler(w, r) {
		return
	}

	log.Printf("Hidden service request from: %s", r.Host)
	if key := parseIdentifier(r); len(key) > 0 {
		log.Printf("With key: %s", key)
		// Confirm on the channel
		c := pending[key]
		if c != nil {
			log.Printf("Onion: Sending confirmation")
			io.WriteString(w, r.RemoteAddr)
			c <- true
			return
		}

		log.Printf("Onion: Finished request")
		io.WriteString(w, "Theoretical Success, but slow connection!")
		// Else - Looks like a timeout probably occurred?
		// but... if the onion page loaded... you're on tor...?
	}
}

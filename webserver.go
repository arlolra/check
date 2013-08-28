package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func webHandler(w http.ResponseWriter, r *http.Request) {
	// Handle asset requests (favicon etc)
	if assetHandler(w, r) {
		return
	}

	log.Printf("Web request from: %s", r.Host)
	if key := parseIdentifier(r); len(key) > 0 {
		waitForHiddenService(w, r, key)
	} else {
		// Generate a new key
		// TODO: Should ensure key is not already in the pending list
		key := generateIdentifier(32)
		pending[key] = make(chan bool, 1)

		// Load initial page with redirects to keyed request
		io.WriteString(w, fmt.Sprintf("<iframe src='http://%s/?id=%s'></iframe>", ONION_DOMAIN, key))
		io.WriteString(w, fmt.Sprintf("<iframe src='http://%s/?id=%s'></iframe>", r.Host, key))

		//http.Redirect(w, r, fmt.Sprintf("?id=%s", key), 302)
	}
}

func waitForHiddenService(w http.ResponseWriter, r *http.Request, key Identifier) {
	log.Printf("Waiting for onion confirmation with key: %s", key)

	// Remove the channel whenever we're finished with this client
	defer delete(pending, key)

	for i := 1; i < MAX_TIMEOUTS+1; i++ {
		// Since we have an ID from a regular web request, it should be ajax/iframe,
		// so we can hang the request until we're ready to pass or fail it
		select {
		case v := <-pending[key]:
			// Success!
			log.Printf("Recieved pending confirmation: %v for key: %s", v, key)
			io.WriteString(w, "Success from web!!!!!!")
			return

		case <-time.After(10 * time.Second): // * time.Minute):
			// Timeout
			log.Printf("Timeout %v/%v  for %s", i, MAX_TIMEOUTS, key)
			io.WriteString(w, fmt.Sprintf("Timeout %v of %v\n", i, MAX_TIMEOUTS))

			// Flush the progress to the client
			// TODO : Send more data or else Chrome doesn't show anything (js could handle this)
			w.(http.Flusher).Flush()
		}
	}
}

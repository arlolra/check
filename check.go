package main

/*
	Alternative implementation - probably a bit slower but less reliant on exit list's freshness:

	On initial request, render 2 iframes (or js enabled page) and generate a unique key
		Create a channel mapped to the unique key
		One of the iframes will make a request to the hidden service with the key
			If the hidden service is hit, then send a message on the keyed channel
		One of the iframes will make a request to the web service with the key
			Hangs the request until a message is recieved on the keyed channel
			If a timeout is reached, retry a few times
*/
var (
	HIDDEN_SERVICE_HOSTNAME_PATH = env("HIDDEN_SERVICE_HOSTNAME_PATH", "./hidden_service/hostname")
	ONION_PORT                   = env("ONIONPORT", "8000")
	WEB_PORT                     = env("PORT", "8080")
	MAX_TIMEOUTS                 = 5
	ONION_DOMAIN                 = loadHiddenServiceHostname()

	pending = make(map[Identifier]chan bool)
)

func main() {
	setupCPU()
	go startServer(ONION_PORT, "Hidden", hiddenServiceHandler)
	startServer(WEB_PORT, "Web", webHandler)
}

package main

import (
	"check"
)

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
	ONION_PORT = check.Env("ONIONPORT", "8000")
	WEB_PORT   = check.Env("PORT", "8080")
)

func main() {
	check.SetupCPU()
	go check.StartServer(ONION_PORT, "Hidden", check.HiddenServiceHandler)
	check.StartServer(WEB_PORT, "Web", check.WebHandler)
}

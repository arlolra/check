package check

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

type Identifier string

func parseIdentifier(r *http.Request) Identifier {
	return Identifier(r.FormValue("id"))
}

func generateIdentifier(length int) Identifier {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		log.Fatal(fmt.Printf("Unable to read from /dev/urandom\n%v", err))
	}

	bytes := make([]byte, length)
	f.Read(bytes)
	f.Close()

	return Identifier(fmt.Sprintf("%x", bytes))
}

func assetHandler(w http.ResponseWriter, r *http.Request) bool {
	if len(r.URL.Path) > 1 {
		// TODO: Serve static assets
		return true
	}

	return false
}

func StartServer(port, serverName string, handler http.HandlerFunc) {
	log.Printf("%s service listening on port: %s\n", serverName, port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), http.HandlerFunc(handler)))
}

// Looks for an environment variable `varname`,
// else use provided `defaultval` value
func Env(varname, defaultval string) string {
	temp := os.Getenv(varname)
	if len(temp) == 0 {
		temp = defaultval
	}

	return temp
}

// Ensure we're running goroutines across all available logical cores
func SetupCPU() {
	cpuCount := runtime.NumCPU()
	runtime.GOMAXPROCS(cpuCount)
	log.Printf("Starting services with %v CPUs active", cpuCount)
}

// Gets the domain for the hidden service from the description file
func loadHiddenServiceHostname() string {
	bytes, err := ioutil.ReadFile(HIDDEN_SERVICE_HOSTNAME_PATH)

	if err != nil {
		log.Fatal("Unable to load the hidden service's hostname file")
	}

	hostname := strings.Trim(string(bytes), "\r\n ")
	log.Printf("Using '%s' as hidden service domain", hostname)

	return hostname
}

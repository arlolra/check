package main

import (
    "strings"
    "bufio"
    "time"
    "fmt"
    "os"
    "log"
    "net/http"
    "strconv"
)

type Port struct {
	min int
	max int
}

type Policy struct {
	accept bool
	ports  []Port
}

func (p Policy) CanExit(exitPort int) bool {
	if len(p.ports) == 0 {
		return false
	}
	for _, port := range p.ports {
		if port.min <= exitPort && exitPort <= port.max {
			return p.accept
		}
	}
	return !p.accept
}

type Exits struct {
    list       map[string]Policy
    updateTime time.Time
}

func (e *Exits) Dump(w *http.ResponseWriter, port int) {
    fmt.Fprintf(*w, "# This is a list of all Tor exit nodes that can contact %s on Port %d #\n", "X", port)
	fmt.Fprintf(*w, "# You can update this list by visiting https://check.torproject.org/cgi-bin/TorBulkExitList.py?ip=%s%d #\n", "X", port)
	fmt.Fprintf(*w, "# This file was generated on %v #\n", e.updateTime.UTC().Format(time.UnixDate))

	for key, val := range e.list {
		if val.CanExit(port) {
			fmt.Fprintf(*w, "%s\n", key)
		}
	}
}

func (e *Exits) Load() {
	file, err := os.Open("data/exit-policies")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	exits := make(map[string]Policy)
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		strs := strings.Fields(scan.Text())
        if len(strs) > 0 {
            policy := Policy{}
            if strs[1] == "accept" {
                policy.accept = true
            }
            ports := strings.Split(strs[2], ",")
            for _, p := range ports {
                s := strings.Split(p, "-")
                min, err := strconv.Atoi(s[0])
                if err != nil {
                    log.Fatal(err)
                }
                port := Port{
                    min: min,
                    max: min,
                }
                if len(s) > 1 {
                    port.max, err = strconv.Atoi(s[1])
                    if err != nil {
                        log.Fatal(err)
                    }
                }
                policy.ports = append(policy.ports, port)
            }
            exits[strs[0]] = policy
        }
    }

    if err = scan.Err(); err != nil {
        log.Fatal(err)
    }

    // swap in exits
    e.list = exits
    e.updateTime = time.Now()
}

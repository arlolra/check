package check

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
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
	reloadChan chan os.Signal
}

func (e *Exits) Dump(port int) string {
	str := fmt.Sprintf("# This is a list of all Tor exit nodes that can contact %s on Port %d #\n", "X", port)
	str += fmt.Sprintf("# You can update this list by visiting https://check.torproject.org/cgi-bin/TorBulkExitList.py?ip=%s%d #\n", "X", port)
	str += fmt.Sprintf("# This file was generated on %v #\n", e.updateTime.UTC().Format(time.UnixDate))

	for key, val := range e.list {
		if val.CanExit(port) {
			str += fmt.Sprintf("%s\n", key)
		}
	}

	return str
}

func (e *Exits) IsTor(remoteAddr string, port int) bool {
	if net.ParseIP(remoteAddr).To4() == nil {
		return false
	}
	return e.list[remoteAddr].CanExit(port)
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

func (e *Exits) Run() {
	e.reloadChan = make(chan os.Signal, 1)
	signal.Notify(e.reloadChan, syscall.SIGUSR2)
	go func() {
		for {
			<-e.reloadChan
			e.Load()
			log.Println("Exit list reloaded.")
		}
	}()
	e.Load()
}

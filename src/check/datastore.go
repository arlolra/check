package check

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// stem's ExitPolicyRule class, sort of
type Rule struct {
	IsAccept  bool
	Address   string
	AddressIP net.IP
	MinPort   int
	MaxPort   int
}

func ValidPort(port int) bool {
	return port >= 0 && port < 65536
}

func (r Rule) IsMatch(address net.IP, port int) bool {
	if r.AddressIP != nil && !r.AddressIP.Equal(address) {
		return false
	}
	if port < r.MinPort || port > r.MaxPort {
		return false
	}
	return true
}

type AddressPort struct {
	Address string
	Port    int
}

type Policy struct {
	Address          string
	Rules            []Rule
	IsAllowedDefault bool
	CanExitCache     map[AddressPort]bool
}

func (p Policy) CanExit(ap AddressPort) (can bool) {
	if can, ok := p.CanExitCache[ap]; ok {
		return can // Explicit return for shadowed var
	}
	// Update the cache *after* we return
	defer func() { p.CanExitCache[ap] = can }()

	addr := net.ParseIP(ap.Address)
	if addr != nil && ValidPort(ap.Port) {
		for _, rule := range p.Rules {
			if rule.IsMatch(addr, ap.Port) {
				can = rule.IsAccept
				return
			}
		}
	}

	can = p.IsAllowedDefault
	return
}

type Exits struct {
	List        map[string]Policy
	UpdateTime  time.Time
	ReloadChan  chan os.Signal
	isTorLookup map[string]bool
}

func (e *Exits) Dump(ip string, port int) string {
	// This should cause less GC
	var buf bytes.Buffer

	e.getAllExits(ip, port, func(can_exit_ip string) {
		buf.WriteString(can_exit_ip)
		buf.WriteRune('\n')
	})

	return buf.String()
}

func (e *Exits) getAllExits(ip string, port int, fn func(ip string)) {
	ap := AddressPort{ip, port}
	for key, val := range e.List {
		if val.CanExit(ap) {
			fn(key)
		}
	}
}

var DefaultTarget = AddressPort{"38.229.70.31", 443}

func (e *Exits) IsTor(remoteAddr string) bool {
	return e.isTorLookup[remoteAddr]
}

func (e *Exits) Load() {
	file, err := os.Open("data/exit-policies")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	exits := make(map[string]Policy)
	dec := json.NewDecoder(file)
	for {
		var p Policy
		if err = dec.Decode(&p); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		for _, r := range p.Rules {
			r.AddressIP = net.ParseIP(r.Address)
		}
		p.CanExitCache = make(map[AddressPort]bool)
		exits[p.Address] = p
	}

	// swap in exits
	e.List = exits
	e.UpdateTime = time.Now()

	// Precompute after the new list is swapped
	e.preComputeTorList()
}

func (e *Exits) preComputeTorList() {
	newmap := make(map[string]bool)
	e.getAllExits(DefaultTarget.Address, DefaultTarget.Port, func(ip string) {
		newmap[ip] = true
	})

	e.isTorLookup = newmap
}

func (e *Exits) Run() {
	e.ReloadChan = make(chan os.Signal, 1)
	signal.Notify(e.ReloadChan, syscall.SIGUSR2)
	go func() {
		for {
			<-e.ReloadChan
			e.Load()
			log.Println("Exit list reloaded.")
		}
	}()
	e.Load()
}

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

type Rule struct {
	IsAccept          bool
	IsAddressWildcard bool
	Address           string
	Mask              string
	IP                net.IP
	IPNet             *net.IPNet
	MinPort           int
	MaxPort           int
}

func (r Rule) IsMatch(address net.IP, port int) bool {
	if !r.IsAddressWildcard {
		if r.IPNet != nil && !r.IPNet.Contains(address) {
			return false
		} else if r.IP != nil && !r.IP.Equal(address) {
			return false
		}
	}
	if port < r.MinPort || port > r.MaxPort {
		return false
	}
	return true
}

func ValidPort(port int) bool {
	return port >= 0 && port < 65536
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
		return can // explicit return for shadowed var
	}
	// update the cache after we return
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
	IsTorLookup map[string]bool
}

func (e *Exits) Dump(ip string, port int) string {
	ap := AddressPort{ip, port}
	var buf bytes.Buffer

	e.GetAllExits(ap, func(exit string) {
		buf.WriteString(exit)
		buf.WriteRune('\n')
	})

	return buf.String()
}

func (e *Exits) GetAllExits(ap AddressPort, fn func(ip string)) {
	for key, val := range e.List {
		if val.CanExit(ap) {
			fn(key)
		}
	}
}

var DefaultTarget = AddressPort{"38.229.70.31", 443}

func (e *Exits) PreComputeTorList() {
	newmap := make(map[string]bool)
	e.GetAllExits(DefaultTarget, func(ip string) {
		newmap[ip] = true
	})
	e.IsTorLookup = newmap
}

func (e *Exits) IsTor(remoteAddr string) bool {
	return e.IsTorLookup[remoteAddr]
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
		for i := range p.Rules {
			r := &p.Rules[i]
			if !r.IsAddressWildcard {
				r.IP = net.ParseIP(r.Address)
				if mask := net.ParseIP(r.Mask); r.IP != nil && mask != nil {
					m := make(net.IPMask, len(mask))
					copy(m, mask)
					r.IPNet = &net.IPNet{r.IP.Mask(m), m}
				}
			}
		}
		p.CanExitCache = make(map[AddressPort]bool)
		exits[p.Address] = p
	}

	// swap in exits
	e.List = exits
	e.UpdateTime = time.Now()

	// precompute IsTor
	e.PreComputeTorList()
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

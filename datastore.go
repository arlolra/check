package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sort"
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
		if r.IPNet != nil {
			if !r.IPNet.Contains(address) {
				return false
			}
		} else {
			if !r.IP.Equal(address) {
				return false
			}
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
}

func (p Policy) CanExit(ap AddressPort) bool {
	addr := net.ParseIP(ap.Address)
	if addr != nil && ValidPort(ap.Port) {
		for _, rule := range p.Rules {
			if rule.IsMatch(addr, ap.Port) {
				return rule.IsAccept
			}
		}
	}
	return p.IsAllowedDefault
}

type PolicyList []Policy

func (p PolicyList) Less(i, j int) bool {
	return p[i].Address < p[j].Address
}

func (p PolicyList) Len() int {
	return len(p)
}

func (p PolicyList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type Exits struct {
	List        PolicyList
	UpdateTime  time.Time
	ReloadChan  chan os.Signal
	IsTorLookup map[string]bool
}

func (e *Exits) Dump(w io.Writer, ip string, port int) {
	ap := AddressPort{ip, port}
	e.GetAllExits(ap, func(exit string) {
		w.Write([]byte(exit + "\n"))
	})
}

func (e *Exits) GetAllExits(ap AddressPort, fn func(ip string)) {
	for _, val := range e.List {
		if val.CanExit(ap) {
			fn(val.Address)
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

func (e *Exits) Load(source io.Reader) error {
	var exits PolicyList
	dec := json.NewDecoder(source)
	for {
		var p Policy
		if err := dec.Decode(&p); err == io.EOF {
			break
		} else if err != nil {
			return err
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
		exits = append(exits, p)
	}

	// sort -n
	sort.Sort(exits)

	// swap in exits
	e.List = exits
	e.UpdateTime = time.Now()

	// precompute IsTor
	e.PreComputeTorList()

	return nil
}

func (e *Exits) LoadFromFile() {
	file, err := os.Open(os.ExpandEnv("${TORCHECKBASE}data/exit-policies"))
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	if err = e.Load(file); err != nil {
		log.Fatal(err)
	}
}

func (e *Exits) Run() {
	e.ReloadChan = make(chan os.Signal, 1)
	signal.Notify(e.ReloadChan, syscall.SIGUSR2)
	go func() {
		for {
			<-e.ReloadChan
			e.LoadFromFile()
			log.Println("Exit list reloaded.")
		}
	}()
	e.LoadFromFile()
}

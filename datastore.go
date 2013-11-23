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
	Fingerprint      string
	Address          string
	Rules            []Rule
	IsAllowedDefault bool
	Tminus           int
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

type ExitInfo struct {
	Address     string
	Fingerprint string
}

func (e ExitInfo) toJSON() (b []byte) {
	j, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return
	}
	return j
}

type Exits struct {
	List        PolicyList
	UpdateTime  time.Time
	ReloadChan  chan os.Signal
	IsTorLookup map[string]string
}

func (e *Exits) Dump(w io.Writer, tminus int, ip string, port int) {
	ap := AddressPort{ip, port}
	e.GetAllExits(ap, tminus, func(exit string, _ string, _ int) {
		w.Write([]byte(exit + "\n"))
	})
}

func (e *Exits) DumpJSON(w io.Writer, tminus int, ip string, port int) {
	ap := AddressPort{ip, port}
	Prefix := []byte(",\n")
	w.Write([]byte("["))
	e.GetAllExits(ap, tminus, func(address string, fingerprint string, ind int) {
		if ind > 0 {
			w.Write(Prefix)
		}
		w.Write(ExitInfo{address, fingerprint}.toJSON())
	})
	w.Write([]byte("]"))
}

func (e *Exits) GetAllExits(ap AddressPort, tminus int, fn func(string, string, int)) {
	ind := 0
	for _, val := range e.List {
		if val.Tminus <= tminus && val.CanExit(ap) {
			fn(val.Address, val.Fingerprint, ind)
			ind += 1
		}
	}
}

var DefaultTarget = AddressPort{"38.229.72.22", 443}

func (e *Exits) PreComputeTorList() {
	newmap := make(map[string]string)
	e.GetAllExits(DefaultTarget, 16, func(ip string, fingerprint string, _ int) {
		newmap[ip] = fingerprint
	})
	e.IsTorLookup = newmap
}

func (e *Exits) IsTor(remoteAddr string) (fingerprint string, ok bool) {
	fingerprint, ok = e.IsTorLookup[remoteAddr]
	return
}

func (e *Exits) Update(exits PolicyList) PolicyList {
	m := make(map[string]Policy)

	for _, p := range e.List {
		p.Tminus = p.Tminus + 1
		m[p.Fingerprint] = p
	}

	for _, p := range exits {
		m[p.Fingerprint] = p
	}

	i := 0
	exits = make(PolicyList, len(m))
	for _, p := range m {
		exits[i] = p
		i = i + 1
	}
	return exits
}

func (e *Exits) Load(source io.Reader, update bool) error {
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

	// bump entries by an hour that aren't in this list
	if update {
		exits = e.Update(exits)
	}

	// sort -n
	sort.Sort(exits)

	e.List = exits
	e.UpdateTime = time.Now()
	e.PreComputeTorList()

	return nil
}

func (e *Exits) LoadFromFile(filePath string, update bool) {
	file, err := os.Open(os.ExpandEnv(filePath))
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	if err = e.Load(file, update); err != nil {
		log.Fatal(err)
	}
}

func (e *Exits) Run(filePath string) {
	e.ReloadChan = make(chan os.Signal, 1)
	signal.Notify(e.ReloadChan, syscall.SIGUSR2)
	go func() {
		for {
			<-e.ReloadChan
			e.LoadFromFile(filePath, true)
			log.Println("Exit list updated.")
		}
	}()
	e.LoadFromFile(filePath, false)
}

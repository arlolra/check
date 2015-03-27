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

type CanExitCache struct {
	ap  AddressPort
	can bool
}

type Policy struct {
	Fingerprint      string
	Address          []string
	Rules            []Rule
	IsAllowedDefault bool
	Tminus           int
	CacheLast        CanExitCache
}

func (p Policy) CanExit(ap AddressPort) (can bool) {
	if p.CacheLast.ap == ap {
		can = p.CacheLast.can
		return
	}

	// update the cache *after* we return
	defer func() {
		p.CacheLast = CanExitCache{ap, can}
	}()

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

type PolicyAddress struct {
	Policy  Policy
	Address string
}

type PolicyList []PolicyAddress

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
	var last string
	e.GetAllExits(ap, tminus, func(exit string, _ string, _ int) {
		if exit != last {
			w.Write([]byte(exit + "\n"))
			last = exit
		}
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
		if val.Policy.Tminus <= tminus && val.Policy.CanExit(ap) {
			fn(val.Address, val.Policy.Fingerprint, ind)
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

func InsertUnique(arr *[]string, a string) {
	for _, b := range *arr {
		if a == b {
			return
		}
	}
	*arr = append(*arr, a)
}

func (e *Exits) Update(exits []Policy, update bool) {
	m := make(map[string]Policy)

	// bump entries by an hour that aren't in the new exit list
	if update {
		for _, p := range e.List {
			if _, ok := m[p.Policy.Fingerprint]; !ok {
				p.Policy.Tminus = p.Policy.Tminus + 1
				m[p.Policy.Fingerprint] = p.Policy
			}
		}
	}

	// keep all unique ips we've seen
	for _, p := range exits {
		if q, ok := m[p.Fingerprint]; ok {
			for _, a := range q.Address {
				InsertUnique(&p.Address, a)
			}
		}
		m[p.Fingerprint] = p
	}

	var pl PolicyList
	for _, p := range m {
		for _, a := range p.Address {
			pl = append(pl, PolicyAddress{p, a})
		}
	}

	// sort -n
	sort.Sort(pl)

	e.List = pl
}

func (e *Exits) Load(source io.Reader, update bool) error {
	var exits []Policy
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

	e.Update(exits, update)
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

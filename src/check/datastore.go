package check

import (
	"encoding/json"
	"github.com/Ryman/intstab"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"
)

func ValidPort(port int) bool {
	return port >= 0 && port < 65536
}

type AddressPort struct {
	Address string
	Port    int
}

type Exits struct {
	List       intstab.IntervalStabber
	UpdateTime time.Time
	ReloadChan chan os.Signal
	TorIPs     map[string]bool
}

func (e *Exits) IsAllowed(address net.IP, port int, cb func([]byte)) {
	rules, err := e.List.Intersect(uint16(port))
	if err != nil {
		return
	}

	sort.Sort(OrderedRuleIntervalSlice(rules))

	// sorted rules are ordered by Policy.Id (ascending)
	lastPolicyResult := -1
	for _, i := range rules {
		r := i.Tag.(*Rule)
		if lastPolicyResult >= r.ParentPolicy.Id {
			continue
		}
		if r.IsMatch(address) {
			lastPolicyResult = r.ParentPolicy.Id
			if r.IsAccept {
				cb(r.ParentPolicy.AddressNewLine)
			}
		}
	}
}

func (e *Exits) Dump(w io.Writer, ip string, port int) {
	address := net.ParseIP(ip)
	if address == nil || !ValidPort(port) {
		return
	}
	e.IsAllowed(address, port, func(ip []byte) {
		w.Write(ip)
	})
}

var DefaultTarget = AddressPort{"38.229.70.31", 443}

func (e *Exits) PreComputeTorList() {
	newmap := make(map[string]bool, len(e.TorIPs))
	addr := net.ParseIP(DefaultTarget.Address)
	e.IsAllowed(addr, DefaultTarget.Port, func(ip []byte) {
		newmap[string(ip[:len(ip)-1])] = true
	})
	e.TorIPs = newmap
}

func (e *Exits) IsTor(remoteAddr string) bool {
	return e.TorIPs[remoteAddr]
}

func (e *Exits) Load(source io.Reader) error {
	// maybe more intervals?
	intervals := make(intstab.IntervalSlice, 0, 30000)

	dec := json.NewDecoder(source)
	for i := 0; true; i++ {
		var p Policy
		if err := dec.Decode(&p); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		p.AddressNewLine = []byte(p.Address + "\n")
		p.Id = i
		for r := range p.IterateProcessedRules() {
			tag := &intstab.Interval{uint16(r.MinPort), uint16(r.MaxPort), r}
			intervals = append(intervals, tag)
		}
	}

	// swap in exits if no errors
	if list, err := intstab.NewIntervalStabber(intervals); err != nil {
		log.Print("Failed to create new IntervalStabber: ", err)
		return err
	} else {
		e.List = list
		e.PreComputeTorList()
	}

	e.UpdateTime = time.Now()
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

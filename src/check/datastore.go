package check

import (
	"bytes"
	"encoding/json"
	_ "fmt"
	"github.com/Ryman/intstab"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// stem's ExitPolicyRule class, sort of
type Rule struct {
	IsAccept      bool
	Address       string
	AddressIP     net.IP
	MinPort       int
	MaxPort       int
	PolicyAddress string
	// Optimisation:
	// Reduced allocs from appending \n to each line in dumps
	PolicyAddressNewLine []byte
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

func (r *Rule) IsAllowed(ip net.IP) bool {
	return r.IsAccept && (r.AddressIP == nil || r.AddressIP.Equal(ip))
}

// func (p Policy) CanExit(ap AddressPort) bool {
// 	can, ok := p.CanExitCache[ap]
// 	if !ok {
// 		can = p.IsAllowedDefault
// 		for _, rule := range p.Rules {
// 			if rule.IsMatch(ap) {
// 				can = rule.IsAccept
// 				break
// 			}
// 		}
// 		p.CanExitCache[ap] = can
// 	}
// 	return can
// }

type Exits struct {
	List       intstab.IntervalStabber
	UpdateTime time.Time
	ReloadChan chan os.Signal
	torIPs     map[string]bool
}

func (e *Exits) Dump(w io.Writer, ip string, port int) {
	address := net.ParseIP(ip)
	if address == nil || !ValidPort(port) {
		return // TODO: Return error
	}

	rules, err := e.List.Intersect(uint16(port))
	if err != nil {
		return // TODO: Return error
	}

	// TODO: exclude ips that were already included
	for _, i := range rules {
		// TODO: Remove this type assertion? Seems to be triggering memmoves
		if r := i.Tag.(Rule); r.IsAllowed(address) {
			w.Write(r.PolicyAddressNewLine)
		}
	}
}

var DefaultTarget = AddressPort{"38.229.70.31", 443}

// This is pretty wastefully implemented, but it's run once per hour
// so it's not a big deal unless it leaks its memory. Check for that!
func (e *Exits) PreComputeTorList() {
	newmap := make(map[string]bool, len(e.torIPs))
	buf := new(bytes.Buffer)
	e.Dump(buf, DefaultTarget.Address, DefaultTarget.Port)
	strIPs := strings.Split(buf.String(), "\n")

	for _, s := range strIPs {
		s = strings.Trim(s, " ")
		newmap[s] = (s != "")
	}

	e.torIPs = newmap
}

func (e *Exits) IsTor(remoteAddr string) bool {
	return e.torIPs[remoteAddr]
}

func (e *Exits) Load(source io.Reader) error {
	e.torIPs = make(map[string]bool)
	intervals := make(intstab.IntervalSlice, 0, 30000)

	dec := json.NewDecoder(source)
	for {
		var p Policy
		if err := dec.Decode(&p); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		// TODO: Turn !IsAccept rules into partitions if
		// the policy's acceptdefault is true
		// BUG: if acceptdefault is true, but they have
		// blocked port 80, then we need to have a range
		// 1-79 and 81-65535
		// Add tests for this
		for _, r := range p.Rules {
			r.AddressIP = net.ParseIP(r.Address)
			r.PolicyAddress = p.Address
			r.PolicyAddressNewLine = []byte(p.Address + "\n")
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

// Helper to ease testing for Load
func (e *Exits) loadFromFile() {
	file, err := os.Open(os.ExpandEnv("${TORCHECKBASE}data/exit-policies"))
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	e.Load(file)
}

func (e *Exits) Run() {
	e.ReloadChan = make(chan os.Signal, 1)
	signal.Notify(e.ReloadChan, syscall.SIGUSR2)
	go func() {
		for {
			<-e.ReloadChan
			e.loadFromFile()
			log.Println("Exit list reloaded.")
		}
	}()
	e.loadFromFile()
}

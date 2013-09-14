package check

import (
	"bytes"
	"encoding/json"
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
		if r := i.Tag.(*Rule); r.IsAllowed(address) {
			w.Write(r.PolicyAddressNewLine)
		}
	}
}

var DefaultTarget = AddressPort{"38.229.70.31", 443}

// This is pretty wastefully implemented, but it's run once per hour
// so it's not a big deal unless it leaks its memory. Check for that!
func (e *Exits) PreComputeTorList() {
	newmap := make(map[string]bool, len(e.TorIPs))
	buf := new(bytes.Buffer)
	e.Dump(buf, DefaultTarget.Address, DefaultTarget.Port)
	strIPs := strings.Split(buf.String(), "\n")

	for _, s := range strIPs {
		s = strings.Trim(s, " ")
		newmap[s] = (s != "")
	}

	e.TorIPs = newmap
}

func (e *Exits) IsTor(remoteAddr string) bool {
	return e.TorIPs[remoteAddr]
}

func (e *Exits) Load(source io.Reader) error {
	// maybe more intervals?
	intervals := make(intstab.IntervalSlice, 0, 30000)

	dec := json.NewDecoder(source)
	for {
		var p Policy
		if err := dec.Decode(&p); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
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

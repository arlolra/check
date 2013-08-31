package check

import (
	"encoding/json"
	"fmt"
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

func (r Rule) IsMatch(ap AddressPort) bool {
	address := net.ParseIP(ap.Address)
	if address == nil {
		return false
	}
	if r.AddressIP != nil && !r.AddressIP.Equal(address) {
		return false
	}
	if !ValidPort(ap.Port) || ap.Port < r.MinPort || ap.Port > r.MaxPort {
		return false
	}
	return true
}

type AddressPort struct {
	Address string
	Port    int
}

type Policy struct {
	Address      string
	Rules        []Rule
	CanExitCache map[AddressPort]bool
}

func (p Policy) CanExit(ap AddressPort) bool {
	can, ok := p.CanExitCache[ap]
	if !ok {
		can = false
		for _, rule := range p.Rules {
			if rule.IsMatch(ap) {
				can = rule.IsAccept
				break
			}
		}
		p.CanExitCache[ap] = can
	}
	return can
}

type Exits struct {
	List       map[string]Policy
	UpdateTime time.Time
	ReloadChan chan os.Signal
}

func (e *Exits) Dump(ip string, port int) string {
	ap := AddressPort{ip, port}
	var str string
	for key, val := range e.List {
		if val.CanExit(ap) {
			str += fmt.Sprintf("%s\n", key)
		}
	}
	return str
}

var DefaultTarget = AddressPort{"38.229.70.31", 443}

func (e *Exits) IsTor(remoteAddr string) bool {
	policy, ok := e.List[remoteAddr]
	if ok {
		return policy.CanExit(DefaultTarget)
	}
	return false
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

package check

import (
	"log"
	"net"
)

type Policy struct {
	Address          string
	Rules            []Rule
	IsAllowedDefault bool
}

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

func (r *Rule) IsAllowed(ip net.IP) bool {
	// for accept rules, check if we've got match (or wildcard)
	// for reject, don't
	return r.IsAccept == (r.AddressIP == nil || r.AddressIP.Equal(ip))
}

// precompute some data for the rule based on parent policy
func (r Rule) Finalize(p *Policy) *Rule {
	r.AddressIP = net.ParseIP(r.Address)
	r.PolicyAddress = p.Address
	r.PolicyAddressNewLine = []byte(p.Address + "\n")

	return &r
}

// Transform any reject rules into a set of accept rules
// e.g. {x, y} => {..., x-1}, {y+1, ...}
func (p *Policy) GetRulesDefaultAccept(ch chan *Rule) {
	// TODO: Test exit relays to see what happens
	// if you specify accept 80, reject 80 in both orders
	// NOTE: This includes 0 as a valid port
	openPorts := new([65536]bool)
	for i, _ := range openPorts {
		openPorts[i] = true
	}

	// block all rejected ports even wildcarded
	// the reject rules must be checked at query time
	for _, r := range p.Rules {
		if !r.IsAccept {
			// block every port in its range
			for i := r.MinPort; i <= r.MaxPort; i++ {
				if !openPorts[i] {
					log.Print("Found overlapping reject rule in Policy: %v", p)
					log.Print("Port %v is rejected twice for all IPs", i)
				}
				// 'close' the port
				openPorts[i] = false
			}
		}
	}

	// create accept ranges for ports open to wildcard addresses
	var currentRule *Rule
	for i, open := range openPorts {
		if !open && currentRule != nil {
			// finish a range
			currentRule.MaxPort = i - 1
			ch <- currentRule.Finalize(p)
			currentRule = nil
		} else if i == 65535 && currentRule != nil {
			// Close the final range if it's not blocked
			currentRule.MaxPort = i
			ch <- currentRule.Finalize(p)
			currentRule = nil
		} else if open && currentRule == nil {
			// start a new open range
			currentRule = new(Rule)
			currentRule.MinPort = i
			currentRule.IsAccept = true
		} // else we're in the middle of a range, do nothing
	}

	// add all rules that are ip specific
	// ip specific rules need to check at query time
	for _, r := range p.Rules {
		if r.Address != "" {
			ch <- r.Finalize(p)
		}
	}
}

func (p *Policy) GetRulesDefaultReject(ch chan *Rule) {
	// default is reject, ignore any reject rules (redundant)
	for _, r := range p.Rules {
		if r.IsAccept {
			ch <- r.Finalize(p)
		}
	}
}

func (p *Policy) IterateProcessedRules() <-chan *Rule {
	ch := make(chan *Rule, 1000)

	// This is an iterator (<-chan *Rule)
	go func() {
		if p.IsAllowedDefault {
			p.GetRulesDefaultAccept(ch)
		} else {
			p.GetRulesDefaultReject(ch)
		}
		// close or deadlock!
		close(ch)
	}()

	return ch
}

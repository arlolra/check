package check

import "net"

type Policy struct {
	Address          string
	Rules            []Rule
	IsAllowedDefault bool
}

type Rule struct {
	IsAccept             bool
	IsAddressWildcard    bool
	Address              string
	Mask                 string
	MinPort              int
	MaxPort              int
	IP                   net.IP
	IPNet                *net.IPNet
	PolicyAddress        string
	PolicyAddressNewLine []byte
	Order                int
}

func (r *Rule) IsMatch(ip net.IP) bool {
	if !r.IsAddressWildcard {
		if r.IPNet != nil && !r.IPNet.Contains(ip) {
			return false
		} else {
			return r.IP.Equal(ip)
		}
	}
	return true
}

// precompute some data for the rule based on parent policy
func (r Rule) Finalize(p *Policy) *Rule {
	if !r.IsAddressWildcard {
		r.IP = net.ParseIP(r.Address)
		if mask := net.ParseIP(r.Mask); r.IP != nil && mask != nil {
			m := make(net.IPMask, len(mask))
			copy(m, mask)
			r.IPNet = &net.IPNet{r.IP.Mask(m), m}
		}
	}
	r.PolicyAddress = p.Address
	r.PolicyAddressNewLine = []byte(p.Address + "\n")
	return &r
}

func (p *Policy) IterateProcessedRules() <-chan *Rule {
	ch := make(chan *Rule, 1000)

	go func() {
		var i int
		for i = range p.Rules {
			r := &p.Rules[i]
			r.Order = i
			ch <- r.Finalize(p)
		}
		if p.IsAllowedDefault {
			nr := new(Rule)
			nr.MinPort = 0
			nr.MaxPort = 65535
			nr.IsAccept = true
			nr.IsAddressWildcard = true
			nr.Order = i + 1
			ch <- nr.Finalize(p)
		}
		close(ch)
	}()

	return ch
}

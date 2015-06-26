package main

import "testing"

var UserAgents = map[string]bool{
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:10.0.2) Gecko/20100101 Firefox/10.0.2":                                    false,
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:11.0) Gecko/20100101 Firefox/11.0":                                        false,
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/27.0.1453.110 Safari/537.36": false,
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_4) AppleWebKit/536.30.1 (KHTML, like Gecko) Version/6.0.5 Safari/536.30.1":    false,
	"Mozilla/5.0 (Windows NT 6.1; rv:10.0) Gecko/20100101 Firefox/10.0":                                                        true,
	"Mozilla/5.0 (Windows NT 6.1; rv:17.0) Gecko/20100101 Firefox/17.0":                                                        true,
	"Mozilla/5.0 (Windows NT 6.1; rv:24.0) Gecko/20100101 Firefox/24.0":                                                        true,
}

func TestLikelyTBB(t *testing.T) {
	for k, v := range UserAgents {
		if LikelyTBB(k) != v {
			t.Errorf("Expected \"%s\" to be: %t", k, v)
		}
	}
}

package main

import (
	"bytes"
	"strings"
	"testing"
)

func dependOn(t *testing.T, deps ...func(*testing.T)) {
	// stops things getting logged into other test logs
	temp := new(testing.T)
	for _, dependancy := range deps {
		dependancy(temp)

		if temp.Skipped() || temp.Failed() {
			t.Skip("Dependency failed, Skipping.")
		}
	}
}

func checkDump(t *testing.T, dumpStr string, expected ...string) {
	found := make(map[string]bool, len(expected))

	if i := len(dumpStr); i > 0 && dumpStr[i-1:i] == "\n" {
		t.Log("Removing trailing newline from dump string")
		dumpStr = dumpStr[0 : i-1]
	}

	if len(dumpStr) > 0 {
		for _, s := range strings.Split(dumpStr, "\n") {
			found[s] = true
		}
	}

	if l, x := len(found), len(expected); l != x {
		t.Errorf("Found %d ips, expected %d", l, x)
	}

	for _, x := range expected {
		if !found[x] {
			t.Errorf("Missing [%s] from dump", x)
		}
	}

	if !t.Failed() {
		t.Log("All expected values found\n")
	} else {
		t.Log("Dump was:\n-----DUMP-----\n", dumpStr, "\n---END DUMP---\n")
	}
}

func (e *Exits) assertIsTor(t *testing.T, ip string, expected bool) {
	if _, ok := e.IsTor(ip); ok != expected {
		t.Errorf("Failed IsTor Assert for %s, got %v but wanted %v", ip, ok, expected)
	}
}

func setupExitList(t *testing.T, testData string) (e *Exits) {
	e = new(Exits)
	err := e.Load(strings.NewReader(testData), false)
	if err != nil {
		t.Fatal("Failed to load data")
	}
	return
}

func TestExitListLoading(t *testing.T) {
	testData := `{"Rules": [{"IsAccept": true, "MinPort": 706, "MaxPort": 706, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 993, "MaxPort": 993, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 995, "MaxPort": 995, "Address": null, "IsAddressWildcard": true}], "IsAllowedDefault": false, "Address": ["83.227.52.198"], "Fingerprint": "1"}
				 {"Rules": [{"IsAccept": true, "MinPort": 20, "MaxPort": 23, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 43, "MaxPort": 43, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 53, "MaxPort": 53, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 79, "MaxPort": 81, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 88, "MaxPort": 88, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 110, "MaxPort": 110, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 143, "MaxPort": 143, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 194, "MaxPort": 194, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 220, "MaxPort": 220, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 443, "MaxPort": 443, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 464, "MaxPort": 465, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 543, "MaxPort": 544, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 563, "MaxPort": 563, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 587, "MaxPort": 587, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 706, "MaxPort": 706, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 749, "MaxPort": 749, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 873, "MaxPort": 873, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 902, "MaxPort": 904, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 981, "MaxPort": 981, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 989, "MaxPort": 995, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 1194, "MaxPort": 1194, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 1220, "MaxPort": 1220, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 1293, "MaxPort": 1293, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 1500, "MaxPort": 1500, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 1723, "MaxPort": 1723, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 1863, "MaxPort": 1863, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 2082, "MaxPort": 2083, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 2086, "MaxPort": 2087, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 2095, "MaxPort": 2096, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 3128, "MaxPort": 3128, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 3389, "MaxPort": 3389, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 3690, "MaxPort": 3690, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 4321, "MaxPort": 4321, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 4643, "MaxPort": 4643, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 5050, "MaxPort": 5050, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 5190, "MaxPort": 5190, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 5222, "MaxPort": 5223, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 5228, "MaxPort": 5228, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 5900, "MaxPort": 5900, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 6666, "MaxPort": 6667, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 6679, "MaxPort": 6679, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 6697, "MaxPort": 6697, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 8000, "MaxPort": 8000, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 8008, "MaxPort": 8008, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 8080, "MaxPort": 8080, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 8087, "MaxPort": 8088, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 8443, "MaxPort": 8443, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 8888, "MaxPort": 8888, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 9418, "MaxPort": 9418, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 9999, "MaxPort": 10000, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 19294, "MaxPort": 19294, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 19638, "MaxPort": 19638, "Address": null, "IsAddressWildcard": true}], "IsAllowedDefault": false, "Address": ["91.121.43.80"], "Fingerprint": "2"}`
	exits := setupExitList(t, testData)

	// Valid tor exit
	exits.assertIsTor(t, "91.121.43.80", true)

	// Invalid tor exit
	exits.assertIsTor(t, "91.121.43.4", false)

	// check both exits are listed for 995
	// Accept either ordering of output
	expectDump(t, exits, "38.229.70.31", 995, "91.121.43.80", "83.227.52.198")
}

func expectDump(t *testing.T, e *Exits, ip string, port int, expected ...string) {
	buf := new(bytes.Buffer)
	e.Dump(buf, 16, ip, port)
	checkDump(t, buf.String(), expected...)
}

func TestIsAcceptRules(t *testing.T) {
	testData := `{"Rules": [{"IsAccept": false, "MinPort": 706, "MaxPort": 706, "Address": null, "IsAddressWildcard": true}, {"IsAccept": false, "MinPort": 5000, "MaxPort": 55000, "Address": null, "IsAddressWildcard": true}], "IsAllowedDefault": false, "Address": ["111.111.111.111"], "Fingerprint": "1"}
				 {"Rules": [{"IsAccept": true, "MinPort": 706, "MaxPort": 706, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 5000, "MaxPort": 55000, "Address": null, "IsAddressWildcard": true}], "IsAllowedDefault": false, "Address": ["222.222.222.222"], "Fingerprint": "2"}`
	exits := setupExitList(t, testData)

	// one should fail, the other should be OK
	expectDump(t, exits, "38.229.70.31", 706, "222.222.222.222")
	// ensure ranges work
	expectDump(t, exits, "38.229.70.31", 50000, "222.222.222.222")
}

func TestIsDefaultAllowedPolicy(t *testing.T) {
	testData := `{"Rules": [{"IsAccept": false, "MinPort": 706, "MaxPort": 706, "Address": null, "IsAddressWildcard": true}, {"IsAccept": false, "MinPort": 5000, "MaxPort": 55000, "Address": null, "IsAddressWildcard": true}], "IsAllowedDefault": true, "Address": ["111.111.111.111"], "Fingerprint": "1"}
				 {"Rules": [{"IsAccept": true, "MinPort": 706, "MaxPort": 706, "Address": null, "IsAddressWildcard": true}, {"IsAccept": true, "MinPort": 5000, "MaxPort": 55000, "Address": null, "IsAddressWildcard": true}], "IsAllowedDefault": false, "Address": ["222.222.222.222"], "Fingerprint": "2"}`
	exits := setupExitList(t, testData)

	// first one should be allowing everything but his blocked port,
	// second one should allow only on port 706
	expectDump(t, exits, "38.229.70.31", 200, "111.111.111.111")
	expectDump(t, exits, "38.229.70.31", 706, "222.222.222.222")
	// ensure ranges work (expect only second ip)
	expectDump(t, exits, "38.229.70.31", 50000, "222.222.222.222")
}

func TestRulesNonWildcard(t *testing.T) {
	// Testing load
	testData := `{"Rules": [{"IsAccept": false, "MinPort": 706, "MaxPort": 706, "Address": "38.229.70.31"}, {"IsAccept": false, "MinPort": 5000, "MaxPort": 55000, "Address": "38.229.70.31"}], "IsAllowedDefault": true, "Address": ["111.111.111.111"], "Fingerprint": "1"}
				 {"Rules": [{"IsAccept": true, "MinPort": 706, "MaxPort": 706, "Address": "38.229.70.31"}, {"IsAccept": true, "MinPort": 5000, "MaxPort": 55000, "Address": "38.229.70.31"}], "IsAllowedDefault": false, "Address": ["222.222.222.222"], "Fingerprint": "2"}`
	exits := setupExitList(t, testData)

	// first one should reject due to ip
	// second one should accept due to ip
	expectDump(t, exits, "38.229.70.31", 706, "222.222.222.222")
	// ensure ranges work
	expectDump(t, exits, "38.229.70.31", 53000, "222.222.222.222")

	// check an ip that doesn't match the rules
	// first should accept due to default policy
	// second should reject due to default policy
	dependOn(t, TestIsDefaultAllowedPolicy)
	expectDump(t, exits, "32.32.32.32", 706, "111.111.111.111")
	// ensure ranges work
	expectDump(t, exits, "32.32.32.32", 53000, "111.111.111.111")
}

func TestMaskedIP(t *testing.T) {
	testData := `{"Rules": [{"MaxPort": 65535, "IsAddressWildcard": false, "Mask": "255.0.0.0", "Address": "0.0.0.0", "IsAccept": false, "MinPort": 1}, {"MaxPort": 65535, "IsAddressWildcard": false, "Mask": "255.255.0.0", "Address": "169.254.0.0", "IsAccept": false, "MinPort": 1}], "IsAllowedDefault": true, "Address": ["111.111.111.111"], "Fingerprint": "1"}`
	exits := setupExitList(t, testData)
	expectDump(t, exits, "0.1.2.3", 123)
	expectDump(t, exits, "169.254.111.111", 345)
	expectDump(t, exits, "1.1.2.3", 123, "111.111.111.111")
}

func TestDoubleReject(t *testing.T) {
	testData := `{"Rules": [{"IsAccept": false, "MinPort": 80, "MaxPort": 80, "Address": "222.222.222.222"}, {"IsAccept": false, "MinPort": 80, "MaxPort": 80, "Address": "123.123.123.123"}], "IsAllowedDefault": true, "Address": ["111.111.111.111"], "Fingerprint": "1"}`
	exits := setupExitList(t, testData)
	expectDump(t, exits, "222.222.222.222", 80)
}

func TestRejectWithDefaultReject(t *testing.T) {
	testData := `{"Rules": [{"IsAccept": false, "MinPort": 80, "MaxPort": 80, "Address": "222.222.222.222"}, {"IsAccept": true, "MinPort": 80, "MaxPort": 80, "Address": "", "IsAddressWildcard": true}], "IsAllowedDefault": false, "Address": ["111.111.111.111"], "Fingerprint": "1"}`
	exits := setupExitList(t, testData)
	// Should reject
	expectDump(t, exits, "222.222.222.222", 80)
	// Should accept on port 80 only
	expectDump(t, exits, "222.222.222.111", 80, "111.111.111.111")
	expectDump(t, exits, "222.222.222.111", 81)
}

func TestMatchedRuleOrdering(t *testing.T) {
	/* SPEC: These lines describe an "exit policy": the rules that an OR follows
	   when deciding whether to allow a new stream to a given address. The
	   'exitpattern' syntax is described below. There MUST be at least one
	   such entry. The rules are considered in order; if no rule matches,
	   the address will be accepted. For clarity, the last such entry SHOULD
	   be accept : or reject :. */
	testData := `{"Rules": [{"IsAccept": false, "MinPort": 80, "MaxPort": 80, "Address": "222.222.222.222"}, {"IsAccept": true, "MinPort": 80, "MaxPort": 80, "Address": "222.222.222.222"}], "IsAllowedDefault": false, "Address": ["111.111.111.111"], "Fingerprint": "1"}`
	exits := setupExitList(t, testData)
	// Should match the reject rule first
	expectDump(t, exits, "222.222.222.222", 80)

	testData = `{"Rules": [{"IsAccept": true, "MinPort": 80, "MaxPort": 80, "Address": "222.222.222.222"}, {"IsAccept": false, "MinPort": 80, "MaxPort": 80, "Address": "222.222.222.222"}], "IsAllowedDefault": false, "Address": ["111.111.111.111"], "Fingerprint": "1"}`
	exits = setupExitList(t, testData)
	// Should match the accept rule first
	expectDump(t, exits, "222.222.222.222", 80, "111.111.111.111")
}

func TestPastHours(t *testing.T) {
	testData := `{"Rules": [{"IsAccept": true, "MinPort": 80, "MaxPort": 80, "Address": null, "IsAddressWildcard": true}], "IsAllowedDefault": false, "Address": ["111.111.111.111"], "Fingerprint": "1", "Tminus": 4}
	{"Rules": [{"IsAccept": true, "MinPort": 80, "MaxPort": 80, "Address": null, "IsAddressWildcard": true}], "IsAllowedDefault": false, "Address": ["222.222.222.222"], "Fingerprint": "2", "Tminus": 17}`
	exits := setupExitList(t, testData)
	// Should reject
	expectDump(t, exits, "123.123.123.123", 80, "111.111.111.111")
}

func BenchmarkIsTor(b *testing.B) {
	e := new(Exits)
	e.LoadFromFile("data/exit-policies", false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.IsTor("91.121.43.80")
		e.IsTor("91.121.43.4")
	}
}

func BenchmarkDumpList(b *testing.B) {
	e := new(Exits)
	e.LoadFromFile("data/exit-policies", false)
	buf := new(bytes.Buffer)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Dump(buf, 16, DefaultTarget.Address, DefaultTarget.Port)
		buf.Reset()
	}
}

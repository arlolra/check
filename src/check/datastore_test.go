package check

import (
	"bytes"
	"strings"
	"testing"
)

func dependOn(t *testing.T, deps ...func(*testing.T)) {
	// Stops things getting logged into other test logs (less confusion)
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

	for _, s := range strings.Split(dumpStr, "\n") {
		found[s] = true
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

// func (p *Policy) checkPolicyCanExit(t *testing.T, ip string, port int, expected bool) {
// 	if ok := p.CanExit(AddressPort{ip, port}); ok != expected {
// 		t.Errorf("CanExit Error: Got %v, Expected %v", ok, expected)
// 	}
// }

func (e *Exits) assertIsTor(t *testing.T, ip string, expected bool) {
	if ok := e.IsTor(ip); ok != expected {
		t.Errorf("Failed IsTor Assert for %s, got %v but wanted %v", ip, ok, expected)
	}
}

func setupExitList(t *testing.T, testData string) (e *Exits) {
	e = new(Exits)

	err := e.Load(strings.NewReader(testData))
	if err != nil {
		t.Fatal("Failed to load data")
	}

	return
}

func TestExitListLoading(t *testing.T) {
	// Testing load
	testData := `{"Rules": [{"IsAccept": true, "MinPort": 706, "MaxPort": 706, "Address": null}, {"IsAccept": true, "MinPort": 993, "MaxPort": 993, "Address": null}, {"IsAccept": true, "MinPort": 995, "MaxPort": 995, "Address": null}], "IsAllowedDefault": false, "Address": "83.227.52.198"}
				 {"Rules": [{"IsAccept": true, "MinPort": 20, "MaxPort": 23, "Address": null}, {"IsAccept": true, "MinPort": 43, "MaxPort": 43, "Address": null}, {"IsAccept": true, "MinPort": 53, "MaxPort": 53, "Address": null}, {"IsAccept": true, "MinPort": 79, "MaxPort": 81, "Address": null}, {"IsAccept": true, "MinPort": 88, "MaxPort": 88, "Address": null}, {"IsAccept": true, "MinPort": 110, "MaxPort": 110, "Address": null}, {"IsAccept": true, "MinPort": 143, "MaxPort": 143, "Address": null}, {"IsAccept": true, "MinPort": 194, "MaxPort": 194, "Address": null}, {"IsAccept": true, "MinPort": 220, "MaxPort": 220, "Address": null}, {"IsAccept": true, "MinPort": 443, "MaxPort": 443, "Address": null}, {"IsAccept": true, "MinPort": 464, "MaxPort": 465, "Address": null}, {"IsAccept": true, "MinPort": 543, "MaxPort": 544, "Address": null}, {"IsAccept": true, "MinPort": 563, "MaxPort": 563, "Address": null}, {"IsAccept": true, "MinPort": 587, "MaxPort": 587, "Address": null}, {"IsAccept": true, "MinPort": 706, "MaxPort": 706, "Address": null}, {"IsAccept": true, "MinPort": 749, "MaxPort": 749, "Address": null}, {"IsAccept": true, "MinPort": 873, "MaxPort": 873, "Address": null}, {"IsAccept": true, "MinPort": 902, "MaxPort": 904, "Address": null}, {"IsAccept": true, "MinPort": 981, "MaxPort": 981, "Address": null}, {"IsAccept": true, "MinPort": 989, "MaxPort": 995, "Address": null}, {"IsAccept": true, "MinPort": 1194, "MaxPort": 1194, "Address": null}, {"IsAccept": true, "MinPort": 1220, "MaxPort": 1220, "Address": null}, {"IsAccept": true, "MinPort": 1293, "MaxPort": 1293, "Address": null}, {"IsAccept": true, "MinPort": 1500, "MaxPort": 1500, "Address": null}, {"IsAccept": true, "MinPort": 1723, "MaxPort": 1723, "Address": null}, {"IsAccept": true, "MinPort": 1863, "MaxPort": 1863, "Address": null}, {"IsAccept": true, "MinPort": 2082, "MaxPort": 2083, "Address": null}, {"IsAccept": true, "MinPort": 2086, "MaxPort": 2087, "Address": null}, {"IsAccept": true, "MinPort": 2095, "MaxPort": 2096, "Address": null}, {"IsAccept": true, "MinPort": 3128, "MaxPort": 3128, "Address": null}, {"IsAccept": true, "MinPort": 3389, "MaxPort": 3389, "Address": null}, {"IsAccept": true, "MinPort": 3690, "MaxPort": 3690, "Address": null}, {"IsAccept": true, "MinPort": 4321, "MaxPort": 4321, "Address": null}, {"IsAccept": true, "MinPort": 4643, "MaxPort": 4643, "Address": null}, {"IsAccept": true, "MinPort": 5050, "MaxPort": 5050, "Address": null}, {"IsAccept": true, "MinPort": 5190, "MaxPort": 5190, "Address": null}, {"IsAccept": true, "MinPort": 5222, "MaxPort": 5223, "Address": null}, {"IsAccept": true, "MinPort": 5228, "MaxPort": 5228, "Address": null}, {"IsAccept": true, "MinPort": 5900, "MaxPort": 5900, "Address": null}, {"IsAccept": true, "MinPort": 6666, "MaxPort": 6667, "Address": null}, {"IsAccept": true, "MinPort": 6679, "MaxPort": 6679, "Address": null}, {"IsAccept": true, "MinPort": 6697, "MaxPort": 6697, "Address": null}, {"IsAccept": true, "MinPort": 8000, "MaxPort": 8000, "Address": null}, {"IsAccept": true, "MinPort": 8008, "MaxPort": 8008, "Address": null}, {"IsAccept": true, "MinPort": 8080, "MaxPort": 8080, "Address": null}, {"IsAccept": true, "MinPort": 8087, "MaxPort": 8088, "Address": null}, {"IsAccept": true, "MinPort": 8443, "MaxPort": 8443, "Address": null}, {"IsAccept": true, "MinPort": 8888, "MaxPort": 8888, "Address": null}, {"IsAccept": true, "MinPort": 9418, "MaxPort": 9418, "Address": null}, {"IsAccept": true, "MinPort": 9999, "MaxPort": 10000, "Address": null}, {"IsAccept": true, "MinPort": 19294, "MaxPort": 19294, "Address": null}, {"IsAccept": true, "MinPort": 19638, "MaxPort": 19638, "Address": null}], "IsAllowedDefault": false, "Address": "91.121.43.80"}`
	exits := setupExitList(t, testData) //, 2)

	// TODO: Update to relative implementation if still needed
	// p := exits.List["83.227.52.198"]
	// if addr := p.Address; addr != "83.227.52.198" {
	// 	t.Errorf("Unexpected address for policy, got %s but expected %s", addr, "83.227.52.198")
	// }

	//p.checkPolicyCanExit(t, "38.229.70.31", 443, false)
	//p.checkPolicyCanExit(t, "38.229.70.31", 995, true)

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
	e.Dump(buf, ip, port)
	checkDump(t, buf.String(), expected...)
}

func TestIsAcceptRules(t *testing.T) {
	// Testing load
	testData := `{"Rules": [{"IsAccept": false, "MinPort": 706, "MaxPort": 706, "Address": null}, {"IsAccept": false, "MinPort": 5000, "MaxPort": 55000, "Address": null}], "IsAllowedDefault": false, "Address": "111.111.111.111"}
				 {"Rules": [{"IsAccept": true, "MinPort": 706, "MaxPort": 706, "Address": null}, {"IsAccept": true, "MinPort": 5000, "MaxPort": 55000, "Address": null}], "IsAllowedDefault": false, "Address": "222.222.222.222"}`
	exits := setupExitList(t, testData)

	// one should fail, the other should be OK
	expectDump(t, exits, "38.229.70.31", 706, "222.222.222.222")
	// ensure ranges work
	expectDump(t, exits, "38.229.70.31", 50000, "222.222.222.222")
}

func TestIsDefaultAllowedPolicy(t *testing.T) {
	// Testing load
	testData := `{"Rules": [{"IsAccept": false, "MinPort": 706, "MaxPort": 706, "Address": null}, {"IsAccept": false, "MinPort": 5000, "MaxPort": 55000, "Address": null}], "IsAllowedDefault": true, "Address": "111.111.111.111"}
				 {"Rules": [{"IsAccept": true, "MinPort": 706, "MaxPort": 706, "Address": null}, {"IsAccept": true, "MinPort": 5000, "MaxPort": 55000, "Address": null}], "IsAllowedDefault": false, "Address": "222.222.222.222"}`
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
	testData := `{"Rules": [{"IsAccept": false, "MinPort": 706, "MaxPort": 706, "Address": "38.229.70.31"}, {"IsAccept": false, "MinPort": 5000, "MaxPort": 55000, "Address": "38.229.70.31"}], "IsAllowedDefault": true, "Address": "111.111.111.111"}
				 {"Rules": [{"IsAccept": true, "MinPort": 706, "MaxPort": 706, "Address": "38.229.70.31"}, {"IsAccept": true, "MinPort": 5000, "MaxPort": 55000, "Address": "38.229.70.31"}], "IsAllowedDefault": false, "Address": "222.222.222.222"}`
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
	// TODO
}

func BenchmarkIsTor(b *testing.B) {
	e := new(Exits)
	e.loadFromFile()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.IsTor("91.121.43.80")
		e.IsTor("91.121.43.4")
	}
}

func BenchmarkDumpList(b *testing.B) {
	e := new(Exits)
	e.loadFromFile()

	buf := new(bytes.Buffer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Dump(buf, DefaultTarget.Address, DefaultTarget.Port)
		buf.Reset()
	}
}

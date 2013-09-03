package check

import (
	"strings"
	"testing"
)

func TestExitListLoading(t *testing.T) {
	// Testing load
	exits := new(Exits) //new(check.Exits)
	testData := `{"Rules": [{"IsAccept": true, "MinPort": 706, "MaxPort": 706, "Address": null}, {"IsAccept": true, "MinPort": 993, "MaxPort": 993, "Address": null}, {"IsAccept": true, "MinPort": 995, "MaxPort": 995, "Address": null}], "IsAllowedDefault": false, "Address": "83.227.52.198"}
				 {"Rules": [{"IsAccept": true, "MinPort": 20, "MaxPort": 23, "Address": null}, {"IsAccept": true, "MinPort": 43, "MaxPort": 43, "Address": null}, {"IsAccept": true, "MinPort": 53, "MaxPort": 53, "Address": null}, {"IsAccept": true, "MinPort": 79, "MaxPort": 81, "Address": null}, {"IsAccept": true, "MinPort": 88, "MaxPort": 88, "Address": null}, {"IsAccept": true, "MinPort": 110, "MaxPort": 110, "Address": null}, {"IsAccept": true, "MinPort": 143, "MaxPort": 143, "Address": null}, {"IsAccept": true, "MinPort": 194, "MaxPort": 194, "Address": null}, {"IsAccept": true, "MinPort": 220, "MaxPort": 220, "Address": null}, {"IsAccept": true, "MinPort": 443, "MaxPort": 443, "Address": null}, {"IsAccept": true, "MinPort": 464, "MaxPort": 465, "Address": null}, {"IsAccept": true, "MinPort": 543, "MaxPort": 544, "Address": null}, {"IsAccept": true, "MinPort": 563, "MaxPort": 563, "Address": null}, {"IsAccept": true, "MinPort": 587, "MaxPort": 587, "Address": null}, {"IsAccept": true, "MinPort": 706, "MaxPort": 706, "Address": null}, {"IsAccept": true, "MinPort": 749, "MaxPort": 749, "Address": null}, {"IsAccept": true, "MinPort": 873, "MaxPort": 873, "Address": null}, {"IsAccept": true, "MinPort": 902, "MaxPort": 904, "Address": null}, {"IsAccept": true, "MinPort": 981, "MaxPort": 981, "Address": null}, {"IsAccept": true, "MinPort": 989, "MaxPort": 995, "Address": null}, {"IsAccept": true, "MinPort": 1194, "MaxPort": 1194, "Address": null}, {"IsAccept": true, "MinPort": 1220, "MaxPort": 1220, "Address": null}, {"IsAccept": true, "MinPort": 1293, "MaxPort": 1293, "Address": null}, {"IsAccept": true, "MinPort": 1500, "MaxPort": 1500, "Address": null}, {"IsAccept": true, "MinPort": 1723, "MaxPort": 1723, "Address": null}, {"IsAccept": true, "MinPort": 1863, "MaxPort": 1863, "Address": null}, {"IsAccept": true, "MinPort": 2082, "MaxPort": 2083, "Address": null}, {"IsAccept": true, "MinPort": 2086, "MaxPort": 2087, "Address": null}, {"IsAccept": true, "MinPort": 2095, "MaxPort": 2096, "Address": null}, {"IsAccept": true, "MinPort": 3128, "MaxPort": 3128, "Address": null}, {"IsAccept": true, "MinPort": 3389, "MaxPort": 3389, "Address": null}, {"IsAccept": true, "MinPort": 3690, "MaxPort": 3690, "Address": null}, {"IsAccept": true, "MinPort": 4321, "MaxPort": 4321, "Address": null}, {"IsAccept": true, "MinPort": 4643, "MaxPort": 4643, "Address": null}, {"IsAccept": true, "MinPort": 5050, "MaxPort": 5050, "Address": null}, {"IsAccept": true, "MinPort": 5190, "MaxPort": 5190, "Address": null}, {"IsAccept": true, "MinPort": 5222, "MaxPort": 5223, "Address": null}, {"IsAccept": true, "MinPort": 5228, "MaxPort": 5228, "Address": null}, {"IsAccept": true, "MinPort": 5900, "MaxPort": 5900, "Address": null}, {"IsAccept": true, "MinPort": 6666, "MaxPort": 6667, "Address": null}, {"IsAccept": true, "MinPort": 6679, "MaxPort": 6679, "Address": null}, {"IsAccept": true, "MinPort": 6697, "MaxPort": 6697, "Address": null}, {"IsAccept": true, "MinPort": 8000, "MaxPort": 8000, "Address": null}, {"IsAccept": true, "MinPort": 8008, "MaxPort": 8008, "Address": null}, {"IsAccept": true, "MinPort": 8080, "MaxPort": 8080, "Address": null}, {"IsAccept": true, "MinPort": 8087, "MaxPort": 8088, "Address": null}, {"IsAccept": true, "MinPort": 8443, "MaxPort": 8443, "Address": null}, {"IsAccept": true, "MinPort": 8888, "MaxPort": 8888, "Address": null}, {"IsAccept": true, "MinPort": 9418, "MaxPort": 9418, "Address": null}, {"IsAccept": true, "MinPort": 9999, "MaxPort": 10000, "Address": null}, {"IsAccept": true, "MinPort": 19294, "MaxPort": 19294, "Address": null}, {"IsAccept": true, "MinPort": 19638, "MaxPort": 19638, "Address": null}], "IsAllowedDefault": false, "Address": "91.121.43.80"}`
	exits.Load(strings.NewReader(testData))

	if len(exits.List) != 2 {
		t.Errorf("Parsed an incorrect number(%v) of policies from the test data",
			len(exits.List))
	}

	p := exits.List["83.227.52.198"]
	if p.Address != "83.227.52.198" {
		t.Errorf("Failed to parse exitList address")
	}

	bad := AddressPort{"38.229.70.31", 443}
	if p.CanExit(bad) {
		t.Errorf("Wrong answer for bad CanExit %v", bad.Port)
	}

	good := AddressPort{"38.229.70.31", 995}
	if !p.CanExit(good) {
		t.Errorf("Wrong answer for good CanExit %v", good.Port)
	}

	// Valid tor exit
	if !exits.IsTor("91.121.43.80") {
		t.Errorf("Wrong answer checking IsTor")
	}

	// check both exits are listed for 995
	// Accept either ordering of output
	addressDump := exits.Dump(good.Address, 995)
	if !(addressDump == "91.121.43.80\n83.227.52.198\n" ||
		addressDump == "83.227.52.198\n91.121.43.80\n") {
		t.Errorf("The exit dump was not as expected:\n%v", addressDump)
	}
}

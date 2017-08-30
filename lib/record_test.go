package lib

import "testing"

func initTest(t *testing.T, gtype string, param string) *Record {
	t.Logf("Check that %s entries are parsed correctly...", gtype)
	record := Record{}
	result := record.Parse(param)
	if !result {
		t.Errorf("Expected %s entry to be parsed correctly with %q, but it failed.", gtype, param)
	}
	return &record
}

func testLink(t *testing.T, record *Record, gtype string, shouldBe bool) {
	if record.IsLink() != shouldBe {
		if shouldBe {
			t.Errorf("Expected %s entry to be a link, but it wasn't.", gtype)
		} else {
			t.Errorf("Expected %s entry not to be a link, but it was.", gtype)
		}
	}
}

func testString(t *testing.T, record *Record, gtype string, display string) {
	if record.ToString() != display {
		t.Errorf("Expected %s entry to be displayed has %s, but it was %s", gtype, display, record.ToString())
	}
}

func TestFailParsing(t *testing.T) {
	param := ""
	record := Record{}
	result := record.Parse(param)
	if result {
		t.Errorf("Expected parsing to fail for %q, but it succeeded.", param)
	}
}

func TestCreateAddress(t *testing.T) {
	gtype := "Sub Menu"
	param := "0123\t/req\tgo.server.net\t42"
	record := initTest(t, gtype, param)

	address := "gopher://go.server.net:42/?q=/req&t=0"
	if record.Address != address {
		t.Errorf("Expected address for %q to be %q, but it was %q instead.", param, address, record.Address)
	}
}

func TestFile(t *testing.T) {
	gtype := "File"
	record := initTest(t, gtype, "0123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[file] 123")
}

func TestSubMenu(t *testing.T) {
	gtype := "Sub Menu"
	record := initTest(t, gtype, "1123")
	testLink(t, record, gtype, true)
	testString(t, record, gtype, "[menu] 123")
}

func TestCCSO(t *testing.T) {
	gtype := "CCSO"
	record := initTest(t, gtype, "2123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[ccso] 123")
}

func TestError(t *testing.T) {
	gtype := "File"
	record := initTest(t, gtype, "3123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[error] 123")
}

func TestBinHex(t *testing.T) {
	gtype := "BinHex"
	record := initTest(t, gtype, "4123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[binary] 123")
}

func TestDOS(t *testing.T) {
	gtype := "DOS"
	record := initTest(t, gtype, "5123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[binary] 123")
}

func TestUUEncoded(t *testing.T) {
	gtype := "UUEncoded"
	record := initTest(t, gtype, "6123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[binary] 123")
}

func TestSearch(t *testing.T) {
	gtype := "Search"
	record := initTest(t, gtype, "7123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[search] 123")
}

func TestTelnet(t *testing.T) {
	gtype := "Telnet"
	record := initTest(t, gtype, "8123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[telnet] 123")
}

func TestBinary(t *testing.T) {
	gtype := "Binary"
	record := initTest(t, gtype, "9123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[binary] 123")
}

func TestRedundant(t *testing.T) {
	gtype := "Redundant"
	record := initTest(t, gtype, "+123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[mirror] 123")
}

func TestTelnet3270(t *testing.T) {
	gtype := "Telnet 3270"
	record := initTest(t, gtype, "T123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[telnet] 123")
}

func TestGIF(t *testing.T) {
	gtype := "GIF"
	record := initTest(t, gtype, "g123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[image] 123")
}

func TestImage(t *testing.T) {
	gtype := "Image"
	record := initTest(t, gtype, "I123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[image] 123")
}

func TestHTML(t *testing.T) {
	gtype := "HTML"
	record := initTest(t, gtype, "h123")
	testLink(t, record, gtype, true)
	testString(t, record, gtype, "[html] 123")
}

func TestInfo(t *testing.T) {
	gtype := "Informational"
	record := initTest(t, gtype, "i123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "123")
}

func TestSound(t *testing.T) {
	gtype := "Sound"
	record := initTest(t, gtype, "s123")
	testLink(t, record, gtype, false)
	testString(t, record, gtype, "[sound] 123")
}

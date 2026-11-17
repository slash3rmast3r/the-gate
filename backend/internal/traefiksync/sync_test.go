package traefiksync

import "testing"

func TestSanitize(t *testing.T) {
	got := sanitize("App.Example.COM")
	if got != "app-example-com" {
		t.Fatalf("sanitize host: got %q", got)
	}
}

func TestParseCIDRs(t *testing.T) {
	got := parseCIDRs(" 10.0.0.0/8 , 192.168.1.0/24 ")
	if len(got) != 2 || got[0] != "10.0.0.0/8" || got[1] != "192.168.1.0/24" {
		t.Fatalf("unexpected cidrs: %#v", got)
	}
	if parseCIDRs("") != nil {
		t.Fatal("empty should be nil")
	}
}

func TestWWWPlan(t *testing.T) {
	hosts, from, to := wwwPlan("example.com", "force_www")
	if len(hosts) != 1 || hosts[0] != "www.example.com" || from != "example.com" || to != "www.example.com" {
		t.Fatalf("force_www: %#v %q %q", hosts, from, to)
	}
	hosts, from, to = wwwPlan("www.example.com", "force_apex")
	if len(hosts) != 1 || hosts[0] != "example.com" || from != "www.example.com" || to != "example.com" {
		t.Fatalf("force_apex: %#v %q %q", hosts, from, to)
	}
}

func TestParseHeaderMap(t *testing.T) {
	m := parseHeaderMap("X-A: 1\n# comment\nX-B: two")
	if m["X-A"] != "1" || m["X-B"] != "two" {
		t.Fatalf("%#v", m)
	}
}

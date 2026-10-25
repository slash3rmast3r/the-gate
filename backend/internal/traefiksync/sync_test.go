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

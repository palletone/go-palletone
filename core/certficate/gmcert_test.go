package certficate

import "testing"

func TestGMGenCert(t *testing.T) {
	address := "P1CkorqRjxs8cQeQ7NQkGcpiLMSSiJYMX1a"
	certbyte, err := GenSMCert(address)
	if err != nil {
		t.Log(err)
	}
	t.Log(certbyte)
	}

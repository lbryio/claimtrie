package tests

import (
	"../claim"
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func TestNormalization(t *testing.T) {
	t.Log("ICU Version: " + claim.IcuVersion())
	assertEqual(t, "test", string(claim.Normalize([]byte("TESt"))))
	assertEqual(t, "test 23", string(claim.Normalize([]byte("tesT 23"))))
	assertEqual(t, "\xFF", string(claim.Normalize([]byte("\xFF"))))
	assertEqual(t, "\xC3\x28", string(claim.Normalize([]byte("\xC3\x28"))))
	assertEqual(t, "\xCF\x89", string(claim.Normalize([]byte("\xE2\x84\xA6"))))
	assertEqual(t, "\xD1\x84", string(claim.Normalize([]byte("\xD0\xA4"))))
	assertEqual(t, "\xD5\xA2", string(claim.Normalize([]byte("\xD4\xB2"))))
	assertEqual(t, "\xE3\x81\xB5\xE3\x82\x99", string(claim.Normalize([]byte("\xE3\x81\xB6"))))
	assertEqual(t, "\xE1\x84\x81\xE1\x85\xAA\xE1\x86\xB0", string(claim.Normalize([]byte("\xEA\xBD\x91"))))
}
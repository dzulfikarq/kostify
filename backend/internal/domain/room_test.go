package domain

import (
	"database/sql/driver"
	"testing"
)

func TestStringArrayRoundtrip(t *testing.T) {
	in := StringArray{"ac", "wifi"}
	v, err := in.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	s, ok := v.(string)
	if !ok || s != `["ac","wifi"]` {
		t.Fatalf("dapat %v (%T)", v, v)
	}
	var out StringArray
	if err := out.Scan(s); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(out) != 2 || out[0] != "ac" || out[1] != "wifi" {
		t.Fatalf("dapat %v", out)
	}
}

func TestStringArrayNilAndBytes(t *testing.T) {
	var a StringArray
	v, _ := a.Value()
	if v != "[]" {
		t.Fatalf("nil harus [], dapat %v", v)
	}
	var b StringArray
	if err := b.Scan([]byte(`["x"]`)); err != nil || len(b) != 1 || b[0] != "x" {
		t.Fatalf("scan bytes: %v %v", b, err)
	}
	if err := (&a).Scan(123); err == nil {
		t.Fatal("tipe salah harus error")
	}
}

func TestOwnerAllowedRoomTransition(t *testing.T) {
	cases := []struct {
		from, to RoomStatus
		want     bool
	}{
		{RoomAvailable, RoomMaintenance, true},
		{RoomMaintenance, RoomAvailable, true},
		{RoomAvailable, RoomBooked, false},
		{RoomAvailable, RoomAvailable, false},
		{RoomActive, RoomAvailable, false},
	}
	for _, c := range cases {
		if got := OwnerAllowedRoomTransition(c.from, c.to); got != c.want {
			t.Errorf("%s->%s: dapat %v, mau %v", c.from, c.to, got, c.want)
		}
	}
}

var _ driver.Valuer = StringArray{}

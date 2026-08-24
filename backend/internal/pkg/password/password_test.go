package password

import "testing"

func TestHashVerify(t *testing.T) {
	h, err := Hash("Rahasia123", 4)
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	if !Verify(h, "Rahasia123") {
		t.Error("verify harus cocok")
	}
	if Verify(h, "salah123") {
		t.Error("verify tidak boleh cocok untuk password salah")
	}
}

func TestWeaknesses(t *testing.T) {
	cases := []struct {
		pw      string
		wantLen int
	}{
		{"Rahasia123", 0},
		{"Rahasiaaa", 1},
		{"tanpaangka", 1},
		{"12345678", 1},
	}
	for _, c := range cases {
		got := Weaknesses(c.pw)
		if len(got) != c.wantLen {
			t.Errorf("Weaknesses(%q) = %v, want %d masalah", c.pw, got, c.wantLen)
		}
	}
}

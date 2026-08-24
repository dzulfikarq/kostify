package domain

import "testing"

func TestBookingStatusValid(t *testing.T) {
	for _, s := range []BookingStatus{
		BookingPending, BookingSurvey, BookingBooked, BookingActive,
		BookingCompleted, BookingCancelled, BookingRejected, BookingExpired,
	} {
		if !s.IsValid() {
			t.Errorf("%s harus valid", s)
		}
	}
	if BookingStatus("hacked").IsValid() {
		t.Error("status asing tidak boleh valid")
	}
}

// ponytail: state machine dijaga di repo.Transition (from harus match); test ini
// mendokumentasikan transisi legal sebagai referensi.
func TestBookingLegalTransitions(t *testing.T) {
	legal := map[BookingStatus][]BookingStatus{
		BookingPending:   {BookingSurvey, BookingRejected, BookingCancelled, BookingExpired},
		BookingSurvey:    {BookingBooked, BookingRejected, BookingCancelled},
		BookingBooked:    {BookingActive, BookingExpired},
		BookingActive:    {BookingCompleted},
		BookingCompleted: {},
		BookingCancelled: {},
		BookingRejected:  {},
		BookingExpired:   {},
	}
	for from, tos := range legal {
		for _, to := range tos {
			if !to.IsValid() || !from.IsValid() {
				t.Errorf("transisi %s→%s pakai status tidak valid", from, to)
			}
		}
	}
}

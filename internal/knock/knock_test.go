package knock

import (
	"testing"
	"time"
)

func TestTrackerCompletesOnCorrectSequence(t *testing.T) {
	tr := NewTracker([]int{4000, 4001, 4002}, 5*time.Second)
	now := time.Now()
	if tr.Knock("1.2.3.4", 4000, now) {
		t.Fatal("completed too early after 1 knock")
	}
	if tr.Knock("1.2.3.4", 4001, now) {
		t.Fatal("completed too early after 2 knocks")
	}
	if !tr.Knock("1.2.3.4", 4002, now) {
		t.Fatal("expected completion after full sequence")
	}
}

func TestTrackerWrongPortResets(t *testing.T) {
	tr := NewTracker([]int{4000, 4001, 4002}, 5*time.Second)
	now := time.Now()
	tr.Knock("1.2.3.4", 4000, now)
	tr.Knock("1.2.3.4", 9999, now) // wrong -> reset
	if tr.Knock("1.2.3.4", 4001, now) {
		t.Fatal("must not complete: sequence was reset by wrong port")
	}
	// Correct sequence still works afterwards.
	tr.Knock("1.2.3.4", 4000, now)
	tr.Knock("1.2.3.4", 4001, now)
	if !tr.Knock("1.2.3.4", 4002, now) {
		t.Fatal("expected completion on a fresh correct sequence")
	}
}

func TestTrackerWindowExpiry(t *testing.T) {
	tr := NewTracker([]int{4000, 4001}, 100*time.Millisecond)
	now := time.Now()
	tr.Knock("1.2.3.4", 4000, now)
	if tr.Knock("1.2.3.4", 4001, now.Add(200*time.Millisecond)) {
		t.Fatal("must not complete: second knock outside window")
	}
}

func TestTrackerIsolatesSources(t *testing.T) {
	tr := NewTracker([]int{4000, 4001}, 5*time.Second)
	now := time.Now()
	tr.Knock("1.1.1.1", 4000, now)
	// A different IP's knock must not advance the first IP.
	if tr.Knock("2.2.2.2", 4001, now) {
		t.Fatal("different source should not complete")
	}
	if !tr.Knock("1.1.1.1", 4001, now) {
		t.Fatal("original source should complete its own sequence")
	}
}

func TestTrackerRestartOnFirstPort(t *testing.T) {
	tr := NewTracker([]int{4000, 4001, 4002}, 5*time.Second)
	now := time.Now()
	tr.Knock("1.2.3.4", 4000, now)
	tr.Knock("1.2.3.4", 4001, now)
	tr.Knock("1.2.3.4", 4000, now) // wrong here, but it's the first port -> restart at idx 1
	tr.Knock("1.2.3.4", 4001, now)
	if !tr.Knock("1.2.3.4", 4002, now) {
		t.Fatal("expected completion after restart-on-first-port")
	}
}

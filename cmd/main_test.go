package main

import (
	"testing"
	"time"
)

const (
	onlineState  = `(207.3 207.3 206.8 017 49.9 13.2 25.0 00001001`
	offlineState = `(207.3 207.3 206.8 017 49.9 13.2 25.0 10001001`
)

func TestNoShutdown(t *testing.T) {
	count := 0
	runShutdown = func() {
		count++
	}

	state := onlineState
	queryState = func(*config) (string, error) {
		return state, nil
	}

	cfg := config{delay: 500 * time.Millisecond}
	deadline := time.Time{}

	checkStatus(&cfg, &deadline)
	if count != 0 {
		t.Fatalf("Shutdown called %d times", count)
	}

	time.Sleep(1 * time.Second)

	checkStatus(&cfg, &deadline)
	if count != 0 {
		t.Fatalf("Shutdown called %d times", count)
	}

	state = offlineState

	checkStatus(&cfg, &deadline)
	if count != 0 {
		t.Fatalf("Shutdown called %d times", count)
	}

	time.Sleep(150 * time.Millisecond)

	checkStatus(&cfg, &deadline)
	if count != 0 {
		t.Fatalf("Shutdown called %d times", count)
	}

	time.Sleep(150 * time.Millisecond)
	state = onlineState

	checkStatus(&cfg, &deadline)
	if count != 0 {
		t.Fatalf("Shutdown called %d times", count)
	}

	time.Sleep(2 * time.Second)
	checkStatus(&cfg, &deadline)
	if count != 0 {
		t.Fatalf("Shutdown called %d times", count)
	}
}

func TestWithShutdown(t *testing.T) {
	count := 0
	runShutdown = func() {
		count++
	}

	state := onlineState
	queryState = func(*config) (string, error) {
		return state, nil
	}

	cfg := config{delay: 500 * time.Millisecond}
	deadline := time.Time{}

	checkStatus(&cfg, &deadline)
	if count != 0 {
		t.Fatalf("Shutdown called %d times", count)
	}

	state = offlineState
	checkStatus(&cfg, &deadline)
	if count != 0 {
		t.Fatalf("Shutdown called %d times", count)
	}

	time.Sleep(1 * time.Second)
	checkStatus(&cfg, &deadline)
	if count != 1 {
		t.Fatalf("Shutdown was not called")
	}
}

package ism

import (
	"testing"
)

func TestNoMatchesEmpty(t *testing.T) {
	a, err := InitAutomaton(nil)
	if err != nil {
		t.Fatalf("InitAutomaton failed: %v", err)
	}

	m, err := a.Match([]rune("apple banana carrot"))
	if err != nil {
		t.Fatalf("Match failed: %v", err)
	}

	if len(m) != 0 {
		t.Fatalf("Match returned matches that should not exist")
	}
}

func TestZeroMatches(t *testing.T) {
	strings := [][]rune{
		[]rune("apple"),
		[]rune("banana"),
		[]rune("carrot"),
	}

	a, err := InitAutomaton(strings)
	if err != nil {
		t.Fatalf("InitAutomaton failed: %v", err)
	}

	m, err := a.Match([]rune("tomato purple yellow"))
	if err != nil {
		t.Fatalf("Match failed: %v", err)
	}

	if len(m) != 0 {
		t.Fatalf("Match returned matches that should not exist")
	}
}

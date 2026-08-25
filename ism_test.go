package ism

import (
	"slices"
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

type ts struct {
	s []rune
	o int
}

func to(p []ts) [][]rune {
	b := make([][]rune, 0, 100)
	for _, c := range p {
		b = append(b, c.s)
	}

	return b
}

func testNeedlesOnHaystack(t *testing.T, needles []ts, haystack []rune) {
	a, err := InitAutomaton(to(needles))
	if err != nil {
		t.Fatalf("InitAutomaton failed: %v", err)
	}

	matches, err := a.Match(haystack)
	if err != nil {
		t.Fatalf("Match failed: %v", err)
	}

	for n, needle := range needles {
		isMatching := false

		for _, match := range matches {
			if slices.Equal(needle.s, match.Word) && needle.o == matches[n].Off {
				isMatching = true
			}
		}

		if !isMatching {
			t.Fatalf("match case %d: cannot find match where off=%d word=%s", n, needle.o, string(needle.s))
		}
	}
}

func TestMatchBasicSeveral(t *testing.T) {
	needles := []ts{
		ts{s: []rune("sir"), o: 3},
		ts{s: []rune("cat"), o: 7},
		ts{s: []rune("slam"), o: 35},
		ts{s: []rune("car"), o: 44},
		ts{s: []rune("clam"), o: 66},
	}

	testNeedlesOnHaystack(t, needles, []rune("sir cat, you really should not slam your car door shut like a clam\""))
}

func TestMatchPrefixes(t *testing.T) {
	needles := []ts{
		// note: this is omitted because it will trigger on every case which the harness cannot differentiate
		// ts{s: []rune("p"), o: 1},
		ts{s: []rune("pu"), o: 2},
		ts{s: []rune("pur"), o: 3},
		ts{s: []rune("purp"), o: 4},
		ts{s: []rune("purpl"), o: 5},
		ts{s: []rune("purple"), o: 6},
	}

	testNeedlesOnHaystack(t, needles, []rune("purple yellow orange"))
}

func TestMatchSuffixes(t *testing.T) {
	needles := []ts{
		ts{s: []rune("purple"), o: 6},
		ts{s: []rune("urple"), o: 6},
		ts{s: []rune("rple"), o: 6},
		ts{s: []rune("ple"), o: 6},
		ts{s: []rune("le"), o: 6},
	}

	testNeedlesOnHaystack(t, needles, []rune("purple rock"))
}

// init match add match where both times should be zero results
func TestMatchAddSameZero(t *testing.T) {
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

	if err = a.AddString([]rune("detective")); err != nil {
		t.Fatalf("AddString failed: %v", err)
	}

	m, err = a.Match([]rune("tomato purple yellow"))
	if err != nil {
		t.Fatalf("Match failed: %v", err)
	}

	if len(m) != 0 {
		t.Fatalf("Match returned matches that should not exist")
	}
}

func TestMatchAddSameMatches(t *testing.T) {
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

	m2, err := a.Match([]rune("tomato purple yellow"))
	if err != nil {
		t.Fatalf("second Match call failed: %v", err)
	}

	if len(m) != len(m2) {
		t.Fatalf("matches have different lengths, %d != %d", len(m), len(m2))
	}

	for i, match := range m {
		if match.Off != m2[i].Off {
			t.Fatalf("match %d: offset mismatch. %d != %d", i, match.Off, m2[i].Off)
		}

		if len(match.Word) != len(m2[i].Word) {
			t.Fatalf("match %d: words lens mismatch. %d != %d", i, len(match.Word), len(m2[i].Word))
		}

		if !slices.Equal(match.Word, m2[i].Word) {
			t.Fatalf("match %d: word mismatch. want=%q have=%q", i, string(match.Word), string(m2[i].Word))
		}
	}
}

func TestMatchAddNewMatches(t *testing.T) {
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

	m2, err := a.Match([]rune("tomato purple yellow"))
	if err != nil {
		t.Fatalf("second Match call failed: %v", err)
	}

	if len(m) != len(m2) {
		t.Fatalf("matches have different lengths, %d != %d", len(m), len(m2))
	}
}

func BenchmarkInit(b *testing.B) {
	master := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzzyxwvutsrqponmlkjihgfedcba9876543210ZYXWVUTSRQPONMLKJIHGFEDCBA")
	strings := make([][]rune, 0, 10000)
	for i := range 10000 {
		j := i % len(master)
		strings = append(strings, master[j:])
	}

	for b.Loop() {
		_, err := InitAutomaton(strings)
		if err != nil {
			b.Fatalf("InitAutomaton failed: %v", err)
		}
	}
}

func BenchmarkAddString(b *testing.B) {
	master := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzzyxwvutsrqponmlkjihgfedcba9876543210ZYXWVUTSRQPONMLKJIHGFEDCBA")
	strings := make([][]rune, 0, 10000)
	for i := range 10000 {
		j := i % len(master)
		strings = append(strings, master[j:])
	}

	a, err := InitAutomaton([][]rune{})
	if err != nil {
		b.Fatalf("InitAutomaton failed: %v", err)
	}

	i := 0
	for ; b.Loop(); i++ {
		j := i % len(strings)
		err := a.AddString(strings[j])
		if err != nil {
			b.Fatalf("AddString failed: %v", err)
		}
	}
}

func BenchmarkMatchBasic(b *testing.B) {
	strings := [][]rune{
		[]rune("apple"),
		[]rune("banana"),
		[]rune("carrot"),
	}

	a, err := InitAutomaton(strings)
	if err != nil {
		b.Fatalf("InitAutomaton failed: %v", err)
	}

	k := []rune("purpleapplebantabananacarrotpureriqwer")

	for b.Loop() {
		_, err := a.Match(k)
		if err != nil {
			b.Fatalf("Match failed: %v", err)
		}
	}
}

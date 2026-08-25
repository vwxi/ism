// go implementation of bertrand meyer's "incremental string matching" (https://se.inf.ethz.ch/~meyer/publications/string/string_matching.pdf)
// this works on rune slices for compat with different languages
package ism

import (
	"errors"
	"slices"
)

var ErrNullPointer = errors.New("null pointer dereference")
var ErrOutputDoesNotExist = errors.New("output set does not exist")
var ErrFailDoesNotExist = errors.New("failing link does not exist")
var ErrInvFailDoesNotExist = errors.New("inverse failing set does not exist")
var ErrChildDoesNotExist = errors.New("child does not exist")
var ErrTrieFindFailed = errors.New("trie find failed")
var ErrEmptyNodeData = errors.New("empty node data")

type node struct {
	root bool

	parent *node

	data []rune

	// trie T
	T map[rune]*node

	// output set (keys are indices into keywords array)
	O map[uint64]struct{}

	// failing
	F  *node
	IF map[*node]struct{}
}

// Automaton represents an incremental version of an Aho-Corasick automaton per the paper
type Automaton struct {
	root *node

	keywords [][]rune
}

type Match struct {
	Off int

	Word []rune
}

func (n *node) AddOutput(kwIdx uint64) error {
	if n.O == nil {
		return ErrOutputDoesNotExist
	}

	if n.IF == nil {
		return ErrInvFailDoesNotExist
	}

	// O[n] := O[n] U {s}
	n.O[kwIdx] = struct{}{}

	// for x in IF[n] loop enter_output(x, s) end for
	for x, _ := range n.IF {
		if err := x.AddOutput(kwIdx); err != nil {
			return err
		}
	}

	return nil
}

func (t *Automaton) addChild(n *node, c rune, data []rune) error {
	if n == nil {
		return ErrNullPointer
	}

	// n'
	np := &node{
		root:   false,
		parent: n,
		data:   data,
		T:      make(map[rune]*node),
		O:      make(map[uint64]struct{}),
		IF:     make(map[*node]struct{}),
	}

	// T[n, c] := n'
	n.T[c] = np

	// complete_failure(n, c)
	if err := t.completeFailure(n, c); err != nil {
		return err
	}

	if np.F == nil {
		return ErrFailDoesNotExist
	}

	// IF[F[n']] := IF[F[n']] U {n'}
	// avoid adding a node to its own IF (which would create a cycle)
	if np.F != nil && np.F != np {
		np.F.IF[np] = struct{}{}
	}

	// complete_inverse(n, n', c)
	return t.completeInverse(n, np, c)
}

// AddString adds a needle string to the automaton
func (t *Automaton) AddString(needle []rune) error {
	n := t.root

	t.keywords = append(t.keywords, needle)
	idx := uint64(len(t.keywords) - 1)

	for i, c := range needle {
		if _, has := n.T[c]; !has {
			if err := t.addChild(n, c, needle[:(i+1)]); err != nil {
				return err
			}
		}

		n = n.T[c]
	}

	return n.AddOutput(idx)
}

// RemoveString is the inverse operation to AddString. it removes a needle string from the automaton
func (t *Automaton) RemoveString(needle []rune) error {
	// no-op if no keywords
	if len(t.keywords) == 0 {
		return nil
	}

	n := t.root

	// get intIdx
	intIdx := -1
	for idx, kw := range t.keywords {
		if slices.Equal(kw, needle) {
			intIdx = idx
			break
		}
	}
	if intIdx == -1 {
		return ErrTrieFindFailed
	}
	kwIdx := uint64(intIdx)

	// dfs to get node and its parent
	for _, c := range needle {
		if k, has := n.T[c]; !has {
			return ErrTrieFindFailed
		} else {
			n = k
		}
	}

	if n == nil {
		return ErrNullPointer
	}

	// remove references. go up fail to remove self. go down IF to remove self
	if n.F != nil {
		delete(n.F.IF, n)
	}

	if len(n.data) == 0 {
		return ErrEmptyNodeData
	}

	// for every inverse fail remove their fail link and any output link to us
	for i, _ := range n.IF {
		if i == nil {
			return ErrNullPointer
		}

		i.F = nil

		delete(i.O, kwIdx)
		delete(n.IF, i)
	}

	delete(n.O, kwIdx)

	// prune tree
	for n.parent != nil {
		p := n.parent

		if len(n.data) == 0 {
			return ErrEmptyNodeData
		}

		last := n.data[len(n.data)-1]

		// do not touch a node that is being referred to
		if len(n.T) != 0 || len(n.O) != 0 {
			break
		}

		if len(n.IF) != 0 {
			break
		}

		delete(p.T, last)
		n = p
	}

	return nil
}

// bfs is a helper for breadth first search
func (n *node) bfs(fn func(c *node) error) error {
	q := []*node{n}
	v := make(map[*node]struct{})

	for len(q) != 0 {
		curr := q[0]
		v[curr] = struct{}{}
		q = q[1:]

		for _, child := range curr.T {
			if _, has := v[child]; !has {
				q = append(q, child)
				v[child] = struct{}{}
			}
		}

		if !curr.root {
			if err := fn(curr); err != nil {
				return err
			}
		}
	}

	return nil
}

// findNode traverses a node based on a given string and returns a node if found
func (t *Automaton) findNode(str []rune) (*node, error) {
	n := t.root
	for _, c := range str {
		if _, has := n.T[c]; has {
			n = n.T[c]
		} else {
			return nil, ErrTrieFindFailed
		}
	}

	return n, nil
}

func (t *Automaton) longestProperSuffix(str []rune) (*node, error) {
	var n *node
	var err error
	if len(str) == 0 {
		return t.root, nil
	}

	for i := 1; i < len(str); i++ {
		currStr := str[i:]
		n, err = t.findNode(currStr)
		if err == nil { // first match will be the longest
			return n, nil
		}
	}

	return n, err
}

// lps finds the longest proper suffix of a given string and returns the root if there isn't one
func (t *Automaton) lps(n *node) *node {
	lpsRes, err := t.longestProperSuffix(n.data)
	if err != nil || lpsRes == nil {
		return t.root
	} else {
		return lpsRes
	}
}

func (t *Automaton) completeFailure(n *node, c rune) error {
	// n' := T[n, c]; m := n
	np, has := n.T[c]

	if !has {
		return t.addChild(n, c, append(n.data, c))
	}

	m := n

	for { // repeat
		// m := F[m]
		m = m.F

		mp, has := m.T[c]
		if m.root || (has && mp != np) {
			break
		}
	}

	if m == nil {
		return ErrNullPointer
	}

	// m' := T[m, c]
	mp, has := m.T[c]

	if !has || mp == np {
		// if there is no child and we are at a root node
		// then the fail link goes to root
		if m.root {
			mp = t.root
		} else {
			return ErrChildDoesNotExist
		}
	}

	// F[n'] := m'
	np.F = mp
	// O[n'] := O[n'] U O[m']
	for mc, _ := range mp.O {
		np.O[mc] = struct{}{}
	}

	return nil
}

func (t *Automaton) completeInverse(n *node, np *node, c rune) error {
	// this should never fire because every node should have an IF
	if n.IF == nil {
		return ErrInvFailDoesNotExist
	}

	for x, _ := range n.IF {
		// if T[x, c] exists
		// x' := T[x, c]
		if xp, has := x.T[c]; has {
			if xp.F == nil {
				return ErrFailDoesNotExist
			}

			// IF[F[x']] = IF[F[x']] - {x'}
			delete(xp.F.IF, xp)

			// F[x'] := n'; IF[n'] := IF[n'] U {x'};
			xp.F = np
			np.IF[xp] = struct{}{}
		} else {
			// try recursively with nodes having x as proper suffix
			if err := t.completeInverse(x, np, c); err != nil {
				return err
			}
		}
	}

	return nil
}

// buildFailure builds the failure function for a given tree
func (t *Automaton) buildFailure() error {
	err := t.root.bfs(func(n *node) error {
		n.F = t.lps(n)

		return nil
	})
	if err != nil {
		return err
	}

	return t.root.bfs(func(n *node) error {
		for c, _ := range n.T {
			if err := t.completeFailure(n, c); err != nil {
				return err
			}
		}

		return nil
	})
}

// InitAutomaton builds the automaton. on an empty slice it will return an empty automaton
func InitAutomaton(strings [][]rune) (*Automaton, error) {
	rootNode := &node{
		root: true,
		T:    make(map[rune]*node),
		O:    make(map[uint64]struct{}),
		IF:   make(map[*node]struct{}),
	}

	rootNode.F = rootNode

	tree := &Automaton{
		root: rootNode,
	}

	if strings == nil {
		return tree, nil
	}

	// make children from all of the provided strings
	for _, str := range strings {
		if err := tree.AddString(str); err != nil {
			return nil, err
		}
	}

	if err := tree.buildFailure(); err != nil {
		return nil, err
	}

	return tree, nil
}

// Match returns a list of all rune slice matches in a given haystack
//
// the offsets returned correspond to the character immediately proceeding the matched keyword
func (t *Automaton) Match(haystack []rune) ([]Match, error) {
	matches := make([]Match, 0, 100)
	n := t.root

	for i, c := range haystack {
		for !n.root && n.T[c] == nil {
			if n.F == n {
				n = t.root
				break
			}

			n = n.F
		}

		if n.T[c] != nil {
			n = n.T[c]

			if len(n.O) != 0 {
				for w, _ := range n.O {
					if w >= uint64(len(t.keywords)) {
						return matches, ErrOutputDoesNotExist
					}

					matches = append(matches, Match{
						Off:  i + 1,
						Word: t.keywords[w],
					})
				}
			}
		}
	}

	return matches, nil
}

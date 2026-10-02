package core

import (
	"reflect"
	"testing"
)

func TestPuzzleSolutionsAndSolvedPersistence(t *testing.T) {
	for _, tc := range []struct {
		kind     PuzzleKind
		solution []int
	}{
		{RunePuzzle, []int{1, 0, 2, 1}},
		{CircuitPuzzle, []int{0, 2, 4, 6, 8}},
		{OfferingPuzzle, []int{0, 0, 1}},
	} {
		p := NewPuzzle(tc.kind)
		if p.Solved || p.Name() == "" || p.Clue() == "" {
			t.Fatal("invalid initial puzzle")
		}
		for i, choice := range tc.solution {
			p.Select(choice)
			if p.Solved && i < len(tc.solution)-1 {
				t.Fatal("puzzle solved before complete solution")
			}
		}
		if !p.Solved {
			t.Fatalf("kind %d did not solve", tc.kind)
		}
		before := *p
		before.Choices = append([]int(nil), p.Choices...)
		p.Reset()
		p.Select(0)
		if !reflect.DeepEqual(before, *p) {
			t.Fatal("solved puzzle changed")
		}
	}
}

func TestPuzzleFailureResetAndInvalidChoices(t *testing.T) {
	for _, kind := range []PuzzleKind{RunePuzzle, CircuitPuzzle, OfferingPuzzle} {
		p := NewPuzzle(kind)
		before := *p
		p.Select(-1)
		p.Select(p.OptionCount())
		if !reflect.DeepEqual(before, *p) {
			t.Fatal("invalid selection changed puzzle")
		}
		p.Select(0)
		p.Reset()
		if !reflect.DeepEqual(before, *p) {
			t.Fatal("reset did not restore initial state")
		}
	}
	for _, kind := range []PuzzleKind{RunePuzzle, OfferingPuzzle} {
		p := NewPuzzle(kind)
		count := 3
		if kind == RunePuzzle {
			count = 4
		}
		for i := 0; i < count; i++ {
			p.Select(2)
			if i < count-1 && p.Count != i+1 {
				t.Fatal("failure reset too early")
			}
		}
		if p.Solved || p.Count != 0 || p.Sum != 0 || len(p.Choices) != 0 {
			t.Fatal("failed attempt not reset")
		}
	}
	for _, order := range [][]int{{1, 0, 0}, {0, 1, 0}} {
		p := NewPuzzle(OfferingPuzzle)
		for _, choice := range order {
			p.Select(choice)
		}
		if !p.Solved {
			t.Fatal("offering order changed correct sum")
		}
	}
}

func TestCircuitNeighborsAndReversibility(t *testing.T) {
	p := NewPuzzle(CircuitPuzzle)
	initial := p.Lamps
	p.Select(0)
	if p.Lamps != initial^(1<<0|1<<1|1<<3) {
		t.Fatal("corner toggle wraps across grid")
	}
	p.Select(0)
	if p.Lamps != initial {
		t.Fatal("toggle not reversible")
	}
	p.Select(4)
	if p.Lamps != initial^(1<<4|1<<1|1<<3|1<<5|1<<7) {
		t.Fatal("center neighbors incorrect")
	}
}

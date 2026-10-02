package core

type PuzzleKind int

const (
	RunePuzzle PuzzleKind = iota
	CircuitPuzzle
	OfferingPuzzle
)

type Puzzle struct {
	Kind       PuzzleKind
	Solved     bool
	Choices    []int
	Lamps      uint16
	Sum, Count int
}

func NewPuzzle(kind PuzzleKind) *Puzzle {
	p := &Puzzle{Kind: kind}
	p.Reset()
	return p
}

func (p *Puzzle) Name() string {
	switch p.Kind {
	case RunePuzzle:
		return "Runes of the cycle"
	case CircuitPuzzle:
		return "Nine lamps"
	default:
		return "Three offerings"
	}
}

func (p *Puzzle) Clue() string {
	switch p.Kind {
	case RunePuzzle:
		return "Night yields to dawn; stars return us to night."
	case CircuitPuzzle:
		return "Light all nine. Flip a stone and its up/down/left/right neighbors."
	default:
		return "Three offerings weigh seven. Count the flame twice."
	}
}

func (p *Puzzle) OptionCount() int {
	if p.Kind == CircuitPuzzle {
		return 9
	}
	return 3
}

func (p *Puzzle) toggle(index int) {
	x, y := index%3, index/3
	for _, d := range []Point{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nx, ny := x+d.X, y+d.Y
		if nx >= 0 && nx < 3 && ny >= 0 && ny < 3 {
			p.Lamps ^= 1 << (ny*3 + nx)
		}
	}
}

func (p *Puzzle) Reset() {
	if p.Solved {
		return
	}
	p.Choices, p.Sum, p.Count = nil, 0, 0
	p.Lamps = 0
	if p.Kind == CircuitPuzzle {
		p.Lamps = 511
		for _, index := range []int{0, 2, 4, 6, 8} {
			p.toggle(index)
		}
	}
}

func (p *Puzzle) Select(index int) string {
	if p.Solved {
		return "The tablet already shines."
	}
	if index < 0 || index >= p.OptionCount() {
		return "Choose a marked stone."
	}
	if p.Kind == CircuitPuzzle {
		p.toggle(index)
		p.Solved = p.Lamps == 511
	} else {
		p.Choices = append(p.Choices, index)
		p.Count = len(p.Choices)
		if p.Kind == RunePuzzle {
			if p.Count < 4 {
				return "The cycle remembers your rune."
			}
			p.Solved = p.Choices[0] == 1 && p.Choices[1] == 0 && p.Choices[2] == 2 && p.Choices[3] == 1
		} else {
			p.Sum += [...]int{2, 3, 5}[index]
			if p.Count < 3 {
				return "The altar weighs your offering."
			}
			p.Solved = p.Sum == 7
		}
		if !p.Solved {
			p.Reset()
			return "The pattern fades. Try again."
		}
	}
	if p.Solved {
		return "The tablet shines. The seal is lifted!"
	}
	return "The lamps shift. Light all nine."
}

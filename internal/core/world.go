package core

type AreaID int

const (
	Reach AreaID = iota
	Shrine
)

type Tile byte

const (
	Grass    Tile = '.'
	Path     Tile = ','
	Water    Tile = '~'
	Wall     Tile = '#'
	Keep     Tile = 'K'
	Portal   Tile = 'S'
	Merchant Tile = 'M'
	Lever    Tile = 'L'
	Gate     Tile = 'D'
	Altar    Tile = 'A'
	Tablet   Tile = 'R'
)

type Chest struct {
	Item, Count int
	Opened      bool
}

type Encounter struct {
	Kind    EnemyKind
	Cleared bool
	Echo    bool
}

type World struct {
	Tiles      [][]Tile
	Chests     map[Point]*Chest
	Encounters map[Point]*Encounter
	Switches   map[Point]bool
	Puzzles    map[Point]*Puzzle
}

func NewWorld() World {
	rows := []string{
		"#####################",
		"#....~~~............#",
		"#....~~~..######....#",
		"#....#....#.S..#....#",
		"#..####...#....#....#",
		"#..#..#...###.##....#",
		"#.M#..#............K#",
		"#,,,,,,,,,.....##...#",
		"#..#..#.,......##...#",
		"#..####.,.~~~.......#",
		"#.......,..~~~...####",
		"#####################",
	}
	w := World{
		Chests:     map[Point]*Chest{{6, 4}: {Item: 1, Count: 1}, {14, 8}: {Item: 0, Count: 2}},
		Encounters: map[Point]*Encounter{{6, 7}: {Kind: EnemyWisp}, {9, 7}: {Kind: EnemyWisp}},
	}
	for _, row := range rows {
		w.Tiles = append(w.Tiles, []Tile(row))
	}
	return w
}

func NewShrineWorld() World {
	w := World{
		Chests: map[Point]*Chest{{35, 18}: {Item: 1, Count: 1}},
		Encounters: map[Point]*Encounter{
			{6, 12}: {Kind: EnemyDragon}, {33, 12}: {Kind: EnemyWisp},
			{6, 18}: {Kind: EnemyDragon, Echo: true}, {6, 15}: {Kind: EnemyWisp, Echo: true},
			{33, 18}: {Kind: EnemyWisp, Echo: true}, {33, 15}: {Kind: EnemyDragon, Echo: true},
			{20, 17}: {Kind: EnemyDragon, Echo: true}, {20, 11}: {Kind: EnemyWisp, Echo: true},
		},
		Switches: map[Point]bool{{6, 4}: false, {33, 4}: false},
		Puzzles:  map[Point]*Puzzle{{4, 8}: NewPuzzle(RunePuzzle), {34, 8}: NewPuzzle(CircuitPuzzle), {19, 14}: NewPuzzle(OfferingPuzzle)},
	}
	for y := 0; y < 24; y++ {
		row := make([]Tile, 40)
		for x := range row {
			row[x] = Wall
		}
		w.Tiles = append(w.Tiles, row)
	}
	carve := func(x0, y0, x1, y1 int) {
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				w.Tiles[y][x] = Path
			}
		}
	}
	carve(2, 21, 33, 21)
	carve(17, 18, 23, 21)
	carve(6, 5, 6, 21)
	carve(33, 5, 33, 21)
	carve(4, 3, 8, 7)
	carve(31, 3, 35, 7)
	carve(20, 5, 20, 21)
	carve(18, 2, 22, 4)
	carve(33, 18, 35, 18)
	carve(4, 8, 8, 10)
	carve(31, 8, 35, 10)
	carve(18, 13, 22, 15)
	w.Tiles[21][2] = Portal
	w.Tiles[4][6], w.Tiles[4][33] = Lever, Lever
	w.Tiles[5][20], w.Tiles[3][20] = Gate, Altar
	w.Tiles[10][20] = Gate
	for point := range w.Puzzles {
		w.Tiles[point.Y][point.X] = Tablet
	}
	return w
}

func (w World) TileAt(p Point) Tile {
	if p.Y < 0 || p.Y >= len(w.Tiles) || p.X < 0 || p.X >= len(w.Tiles[p.Y]) {
		return Wall
	}
	return w.Tiles[p.Y][p.X]
}

func (w World) CanMove(p Point) bool {
	if chest := w.Chests[p]; chest != nil && !chest.Opened {
		return false
	}
	return Walkable(byte(w.TileAt(p)))
}

func (w World) ActivateSwitch(p Point) bool {
	active, exists := w.Switches[p]
	if !exists || active {
		return false
	}
	tablet := Point{4, 8}
	if p == (Point{33, 4}) {
		tablet = Point{34, 8}
	}
	if puzzle := w.Puzzles[tablet]; puzzle == nil || !puzzle.Solved {
		return false
	}
	w.Switches[p] = true
	w.RefreshGates()
	return true
}

func (w World) RefreshGates() {
	if puzzle := w.Puzzles[Point{19, 14}]; puzzle != nil && puzzle.Solved {
		w.Tiles[10][20] = Path
	}
	if len(w.Switches) == 0 {
		return
	}
	for _, on := range w.Switches {
		if !on {
			return
		}
	}
	w.Tiles[5][20] = Path
}

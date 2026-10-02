package core

import "testing"

func reachable(w World, start Point, avoid Point) map[Point]bool {
	seen := map[Point]bool{start: true}
	queue := []Point{start}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, d := range []Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			next := Point{p.X + d.X, p.Y + d.Y}
			if !seen[next] && next != avoid && w.CanMove(next) {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return seen
}

func adjacentReachable(seen map[Point]bool, p Point) bool {
	return seen[Point{p.X - 1, p.Y}] || seen[Point{p.X + 1, p.Y}] || seen[Point{p.X, p.Y - 1}] || seen[Point{p.X, p.Y + 1}]
}

func TestWorldDimensionsAndChestMovement(t *testing.T) {
	for _, area := range []struct {
		world         World
		width, height int
	}{{NewWorld(), 21, 12}, {NewShrineWorld(), 40, 24}} {
		if len(area.world.Tiles) != area.height {
			t.Fatal("wrong height")
		}
		for y, row := range area.world.Tiles {
			if len(row) != area.width {
				t.Fatalf("row %d has width %d", y, len(row))
			}
			for x, tile := range row {
				if (x == 0 || x == area.width-1 || y == 0 || y == area.height-1) && tile != Wall {
					t.Fatalf("unsealed boundary %d,%d", x, y)
				}
			}
		}
		for p, chest := range area.world.Chests {
			if area.world.CanMove(p) {
				t.Fatal("unopened chest is walkable")
			}
			chest.Opened = true
			if area.world.TileAt(p) != Wall && !area.world.CanMove(p) {
				t.Fatal("opened chest still blocks floor")
			}
		}
		if area.world.CanMove(Point{-1, 0}) {
			t.Fatal("out of bounds is walkable")
		}
	}
}

func TestShrineRoutesAndSwitchOrders(t *testing.T) {
	left, right, altar := Point{6, 4}, Point{33, 4}, Point{20, 3}
	for _, order := range [][2]Point{{left, right}, {right, left}} {
		w := NewShrineWorld()
		seen := reachable(w, Point{3, 21}, Point{-1, -1})
		for _, p := range []Point{left, right, {35, 18}, {4, 8}, {34, 8}, {19, 14}} {
			if !adjacentReachable(seen, p) {
				t.Fatalf("cannot reach interaction at %v", p)
			}
		}
		if w.ActivateSwitch(left) || w.ActivateSwitch(right) {
			t.Fatal("unsolved wing enabled lamp")
		}
		if w.CanMove(Point{20, 10}) {
			t.Fatal("offering seal starts open")
		}
		for _, index := range []int{1, 0, 2, 1} {
			w.Puzzles[Point{4, 8}].Select(index)
		}
		for _, index := range []int{0, 2, 4, 6, 8} {
			w.Puzzles[Point{34, 8}].Select(index)
		}
		if !seen[Point{2, 21}] || adjacentReachable(seen, altar) {
			t.Fatal("portal unavailable or gate bypassable")
		}
		for _, p := range []Point{{6, 12}, {33, 12}} {
			blocked := reachable(w, Point{3, 21}, p)
			if adjacentReachable(blocked, Point{p.X, 4}) {
				t.Fatalf("encounter at %v can be bypassed", p)
			}
		}
		if !w.ActivateSwitch(order[0]) || w.ActivateSwitch(order[0]) || w.ActivateSwitch(Point{1, 1}) {
			t.Fatal("switch did not latch")
		}
		if w.CanMove(Point{20, 5}) {
			t.Fatal("one switch opened gate")
		}
		if !w.ActivateSwitch(order[1]) || !w.CanMove(Point{20, 5}) {
			t.Fatal("both switches did not open gate")
		}
		if adjacentReachable(reachable(w, Point{3, 21}, Point{-1, -1}), altar) {
			t.Fatal("offering gate can be bypassed")
		}
		for _, index := range []int{0, 0, 1} {
			w.Puzzles[Point{19, 14}].Select(index)
		}
		w.RefreshGates()
		if !adjacentReachable(reachable(w, Point{3, 21}, Point{-1, -1}), altar) {
			t.Fatal("opened altar inaccessible")
		}
	}
}

func TestShrineEchoEncountersAndFloors(t *testing.T) {
	w := NewShrineWorld()
	echoes := 0
	for p, encounter := range w.Encounters {
		if !w.CanMove(p) {
			t.Fatalf("encounter inaccessible at %v", p)
		}
		if encounter.Echo {
			echoes++
		}
	}
	if echoes != 6 {
		t.Fatalf("got %d echoes", echoes)
	}
	for _, row := range w.Tiles {
		for _, tile := range row {
			if tile == Grass {
				t.Fatal("shrine has random encounter grass")
			}
		}
	}
	for p := range w.Puzzles {
		if w.CanMove(p) {
			t.Fatal("tablet walkable")
		}
	}
}

func TestReachInteractions(t *testing.T) {
	w := NewWorld()
	seen := reachable(w, Point{2, 7}, Point{-1, -1})
	for _, p := range []Point{{2, 6}, {6, 4}, {14, 8}} {
		if !adjacentReachable(seen, p) {
			t.Fatalf("unreachable interaction %v", p)
		}
	}
	for _, p := range []Point{{12, 3}, {19, 6}, {6, 7}, {9, 7}} {
		if !seen[p] {
			t.Fatalf("unreachable destination %v", p)
		}
	}
}

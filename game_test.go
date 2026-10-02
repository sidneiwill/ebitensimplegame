package main

import (
	"encoding/binary"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"gogame/internal/core"
)

func TestSynthesizedSound(t *testing.T) {
	pcm := synthesizeSound(48000, []float64{440, 660}, .1)
	if len(pcm) != 38400 {
		t.Fatal("incorrect PCM duration")
	}
	peak := int16(0)
	for i := 0; i < len(pcm); i += 4 {
		left := int16(binary.LittleEndian.Uint16(pcm[i:]))
		right := int16(binary.LittleEndian.Uint16(pcm[i+2:]))
		if left != right || left > 3000 || left < -3000 {
			t.Fatal("PCM stereo mismatch or excessive volume")
		}
		if left > peak {
			peak = left
		}
	}
	if peak == 0 || binary.LittleEndian.Uint16(pcm) != 0 {
		t.Fatal("sound is silent or attack envelope starts abruptly")
	}
	if tail := int16(binary.LittleEndian.Uint16(pcm[len(pcm)-4:])); tail > 10 || tail < -10 {
		t.Fatal("release envelope ends abruptly")
	}
}

func TestTitleMenuFlow(t *testing.T) {
	bank := &SoundBank{Muted: true}
	g := NewGame(1, bank)
	if g.mode != ModeTitle {
		t.Fatal("new game did not open at title menu")
	}

	for _, ending := range []Mode{ModeVictory, ModeGameOver} {
		g.mode = ending
		g.returnToTitle()
		if g.mode != ModeTitle {
			t.Fatalf("ending %d did not return to title", ending)
		}
	}

	g.progress.BladeBought = true
	g.player.Level = 4
	g.worlds[core.Shrine].Switches[Point{X: 6, Y: 4}] = true
	g.startJourney()
	if g.mode != ModeExplore || g.sounds != bank || !bank.Muted || g.area != core.Reach || g.progress != (core.Progress{}) || g.player.Level != 1 || g.worlds[core.Shrine].Switches[Point{X: 6, Y: 4}] {
		t.Fatal("begin journey did not reset the run and preserve sound preference")
	}
}

func TestEscapeReturnsToTitleOutsideNestedMenus(t *testing.T) {
	g := NewGame(1, nil)
	if !g.escapeReturnsToTitle() {
		t.Fatal("escape should leave exploration for the title menu")
	}
	g.mode = ModeCombat
	g.combat = NewCombat(core.NewEnemy(EnemyKnight))
	if !g.escapeReturnsToTitle() {
		t.Fatal("escape should leave combat for the title menu")
	}
	g.combat.Phase = PhaseChooseItem
	if g.escapeReturnsToTitle() {
		t.Fatal("escape should cancel item selection before leaving combat")
	}
	for _, mode := range []Mode{ModeInventory, ModeShop, ModePuzzle} {
		g.mode = mode
		if g.escapeReturnsToTitle() {
			t.Fatalf("escape should back out of mode %d before leaving the game", mode)
		}
	}
}

func advanceMessages(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; g.combat != nil && g.combat.Phase == PhaseMessage; i++ {
		if i >= 4 {
			t.Fatal("combat message did not resolve")
		}
		g.combat.Ticks = 0
		g.updateCombat()
	}
}

func winBattle(t *testing.T, g *Game) int {
	t.Helper()
	turns := 0
	advanceMessages(t, g)
	for i := 0; g.mode == ModeCombat; i++ {
		if i >= 100 {
			t.Fatal("battle did not finish")
		}
		command := core.Fight
		if g.combat.Intent() == core.Heavy {
			command = core.Guard
		}
		if g.combat.Intent() == core.Charge && g.player.HP <= g.player.MaxHP-12 && g.inventory.Counts[Potion] > 0 {
			command = core.Item
		}
		g.commitCombat(command, Potion)
		turns++
		advanceMessages(t, g)
	}
	if g.mode != ModeExplore && g.mode != ModeVictory {
		t.Fatal("battle ended in defeat")
	}
	return turns
}

func TestFullExplorationRoute(t *testing.T) {
	g := NewGame(1, nil)
	g.startJourney()
	moves, battles, turns := 0, 0, 0
	fight := func() {
		if g.mode == ModeCombat {
			battles++
			turns += winBattle(t, g)
		}
	}
	walkTo := func(target Point) {
		t.Helper()
		start, area := g.player.Pos, g.area
		parents := map[Point]Point{start: start}
		queue := []Point{start}
		for i := 0; i < len(queue); i++ {
			pos := queue[i]
			if pos == target {
				break
			}
			for _, d := range []Point{{X: 0, Y: -1}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: -1, Y: 0}} {
				next := Point{X: pos.X + d.X, Y: pos.Y + d.Y}
				_, seen := parents[next]
				if seen || !g.world().CanMove(next) {
					continue
				}
				tile := g.world().TileAt(next)
				if (tile == Portal || tile == Keep) && next != target {
					continue
				}
				parents[next] = pos
				queue = append(queue, next)
			}
		}
		if _, found := parents[target]; !found {
			t.Fatalf("no route from %+v to %+v in area %d", start, target, area)
		}
		var path []Point
		for p := target; p != start; p = parents[p] {
			path = append(path, p)
		}
		for i := len(path) - 1; i >= 0; i-- {
			next := path[i]
			g.move(next.X-g.player.Pos.X, next.Y-g.player.Pos.Y)
			moves++
			fight()
			if g.area != area {
				if i != 0 {
					t.Fatal("portal interrupted waypoint route")
				}
				return
			}
			if g.player.Pos != next {
				t.Fatalf("movement blocked at %+v", next)
			}
		}
	}
	interact := func(dx, dy int) { g.player.FacingX, g.player.FacingY = dx, dy; g.interact() }
	puzzle := func(stand Point, dx, dy int, choices ...int) {
		walkTo(stand)
		interact(dx, dy)
		if g.mode != ModePuzzle {
			t.Fatal("route failed to open tablet")
		}
		for _, choice := range choices {
			g.menuCursor = choice
			g.selectPuzzle()
		}
		if !g.puzzle.Solved {
			t.Fatal("route failed to solve tablet")
		}
		g.mode = ModeExplore
	}
	walkTo(Point{X: 10, Y: 7})
	walkTo(Point{X: 2, Y: 7})
	interact(0, -1)
	g.menuCursor = 1
	g.buySelected()
	if !g.progress.BladeBought {
		t.Fatal("main path could not afford blade")
	}
	g.mode = ModeExplore
	walkTo(Point{X: 12, Y: 3})
	if g.area != core.Shrine {
		t.Fatal("route failed to enter shrine")
	}
	puzzle(Point{X: 5, Y: 8}, -1, 0, 1, 0, 2, 1)
	walkTo(Point{X: 6, Y: 5})
	interact(0, -1)
	puzzle(Point{X: 33, Y: 8}, 1, 0, 0, 2, 4, 6, 8)
	walkTo(Point{X: 33, Y: 5})
	interact(0, -1)
	walkTo(Point{X: 34, Y: 18})
	interact(1, 0)
	if !g.world().Chests[Point{X: 35, Y: 18}].Opened {
		t.Fatal("route missed treasure")
	}
	puzzle(Point{X: 20, Y: 14}, -1, 0, 0, 0, 1)
	walkTo(Point{X: 20, Y: 4})
	interact(0, -1)
	fight()
	if !g.progress.ShrineCleared || g.mode != ModeExplore {
		t.Fatal("route failed to earn shrine blessing")
	}
	for point, encounter := range g.world().Encounters {
		if !encounter.Cleared {
			t.Fatalf("route missed shrine encounter %+v", point)
		}
	}
	for point, p := range g.world().Puzzles {
		if !p.Solved {
			t.Fatalf("route missed puzzle %+v", point)
		}
	}
	for point, on := range g.world().Switches {
		if !on {
			t.Fatalf("route missed lamp %+v", point)
		}
	}
	walkTo(Point{X: 2, Y: 21})
	if g.area != core.Reach {
		t.Fatal("route failed to leave shrine")
	}
	walkTo(Point{X: 19, Y: 6})
	if g.mode != ModeVictory || g.player.HP <= 0 {
		t.Fatal("full exploration route did not win")
	}
	t.Logf("Full route: %d moves, %d battles, %d committed combat turns; final level %d, HP %d/%d", moves, battles, turns, g.player.Level, g.player.HP, g.player.MaxHP)
}

func TestGameFlow(t *testing.T) {
	g := NewGame(1, nil)
	g.player.FacingY = -1
	g.interact()
	if g.mode != ModeShop {
		t.Fatal("starting merchant interaction failed")
	}
	g.mode = ModeExplore
	for _, pos := range []Point{{X: 6, Y: 7}, {X: 9, Y: 7}} {
		g.player.Pos = Point{X: pos.X - 1, Y: pos.Y}
		g.move(1, 0)
		encounter := g.world().Encounters[pos]
		if g.mode != ModeCombat || g.combat.Encounter != encounter {
			t.Fatal("movement did not trigger fixed encounter")
		}
		winBattle(t, g)
		if !encounter.Cleared {
			t.Fatal("defeated fixed encounter not cleared")
		}
		g.player.Pos = Point{X: pos.X - 1, Y: pos.Y}
		g.move(1, 0)
		if g.mode != ModeExplore {
			t.Fatal("cleared encounter repeated")
		}
	}
	if g.player.Level != 2 || g.player.Gold != 12 {
		t.Fatalf("main path rewards: %+v", g.player)
	}
	g.mode, g.menuCursor = ModeShop, 1
	g.buySelected()
	if !g.progress.BladeBought || g.player.Attack != 13 || g.player.Gold != 0 {
		t.Fatal("blade purchase failed")
	}
	g.player.Gold = 12
	g.buySelected()
	if g.player.Gold != 12 || g.player.Attack != 13 {
		t.Fatal("duplicate purchase changed balances")
	}
	g.mode = ModeExplore
	g.player.HP = g.player.MaxHP
	g.beginCombat(core.NewEnemy(EnemyGuardian), nil)
	advanceMessages(t, g)
	g.combat.Phase = PhaseChooseItem
	turn, count := g.combat.Turn, g.inventory.Counts[Potion]
	g.commitCombat(core.Item, Potion)
	advanceMessages(t, g)
	if g.combat.Phase != PhaseChooseItem || g.combat.EnemyPending || g.combat.Turn != turn || g.inventory.Counts[Potion] != count {
		t.Fatal("failed item advanced combat")
	}
	// Seed 1 succeeds on the first escape roll, independent of combat actions.
	g.commitCombat(core.Run, 0)
	advanceMessages(t, g)
	if g.mode != ModeExplore || g.progress.ShrineCleared || g.player.Gold != 12 {
		t.Fatal("guardian escape gave reward or failed")
	}
	g.beginCombat(core.NewEnemy(EnemyGuardian), nil)
	winBattle(t, g)
	if !g.progress.ShrineCleared || g.mode != ModeExplore || g.player.MaxHP != 46 {
		t.Fatal("guardian victory did not bless or ended run")
	}
	g.worlds[core.Reach].Chests[Point{X: 14, Y: 8}].Opened = true
	g.changeArea()
	if g.world().ActivateSwitch(Point{X: 6, Y: 4}) || g.world().ActivateSwitch(Point{X: 33, Y: 4}) {
		t.Fatal("lamp activated before tablet solved")
	}
	for _, tc := range []struct {
		stand   Point
		dx, dy  int
		tablet  Point
		choices []int
	}{
		{Point{X: 5, Y: 8}, -1, 0, Point{X: 4, Y: 8}, []int{1, 0, 2, 1}},
		{Point{X: 33, Y: 8}, 1, 0, Point{X: 34, Y: 8}, []int{0, 2, 4, 6, 8}},
		{Point{X: 20, Y: 14}, -1, 0, Point{X: 19, Y: 14}, []int{0, 0, 1}},
	} {
		g.player.Pos, g.player.FacingX, g.player.FacingY = tc.stand, tc.dx, tc.dy
		g.interact()
		if g.mode != ModePuzzle || g.puzzle != g.world().Puzzles[tc.tablet] {
			t.Fatal("tablet interaction did not open correct puzzle")
		}
		for _, choice := range tc.choices {
			g.menuCursor = choice
			g.selectPuzzle()
		}
		if !g.puzzle.Solved {
			t.Fatal("puzzle adapter did not solve tablet")
		}
		g.mode = ModeExplore
	}
	g.world().RefreshGates()
	if g.world().TileAt(Point{X: 20, Y: 10}) != Path {
		t.Fatal("offering did not open inner gate")
	}
	g.world().ActivateSwitch(Point{X: 6, Y: 4})
	g.world().ActivateSwitch(Point{X: 33, Y: 4})
	g.world().Chests[Point{X: 35, Y: 18}].Opened = true
	g.changeArea()
	g.changeArea()
	if !g.world().Switches[Point{X: 6, Y: 4}] || !g.world().Chests[Point{X: 35, Y: 18}].Opened || !g.worlds[core.Reach].Chests[Point{X: 14, Y: 8}].Opened || !g.progress.BladeBought || !g.progress.ShrineCleared {
		t.Fatal("area revisit lost progress")
	}
	for _, puzzle := range g.world().Puzzles {
		if !puzzle.Solved {
			t.Fatal("area revisit lost puzzle solution")
		}
	}
	bank := &SoundBank{Muted: true}
	g.sounds = bank
	g.restart()
	if g.sounds != bank || !bank.Muted || g.mode != ModeExplore || g.area != core.Reach || g.progress != (core.Progress{}) || g.player.Level != 1 || g.player.MaxHP != 32 || g.player.Attack != 9 || g.inventory.Counts[Potion] != 2 || g.worlds[core.Shrine].Switches[Point{X: 6, Y: 4}] || g.worlds[core.Reach].Chests[Point{X: 14, Y: 8}].Opened {
		t.Fatal("restart failed to reset run and preserve audio")
	}
	for _, puzzle := range g.worlds[core.Shrine].Puzzles {
		if puzzle.Solved {
			t.Fatal("restart retained puzzle solution")
		}
	}
	g.player.Pos = Point{X: 12, Y: 4}
	g.move(0, -1)
	if g.area != core.Shrine || g.player.Pos != (Point{X: 3, Y: 21}) {
		t.Fatal("moving onto portal did not enter shrine")
	}
	g.move(-1, 0)
	if g.area != core.Reach || g.player.Pos != (Point{X: 12, Y: 4}) {
		t.Fatal("moving onto return portal failed")
	}
	g.player.Pos = Point{X: 18, Y: 6}
	g.move(1, 0)
	if g.mode != ModeCombat || !g.combat.Enemy.Boss {
		t.Fatal("keep entry did not start Warden")
	}
}

type screenCase struct {
	name  string
	setup func(*Game)
}
type screenshotGame struct {
	game     *Game
	cases    []screenCase
	folder   string
	index    int
	captured bool
	err      error
	bank     *SoundBank
	started  bool
	wait     int
}

func (s *screenshotGame) Layout(w, h int) (int, int) { return s.game.Layout(w, h) }
func (s *screenshotGame) Update() error {
	if s.err != nil {
		return s.err
	}
	if !s.started {
		s.started = true
		s.bank.Play(SoundConfirm)
		s.wait = 8
	}
	if s.wait > 0 {
		s.wait--
		return nil
	}
	if s.captured {
		s.bank.Stop()
		s.index++
		s.captured = false
		if s.index == len(s.cases) {
			return ebiten.Termination
		}
		s.bank.Muted = false
		s.game = NewGame(1, s.bank)
		s.game.mode = ModeExplore
		s.cases[s.index].setup(s.game)
		if s.index == len(s.cases)-1 {
			s.bank.Play(SoundVictory)
			s.wait = 60
		}
	}
	return nil
}
func (s *screenshotGame) Draw(screen *ebiten.Image) {
	s.game.Draw(screen)
	if s.captured || s.err != nil {
		return
	}
	img := image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight))
	screen.ReadPixels(img.Pix)
	f, err := os.Create(filepath.Join(s.folder, s.cases[s.index].name+".png"))
	if err != nil {
		s.err = err
		return
	}
	s.err = png.Encode(f, img)
	if err := f.Close(); s.err == nil {
		s.err = err
	}
	s.captured = true
}

func TestRenderedScreens(t *testing.T) {
	folder := os.Getenv("GOGAME_SCREENSHOTS")
	if folder == "" {
		t.Skip("set GOGAME_SCREENSHOTS to capture rendered screens")
	}
	if err := os.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	shrine := func(g *Game, p Point) {
		g.area = core.Shrine
		g.player.Pos = p
		g.message = "Solve the wing tablets to light both lamps. The inner seal awaits an offering."
	}
	solve := func(g *Game, point Point, choices ...int) {
		for _, choice := range choices {
			g.world().Puzzles[point].Select(choice)
		}
		g.world().RefreshGates()
	}
	cases := []screenCase{
		{"title-menu", func(g *Game) { g.mode = ModeTitle }},
		{"reach-start", func(g *Game) { g.mode = ModeExplore }},
		{"reach-east-edge", func(g *Game) { g.player.Pos = Point{X: 18, Y: 6} }},
		{"shrine-entry", func(g *Game) { shrine(g, Point{X: 3, Y: 21}) }},
		{"shrine-west-lamp", func(g *Game) {
			shrine(g, Point{X: 6, Y: 5})
			solve(g, Point{X: 4, Y: 8}, 1, 0, 2, 1)
			g.world().ActivateSwitch(Point{X: 6, Y: 4})
			g.message = "One lamp burns. Find the other wing."
		}},
		{"shrine-east-lamp", func(g *Game) {
			shrine(g, Point{X: 33, Y: 5})
			solve(g, Point{X: 34, Y: 8}, 0, 2, 4, 6, 8)
			g.world().ActivateSwitch(Point{X: 33, Y: 4})
		}},
		{"shrine-locked-gate", func(g *Game) {
			shrine(g, Point{X: 20, Y: 6})
			g.message = "The altar is sealed. Light both wing lamps."
		}},
		{"shrine-open-altar", func(g *Game) {
			shrine(g, Point{X: 20, Y: 4})
			solve(g, Point{X: 4, Y: 8}, 1, 0, 2, 1)
			solve(g, Point{X: 34, Y: 8}, 0, 2, 4, 6, 8)
			solve(g, Point{X: 19, Y: 14}, 0, 0, 1)
			g.world().ActivateSwitch(Point{X: 6, Y: 4})
			g.world().ActivateSwitch(Point{X: 33, Y: 4})
			g.message = "Both lamps burn. The northern altar gate opens!"
		}},
		{"puzzle-runes", func(g *Game) {
			shrine(g, Point{X: 5, Y: 8})
			g.mode = ModePuzzle
			g.puzzle = g.world().Puzzles[Point{X: 4, Y: 8}]
			g.puzzle.Select(1)
			g.puzzle.Select(0)
			g.puzzleMessage = "The cycle remembers your rune."
		}},
		{"puzzle-circuit", func(g *Game) {
			shrine(g, Point{X: 33, Y: 8})
			g.mode = ModePuzzle
			g.puzzle = g.world().Puzzles[Point{X: 34, Y: 8}]
			g.menuCursor = 4
			g.puzzle.Select(4)
			g.puzzleMessage = "The lamps shift. Light all nine."
		}},
		{"puzzle-offerings", func(g *Game) {
			shrine(g, Point{X: 20, Y: 14})
			g.mode = ModePuzzle
			g.puzzle = g.world().Puzzles[Point{X: 19, Y: 14}]
			g.puzzle.Select(0)
			g.puzzle.Select(0)
			g.puzzleMessage = "The altar weighs your offering."
		}},
		{"reach-muted", func(g *Game) { g.sounds.Muted = true }},
		{"merchant", func(g *Game) {
			g.mode = ModeShop
			g.player.Gold = 12
			g.message = "Guard blocks charged strikes. Heal on charge turns."
		}},
		{"merchant-owned", func(g *Game) {
			g.mode = ModeShop
			g.progress.BladeBought = true
			g.menuCursor = 1
			g.message = "Blade already purchased."
		}},
		{"inventory", func(g *Game) { g.mode = ModeInventory; g.inventory.Add(MoonHerb, 1) }},
		{"combat-charged", func(g *Game) {
			g.beginCombat(core.NewEnemy(EnemyKnight), nil)
			g.combat.Turn = 2
			g.combat.Phase = PhaseChoose
		}},
		{"combat-message", func(g *Game) { g.beginCombat(core.NewEnemy(EnemyWisp), nil) }},
		{"combat-items", func(g *Game) {
			g.beginCombat(core.NewEnemy(EnemyGuardian), nil)
			g.combat.Phase = PhaseChooseItem
			g.inventory.Add(MoonHerb, 1)
		}},
		{"combat-brace", func(g *Game) {
			g.beginCombat(core.NewEnemy(EnemyDragon), nil)
			g.combat.Phase = PhaseChoose
			g.combat.Braced = true
		}},
		{"combat-hit-feedback", func(g *Game) {
			g.beginCombat(core.NewEnemy(EnemyGuardian), nil)
			g.combat.Message = "Blade hits for 9."
			g.feedback = "-9"
			g.feedbackEnemy = true
			g.feedbackTicks = 40
			g.enemyHitTicks = 8
		}},
		{"combat-block-feedback", func(g *Game) {
			g.beginCombat(core.NewEnemy(EnemyKnight), nil)
			g.combat.Turn = 3
			g.combat.Message = "Charged strike blocked!"
			g.feedback = "BLOCK"
			g.feedbackTicks = 40
		}},
		{"victory", func(g *Game) { g.mode = ModeVictory }},
		{"game-over", func(g *Game) { g.mode = ModeGameOver; g.player.HP = 0 }},
	}
	bank := NewSoundBank()
	s := &screenshotGame{game: NewGame(1, bank), cases: cases, folder: folder, bank: bank}
	cases[0].setup(s.game)
	ebiten.SetWindowSize(screenWidth*3, screenHeight*3)
	ebiten.SetWindowTitle("Ashen Crown visual checks")
	if err := ebiten.RunGame(s); err != nil {
		t.Fatal(err)
	}
	if s.err != nil {
		t.Fatal(s.err)
	}
	if s.index != len(cases) {
		t.Fatalf("captured %d of %d screens", s.index, len(cases))
	}
}

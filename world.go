package main

import (
	"gogame/internal/core"

	"github.com/hajimehoshi/ebiten/v2"
)

type Tile = core.Tile

const (
	Grass    = core.Grass
	Path     = core.Path
	Water    = core.Water
	Wall     = core.Wall
	Keep     = core.Keep
	Portal   = core.Portal
	Merchant = core.Merchant
	Lever    = core.Lever
	Gate     = core.Gate
	Altar    = core.Altar
	Tablet   = core.Tablet
)

func (g *Game) updateExplore() {
	if justPressed(ebiten.KeyI, ebiten.KeyTab) {
		g.mode = ModeInventory
		g.menuCursor = 0
		g.play(SoundConfirm)
		return
	}
	if justPressed(ebiten.KeyE, ebiten.KeyZ, ebiten.KeyEnter, ebiten.KeySpace) {
		g.interact()
		return
	}
	if g.stepCooldown > 0 {
		g.stepCooldown--
		return
	}
	dx, dy := 0, 0
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW):
		dy = -1
	case ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS):
		dy = 1
	case ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA):
		dx = -1
	case ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD):
		dx = 1
	default:
		return
	}
	g.move(dx, dy)
}

func (g *Game) move(dx, dy int) {
	g.player.FacingX, g.player.FacingY = dx, dy
	next := Point{X: g.player.Pos.X + dx, Y: g.player.Pos.Y + dy}
	w := g.world()
	if !w.CanMove(next) {
		g.stepCooldown = 5
		return
	}
	g.player.Pos = next
	g.player.WalkBit ^= 1
	g.stepCooldown = 8
	tile := w.TileAt(next)
	if tile == Portal {
		g.changeArea()
		return
	}
	if tile == Keep {
		g.beginCombat(core.NewEnemy(core.EnemyKnight), nil)
		return
	}
	if encounter := w.Encounters[next]; encounter != nil && !encounter.Cleared {
		enemy := core.NewEnemy(encounter.Kind)
		if encounter.Echo {
			enemy.Name = "Echo " + enemy.Name
			enemy.HP += 8
			enemy.MaxHP += 8
			enemy.Attack++
			enemy.XP, enemy.Gold = 2, 3
		}
		g.beginCombat(enemy, encounter)
		return
	}
	if g.area == core.Reach && tile == Grass && g.rng.Intn(100) < 11 {
		g.beginCombat(randomEnemy(g.rng.Intn(2)), nil)
	}
}

func (g *Game) changeArea() {
	if g.area == core.Reach {
		g.area = core.Shrine
		g.player.Pos = Point{X: 3, Y: 21}
		g.say("Solve the wing tablets to light both lamps. The inner seal awaits an offering.")
	} else {
		g.area = core.Reach
		g.player.Pos = Point{X: 12, Y: 4}
		g.say("Verdant Reach. The keep waits to the east.")
	}
	g.player.FacingX = 0
	g.player.FacingY = 1
	g.stepCooldown = 8
	g.play(SoundConfirm)
}

func (g *Game) interact() {
	front := Point{X: g.player.Pos.X + g.player.FacingX, Y: g.player.Pos.Y + g.player.FacingY}
	w := g.world()
	if chest := w.Chests[front]; chest != nil {
		if chest.Opened {
			g.say("The coffer is empty.")
			return
		}
		chest.Opened = true
		g.inventory.Add(chest.Item, chest.Count)
		g.say("Found " + itoa(chest.Count) + " " + itemNames[chest.Item] + "!")
		g.play(SoundReward)
		return
	}
	switch w.TileAt(front) {
	case Merchant:
		g.mode = ModeShop
		g.menuCursor = 0
		g.say("Guard blocks charged strikes. Heal on charge turns.")
		g.play(SoundConfirm)
	case Portal:
		g.changeArea()
	case Lever:
		if !w.ActivateSwitch(front) {
			if w.Switches[front] {
				g.say("This lamp already burns.")
			} else {
				g.say("Solve this wing's tablet to awaken its lamp.")
			}
			return
		}
		if w.TileAt(Point{X: 20, Y: 5}) == Path {
			g.say("Both lamps burn. The northern altar gate opens!")
		} else {
			g.say("One lamp burns. Find the lamp in the other wing.")
		}
		g.play(SoundReward)
	case Gate:
		if front.Y == 10 {
			g.say("The inner seal awaits three spectral offerings.")
		} else {
			g.say("The altar is sealed. Light both wing lamps.")
		}
	case Tablet:
		g.puzzle = w.Puzzles[front]
		g.puzzleMessage = ""
		g.mode, g.menuCursor = ModePuzzle, 0
		g.play(SoundConfirm)
	case Altar:
		if g.progress.ShrineCleared {
			g.say("The blessing is yours. The altar rests.")
			return
		}
		g.beginCombat(core.NewEnemy(core.EnemyGuardian), nil)
	case Keep:
		g.say("The Cinder Warden waits. Step through the gate.")
	default:
		g.say("Only wind answers.")
	}
}

func (g *Game) updateInventory() {
	g.navigate(&g.menuCursor, len(itemNames))
	if justPressed(ebiten.KeyEscape, ebiten.KeyX, ebiten.KeyI, ebiten.KeyTab) {
		g.mode = ModeExplore
		g.play(SoundConfirm)
		return
	}
	if justPressed(ebiten.KeyEnter, ebiten.KeyZ, ebiten.KeySpace) {
		id := g.menuCursor
		if healed, ok := HealWithItem(&g.player, &g.inventory, id); ok {
			g.say(itemNames[id] + " restored " + itoa(healed) + " HP.")
			g.play(SoundHeal)
		} else if g.inventory.Counts[id] == 0 {
			g.say("None left.")
			g.play(SoundError)
		} else {
			g.say("HP already full.")
			g.play(SoundError)
		}
	}
}

func (g *Game) updateShop() {
	g.navigate(&g.menuCursor, 2)
	if justPressed(ebiten.KeyEscape, ebiten.KeyX) {
		g.mode = ModeExplore
		g.play(SoundConfirm)
		return
	}
	if justPressed(ebiten.KeyEnter, ebiten.KeyZ, ebiten.KeySpace) {
		g.buySelected()
	}
}

func (g *Game) buySelected() {
	message, ok := core.Purchase(&g.player, g.inventory.Counts, &g.progress, g.menuCursor == 1)
	g.say(message)
	if ok {
		g.play(SoundConfirm)
	} else {
		g.play(SoundError)
	}
}

func (g *Game) updatePuzzle() {
	if justPressed(ebiten.KeyEscape, ebiten.KeyX) {
		g.mode = ModeExplore
		if g.puzzle.Solved {
			switch g.puzzle.Kind {
			case core.RunePuzzle:
				g.say("Rune chamber complete. Light the western lamp.")
			case core.CircuitPuzzle:
				g.say("Circuit restored. Light the eastern lamp.")
			case core.OfferingPuzzle:
				g.say("Offering accepted. The inner gate opens.")
			}
		}
		g.play(SoundConfirm)
		return
	}
	if justPressed(ebiten.KeyR) {
		g.puzzle.Reset()
		g.puzzleMessage = "The chamber is reset."
		if g.puzzle.Solved {
			g.puzzleMessage = "This chamber is already complete."
		}
		g.play(SoundConfirm)
		return
	}
	if g.puzzle.Kind == core.CircuitPuzzle {
		before := g.menuCursor
		switch {
		case justPressed(ebiten.KeyArrowUp, ebiten.KeyW):
			g.menuCursor = (g.menuCursor + 6) % 9
		case justPressed(ebiten.KeyArrowDown, ebiten.KeyS):
			g.menuCursor = (g.menuCursor + 3) % 9
		case justPressed(ebiten.KeyArrowLeft, ebiten.KeyA):
			g.menuCursor = (g.menuCursor/3)*3 + (g.menuCursor+2)%3
		case justPressed(ebiten.KeyArrowRight, ebiten.KeyD):
			g.menuCursor = (g.menuCursor/3)*3 + (g.menuCursor+1)%3
		}
		if before != g.menuCursor {
			g.play(SoundMenu)
		}
	} else {
		g.navigate(&g.menuCursor, g.puzzle.OptionCount())
	}
	if justPressed(ebiten.KeyEnter, ebiten.KeyZ, ebiten.KeySpace) {
		g.selectPuzzle()
	}
}

func (g *Game) selectPuzzle() {
	solved := g.puzzle.Solved
	g.puzzleMessage = g.puzzle.Select(g.menuCursor)
	g.world().RefreshGates()
	if !solved && g.puzzle.Solved {
		g.play(SoundReward)
	} else if g.puzzleMessage == "The pattern fades. Try again." {
		g.play(SoundError)
	} else {
		g.play(SoundConfirm)
	}
}

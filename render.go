package main

import (
	"image/color"
	"strconv"
	"strings"

	"gogame/internal/core"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

var (
	ink   = color.RGBA{0x0c, 0x11, 0x17, 0xff}
	navy  = color.RGBA{0x17, 0x23, 0x31, 0xff}
	green = color.RGBA{0x20, 0x37, 0x31, 0xff}
	moss  = color.RGBA{0x43, 0x64, 0x48, 0xff}
	sage  = color.RGBA{0x8e, 0xad, 0x78, 0xff}
	bone  = color.RGBA{0xc5, 0xc3, 0xa4, 0xff}
	cream = color.RGBA{0xff, 0xf1, 0xc8, 0xff}
	ember = color.RGBA{0xf0, 0x7b, 0x48, 0xff}
	gold  = color.RGBA{0xe5, 0xb5, 0x55, 0xff}
	cyan  = color.RGBA{0x78, 0xc8, 0xbe, 0xff}
	blood = color.RGBA{0xb8, 0x4b, 0x49, 0xff}
)

type tileTextureKey struct {
	tile    Tile
	variant int
	frame   int
}

func itoa(n int) string { return strconv.Itoa(n) }

func rect(dst *ebiten.Image, x, y, w, h int, c color.Color) {
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), c, false)
}

func label(dst *ebiten.Image, s string, x, y int, c color.Color) {
	if c != ink {
		text.Draw(dst, s, basicfont.Face7x13, x+1, y+1, ink)
	}
	text.Draw(dst, s, basicfont.Face7x13, x, y, c)
}

func wrap(s string, maxChars int) []string {
	lines := []string{""}
	for _, word := range strings.Fields(s) {
		last := len(lines) - 1
		candidate := word
		if lines[last] != "" {
			candidate = lines[last] + " " + word
		}
		if len(candidate) <= maxChars {
			lines[last] = candidate
		} else {
			lines = append(lines, word)
		}
	}
	return lines
}

func panel(dst *ebiten.Image, x, y, w, h int) {
	rect(dst, x, y, w, h, bone)
	rect(dst, x+1, y+1, w-2, h-2, navy)
	rect(dst, x+2, y+2, w-4, h-4, ink)
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.mode {
	case ModeTitle:
		g.drawTitle(screen)
	case ModeExplore:
		g.drawWorld(screen)
	case ModeInventory:
		g.drawInventory(screen)
	case ModeShop:
		g.drawShop(screen)
	case ModePuzzle:
		g.drawPuzzle(screen)
	case ModeCombat:
		g.drawCombat(screen)
	case ModeGameOver:
		g.drawEnd(screen, "FALLEN", "The crown remains in shadow.", blood)
	case ModeVictory:
		g.drawEnd(screen, "CROWN RECLAIMED", "Dawn returns to Verdant Reach.", gold)
	}
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	g.ensureTileTextures()
	screen.Fill(ink)
	w := &g.worlds[core.Reach]
	for y, row := range w.Tiles {
		for x, tile := range row {
			g.drawTile(screen, x*tileSize-tileSize/2, y*tileSize, tile, x, y)
		}
	}
	for point, chest := range w.Chests {
		g.drawChest(screen, point.X*tileSize-tileSize/2, point.Y*tileSize, chest.Opened)
	}
	g.drawHero(screen, 2*tileSize-tileSize/2, 7*tileSize)
	rect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0x08, 0x0d, 0x12, 0xb8})

	panel(screen, 38, 27, 244, 120)
	rect(screen, 148, 41, 24, 3, gold)
	rect(screen, 148, 34, 4, 7, gold)
	rect(screen, 158, 30, 4, 11, cream)
	rect(screen, 168, 34, 4, 7, gold)
	centeredLabel(screen, "ASHEN CROWN", 67, gold)
	centeredLabel(screen, "VERDANT REACH AWAITS", 86, cream)
	panel(screen, 90, 101, 140, 22)
	centeredLabel(screen, "BEGIN JOURNEY", 117, gold)
	centeredLabel(screen, "Z / ENTER / SPACE", 139, sage)

	panel(screen, 4, 154, 312, 34)
	centeredLabel(screen, "ARROWS / WASD MOVE   I ITEMS", 169, cream)
	soundHint := "M SFX ON"
	if g.sounds != nil && g.sounds.Muted {
		soundHint = "M SFX OFF"
	}
	centeredLabel(screen, "Z / ENTER / SPACE ACT   "+soundHint, 183, sage)
}

func centeredLabel(dst *ebiten.Image, s string, y int, c color.Color) {
	label(dst, s, (screenWidth-len(s)*7)/2, y, c)
}

func (g *Game) drawWorld(screen *ebiten.Image) {
	g.ensureTileTextures()
	if g.viewport == nil {
		g.viewport = ebiten.NewImage(screenWidth, viewRows*tileSize)
	}
	w := g.world()
	cameraX := max(0, min(g.player.Pos.X-10, len(w.Tiles[0])-20))
	cameraY := max(0, min(g.player.Pos.Y-4, len(w.Tiles)-viewRows))
	g.viewport.Fill(ink)
	for y := 0; y < viewRows; y++ {
		for x := 0; x < 20; x++ {
			point := Point{X: cameraX + x, Y: cameraY + y}
			tile := w.TileAt(point)
			g.drawTile(g.viewport, x*tileSize, y*tileSize, tile, point.X, point.Y)
			if tile == Lever && w.Switches[point] {
				rect(g.viewport, x*tileSize+5, y*tileSize+3, 6, 5, cyan)
			}
			if tile == Altar && g.progress.ShrineCleared {
				rect(g.viewport, x*tileSize+5, y*tileSize+3, 6, 4, cyan)
			}
			if tile == Tablet && w.Puzzles[point].Solved {
				rect(g.viewport, x*tileSize+4, y*tileSize+3, 8, 2, cyan)
			}
			if encounter := w.Encounters[point]; encounter != nil && !encounter.Cleared {
				rect(g.viewport, x*tileSize+6, y*tileSize+3, 4, 6, ember)
				rect(g.viewport, x*tileSize+7, y*tileSize+11, 2, 2, cream)
			}
		}
	}
	for point, chest := range w.Chests {
		g.drawChest(g.viewport, (point.X-cameraX)*tileSize, (point.Y-cameraY)*tileSize, chest.Opened)
	}
	g.drawHero(g.viewport, (g.player.Pos.X-cameraX)*tileSize, (g.player.Pos.Y-cameraY)*tileSize)
	screen.Fill(ink)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, viewTop)
	screen.DrawImage(g.viewport, op)
	panel(screen, 4, 2, 312, 24)
	label(screen, "HP "+itoa(g.player.HP)+"/"+itoa(g.player.MaxHP), 10, 18, cream)
	label(screen, "LV "+itoa(g.player.Level), 111, 18, gold)
	label(screen, "G "+itoa(g.player.Gold), 166, 18, gold)
	label(screen, "ESC MENU", 202, 18, sage)
	mute := "SFX ON"
	if g.sounds != nil && g.sounds.Muted {
		mute = "SFX OFF"
	}
	label(screen, mute, 264, 18, cyan)
	message := g.message
	if g.messageTicks == 0 {
		message = "Reach | Z interact  I items  M mute"
		if g.area == core.Shrine {
			count := 0
			for _, on := range w.Switches {
				if on {
					count++
				}
			}
			message = "Shrine | Lamps " + itoa(count) + "/2 | Explore both wings"
			if count == 2 {
				message = "Lamps lit | Offering tablet opens inner seal"
				if w.TileAt(Point{X: 20, Y: 10}) == Path {
					message = "Altar open | Return stairs are southwest"
				}
			}
			if g.progress.ShrineCleared {
				message = "Shrine blessed | Return south to the Reach"
			}
		}
	}
	panel(screen, 4, 159, 312, 29)
	for i, line := range wrap(message, 42) {
		if i == 2 {
			break
		}
		label(screen, line, 10, 172+i*12, cream)
	}
}

func (g *Game) drawTile(dst *ebiten.Image, px, py int, tile Tile, tx, ty int) {
	v := (tx*17 + ty*31) % 3
	frame := 0
	if tile == Water {
		frame = (g.frame / 20) % 2
	}
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	op.GeoM.Translate(float64(px), float64(py))
	dst.DrawImage(g.tileTextures[tileTextureKey{tile: tile, variant: v, frame: frame}], op)
}

func (g *Game) ensureTileTextures() {
	if g.tileTextures != nil {
		return
	}
	g.tileTextures = make(map[tileTextureKey]*ebiten.Image)
	for _, tile := range []Tile{Grass, Path, Water, Wall, Keep, Portal, Merchant, Lever, Gate, Altar, Tablet} {
		frames := 1
		if tile == Water {
			frames = 2
		}
		for variant := 0; variant < 3; variant++ {
			for frame := 0; frame < frames; frame++ {
				texture := ebiten.NewImage(tileSize, tileSize)
				paintTileTexture(texture, tile, variant, frame)
				g.tileTextures[tileTextureKey{tile: tile, variant: variant, frame: frame}] = texture
			}
		}
	}
}

func paintTileTexture(dst *ebiten.Image, tile Tile, v, frame int) {
	switch tile {
	case Grass:
		rect(dst, 0, 0, tileSize, tileSize, green)
		rect(dst, 0, 13, tileSize, 3, color.RGBA{0x19, 0x2d, 0x2a, 0xff})
		if v != 0 {
			x := 2 + v*4
			rect(dst, x, 4, 2, 5+v, moss)
			rect(dst, x+1, 3, 1, 2, sage)
			rect(dst, 11-v, 10, 2, 3, moss)
		}
	case Path:
		rect(dst, 0, 0, tileSize, tileSize, color.RGBA{0x8d, 0x86, 0x65, 0xff})
		rect(dst, 1, 1, 14, 14, color.RGBA{0xb0, 0xa7, 0x7f, 0xff})
		rect(dst, 2+v*4, 4+v*3, 3, 1, moss)
		rect(dst, 10-v, 10, 2, 2, color.RGBA{0x8d, 0x86, 0x65, 0xff})
	case Water:
		rect(dst, 0, 0, tileSize, tileSize, navy)
		rect(dst, (frame*3+v*5)%12, 4, 6, 2, cyan)
		rect(dst, (8+v*3)%13, 11, 4, 1, sage)
		rect(dst, 0, 14, tileSize, 2, color.RGBA{0x12, 0x1c, 0x29, 0xff})
	case Wall:
		rect(dst, 0, 0, tileSize, tileSize, ink)
		rect(dst, 1, 1, 14, 14, moss)
		rect(dst, 3, 3, 10, 10, green)
		rect(dst, 0, 0, 16, 2, sage)
		rect(dst, 7, 3, 1, 10, color.RGBA{0x17, 0x2a, 0x27, 0xff})
		if v == 0 {
			rect(dst, 4, 6, 3, 1, sage)
		}
	case Keep:
		rect(dst, 0, 0, tileSize, tileSize, color.RGBA{0x75, 0x6c, 0x55, 0xff})
		rect(dst, 1, 0, 14, 16, bone)
		rect(dst, 3, 2, 10, 14, navy)
		rect(dst, 5, 6, 6, 10, ink)
		rect(dst, 5, 3, 6, 2, ember)
		rect(dst, 7, 7, 2, 4, gold)
	case Portal:
		paintTileTexture(dst, Path, v, 0)
		for step := 0; step < 4; step++ {
			rect(dst, 3, 3+step*3, 10, 2, navy)
		}
		rect(dst, 5, 3, 6, 2, cyan)
	case Merchant:
		paintTileTexture(dst, Grass, v, 0)
		rect(dst, 4, 6, 8, 8, navy)
		rect(dst, 5, 4, 6, 5, bone)
		rect(dst, 2, 3, 12, 3, gold)
		rect(dst, 4, 1, 8, 3, gold)
		rect(dst, 6, 8, 2, 1, ink)
		rect(dst, 10, 11, 3, 3, ember)
	case Lever:
		paintTileTexture(dst, Path, v, 0)
		rect(dst, 3, 11, 10, 3, navy)
		rect(dst, 6, 6, 4, 6, bone)
		rect(dst, 5, 3, 6, 5, ember)
	case Gate:
		paintTileTexture(dst, Wall, v, 0)
		rect(dst, 2, 2, 12, 12, navy)
		for x := 3; x < 14; x += 4 {
			rect(dst, x, 2, 2, 12, gold)
		}
	case Altar:
		paintTileTexture(dst, Path, v, 0)
		rect(dst, 2, 11, 12, 4, bone)
		rect(dst, 4, 8, 8, 3, gold)
		rect(dst, 5, 3, 6, 5, cyan)
		rect(dst, 7, 1, 2, 2, cream)
	case Tablet:
		paintTileTexture(dst, Path, v, 0)
		rect(dst, 2, 1, 12, 14, bone)
		rect(dst, 3, 2, 10, 12, navy)
		label(dst, "?", 5, 12, cyan)
	}
}

func (g *Game) drawHero(dst *ebiten.Image, x, y int) {
	rect(dst, x+2, y+14, 12, 2, ink)
	if g.player.WalkBit == 0 {
		rect(dst, x+4, y+12, 3, 3, ink)
		rect(dst, x+9, y+12, 3, 3, ink)
	} else {
		rect(dst, x+3, y+12, 3, 3, ink)
		rect(dst, x+10, y+12, 3, 3, ink)
	}
	rect(dst, x+3, y+5, 10, 8, navy)
	rect(dst, x+4, y+3, 8, 5, sage)
	rect(dst, x+5, y+4, 6, 4, bone)
	rect(dst, x+5, y+6, 2, 1, ink)
	rect(dst, x+9, y+6, 2, 1, ink)
	rect(dst, x+2, y+7, 3, 3, ember)
	rect(dst, x+11, y+9, 3, 2, gold)
}

func (g *Game) drawChest(dst *ebiten.Image, x, y int, opened bool) {
	rect(dst, x+2, y+14, 12, 1, ink)
	rect(dst, x+3, y+7, 10, 7, gold)
	rect(dst, x+4, y+9, 8, 4, navy)
	rect(dst, x+7, y+9, 2, 3, ember)
	if opened {
		rect(dst, x+3, y+3, 10, 3, gold)
		rect(dst, x+4, y+2, 8, 1, cream)
	} else {
		rect(dst, x+3, y+4, 10, 4, bone)
		rect(dst, x+4, y+5, 8, 2, gold)
	}
}

func (g *Game) drawCombat(screen *ebiten.Image) {
	c := g.combat
	screen.Fill(navy)
	rect(screen, 0, 62, screenWidth, 58, green)
	for x := 0; x < screenWidth; x += 16 {
		rect(screen, x, 88+(x/16%3)*4, 12, 30, moss)
	}
	label(screen, c.Enemy.Name, 178, 18, cream)
	label(screen, "HP "+itoa(g.player.HP)+"/"+itoa(g.player.MaxHP)+" LV "+itoa(g.player.Level), 10, 18, cream)
	rect(screen, 178, 25, 128, 5, ink)
	rect(screen, 178, 25, 128*c.Enemy.HP/max(1, c.Enemy.MaxHP), 5, ember)
	label(screen, "NEXT: "+c.Intent().String(), 10, 42, gold)
	hint := "Choose your command."
	if c.Intent() == core.Heavy {
		hint = "GUARD blocks the charged strike."
	}
	if c.Intent() == core.Charge {
		hint = "Safe turn: attack or heal."
	}
	if c.Braced {
		hint = "BRACED: enemy defense +3 this turn."
	}
	label(screen, hint, 10, 57, cream)
	g.drawHero(screen, 45, 88)
	g.drawEnemy(screen, c.Enemy.Kind, 222, 70)
	if g.enemyHitTicks > 0 {
		rect(screen, 219, 69, 44, 2, cream)
		rect(screen, 219, 121, 44, 2, cream)
		rect(screen, 219, 69, 2, 54, cream)
		rect(screen, 261, 69, 2, 54, cream)
	}
	if g.playerHitTicks > 0 {
		rect(screen, 44, 87, 18, 18, blood)
	}
	if g.feedbackTicks > 0 {
		x := 32
		y := 79
		if g.feedbackEnemy {
			x = 223
			y = 69
		}
		feedbackColor := blood
		if g.feedback == "BLOCK" || strings.HasPrefix(g.feedback, "+") {
			feedbackColor = cyan
		}
		label(screen, g.feedback, x, y, feedbackColor)
	}
	panel(screen, 4, 124, 312, 64)
	if c.Phase == PhaseChooseItem {
		label(screen, "ITEMS", 12, 140, gold)
		for i, name := range itemNames {
			mark := "  "
			if i == c.ItemCursor {
				mark = "> "
			}
			label(screen, mark+name+" x"+itoa(g.inventory.Counts[i]), 12, 155+i*13, cream)
		}
		label(screen, "Heal "+itoa(itemHealing[c.ItemCursor])+" HP", 185, 146, cyan)
		label(screen, "Z use / X back", 185, 169, bone)
	} else {
		commands := []string{"FIGHT", "ITEM", "GUARD", "RUN"}
		for i, command := range commands {
			y := 141 + i*11
			if c.Phase == PhaseChoose && i == g.menuCursor {
				rect(screen, 11, y-9, 76, 10, gold)
				label(screen, ">", 14, y, ink)
				label(screen, command, 25, y, ink)
			} else {
				label(screen, command, 25, y, bone)
			}
		}
		label(screen, "LV "+itoa(g.player.Level), 112, 143, gold)
		hpColor := cream
		if g.playerHitTicks > 0 {
			hpColor = blood
		}
		label(screen, "HP "+itoa(g.player.HP)+"/"+itoa(g.player.MaxHP), 112, 156, hpColor)
		label(screen, "POTION x"+itoa(g.inventory.Counts[Potion]), 112, 169, cyan)
	}
	if c.Phase == PhaseMessage {
		panel(screen, 105, 128, 205, 54)
		for i, line := range wrap(c.Message, 27) {
			if i == 2 {
				break
			}
			label(screen, line, 112, 147+i*13, cream)
		}
		label(screen, "Z: continue", 210, 175, sage)
	}
}

func (g *Game) drawEnemy(dst *ebiten.Image, kind EnemyKind, x, y int) {
	switch kind {
	case EnemyDragon:
		rect(dst, x+8, y+34, 22, 13, green)
		rect(dst, x+4, y+27, 30, 11, moss)
		rect(dst, x+8, y+18, 22, 13, moss)
		rect(dst, x+14, y+13, 10, 8, sage)
		rect(dst, x+10, y+29, 4, 4, cream)
		rect(dst, x+23, y+29, 4, 4, cream)
		rect(dst, x+11, y+30, 2, 3, ink)
		rect(dst, x+24, y+30, 2, 3, ink)
		rect(dst, x+9, y+22, 6, 2, sage)
	case EnemyWisp:
		rect(dst, x+17, y+8, 5, 8, ember)
		rect(dst, x+12, y+14, 16, 24, ember)
		rect(dst, x+8, y+20, 24, 13, gold)
		rect(dst, x+12, y+18, 16, 14, cream)
		rect(dst, x+16, y+22, 3, 4, navy)
		rect(dst, x+23, y+22, 3, 4, navy)
		rect(dst, x+15, y+29, 10, 2, ember)
	case EnemyKnight, EnemyGuardian:
		rect(dst, x+8, y+25, 22, 22, navy)
		rect(dst, x+4, y+33, 30, 12, ink)
		rect(dst, x+10, y+5, 20, 22, bone)
		rect(dst, x+12, y+10, 16, 11, navy)
		rect(dst, x+13, y+15, 4, 3, blood)
		rect(dst, x+23, y+15, 4, 3, blood)
		rect(dst, x+11, y+25, 16, 4, ember)
		rect(dst, x+35, y+12, 3, 35, cyan)
		rect(dst, x+33, y+8, 7, 5, bone)
	}
	if kind == EnemyGuardian {
		rect(dst, x+11, y+25, 16, 4, cyan)
		rect(dst, x+13, y+15, 4, 3, cyan)
		rect(dst, x+23, y+15, 4, 3, cyan)
	}
	rect(dst, x, y+48, 38, 3, ink)
}

func (g *Game) drawInventory(screen *ebiten.Image) {
	screen.Fill(ink)
	panel(screen, 4, 4, 312, 184)
	label(screen, "FIELD LEDGER", 12, 20, gold)
	label(screen, "HP "+itoa(g.player.HP)+"/"+itoa(g.player.MaxHP)+" G "+itoa(g.player.Gold), 166, 20, cream)
	rect(screen, 10, 27, 300, 1, moss)
	items := []int{Potion, MoonHerb}
	for i, id := range items {
		y := 48 + i*24
		if i == g.menuCursor {
			rect(screen, 10, y-13, 182, 20, gold)
			label(screen, "> "+itemNames[id], 16, y, ink)
		} else {
			label(screen, "  "+itemNames[id], 16, y, cream)
		}
		label(screen, "x"+itoa(g.inventory.Counts[id]), 158, y, cyan)
	}
	rect(screen, 199, 28, 1, 135, moss)
	selected := items[g.menuCursor]
	label(screen, itemNames[selected], 210, 49, gold)
	desc := "Restores 12 HP."
	if selected == MoonHerb {
		desc = "Restores 25 HP."
	}
	label(screen, desc, 210, 66, cream)
	label(screen, "Z  USE", 210, 129, ember)
	label(screen, "X  BACK", 210, 143, moss)
	label(screen, "ATK "+itoa(g.player.Attack)+" DEF "+itoa(g.player.Defense), 12, 145, bone)
	if g.messageTicks > 0 {
		for i, line := range wrap(g.message, 42) {
			if i == 2 {
				break
			}
			label(screen, line, 12, 165+i*12, cream)
		}
	} else {
		label(screen, "Choose an item.", 12, 178, sage)
	}
}

func (g *Game) drawShop(screen *ebiten.Image) {
	screen.Fill(ink)
	panel(screen, 4, 4, 312, 184)
	label(screen, "TRAVELLING MERCHANT", 12, 23, gold)
	label(screen, "GOLD "+itoa(g.player.Gold), 218, 23, cream)
	rect(screen, 10, 30, 300, 1, moss)
	rows := []string{"Potion            4g", "Tempered blade   12g"}
	if g.progress.BladeBought {
		rows[1] = "Tempered blade  OWNED"
	}
	for i, row := range rows {
		y := 57 + i*26
		if i == g.menuCursor {
			rect(screen, 10, y-15, 300, 22, gold)
			label(screen, "> "+row, 16, y, ink)
		} else {
			label(screen, "  "+row, 16, y, bone)
		}
	}
	description := "Restores 12 HP. Potions x" + itoa(g.inventory.Counts[Potion])
	if g.menuCursor == 1 {
		description = "Permanent +2 Attack. One per run."
	}
	label(screen, description, 12, 112, cream)
	label(screen, "Z buy / X back / M mute", 12, 136, cyan)
	for i, line := range wrap(g.message, 42) {
		if i == 2 {
			break
		}
		label(screen, line, 12, 162+i*13, cream)
	}
}

func (g *Game) drawPuzzle(screen *ebiten.Image) {
	p := g.puzzle
	screen.Fill(ink)
	panel(screen, 4, 4, 312, 184)
	label(screen, p.Name(), 12, 23, gold)
	if p.Solved {
		label(screen, "SOLVED", 252, 23, cyan)
	}
	for i, line := range wrap(p.Clue(), 42) {
		label(screen, line, 12, 43+i*13, cream)
	}
	status := ""
	if p.Kind == core.CircuitPuzzle {
		count := 0
		for i := 0; i < 9; i++ {
			x, y := 42+(i%3)*80, 72+(i/3)*24
			border := moss
			if i == g.menuCursor {
				border = cyan
			}
			rect(screen, x-2, y-2, 66, 23, border)
			fill, textColor, state := navy, cream, "OFF"
			if p.Lamps&(1<<i) != 0 {
				fill = gold
				textColor = ink
				state = "ON"
				count++
			}
			rect(screen, x, y, 62, 19, fill)
			label(screen, itoa(i+1)+": "+state, x+5, y+13, textColor)
			if i == g.menuCursor {
				label(screen, ">", x-10, y+13, cyan)
			}
		}
		status = "Lamps " + itoa(count) + "/9"
	} else {
		options := []string{"SUN", "MOON", "STAR"}
		if p.Kind == core.OfferingPuzzle {
			options = []string{"EMBER - weight 2", "MOON - weight 3", "STAR - weight 5"}
		}
		for i, option := range options {
			y := 87 + i*21
			if i == g.menuCursor {
				rect(screen, 10, y-14, 300, 18, gold)
				label(screen, "> "+option, 16, y, ink)
			} else {
				label(screen, "  "+option, 16, y, bone)
			}
		}
		if p.Kind == core.RunePuzzle {
			names := []string{}
			for _, choice := range p.Choices {
				names = append(names, options[choice])
			}
			status = "Cycle " + itoa(p.Count) + "/4: " + strings.Join(names, " ")
		} else {
			status = "Spectral offerings " + itoa(p.Count) + "/3 | Weight " + itoa(p.Sum) + "/7"
		}
	}
	label(screen, status, 12, 151, cyan)
	label(screen, g.puzzleMessage, 12, 165, cream)
	label(screen, "Z choose  R reset  X back", 12, 182, sage)
}

func (g *Game) drawEnd(screen *ebiten.Image, title, subtitle string, accent color.Color) {
	screen.Fill(ink)
	for y := 0; y < screenHeight; y += 8 {
		for x := (y / 8) % 2 * 8; x < screenWidth; x += 16 {
			rect(screen, x, y, 2, 2, green)
		}
	}
	panel(screen, 46, 56, 228, 80)
	label(screen, title, 160-len(title)*3, 82, accent)
	label(screen, subtitle, 160-len(subtitle)*3, 102, cream)
	centeredLabel(screen, "Z / ENTER / SPACE  TITLE MENU", 124, sage)
}

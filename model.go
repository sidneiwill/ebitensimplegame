package main

import (
	"math/rand"

	"gogame/internal/core"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 320
	screenHeight = 192
	tileSize     = 16
	viewRows     = 8
	viewTop      = 28
)

type Mode uint8

const (
	ModeExplore Mode = iota
	ModeCombat
	ModeInventory
	ModeShop
	ModePuzzle
	ModeGameOver
	ModeVictory
	ModeTitle
)

type Point = core.Point
type Player = core.Player

const (
	Potion   = core.Potion
	MoonHerb = core.MoonHerb
)

var itemNames = core.ItemNames
var itemHealing = core.ItemHealing

type Inventory struct{ Counts map[int]int }

func NewInventory() Inventory { return Inventory{Counts: map[int]int{Potion: 2}} }
func (i *Inventory) Add(id int, count int) {
	if count > 0 {
		i.Counts[id] += count
	}
}
func HealWithItem(p *Player, inv *Inventory, id int) (int, bool) {
	if id < 0 || id >= len(itemHealing) {
		return 0, false
	}
	hp, healed, ok := core.Heal(p.HP, p.MaxHP, itemHealing[id], inv.Counts, id)
	p.HP = hp
	return healed, ok
}

type Game struct {
	mode                                          Mode
	worlds                                        [2]core.World
	area                                          core.AreaID
	player                                        Player
	inventory                                     Inventory
	progress                                      core.Progress
	puzzle                                        *core.Puzzle
	puzzleMessage                                 string
	combat                                        *Combat
	rng                                           *rand.Rand
	sounds                                        *SoundBank
	message                                       string
	messageTicks, menuCursor, stepCooldown, frame int
	feedback                                      string
	feedbackEnemy                                 bool
	feedbackTicks, enemyHitTicks, playerHitTicks  int
	tileTextures                                  map[tileTextureKey]*ebiten.Image
	viewport                                      *ebiten.Image
}

func NewGame(seed int64, sounds *SoundBank) *Game {
	return &Game{
		mode: ModeTitle, worlds: [2]core.World{core.NewWorld(), core.NewShrineWorld()},
		player:    Player{Pos: Point{X: 2, Y: 7}, HP: 32, MaxHP: 32, Attack: 9, Defense: 3, Level: 1, FacingY: 1},
		inventory: NewInventory(), rng: rand.New(rand.NewSource(seed)), sounds: sounds,
		message: "Reach the keep. Explore the shrine.", messageTicks: 240,
	}
}
func (g *Game) world() *core.World         { return &g.worlds[g.area] }
func (g *Game) Layout(_, _ int) (int, int) { return screenWidth, screenHeight }
func justPressed(keys ...ebiten.Key) bool {
	for _, key := range keys {
		if inpututil.IsKeyJustPressed(key) {
			return true
		}
	}
	return false
}
func (g *Game) play(id SoundID) {
	if g.sounds != nil {
		g.sounds.Play(id)
	}
}
func (g *Game) navigate(cursor *int, count int) {
	if justPressed(ebiten.KeyArrowUp, ebiten.KeyW) {
		*cursor = (*cursor + count - 1) % count
		g.play(SoundMenu)
	}
	if justPressed(ebiten.KeyArrowDown, ebiten.KeyS) {
		*cursor = (*cursor + 1) % count
		g.play(SoundMenu)
	}
}
func (g *Game) restart() {
	sounds := g.sounds
	if sounds != nil {
		sounds.Stop()
	}
	*g = *NewGame(g.rng.Int63(), sounds)
	g.mode = ModeExplore
}
func (g *Game) startJourney() {
	g.restart()
	g.play(SoundConfirm)
}
func (g *Game) returnToTitle() {
	g.mode = ModeTitle
	g.menuCursor = 0
	g.play(SoundConfirm)
}
func (g *Game) escapeReturnsToTitle() bool {
	switch g.mode {
	case ModeExplore, ModeGameOver, ModeVictory:
		return true
	case ModeCombat:
		return g.combat == nil || g.combat.Phase != PhaseChooseItem
	default:
		return false
	}
}
func (g *Game) Update() error {
	g.frame++
	for _, ticks := range []*int{&g.messageTicks, &g.feedbackTicks, &g.enemyHitTicks, &g.playerHitTicks} {
		if *ticks > 0 {
			*ticks--
		}
	}
	if justPressed(ebiten.KeyEscape) && g.escapeReturnsToTitle() {
		g.returnToTitle()
		return nil
	}
	if g.sounds != nil && justPressed(ebiten.KeyM) {
		g.sounds.Muted = !g.sounds.Muted
		if g.sounds.Muted {
			g.sounds.Stop()
		} else {
			g.play(SoundConfirm)
		}
	}
	switch g.mode {
	case ModeTitle:
		if justPressed(ebiten.KeyZ, ebiten.KeyEnter, ebiten.KeySpace) {
			g.startJourney()
		}
	case ModeExplore:
		g.updateExplore()
	case ModeInventory:
		g.updateInventory()
	case ModeShop:
		g.updateShop()
	case ModePuzzle:
		g.updatePuzzle()
	case ModeCombat:
		g.updateCombat()
	case ModeGameOver, ModeVictory:
		if justPressed(ebiten.KeyZ, ebiten.KeyEnter, ebiten.KeySpace) {
			g.returnToTitle()
		}
	}
	return nil
}
func (g *Game) say(message string) { g.message = message; g.messageTicks = 240 }
func (g *Game) beginCombat(enemy Enemy, encounter *core.Encounter) {
	g.combat = NewCombat(enemy)
	g.combat.Encounter = encounter
	g.menuCursor = 0
	g.mode = ModeCombat
	g.feedbackTicks = 0
	g.enemyHitTicks = 0
	g.playerHitTicks = 0
	if enemy.Kind == EnemyWisp {
		g.combat.Message = "Wisp is charged. GUARD blocks it!"
	}
}

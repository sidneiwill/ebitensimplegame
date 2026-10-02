package main

import (
	"gogame/internal/core"

	"github.com/hajimehoshi/ebiten/v2"
)

type EnemyKind = core.EnemyKind

const (
	EnemyDragon   = core.EnemyDragon
	EnemyWisp     = core.EnemyWisp
	EnemyKnight   = core.EnemyKnight
	EnemyGuardian = core.EnemyGuardian
)

type Enemy = core.Enemy
type CombatPhase uint8

const (
	PhaseChoose CombatPhase = iota
	PhaseChooseItem
	PhaseMessage
)

type Combat struct {
	core.Battle
	Phase, ReturnPhase CombatPhase
	Message            string
	Ticks, ItemCursor  int
	EnemyPending       bool
	Encounter          *core.Encounter
}

func NewCombat(enemy Enemy) *Combat {
	return &Combat{Battle: core.Battle{Enemy: enemy}, Phase: PhaseMessage, Message: enemy.Name + " blocks the path!", Ticks: 55}
}
func randomEnemy(n int) Enemy { return core.NewEnemy(core.EnemyKind(n)) }

func (g *Game) updateCombat() {
	c := g.combat
	if c.Phase == PhaseMessage {
		if c.Ticks > 0 {
			c.Ticks--
			if !justPressed(ebiten.KeyEnter, ebiten.KeyZ, ebiten.KeySpace) {
				return
			}
		}
		if c.Finished {
			g.finishCombat()
			return
		}
		if c.EnemyPending {
			c.EnemyPending = false
			intent, guarding := c.Intent(), c.Guarding
			message, damage := c.Respond(&g.player)
			c.Message = message
			c.Ticks = 45
			c.ReturnPhase = PhaseChoose
			if damage > 0 {
				g.playerHitTicks = 8
				g.feedback = "-" + itoa(damage)
				g.feedbackEnemy = false
				g.feedbackTicks = 40
				g.play(SoundHit)
			} else if intent == core.Heavy && guarding {
				g.feedback = "BLOCK"
				g.feedbackEnemy = false
				g.feedbackTicks = 40
				g.play(SoundBlock)
			} else {
				g.play(SoundConfirm)
			}
			return
		}
		c.Phase = c.ReturnPhase
		return
	}
	if c.Phase == PhaseChooseItem {
		g.navigate(&c.ItemCursor, len(itemNames))
		if justPressed(ebiten.KeyEscape, ebiten.KeyX) {
			c.Phase = PhaseChoose
			g.play(SoundConfirm)
			return
		}
		if justPressed(ebiten.KeyEnter, ebiten.KeyZ, ebiten.KeySpace) {
			g.commitCombat(core.Item, c.ItemCursor)
		}
		return
	}
	g.navigate(&g.menuCursor, 4)
	if !justPressed(ebiten.KeyEnter, ebiten.KeyZ, ebiten.KeySpace) {
		return
	}
	if g.menuCursor == int(core.Item) {
		c.Phase = PhaseChooseItem
		c.ItemCursor = 0
		g.play(SoundConfirm)
		return
	}
	g.commitCombat(core.Command(g.menuCursor), 0)
}

func (g *Game) commitCombat(command core.Command, item int) {
	c := g.combat
	hp, enemyHP := g.player.HP, c.Enemy.HP
	escape := false
	if command == core.Run && !c.Enemy.Boss {
		escape = g.rng.Intn(100) >= 40
	}
	message, committed := c.Commit(&g.player, g.inventory.Counts, command, item, escape)
	c.Message = message
	c.Phase = PhaseMessage
	c.Ticks = 40
	c.ReturnPhase = PhaseChoose
	if !committed {
		if command == core.Item {
			c.ReturnPhase = PhaseChooseItem
		}
		g.play(SoundError)
		return
	}
	c.EnemyPending = !c.Finished
	switch command {
	case core.Fight:
		g.enemyHitTicks = 8
		g.feedback = "-" + itoa(enemyHP-c.Enemy.HP)
		g.feedbackEnemy = true
		g.feedbackTicks = 40
		g.play(SoundHit)
	case core.Item:
		g.feedback = "+" + itoa(g.player.HP-hp)
		g.feedbackEnemy = false
		g.feedbackTicks = 40
		g.play(SoundHeal)
	default:
		g.play(SoundConfirm)
	}
}

func (g *Game) finishCombat() {
	c := g.combat
	if g.player.HP == 0 {
		g.mode = ModeGameOver
		g.combat = nil
		g.play(SoundError)
		return
	}
	if c.Award(&g.player, &g.progress) {
		if c.Encounter != nil {
			c.Encounter.Cleared = true
		}
		if c.Enemy.Boss {
			g.mode = ModeVictory
			g.combat = nil
			g.play(SoundVictory)
			return
		}
		if c.Enemy.Kind == EnemyGuardian {
			g.say("Shrine blessing: +8 max HP. Restored!")
		} else {
			g.say("Won " + itoa(c.Enemy.XP) + " XP and " + itoa(c.Enemy.Gold) + " gold.")
		}
		g.play(SoundReward)
	}
	g.mode = ModeExplore
	g.combat = nil
	g.stepCooldown = 8
}

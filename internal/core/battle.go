package core

import "fmt"

type Point struct{ X, Y int }

type Player struct {
	Pos                        Point
	HP, MaxHP, Attack, Defense int
	Level, XP, Gold            int
	FacingX, FacingY, WalkBit  int
}

const (
	Potion = iota
	MoonHerb
)

var ItemNames = [2]string{"Potion", "Moon herb"}
var ItemHealing = [2]int{12, 25}

type EnemyKind uint8

const (
	EnemyDragon EnemyKind = iota
	EnemyWisp
	EnemyKnight
	EnemyGuardian
)

type Enemy struct {
	Name                                 string
	HP, MaxHP, Attack, Defense, XP, Gold int
	Kind                                 EnemyKind
	Boss                                 bool
}

func NewEnemy(kind EnemyKind) Enemy {
	switch kind {
	case EnemyDragon:
		return Enemy{"Snap Dragon", 22, 22, 6, 1, 8, 4, kind, false}
	case EnemyWisp:
		return Enemy{"Ember Wisp", 20, 20, 8, 0, 10, 6, kind, false}
	case EnemyGuardian:
		return Enemy{"Shrine Guardian", 48, 48, 9, 2, 16, 8, kind, false}
	default:
		return Enemy{"Cinder Warden", 80, 80, 11, 3, 30, 25, EnemyKnight, true}
	}
}

type Action uint8

const (
	Strike Action = iota
	Charge
	Heavy
	Brace
)

func (a Action) String() string {
	return [...]string{"STRIKE", "CHARGE (safe turn)", "CHARGED STRIKE", "BRACE"}[a]
}

type Command uint8

const (
	Fight Command = iota
	Item
	Guard
	Run
)

type Battle struct {
	Enemy                                        Enemy
	Turn                                         int
	Guarding, Braced, Finished, Escaped, Awarded bool
}

func (b *Battle) Intent() Action {
	switch b.Enemy.Kind {
	case EnemyDragon:
		if b.Turn%2 == 1 {
			return Brace
		}
		return Strike
	case EnemyWisp:
		if b.Turn%2 == 1 {
			return Charge
		}
		return Heavy
	default:
		return [...]Action{Strike, Charge, Heavy}[b.Turn%3]
	}
}

// Commit changes the player turn only when an action succeeds.
func (b *Battle) Commit(p *Player, counts map[int]int, command Command, item int, escape bool) (string, bool) {
	if b.Finished || p.HP <= 0 {
		return "Battle is over.", false
	}
	var message string
	switch command {
	case Fight:
		defense := b.Enemy.Defense
		if b.Braced {
			defense += 3
		}
		damage := Damage(p.Attack, defense)
		b.Enemy.HP = max(0, b.Enemy.HP-damage)
		message = fmt.Sprintf("Blade hits for %d.", damage)
		if b.Enemy.HP == 0 {
			b.Finished = true
			message = b.Enemy.Name + " falls!"
		}
	case Item:
		if item < 0 || item >= len(ItemHealing) {
			return "Unknown item.", false
		}
		if counts[item] <= 0 {
			return "None left.", false
		}
		hp, healed, ok := Heal(p.HP, p.MaxHP, ItemHealing[item], counts, item)
		if !ok {
			return "HP already full.", false
		}
		p.HP = hp
		message = fmt.Sprintf("%s restores %d HP.", ItemNames[item], healed)
	case Guard:
		b.Guarding = true
		message = "Guard blocks charged attacks."
	case Run:
		if !b.Enemy.Boss && escape {
			b.Finished, b.Escaped = true, true
			message = "You escaped."
		} else {
			message = "Escape failed."
		}
	default:
		return "Unknown command.", false
	}
	b.Braced = false
	return message, true
}

func (b *Battle) Respond(p *Player) (string, int) {
	if b.Finished || p.HP <= 0 {
		return "", 0
	}
	action := b.Intent()
	b.Turn++
	guarding := b.Guarding
	b.Guarding = false
	switch action {
	case Charge:
		return b.Enemy.Name + " gathers power.", 0
	case Brace:
		b.Braced = true
		return b.Enemy.Name + " braces (+3 defense).", 0
	}
	attack := b.Enemy.Attack
	if action == Heavy {
		switch b.Enemy.Kind {
		case EnemyWisp:
			attack += 4
		case EnemyGuardian:
			attack += 6
		default:
			attack += 8
		}
		if guarding {
			return "Charged strike blocked!", 0
		}
	}
	damage := Damage(attack, p.Defense)
	if guarding {
		damage = max(1, (damage+1)/2)
	}
	p.HP = max(0, p.HP-damage)
	if p.HP == 0 {
		b.Finished = true
	}
	return fmt.Sprintf("%s strikes for %d.", b.Enemy.Name, damage), damage
}

type Progress struct{ BladeBought, ShrineCleared bool }

func Purchase(p *Player, counts map[int]int, progress *Progress, blade bool) (string, bool) {
	cost := 4
	if blade {
		cost = 12
		if progress.BladeBought {
			return "Blade already purchased.", false
		}
	}
	if p.Gold < cost {
		return "Not enough gold.", false
	}
	p.Gold -= cost
	if blade {
		progress.BladeBought = true
		p.Attack += 2
		return "Tempered blade: +2 attack.", true
	}
	counts[Potion]++
	return "Bought a Potion.", true
}

func (b *Battle) Award(p *Player, progress *Progress) bool {
	if b.Awarded || !b.Finished || b.Escaped || b.Enemy.HP != 0 || p.HP <= 0 {
		return false
	}
	b.Awarded = true
	if b.Enemy.Kind == EnemyGuardian && progress.ShrineCleared {
		return false
	}
	p.XP += b.Enemy.XP
	p.Gold += b.Enemy.Gold
	for p.XP >= p.Level*20 {
		p.Level++
		p.MaxHP += 6
		p.Attack += 2
		p.HP = p.MaxHP
	}
	if b.Enemy.Kind == EnemyGuardian {
		progress.ShrineCleared = true
		p.MaxHP += 8
		p.HP = p.MaxHP
	}
	return true
}

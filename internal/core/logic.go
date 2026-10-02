package core

func Damage(attack, defense int) int {
	return max(1, attack-defense)
}

func Consume(counts map[int]int, id int) bool {
	if counts[id] <= 0 {
		return false
	}
	counts[id]--
	return true
}

func Heal(hp, maxHP, amount int, counts map[int]int, itemID int) (newHP, healed int, ok bool) {
	if hp >= maxHP || !Consume(counts, itemID) {
		return hp, 0, false
	}
	newHP = min(maxHP, hp+amount)
	return newHP, newHP - hp, true
}

func Walkable(tile byte) bool {
	return tile == byte(Grass) || tile == byte(Path) || tile == byte(Keep) || tile == byte(Portal)
}

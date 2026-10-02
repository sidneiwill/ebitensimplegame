package core

import "testing"

func TestDamageHasMinimum(t *testing.T) {
	if got := Damage(2, 99); got != 1 {
		t.Fatalf("Damage(2, 99) = %d, want 1", got)
	}
	if got := Damage(10, 3); got != 7 {
		t.Fatalf("Damage(10, 3) = %d, want 7", got)
	}
}

func TestConsumeNeverGoesNegative(t *testing.T) {
	counts := map[int]int{}
	if Consume(counts, 0) {
		t.Fatal("empty inventory consumed item")
	}
	if counts[0] != 0 {
		t.Fatalf("count = %d, want 0", counts[0])
	}
}

func TestHealingCapsAtMaxHP(t *testing.T) {
	counts := map[int]int{0: 1}
	hp, healed, ok := Heal(29, 32, 12, counts, 0)
	if !ok || healed != 3 || hp != 32 {
		t.Fatalf("heal result = hp %d, healed %d, ok %v", hp, healed, ok)
	}
	if counts[0] != 0 {
		t.Fatalf("count = %d, want 0", counts[0])
	}
}

func TestFullHealthDoesNotConsumeItem(t *testing.T) {
	counts := map[int]int{0: 1}
	if _, _, ok := Heal(32, 32, 12, counts, 0); ok {
		t.Fatal("full-health player used item")
	}
	if counts[0] != 1 {
		t.Fatalf("count = %d, want 1", counts[0])
	}
}

func TestWalkable(t *testing.T) {
	for _, tile := range []byte{'#', '~'} {
		if Walkable(tile) {
			t.Fatalf("tile %q should block movement", tile)
		}
	}
	for _, tile := range []byte{'.', ',', 'K'} {
		if !Walkable(tile) {
			t.Fatalf("tile %q should allow movement", tile)
		}
	}
}

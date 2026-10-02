package core

import "testing"

func TestEnemyPatternsAndGuard(t *testing.T) {
	for _, tc := range []struct {
		kind    EnemyKind
		pattern []Action
		bonus   int
	}{
		{EnemyDragon, []Action{Strike, Brace}, 0}, {EnemyWisp, []Action{Heavy, Charge}, 4},
		{EnemyGuardian, []Action{Strike, Charge, Heavy}, 6}, {EnemyKnight, []Action{Strike, Charge, Heavy}, 8},
	} {
		b := Battle{Enemy: NewEnemy(tc.kind)}
		p := Player{HP: 1000, MaxHP: 1000, Defense: 3}
		for i := 0; i < len(tc.pattern)*2; i++ {
			want := tc.pattern[i%len(tc.pattern)]
			if b.Intent() != want {
				t.Fatalf("kind %d turn %d: intent %v", tc.kind, i, b.Intent())
			}
			before := p.HP
			if _, ok := b.Commit(&p, nil, Guard, 0, false); !ok {
				t.Fatal("guard rejected")
			}
			_, damage := b.Respond(&p)
			expected := 0
			if want == Strike {
				expected = max(1, (Damage(b.Enemy.Attack, p.Defense)+1)/2)
			}
			if damage != expected || p.HP != before-expected || p.Defense != 3 || b.Guarding || b.Turn != i+1 {
				t.Fatalf("guard response kind %d turn %d: %+v damage %d player %+v", tc.kind, i, b, damage, p)
			}
		}
		for i, a := range tc.pattern {
			if a == Heavy {
				b = Battle{Enemy: NewEnemy(tc.kind), Turn: i}
				_, damage := b.Respond(&p)
				if damage != Damage(b.Enemy.Attack+tc.bonus, p.Defense) {
					t.Fatalf("heavy bonus kind %d: %d", tc.kind, damage)
				}
			}
		}
	}
	for _, defense := range []int{1, 100} {
		b := Battle{Enemy: NewEnemy(EnemyDragon)}
		p := Player{HP: 100, MaxHP: 100, Defense: defense}
		b.Commit(&p, nil, Guard, 0, false)
		_, got := b.Respond(&p)
		want := max(1, (Damage(6, defense)+1)/2)
		if got != want {
			t.Fatalf("guard rounding/minimum: got %d want %d", got, want)
		}
	}
}

func TestBraceAndItemTransactions(t *testing.T) {
	for _, command := range []Command{Fight, Item, Guard, Run} {
		b := Battle{Enemy: NewEnemy(EnemyDragon), Braced: true}
		p := Player{HP: 10, MaxHP: 30, Attack: 10}
		counts := map[int]int{Potion: 1}
		if _, ok := b.Commit(&p, counts, command, Potion, false); !ok || b.Braced {
			t.Fatalf("brace did not expire for %d", command)
		}
		if command == Fight && b.Enemy.HP != 16 {
			t.Fatal("brace defense not applied")
		}
		if command == Item && (p.HP != 22 || counts[Potion] != 0) {
			t.Fatal("healing transaction")
		}
	}
	for _, tc := range []struct{ hp, item, count int }{{30, Potion, 1}, {10, Potion, 0}, {10, -1, 1}, {10, 2, 1}} {
		b := Battle{Enemy: NewEnemy(EnemyDragon), Braced: true, Turn: 1}
		p := Player{HP: tc.hp, MaxHP: 30}
		counts := map[int]int{Potion: tc.count, MoonHerb: 1}
		before := b
		if _, ok := b.Commit(&p, counts, Item, tc.item, false); ok || b != before || p.HP != tc.hp || counts[Potion] != tc.count || counts[MoonHerb] != 1 {
			t.Fatalf("failed item mutated state: %+v", tc)
		}
	}
	b := Battle{Enemy: NewEnemy(EnemyDragon)}
	p := Player{HP: 20, MaxHP: 30}
	counts := map[int]int{MoonHerb: 1}
	b.Commit(&p, counts, Item, MoonHerb, false)
	if p.HP != 30 || counts[MoonHerb] != 0 {
		t.Fatal("herb healing should cap at maximum")
	}
}

func TestBattleEndsBeforeResponse(t *testing.T) {
	for _, command := range []Command{Fight, Run} {
		b := Battle{Enemy: NewEnemy(EnemyDragon)}
		p := Player{HP: 30, MaxHP: 30, Attack: 100}
		b.Commit(&p, nil, command, 0, true)
		message, damage := b.Respond(&p)
		if !b.Finished || p.HP != 30 || b.Turn != 0 || message != "" || damage != 0 {
			t.Fatalf("response after battle ends: %+v", b)
		}
		if _, ok := b.Commit(&p, nil, Fight, 0, false); ok {
			t.Fatal("finished battle accepted action")
		}
	}
	b := Battle{Enemy: NewEnemy(EnemyKnight)}
	p := Player{HP: 30, MaxHP: 30}
	b.Commit(&p, nil, Run, 0, true)
	b.Respond(&p)
	if b.Finished || b.Escaped || b.Turn != 1 || p.HP >= 30 {
		t.Fatal("boss escape must fail and cost turn")
	}
}

func TestPurchasesAndRewards(t *testing.T) {
	p := Player{HP: 20, MaxHP: 32, Attack: 9, Defense: 3, Level: 1, Gold: 3}
	counts := map[int]int{}
	progress := Progress{}
	if _, ok := Purchase(&p, counts, &progress, false); ok || p.Gold != 3 || counts[Potion] != 0 {
		t.Fatal("unaffordable potion")
	}
	if _, ok := Purchase(&p, counts, &progress, true); ok || progress.BladeBought || p.Attack != 9 {
		t.Fatal("unaffordable blade")
	}
	p.Gold = 20
	Purchase(&p, counts, &progress, false)
	Purchase(&p, counts, &progress, false)
	Purchase(&p, counts, &progress, true)
	if p.Gold != 0 || counts[Potion] != 2 || p.Attack != 11 || !progress.BladeBought {
		t.Fatal("purchase balances")
	}
	p.Gold = 20
	if _, ok := Purchase(&p, counts, &progress, true); ok || p.Gold != 20 || p.Attack != 11 {
		t.Fatal("duplicate blade purchase")
	}
	p.XP = 10
	b := Battle{Enemy: NewEnemy(EnemyGuardian), Finished: true}
	b.Enemy.HP = 0
	if !b.Award(&p, &progress) || !progress.ShrineCleared || p.Level != 2 || p.MaxHP != 46 || p.HP != 46 || p.Attack != 13 {
		t.Fatalf("guardian reward %+v", p)
	}
	before := p
	if b.Award(&p, &progress) || p != before {
		t.Fatal("duplicate reward")
	}
	b = Battle{Enemy: NewEnemy(EnemyGuardian), Finished: true}
	b.Enemy.HP = 0
	if b.Award(&p, &progress) || p != before {
		t.Fatal("guardian reward from another battle duplicated")
	}
	b = Battle{Enemy: NewEnemy(EnemyGuardian), Finished: true, Escaped: true}
	if b.Award(&p, &progress) || p != before {
		t.Fatal("escape awarded rewards")
	}
	b = Battle{Enemy: NewEnemy(EnemyKnight), Finished: true}
	b.Enemy.HP = 0
	b.Award(&p, &progress)
	if p.Level != 3 || p.MaxHP != 52 || p.Attack != 15 || p.HP != 52 {
		t.Fatalf("level lost permanent upgrades: %+v", p)
	}
}

func TestWardenRoutes(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		hp, attack, minHeavy int
	}{{"direct", 38, 11, 0}, {"full", 52, 15, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			p := Player{HP: tc.hp, MaxHP: tc.hp, Attack: tc.attack, Defense: 3}
			b := Battle{Enemy: NewEnemy(EnemyKnight)}
			counts := map[int]int{Potion: 2}
			heavy := 0
			for turns := 0; !b.Finished && turns < 100; turns++ {
				command := Fight
				if b.Intent() == Heavy {
					command = Guard
					heavy++
				}
				if b.Intent() == Charge && p.HP <= 26 && counts[Potion] > 0 {
					command = Item
				}
				if _, ok := b.Commit(&p, counts, command, Potion, false); !ok {
					t.Fatal("route command rejected")
				}
				b.Respond(&p)
			}
			if b.Enemy.HP != 0 || p.HP <= 0 || heavy < tc.minHeavy {
				t.Fatalf("route failed: player %+v battle %+v heavies %d", p, b, heavy)
			}
		})
	}
}

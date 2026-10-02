# Ashen Crown

Small old-school JRPG built in Go with Ebitengine. Explore Verdant Reach, trade with a merchant, discover an optional shrine dungeon, and defeat the Cinder Warden through turn-based combat.

## Run

Requires Go 1.24+ and desktop graphics and audio libraries supported by Ebitengine.

```sh
make run
```

The game opens at the title menu. Choose **Begin Journey** to start; close the window to quit.

`make run` also handles Fedora systems that provide `libXxf86vm.so.1` without the development linker symlink. On other Linux distributions, install the Ebitengine desktop dependencies documented for that distribution.

## Controls

| Key | Action |
| --- | --- |
| Arrow keys / WASD | Move or choose command |
| Z / Enter / Space | Confirm or interact |
| X | Back |
| Escape | Return to title while exploring or fighting; back out of menus |
| I / Tab | Open inventory |
| M | Mute or unmute sound effects |
| R | Reset the current puzzle |

Walk onto the keep gate at the east edge to face the boss. Chests, the merchant, shrine lamps, and the altar respond when facing them and pressing interact. Cyan stairs connect the Reach and the shrine.

Confirm on the victory or defeat screen returns to the title menu. Choose **Begin Journey** to reset progress and start a new run; the sound mute preference persists.

Press Escape while exploring or fighting to leave the current run and return to the title menu. Escape still backs out of inventory, shop, puzzle, and item-selection screens first.

## Combat and preparation

The enemy's **NEXT** action is always visible. **GUARD** blocks charged strikes completely and halves ordinary damage. **CHARGE** is a safe turn to attack or heal. Dragons alternate striking and bracing; Wisps begin charged. Items can be selected in combat; cancelling or choosing an unusable item costs no turn.

The merchant near the starting path sells Potions for 4 gold and a one-time Tempered Blade for 12 gold (+2 Attack). The two marked Wisp encounters along the main path provide enough XP for level 2 and enough gold for the blade. Marked encounters remain until defeated; escaping does not clear them.

The optional shrine has three puzzle chambers: rune ordering in the western wing, a nine-lamp circuit in the eastern wing, and spectral offerings beside the central passage. Solve each wing's tablet to awaken its lamp; both lamps open the northern altar gate. The offering chamber opens the inner seal. Puzzles are resettable with R and consume no inventory items.

Additional marked Echo encounters guard the passages, awarding gold and a little XP. The treasure detour holds a Moon Herb. Defeat the altar guardian for +8 maximum HP and a full heal. You may return to the Reach at any time. Progress persists between areas and resets when starting a new run; mute preference persists across restarts.

## Verify

```sh
make test
make vet
make build
```

With a desktop display available, run the integration and rendering checks:

```sh
make test-ui
```

The rendering check briefly opens a game window and saves screenshots of the maps, menus, combat states, and endings in `.build/screenshots`. Headless CI can use `make test`; it verifies the combat rules, purchases, rewards, and shrine reachability without starting graphics or audio.

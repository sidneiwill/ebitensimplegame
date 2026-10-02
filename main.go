package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(screenWidth*3, screenHeight*3)
	ebiten.SetWindowTitle("Ashen Crown")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	game := NewGame(1, NewSoundBank())
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

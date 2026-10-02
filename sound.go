package main

import (
	"encoding/binary"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

type SoundID int

const (
	SoundMenu SoundID = iota
	SoundConfirm
	SoundError
	SoundHit
	SoundBlock
	SoundHeal
	SoundReward
	SoundVictory
)

type SoundBank struct {
	Muted   bool
	players [8]*audio.Player
}

func NewSoundBank() *SoundBank {
	context := audio.CurrentContext()
	if context == nil {
		context = audio.NewContext(48000)
	}
	bank := &SoundBank{}
	sounds := [...]struct {
		notes    []float64
		duration float64
	}{
		{[]float64{660}, .045},
		{[]float64{660, 880}, .07},
		{[]float64{220, 165}, .09},
		{[]float64{130, 65}, .06},
		{[]float64{330, 220}, .065},
		{[]float64{440, 550, 660}, .09},
		{[]float64{660, 880, 1100}, .09},
		{[]float64{523.25, 659.25, 783.99, 1046.5, 783.99, 1046.5}, .16},
	}
	for id, sound := range sounds {
		bank.players[id] = context.NewPlayerFromBytes(synthesizeSound(context.SampleRate(), sound.notes, sound.duration))
	}
	return bank
}

// synthesizeSound emits quiet stereo PCM with an envelope to avoid clicks.
func synthesizeSound(sampleRate int, notes []float64, duration float64) []byte {
	frames := int(float64(sampleRate) * duration)
	pcm := make([]byte, frames*len(notes)*4)
	for note, frequency := range notes {
		for frame := 0; frame < frames; frame++ {
			t := float64(frame) / float64(sampleRate)
			envelope := math.Min(1, t/.005) * math.Min(1, (duration-t)/.025)
			wave := math.Sin(2*math.Pi*frequency*t) + .25*math.Sin(6*math.Pi*frequency*t)
			sample := uint16(int16(wave * envelope * 2400))
			offset := (note*frames + frame) * 4
			binary.LittleEndian.PutUint16(pcm[offset:], sample)
			binary.LittleEndian.PutUint16(pcm[offset+2:], sample)
		}
	}
	return pcm
}

func (bank *SoundBank) Play(id SoundID) {
	if bank == nil || bank.Muted || id < 0 || int(id) >= len(bank.players) {
		return
	}
	player := bank.players[id]
	if player == nil {
		return
	}
	player.Pause()
	if err := player.Rewind(); err != nil {
		log.Printf("rewind sound %d: %v", id, err)
		return
	}
	player.Play()
}

func (bank *SoundBank) Stop() {
	if bank == nil {
		return
	}
	for id, player := range bank.players {
		if player == nil {
			continue
		}
		player.Pause()
		if err := player.Rewind(); err != nil {
			log.Printf("rewind sound %d: %v", id, err)
		}
	}
}

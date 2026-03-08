// This script shows a chaser light using the SetSegment method.
//
// This configures a 4-digit 7-segment Common Cathode display for raspberry pi pico 2

package main

import (
	"machine"
	"time"

	"github.com/niekdomi/sevseg"
)

func main() {
	displayConfig := sevseg.Config{
		Hardware: sevseg.CommonCathode,
		DigitPins: []machine.Pin{
			machine.GP0,
			machine.GP1,
			machine.GP2,
			machine.GP3,
		},
		SegmentPins: []machine.Pin{
			machine.GP4,  // A
			machine.GP5,  // B
			machine.GP6,  // C
			machine.GP7,  // D
			machine.GP8,  // E
			machine.GP9,  // F
			machine.GP10, // G
			machine.GP11, // DP
		},
		UseLeadingZeros: false,
	}

	display, ok := sevseg.NewSevSeg(displayConfig)
	if !ok {
		return
	}

	chaserLight := [][]uint8{
		{0b00000001, 0b00000000, 0b00000000, 0b00000000},
		{0b00000010, 0b00000000, 0b00000000, 0b00000000},
		{0b00000100, 0b00000000, 0b00000000, 0b00000000},
		{0b00001000, 0b00000000, 0b00000000, 0b00000000},
		{0b00000000, 0b00001000, 0b00000000, 0b00000000},
		{0b00000000, 0b00000000, 0b00001000, 0b00000000},
		{0b00000000, 0b00000000, 0b00000000, 0b00001000},
		{0b00000000, 0b00000000, 0b00000000, 0b00010000},
		{0b00000000, 0b00000000, 0b00000000, 0b00100000},
		{0b00000000, 0b00000000, 0b00000000, 0b00000001},
		{0b00000000, 0b00000000, 0b00000001, 0b00000000},
		{0b00000000, 0b00000001, 0b00000000, 0b00000000},
		{0b00000001, 0b00000000, 0b00000000, 0b00000000},
	}

	nextSegment := time.NewTicker(50 * time.Millisecond)
	defer nextSegment.Stop()

	refreshTicker := time.NewTicker(1 * time.Millisecond)
	defer refreshTicker.Stop()

	i := 0
	for {
		select {
		case <-nextSegment.C:
			display.SetSegment(chaserLight[i])
			i = (i + 1) % len(chaserLight)
		case <-refreshTicker.C:
			display.Refresh()
		}
	}
}

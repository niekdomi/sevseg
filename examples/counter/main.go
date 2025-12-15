// This script displays a counting sequence from -9 to 99 on a 2-digit 7-segment display.
// Demonstrates basic number display functionality.
//
// This configures a 2-digit 7-segment Common Cathode display for arduino-nano

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
			machine.D3,
			machine.D2,
		},
		SegmentPins: []machine.Pin{
			machine.D4,  // A
			machine.D5,  // B
			machine.D6,  // C
			machine.D7,  // D
			machine.D8,  // E
			machine.D9,  // F
			machine.D10, // G
			machine.D11, // DP
		},
		UseLeadingZeros: true,
	}

	display, ok := sevseg.NewSevSeg(displayConfig)
	if !ok {
		return
	}

	counter := int32(-9)
	refreshCounter := 0
	refreshesPerUpdate := 200

	display.SetNumber(counter)

	for {
		refreshCounter++
		if refreshCounter >= refreshesPerUpdate {
			counter++
			if counter > 99 {
				counter = -9
			}
			display.SetNumber(counter)
			refreshCounter = 0
		}

		display.Refresh()
		time.Sleep(1 * time.Millisecond)
	}
}

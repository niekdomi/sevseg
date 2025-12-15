// This script cycles through all displayable letters of the alphabet.
// Shows which letters can be displayed on a 7-segment display.
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
		UseLeadingZeros: false,
	}

	display, ok := sevseg.NewSevSeg(displayConfig)
	if !ok {
		return
	}

	for {
		for i := 'A'; i <= 'Z'; i++ {
			display.SetText(string(i))

			for range 400 {
				display.Refresh()
				time.Sleep(1 * time.Millisecond)
			}
		}
	}
}

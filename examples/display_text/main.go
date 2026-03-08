// This script cycles through different text messages on a 2-digit 7-segment display.
// Demonstrates text display capabilities with supported characters.
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

	messages := []string{"Hi", "Go", "On", "Ab", "Cd", "EF"}
	messageIndex := 0
	refreshCounter := 0
	refreshesPerUpdate := 500

	display.SetText(messages[messageIndex])

	for {
		refreshCounter++
		if refreshCounter >= refreshesPerUpdate {
			messageIndex++
			if messageIndex >= len(messages) {
				messageIndex = 0
			}
			display.SetText(messages[messageIndex])
			refreshCounter = 0
		}

		display.Refresh()
		time.Sleep(1 * time.Millisecond)
	}
}

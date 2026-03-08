// This script displays numbers with decimal points.
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

	examples := []struct {
		number  int32
		decimal uint8
	}{
		{42, 0},
		{42, 1},
		{73, 0},
		{18, 1},
	}

	index := 0

	for {
		example := examples[index]
		display.SetNumberWithDecimal(example.number, example.decimal)

		index = (index + 1) % len(examples)

		for range 500 {
			display.Refresh()
			time.Sleep(1 * time.Millisecond)
		}

	}
}

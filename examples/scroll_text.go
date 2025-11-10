// This script shows the ability to scroll text on a 4-digit 7-segment display.
//
// This configures a 4-digit 7-segment Common Cathode display for raspberry pi pico 2

package main

import (
	"machine"
	"time"

	"github.com/domi413/sevseg"
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
		PWMType:         sevseg.SoftwarePWM,
		UseLeadingZeros: false,
	}

	display, ok := sevseg.NewSevSeg(displayConfig)
	if !ok {
		return
	}

	if !display.SetText("Crazy display test") {
		return
	}

	scrollTicker := time.NewTicker(200 * time.Millisecond)
	defer scrollTicker.Stop()

	refreshTicker := time.NewTicker(1 * time.Millisecond)
	defer refreshTicker.Stop()

	for {
		select {
		case <-scrollTicker.C:
			display.ScrollTextLeft()
		case <-refreshTicker.C:
			display.Refresh()
		}
	}
}

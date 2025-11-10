# TinyGo 7-Segment Display Library

A comprehensive TinyGo library for controlling multiplexed 7-segment displays
with support for both common-anode and common-cathode configurations. This
library supports numbers up to `int32`, making it ideal for displays with up to
8 digits.

This library is inspired by the popular
[SevSeg](https://github.com/DeanIsMe/SevSeg/) library for Arduino, adapted and
improved for TinyGo.

## Installation

Include `import "github.com/domi413/sevseg"` in your TinyGo project and run the
following commands to set up the module:

```bash
go mod init {your-project} # May already be done
go get github.com/domi413/sevseg@main
```

## Hardware Setup

### 7-Segment Display Pinout

```
 AAA
F   B
F   B
 GGG
E   C
E   C
 DDD  DP
```

### Pin Mapping

The library expects segment pins in this order:

1. Segment A
2. Segment B
3. Segment C
4. Segment D
5. Segment E
6. Segment F
7. Segment G
8. Decimal Point (DP) (optional, required for decimal points and certain symbols)

### Brightness Control

Brightness control requires PWM-capable pins for the `SegmentPins` when using
`HardwarePWM`. Ensure your microcontroller supports PWM on the chosen pins and
configure them in the `Config.PWMPins` field if `HardwarePWM` is selected.
Software PWM is also supported but may require more CPU resources.

## Simple Example for a 2-Digit Common-Cathode Display on an Arduino Nano

```go
package main

import (
	"machine"
	"time"
	"github.com/domi413/sevseg"
)

func main() {
	displayConfig := sevseg.Config{
		Hardware: sevseg.CommonCathode,
		PWMType:  sevseg.SoftwarePWM,
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
		panic("Failed to create display")
	}

	if !display.SetNumber(69) {
		panic("Failed to set number")
	}

	for {
		// Refresh the display at >100Hz to avoid flicker
		display.Refresh()
		time.Sleep(time.Millisecond * 10)
	}
}
```

## API Reference

### Configuration

```go
type Config struct {
	Hardware        displayType     // CommonAnode or CommonCathode
	PWMType         pwmType         // SoftwarePWM or HardwarePWM
	DigitPins       []machine.Pin   // Pins for multiplexing the digits
	SegmentPins     []machine.Pin   // Pins controlling segments (A-G, optionally DP)
	UseLeadingZeros bool            // Whether to display leading zeros for numbers
	// PWMPins      []machine.PWM   // PWM timers for HardwarePWM (NOT IMPLEMENTED YET)
}
```

### Methods

#### `NewSevSeg(config Config) (*SevSeg, bool)`

Creates a new `SevSeg` instance with the provided configuration. Returns the
display instance and a boolean indicating success (`true`) or failure
(`false`).

- **Failure cases**: Invalid configuration (e.g., no digit pins, fewer than 7
  or more than 8 segment pins).

#### `DisplayTest(delayMS uint16)`

Tests the display by iterating through each segment (A-G, DP) for each digit.
Does not require external `Refresh()` calls, as it handles refreshing
internally.

- **Parameter**: `delayMS` specifies the duration (in milliseconds) each
  segment is displayed.

#### `Toggle(enable bool)`

Toggles the display on or off. The blinking interval must be managed by the
user by alternating the `enable` parameter.

#### `Clear()`

Clears the display by setting all segments to blank (space).

#### `Off()`

Turns off the display immediately by clearing all digit and segment pins,
without requiring a `Refresh()` call.

#### `On()`

Turns on the display, restoring the previous brightness (defaults to 100% if
previously set to 0).

#### `GetDisplayWidth() uint8`

Returns the number of digits in the display.

#### `IsCharacterSupported(char byte) bool`

Checks if a specific character can be displayed on the 7-segment display.

- **Returns**: `true` if supported, `false` otherwise.

#### `SetBrightness(brightness uint8)`

Sets the display brightness as a percentage (0–100). Values above 100 are
clamped to 100. A value of 0 disables the display.

- **Note**: Requires PWM-capable pins for `HardwarePWM` or sufficient CPU
  resources for `SoftwarePWM`.

#### `SetNumber(number int32) bool`

Sets a number (up to `int32`) to be displayed. Supports positive and negative
numbers. Leading zeros are displayed only if `UseLeadingZeros` is `true`.

- **Returns**: `true` on success, `false` if the number exceeds the display’s
  digit capacity.

#### `SetNumberFloat(number float32, decimalPlaces uint8) bool`

Displays a floating-point number with the specified number of decimal places.

- **Parameters**:
  - `number`: The float to display.
  - `decimalPlaces`: Number of decimal places to show.
- **Returns**: `true` on success, `false` if the number exceeds capacity or
  `decimalPlaces` is 0.

#### `SetNumberWithDecimal(number int32, decimalPointPosition uint8) bool`

Displays a number with a decimal point at the specified position (zero-indexed
from the right, e.g., `1234` with `decimalPointPosition=1` displays `123.4`).

- **Returns**: `true` on success, `false` if the number exceeds capacity, the
  decimal point position is invalid, or the display lacks a decimal point pin.

#### `SetNumberWithMultipleDecimals(number int32, decimalPointsPositions []uint8) bool`

Displays a number with multiple decimal points at specified positions
(zero-indexed from the right, e.g., `1234` with `[]uint8{1, 2}` displays
`12.3.4`).

- **Returns**: `true` on success, `false` if the number exceeds capacity, any
  decimal point position is invalid, or the display lacks a decimal point pin.

#### `SetHex(number uint32) bool`

Displays a number in hexadecimal format.

- **Returns**: `true` on success, `false` if the number exceeds the display’s
  digit capacity.

#### `SetTemperature(temperature float32, decimalPlaces uint8) bool`

Displays a temperature with a degree symbol (`°`). Requires at least 2 digits.

- **Parameters**:
  - `temperature`: The temperature to display.
  - `decimalPlaces`: Number of decimal places.
- **Returns**: `true` on success, `false` if the number exceeds capacity or the
  display has fewer than 2 digits.

#### `SetTemperatureWithUnit(temperature float32, decimalPlaces uint8, unit tempUnit) bool`

Displays a temperature with a degree symbol and unit (`°C` or `°F`). Requires
at least 3 digits.

- **Parameters**:
  - `temperature`: The temperature to display.
  - `decimalPlaces`: Number of decimal places.
  - `unit`: `TemperatureUnit.Celsius` or `TemperatureUnit.Fahrenheit`.
- **Returns**: `true` on success, `false` if the number exceeds capacity or the
  display has fewer than 3 digits.

#### `SetSegment(pattern []uint8) bool`

Sets a custom segment pattern for each digit. The pattern is a bitmask where
each bit corresponds to a segment (A-G, DP).

- **Example**: To display the following pattern:

```
‾│ │‾  │‾  ‾│
_|.│_  │_. _│
```

Use `[]uint8{0b10001111, 0b00111001, 0b10111001, 0b00001111}` (right to left).

- **Returns**: `true` on success, `false` if the pattern length exceeds the
  number of digits.

#### `SetText(text string) bool`

Displays text on the 7-segment display. Text is written from left to right. If
the text is shorter than the display, remaining digits (on the right) are
cleared. If longer, use `ScrollTextLeft` or `ScrollTextRight`.

- **Supported Characters**:
  - Digits: `0-9`
  - Letters: `A-Z` (case insensitive)
  - Special: ` ` (space), `-` (minus), `.` (decimal point), `°` (degree), `_`
    (underscore)
- **Returns**: `true` on success, `false` if the text contains unsupported
  characters or is too long (without scrolling).

#### `ScrollTextLeft()`

Scrolls the displayed text left by one digit. No effect if the text length is
less than or equal to the display width.

#### `ScrollTextRight()`

Scrolls the displayed text right by one digit. No effect if the text length is
less than or equal to the display width.

#### `Refresh() bool`

Refreshes the display by cycling through each digit. Must be called frequently
(recommended >100Hz, e.g., every 10ms) to maintain a stable, flicker-free
display.
todo:

- **Note**: Do not call `Refresh()` too frequently (e.g., in a tight loop
  without delay), as this may overload the microcontroller. Use a delay (e.g.,
  `time.Sleep(time.Millisecond * 10)`) to achieve a reasonable refresh rate.
- **Returns**: `true` on success, `false` if the display is not initialized or
  disabled.todo:

## Troubleshooting

### Display is Dim or Flickering

- Ensure `Refresh()` is called with at least 100Hz (e.g., every 10ms).
- Verify resistor values (too high resistance can cause dimming).
- Check the power supply’s current capacity.
- For brightness control, ensure PWM pins are correctly configured if using
  `HardwarePWM`.

### Numbers Appear Backwards

- Verify digit pins are connected in the correct order.
- Ensure segment pins match the expected order (A-G, DP).

### Display Shows Wrong Characters

- Confirm the correct display type (`CommonAnode` or `CommonCathode`).
- Check segment pin wiring order.

### Some Segments Don’t Light Up

- Inspect for loose connections.
- Verify resistor values and connections.
- Test segments with a multimeter.
- Ensure `Refresh()` is called frequently enough.

### Method Returns `false`

- For `SetNumber`, ensure the number fits within the display’s digits (up to
  `int32` for 8 digits).
- For decimal operations, ensure 8 segment pins (including DP) are defined.
- For text operations, verify all characters are supported and text length is
  appropriate.
- For segment patterns, ensure the pattern length does not exceed the digit
  count.

## Error Handling

This library uses boolean return values for memory efficiency in embedded
environments:

- `true`: Operation successful.
- `false`: An error occurred (e.g., invalid input, configuration, or capacity
  exceeded).

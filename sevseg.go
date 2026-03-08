//go:build tinygo

// Package sevseg is a library for controlling 7-segment displays.
package sevseg

import (
	"machine"
	"time"
)

// TempUnit defines the temperature unit for temperature display.
type TempUnit uint8

const (
	Celsius TempUnit = iota
	Fahrenheit
)

// HardwarePWM and SoftwarePWM define the type of PWM used for brightness control.
type PWMType uint8

const (
	SoftwarePWM PWMType = iota
	HardwarePWM
)

// type pwmChannelMap struct {
// 	pwm     machine.PWM
// 	channel uint8
// }

// CommonAnode and CommonCathode define the type of 7-segment display.
type DisplayType uint8

const (
	CommonAnode DisplayType = iota
	CommonCathode
)

// Config holds the configuration for a 7-segment display.
type Config struct {
	// Hardware defines the type of 7-segment display.
	// It can be either CommonAnode or CommonCathode.
	Hardware DisplayType

	// PWM defines the type of PWM used for brightness control.
	//
	// If you want to use the hardware PWM you need to configure PWMTimers and
	// PWMPins.
	PWMType PWMType

	// PWMPins defines the PWM pins e.g., [machine.Timer0, machine.Timer1] or
	// [machine.PWM3, machine.PWM4] depending on the board.
	// PWMPins []machine.PWM

	// DigitPins defines the pins used control/multiplex the digits.
	DigitPins []machine.Pin

	// SegmentPins defines the pins used to control the segments of the display.
	// Normally, these are 7 or 8 pins, depending on whether a decimal point is
	// used.
	SegmentPins []machine.Pin

	// UseLeadingZeros defines whether leading zeros should be displayed.
	UseLeadingZeros bool
}

// SevSeg represents a 7-segment display.
type SevSeg struct {
	config          DisplayType
	pwm             PWMType
	digitPins       []machine.Pin
	segmentPins     []machine.Pin
	useLeadingZeros bool

	// Internal state
	enabled    bool
	brightness uint8
	// pwmChannels map[machine.Pin]pwmChannelMap

	// Text scrolling state
	scrollPosition uint8
	textPattern    []uint8

	// Refresh state
	pwmCounter            uint8
	currentDigitToRefresh uint8
	updatedDisplay        []uint8
}

// NewSevSeg creates a new instance of sevSeg with the provided configuration.
func NewSevSeg(cfg Config) (*SevSeg, bool) {
	if len(cfg.DigitPins) == 0 || len(cfg.SegmentPins) < 7 || len(cfg.SegmentPins) > 8 {
		return nil, false
	}

	for _, pin := range cfg.DigitPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}

	for _, pin := range cfg.SegmentPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}

	s := &SevSeg{
		config:          cfg.Hardware,
		pwm:             cfg.PWMType,
		digitPins:       cfg.DigitPins,
		segmentPins:     cfg.SegmentPins,
		useLeadingZeros: cfg.UseLeadingZeros,
		brightness:      100,
		enabled:         true,
		// pwmChannels:           make(map[machine.Pin]pwmChannelMap),
		updatedDisplay:        make([]uint8, len(cfg.DigitPins)),
		currentDigitToRefresh: 0,
	}

	// if s.pwm == HardwarePWM && !s.configurePWM(cfg.PWMPins) {
	// 	return nil, false
	// }

	s.clearDigitPins()
	s.clearSegmentPins()

	return s, true
}

// DisplayTest is a standalone method that can be used to test the functionality
// of the display or if it's correctly wired up. It will iterate over each
// segment and digit of the display.
//
// A -> B -> C -> D -> E -> F -> G -> DP
//
// Note that this method must not require to call Refresh externally.
func (s *SevSeg) DisplayTest(delayMS uint16) {
	segmentPatterns := []uint8{
		0b00000001, // Segment A
		0b00000010, // Segment B
		0b00000100, // Segment C
		0b00001000, // Segment D
		0b00010000, // Segment E
		0b00100000, // Segment F
		0b01000000, // Segment G
		0b10000000, // Segment DP
	}

	for i := range len(s.digitPins) {
		for j := range len(s.segmentPins) {
			s.updatedDisplay[i] = segmentPatterns[j]

			for range delayMS {
				s.Refresh()
				time.Sleep(time.Millisecond)
			}

			s.updatedDisplay[i] = s.getSegmentCode(36) // BLANK
		}
	}
}

// IsEnabled returns whether the display is currently enabled
func (s *SevSeg) IsEnabled() bool {
	return s.enabled
}

// Toggle can be used to toggle/blink the display. A boolean value is passed to
// enable or disable the display.
//
// Since this library doesn't handle timing, the blinking interval must be
// handled by the user by passing a toggling boolean value.
func (s *SevSeg) Toggle(enable bool) {
	s.enabled = enable
}

// Clear clears the display by setting all segments to blank.
func (s *SevSeg) Clear() {
	for i := range s.updatedDisplay {
		s.updatedDisplay[i] = s.getSegmentCode(36) // BLANK
	}
}

// Off turns the display off by setting all digit and segment pins to their
// respective off state, depending on the display type.
//
// This turns off the display immediately without calling Refresh.
func (s *SevSeg) Off() {
	s.enabled = false
	s.clearDigitPins()
	s.clearSegmentPins()
}

// On turns the display on.
func (s *SevSeg) On() {
	s.enabled = true

	if s.brightness == 0 {
		s.brightness = 100
	}
}

// GetDisplayWidth returns the amount of digits the display has.
func (s *SevSeg) GetDisplayWidth() uint8 {
	return uint8(len(s.digitPins))
}

// IsCharacterSupported checks if a specific character can be displayed.
func (s *SevSeg) IsCharacterSupported(char byte) bool {
	_, ok := s.charToSegmentPattern(char)
	return ok
}

// GetBrightness returns the current brightness level (0-100)
func (s *SevSeg) GetBrightness() uint8 {
	return s.brightness
}

// SetBrightness sets the brightness of the display.
//
// Takes the brightness level in percentage (0-100) as an argument.
// Any value greater than 100 will be clamped to 100.
func (s *SevSeg) SetBrightness(brightness uint8) {
	if brightness == 0 {
		s.enabled = false
	} else {
		s.enabled = true
	}

	if brightness > 100 {
		s.brightness = 100
	} else {
		s.brightness = brightness
	}
}

// SetNumber sets the number to be displayed.
func (s *SevSeg) SetNumber(number int32) bool {
	if !s.checkAvailableDigits(number, 10) {
		return false
	}

	s.setNumberInitPattern()

	isNegative := number < 0
	if isNegative {
		number = -number
	}

	position := 0
	if number == 0 {
		s.updatedDisplay[position] = s.getSegmentCode(0) // ZERO
	} else {
		for number > 0 && position < len(s.digitPins) {
			digit := uint8(number % 10)
			s.updatedDisplay[position] = s.getSegmentCode(digit)
			number /= 10

			position++
		}
	}

	if isNegative {
		s.updatedDisplay[position] = s.getSegmentCode(37) // MINUS
	}

	return true
}

// SetNumberFloat takes a float number as argument and displays it with a
// specified number of decimal places.
func (s *SevSeg) SetNumberFloat(number float32, decimalPlaces uint8) bool {
	if decimalPlaces <= 0 {
		return false
	}

	scale := int32(1)
	for range decimalPlaces {
		scale *= 10
	}

	scaled := int32(number * float32(scale))

	if !s.SetNumberWithDecimal(scaled, decimalPlaces) {
		return false
	}

	return true
}

// SetNumberWithDecimal sets the number to be displayed, including a decimal
// point at a specified position.
//
// decimalPointPosition specifies the position of the decimal point from right to
// left since the LSB is the right most digit.
//
// E.g. for a 4-digit display, decimalPointPosition = 1 would look like this: 000.0
func (s *SevSeg) SetNumberWithDecimal(number int32, decimalPointPosition uint8) bool {
	return s.SetNumberWithMultipleDecimals(number, []uint8{decimalPointPosition})
}

// SetNumberWithMultipleDecimals sets the number to be displayed, including
// multiple decimal points at specified positions.
//
// decimalPointsPositions is a slice of positions for the decimal points from right
// to left, since the LSB is the right most digit.
//
// E.g. for a 4-digit display, decimalPointsPositions = []uint{1, 2} would look like
// this: 00.0.0
func (s *SevSeg) SetNumberWithMultipleDecimals(number int32, decimalPointsPositions []uint8) bool {
	if len(decimalPointsPositions) == 0 {
		return false
	}

	for _, decimalPos := range decimalPointsPositions {
		if decimalPos > uint8(len(s.digitPins)) {
			return false
		}
	}

	if len(s.segmentPins) < 8 {
		return false
	}

	if !s.SetNumber(number) {
		return false
	}

	for _, decimalPos := range decimalPointsPositions {
		if decimalPos > uint8(len(s.digitPins)) {
			return false
		}

		s.updatedDisplay[decimalPos] |= s.getSegmentCode(38) // DECIMAL POINT
	}

	return true
}

// SetHex sets the number to be displayed as a hexadecimal value.
func (s *SevSeg) SetHex(number uint32) bool {
	if !s.checkAvailableDigits(int32(number), 16) {
		return false
	}

	s.setNumberInitPattern()

	position := 0
	if number == 0 {
		s.updatedDisplay[position] = s.getSegmentCode(0) // ZERO

		return true
	}

	for number > 0 && position < len(s.digitPins) {
		digit := uint8(number) % 16
		s.updatedDisplay[position] = s.getSegmentCode(digit)
		number /= 16
		position++

	}

	return true
}

// SetTemperature sets the temperature to be displayed with a ° character.
func (s *SevSeg) SetTemperature(temperature float32, decimalPlaces uint8) bool {
	if len(s.digitPins) <= 1 {
		return false // We need at least 2 digits to display a number
	}

	scale := int32(1)
	for range decimalPlaces + 1 { // Additional *10 for the ° Character
		scale *= 10
	}

	scaled := int32(temperature * float32(scale))

	if !s.checkAvailableDigits(int32(scaled), 10) {
		return false
	}

	if decimalPlaces > 0 {
		// Scale temperature by 10 to reserve space for ° symbol
		if !s.SetNumberWithDecimal(scaled, decimalPlaces+1) {
			return false
		}
	} else {
		if !s.SetNumber(scaled) {
			return false
		}
	}

	s.updatedDisplay[0] = s.getSegmentCode(39) // DEGREE

	return true
}

// SetTemperatureWithUnit sets the temperature to be displayed in °C or °F.
// Note that two digits are required to show °C / °F
func (s *SevSeg) SetTemperatureWithUnit(temperature float32, decimalPlaces uint8, unit tempUnit) bool {
	if len(s.digitPins) <= 2 {
		return false // We need at least 3 digits to display a number
	}

	// Scale temperature by 10 to reserve space for unit symbol (C/F)
	adjustedDecimalPlaces := decimalPlaces
	if decimalPlaces > 0 {
		adjustedDecimalPlaces++ // Move decimal point
	}
	if !s.SetTemperature(temperature*10, adjustedDecimalPlaces) {
		return false
	}

	s.updatedDisplay[1] = s.getSegmentCode(39) // DEGREE
	if unit == Celsius {
		s.updatedDisplay[0] = s.getSegmentCode(12) // 'C'
	} else {
		s.updatedDisplay[0] = s.getSegmentCode(15) // 'F'
	}

	return true
}

// SetSegment can be used to display any arbitrary segment pattern.
// E.g. to display the following pattern:
//
//		‾│ │‾  │‾  ‾│
//		_|.│_  │_. _│
//	    4   3   2  1
//
// Would equal the bit pattern:
//
// {0b00001111, 0b10111001, 0b00111001, 0b10001111}
//
//	1           2           3           4
//
// (The numbers represent the respective digit)
//
// Note that the left most segment is the last pattern in the slice.
//
// The segments will be displayed from right to left. This means that if fewer
// segments are defined than digits available, the remaining segments (on the
// left) will be cleared.
func (s *SevSeg) SetSegment(pattern []uint8) bool {
	if len(pattern) > len(s.digitPins) {
		return false
	}

	copy(s.updatedDisplay, pattern)

	return true
}

// SetText displays a text.
//
// If the text is longer than the number of digits, an error is returned.
//
// The text is written from left to right, meaning that if the text is shorter
// than the number of digits, the remaining segments (on the right) will be cut
// off. You can use ScrollTextLeft or ScrollTextRight to scroll the text.
func (s *SevSeg) SetText(text string) bool {
	s.Clear()

	s.scrollPosition = 0

	textLength := len(text)
	displayWidth := len(s.digitPins)
	reservedTextLength := textLength

	if textLength > displayWidth {
		reservedTextLength += displayWidth
	}
	s.textPattern = make([]uint8, reservedTextLength)

	for i, char := range []byte(text) {
		segment, ok := s.charToSegmentPattern(char)
		if !ok {
			return false
		}

		s.textPattern[i] = segment
	}

	if textLength > displayWidth {
		for i := range displayWidth {
			s.textPattern[textLength+i] = s.getSegmentCode(36) // BLANK
		}
	}

	s.updateDisplayFromPatterns()

	return true
}

// ScrollTextLeft scrolls the text to the left by one digit/segment.
func (s *SevSeg) ScrollTextLeft() {
	patternLength := uint8(len(s.textPattern))

	if patternLength <= uint8(len(s.digitPins)) {
		return
	}

	s.scrollPosition = (s.scrollPosition + 1) % patternLength

	s.updateDisplayFromPatterns()
}

// ScrollTextRight scrolls the text to the right by one digit/segment.
func (s *SevSeg) ScrollTextRight() {
	patternLength := uint8(len(s.textPattern))

	if patternLength <= uint8(len(s.digitPins)) {
		return
	}

	s.scrollPosition = (s.scrollPosition - 1 + patternLength) % patternLength

	s.updateDisplayFromPatterns()
}

// Refresh updates the display. Must be called periodically, ideally with >100Hz
// to avoid flicker.
func (s *SevSeg) Refresh() bool {
	if len(s.updatedDisplay) == 0 {
		return false
	}

	s.clearDigitPins()

	if s.pwm == SoftwarePWM {
		s.softwarePWM()
	} else {
		// s.hardwarePWM()
	}

	if !s.enabled {
		return false
	}

	s.setSegmentPins()

	// Turn on the current digit
	if s.currentDigitToRefresh < uint8(len(s.digitPins)) {
		if s.config == CommonCathode {
			s.digitPins[s.currentDigitToRefresh].Low()
		} else {
			s.digitPins[s.currentDigitToRefresh].High()
		}
	}

	s.currentDigitToRefresh = (s.currentDigitToRefresh + 1) % uint8(len(s.digitPins))

	return true
}

// checkAvailableDigits checks if the number can fit within the specified number
// of digits.
func (s *SevSeg) checkAvailableDigits(number int32, base uint8) bool {
	if number == 0 {
		return len(s.digitPins) >= 1
	}

	count := 0
	if number < 0 {
		count = 1
		number = -number
	}

	for number > 0 {
		count++
		number /= int32(base)
	}

	return count <= len(s.digitPins)
}

// charToSegmentPattern converts a character to its corresponding segment
// pattern.
func (s *SevSeg) charToSegmentPattern(char byte) (uint8, bool) {
	if char >= 'a' && char <= 'z' {
		// Since we can't differ between upper and lower case letters, we
		// convert lower-case letters to upper-case.
		char = char - 'a' + 'A'
	}

	switch {
	case char >= '0' && char <= '9':
		return s.getSegmentCode(char - '0'), true
	case char >= 'A' && char <= 'Z':
		return s.getSegmentCode(char - 'A' + 10), true
	case char == ' ':
		return s.getSegmentCode(36), true
	case char == '-':
		return s.getSegmentCode(37), true
	case char == '.':
		return s.getSegmentCode(38), true
	case char == '*':
		return s.getSegmentCode(39), true
	case char == '_':
		return s.getSegmentCode(40), true
	}

	return 0, false
}

// clearDigitPins turns off all digit pins.
func (s *SevSeg) clearDigitPins() {
	for _, pin := range s.digitPins {
		if s.config == CommonCathode {
			pin.High()
		} else {
			pin.Low()
		}
	}
}

// clearSegmentPins turns off all segment pins.
func (s *SevSeg) clearSegmentPins() {
	for _, pin := range s.segmentPins {
		if s.config == CommonCathode {
			pin.Low()
		} else {
			pin.High()
		}
	}
}

// configurePWM sets up PWM channels for segment pins if HardwarePWM is used.
// func (s *SevSeg) configurePWM(pwmPins []machine.PWM) bool {
// 	for _, timer := range pwmPins {
// 		timer.Configure(machine.PWMConfig{})
// 	}

// 	for _, segmentPin := range s.segmentPins {
// 		found := false
// 		for _, timer := range pwmPins {
// 			if ch, err := timer.Channel(segmentPin); err == nil {
// 				s.pwmChannels[segmentPin] = pwmChannelMap{pwm: timer, channel: ch}
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			return false
// 		}
// 	}
// 	return true
// }

// setNumberInitPattern sets the initial pattern for the display when a number
// is set.
func (s *SevSeg) setNumberInitPattern() {
	initPattern := s.getSegmentCode(36) // BLANK
	if s.useLeadingZeros {
		initPattern = s.getSegmentCode(0) // ZERO
	}

	for i := range s.updatedDisplay {
		s.updatedDisplay[i] = initPattern
	}
}

// setSegmentPins sets the segment pins according to the current digit to
// refresh and the updated display pattern.
func (s *SevSeg) setSegmentPins() {
	pattern := s.updatedDisplay[s.currentDigitToRefresh]

	if s.config == CommonAnode {
		pattern = ^pattern
	}

	for i, pin := range s.segmentPins {
		if (pattern & (1 << i)) != 0 {
			pin.High()
		} else {
			pin.Low()
		}
	}
}

// hardwarePWM is a hardware controlled PWM that sets the segments on the
// display with the according brightness.
// func (s *SevSeg) hardwarePWM() {
// 	s.enabled = s.brightness > 0

// 	if !s.enabled {
// 		return
// 	}

// 	// FIXME:
// 	if s.currentDigitToRefresh < uint8(len(s.digitPins)) {
// 		if channelMap, exists := s.pwmChannels[s.digitPins[s.currentDigitToRefresh]]; exists {
// 			duty := (channelMap.pwm.Top() * uint32(s.brightness)) / 100
// 			channelMap.pwm.Set(channelMap.channel, duty)
// 		}
// 	}
// }

// softwarePWM is a software controlled PWM that sets the segments on the
// display with the according brightness.
func (s *SevSeg) softwarePWM() {
	const pwmPeriod = uint8(10)

	s.pwmCounter = (s.pwmCounter + 1) % pwmPeriod

	// Enable display only during "on" portion of PWM cycle
	// Special cases: 0 = always off, 10 = always on
	brightnessLevel := (s.brightness + 9) / 10
	s.enabled = brightnessLevel > 0 && (brightnessLevel >= 10 || s.pwmCounter < brightnessLevel)
}

// updateDisplayFromPatterns updates the display buffer from the text pattern.
func (s *SevSeg) updateDisplayFromPatterns() {
	displayWidth := len(s.digitPins)
	patternLength := len(s.textPattern)

	if patternLength > displayWidth {
		for i := range displayWidth {
			patternIndex := (int(s.scrollPosition) + i) % patternLength
			s.updatedDisplay[displayWidth-1-i] = s.textPattern[patternIndex]
		}

		return
	}

	blankPattern := s.getSegmentCode(36) // BLANK
	for i := range displayWidth {
		if i < patternLength {
			s.updatedDisplay[displayWidth-1-i] = s.textPattern[i]
		} else {
			s.updatedDisplay[displayWidth-1-i] = blankPattern
		}
	}
}

// getSegmentCode returns the segment code for a given index.
func (s *SevSeg) getSegmentCode(index uint8) uint8 {
	codes := []uint8{
		// GFEDCBA   Index   ASCII   Symbol   7-segment map:
		0b00111111, // 0       0      '0'          AAA
		0b0000110,  // 1       1      '1'         F   B
		0b01011011, // 2       2      '2'         F   B
		0b01001111, // 3       3      '3'          GGG
		0b01100110, // 4       4      '4'         E   C
		0b01101101, // 5       5      '5'         E   C
		0b01111101, // 6       6      '6'          DDD
		0b00000111, // 7       7      '7'
		0b01111111, // 8       8      '8'
		0b01101111, // 9       9      '9'

		0b01110111, // 10     65      'A'
		0b01111100, // 11     66      'b'
		0b00111001, // 12     67      'C'
		0b01011110, // 13     68      'd'
		0b01111001, // 14     69      'E'
		0b01110001, // 15     70      'F'
		0b00111101, // 16     71      'G'
		0b01110110, // 17     72      'H'
		0b00110000, // 18     73      'I'
		0b00001110, // 19     74      'J'
		0b01110110, // 20     75      'K'  Same as 'H'
		0b00111000, // 21     76      'L'
		0b00000000, // 22     77      'M'  NO DISPLAY
		0b01010100, // 23     78      'n'
		0b00111111, // 24     79      'O'
		0b01110011, // 25     80      'P'
		0b01100111, // 26     81      'q'
		0b01010000, // 27     82      'r'
		0b01101101, // 28     83      'S'
		0b01111000, // 29     84      't'
		0b00111110, // 30     85      'U'
		0b00111110, // 31     86      'V'  Same as 'U'
		0b00000000, // 32     87      'W'  NO DISPLAY
		0b01110110, // 33     88      'X'  Same as 'H'
		0b01101110, // 34     89      'y'
		0b01011011, // 35     90      'Z'  Same as '2'

		0b00000000, // 36     32      ' '  BLANK
		0b01000000, // 37     45      '-'  DASH / MINUS
		0b10000000, // 38     46      '.'  PERIOD / DECIMAL POINT
		0b01100011, // 39     42      '°'  DEGREE
		0b00001000, // 40     95      '_'  UNDERSCORE
	}

	return codes[index]
}

// Package width provides display width calculation for Unicode strings.
// This is used for aligning text in terminal output where characters
// may occupy different numbers of columns (e.g., CJK characters are
// typically 2 columns wide, while ASCII is 1 column).
//
// The implementation follows Unicode Standard Annex #11 (East Asian Width)
// with simplifications suitable for terminal display.
package width

// RuneWidth returns the number of terminal columns needed to display r.
// Returns 0 for non-printable/combining characters, 1 for narrow characters,
// and 2 for wide characters (CJK, fullwidth, wide emoji).
func RuneWidth(r rune) int {
	// Control characters and non-printables
	if r < 0x20 || (r >= 0x7F && r <= 0x9F) {
		return 0
	}

	// Soft hyphen
	if r == 0x00AD {
		return 0
	}

	// Combining characters (zero width)
	if isCombining(r) {
		return 0
	}

	// Zero-width characters
	if isZeroWidth(r) {
		return 0
	}

	// Wide characters (2 columns)
	if isWide(r) {
		return 2
	}

	// Default: narrow (1 column)
	return 1
}

// StringWidth returns the total display width of s in terminal columns.
func StringWidth(s string) int {
	w := 0
	for _, r := range s {
		w += RuneWidth(r)
	}
	return w
}

// Truncate truncates s to fit within maxWidth display columns.
// If the string exceeds maxWidth, it is truncated and suffix is appended.
// The result (including suffix) will not exceed maxWidth columns.
func Truncate(s string, maxWidth int, suffix string) string {
	sw := StringWidth(s)
	if sw <= maxWidth {
		return s
	}

	suffixWidth := StringWidth(suffix)
	targetWidth := maxWidth - suffixWidth
	if targetWidth <= 0 {
		// Not enough room for suffix, just truncate without it
		targetWidth = maxWidth
		suffix = ""
	}

	var result []rune
	currentWidth := 0
	for _, r := range s {
		rw := RuneWidth(r)
		if currentWidth+rw > targetWidth {
			break
		}
		result = append(result, r)
		currentWidth += rw
	}

	return string(result) + suffix
}

// runeRange represents a range of Unicode code points [lo, hi] inclusive.
type runeRange struct {
	lo, hi rune
}

// inRanges returns true if r falls within any of the given ranges.
func inRanges(r rune, ranges []runeRange) bool {
	for _, rng := range ranges {
		if r >= rng.lo && r <= rng.hi {
			return true
		}
	}
	return false
}

// combiningRanges defines Unicode ranges for combining diacritical marks
// and similar characters that overlay the previous character (zero width).
var combiningRanges = []runeRange{
	{0x0300, 0x036F}, // Combining Diacritical Marks
	{0x1AB0, 0x1AFF}, // Combining Diacritical Marks Extended
	{0x1DC0, 0x1DFF}, // Combining Diacritical Marks Supplement
	{0x20D0, 0x20FF}, // Combining Diacritical Marks for Symbols
	{0xFE20, 0xFE2F}, // Combining Half Marks
}

// isCombining returns true for combining diacritical marks and similar
// characters that overlay the previous character (zero width).
func isCombining(r rune) bool {
	return inRanges(r, combiningRanges)
}

// zeroWidthRanges defines Unicode ranges for zero-width characters.
var zeroWidthRanges = []runeRange{
	{0x200B, 0x200F},   // Zero-width space, non-joiner, joiner
	{0x2028, 0x202F},   // Line/paragraph separators, format chars
	{0x2060, 0x206F},   // Invisible format indicators
	{0xFE00, 0xFE0F},   // Variation selectors
	{0xFEFF, 0xFEFF},   // Byte order mark
	{0xE0100, 0xE01EF}, // Variation selectors supplement
}

// isZeroWidth returns true for characters that take zero terminal columns.
func isZeroWidth(r rune) bool {
	return inRanges(r, zeroWidthRanges)
}

// wideRanges defines Unicode ranges for characters that occupy 2 terminal columns.
// This includes CJK characters, fullwidth forms, and wide emoji.
var wideRanges = []runeRange{
	// Hangul Jamo
	{0x1100, 0x115F},

	// Miscellaneous symbols (wide emoji subset)
	{0x231A, 0x231B}, // watch, hourglass
	{0x2329, 0x232A}, // angle brackets
	{0x23E9, 0x23F3}, // media control symbols
	{0x23F8, 0x23FA},

	// Misc symbols
	{0x25FD, 0x25FE},
	{0x2614, 0x2615}, // umbrella, hot beverage
	{0x2648, 0x2653}, // zodiac
	{0x267F, 0x267F}, // wheelchair
	{0x2693, 0x2693}, // anchor
	{0x26A1, 0x26A1}, // high voltage
	{0x26AA, 0x26AB}, // circles
	{0x26BD, 0x26BE}, // soccer, baseball
	{0x26C4, 0x26C5}, // snowman, sun
	{0x26CE, 0x26CE}, // ophiuchus
	{0x26D4, 0x26D4}, // no entry
	{0x26EA, 0x26EA}, // church
	{0x26F2, 0x26F3}, // fountain, golf
	{0x26F5, 0x26F5}, // sailboat
	{0x26FA, 0x26FA}, // tent
	{0x26FD, 0x26FD}, // fuel pump
	{0x2705, 0x2705}, // check mark
	{0x270A, 0x270B}, // fist, hand
	{0x2728, 0x2728}, // sparkles
	{0x274C, 0x274C}, // cross mark
	{0x274E, 0x274E}, // cross mark
	{0x2753, 0x2755}, // question marks
	{0x2757, 0x2757}, // exclamation
	{0x2795, 0x2797}, // math symbols
	{0x27B0, 0x27B0}, // curly loop
	{0x27BF, 0x27BF}, // curly loop
	{0x2B1B, 0x2B1C}, // squares
	{0x2B50, 0x2B50}, // star
	{0x2B55, 0x2B55}, // circle

	// CJK Radicals Supplement through CJK Unified Ideographs
	{0x2E80, 0x2EF3},
	{0x2F00, 0x2FD5}, // Kangxi Radicals
	{0x2FF0, 0x2FFF}, // Ideographic Description
	{0x3000, 0x303E}, // CJK Symbols and Punctuation
	{0x3041, 0x3096}, // Hiragana
	{0x3099, 0x30FF}, // Hiragana/Katakana
	{0x3105, 0x312F}, // Bopomofo
	{0x3131, 0x318E}, // Hangul Compatibility Jamo
	{0x3190, 0x31E3}, // Kanbun, Bopomofo Extended
	{0x31F0, 0x321E}, // Katakana Phonetic Extensions, Enclosed CJK
	{0x3220, 0x3247},
	{0x3250, 0x4DBF}, // CJK blocks
	{0x4E00, 0x9FFF}, // CJK Unified Ideographs
	{0xA960, 0xA97F}, // Hangul Jamo Extended-A
	{0xAC00, 0xD7A3}, // Hangul Syllables

	// CJK Compatibility Ideographs
	{0xF900, 0xFAFF},

	// Vertical Forms, CJK Compatibility Forms
	{0xFE10, 0xFE19},
	{0xFE30, 0xFE6F},

	// Fullwidth Forms
	{0xFF01, 0xFF60},
	{0xFFE0, 0xFFE6},

	// CJK extensions and supplements (Plane 2)
	{0x1F300, 0x1F5FF}, // Misc Symbols and Pictographs
	{0x1F600, 0x1F64F}, // Emoticons
	{0x1F680, 0x1F6FF}, // Transport and Map Symbols
	{0x1F700, 0x1F77F}, // Alchemical Symbols
	{0x1F780, 0x1F7FF}, // Geometric Shapes Extended
	{0x1F800, 0x1F8FF}, // Supplemental Arrows-C
	{0x1F900, 0x1F9FF}, // Supplemental Symbols and Pictographs
	{0x1FA00, 0x1FA6F}, // Chess Symbols
	{0x1FA70, 0x1FAFF}, // Symbols and Pictographs Extended-A

	// CJK Unified Ideographs Extension B through G (Plane 2)
	{0x20000, 0x2FFFD},
	// Plane 3
	{0x30000, 0x3FFFD},
}

// isWide returns true for characters that occupy 2 terminal columns.
// This includes CJK characters, fullwidth forms, and wide emoji.
func isWide(r rune) bool {
	return inRanges(r, wideRanges)
}

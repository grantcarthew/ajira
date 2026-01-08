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

// isCombining returns true for combining diacritical marks and similar
// characters that overlay the previous character (zero width).
func isCombining(r rune) bool {
	// Combining Diacritical Marks
	if r >= 0x0300 && r <= 0x036F {
		return true
	}
	// Combining Diacritical Marks Extended
	if r >= 0x1AB0 && r <= 0x1AFF {
		return true
	}
	// Combining Diacritical Marks Supplement
	if r >= 0x1DC0 && r <= 0x1DFF {
		return true
	}
	// Combining Diacritical Marks for Symbols
	if r >= 0x20D0 && r <= 0x20FF {
		return true
	}
	// Combining Half Marks
	if r >= 0xFE20 && r <= 0xFE2F {
		return true
	}
	return false
}

// isZeroWidth returns true for characters that take zero terminal columns.
func isZeroWidth(r rune) bool {
	// Zero-width space, non-joiner, joiner
	if r >= 0x200B && r <= 0x200F {
		return true
	}
	// Line/paragraph separators, format chars
	if r >= 0x2028 && r <= 0x202F {
		return true
	}
	// Invisible format indicators
	if r >= 0x2060 && r <= 0x206F {
		return true
	}
	// Byte order mark
	if r == 0xFEFF {
		return true
	}
	// Variation selectors
	if r >= 0xFE00 && r <= 0xFE0F {
		return true
	}
	// Variation selectors supplement
	if r >= 0xE0100 && r <= 0xE01EF {
		return true
	}
	return false
}

// isWide returns true for characters that occupy 2 terminal columns.
// This includes CJK characters, fullwidth forms, and wide emoji.
func isWide(r rune) bool {
	// Hangul Jamo
	if r >= 0x1100 && r <= 0x115F {
		return true
	}

	// Miscellaneous symbols (wide emoji subset)
	if r == 0x231A || r == 0x231B { // watch, hourglass
		return true
	}
	if r == 0x2329 || r == 0x232A { // angle brackets
		return true
	}
	if r >= 0x23E9 && r <= 0x23F3 { // media control symbols
		return true
	}
	if r >= 0x23F8 && r <= 0x23FA {
		return true
	}

	// Misc symbols
	if r == 0x25FD || r == 0x25FE {
		return true
	}
	if r >= 0x2614 && r <= 0x2615 { // umbrella, hot beverage
		return true
	}
	if r >= 0x2648 && r <= 0x2653 { // zodiac
		return true
	}
	if r == 0x267F { // wheelchair
		return true
	}
	if r == 0x2693 { // anchor
		return true
	}
	if r == 0x26A1 { // high voltage
		return true
	}
	if r == 0x26AA || r == 0x26AB { // circles
		return true
	}
	if r >= 0x26BD && r <= 0x26BE { // soccer, baseball
		return true
	}
	if r >= 0x26C4 && r <= 0x26C5 { // snowman, sun
		return true
	}
	if r == 0x26CE { // ophiuchus
		return true
	}
	if r == 0x26D4 { // no entry
		return true
	}
	if r == 0x26EA { // church
		return true
	}
	if r >= 0x26F2 && r <= 0x26F3 { // fountain, golf
		return true
	}
	if r == 0x26F5 { // sailboat
		return true
	}
	if r == 0x26FA { // tent
		return true
	}
	if r == 0x26FD { // fuel pump
		return true
	}
	if r == 0x2705 { // check mark
		return true
	}
	if r >= 0x270A && r <= 0x270B { // fist, hand
		return true
	}
	if r == 0x2728 { // sparkles
		return true
	}
	if r == 0x274C || r == 0x274E { // cross marks
		return true
	}
	if r >= 0x2753 && r <= 0x2755 { // question marks
		return true
	}
	if r == 0x2757 { // exclamation
		return true
	}
	if r >= 0x2795 && r <= 0x2797 { // math symbols
		return true
	}
	if r == 0x27B0 || r == 0x27BF { // curly loop
		return true
	}
	if r >= 0x2B1B && r <= 0x2B1C { // squares
		return true
	}
	if r == 0x2B50 { // star
		return true
	}
	if r == 0x2B55 { // circle
		return true
	}

	// CJK Radicals Supplement through CJK Unified Ideographs
	if r >= 0x2E80 && r <= 0x2EF3 {
		return true
	}
	if r >= 0x2F00 && r <= 0x2FD5 { // Kangxi Radicals
		return true
	}
	if r >= 0x2FF0 && r <= 0x2FFF { // Ideographic Description
		return true
	}
	if r >= 0x3000 && r <= 0x303E { // CJK Symbols and Punctuation
		return true
	}
	if r >= 0x3041 && r <= 0x3096 { // Hiragana
		return true
	}
	if r >= 0x3099 && r <= 0x30FF { // Hiragana/Katakana
		return true
	}
	if r >= 0x3105 && r <= 0x312F { // Bopomofo
		return true
	}
	if r >= 0x3131 && r <= 0x318E { // Hangul Compatibility Jamo
		return true
	}
	if r >= 0x3190 && r <= 0x31E3 { // Kanbun, Bopomofo Extended
		return true
	}
	if r >= 0x31F0 && r <= 0x321E { // Katakana Phonetic Extensions, Enclosed CJK
		return true
	}
	if r >= 0x3220 && r <= 0x3247 {
		return true
	}
	if r >= 0x3250 && r <= 0x4DBF { // CJK blocks
		return true
	}
	if r >= 0x4E00 && r <= 0x9FFF { // CJK Unified Ideographs
		return true
	}
	if r >= 0xA960 && r <= 0xA97F { // Hangul Jamo Extended-A
		return true
	}
	if r >= 0xAC00 && r <= 0xD7A3 { // Hangul Syllables
		return true
	}

	// CJK Compatibility Ideographs
	if r >= 0xF900 && r <= 0xFAFF {
		return true
	}

	// Vertical Forms, CJK Compatibility Forms
	if r >= 0xFE10 && r <= 0xFE19 {
		return true
	}
	if r >= 0xFE30 && r <= 0xFE6F {
		return true
	}

	// Fullwidth Forms
	if r >= 0xFF01 && r <= 0xFF60 {
		return true
	}
	if r >= 0xFFE0 && r <= 0xFFE6 {
		return true
	}

	// CJK extensions and supplements (Plane 2)
	if r >= 0x1F300 && r <= 0x1F5FF { // Misc Symbols and Pictographs
		return true
	}
	if r >= 0x1F600 && r <= 0x1F64F { // Emoticons
		return true
	}
	if r >= 0x1F680 && r <= 0x1F6FF { // Transport and Map Symbols
		return true
	}
	if r >= 0x1F700 && r <= 0x1F77F { // Alchemical Symbols
		return true
	}
	if r >= 0x1F780 && r <= 0x1F7FF { // Geometric Shapes Extended
		return true
	}
	if r >= 0x1F800 && r <= 0x1F8FF { // Supplemental Arrows-C
		return true
	}
	if r >= 0x1F900 && r <= 0x1F9FF { // Supplemental Symbols and Pictographs
		return true
	}
	if r >= 0x1FA00 && r <= 0x1FA6F { // Chess Symbols
		return true
	}
	if r >= 0x1FA70 && r <= 0x1FAFF { // Symbols and Pictographs Extended-A
		return true
	}

	// CJK Unified Ideographs Extension B through G (Plane 2)
	if r >= 0x20000 && r <= 0x2FFFD {
		return true
	}
	// Plane 3
	if r >= 0x30000 && r <= 0x3FFFD {
		return true
	}

	return false
}

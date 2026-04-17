package width

import (
	"strings"
	"testing"
)

func TestRuneWidth_ASCII(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		{"space", ' ', 1},
		{"letter a", 'a', 1},
		{"letter Z", 'Z', 1},
		{"digit 0", '0', 1},
		{"digit 9", '9', 1},
		{"exclamation", '!', 1},
		{"tilde", '~', 1},
		{"at sign", '@', 1},
		{"hash", '#', 1},
		{"dollar", '$', 1},
		{"percent", '%', 1},
		{"ampersand", '&', 1},
		{"asterisk", '*', 1},
		{"open paren", '(', 1},
		{"close paren", ')', 1},
		{"hyphen", '-', 1},
		{"underscore", '_', 1},
		{"equals", '=', 1},
		{"plus", '+', 1},
		{"open bracket", '[', 1},
		{"close bracket", ']', 1},
		{"open brace", '{', 1},
		{"close brace", '}', 1},
		{"pipe", '|', 1},
		{"backslash", '\\', 1},
		{"semicolon", ';', 1},
		{"colon", ':', 1},
		{"single quote", '\'', 1},
		{"double quote", '"', 1},
		{"comma", ',', 1},
		{"period", '.', 1},
		{"less than", '<', 1},
		{"greater than", '>', 1},
		{"slash", '/', 1},
		{"question mark", '?', 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RuneWidth(tt.r); got != tt.want {
				t.Errorf("RuneWidth(%q) = %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}

func TestRuneWidth_ControlCharacters(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		{"null", 0x00, 0},
		{"bell", 0x07, 0},
		{"backspace", 0x08, 0},
		{"tab", 0x09, 0},
		{"newline", 0x0A, 0},
		{"carriage return", 0x0D, 0},
		{"escape", 0x1B, 0},
		{"unit separator", 0x1F, 0},
		{"DEL", 0x7F, 0},
		{"C1 control start", 0x80, 0},
		{"C1 control end", 0x9F, 0},
		{"soft hyphen", 0x00AD, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RuneWidth(tt.r); got != tt.want {
				t.Errorf("RuneWidth(0x%04X) = %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}

func TestRuneWidth_CombiningCharacters(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		{"combining grave", 0x0300, 0},
		{"combining acute", 0x0301, 0},
		{"combining circumflex", 0x0302, 0},
		{"combining tilde", 0x0303, 0},
		{"combining macron", 0x0304, 0},
		{"combining diaeresis", 0x0308, 0},
		{"combining ring above", 0x030A, 0},
		{"combining caron", 0x030C, 0},
		{"combining cedilla", 0x0327, 0},
		{"combining end of block", 0x036F, 0},
		// Extended combining marks
		{"combining extended start", 0x1AB0, 0},
		{"combining extended end", 0x1AFF, 0},
		// Supplement
		{"combining supplement start", 0x1DC0, 0},
		// For symbols
		{"combining symbols start", 0x20D0, 0},
		// Half marks
		{"combining half marks start", 0xFE20, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RuneWidth(tt.r); got != tt.want {
				t.Errorf("RuneWidth(0x%04X) = %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}

func TestRuneWidth_ZeroWidthCharacters(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		{"zero-width space", 0x200B, 0},
		{"zero-width non-joiner", 0x200C, 0},
		{"zero-width joiner", 0x200D, 0},
		{"left-to-right mark", 0x200E, 0},
		{"right-to-left mark", 0x200F, 0},
		{"line separator", 0x2028, 0},
		{"paragraph separator", 0x2029, 0},
		{"word joiner", 0x2060, 0},
		{"function application", 0x2061, 0},
		{"invisible times", 0x2062, 0},
		{"invisible separator", 0x2063, 0},
		{"byte order mark", 0xFEFF, 0},
		{"variation selector 1", 0xFE00, 0},
		{"variation selector 16", 0xFE0F, 0},
		{"variation selector supplement start", 0xE0100, 0},
		{"variation selector supplement end", 0xE01EF, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RuneWidth(tt.r); got != tt.want {
				t.Errorf("RuneWidth(0x%04X) = %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}

func TestRuneWidth_CJK(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		// Hiragana
		{"hiragana a", 'あ', 2},
		{"hiragana i", 'い', 2},
		{"hiragana u", 'う', 2},
		{"hiragana ka", 'か', 2},
		{"hiragana n", 'ん', 2},
		// Katakana
		{"katakana a", 'ア', 2},
		{"katakana i", 'イ', 2},
		{"katakana ka", 'カ', 2},
		{"katakana n", 'ン', 2},
		// CJK Ideographs
		{"kanji one", '一', 2},
		{"kanji two", '二', 2},
		{"kanji three", '三', 2},
		{"kanji mountain", '山', 2},
		{"kanji river", '川', 2},
		{"kanji sun/day", '日', 2},
		{"kanji moon/month", '月', 2},
		{"kanji fire", '火', 2},
		{"kanji water", '水', 2},
		{"kanji tree/wood", '木', 2},
		{"kanji gold/metal", '金', 2},
		{"kanji soil/earth", '土', 2},
		{"kanji person", '人', 2},
		{"kanji big", '大', 2},
		{"kanji small", '小', 2},
		// Chinese
		{"chinese ni", '你', 2},
		{"chinese hao", '好', 2},
		{"chinese wo", '我', 2},
		{"chinese shi", '是', 2},
		{"chinese de", '的', 2},
		// Hangul
		{"hangul ga", '가', 2},
		{"hangul na", '나', 2},
		{"hangul da", '다', 2},
		{"hangul annyeong", '안', 2},
		{"hangul nyeong", '녕', 2},
		// Fullwidth forms
		{"fullwidth A", 'Ａ', 2},
		{"fullwidth Z", 'Ｚ', 2},
		{"fullwidth 0", '０', 2},
		{"fullwidth 9", '９', 2},
		{"fullwidth exclamation", '！', 2},
		// CJK punctuation
		{"ideographic comma", '、', 2},
		{"ideographic full stop", '。', 2},
		{"left corner bracket", '「', 2},
		{"right corner bracket", '」', 2},
		{"ideographic space", '　', 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RuneWidth(tt.r); got != tt.want {
				t.Errorf("RuneWidth(%q) = %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}

func TestRuneWidth_Emoji(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		// Common emoji
		{"grinning face", '😀', 2},
		{"fire", '🔥', 2},
		{"heart", '❤', 1}, // This is a text presentation emoji, width 1
		{"red heart", '💖', 2},
		{"thumbs up", '👍', 2},
		{"clapping hands", '👏', 2},
		{"folded hands", '🙏', 2},
		{"check mark", '✅', 2},
		{"cross mark", '❌', 2},
		{"warning", '⚠', 1}, // Text presentation
		{"star", '⭐', 2},
		{"sparkles", '✨', 2},
		{"rocket", '🚀', 2},
		{"party popper", '🎉', 2},
		{"trophy", '🏆', 2},
		{"bug", '🐛', 2},
		{"wrench", '🔧', 2},
		{"hammer", '🔨', 2},
		{"gear", '⚙', 1}, // Text presentation
		{"magnifying glass", '🔍', 2},
		{"light bulb", '💡', 2},
		{"memo", '📝', 2},
		{"calendar", '📅', 2},
		{"clock", '🕐', 2},
		{"hourglass", '⏳', 2},
		{"watch", '⌚', 2},
		// Transport
		{"car", '🚗', 2},
		{"airplane", '✈', 1}, // Text presentation
		{"ship", '🚢', 2},
		// Nature
		{"sun", '☀', 1},       // Text presentation
		{"cloud", '☁', 1},     // Text presentation
		{"umbrella", '☂', 1},  // Text presentation
		{"snowflake", '❄', 1}, // Text presentation
		{"rainbow", '🌈', 2},
		// Food
		{"pizza", '🍕', 2},
		{"hamburger", '🍔', 2},
		{"coffee", '☕', 2},
		{"beer", '🍺', 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RuneWidth(tt.r); got != tt.want {
				t.Errorf("RuneWidth(%q U+%04X) = %d, want %d", tt.r, tt.r, got, tt.want)
			}
		})
	}
}

func TestRuneWidth_LatinExtended(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		// Latin-1 Supplement
		{"copyright", '©', 1},
		{"registered", '®', 1},
		{"degree", '°', 1},
		{"plus minus", '±', 1},
		{"micro", 'µ', 1},
		{"pilcrow", '¶', 1},
		// Accented letters (precomposed)
		{"a grave", 'à', 1},
		{"a acute", 'á', 1},
		{"a circumflex", 'â', 1},
		{"a tilde", 'ã', 1},
		{"a umlaut", 'ä', 1},
		{"a ring", 'å', 1},
		{"ae ligature", 'æ', 1},
		{"c cedilla", 'ç', 1},
		{"e grave", 'è', 1},
		{"e acute", 'é', 1},
		{"n tilde", 'ñ', 1},
		{"o umlaut", 'ö', 1},
		{"u umlaut", 'ü', 1},
		{"sharp s", 'ß', 1},
		// Latin Extended-A
		{"a macron", 'ā', 1},
		{"c caron", 'č', 1},
		{"d stroke", 'đ', 1},
		{"e macron", 'ē', 1},
		{"l stroke", 'ł', 1},
		{"o macron", 'ō', 1},
		{"s caron", 'š', 1},
		{"z caron", 'ž', 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RuneWidth(tt.r); got != tt.want {
				t.Errorf("RuneWidth(%q) = %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}

func TestRuneWidth_Symbols(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		// Currency
		{"dollar", '$', 1},
		{"cent", '¢', 1},
		{"pound", '£', 1},
		{"yen", '¥', 1},
		{"euro", '€', 1},
		// Arrows
		{"left arrow", '←', 1},
		{"up arrow", '↑', 1},
		{"right arrow", '→', 1},
		{"down arrow", '↓', 1},
		// Math
		{"infinity", '∞', 1},
		{"not equal", '≠', 1},
		{"less or equal", '≤', 1},
		{"greater or equal", '≥', 1},
		{"approximately", '≈', 1},
		// Box drawing
		{"box light horizontal", '─', 1},
		{"box light vertical", '│', 1},
		{"box light corner", '┌', 1},
		// Miscellaneous
		{"bullet", '•', 1},
		{"ellipsis", '…', 1},
		{"em dash", '—', 1},
		{"en dash", '–', 1},
		{"left single quote", '\u2018', 1},
		{"right single quote", '\u2019', 1},
		{"left double quote", '\u201C', 1},
		{"right double quote", '\u201D', 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RuneWidth(tt.r); got != tt.want {
				t.Errorf("RuneWidth(%q) = %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}

func TestStringWidth_Basic(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"empty", "", 0},
		{"single ASCII", "a", 1},
		{"ASCII word", "hello", 5},
		{"ASCII sentence", "Hello, World!", 13},
		{"numbers", "12345", 5},
		{"mixed ASCII", "abc123!@#", 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringWidth(tt.s); got != tt.want {
				t.Errorf("StringWidth(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestStringWidth_CJK(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"single hiragana", "あ", 2},
		{"hiragana word", "あいう", 6},
		{"single kanji", "日", 2},
		{"kanji word", "日本語", 6},
		{"mixed hiragana kanji", "こんにちは", 10},
		{"hangul word", "안녕하세요", 10},
		{"chinese greeting", "你好", 4},
		{"chinese sentence", "你好世界", 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringWidth(tt.s); got != tt.want {
				t.Errorf("StringWidth(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestStringWidth_Mixed(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"ASCII and kanji", "Hello日本", 9},    // 5 + 4
		{"kanji and ASCII", "日本語ABC", 9},     // 6 + 3
		{"emoji and ASCII", "Hello🔥", 7},     // 5 + 2
		{"ASCII emoji ASCII", "Hi🎉Bye", 7},   // 2 + 2 + 3
		{"fullwidth and ASCII", "ＡＢＣabc", 9}, // 6 + 3
		{"complex mix", "Hello世界🌍!", 12},     // 5 + 4 + 2 + 1
		{"issue key style", "PROJ-123", 8},
		{"issue with CJK summary", "PROJ-123: 日本語タスク", 22}, // 10 + 12
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringWidth(tt.s); got != tt.want {
				t.Errorf("StringWidth(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestStringWidth_WithCombining(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		// Combining characters should not add width
		{"e with combining acute", "e\u0301", 1},         // e + combining acute = é
		{"a with combining ring", "a\u030A", 1},          // a + combining ring = å
		{"n with combining tilde", "n\u0303", 1},         // n + combining tilde = ñ
		{"o with combining umlaut", "o\u0308", 1},        // o + combining umlaut = ö
		{"multiple combining", "e\u0301\u0327", 1},       // e + acute + cedilla
		{"word with combining", "cafe\u0301", 4},         // café
		{"resume with accents", "re\u0301sume\u0301", 6}, // résumé
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringWidth(tt.s); got != tt.want {
				t.Errorf("StringWidth(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestStringWidth_WithZeroWidth(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"with ZWS", "ab\u200Bcd", 4},  // zero-width space
		{"with ZWNJ", "ab\u200Ccd", 4}, // zero-width non-joiner
		{"with ZWJ", "ab\u200Dcd", 4},  // zero-width joiner
		{"with BOM", "\uFEFFhello", 5}, // byte order mark
		{"with word joiner", "hello\u2060world", 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringWidth(tt.s); got != tt.want {
				t.Errorf("StringWidth(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestStringWidth_RealWorldExamples(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		// Jira-like issue summaries
		{"bug report", "Fix login button not working", 28},
		{"feature request", "Add dark mode support", 21},
		{"with emoji prefix", "🐛 Fix null pointer exception", 29},    // 2 + 27
		{"with check emoji", "✅ Completed: Update dependencies", 33}, // 2 + 31
		// Usernames
		{"simple username", "john.doe", 8},
		{"email username", "user@example.com", 16},
		// Japanese issue summary
		{"japanese task", "ログイン機能の修正", 18},
		// Korean issue summary
		{"korean task", "로그인 버그 수정", 16}, // 6 + 1 + 4 + 1 + 4
		// Chinese issue summary
		{"chinese task", "修复登录问题", 12},
		// Mixed language
		{"mixed JP-EN", "Fix バグ in login", 17}, // 4 + 4 + 9
		// Status labels
		{"status done", "Done", 4},
		{"status in progress", "In Progress", 11},
		{"status JP", "完了", 4},
		// Priority
		{"priority high", "High", 4},
		{"priority JP", "高", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringWidth(tt.s); got != tt.want {
				t.Errorf("StringWidth(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestStringWidth_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"only spaces", "     ", 5},
		{"only tabs", "\t\t", 0},         // tabs are control chars
		{"only newlines", "\n\n\n", 0},   // newlines are control chars
		{"spaces and tabs", "  \t  ", 4}, // 2 + 0 + 2
		{"long ASCII", strings.Repeat("a", 100), 100},
		{"long CJK", strings.Repeat("日", 50), 100},
		{"alternating", "a日b月c", 7}, // 1+2+1+2+1
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringWidth(tt.s); got != tt.want {
				t.Errorf("StringWidth(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

// Benchmarks
func BenchmarkRuneWidth_ASCII(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RuneWidth('a')
	}
}

func BenchmarkRuneWidth_CJK(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RuneWidth('日')
	}
}

func BenchmarkRuneWidth_Emoji(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RuneWidth('🔥')
	}
}

func BenchmarkStringWidth_Short(b *testing.B) {
	s := "Hello, World!"
	for i := 0; i < b.N; i++ {
		StringWidth(s)
	}
}

func BenchmarkStringWidth_CJK(b *testing.B) {
	s := "日本語テキスト"
	for i := 0; i < b.N; i++ {
		StringWidth(s)
	}
}

func BenchmarkStringWidth_Mixed(b *testing.B) {
	s := "Hello世界🌍Test日本語"
	for i := 0; i < b.N; i++ {
		StringWidth(s)
	}
}

func BenchmarkStringWidth_Long(b *testing.B) {
	s := strings.Repeat("Hello世界🌍", 100)
	for i := 0; i < b.N; i++ {
		StringWidth(s)
	}
}

func TestTruncate_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		suffix   string
		want     string
	}{
		{"empty", "", 10, "...", ""},
		{"shorter_than_max", "hello", 10, "...", "hello"},
		{"exact_length", "hello", 5, "...", "hello"},
		{"needs_truncation", "hello world", 8, "...", "hello..."},
		{"truncate_to_zero", "hello", 0, "...", ""},
		{"suffix_too_long", "hello", 2, "...", "he"},
		{"no_suffix", "hello world", 5, "", "hello"},
		{"single_char_suffix", "hello world", 6, ".", "hello."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxWidth, tt.suffix)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d, %q) = %q, want %q",
					tt.input, tt.maxWidth, tt.suffix, got, tt.want)
			}
			// Verify result doesn't exceed maxWidth
			gotWidth := StringWidth(got)
			if gotWidth > tt.maxWidth {
				t.Errorf("Truncate result width %d exceeds maxWidth %d", gotWidth, tt.maxWidth)
			}
		})
	}
}

func TestTruncate_CJK(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		suffix   string
		want     string
	}{
		{"CJK_no_truncation", "日本語", 6, "...", "日本語"},
		{"CJK_truncate_one", "日本語テスト", 9, "...", "日本語..."},
		{"CJK_truncate_to_fit", "日本語テスト", 8, "...", "日本..."},
		{"CJK_mixed_ASCII", "Hello世界", 10, "...", "Hello世界"},
		{"CJK_mixed_truncate", "Hello世界Test", 10, "...", "Hello世..."},
		{"hangul", "안녕하세요", 7, "...", "안녕..."},
		{"chinese", "你好世界", 5, "...", "你..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxWidth, tt.suffix)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d, %q) = %q, want %q",
					tt.input, tt.maxWidth, tt.suffix, got, tt.want)
			}
			gotWidth := StringWidth(got)
			if gotWidth > tt.maxWidth {
				t.Errorf("Truncate result width %d exceeds maxWidth %d", gotWidth, tt.maxWidth)
			}
		})
	}
}

func TestTruncate_Emoji(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		suffix   string
		want     string
	}{
		{"emoji_fits", "🔥🚀", 4, "...", "🔥🚀"},
		{"emoji_truncate", "🔥🚀✨⭐", 5, "...", "🔥..."},
		{"emoji_with_text", "Bug🐛Fix", 8, "...", "Bug🐛Fix"},
		{"text_then_emoji", "Test🔥", 6, "...", "Test🔥"},
		{"text_then_emoji_truncate", "Testing🔥", 7, "...", "Test..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxWidth, tt.suffix)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d, %q) = %q, want %q",
					tt.input, tt.maxWidth, tt.suffix, got, tt.want)
			}
			gotWidth := StringWidth(got)
			if gotWidth > tt.maxWidth {
				t.Errorf("Truncate result width %d exceeds maxWidth %d", gotWidth, tt.maxWidth)
			}
		})
	}
}

func TestTruncate_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		suffix   string
		wantMax  int // verify result is at most this width
	}{
		{"wide_char_boundary", "あいうえお", 7, "...", 7},
		{"wide_char_exact", "あいう", 6, "...", 6},
		{"combining_chars", "e\u0301e\u0301e\u0301", 3, "...", 3},
		{"zero_width_chars", "a\u200Bb\u200Bc", 3, "...", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxWidth, tt.suffix)
			gotWidth := StringWidth(got)
			if gotWidth > tt.wantMax {
				t.Errorf("Truncate(%q, %d, %q) width = %d, want <= %d",
					tt.input, tt.maxWidth, tt.suffix, gotWidth, tt.wantMax)
			}
		})
	}
}

func TestTruncate_RealWorld(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		suffix   string
	}{
		{"issue_summary_EN", "Fix authentication bug in login page", 60, "..."},
		{"issue_summary_JP", "ログイン機能のバグを修正する必要があります", 60, "..."},
		{"issue_summary_mixed", "Fix the 日本語 encoding issue in exports", 50, "..."},
		{"long_summary", "This is a very long issue summary that definitely needs to be truncated to fit within the display width", 60, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxWidth, tt.suffix)
			gotWidth := StringWidth(got)
			if gotWidth > tt.maxWidth {
				t.Errorf("Truncate result width %d exceeds maxWidth %d for %q",
					gotWidth, tt.maxWidth, tt.input)
			}
			// If original fits, should be unchanged
			if StringWidth(tt.input) <= tt.maxWidth && got != tt.input {
				t.Errorf("Truncate modified string that already fits: got %q, want %q",
					got, tt.input)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  string
	}{
		{"ascii_shorter", "abc", 6, "abc   "},
		{"ascii_equal", "abcdef", 6, "abcdef"},
		{"ascii_longer", "abcdefgh", 6, "abcdefgh"},
		{"empty_padded", "", 4, "    "},
		{"empty_zero_width", "", 0, ""},
		{"zero_width_target", "abc", 0, "abc"},
		{"negative_width", "abc", -1, "abc"},
		{"cjk_shorter", "日本", 6, "日本  "},
		{"cjk_equal", "日本語", 6, "日本語"},
		{"cjk_longer", "日本語字", 6, "日本語字"},
		{"mixed_shorter", "a日b", 6, "a日b  "},
		{"emoji", "🎉", 4, "🎉  "},
		{"single_space", "a", 1, "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PadRight(tt.input, tt.width)
			if got != tt.want {
				t.Errorf("PadRight(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
			}
			// Result width must be at least the target (unless input already exceeds it).
			gotWidth := StringWidth(got)
			inputWidth := StringWidth(tt.input)
			if inputWidth < tt.width && gotWidth != tt.width {
				t.Errorf("PadRight(%q, %d) result width = %d, want %d",
					tt.input, tt.width, gotWidth, tt.width)
			}
		})
	}
}

func TestPadRight_PreservesPrefix(t *testing.T) {
	got := PadRight("hello", 10)
	if !strings.HasPrefix(got, "hello") {
		t.Errorf("PadRight result %q does not start with input", got)
	}
	if len(got) != 10 {
		t.Errorf("PadRight ASCII result length = %d, want 10", len(got))
	}
}

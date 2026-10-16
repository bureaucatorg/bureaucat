package mattermostbot

import "testing"

func TestSanitize(t *testing.T) {
	const app = "https://bc.example.com/"
	tests := []struct{ name, in, want string }{
		{"relative link made absolute", "see [ENG-42](/projects/ENG/tasks/42)", "see [ENG-42](https://bc.example.com/projects/ENG/tasks/42)"},
		{"absolute app link kept", "[x](https://bc.example.com/projects)", "[x](https://bc.example.com/projects)"},
		{"external link reduced to label", "[docs](https://evil.example/?d=1)", "docs"},
		{"lookalike host is external", "[x](https://bc.example.com.evil.example/a)", "x"},
		{"protocol-relative is external", "[x](//evil.example/a)", "x"},
		{"image becomes alt text", "![secret](https://evil.example/p.png?d=1)", "secret"},
		{"app image still dropped", "![logo](/logo.png)", "logo"},
		{"bare external url defused", "go to https://evil.example/x now", "go to `https://evil.example/x` now"},
		{"www url defused", "www.evil.example", "`www.evil.example`"},
		{"url inside dropped label defused", "[https://evil.example](https://evil.example)", "`https://evil.example`"},
		{"bare app url kept", "https://bc.example.com/projects/ENG", "https://bc.example.com/projects/ENG"},
		{"mass mentions defused", "hey @channel and @ALL", "hey `@channel` and `@ALL`"},
		{"mention in app link label defused", "[@here](/projects)", "[`@here`](https://bc.example.com/projects)"},
		{"normal mention untouched", "ask @sam", "ask @sam"},
		{"blank line before table", "ENG\n| A | B |\n|---|---|\n| 1 | 2 |", "ENG\n\n| A | B |\n|---|---|\n| 1 | 2 |"},
		{"table already separated", "ENG\n\n| A | B |\n| :-- | --: |", "ENG\n\n| A | B |\n| :-- | --: |"},
		{"table at start untouched", "| A | B |\n|---|---|", "| A | B |\n|---|---|"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitize(tt.in, app); got != tt.want {
				t.Errorf("sanitize(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDecision(t *testing.T) {
	for in, want := range map[string][2]bool{
		"yes": {true, true}, "Y": {true, true}, " Approve! ": {true, true},
		"no": {false, true}, "Reject.": {false, true},
		"yes please delete it": {false, false}, "": {false, false},
	} {
		approve, ok := decision(in)
		if approve != want[0] || ok != want[1] {
			t.Errorf("decision(%q) = %v, %v; want %v, %v", in, approve, ok, want[0], want[1])
		}
	}
}

package suggest

import "testing"

func TestParseContext(t *testing.T) {
	context, ok := ParseContext("git branch --sho", 16)
	if !ok {
		t.Fatal("ParseContext rejected a valid cursor")
	}

	if got, want := context.Prefix, "--sho"; got != want {
		t.Fatalf("Prefix = %q, want %q", got, want)
	}
	if got, want := context.TokenStart, 11; got != want {
		t.Fatalf("TokenStart = %d, want %d", got, want)
	}
	if got, want := context.TokenEnd, 16; got != want {
		t.Fatalf("TokenEnd = %d, want %d", got, want)
	}
	if got, want := len(context.WordsBefore), 2; got != want {
		t.Fatalf("len(WordsBefore) = %d, want %d", got, want)
	}
}

func TestSuggest(t *testing.T) {
	tests := []struct {
		name string
		line string
		want []string
	}{
		{name: "ordinary command", line: "go test ", want: nil},
		{name: "git subcommand", line: "git br", want: []string{"branch"}},
		{name: "branch option", line: "git branch --sh", want: []string{"--show-current"}},
		{name: "all branch options", line: "git branch --", want: []string{"--show-current", "--merged", "--no-merged", "--delete"}},
		{name: "unknown option", line: "git branch --unknown", want: nil},
		{name: "completed subcommand", line: "git status", want: nil},
		{name: "completed option", line: "git status --short", want: nil},
		{name: "exclude used option", line: "git status --short ", want: []string{"--branch", "--porcelain"}},
		{name: "exclude used option with prefix", line: "git status --short --", want: []string{"--branch", "--porcelain"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))
			if len(got) != len(test.want) {
				t.Fatalf("Suggest() returned %d values, want %d: %#v", len(got), len(test.want), got)
			}
			for index, want := range test.want {
				if got[index].Value != want {
					t.Fatalf("suggestion %d = %q, want %q", index, got[index].Value, want)
				}
			}
		})
	}
}

func TestSuggestUsesCursorPrefix(t *testing.T) {
	line := "git branch --show-current"
	got := Suggest(line, len([]rune("git branch --sh")))
	if len(got) != 1 || got[0].Value != "--show-current" {
		t.Fatalf("Suggest() = %#v, want --show-current", got)
	}
}

func TestSuggestsCommonSubcommands(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git stat", want: "status"},
		{line: "git com", want: "commit"},
		{line: "git swi", want: "switch"},
		{line: "git reb", want: "rebase"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(
				test.line,
				len([]rune(test.line)),
			)

			if len(got) != 1 {
				t.Fatalf(
					"Suggest(%q) returned %d candidates: %#v",
					test.line,
					len(got),
					got,
				)
			}
			if got[0].Value != test.want {
				t.Fatalf(
					"Suggest(%q) = %q, want %q",
					test.line,
					got[0].Value,
					test.want,
				)
			}
		})
	}
}

func TestSuggestsAmbiguousPrefix(t *testing.T) {
	got := Suggest("git sta", len([]rune("git sta")))
	want := []string{"status", "stash"}

	if len(got) != len(want) {
		t.Fatalf("Suggest(\"git sta\") returned %d candidates: %#v", len(got), got)
	}
	for index, value := range want {
		if got[index].Value != value {
			t.Fatalf("candidate %d = %q, want %q", index, got[index].Value, value)
		}
	}
}

func TestSuggestsCommonOptions(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{
			line: "git status --sh",
			want: "--short",
		},
		{
			line: "git commit --am",
			want: "--amend",
		},
		{
			line: "git switch --cr",
			want: "--create",
		},
		{
			line: "git log --gra",
			want: "--graph",
		},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(
				test.line,
				len([]rune(test.line)),
			)

			if len(got) != 1 {
				t.Fatalf(
					"Suggest(%q) returned %d candidates: %#v",
					test.line,
					len(got),
					got,
				)
			}
			if got[0].Value != test.want {
				t.Fatalf(
					"Suggest(%q) = %q, want %q",
					test.line,
					got[0].Value,
					test.want,
				)
			}
		})
	}
}

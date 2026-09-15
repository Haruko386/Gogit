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
		{name: "empty git subcommand", line: "git ", want: nil},
		{name: "git subcommand", line: "git br", want: []string{"branch"}},
		{name: "empty branch option", line: "git branch ", want: nil},
		{name: "branch option", line: "git branch --sh", want: []string{"--show-current"}},
		{name: "all branch options", line: "git branch --", want: []string{"--show-current", "--merged", "--no-merged", "--delete"}},
		{name: "unknown option", line: "git branch --unknown", want: nil},
		{name: "completed subcommand", line: "git status", want: nil},
		{name: "completed option", line: "git status --short", want: nil},
		{name: "empty option after used option", line: "git status --short ", want: nil},
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
		{line: "git clo", want: "clone"},
		{line: "git ini", want: "init"},
		{line: "git rem", want: "remote"},
		{line: "git ta", want: "tag"},
		{line: "git rese", want: "reset"},
		{line: "git reve", want: "revert"},
		{line: "git cherry-p", want: "cherry-pick"},
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
		{
			line: "git clone --dep",
			want: "--depth",
		},
		{
			line: "git init --initial",
			want: "--initial-branch",
		},
		{
			line: "git remote --ver",
			want: "--verbose",
		},
		{
			line: "git tag --ann",
			want: "--annotate",
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

func TestAnalyzeCommitMessageValueHint(t *testing.T) {
	tests := []string{
		"git commit --message",
		"git commit --message ",
		`git commit --message "fix bug`,
		`git commit --message "fix bug"`,
		`git commit --message=fix`,
		`git commit -m "fix"`,
	}

	for _, line := range tests {
		t.Run(line, func(t *testing.T) {
			result := Analyze(line, len([]rune(line)))
			if result.Hint == nil {
				t.Fatalf("Analyze(%q) returned no value hint", line)
			}
			if got, want := result.Hint.Name, "message"; got != want {
				t.Fatalf("hint name = %q, want %q", got, want)
			}
			if len(result.Suggestions) != 0 {
				t.Fatalf("value position returned suggestions: %#v", result.Suggestions)
			}
		})
	}
}

func TestAnalyzeHidesOptionsAfterCommitMessageAtEmptyToken(t *testing.T) {
	tests := []string{
		`git commit --message "fix bug" `,
		`git commit --message="fix bug" `,
		`git commit --message "" `,
	}

	for _, line := range tests {
		t.Run(line, func(t *testing.T) {
			result := Analyze(line, len([]rune(line)))
			if result.Hint != nil {
				t.Fatalf("Analyze(%q) returned unexpected hint: %#v", line, result.Hint)
			}
			if len(result.Suggestions) != 0 {
				t.Fatalf("Analyze(%q) returned suggestions for an empty token: %#v", line, result.Suggestions)
			}
		})
	}
}

func TestAnalyzeSuggestsRepeatableMessageAfterPrefix(t *testing.T) {
	tests := []string{
		`git commit --message "fix bug" --m`,
		`git commit --message="fix bug" --m`,
		`git commit --message "" --m`,
	}

	for _, line := range tests {
		t.Run(line, func(t *testing.T) {
			result := Analyze(line, len([]rune(line)))
			if result.Hint != nil {
				t.Fatalf("Analyze(%q) returned unexpected hint: %#v", line, result.Hint)
			}

			foundRepeatableMessage := false
			for _, candidate := range result.Suggestions {
				if candidate.Value == "--message" {
					foundRepeatableMessage = true
				}
			}
			if !foundRepeatableMessage {
				t.Fatal("repeatable --message was not suggested")
			}
		})
	}
}

func TestAnalyzeSuggestsRepositoryBranches(t *testing.T) {
	branches := []Suggestion{
		{Value: "dev", Kind: KindBranch},
		{Value: "feature/login", Kind: KindBranch},
		{Value: "main", Kind: KindBranch},
		{Value: "origin/dev", Kind: KindBranch},
		{Value: "origin/main", Kind: KindBranch},
	}

	tests := []struct {
		line string
		want []string
	}{
		{line: "git switch fe", want: []string{"feature/login"}},
		{line: "git checkout de", want: []string{"dev"}},
		{line: "git merge ma", want: []string{"main"}},
		{line: "git rebase or", want: []string{"origin/dev", "origin/main"}},
		{line: "git reset --hard or", want: []string{"origin/dev", "origin/main"}},
		{line: "git pull origin ma", want: []string{"main"}},
		{line: "git push origin fe", want: []string{"feature/login"}},
		{line: "git revert ma", want: []string{"main"}},
		{line: "git cherry-pick fe", want: []string{"feature/login"}},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			result := AnalyzeWithBranches(
				test.line,
				len([]rune(test.line)),
				branches,
			)

			if len(result.Suggestions) != len(test.want) {
				t.Fatalf(
					"AnalyzeWithBranches(%q) returned %d candidates, want %d: %#v",
					test.line,
					len(result.Suggestions),
					len(test.want),
					result.Suggestions,
				)
			}
			for index, want := range test.want {
				if result.Suggestions[index].Value != want {
					t.Fatalf(
						"candidate %d = %q, want %q",
						index,
						result.Suggestions[index].Value,
						want,
					)
				}
			}
		})
	}
}

func TestAnalyzeDoesNotSuggestExistingBranchAsNewBranchName(t *testing.T) {
	branches := []Suggestion{{Value: "feature/login", Kind: KindBranch}}
	for _, line := range []string{
		"git switch --create fe",
		"git checkout -b fe",
	} {
		result := AnalyzeWithBranches(line, len([]rune(line)), branches)
		if len(result.Suggestions) != 0 {
			t.Fatalf("AnalyzeWithBranches(%q) = %#v, want no branches", line, result.Suggestions)
		}
	}
}

func TestAnalyzeNewOptionValueHints(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{
			line: "git clone --branch ",
			want: "branch",
		},
		{
			line: "git clone --depth ",
			want: "depth",
		},
		{
			line: "git clone -o ",
			want: "name",
		},
		{
			line: "git init -b ",
			want: "branch",
		},
		{
			line: "git init --template ",
			want: "directory",
		},
		{
			line: "git tag -m ",
			want: "message",
		},
		{
			line: "git tag --contains ",
			want: "commit",
		},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			result := Analyze(
				test.line,
				len([]rune(test.line)),
			)

			if result.Hint == nil {
				t.Fatalf(
					"Analyze(%q) returned no value hint",
					test.line,
				)
			}
			if result.Hint.Name != test.want {
				t.Fatalf(
					"hint name = %q, want %q",
					result.Hint.Name,
					test.want,
				)
			}
			if len(result.Suggestions) != 0 {
				t.Fatalf(
					"value position returned suggestions: %#v",
					result.Suggestions,
				)
			}
		})
	}
}

func TestSuggestsRemoteSubcommands(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{
			line: "git remote ad",
			want: "add",
		},
		{
			line: "git remote ren",
			want: "rename",
		},
		{
			line: "git remote rem",
			want: "remove",
		},
		{
			line: "git remote set-h",
			want: "set-head",
		},
		{
			line: "git remote set-b",
			want: "set-branches",
		},
		{
			line: "git remote get",
			want: "get-url",
		},
		{
			line: "git remote set-u",
			want: "set-url",
		},
		{
			line: "git remote sho",
			want: "show",
		},
		{
			line: "git remote pru",
			want: "prune",
		},
		{
			line: "git remote upd",
			want: "update",
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

func TestRemoteSubcommandsDoNotReplaceRemoteOptions(t *testing.T) {
	got := Suggest(
		"git remote --ver",
		len([]rune("git remote --ver")),
	)

	if len(got) != 1 || got[0].Value != "--verbose" {
		t.Fatalf(
			"Suggest() = %#v, want --verbose",
			got,
		)
	}
}

func TestSuggestsRemoteSubcommandOptions(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{
			line: "git remote add --mir",
			want: "--mirror=",
		},
		{
			line: "git remote rename --pro",
			want: "--progress",
		},
		{
			line: "git remote set-head --au",
			want: "--auto",
		},
		{
			line: "git remote set-branches --ad",
			want: "--add",
		},
		{
			line: "git remote get-url --pu",
			want: "--push",
		},
		{
			line: "git remote set-url --del",
			want: "--delete",
		},
		{
			line: "git remote show --no",
			want: "--no-query",
		},
		{
			line: "git remote prune --dry",
			want: "--dry-run",
		},
		{
			line: "git remote update --pru",
			want: "--prune",
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

func TestAnalyzeRemoteSubcommandOptionValueHints(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{
			line: "git remote add -t ",
			want: "branch",
		},
		{
			line: "git remote add -m ",
			want: "branch",
		},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			result := Analyze(
				test.line,
				len([]rune(test.line)),
			)

			if result.Hint == nil {
				t.Fatalf(
					"Analyze(%q) returned no value hint",
					test.line,
				)
			}

			if result.Hint.Name != test.want {
				t.Fatalf(
					"hint name = %q, want %q",
					result.Hint.Name,
					test.want,
				)
			}

			if len(result.Suggestions) != 0 {
				t.Fatalf(
					"value position returned suggestions: %#v",
					result.Suggestions,
				)
			}
		})
	}
}

func TestRemoteOptionsBeforeSubcommandArePreserved(t *testing.T) {
	got := Suggest(
		"git remote --verbose sho",
		len([]rune("git remote --verbose sho")),
	)

	if len(got) != 1 || got[0].Value != "show" {
		t.Fatalf(
			"Suggest() = %#v, want show",
			got,
		)
	}
}

func TestRemoteSubcommandOptionConflicts(t *testing.T) {
	tests := []string{
		"git remote add --tags --no",
		"git remote set-head -a --del",
		"git remote set-url --add --del",
		"git remote rename --progress --no",
	}

	for _, line := range tests {
		t.Run(line, func(t *testing.T) {
			got := Suggest(line, len([]rune(line)))
			if len(got) != 0 {
				t.Fatalf(
					"Suggest(%q) = %#v, want no conflicting option",
					line,
					got,
				)
			}
		})
	}
}

func TestRemoteAddTrackOptionIsRepeatable(t *testing.T) {
	got := Suggest(
		"git remote add -t main -",
		len([]rune("git remote add -t main -")),
	)

	found := false
	for _, candidate := range got {
		if candidate.Value == "-t" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf(
			"repeatable -t was not suggested: %#v",
			got,
		)
	}
}

func TestSuggestsHistoryEditingOptions(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git reset --har", want: "--hard"},
		{line: "git reset --patc", want: "--patch"},
		{line: "git revert --no-c", want: "--no-commit"},
		{line: "git revert --main", want: "--mainline"},
		{line: "git cherry-pick --f", want: "--ff"},
		{line: "git cherry-pick --emp", want: "--empty"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))

			if len(got) != 1 || got[0].Value != test.want {
				t.Fatalf(
					"Suggest(%q) = %#v, want %q",
					test.line,
					got,
					test.want,
				)
			}
		})
	}
}

func TestAnalyzeHistoryEditingValueHints(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{
			line: "git reset --pathspec-from-file ",
			want: "file",
		},
		{
			line: "git revert -m ",
			want: "parent-number",
		},
		{
			line: "git revert --strategy ",
			want: "strategy",
		},
		{
			line: "git cherry-pick -X ",
			want: "option",
		},
		{
			line: "git cherry-pick --empty=",
			want: "mode",
		},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			result := Analyze(test.line, len([]rune(test.line)))

			if result.Hint == nil {
				t.Fatalf("Analyze(%q) returned no value hint", test.line)
			}
			if result.Hint.Name != test.want {
				t.Fatalf(
					"hint name = %q, want %q",
					result.Hint.Name,
					test.want,
				)
			}
			if len(result.Suggestions) != 0 {
				t.Fatalf(
					"value position returned suggestions: %#v",
					result.Suggestions,
				)
			}
		})
	}
}

func TestHistoryEditingOptionConflicts(t *testing.T) {
	tests := []string{
		"git reset --hard --so",
		"git reset -p --mi",
		"git revert --continue --ab",
		"git revert --edit --no-e",
		"git cherry-pick --skip --con",
		"git cherry-pick --no-edit --ed",
	}

	for _, line := range tests {
		t.Run(line, func(t *testing.T) {
			got := Suggest(line, len([]rune(line)))
			if len(got) != 0 {
				t.Fatalf(
					"Suggest(%q) = %#v, want no conflicting option",
					line,
					got,
				)
			}
		})
	}
}

func TestHistoryEditingStrategyOptionIsRepeatable(t *testing.T) {
	lines := []string{
		"git revert -X ours --str",
		"git cherry-pick --strategy-option=ours --str",
	}

	for _, line := range lines {
		t.Run(line, func(t *testing.T) {
			got := Suggest(line, len([]rune(line)))

			found := false
			for _, candidate := range got {
				if candidate.Value == "--strategy-option" {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf(
					"repeatable --strategy-option was not suggested: %#v",
					got,
				)
			}
		})
	}
}

func TestSequencerActionsDoNotSuggestBranches(t *testing.T) {
	branches := []Suggestion{
		{Value: "main", Kind: KindBranch},
	}

	lines := []string{
		"git revert --continue ma",
		"git cherry-pick --abort ma",
	}

	for _, line := range lines {
		result := AnalyzeWithBranches(
			line,
			len([]rune(line)),
			branches,
		)

		if len(result.Suggestions) != 0 {
			t.Fatalf(
				"AnalyzeWithBranches(%q) = %#v, want no branches",
				line,
				result.Suggestions,
			)
		}
	}
}

func TestSuggestsFileManagementSubcommands(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git cle", want: "clean"},
		{line: "git r", want: "rm"},
		{line: "git m", want: "mv"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))

			found := false
			for _, candidate := range got {
				if candidate.Value == test.want {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf(
					"Suggest(%q) = %#v, missing %q",
					test.line,
					got,
					test.want,
				)
			}
		})
	}
}

func TestSuggestsFileManagementOptions(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git clean --dry", want: "--dry-run"},
		{line: "git clean --inter", want: "--interactive"},
		{line: "git clean --exc", want: "--exclude"},
		{line: "git rm --cach", want: "--cached"},
		{line: "git rm --ignore", want: "--ignore-unmatch"},
		{line: "git rm --pathspec-from", want: "--pathspec-from-file"},
		{line: "git mv --verb", want: "--verbose"},
		{line: "git mv --spa", want: "--sparse"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))

			if len(got) != 1 || got[0].Value != test.want {
				t.Fatalf(
					"Suggest(%q) = %#v, want %q",
					test.line,
					got,
					test.want,
				)
			}
		})
	}
}

func TestAnalyzeFileManagementValueHints(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{
			line: "git clean -e ",
			want: "pattern",
		},
		{
			line: "git rm --pathspec-from-file ",
			want: "file",
		},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			result := Analyze(test.line, len([]rune(test.line)))

			if result.Hint == nil {
				t.Fatalf("Analyze(%q) returned no value hint", test.line)
			}
			if result.Hint.Name != test.want {
				t.Fatalf(
					"hint name = %q, want %q",
					result.Hint.Name,
					test.want,
				)
			}
			if len(result.Suggestions) != 0 {
				t.Fatalf(
					"value position returned suggestions: %#v",
					result.Suggestions,
				)
			}
		})
	}
}

func TestCleanExcludeIsRepeatable(t *testing.T) {
	line := "git clean -e build --ex"
	got := Suggest(line, len([]rune(line)))

	found := false
	for _, candidate := range got {
		if candidate.Value == "--exclude" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("repeatable --exclude was not suggested: %#v", got)
	}
}

func TestCleanModesAreMutuallyExclusive(t *testing.T) {
	tests := []struct {
		line      string
		forbidden string
	}{
		{
			line:      "git clean -x -",
			forbidden: "-X",
		},
		{
			line:      "git clean -X -",
			forbidden: "-x",
		},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))

			for _, candidate := range got {
				if candidate.Value == test.forbidden {
					t.Fatalf(
						"Suggest(%q) returned conflicting option %q: %#v",
						test.line,
						test.forbidden,
						got,
					)
				}
			}
		})
	}
}

func TestSuggestsRebaseOptions(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git rebase --ont", want: "--onto"},
		{line: "git rebase --keep-b", want: "--keep-base"},
		{line: "git rebase --inter", want: "--interactive"},
		{line: "git rebase --autosq", want: "--autosquash"},
		{line: "git rebase --autost", want: "--autostash"},
		{line: "git rebase --update", want: "--update-refs"},
		{line: "git rebase --rebase-m", want: "--rebase-merges"},
		{line: "git rebase --force-r", want: "--force-rebase"},
		{line: "git rebase --edit-t", want: "--edit-todo"},
		{line: "git rebase --show-c", want: "--show-current-patch"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))

			if len(got) != 1 || got[0].Value != test.want {
				t.Fatalf(
					"Suggest(%q) = %#v, want %q",
					test.line,
					got,
					test.want,
				)
			}
		})
	}
}

func TestAnalyzeRebaseValueHints(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git rebase --onto ", want: "revision"},
		{line: "git rebase -x ", want: "command"},
		{line: "git rebase --empty=", want: "mode"},
		{line: "git rebase -s ", want: "strategy"},
		{line: "git rebase -X ", want: "option"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			result := Analyze(test.line, len([]rune(test.line)))

			if result.Hint == nil {
				t.Fatalf("Analyze(%q) returned no value hint", test.line)
			}
			if result.Hint.Name != test.want {
				t.Fatalf(
					"hint name = %q, want %q",
					result.Hint.Name,
					test.want,
				)
			}
			if len(result.Suggestions) != 0 {
				t.Fatalf(
					"value position returned suggestions: %#v",
					result.Suggestions,
				)
			}
		})
	}
}

func TestRebaseOptionConflicts(t *testing.T) {
	tests := []struct {
		line      string
		forbidden string
	}{
		{
			line:      "git rebase --onto main --keep",
			forbidden: "--keep-base",
		},
		{
			line:      "git rebase --keep-base --ont",
			forbidden: "--onto",
		},
		{
			line:      "git rebase --apply --mer",
			forbidden: "--merge",
		},
		{
			line:      "git rebase --autosquash --no-auto",
			forbidden: "--no-autosquash",
		},
		{
			line:      "git rebase --continue --abo",
			forbidden: "--abort",
		},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))

			for _, candidate := range got {
				if candidate.Value == test.forbidden {
					t.Fatalf(
						"Suggest(%q) returned conflicting option %q: %#v",
						test.line,
						test.forbidden,
						got,
					)
				}
			}
		})
	}
}

func TestRebaseOptionsCanRepeatWhereAllowed(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{
			line: "git rebase -x \"go test ./...\" --ex",
			want: "--exec",
		},
		{
			line: "git rebase -X ours --strategy-o",
			want: "--strategy-option",
		},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))

			found := false
			for _, candidate := range got {
				if candidate.Value == test.want {
					found = true
					break
				}
			}

			if !found {
				t.Fatalf(
					"Suggest(%q) = %#v, missing repeatable %q",
					test.line,
					got,
					test.want,
				)
			}
		})
	}
}

func TestRebaseApplyConflictsAreSymmetric(t *testing.T) {
	options := gitOptions["rebase"]
	apply, ok := findOption(options, "--apply")
	if !ok {
		t.Fatal("rebase --apply option is missing")
	}

	conflicts := []string{
		"--strategy",
		"--strategy-option",
		"--autosquash",
		"--interactive",
		"--exec",
		"--empty",
		"--update-refs",
	}

	contains := func(values []string, want string) bool {
		for _, value := range values {
			if value == want {
				return true
			}
		}
		return false
	}

	for _, value := range conflicts {
		if !contains(apply.ConflictsWith, value) {
			t.Errorf("--apply does not conflict with %s", value)
		}

		option, found := findOption(options, value)
		if !found {
			t.Fatalf("rebase %s option is missing", value)
		}
		if !contains(option.ConflictsWith, "--apply") {
			t.Errorf("%s does not conflict with --apply", value)
		}
	}

	keepBase, _ := findOption(options, "--keep-base")
	root, _ := findOption(options, "--root")
	if !contains(keepBase.ConflictsWith, "--root") ||
		!contains(root.ConflictsWith, "--keep-base") {
		t.Error("--keep-base and --root conflicts are not symmetric")
	}
}

func TestRebaseApplyRootConflictDependsOnOnto(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git rebase --root --app", want: ""},
		{line: "git rebase --apply --ro", want: ""},
		{line: "git rebase --root --onto main --app", want: "--apply"},
		{line: "git rebase --apply --onto main --ro", want: "--root"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))
			found := false
			for _, candidate := range got {
				if candidate.Value == test.want {
					found = true
					break
				}
			}

			if test.want == "" && len(got) != 0 {
				t.Fatalf("Suggest(%q) = %#v, want no candidate", test.line, got)
			}
			if test.want != "" && !found {
				t.Fatalf("Suggest(%q) = %#v, missing %q", test.line, got, test.want)
			}
		})
	}
}

func TestSuggestsRepositoryManagementSubcommands(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git worktree ad", want: "add"},
		{line: "git worktree lis", want: "list"},
		{line: "git worktree loc", want: "lock"},
		{line: "git worktree mov", want: "move"},
		{line: "git worktree pru", want: "prune"},
		{line: "git worktree rem", want: "remove"},
		{line: "git worktree rep", want: "repair"},
		{line: "git worktree unl", want: "unlock"},

		{line: "git submodule ad", want: "add"},
		{line: "git submodule sta", want: "status"},
		{line: "git submodule ini", want: "init"},
		{line: "git submodule dei", want: "deinit"},
		{line: "git submodule upd", want: "update"},
		{line: "git submodule set-b", want: "set-branch"},
		{line: "git submodule set-u", want: "set-url"},
		{line: "git submodule sum", want: "summary"},
		{line: "git submodule for", want: "foreach"},
		{line: "git submodule syn", want: "sync"},
		{line: "git submodule abs", want: "absorbgitdirs"},

		{line: "git bisect sta", want: "start"},
		{line: "git bisect goo", want: "good"},
		{line: "git bisect bad", want: ""},
		{line: "git bisect ter", want: "terms"},
		{line: "git bisect ski", want: "skip"},
		{line: "git bisect nex", want: "next"},
		{line: "git bisect rese", want: "reset"},
		{line: "git bisect vis", want: "visualize"},
		{line: "git bisect repl", want: "replay"},
		{line: "git bisect log", want: ""},
		{line: "git bisect run", want: ""},
		{line: "git bisect new", want: ""},
		{line: "git bisect old", want: ""},
		{line: "git bisect vie", want: "view"},
		{line: "git bisect hel", want: "help"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))

			if test.want == "" {
				if len(got) != 0 {
					t.Fatalf(
						"Suggest(%q) = %#v, want completed command",
						test.line,
						got,
					)
				}
				return
			}

			if len(got) != 1 || got[0].Value != test.want {
				t.Fatalf(
					"Suggest(%q) = %#v, want %q",
					test.line,
					got,
					test.want,
				)
			}
		})
	}
}

func TestSuggestsWorktreeOptions(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git worktree add --det", want: "--detach"},
		{line: "git worktree add --no-ch", want: "--no-checkout"},
		{line: "git worktree add --orp", want: "--orphan"},
		{line: "git worktree list --verb", want: "--verbose"},
		{line: "git worktree list --porc", want: "--porcelain"},
		{line: "git worktree lock --rea", want: "--reason"},
		{line: "git worktree prune --dry", want: "--dry-run"},
		{line: "git worktree prune --exp", want: "--expire"},
		{line: "git worktree remove --for", want: "--force"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))

			if len(got) != 1 || got[0].Value != test.want {
				t.Fatalf(
					"Suggest(%q) = %#v, want %q",
					test.line,
					got,
					test.want,
				)
			}
		})
	}
}

func TestAnalyzeWorktreeValueHints(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git worktree add -b ", want: "branch"},
		{line: "git worktree add -B ", want: "branch"},
		{line: "git worktree add --reason ", want: "reason"},
		{line: "git worktree lock --reason ", want: "reason"},
		{line: "git worktree prune --expire ", want: "time"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			result := Analyze(test.line, len([]rune(test.line)))

			if result.Hint == nil {
				t.Fatalf("Analyze(%q) returned no value hint", test.line)
			}
			if result.Hint.Name != test.want {
				t.Fatalf(
					"hint name = %q, want %q",
					result.Hint.Name,
					test.want,
				)
			}
			if len(result.Suggestions) != 0 {
				t.Fatalf(
					"value position returned suggestions: %#v",
					result.Suggestions,
				)
			}
		})
	}
}

func TestWorktreeOptionConflicts(t *testing.T) {
	tests := []struct {
		line      string
		forbidden string
	}{
		{
			line:      "git worktree add --checkout --no-ch",
			forbidden: "--no-checkout",
		},
		{
			line:      "git worktree add --no-checkout --ch",
			forbidden: "--checkout",
		},
		{
			line:      "git worktree add -b feature -B",
			forbidden: "-B",
		},
		{
			line:      "git worktree add -B feature -b",
			forbidden: "-b",
		},
		{
			line:      "git worktree list --verbose --por",
			forbidden: "--porcelain",
		},
		{
			line:      "git worktree list --porcelain --ver",
			forbidden: "--verbose",
		},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))

			for _, candidate := range got {
				if candidate.Value == test.forbidden {
					t.Fatalf(
						"Suggest(%q) returned conflicting option %q: %#v",
						test.line,
						test.forbidden,
						got,
					)
				}
			}
		})
	}
}

func TestSuggestsSubmoduleOptions(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git submodule --qui", want: "--quiet"},
		{line: "git submodule add --bran", want: "--branch"},
		{line: "git submodule add --nam", want: "--name"},
		{line: "git submodule add --refer", want: "--reference"},
		{line: "git submodule add --ref-f", want: "--ref-format"},
		{line: "git submodule add --dep", want: "--depth"},
		{line: "git submodule status --cac", want: "--cached"},
		{line: "git submodule status --recur", want: "--recursive"},
		{line: "git submodule deinit --al", want: "--all"},
		{line: "git submodule update --ini", want: "--init"},
		{line: "git submodule update --rem", want: "--remote"},
		{line: "git submodule update --no-f", want: "--no-fetch"},
		{line: "git submodule update --check", want: "--checkout"},
		{line: "git submodule update --mer", want: "--merge"},
		{line: "git submodule update --reba", want: "--rebase"},
		{line: "git submodule update --job", want: "--jobs"},
		{line: "git submodule update --filt", want: "--filter"},
		{line: "git submodule update --recomm", want: "--recommend-shallow"},
		{line: "git submodule update --no-s", want: "--no-single-branch"},
		{line: "git submodule set-branch --def", want: "--default"},
		{line: "git submodule summary --fil", want: "--files"},
		{line: "git submodule summary --summary", want: "--summary-limit"},
		{line: "git submodule foreach --recur", want: "--recursive"},
		{line: "git submodule sync --recur", want: "--recursive"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))
			if len(got) != 1 || got[0].Value != test.want {
				t.Fatalf(
					"Suggest(%q) = %#v, want %q",
					test.line,
					got,
					test.want,
				)
			}
		})
	}
}

func TestAnalyzeSubmoduleValueHints(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git submodule add --branch ", want: "branch"},
		{line: "git submodule add --name ", want: "name"},
		{line: "git submodule add --reference ", want: "repository"},
		{line: "git submodule add --ref-format ", want: "format"},
		{line: "git submodule add --depth ", want: "depth"},
		{line: "git submodule update --jobs ", want: "count"},
		{line: "git submodule update --filter ", want: "filter"},
		{line: "git submodule set-branch --branch ", want: "branch"},
		{line: "git submodule summary --summary-limit ", want: "count"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			result := Analyze(test.line, len([]rune(test.line)))
			if result.Hint == nil {
				t.Fatalf("Analyze(%q) returned no value hint", test.line)
			}
			if result.Hint.Name != test.want {
				t.Fatalf(
					"hint name = %q, want %q",
					result.Hint.Name,
					test.want,
				)
			}
			if len(result.Suggestions) != 0 {
				t.Fatalf(
					"value position returned suggestions: %#v",
					result.Suggestions,
				)
			}
		})
	}
}

func TestSubmoduleOptionConflicts(t *testing.T) {
	tests := []struct {
		line      string
		forbidden string
	}{
		{
			line:      "git submodule update --checkout --mer",
			forbidden: "--merge",
		},
		{
			line:      "git submodule update --merge --reba",
			forbidden: "--rebase",
		},
		{
			line:      "git submodule update --rebase --check",
			forbidden: "--checkout",
		},
		{
			line:      "git submodule update --recommend-shallow --no-rec",
			forbidden: "--no-recommend-shallow",
		},
		{
			line:      "git submodule update --single-branch --no-s",
			forbidden: "--no-single-branch",
		},
		{
			line:      "git submodule set-branch --default --bran",
			forbidden: "--branch",
		},
		{
			line:      "git submodule summary --cached --fil",
			forbidden: "--files",
		},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))
			for _, candidate := range got {
				if candidate.Value == test.forbidden {
					t.Fatalf(
						"Suggest(%q) returned conflicting option %q: %#v",
						test.line,
						test.forbidden,
						got,
					)
				}
			}
		})
	}
}

func TestSuggestsBisectOptions(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git bisect start --no-c", want: "--no-checkout"},
		{line: "git bisect start --first", want: "--first-parent"},
		{line: "git bisect start --term-b", want: "--term-bad"},
		{line: "git bisect start --term-n", want: "--term-new"},
		{line: "git bisect start --term-g", want: "--term-good"},
		{line: "git bisect start --term-o", want: "--term-old"},
		{line: "git bisect terms --term-g", want: "--term-good"},
		{line: "git bisect terms --term-n", want: "--term-new"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))
			if len(got) != 1 || got[0].Value != test.want {
				t.Fatalf(
					"Suggest(%q) = %#v, want %q",
					test.line,
					got,
					test.want,
				)
			}
		})
	}
}

func TestAnalyzeBisectTermValueHints(t *testing.T) {
	for _, line := range []string{
		"git bisect start --term-bad ",
		"git bisect start --term-new ",
		"git bisect start --term-good ",
		"git bisect start --term-old ",
	} {
		t.Run(line, func(t *testing.T) {
			result := Analyze(line, len([]rune(line)))
			if result.Hint == nil {
				t.Fatalf("Analyze(%q) returned no value hint", line)
			}
			if result.Hint.Name != "term" {
				t.Fatalf(
					"hint name = %q, want %q",
					result.Hint.Name,
					"term",
				)
			}
		})
	}
}

func TestBisectTermOptionsConflict(t *testing.T) {
	tests := []struct {
		line      string
		forbidden string
	}{
		{
			line:      "git bisect start --term-bad broken --term-n",
			forbidden: "--term-new",
		},
		{
			line:      "git bisect start --term-old working --term-g",
			forbidden: "--term-good",
		},
		{
			line:      "git bisect terms --term-good --term-b",
			forbidden: "--term-bad",
		},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))
			for _, candidate := range got {
				if candidate.Value == test.forbidden {
					t.Fatalf(
						"Suggest(%q) returned conflicting option %q: %#v",
						test.line,
						test.forbidden,
						got,
					)
				}
			}
		})
	}
}

func TestAnalyzeSuggestsBranchesForBisectRevisions(t *testing.T) {
	branches := []Suggestion{
		{Value: "main", Kind: KindBranch},
		{Value: "feature/login", Kind: KindBranch},
	}

	tests := []struct {
		line string
		want string
	}{
		{line: "git bisect start ma", want: "main"},
		{line: "git bisect good fe", want: "feature/login"},
		{line: "git bisect bad ma", want: "main"},
		{line: "git bisect new ma", want: "main"},
		{line: "git bisect old fe", want: "feature/login"},
		{line: "git bisect skip ma", want: "main"},
		{line: "git bisect reset ma", want: "main"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			result := AnalyzeWithBranches(
				test.line,
				len([]rune(test.line)),
				branches,
			)

			if len(result.Suggestions) != 1 ||
				result.Suggestions[0].Value != test.want {
				t.Fatalf(
					"AnalyzeWithBranches(%q) = %#v, want %q",
					test.line,
					result.Suggestions,
					test.want,
				)
			}
		})
	}
}

func TestBisectPathspecDoesNotSuggestBranches(t *testing.T) {
	line := "git bisect start main dev -- ma"
	branches := []Suggestion{{Value: "main", Kind: KindBranch}}

	result := AnalyzeWithBranches(
		line,
		len([]rune(line)),
		branches,
	)

	if len(result.Suggestions) != 0 {
		t.Fatalf(
			"AnalyzeWithBranches(%q) = %#v, want no branches after --",
			line,
			result.Suggestions,
		)
	}
}

func TestSuggestsInspectionCommands(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git bla", want: "blame"},
		{line: "git gre", want: "grep"},
		{line: "git short", want: "shortlog"},
		{line: "git des", want: "describe"},
		{line: "git refl", want: "reflog"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))
			if len(got) != 1 || got[0].Value != test.want {
				t.Fatalf(
					"Suggest(%q) = %#v, want %q",
					test.line,
					got,
					test.want,
				)
			}
		})
	}
}

func TestSuggestsAmbiguousInspectionPrefix(t *testing.T) {
	got := Suggest("git sho", len([]rune("git sho")))
	want := []string{"show", "shortlog"}

	if len(got) != len(want) {
		t.Fatalf(
			"Suggest(%q) returned %d candidates, want %d: %#v",
			"git sho",
			len(got),
			len(want),
			got,
		)
	}

	for index, value := range want {
		if got[index].Value != value {
			t.Fatalf(
				"candidate %d = %q, want %q",
				index,
				got[index].Value,
				value,
			)
		}
	}
}

func TestSuggestsReflogSubcommands(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git reflog sho", want: "show"},
		{line: "git reflog lis", want: "list"},
		{line: "git reflog exi", want: "exists"},
		{line: "git reflog wri", want: "write"},
		{line: "git reflog del", want: "delete"},
		{line: "git reflog dro", want: "drop"},
		{line: "git reflog exp", want: "expire"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))
			if len(got) != 1 || got[0].Value != test.want {
				t.Fatalf(
					"Suggest(%q) = %#v, want %q",
					test.line,
					got,
					test.want,
				)
			}
		})
	}
}

func TestSuggestsInspectionOptions(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git show --stat", want: ""},
		{line: "git show --name-st", want: "--name-status"},
		{line: "git show --show-s", want: "--show-signature"},
		{line: "git blame --line-p", want: "--line-porcelain"},
		{line: "git blame --color-b", want: "--color-by-age"},
		{line: "git grep --fixed", want: "--fixed-strings"},
		{line: "git grep --perl", want: "--perl-regexp"},
		{line: "git grep --before", want: "--before-context"},
		{line: "git grep --max-d", want: "--max-depth"},
		{line: "git shortlog --comm", want: "--committer"},
		{line: "git shortlog --numb", want: "--numbered"},
		{line: "git shortlog --summ", want: "--summary"},
		{line: "git describe --cand", want: "--candidates"},
		{line: "git describe --exact", want: "--exact-match"},
		{line: "git describe --first", want: "--first-parent"},
		{line: "git reflog show --date", want: ""},
		{line: "git reflog delete --dry", want: "--dry-run"},
		{line: "git reflog drop --single", want: "--single-worktree"},
		{line: "git reflog expire --expire-u", want: "--expire-unreachable"},
		{line: "git reflog expire --stale", want: "--stale-fix"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))

			if test.want == "" {
				if len(got) != 0 {
					t.Fatalf(
						"Suggest(%q) = %#v, want completed option",
						test.line,
						got,
					)
				}
				return
			}

			if len(got) != 1 || got[0].Value != test.want {
				t.Fatalf(
					"Suggest(%q) = %#v, want %q",
					test.line,
					got,
					test.want,
				)
			}
		})
	}
}

func TestAnalyzeInspectionValueHints(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{line: "git show --format ", want: "format"},
		{line: "git blame --ignore-rev ", want: "revision"},
		{line: "git blame --contents ", want: "file"},
		{line: "git blame -L ", want: "range"},
		{line: "git grep --context ", want: "lines"},
		{line: "git grep -e ", want: "pattern"},
		{line: "git grep -f ", want: "file"},
		{line: "git shortlog --group ", want: "field"},
		{line: "git describe --abbrev ", want: "length"},
		{line: "git describe --match ", want: "pattern"},
		{line: "git reflog show --date ", want: "format"},
		{line: "git reflog expire --expire ", want: "time"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			result := Analyze(test.line, len([]rune(test.line)))
			if result.Hint == nil {
				t.Fatalf("Analyze(%q) returned no value hint", test.line)
			}
			if result.Hint.Name != test.want {
				t.Fatalf(
					"hint name = %q, want %q",
					result.Hint.Name,
					test.want,
				)
			}
			if len(result.Suggestions) != 0 {
				t.Fatalf(
					"value position returned suggestions: %#v",
					result.Suggestions,
				)
			}
		})
	}
}

func TestInspectionOptionConflicts(t *testing.T) {
	tests := []struct {
		line      string
		forbidden string
	}{
		{
			line:      "git show --patch --no-p",
			forbidden: "--no-patch",
		},
		{
			line:      "git show --name-only --name-st",
			forbidden: "--name-status",
		},
		{
			line:      "git blame --porcelain --line-p",
			forbidden: "--line-porcelain",
		},
		{
			line:      "git grep --fixed-strings --perl",
			forbidden: "--perl-regexp",
		},
		{
			line:      "git grep --no-index --cac",
			forbidden: "--cached",
		},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			got := Suggest(test.line, len([]rune(test.line)))
			for _, candidate := range got {
				if candidate.Value == test.forbidden {
					t.Fatalf(
						"Suggest(%q) returned conflicting option %q: %#v",
						test.line,
						test.forbidden,
						got,
					)
				}
			}
		})
	}
}

func TestAnalyzeSuggestsBranchesForInspectionCommands(t *testing.T) {
	branches := []Suggestion{
		{Value: "main", Kind: KindBranch},
		{Value: "feature/login", Kind: KindBranch},
	}

	tests := []struct {
		line string
		want string
	}{
		{line: "git show ma", want: "main"},
		{line: "git blame fe", want: "feature/login"},
		{line: "git shortlog ma", want: "main"},
		{line: "git describe fe", want: "feature/login"},
		{line: "git reflog show ma", want: "main"},
		{line: "git reflog exists fe", want: "feature/login"},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			result := AnalyzeWithBranches(
				test.line,
				len([]rune(test.line)),
				branches,
			)

			if len(result.Suggestions) != 1 ||
				result.Suggestions[0].Value != test.want {
				t.Fatalf(
					"AnalyzeWithBranches(%q) = %#v, want %q",
					test.line,
					result.Suggestions,
					test.want,
				)
			}
		})
	}
}

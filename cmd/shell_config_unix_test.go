//go:build !windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Haruko386/Gogit/internal/protocol"
)

func TestBashPromptHooksCannotReplaceGogitPrompt(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is not installed")
	}

	tests := []struct {
		name       string
		bashRC     string
		wantHook   string
		promptLoop string
	}{
		{
			name:     "scalar PROMPT_COMMAND",
			bashRC:   "PROMPT_COMMAND='GOGIT_TEST_HOOK=scalar; PS1=user-scalar'\n",
			wantHook: "hook=scalar",
			promptLoop: `
eval "$PROMPT_COMMAND"
`,
		},
		{
			name:     "array PROMPT_COMMAND",
			bashRC:   "PROMPT_COMMAND=('GOGIT_TEST_HOOK=array' 'PS1=user-array')\n",
			wantHook: "hook=array",
			promptLoop: `
for action in "${PROMPT_COMMAND[@]}"; do eval "$action"; done
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			writeTestFile(t, filepath.Join(home, ".bashrc"), test.bashRC)

			initFile := filepath.Join(t.TempDir(), "bashrc")
			marker := "bash-hook-test"
			writeTestFile(t, initFile, bashInitScript(marker))

			command := exec.Command(
				bash,
				"--noprofile",
				"--rcfile",
				initFile,
				"-i",
				"-c",
				test.promptLoop+`printf 'hook=%s\nprompt=%s\n' "$GOGIT_TEST_HOOK" "$PS1"`,
			)
			command.Env = append(os.Environ(), "HOME="+home)
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("run Bash fixture: %v\n%s", err, output)
			}

			assertContainsAll(
				t,
				string(output),
				test.wantHook,
				protocol.BeginMarker(marker),
				protocol.EndMarker(marker),
			)
		})
	}
}

func TestZshEnvAndPromptHooksPreserveGogitBootstrap(t *testing.T) {
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh is not installed")
	}

	home := t.TempDir()
	selectedZDOTDIR := filepath.Join(home, "selected-zdotdir")
	if err := os.Mkdir(selectedZDOTDIR, 0o700); err != nil {
		t.Fatal(err)
	}
	writeTestFile(
		t,
		filepath.Join(home, ".zshenv"),
		fmt.Sprintf(
			"export GOGIT_TEST_ZSHENV=loaded\nexport ZDOTDIR=%s\n",
			quoteShell(selectedZDOTDIR),
		),
	)
	writeTestFile(
		t,
		filepath.Join(selectedZDOTDIR, ".zshrc"),
		`export GOGIT_TEST_ZSHRC=selected
gogit_test_user_precmd() {
    PROMPT=user-prompt
    RPROMPT=user-right-prompt
}
typeset -ga precmd_functions
precmd_functions+=(gogit_test_user_precmd)
`,
	)

	bootstrap := t.TempDir()
	writeTestFile(t, filepath.Join(bootstrap, ".zshenv"), zshEnvInitScript())
	marker := "zsh-hook-test"
	writeTestFile(t, filepath.Join(bootstrap, ".zshrc"), zshInitScript(marker))

	script := `
for hook in $precmd_functions; do $hook; done
command sh -c 'printf "child=%s\n" "$GOGIT_TEST_ZSHENV"'
print -r -- "zshrc=$GOGIT_TEST_ZSHRC"
print -r -- "prompt=$PROMPT"
print -r -- "right=$RPROMPT"
`
	command := exec.Command(zsh, "-d", "-i", "-c", script)
	command.Env = append(
		os.Environ(),
		"HOME="+home,
		"ZDOTDIR="+bootstrap,
		"GOGIT_USER_ZDOTDIR="+home,
		"GOGIT_USER_ZDOTDIR_WAS_SET=false",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run Zsh fixture: %v\n%s", err, output)
	}

	assertContainsAll(
		t,
		string(output),
		"child=loaded",
		"zshrc=selected",
		"right=\n",
		protocol.BeginMarker(marker),
		protocol.EndMarker(marker),
	)
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertContainsAll(t *testing.T, output string, values ...string) {
	t.Helper()
	for _, value := range values {
		if !strings.Contains(output, value) {
			t.Fatalf("output does not contain %q: %q", value, output)
		}
	}
}

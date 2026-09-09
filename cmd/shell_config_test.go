package cmd

import (
	"strings"
	"testing"

	"github.com/Haruko386/Gogit/internal/protocol"
)

func TestShellInitScriptsLoadUserConfigBeforeInstallingPrompt(t *testing.T) {
	marker := "shell-config-test"
	tests := []struct {
		name       string
		script     string
		userConfig string
		promptName string
	}{
		{
			name:       "Bash",
			script:     bashInitScript(marker),
			userConfig: `. "$HOME/.bashrc"`,
			promptName: "PS1=",
		},
		{
			name:       "Zsh",
			script:     zshInitScript(marker),
			userConfig: `source "$GOGIT_USER_ZDOTDIR_AFTER_ZSHENV/.zshrc"`,
			promptName: "PROMPT=",
		},
		{
			name:       "POSIX sh",
			script:     posixInitScript(marker),
			userConfig: `. "$GOGIT_USER_ENV"`,
			promptName: "PS1=",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configIndex := strings.Index(test.script, test.userConfig)
			protectionIndex := strings.Index(
				test.script,
				protectPromptEnvironment,
			)
			promptIndex := strings.Index(test.script, test.promptName)
			if configIndex < 0 {
				t.Fatalf("user configuration is not loaded: %q", test.script)
			}
			if protectionIndex <= configIndex || promptIndex <= protectionIndex {
				t.Fatalf("prompt protection and installation must follow user configuration: %q", test.script)
			}
			if !strings.Contains(test.script, protocol.BeginMarker(marker)) ||
				!strings.Contains(test.script, protocol.EndMarker(marker)) {
				t.Fatalf("prompt protocol markers are missing: %q", test.script)
			}
		})
	}
}

func TestZshEnvBootstrapPreservesUserChangesBeforeRestoringBootstrapDirectory(t *testing.T) {
	script := zshEnvInitScript()
	sourceIndex := strings.Index(
		script,
		`source "$__gogit_user_zdotdir/.zshenv"`,
	)
	selectionIndex := strings.Index(
		script,
		`GOGIT_USER_ZDOTDIR_AFTER_ZSHENV=${ZDOTDIR:-$HOME}`,
	)
	restoreIndex := strings.Index(
		script,
		`ZDOTDIR=$__gogit_bootstrap_zdotdir`,
	)

	if sourceIndex < 0 || selectionIndex <= sourceIndex || restoreIndex <= selectionIndex {
		t.Fatalf("unexpected Zsh environment bootstrap order: %q", script)
	}
}

func TestShellInitScriptsAppendFinalPromptHooks(t *testing.T) {
	bashScript := bashInitScript("bash-hook-test")
	for _, expected := range []string{
		"PROMPT_COMMAND+=(__gogit_restore_prompt)",
		`PROMPT_COMMAND="${PROMPT_COMMAND%;};__gogit_restore_prompt"`,
	} {
		if !strings.Contains(bashScript, expected) {
			t.Fatalf("Bash init script does not contain %q: %q", expected, bashScript)
		}
	}

	zshScript := zshInitScript("zsh-hook-test")
	if expected := "precmd_functions+=(__gogit_restore_prompt)"; !strings.Contains(zshScript, expected) {
		t.Fatalf("Zsh init script does not contain %q: %q", expected, zshScript)
	}
}

func TestQuoteShellPreservesSingleQuotes(t *testing.T) {
	if got, want := quoteShell("it's"), `'it'"'"'s'`; got != want {
		t.Fatalf("quoteShell() = %q, want %q", got, want)
	}
}

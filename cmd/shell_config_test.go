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
	bashMarker := "bash-hook-test"
	bashRecovery := protocol.RecoveryName(bashMarker)
	bashScript := bashInitScript(bashMarker)
	for _, expected := range []string{
		"PROMPT_COMMAND+=(" + bashRecovery + ")",
		`PROMPT_COMMAND="${PROMPT_COMMAND%%;};` + bashRecovery + `"`,
	} {
		if !strings.Contains(bashScript, expected) {
			t.Fatalf("Bash init script does not contain %q: %q", expected, bashScript)
		}
	}

	zshMarker := "zsh-hook-test"
	zshScript := zshInitScript(zshMarker)
	if expected := "precmd_functions+=(" + protocol.RecoveryName(zshMarker) + ")"; !strings.Contains(zshScript, expected) {
		t.Fatalf("Zsh init script does not contain %q: %q", expected, zshScript)
	}
}

func TestPOSIXCommandWrapperRestoresPromptAfterUserCommand(t *testing.T) {
	marker := "wrapper-test"
	command := "PS1=user-prompt # deliberately replace the protocol prompt"
	wrapped := wrapPOSIXCommand(command, marker)

	if !strings.Contains(wrapped, command) {
		t.Fatalf("wrapped command lost user input: %q", wrapped)
	}
	if !strings.Contains(wrapped, protocol.RecoveryName(marker)) {
		t.Fatalf("wrapped command does not invoke recovery: %q", wrapped)
	}
	if !strings.Contains(wrapped, "\n}; ") {
		t.Fatalf("recovery is not protected from a trailing comment: %q", wrapped)
	}
	if !strings.Contains(wrapped, "pass-through mode") {
		t.Fatalf("wrapped command has no safe-degradation message: %q", wrapped)
	}
}

func TestPOSIXCommandWrapperLeavesBlankCommandAlone(t *testing.T) {
	if got := wrapPOSIXCommand("   ", "wrapper-test"); got != "   " {
		t.Fatalf("blank command = %q", got)
	}
}

func TestQuoteShellPreservesSingleQuotes(t *testing.T) {
	if got, want := quoteShell("it's"), `'it'"'"'s'`; got != want {
		t.Fatalf("quoteShell() = %q, want %q", got, want)
	}
}

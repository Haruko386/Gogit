package cmd

import (
	"fmt"
	"strings"

	"github.com/Haruko386/Gogit/internal/protocol"
)

const promptFieldSeparator = "\x1f"

const protectPromptEnvironment = "export CONDA_CHANGEPS1=false\nexport VIRTUAL_ENV_DISABLE_PROMPT=1\n"

func bashInitScript(marker string) string {
	prompt := protocol.BeginMarker(marker) +
		`$(if [ -n "$CONDA_DEFAULT_ENV" ]; then ` +
		`printf "%s" "$CONDA_DEFAULT_ENV"; ` +
		`elif [ -n "$VIRTUAL_ENV" ]; then ` +
		`basename "$VIRTUAL_ENV"; fi)` +
		promptFieldSeparator +
		`\w` +
		protocol.EndMarker(marker)

	return fmt.Sprintf(
		"if [ -r \"$HOME/.bashrc\" ]; then . \"$HOME/.bashrc\"; fi\n%sPS1=%s\n",
		protectPromptEnvironment,
		quoteShell(prompt),
	)
}

func zshInitScript(marker string) string {
	prompt := protocol.BeginMarker(marker) +
		`${CONDA_DEFAULT_ENV:-${VIRTUAL_ENV:t}}` +
		promptFieldSeparator +
		`%~` +
		protocol.EndMarker(marker)

	return fmt.Sprintf(
		"if [[ -r \"$GOGIT_USER_ZDOTDIR/.zshrc\" ]]; then source \"$GOGIT_USER_ZDOTDIR/.zshrc\"; fi\n%ssetopt PROMPT_SUBST\nPROMPT=%s\nRPROMPT=\n",
		protectPromptEnvironment,
		quoteShell(prompt),
	)
}

func posixInitScript(marker string) string {
	prompt := protocol.BeginMarker(marker) +
		`$(if [ -n "$CONDA_DEFAULT_ENV" ]; then ` +
		`printf "%s" "$CONDA_DEFAULT_ENV"; ` +
		`elif [ -n "$VIRTUAL_ENV" ]; then ` +
		`basename "$VIRTUAL_ENV"; fi)` +
		promptFieldSeparator +
		`$(pwd)` +
		protocol.EndMarker(marker)

	return fmt.Sprintf(
		"if [ -n \"$GOGIT_USER_ENV\" ] && [ -r \"$GOGIT_USER_ENV\" ]; then . \"$GOGIT_USER_ENV\"; fi\n%sPS1=%s\n",
		protectPromptEnvironment,
		quoteShell(prompt),
	)
}

func quoteShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

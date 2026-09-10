package cmd

import (
	"fmt"
	"strings"

	"github.com/Haruko386/Gogit/internal/protocol"
)

const promptFieldSeparator = "\x1f"

const protectPromptEnvironment = "export CONDA_CHANGEPS1=false\nexport VIRTUAL_ENV_DISABLE_PROMPT=1\n"

func bashInitScript(marker string) string {
	recoveryName := protocol.RecoveryName(marker)
	prompt := protocol.BeginMarker(marker) +
		`$(if [ -n "$CONDA_DEFAULT_ENV" ]; then ` +
		`printf "%s" "$CONDA_DEFAULT_ENV"; ` +
		`elif [ -n "$VIRTUAL_ENV" ]; then ` +
		`basename "$VIRTUAL_ENV"; fi)` +
		promptFieldSeparator +
		`\w` +
		protocol.EndMarker(marker)

	return fmt.Sprintf(
		`if [ -r "$HOME/.bashrc" ]; then . "$HOME/.bashrc"; fi
%s__gogit_prompt=%s
%s() {
    PS1=$__gogit_prompt
}
if [[ $(declare -p PROMPT_COMMAND 2>/dev/null) =~ ^declare[[:space:]]+-[^[:space:]]*a[^[:space:]]*[[:space:]]+PROMPT_COMMAND= ]]; then
    PROMPT_COMMAND+=(%s)
elif [[ -n ${PROMPT_COMMAND-} ]]; then
    PROMPT_COMMAND="${PROMPT_COMMAND%%%%;};%s"
else
    PROMPT_COMMAND=%s
fi
%s
`,
		protectPromptEnvironment,
		quoteShell(prompt),
		recoveryName,
		recoveryName,
		recoveryName,
		recoveryName,
		recoveryName,
	)
}

func zshEnvInitScript() string {
	return `__gogit_bootstrap_zdotdir=$ZDOTDIR
__gogit_user_zdotdir=$GOGIT_USER_ZDOTDIR
if [[ $GOGIT_USER_ZDOTDIR_WAS_SET == true ]]; then
    ZDOTDIR=$__gogit_user_zdotdir
else
    unset ZDOTDIR
fi
if [[ -r "$__gogit_user_zdotdir/.zshenv" ]]; then
    source "$__gogit_user_zdotdir/.zshenv"
fi
GOGIT_USER_ZDOTDIR_AFTER_ZSHENV=${ZDOTDIR:-$HOME}
ZDOTDIR=$__gogit_bootstrap_zdotdir
unset __gogit_bootstrap_zdotdir __gogit_user_zdotdir
`
}

func zshInitScript(marker string) string {
	recoveryName := protocol.RecoveryName(marker)
	prompt := protocol.BeginMarker(marker) +
		`${CONDA_DEFAULT_ENV:-${VIRTUAL_ENV:t}}` +
		promptFieldSeparator +
		`%~` +
		protocol.EndMarker(marker)

	return fmt.Sprintf(
		`if [[ -r "$GOGIT_USER_ZDOTDIR_AFTER_ZSHENV/.zshrc" ]]; then
    source "$GOGIT_USER_ZDOTDIR_AFTER_ZSHENV/.zshrc"
fi
%ssetopt PROMPT_SUBST
typeset -g __gogit_prompt=%s
%s() {
    PROMPT=$__gogit_prompt
    RPROMPT=
}
typeset -ga precmd_functions
precmd_functions+=(%s)
%s
`,
		protectPromptEnvironment,
		quoteShell(prompt),
		recoveryName,
		recoveryName,
		recoveryName,
	)
}

func posixInitScript(marker string) string {
	recoveryName := protocol.RecoveryName(marker)
	prompt := protocol.BeginMarker(marker) +
		`$(if [ -n "$CONDA_DEFAULT_ENV" ]; then ` +
		`printf "%s" "$CONDA_DEFAULT_ENV"; ` +
		`elif [ -n "$VIRTUAL_ENV" ]; then ` +
		`basename "$VIRTUAL_ENV"; fi)` +
		promptFieldSeparator +
		`$(pwd)` +
		protocol.EndMarker(marker)

	return fmt.Sprintf(
		"if [ -n \"$GOGIT_USER_ENV\" ] && [ -r \"$GOGIT_USER_ENV\" ]; then . \"$GOGIT_USER_ENV\"; fi\n%s__gogit_prompt=%s\n%s() { PS1=$__gogit_prompt; }\n%s\n",
		protectPromptEnvironment,
		quoteShell(prompt),
		recoveryName,
		recoveryName,
	)
}

func wrapPOSIXCommand(command, marker string) string {
	if strings.TrimSpace(command) == "" {
		return command
	}
	recoveryName := protocol.RecoveryName(marker)
	return "{\n" + command + "\n}; __gogit_status=$?; " +
		"if command -v " + recoveryName + " >/dev/null 2>&1; then " +
		recoveryName + "; else printf '\\nGogit: prompt protocol recovery failed; " +
		"continuing in pass-through mode.\\n' >&2; fi; (exit $__gogit_status)"
}

func quoteShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

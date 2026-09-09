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
		`if [ -r "$HOME/.bashrc" ]; then . "$HOME/.bashrc"; fi
%s__gogit_prompt=%s
__gogit_restore_prompt() {
    PS1=$__gogit_prompt
}
if [[ $(declare -p PROMPT_COMMAND 2>/dev/null) =~ ^declare[[:space:]]+-[^[:space:]]*a[^[:space:]]*[[:space:]]+PROMPT_COMMAND= ]]; then
    PROMPT_COMMAND+=(__gogit_restore_prompt)
elif [[ -n ${PROMPT_COMMAND-} ]]; then
    PROMPT_COMMAND="${PROMPT_COMMAND%%;};__gogit_restore_prompt"
else
    PROMPT_COMMAND=__gogit_restore_prompt
fi
__gogit_restore_prompt
`,
		protectPromptEnvironment,
		quoteShell(prompt),
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
__gogit_restore_prompt() {
    PROMPT=$__gogit_prompt
    RPROMPT=
}
typeset -ga precmd_functions
precmd_functions+=(__gogit_restore_prompt)
__gogit_restore_prompt
`,
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

//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Haruko386/Gogit/internal/protocol"
)

func systemShell(marker string) (*exec.Cmd, commandWrapper, func(), error) {
	script := powershellInitScript(marker)

	var command *exec.Cmd

	if path, err := exec.LookPath("pwsh.exe"); err == nil {
		command = exec.Command(
			path,
			"-NoLogo",
			"-NoExit",
			"-Command",
			script,
		)
	} else {
		command = exec.Command(
			"powershell.exe",
			"-NoLogo",
			"-NoExit",
			"-Command",
			script,
		)
	}

	command.Env = append(
		os.Environ(),
		"CONDA_CHANGEPS1=false",
		"VIRTUAL_ENV_DISABLE_PROMPT=1",
	)

	// PowerShell wraps commands after PSConsoleHostReadLine returns. The PTY
	// therefore receives and echoes only the command the user actually typed.
	return command, nil, func() {}, nil
}

func powershellWrapperName(marker string) string {
	return protocol.RecoveryName(marker) + "_wrap"
}

func powershellInitScript(marker string) string {
	begin := protocol.BeginMarker(marker)
	end := protocol.EndMarker(marker)
	recoveryName := protocol.RecoveryName(marker)
	wrapperName := powershellWrapperName(marker)

	return fmt.Sprintf(`
$env:CONDA_CHANGEPS1 = 'false'
$env:VIRTUAL_ENV_DISABLE_PROMPT = '1'

$global:__gogit_original_readline = ${function:PSConsoleHostReadLine}

$global:__gogit_prompt_impl = {
    $environmentName = ''

    if (-not [string]::IsNullOrWhiteSpace(
        $env:CONDA_DEFAULT_ENV
    )) {
        $environmentName = $env:CONDA_DEFAULT_ENV
    }
    elseif (-not [string]::IsNullOrWhiteSpace(
        $env:VIRTUAL_ENV
    )) {
		$environmentName = Split-Path -Leaf -Path $env:VIRTUAL_ENV
    }

    $directory = (
        $executionContext.SessionState.Path.CurrentLocation.Path
    )

	return '%s' + $environmentName + [char]31 + $directory + '%s'
}

function global:%s {
    param(
        [Parameter(Mandatory = $true)]
        [AllowEmptyString()]
        [string] $Command
    )

    if ([string]::IsNullOrWhiteSpace($Command)) {
        return $Command
    }

    $scriptText = $Command + [Environment]::NewLine + '$__gogit_command_succeeded = $?'
    $encoded = [System.Convert]::ToBase64String(
        [System.Text.Encoding]::UTF8.GetBytes($scriptText)
    )

    return ('$__gogit_command = [scriptblock]::Create([System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String(''__GOGIT_ENCODED__''))); . $__gogit_command; if (Test-Path Function:\global:%s) { & %s } else { [Console]::Error.WriteLine(''Gogit: prompt protocol recovery failed; continuing in pass-through mode.'') }; if (-not $__gogit_command_succeeded) { Write-Error ''Gogit: preserving command failure status.'' -ErrorAction Ignore }').Replace('__GOGIT_ENCODED__', $encoded)
}

$global:__gogit_readline_impl = {
    if ($null -ne $global:__gogit_original_readline) {
        $command = & $global:__gogit_original_readline
    }
    else {
        $command = [Console]::ReadLine()
    }

    return & %s $command
}

function global:%s {
    Set-Item -Path Function:\global:prompt -Value $global:__gogit_prompt_impl
    Set-Item -Path Function:\global:PSConsoleHostReadLine -Value $global:__gogit_readline_impl
}

& %s
`, begin, end, wrapperName, recoveryName, recoveryName, wrapperName, recoveryName, recoveryName)
}

//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Haruko386/Gogit/internal/protocol"
)

func systemShell(marker string) (*exec.Cmd, commandWrapper, func(), error) {
	recoveryName := protocol.RecoveryName(marker)
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

	return command, func(input string) string {
		return wrapPowerShellCommand(input, recoveryName)
	}, func() {}, nil
}

func wrapPowerShellCommand(input, recoveryName string) string {
	if strings.TrimSpace(input) == "" {
		return input
	}

	return ". {\n" + input + "\n$__gogit_command_succeeded = $?\n}; " +
		"if (Test-Path Function:\\global:" + recoveryName + ") { & " + recoveryName +
		" } else { [Console]::Error.WriteLine('Gogit: prompt protocol recovery failed; continuing in pass-through mode.') }; " +
		"if (-not $__gogit_command_succeeded) { Write-Error 'Gogit: preserving command failure status.' -ErrorAction Ignore }"
}

func powershellInitScript(marker string) string {
	begin := protocol.BeginMarker(marker)
	end := protocol.EndMarker(marker)
	recoveryName := protocol.RecoveryName(marker)

	return fmt.Sprintf(`
$env:CONDA_CHANGEPS1 = 'false'
$env:VIRTUAL_ENV_DISABLE_PROMPT = '1'

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
    Set-Item -Path Function:\global:prompt -Value $global:__gogit_prompt_impl
}

& %s
`, begin, end, recoveryName, recoveryName)
}

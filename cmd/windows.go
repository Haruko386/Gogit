//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Haruko386/Gogit/internal/protocol"
)

func systemShell(marker string) (*exec.Cmd, func(), error) {
	begin := protocol.BeginMarker(marker)
	end := protocol.EndMarker(marker)

	script := fmt.Sprintf(`
$env:CONDA_CHANGEPS1 = 'false'
$env:VIRTUAL_ENV_DISABLE_PROMPT = '1'

function global:prompt {
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
`, begin, end)

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

	return command, func() {}, nil
}

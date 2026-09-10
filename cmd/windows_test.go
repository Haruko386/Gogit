//go:build windows

package cmd

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Haruko386/Gogit/internal/protocol"
)

func TestPowerShellRecoveryRestoresOverwrittenPrompt(t *testing.T) {
	powershell, err := exec.LookPath("pwsh.exe")
	if err != nil {
		powershell, err = exec.LookPath("powershell.exe")
	}
	if err != nil {
		t.Skip("PowerShell is not installed")
	}

	marker := "powershell-recovery-test"
	recoveryName := protocol.RecoveryName(marker)
	script := powershellInitScript(marker) + `
function global:prompt { 'user prompt' }
& ` + recoveryName + `
& prompt
`
	command := exec.Command(powershell, "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", script)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run PowerShell recovery fixture: %v\n%s", err, output)
	}

	wantBegin := protocol.BeginMarker(marker)
	wantEnd := protocol.EndMarker(marker)
	if !strings.Contains(string(output), wantBegin) || !strings.Contains(string(output), wantEnd) {
		t.Fatalf("restored prompt does not contain a protocol frame: %q", output)
	}
}

func TestPowerShellCommandWrapperProtectsRecoveryFromComment(t *testing.T) {
	marker := "powershell-wrapper-test"
	_, wrapper, _, err := systemShell(marker)
	if err != nil {
		t.Fatal(err)
	}

	wrapped := wrapper("function global:prompt { 'user prompt' } # replace prompt")
	if !strings.Contains(wrapped, "\n}; if") {
		t.Fatalf("recovery is not protected from a trailing comment: %q", wrapped)
	}
	if !strings.Contains(wrapped, protocol.RecoveryName(marker)) {
		t.Fatalf("wrapped command does not invoke session recovery: %q", wrapped)
	}
	if !strings.Contains(wrapped, "pass-through mode") {
		t.Fatalf("wrapped command has no safe-degradation message: %q", wrapped)
	}
}

func TestPowerShellCommandWrapperPreservesFailureStatus(t *testing.T) {
	interpreters := []struct {
		name string
		path string
	}{
		{name: "PowerShell Core", path: findExecutable("pwsh.exe")},
		{name: "Windows PowerShell", path: findExecutable("powershell.exe")},
	}

	for _, interpreter := range interpreters {
		t.Run(interpreter.name, func(t *testing.T) {
			if interpreter.path == "" {
				t.Skip(interpreter.name + " is not installed")
			}

			marker := "status-test-" + interpreter.name
			recoveryName := protocol.RecoveryName(marker)
			tests := []struct {
				name       string
				setup      string
				command    string
				wantNative string
			}{
				{
					name:       "native command",
					command:    `cmd.exe /d /c exit 7`,
					wantNative: "7",
				},
				{
					name:       "PowerShell command",
					setup:      `$global:LASTEXITCODE = 41`,
					command:    `Write-Error 'expected failure' -ErrorAction SilentlyContinue`,
					wantNative: "41",
				},
			}

			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					wrapped := wrapPowerShellCommand(test.command, recoveryName)
					script := powershellInitScript(marker) + "\n" + test.setup + "\n" +
						wrapped + `
[Console]::WriteLine("status=$?;native=$LASTEXITCODE")
`
					command := exec.Command(
						interpreter.path,
						"-NoLogo",
						"-NoProfile",
						"-NonInteractive",
						"-Command",
						script,
					)
					output, err := command.CombinedOutput()
					if err != nil {
						t.Fatalf("run status fixture: %v\n%s", err, output)
					}

					want := "status=False;native=" + test.wantNative
					if !strings.Contains(string(output), want) {
						t.Fatalf("output does not contain %q: %q", want, output)
					}
				})
			}
		})
	}
}

func findExecutable(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absolute
}

//go:build windows

package cmd

import (
	"os/exec"
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

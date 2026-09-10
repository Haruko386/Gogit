//go:build windows

package cmd

import (
	"bytes"
	"encoding/base64"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Haruko386/Gogit/internal/protocol"
	"github.com/Haruko386/Gogit/internal/session"
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
	powershell, err := exec.LookPath("pwsh.exe")
	if err != nil {
		powershell, err = exec.LookPath("powershell.exe")
	}
	if err != nil {
		t.Skip("PowerShell is not installed")
	}

	marker := "powershell-wrapper-test"
	input := "function global:prompt { 'user prompt' } # replace prompt"
	encodedInput := base64.StdEncoding.EncodeToString([]byte(input))
	script := powershellInitScript(marker) + `
$inputText = [System.Text.Encoding]::UTF8.GetString(
    [System.Convert]::FromBase64String('` + encodedInput + `')
)
$wrapped = & ` + powershellWrapperName(marker) + ` $inputText
[Console]::WriteLine(
    [System.Convert]::ToBase64String(
        [System.Text.Encoding]::UTF8.GetBytes($wrapped)
    )
)
`
	command := exec.Command(powershell, "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", script)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run PowerShell wrapper fixture: %v\n%s", err, output)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(output)))
	if err != nil {
		t.Fatalf("decode PowerShell wrapper: %v\n%s", err, output)
	}
	wrapped := string(decoded)

	if strings.ContainsAny(wrapped, "\r\n") {
		t.Fatalf(
			"interactive PowerShell wrapper contains a physical newline: %q",
			wrapped,
		)
	}
	if strings.Contains(wrapped, input) {
		t.Fatalf("submitted command was not encoded: %q", wrapped)
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
					encodedInput := base64.StdEncoding.EncodeToString([]byte(test.command))
					script := powershellInitScript(marker) + "\n" + test.setup + `
$inputText = [System.Text.Encoding]::UTF8.GetString(
    [System.Convert]::FromBase64String('` + encodedInput + `')
)
$wrapped = & ` + powershellWrapperName(marker) + ` $inputText
$wrapped += '; [Console]::WriteLine("status=$?;native=$LASTEXITCODE")'
& ([scriptblock]::Create($wrapped))
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

func TestPowerShellInteractiveInputDoesNotEchoProtocolWrapper(t *testing.T) {
	marker := "interactive-wrapper-test"
	command, wrapper, cleanup, err := systemShell(marker)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if wrapper != nil {
		t.Fatal("PowerShell commands must be wrapped after ReadLine, not before PTY input")
	}

	shell := session.New(command, 120, 40)
	if err := shell.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := shell.Close(); err != nil {
			t.Errorf("close PowerShell session: %v", err)
		}
	}()

	events := readStream(shell)
	output := waitForShellOutput(t, events, nil, func(data []byte) bool {
		return bytes.Count(data, []byte(protocol.EndMarker(marker))) >= 1
	})

	input := []byte("Write-Output gogit-interactive-ok\r")
	if err := writeAll(shell, input); err != nil {
		t.Fatal(err)
	}
	output = waitForShellOutput(t, events, output, func(data []byte) bool {
		return bytes.Count(data, []byte(protocol.EndMarker(marker))) >= 2 &&
			bytes.Contains(data, []byte("gogit-interactive-ok"))
	})

	if bytes.Contains(output, []byte("$__gogit_command")) ||
		bytes.Contains(output, []byte("FromBase64String")) {
		t.Fatalf("internal PowerShell wrapper was echoed: %q", output)
	}
}

func waitForShellOutput(
	t *testing.T,
	events <-chan streamEvent,
	initial []byte,
	done func([]byte) bool,
) []byte {
	t.Helper()
	output := append([]byte(nil), initial...)
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()

	for !done(output) {
		select {
		case event := <-events:
			output = append(output, event.data...)
			if event.err != nil {
				t.Fatalf("read PowerShell PTY: %v; output: %q", event.err, output)
			}
		case <-timer.C:
			t.Fatalf("timed out waiting for PowerShell output: %q", output)
		}
	}
	return output
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

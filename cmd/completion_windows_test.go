//go:build windows

package cmd

import "testing"

func TestQuoteCommandArgumentForPowerShell(t *testing.T) {
	if got, want := quoteCommandArgument("release'; Write-Error unsafe"), "'release''; Write-Error unsafe'"; got != want {
		t.Fatalf("quoteCommandArgument() = %q, want %q", got, want)
	}
}

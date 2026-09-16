//go:build !windows

package cmd

import "testing"

func TestQuoteCommandArgumentForPOSIXShell(t *testing.T) {
	if got, want := quoteCommandArgument("release'; echo unsafe"), `'release'"'"'; echo unsafe'`; got != want {
		t.Fatalf("quoteCommandArgument() = %q, want %q", got, want)
	}
}

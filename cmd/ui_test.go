package cmd

import (
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Haruko386/Gogit/internal/protocol"
)

type testShellSession struct {
	output *io.PipeReader
	writes chan []byte
	done   chan struct{}
}

func (s *testShellSession) Start() error                  { return nil }
func (s *testShellSession) Read(data []byte) (int, error) { return s.output.Read(data) }
func (s *testShellSession) Resize(_, _ int) error         { return nil }
func (s *testShellSession) Wait() error                   { <-s.done; return nil }
func (s *testShellSession) Close() error                  { return nil }

func (s *testShellSession) Write(data []byte) (int, error) {
	s.writes <- append([]byte(nil), data...)
	return len(data), nil
}

func TestRunShellUIReplaysCommandsAfterMultilinePaste(t *testing.T) {
	stdinReader, stdinWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := stdinReader.Close(); err != nil {
			t.Errorf("close stdin reader: %v", err)
		}
	})

	var (
		closeStdinWriterOnce sync.Once
		closeStdinWriterErr  error
	)
	closeStdinWriter := func() error {
		closeStdinWriterOnce.Do(func() {
			closeStdinWriterErr = stdinWriter.Close()
		})
		return closeStdinWriterErr
	}
	t.Cleanup(func() {
		if err := closeStdinWriter(); err != nil {
			t.Errorf("close stdin writer: %v", err)
		}
	})

	stdout, err := os.CreateTemp(t.TempDir(), "gogit-output-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := stdout.Close(); err != nil {
			t.Errorf("close stdout capture: %v", err)
		}
	})

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	os.Stdin = stdinReader
	os.Stdout = stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	shellOutput, shellWriter := io.Pipe()
	shellDone := make(chan struct{})
	shell := &testShellSession{
		output: shellOutput,
		writes: make(chan []byte, 2),
		done:   shellDone,
	}
	defer func() {
		close(shellDone)
		if err := shellOutput.Close(); err != nil {
			t.Errorf("close shell output reader: %v", err)
		}
		if err := shellWriter.Close(); err != nil {
			t.Errorf("close shell output writer: %v", err)
		}
	}()

	marker := "multiline-test"
	resizeDone := make(chan error)
	result := make(chan error, 1)
	go func() {
		_, runErr := runShellUI(shell, marker, resizeDone)
		result <- runErr
	}()

	writePromptFrame(t, shellWriter, marker)
	waitForOutput(t, stdout)

	if _, err := stdinWriter.Write([]byte("\x1b[200~git status\r\n")); err != nil {
		t.Fatal(err)
	}

	assertShellWrite(t, shell.writes, "git status\r")
	writePromptFrame(t, shellWriter, marker)

	if _, err := stdinWriter.Write([]byte("go test ./...\r\n\x1b[20")); err != nil {
		t.Fatal(err)
	}
	assertNoShellWrite(t, shell.writes)

	if _, err := stdinWriter.Write([]byte("1~")); err != nil {
		t.Fatal(err)
	}
	assertShellWrite(t, shell.writes, "go test ./...\r")

	if _, err := stdinWriter.Write([]byte("interactive input")); err != nil {
		t.Fatal(err)
	}
	assertShellWrite(t, shell.writes, "interactive input")

	if err := closeStdinWriter(); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("runShellUI returned an error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("runShellUI did not stop after stdin closed")
	}
}

func TestFormatPromptUsesTerminalCellWidth(t *testing.T) {
	_, width := formatPrompt(protocol.Prompt{
		Environment: "开发",
		Directory:   "项目🚀",
	})

	if want := 23; width != want {
		t.Fatalf("prompt width = %d, want %d", width, want)
	}
}

func writePromptFrame(t *testing.T, writer io.Writer, marker string) {
	t.Helper()
	frame := protocol.BeginMarker(marker) +
		string(rune(0x1f)) +
		"test-directory" +
		protocol.EndMarker(marker)
	if _, err := io.WriteString(writer, frame); err != nil {
		t.Fatal(err)
	}
}

func waitForOutput(t *testing.T, output *os.File) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		info, err := output.Stat()
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() > 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("initial prompt was not rendered")
}

func assertShellWrite(t *testing.T, writes <-chan []byte, want string) {
	t.Helper()
	select {
	case got := <-writes:
		if string(got) != want {
			t.Fatalf("shell write = %q, want %q", got, want)
		}
	case <-time.After(time.Second):
		t.Fatalf("shell did not receive %q", want)
	}
}

func assertNoShellWrite(t *testing.T, writes <-chan []byte) {
	t.Helper()
	select {
	case got := <-writes:
		t.Fatalf("unexpected shell write: %q", got)
	case <-time.After(25 * time.Millisecond):
	}
}

package protocol

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	markerPrefix       = "__GOGIT_READY_"
	fieldSeparator     = byte(0x1f)
	maxPromptFrameSize = 64 * 1024
)

type Prompt struct {
	Environment string
	Directory   string
}

// NewMarker creates a marker that identifies the shell prompt belonging to
// this Gogit process.
func NewMarker() (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate shell marker: %w", err)
	}

	return markerPrefix + hex.EncodeToString(random) + "__", nil
}

func BeginMarker(marker string) string {
	return marker + "PROMPT_BEGIN__"
}

func EndMarker(marker string) string {
	return marker + "PROMPT_END__"
}

// RecoveryName returns a shell-safe, session-specific function name. Keeping
// the recovery entry point unpredictable avoids collisions with user config.
func RecoveryName(marker string) string {
	sum := sha256.Sum256([]byte(marker))
	return "__gogit_recover_" + hex.EncodeToString(sum[:8])
}

// Scanner removes shell-ready markers from arbitrarily chunked PTY output.
type Scanner struct {
	begin   []byte
	end     []byte
	pending []byte
	frame   []byte
	inside  bool
}

// NewScanner creates a scanner for one shell session.
func NewScanner(marker string) *Scanner {
	if marker == "" {
		panic("protocol marker must not be empty")
	}

	return &Scanner{
		begin: []byte(BeginMarker(marker)),
		end:   []byte(EndMarker(marker)),
	}
}

// Push accepts the next PTY output chunk.
//
// visible contains bytes that may be printed to the user's terminal.
// prompts contains the state carried by each complete prompt frame.
func (s *Scanner) Push(data []byte) (
	visible []byte,
	prompts []Prompt,
) {
	buffer := make([]byte, 0, len(s.pending)+len(data))
	buffer = append(buffer, s.pending...)
	buffer = append(buffer, data...)
	s.pending = nil

	for len(buffer) > 0 {
		delimiter := s.begin
		if s.inside {
			delimiter = s.end

			// A new begin marker before an end marker means the previous frame
			// was damaged. Resynchronize without exposing either marker.
			beginIndex := bytes.Index(buffer, s.begin)
			endIndex := bytes.Index(buffer, s.end)
			if beginIndex >= 0 && (endIndex < 0 || beginIndex < endIndex) {
				s.frame = nil
				buffer = buffer[beginIndex+len(s.begin):]
				continue
			}
		}

		index := bytes.Index(buffer, delimiter)
		if index >= 0 {
			stable := buffer[:index]

			if s.inside {
				if len(s.frame)+len(stable) > maxPromptFrameSize {
					buffer = buffer[index+len(delimiter):]
					s.frame = nil
					s.inside = false
					s.pending = nil
					continue
				}
				s.frame = append(s.frame, stable...)
			} else {
				visible = append(visible, stable...)
			}

			buffer = buffer[index+len(delimiter):]

			if s.inside {
				prompts = append(prompts, parsePrompt(s.frame))
				s.frame = nil
				s.inside = false
			} else {
				s.inside = true
			}

			continue
		}

		keep := matchingSuffixLength(buffer, delimiter)
		if s.inside {
			keep = max(keep, matchingSuffixLength(buffer, s.begin))
		}
		stable := buffer[:len(buffer)-keep]

		if s.inside {
			if len(s.frame)+len(stable) > maxPromptFrameSize {
				// A corrupt frame must not consume unbounded memory or hide all
				// subsequent shell output forever. Discard it and scan afresh.
				s.frame = nil
				s.inside = false
				s.pending = nil
				break
			}
			s.frame = append(s.frame, stable...)
		} else {
			visible = append(visible, stable...)
		}

		if keep > 0 {
			s.pending = append(
				s.pending,
				buffer[len(buffer)-keep:]...,
			)
		}

		break
	}

	return visible, prompts
}

// Flush resets an incomplete protocol frame. Partial markers are deliberately
// discarded: exposing them would leak the per-session protocol token.
func (s *Scanner) Flush() []byte {
	s.pending = nil
	s.frame = nil
	s.inside = false

	return nil
}

func parsePrompt(data []byte) Prompt {
	environment, directory, found := bytes.Cut(
		data,
		[]byte{fieldSeparator},
	)
	if !found {
		return Prompt{
			Directory: string(data),
		}
	}

	return Prompt{
		Environment: string(environment),
		Directory:   string(directory),
	}
}

// matchingSuffixLength returns the longest suffix of data that is also a
// prefix of marker.
func matchingSuffixLength(data, marker []byte) int {
	maxLength := min(len(data), len(marker)-1)

	for length := maxLength; length > 0; length-- {
		if bytes.Equal(
			data[len(data)-length:],
			marker[:length],
		) {
			return length
		}
	}

	return 0
}

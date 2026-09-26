package history

import "strings"

// History stores commands entered during the current Gogit session.
// TODO: store the history to user log, so that it will always have history
type History struct {
	entries  []string
	position int
	draft    string
	browsing bool
	limit    int
}

func New(entries []string, limit int) History {
	if limit <= 0 {
		limit = DefaultLimit
	}

	history := History{
		limit: limit,
	}

	for _, entry := range entries {
		history.Add(entry)
	}

	history.Reset()
	return history
}

// Reset ends the current history navigation.
func (h *History) Reset() {
	h.position = len(h.entries)
	h.draft = ""
	h.browsing = false
}

// Add records one completed command.
func (h *History) Add(command string) bool {
	if strings.TrimSpace(command) == "" {
		h.Reset()
		return false
	}

	if len(h.entries) > 0 &&
		h.entries[len(h.entries)-1] == command {
		h.Reset()
		return false
	}

	h.entries = append(h.entries, command)

	limit := h.limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if len(h.entries) > limit {
		h.entries = append(
			[]string(nil),
			h.entries[len(h.entries)-limit:]...,
		)
	}

	h.Reset()
	return true
}

// Previous moves to an older command.
// current is preserved as a draft when history browsing begins.
func (h *History) Previous(current string) (string, bool) {
	if len(h.entries) == 0 {
		return "", false
	}

	if !h.browsing {
		h.draft = current
		h.position = len(h.entries)
		h.browsing = true
	}

	if h.position == 0 {
		return "", false
	}

	h.position--
	return h.entries[h.position], true
}

// Next moves to a newer command or restores the original draft.
func (h *History) Next() (string, bool) {
	if !h.browsing {
		return "", false
	}

	if h.position < len(h.entries)-1 {
		h.position++
		return h.entries[h.position], true
	}

	draft := h.draft
	h.Reset()
	return draft, true
}

// Browsing reports whether history navigation is currently active.
func (h *History) Browsing() bool {
	return h.browsing
}

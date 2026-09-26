package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultLimit = 1000

type Store struct {
	path  string
	limit int
}

func NewStore(path string, limit int) *Store {
	if limit <= 0 {
		limit = DefaultLimit
	}

	return &Store{path: path, limit: limit}
}

func DefaultStore() (*Store, error) {
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("locate user config directory: %w", err)
	}

	return NewStore(filepath.Join(configDirectory, "gogit", "history.jsonl"), DefaultLimit), nil
}

func (s Store) Load() ([]string, error) {
	file, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entries := make([]string, 0)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		var command string
		if err := json.Unmarshal(scanner.Bytes(), &command); err != nil {
			// A corrupt record must not prevent Gogit from starting.
			continue
		}

		if strings.TrimSpace(command) == "" {
			continue
		}

		entries = append(entries, command)
		if len(entries) > s.limit {
			entries = append([]string(nil), entries[len(entries)-s.limit:]...)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func (s Store) Append(command string) error {
	if !ShouldPersist(command) {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}

	record, err := json.Marshal(command)
	if err != nil {
		return err
	}
	record = append(record, '\n')

	file, err := os.OpenFile(
		s.path,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0o600,
	)
	if err != nil {
		return err
	}

	if _, err := file.Write(record); err != nil {
		_ = file.Close()
		return err
	}

	return file.Close()
}

func ShouldPersist(command string) bool {
	if strings.TrimSpace(command) == "" {
		return false
	}

	if command[0] == ' ' || command[0] == '\t' {
		return false
	}

	lowerCommand := strings.ToLower(command)
	sensitiveMarkers := []string{
		"--password",
		"--token",
		"--secret",
		"password=",
		"token=",
		"secret=",
		"authorization:",
	}

	for _, marker := range sensitiveMarkers {
		if strings.Contains(lowerCommand, marker) {
			return false
		}
	}

	return true
}

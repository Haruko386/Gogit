package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	entries, _, err := s.readEntries()
	return entries, err
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

	if err := file.Close(); err != nil {
		return err
	}

	return s.compactIfNeeded()
}

func (s Store) compactIfNeeded() error {
	entries, count, err := s.readEntries()
	if err != nil {
		return err
	}
	if count <= 2*s.limit {
		return nil
	}

	directory := filepath.Dir(s.path)
	temporary, err := os.CreateTemp(directory, ".gogit-history-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}

	encoder := json.NewEncoder(temporary)
	for _, entry := range entries {
		if err := encoder.Encode(entry); err != nil {
			_ = temporary.Close()
			return err
		}
	}

	if err := temporary.Close(); err != nil {
		return err
	}

	return replaceFile(temporaryPath, s.path)
}

func (s Store) readEntries() ([]string, int, error) {
	file, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	entries := make([]string, 0, s.limit)
	count := 0
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		var command string
		if err := json.Unmarshal(scanner.Bytes(), &command); err != nil {
			continue
		}
		if strings.TrimSpace(command) == "" {
			continue
		}

		count++
		entries = append(entries, command)
		if len(entries) > s.limit {
			entries = append([]string(nil), entries[len(entries)-s.limit:]...)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}

	return entries, count, nil
}

func replaceFile(source, destination string) error {
	renameErr := os.Rename(source, destination)
	if renameErr == nil {
		return nil
	}
	if runtime.GOOS != "windows" {
		return renameErr
	}

	// Windows cannot rename over an existing file. Move the original aside
	// first so a failed replacement can still restore it.
	backup := source + ".old"
	if err := os.Rename(destination, backup); err != nil {
		return err
	}

	if err := os.Rename(source, destination); err != nil {
		restoreErr := os.Rename(backup, destination)
		return errors.Join(err, restoreErr)
	}

	return os.Remove(backup)
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

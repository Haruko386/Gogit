package suggest

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LoadBranches returns local and remote-tracking branches known to the Git
// repository containing directory. A non-repository is treated as having no
// branches because completion failures must not interrupt terminal input.
func LoadBranches(ctx context.Context, directory string) ([]Suggestion, error) {
	if directory == "" {
		return nil, nil
	}
	directory = expandHomeDirectory(directory)

	command := exec.CommandContext(
		ctx,
		"git",
		"-C",
		directory,
		"for-each-ref",
		"--format=%(HEAD)%09%(refname)%09%(refname:short)",
		"refs/heads",
		"refs/remotes",
	)
	output, err := command.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, nil
		}
		return nil, err
	}

	return parseBranches(output), nil
}

func expandHomeDirectory(directory string) string {
	if directory != "~" &&
		!strings.HasPrefix(directory, "~/") &&
		!strings.HasPrefix(directory, `~\`) {
		return directory
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return directory
	}
	if directory == "~" {
		return home
	}
	return filepath.Join(home, directory[2:])
}

func parseBranches(output []byte) []Suggestion {
	branches := make([]Suggestion, 0)
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		fields := strings.SplitN(scanner.Text(), "\t", 3)
		if len(fields) != 3 || fields[2] == "" {
			continue
		}

		fullName := fields[1]
		shortName := fields[2]
		if strings.HasPrefix(fullName, "refs/remotes/") &&
			strings.HasSuffix(fullName, "/HEAD") {
			continue
		}

		description := "Local branch."
		if strings.HasPrefix(fullName, "refs/remotes/") {
			description = "Remote-tracking branch."
		} else if fields[0] == "*" {
			description = "Current local branch."
		}

		branches = append(branches, Suggestion{
			Value:       shortName,
			Description: description,
			Kind:        KindBranch,
		})
	}

	return branches
}

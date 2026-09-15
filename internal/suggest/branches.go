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

// LoadRepositoryCandidates loads dynamic values from the repository that
// contains directory.
func LoadRepositoryCandidates(
	ctx context.Context,
	directory string,
) (RepositoryCandidates, error) {
	branches, err := LoadBranches(ctx, directory)
	if err != nil {
		return RepositoryCandidates{}, err
	}

	remotes, err := LoadRemotes(ctx, directory)
	if err != nil {
		return RepositoryCandidates{}, err
	}

	tags, err := LoadTags(ctx, directory)
	if err != nil {
		return RepositoryCandidates{}, err
	}

	return RepositoryCandidates{
		Branches: branches,
		Remotes:  remotes,
		Tags:     tags,
	}, nil
}

// LoadBranches returns local and remote-tracking branches known to the Git
// repository containing directory. A non-repository is treated as having no
// branches because completion failures must not interrupt terminal input.
func LoadBranches(ctx context.Context, directory string) ([]Suggestion, error) {
	if directory == "" {
		return nil, nil
	}
	output, err := runRepositoryGit(ctx, directory,
		"for-each-ref",
		"--format=%(HEAD)%09%(refname)%09%(refname:short)",
		"refs/heads",
		"refs/remotes",
	)
	if err != nil {
		return nil, err
	}

	return parseBranches(output), nil
}

func LoadTags(ctx context.Context, directory string) ([]Suggestion, error) {
	output, err := runRepositoryGit(ctx, directory, "tag", "--list")
	if err != nil {
		return nil, err
	}
	return parseNamedSuggestions(output, KindTag, "Repository tag."), nil
}

func LoadRemotes(ctx context.Context, directory string) ([]Suggestion, error) {
	output, err := runRepositoryGit(ctx, directory, "remote")
	if err != nil {
		return nil, err
	}
	return parseNamedSuggestions(output, KindRemote, "Remote repository."), nil
}

// runRepositoryGit runs a Git command in directory. A Git command failure,
// such as directory not being a repository, produces no candidates.
func runRepositoryGit(
	ctx context.Context,
	directory string,
	arguments ...string,
) ([]byte, error) {
	if directory == "" {
		return nil, nil
	}

	directory = expandHomeDirectory(directory)

	commandArguments := make([]string, 0, len(arguments)+2)
	commandArguments = append(commandArguments, "-C", directory)
	commandArguments = append(commandArguments, arguments...)

	command := exec.CommandContext(ctx, "git", commandArguments...)

	output, err := command.Output()
	if err == nil {
		return output, nil
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, ctxErr
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return nil, nil
	}
	return nil, err
}

func parseNamedSuggestions(
	output []byte,
	kind Kind,
	description string,
) []Suggestion {
	suggestions := make([]Suggestion, 0)
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		value := strings.TrimSpace(scanner.Text())
		if value == "" {
			continue
		}

		suggestions = append(suggestions, Suggestion{
			Value:       value,
			Description: description,
			Kind:        kind,
		})
	}

	return suggestions
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

package suggest

import "strings"

// Suggest returns Git completions for the token at cursor.
func Suggest(line string, cursor int) []Suggestion {
	context, ok := ParseContext(line, cursor)
	if !ok ||
		len(context.WordsBefore) == 0 ||
		context.WordsBefore[0] != "git" {
		return nil
	}

	if len(context.WordsBefore) == 1 {
		return matching(
			gitSubcommands,
			context.Prefix,
			nil,
		)
	}

	options, ok := gitOptions[context.WordsBefore[1]]
	if !ok {
		return nil
	}

	used := usedOptions(context.WordsBefore[2:])

	return matching(
		options,
		context.Prefix,
		used,
	)
}

// matching filters candidates to those starting with prefix, skipping the
// candidate whose value already equals prefix and any value present in excluded.
func matching(candidates []Suggestion, prefix string, excluded map[string]struct{}) []Suggestion {
	matched := make([]Suggestion, 0, len(candidates))

	for _, candidate := range candidates {
		// The current token is already complete.
		if candidate.Value == prefix {
			continue
		}

		// The option has already appeared earlier in the command.
		if _, exists := excluded[candidate.Value]; exists {
			continue
		}

		if strings.HasPrefix(candidate.Value, prefix) {
			matched = append(matched, candidate)
		}
	}

	return matched
}

// usedOptions collects the option names that already appear among words, so
// they can be excluded from further suggestions. Values passed with "=" are
// keyed by the option name only.
func usedOptions(words []string) map[string]struct{} {
	used := make(map[string]struct{}, len(words))

	for _, word := range words {
		name, _, _ := strings.Cut(word, "=")
		used[name] = struct{}{}
	}

	return used
}

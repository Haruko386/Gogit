package suggest

import (
	"strings"
	"unicode"
)

// Suggest returns Git completions for the token at cursor.
func Suggest(line string, cursor int) []Suggestion {
	return Analyze(line, cursor).Suggestions
}

// Analyze returns insertable completions and, when applicable, a display-only
// explanation of the value expected by the preceding option.
func Analyze(line string, cursor int) Result {
	context, ok := ParseContext(line, cursor)
	if !ok ||
		len(context.WordsBefore) == 0 ||
		context.WordsBefore[0] != "git" {
		return Result{}
	}

	if len(context.WordsBefore) == 1 {
		return Result{Suggestions: matching(
			gitSubcommands, context.Prefix, nil,
		)}
	}

	options, ok := gitOptions[context.WordsBefore[1]]
	if !ok {
		return Result{}
	}

	words, atBoundary := commandWords(line, cursor)
	if hint := expectedValueHint(words, atBoundary, options); hint != nil {
		return Result{Hint: hint}
	}

	used := usedOptions(context.WordsBefore[2:])

	return Result{Suggestions: matching(
		options, context.Prefix, used,
	)}
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
		if suggestionWasUsed(candidate, excluded) {
			continue
		}

		if strings.HasPrefix(candidate.Value, prefix) {
			matched = append(matched, candidate)
		}
	}

	return matched
}

func suggestionWasUsed(candidate Suggestion, used map[string]struct{}) bool {
	if candidate.Repeatable {
		return false
	}
	if _, exists := used[candidate.Value]; exists {
		return true
	}
	for _, alias := range candidate.Aliases {
		if _, exists := used[alias]; exists {
			return true
		}
	}
	return false
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

func expectedValueHint(
	words []string,
	atBoundary bool,
	options []Suggestion,
) *ValueHint {
	if len(words) < 3 {
		return nil
	}

	for index := 2; index < len(words); index++ {
		name, _, hasEquals := strings.Cut(words[index], "=")
		option, ok := findOption(options, name)
		if !ok || !option.TakesValue {
			continue
		}

		if hasEquals {
			if index == len(words)-1 && !atBoundary {
				return valueHint(option)
			}
			continue
		}

		if index == len(words)-1 {
			return valueHint(option)
		}

		index++ // The following word is this option's value.
		if index == len(words)-1 && !atBoundary {
			return valueHint(option)
		}
	}

	return nil
}

func findOption(options []Suggestion, name string) (Suggestion, bool) {
	for _, option := range options {
		if option.Value == name {
			return option, true
		}
		for _, alias := range option.Aliases {
			if alias == name {
				return option, true
			}
		}
	}
	return Suggestion{}, false
}

func valueHint(option Suggestion) *ValueHint {
	return &ValueHint{
		Name:        option.ValueName,
		Description: option.ValueDescription,
	}
}

// commandWords tokenizes the portion of the command up to cursor while
// treating whitespace inside single or double quotes as part of one word.
func commandWords(line string, cursor int) ([]string, bool) {
	runes := []rune(line)
	if cursor < 0 || cursor > len(runes) {
		return nil, false
	}
	runes = runes[:cursor]

	words := make([]string, 0, 6)
	var current strings.Builder
	var quote rune
	escaped := false
	tokenStarted := false
	atBoundary := len(runes) == 0

	for _, value := range runes {
		if escaped {
			current.WriteRune(value)
			escaped = false
			tokenStarted = true
			atBoundary = false
			continue
		}

		if value == '\\' || value == '`' {
			escaped = true
			tokenStarted = true
			atBoundary = false
			continue
		}

		if quote != 0 {
			if value == quote {
				quote = 0
			} else {
				current.WriteRune(value)
			}
			atBoundary = false
			continue
		}

		if value == '\'' || value == '"' {
			quote = value
			tokenStarted = true
			atBoundary = false
			continue
		}

		if unicode.IsSpace(value) {
			if tokenStarted {
				words = append(words, current.String())
				current.Reset()
				tokenStarted = false
			}
			atBoundary = true
			continue
		}

		current.WriteRune(value)
		tokenStarted = true
		atBoundary = false
	}

	if escaped {
		current.WriteRune('\\')
	}
	if tokenStarted {
		words = append(words, current.String())
	}

	return words, atBoundary && quote == 0
}

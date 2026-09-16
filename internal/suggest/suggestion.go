package suggest

// Kind identifies what a suggestion represents.
type Kind string

const (
	KindSubcommand Kind = "subcommand"
	KindOption     Kind = "option"
	KindBranch     Kind = "branch"
	KindRemote     Kind = "remote"
	KindTag        Kind = "tag"
)

// Suggestion is one value Gogit can insert into the current token.
type Suggestion struct {
	Value            string
	Description      string
	Kind             Kind
	TakesValue       bool
	ValueName        string
	ValueDescription string
	Repeatable       bool
	Aliases          []string
	ConflictsWith    []string
}

// ValueHint explains the value expected after an option. It is display-only
// and must never be inserted as completion text.
type ValueHint struct {
	Name        string
	Description string
}

// Result separates insertable candidates from a display-only value hint.
type Result struct {
	Suggestions []Suggestion
	Hint        *ValueHint
}

// RepositoryCandidates contains dynamic values loaded from the current
// repository. Empty groups are valid when the directory is not a repository.
type RepositoryCandidates struct {
	Branches []Suggestion
	Remotes  []Suggestion
	Tags     []Suggestion
}

// Context describes the token under the cursor. All offsets are rune indexes.
type Context struct {
	WordsBefore []string
	Token       string
	Prefix      string
	TokenStart  int
	TokenEnd    int
}

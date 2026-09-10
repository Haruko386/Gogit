package suggest

var gitSubcommands = []Suggestion{
	{
		Value:       "status",
		Description: "Show the working tree status.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "add",
		Description: "Add file changes to the staging area.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "commit",
		Description: "Record staged changes in repository history.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "branch",
		Description: "List, create, or delete branches.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "switch",
		Description: "Switch branches or create a new branch.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "checkout",
		Description: "Switch branches or restore working tree files.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "restore",
		Description: "Restore working tree files.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "log",
		Description: "Show commit history.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "diff",
		Description: "Show changes between commits and working trees.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "merge",
		Description: "Join development histories together.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "rebase",
		Description: "Reapply commits on top of another base.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "fetch",
		Description: "Download objects and refs from a remote.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "pull",
		Description: "Fetch and integrate changes from a remote.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "push",
		Description: "Update remote refs and transfer objects.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "stash",
		Description: "Temporarily store uncommitted changes.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "clone",
		Description: "Clone a repository into a new directory.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "init",
		Description: "Create an empty Git repository",
		Kind:        KindSubcommand,
	},
	{
		Value:       "remote",
		Description: "Manage tracked remote repositories.",
		Kind:        KindSubcommand,
	},
	{
		Value:       "tag",
		Description: "Create, list, delete, or verify tags",
		Kind:        KindSubcommand,
	},
}

var gitNestedSubcommands = map[string][]Suggestion{
	"remote": {
		{
			Value:       "add",
			Description: "Add a new remote repository.",
			Kind:        KindSubcommand,
		},
		{
			Value:       "rename",
			Description: "Rename an existing remote.",
			Kind:        KindSubcommand,
		},
		{
			Value:       "remove",
			Description: "Remove an existing remote.",
			Kind:        KindSubcommand,
		},
		{
			Value:       "set-head",
			Description: "Set or delete the default branch for a remote.",
			Kind:        KindSubcommand,
		},
		{
			Value:       "set-branches",
			Description: "Change the branches tracked for a remote.",
			Kind:        KindSubcommand,
		},
		{
			Value:       "get-url",
			Description: "Show the URL of a remote.",
			Kind:        KindSubcommand,
		},
		{
			Value:       "set-url",
			Description: "Change the URL of a remote.",
			Kind:        KindSubcommand,
		},
		{
			Value:       "show",
			Description: "Show information about a remote.",
			Kind:        KindSubcommand,
		},
		{
			Value:       "prune",
			Description: "Delete stale remote-tracking references.",
			Kind:        KindSubcommand,
		},
		{
			Value:       "update",
			Description: "Fetch updates for one or more remotes.",
			Kind:        KindSubcommand,
		},
	},
}

var gitOptions = map[string][]Suggestion{
	"branch": {
		{
			Value:       "--show-current",
			Description: "Print the name of the current branch.",
			Kind:        KindOption,
		},
		{
			Value:       "--merged",
			Description: "List branches already merged into the specified commit.",
			Kind:        KindOption,
		},
		{
			Value:       "--no-merged",
			Description: "List branches that have not been merged.",
			Kind:        KindOption,
		},
		{
			Value:       "--delete",
			Description: "Delete a branch.",
			Kind:        KindOption,
		},
	},
	"status": {
		{
			Value:       "--short",
			Description: "Show status in a compact format.",
			Kind:        KindOption,
		},
		{
			Value:       "--branch",
			Description: "Show branch information.",
			Kind:        KindOption,
		},
		{
			Value:       "--porcelain",
			Description: "Use a stable machine-readable format.",
			Kind:        KindOption,
		},
	},
	"add": {
		{
			Value:       "--all",
			Description: "Stage changes from the entire working tree.",
			Kind:        KindOption,
		},
		{
			Value:       "--patch",
			Description: "Interactively choose changes to stage.",
			Kind:        KindOption,
		},
		{
			Value:       "--update",
			Description: "Stage modified and deleted tracked files.",
			Kind:        KindOption,
		},
	},
	"commit": {
		{
			Value:            "--message",
			Description:      "Use the supplied text as the commit message; repeat it to add paragraphs.",
			Kind:             KindOption,
			TakesValue:       true,
			ValueName:        "message",
			ValueDescription: "Enter the commit message. Quotes are recommended.",
			Repeatable:       true,
			Aliases:          []string{"-m"},
		},
		{
			Value:       "--amend",
			Description: "Replace the most recent commit.",
			Kind:        KindOption,
		},
		{
			Value:       "--all",
			Description: "Stage modified and deleted tracked files before committing.",
			Kind:        KindOption,
		},
		{
			Value:       "--no-verify",
			Description: "Skip commit hooks.",
			Kind:        KindOption,
		},
	},
	"switch": {
		{
			Value:       "--create",
			Description: "Create a new branch and switch to it.",
			Kind:        KindOption,
		},
		{
			Value:       "--detach",
			Description: "Switch to a commit without attaching HEAD to a branch.",
			Kind:        KindOption,
		},
	},
	"log": {
		{
			Value:       "--oneline",
			Description: "Show each commit on one line.",
			Kind:        KindOption,
		},
		{
			Value:       "--graph",
			Description: "Draw a text-based commit graph.",
			Kind:        KindOption,
		},
		{
			Value:       "--all",
			Description: "Show commits from all refs.",
			Kind:        KindOption,
		},
	},
	"clone": {
		{
			Value:            "--branch",
			Description:      "Check out the specified branch after cloning.",
			Kind:             KindOption,
			TakesValue:       true,
			ValueName:        "branch",
			ValueDescription: "Enter the branch or tag to check out.",
			Aliases:          []string{"-b"},
		},
		{
			Value:            "--depth",
			Description:      "Create a shallow clone with limited history.",
			Kind:             KindOption,
			TakesValue:       true,
			ValueName:        "depth",
			ValueDescription: "Enter the number of commits to include",
		},
		{
			Value:            "--origin",
			Description:      "Use a remote name other than origin.",
			Kind:             KindOption,
			TakesValue:       true,
			ValueName:        "name",
			ValueDescription: "Enter the name of the remote.",
			Aliases:          []string{"-o"},
		},
		{
			Value:       "--single-branch",
			Description: "Clone only the history of one branch.",
			Kind:        KindOption,
		},
		{
			Value:       "--recurse-submodules",
			Description: "Initialize and clone submodules.",
			Kind:        KindOption,
		},
		{
			Value:       "--no-checkout",
			Description: "Do not checkout HEAD after cloning.",
			Kind:        KindOption,
			Aliases:     []string{"-n"},
		},
		{
			Value:       "--bare",
			Description: "Create a bare repository.",
			Kind:        KindOption,
		},
		{
			Value:       "--mirror",
			Description: "Create a mirror of the source repository.",
			Kind:        KindOption,
		},
	},
	"init": {
		{
			Value:       "--bare",
			Description: "Create a bare repository.",
			Kind:        KindOption,
		},
		{
			Value:            "--initial-branch",
			Description:      "Use the specified name for the initial branch.",
			Kind:             KindOption,
			TakesValue:       true,
			ValueName:        "branch",
			ValueDescription: "Enter the name of the initial branch.",
			Aliases:          []string{"-b"},
		},
		{
			Value:            "--template",
			Description:      "Use templates from the specified directory.",
			Kind:             KindOption,
			TakesValue:       true,
			ValueName:        "directory",
			ValueDescription: "Enter the template directory.",
		},
		{
			Value:            "--separate-git-dir",
			Description:      "Store repository metadata in a separate directory.",
			Kind:             KindOption,
			TakesValue:       true,
			ValueName:        "git-dir",
			ValueDescription: "Enter the Git metadata directory.",
		},
	},
	"remote": {
		{
			Value:       "--verbose",
			Description: "Show remote URLs in addition to remote names",
			Kind:        KindOption,
			Aliases:     []string{"-v"},
		},
	},
	"remote add": {
		{
			Value:       "-f",
			Description: "Fetch the remote immediately after adding it.",
			Kind:        KindOption,
		},
		{
			Value:            "-t",
			Description:      "Track only the specified branch; repeat it to track multiple branches.",
			Kind:             KindOption,
			TakesValue:       true,
			ValueName:        "branch",
			ValueDescription: "Enter a remote branch to track.",
			Repeatable:       true,
		},
		{
			Value:            "-m",
			Description:      "Set the remote's default branch.",
			Kind:             KindOption,
			TakesValue:       true,
			ValueName:        "branch",
			ValueDescription: "Enter the remote's default branch.",
		},
		{
			Value:         "--tags",
			Description:   "Import every tag from the remote.",
			Kind:          KindOption,
			ConflictsWith: []string{"--no-tags"},
		},
		{
			Value:         "--no-tags",
			Description:   "Do not import tags from the remote.",
			Kind:          KindOption,
			ConflictsWith: []string{"--tags"},
		},
		{
			Value:       "--mirror=",
			Description: "Create a fetch or push mirror; append fetch or push.",
			Kind:        KindOption,
		},
	},
	"remote rename": {
		{
			Value:         "--progress",
			Description:   "Show progress while renaming remote references.",
			Kind:          KindOption,
			ConflictsWith: []string{"--no-progress"},
		},
		{
			Value:         "--no-progress",
			Description:   "Do not show progress while renaming remote references.",
			Kind:          KindOption,
			ConflictsWith: []string{"--progress"},
		},
	},
	"remote set-head": {
		{
			Value:         "--auto",
			Description:   "Query the remote and set its default branch automatically.",
			Kind:          KindOption,
			Aliases:       []string{"-a"},
			ConflictsWith: []string{"--delete"},
		},
		{
			Value:         "--delete",
			Description:   "Delete the remote's default-branch reference.",
			Kind:          KindOption,
			Aliases:       []string{"-d"},
			ConflictsWith: []string{"--auto"},
		},
	},
	"remote set-branches": {
		{
			Value:       "--add",
			Description: "Add branches instead of replacing the tracked branch list.",
			Kind:        KindOption,
		},
	},
	"remote get-url": {
		{
			Value:       "--push",
			Description: "Show push URLs instead of fetch URLs.",
			Kind:        KindOption,
		},
		{
			Value:       "--all",
			Description: "Show all URLs for the remote.",
			Kind:        KindOption,
		},
	},
	"remote set-url": {
		{
			Value:       "--push",
			Description: "Change push URLs instead of fetch URLs.",
			Kind:        KindOption,
		},
		{
			Value:         "--add",
			Description:   "Add a new URL instead of replacing an existing URL.",
			Kind:          KindOption,
			ConflictsWith: []string{"--delete"},
		},
		{
			Value:         "--delete",
			Description:   "Delete URLs matching the supplied regular expression.",
			Kind:          KindOption,
			ConflictsWith: []string{"--add"},
		},
	},
	"remote show": {
		{
			Value:       "--no-query",
			Description: "Use cached information without querying the remote.",
			Kind:        KindOption,
			Aliases:     []string{"-n"},
		},
	},
	"remote prune": {
		{
			Value:       "--dry-run",
			Description: "Report stale references without deleting them.",
			Kind:        KindOption,
			Aliases:     []string{"-n"},
		},
	},
	"remote update": {
		{
			Value:       "--prune",
			Description: "Prune stale references while updating remotes.",
			Kind:        KindOption,
			Aliases:     []string{"-p"},
		},
	},
	"tag": {
		{
			Value:       "--list",
			Description: "List tags, optionally matching a pattern",
			Kind:        KindOption,
			Aliases:     []string{"-l"},
		},
		{
			Value:       "--annotate",
			Description: "Create an annotated tag.",
			Kind:        KindOption,
			Aliases:     []string{"-a"},
		},
		{
			Value:            "--message",
			Description:      "Use the supplied annotation message.",
			Kind:             KindOption,
			TakesValue:       true,
			ValueName:        "message",
			ValueDescription: "Enter the tag annotation message.",
			Repeatable:       true,
			Aliases:          []string{"-m"},
		},
		{
			Value:       "--delete",
			Description: "Delete one or more tags.",
			Kind:        KindOption,
			Aliases:     []string{"-d"},
		},
		{
			Value:       "--force",
			Description: "Replace an existing tag.",
			Kind:        KindOption,
			Aliases:     []string{"-f"},
		},
		{
			Value:       "--sign",
			Description: "Create a GPG-signed tag",
			Kind:        KindOption,
			Aliases:     []string{"-s"},
		},
		{
			Value:       "--verify",
			Description: "Verify the GPG signature of a tag",
			Kind:        KindOption,
			Aliases:     []string{"-v"},
		},
		{
			Value:            "--contains",
			Description:      "List tags containing the specified commit.",
			Kind:             KindOption,
			TakesValue:       true,
			ValueName:        "commit",
			ValueDescription: "Enter a commit or revision",
		},
		{
			Value:            "--points-at",
			Description:      "List tags pointing at the specified object.",
			Kind:             KindOption,
			TakesValue:       true,
			ValueName:        "object",
			ValueDescription: "Enter a Git object name.",
		},
	},
}

package suggest

import "testing"

func TestParseBranches(t *testing.T) {
	output := []byte(
		" \trefs/heads/dev\tdev\n" +
			"*\trefs/heads/main\tmain\n" +
			" \trefs/remotes/origin/HEAD\torigin/HEAD\n" +
			" \trefs/remotes/origin/dev\torigin/dev\n",
	)

	got := parseBranches(output)
	if len(got) != 3 {
		t.Fatalf("parseBranches() returned %d branches: %#v", len(got), got)
	}

	tests := []struct {
		value       string
		description string
	}{
		{value: "dev", description: "Local branch."},
		{value: "main", description: "Current local branch."},
		{value: "origin/dev", description: "Remote-tracking branch."},
	}

	for index, test := range tests {
		if got[index].Value != test.value {
			t.Fatalf("branch %d value = %q, want %q", index, got[index].Value, test.value)
		}
		if got[index].Description != test.description {
			t.Fatalf(
				"branch %d description = %q, want %q",
				index,
				got[index].Description,
				test.description,
			)
		}
		if got[index].Kind != KindBranch {
			t.Fatalf("branch %d kind = %q, want %q", index, got[index].Kind, KindBranch)
		}
	}
}

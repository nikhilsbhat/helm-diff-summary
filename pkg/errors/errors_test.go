package errors

import "testing"

func TestDiffSummaryError(t *testing.T) {
	err := &DiffSummaryError{Message: "failed"}
	if err.Error() != "failed" {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}

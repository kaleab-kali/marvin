package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunShowsHelpWithoutArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "marvin analyze <cost-explorer.csv>") {
		t.Fatalf("expected usage in stdout, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunShowsVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got, want := stdout.String(), "marvin "+Version+"\n"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"nope"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), `unknown command "nope"`) {
		t.Fatalf("expected unknown command error, got %q", stderr.String())
	}
}

func TestAnalyzeRequiresPath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"analyze"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "requires a Cost Explorer CSV path") {
		t.Fatalf("expected path error, got %q", stderr.String())
	}
}

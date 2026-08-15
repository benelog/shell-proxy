package executor

import (
	"strings"
	"testing"
	"time"
)

func TestExitCodeShouldBe0(t *testing.T) {
	res, err := New(0).Execute("echo hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", res.ExitCode)
	}
}

func TestExitCodeShouldNotBe0(t *testing.T) {
	res, err := New(6 * time.Second).Execute("exit 3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ExitCode != 3 {
		t.Errorf("exit code = %d, want 3", res.ExitCode)
	}
}

func TestStdoutShouldBeCaptured(t *testing.T) {
	res, err := New(6 * time.Second).Execute("echo 33")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimSpace(res.StandardOutput); got != "33" {
		t.Errorf("stdout = %q, want %q", got, "33")
	}
}

func TestStderrShouldBeCaptured(t *testing.T) {
	res, err := New(6 * time.Second).Execute("echo oops 1>&2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimSpace(res.StandardError); got != "oops" {
		t.Errorf("stderr = %q, want %q", got, "oops")
	}
}

func TestTimeoutIsReported(t *testing.T) {
	res, err := New(200 * time.Millisecond).Execute("sleep 5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.TimedOut {
		t.Errorf("expected TimedOut to be true")
	}
}

func TestPipeIsSupported(t *testing.T) {
	res, err := New(6 * time.Second).Execute("echo hello world | wc -w")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimSpace(res.StandardOutput); got != "2" {
		t.Errorf("stdout = %q, want %q", got, "2")
	}
}

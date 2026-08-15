package auth

import (
	"strings"
	"testing"
)

func TestResolveUsesGivenCredentials(t *testing.T) {
	c, err := Resolve("alice", "hunter2")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if c.Username != "alice" || c.Password != "hunter2" {
		t.Errorf("got %s/%s, want alice/hunter2", c.Username, c.Password)
	}
	if c.DefaultedUsername || c.GeneratedPassword {
		t.Errorf("credentials marked as defaulted, want both explicit")
	}
}

func TestResolveDefaultsToOSUserAndRandomPassword(t *testing.T) {
	c, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if c.Username == "" {
		t.Error("username is empty, want the OS login account")
	}
	if !c.DefaultedUsername || !c.GeneratedPassword {
		t.Error("defaults were not flagged as such")
	}
	if len(c.Password) < 16 {
		t.Errorf("generated password %q is shorter than expected", c.Password)
	}
}

func TestResolveGeneratesADifferentPasswordEachTime(t *testing.T) {
	first, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	second, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if first.Password == second.Password {
		t.Error("two runs produced the same password")
	}
}

func TestOSUsernameStripsWindowsDomain(t *testing.T) {
	// The stripping logic is inlined in osUsername; check it does not leak a
	// domain prefix on any platform.
	if name := osUsername(); strings.ContainsAny(name, `\/`) {
		t.Errorf("osUsername() = %q, want no domain separator", name)
	}
}

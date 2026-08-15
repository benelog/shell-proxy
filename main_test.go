package main

import "testing"

func TestParsePortDefaultsWhenEmpty(t *testing.T) {
	if got := parsePort(nil); got != DefaultPort {
		t.Errorf("parsePort(nil) = %d, want %d", got, DefaultPort)
	}
}

func TestParsePortDefaultsWhenNotANumber(t *testing.T) {
	if got := parsePort([]string{"abc"}); got != DefaultPort {
		t.Errorf("parsePort([abc]) = %d, want %d", got, DefaultPort)
	}
}

func TestParsePortParsesNumber(t *testing.T) {
	if got := parsePort([]string{"18083"}); got != 18083 {
		t.Errorf("parsePort([18083]) = %d, want 18083", got)
	}
}

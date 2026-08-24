package xdg

import (
	"path/filepath"
	"strings"
	"testing"
)

// A relative XDG value would root a separate registry under every working
// directory, and point uninstall's RemoveAll at whichever one it was run
// from. The spec says to ignore it, and so do we.
func TestRelativeXDGValuesAreIgnored(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "relcfg")

	got := ConfigDir(App)
	if !filepath.IsAbs(got) {
		t.Errorf("ConfigDir = %q, want an absolute path", got)
	}
	if strings.Contains(got, "relcfg") {
		t.Errorf("ConfigDir = %q, want the relative value ignored", got)
	}
}

func TestAbsoluteXDGValuesAreUsed(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/cfg")
	if got, want := ConfigDir(App), filepath.Join("/tmp/cfg", App); got != want {
		t.Errorf("ConfigDir = %q, want %q", got, want)
	}
}

func TestTheDefaultSitsUnderHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	if got, want := ConfigDir(App), filepath.Join(Home(), ".config", App); got != want {
		t.Errorf("ConfigDir = %q, want %q", got, want)
	}
}

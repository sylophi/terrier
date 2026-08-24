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
	t.Setenv("XDG_DATA_HOME", "../reldata")

	for name, got := range map[string]string{"ConfigDir": ConfigDir(App), "DataDir": DataDir(App)} {
		if !filepath.IsAbs(got) {
			t.Errorf("%s = %q, want an absolute path", name, got)
		}
		if strings.Contains(got, "relcfg") || strings.Contains(got, "reldata") {
			t.Errorf("%s = %q, want the relative value ignored", name, got)
		}
	}
}

func TestAbsoluteXDGValuesAreUsed(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/cfg")
	if got, want := ConfigDir(App), filepath.Join("/tmp/cfg", App); got != want {
		t.Errorf("ConfigDir = %q, want %q", got, want)
	}
}

func TestDefaultsSitUnderHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	if got, want := ConfigDir(App), filepath.Join(Home(), ".config", App); got != want {
		t.Errorf("ConfigDir = %q, want %q", got, want)
	}
	if got, want := DataDir(App), filepath.Join(Home(), ".local", "share", App); got != want {
		t.Errorf("DataDir = %q, want %q", got, want)
	}
}

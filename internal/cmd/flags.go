package cmd

import (
	"fmt"
	"strings"
)

// flags declares the boolean switches a command accepts, keyed without
// dashes. Registering several keys against the same pointer gives a flag
// its aliases. Terrier has no flag that takes a value, which keeps
// parsing to this.
type flags map[string]*bool

// yesFlag is the one flag set with an alias worth spelling in a single
// place. A lone --json is clearer written out at the command that takes it.
func yesFlag(target *bool) flags {
	return flags{"y": target, "yes": target}
}

// parse pulls the declared flags out of args and returns the positionals.
// Flags may appear anywhere, and a bare "--" ends flag parsing so a path
// beginning with a dash can still be written literally.
func (f flags) parse(args []string, usage string) ([]string, error) {
	positional := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if len(arg) < 2 || !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			continue
		}
		name := strings.TrimLeft(arg, "-")
		target, declared := f[name]
		if !declared {
			return nil, fmt.Errorf("unknown flag: %s\n%s", arg, usage)
		}
		*target = true
	}
	return positional, nil
}

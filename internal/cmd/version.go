package cmd

import (
	"fmt"
)

// Contract is the version of terrier's promises to other tools: the shape
// of what --json prints, and the exit codes commands answer with.
//
// It is deliberately not the release version. Fields are only ever added
// to the JSON, so a tool written against contract 1 keeps working across
// every release that stays compatible, and this number does not move.
// It moves only when something a tool could already be relying on
// changes or goes away, which is the one case where a tool is better off
// refusing to run than guessing.
//
// A tool should read this, compare it against the highest contract it was
// written for, and stop with a clear message when terrier reports a higher
// one.
const Contract = 1

const versionUsage = "usage: terrier version [--json]"

// Version prints what is installed. The --json form is what a tool reads
// to decide whether it understands this terrier.
func Version(args []string, version string) error {
	var asJSON bool
	rest, err := flags{"json": &asJSON}.parse(args, versionUsage)
	if err != nil {
		return err
	}
	if len(rest) > 0 {
		return fmt.Errorf("unexpected argument: %s\n%s", rest[0], versionUsage)
	}

	if asJSON {
		return writeJSON(struct {
			Version  string `json:"version"`
			Contract int    `json:"contract"`
		}{version, Contract})
	}
	fmt.Println(version)
	return nil
}

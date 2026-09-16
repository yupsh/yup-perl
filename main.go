// Command yup-perl is the CLI wrapper around github.com/gloo-foo/cmd-perl.
package main

import (
	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-perl"
	urf "github.com/urfave/cli/v3"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const (
	name          = "perl"
	flagLoop      = "loop"
	flagPrint     = "print"
	flagAutoSplit = "autosplit"
)

// Error is the package's sentinel error type, so every emitted error path is
// comparable with errors.Is.
type Error string

func (e Error) Error() string { return string(e) }

// ErrMissingScript is emitted when no script operand is supplied.
const ErrMissingScript Error = "no script given"

// synopsis is the multi-line --help usage block; urfave/cli indents it three
// spaces, so the lines stay flush-left.
const synopsis = `perl [OPTIONS] SCRIPT

Run SCRIPT over standard input via perl -e.`

// spec declares the perl wrapper: a stdin filter whose single operand is the
// perl script and whose flags configure the -e loop.
var spec = clix.Spec{
	Name:     name,
	Summary:  "perl command wrapper for yupsh",
	Synopsis: synopsis,
	Build:    build,
	Flags: []urf.Flag{
		&urf.BoolFlag{
			Name:    flagLoop,
			Aliases: []string{"n"},
			Usage:   "assume loop around script",
			Sources: urf.EnvVars("YUP_PERL_LOOP"),
		},
		&urf.BoolFlag{
			Name:    flagPrint,
			Aliases: []string{"p"},
			Usage:   "assume loop like -n but print line also",
			Sources: urf.EnvVars("YUP_PERL_PRINT"),
		},
		&urf.BoolFlag{
			Name:    flagAutoSplit,
			Aliases: []string{"a"},
			Usage:   "autosplit mode with -n or -p",
			Sources: urf.EnvVars("YUP_PERL_AUTOSPLIT"),
		},
	},
}

// build maps the invocation to perl's pipeline: standard input feeds perl, whose
// single operand is the script. A bare invocation is a usage error.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	operands := inv.Args.Args().Slice()
	if len(operands) == 0 {
		return nil, nil, ErrMissingScript
	}
	opts := append(options(inv.Args), command.PerlScript(operands[0]))
	return clix.Stdin(inv.Stdin), command.Perl(opts...), nil
}

// options folds the parsed flags into perl's option values. -p implies -n, so it
// enables both the print and loop options.
func options(c *urf.Command) []any {
	var opts []any
	if c.Bool(flagLoop) {
		opts = append(opts, command.PerlLoop)
	}
	if c.Bool(flagPrint) {
		opts = append(opts, command.PerlPrint, command.PerlLoop)
	}
	if c.Bool(flagAutoSplit) {
		opts = append(opts, command.PerlAutoSplit)
	}
	return opts
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }

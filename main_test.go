package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	clix "github.com/gloo-foo/cli"
	"github.com/spf13/afero"
	urf "github.com/urfave/cli/v3"
)

// parse runs args through a bare command carrying the wrapper's flags and
// returns the parsed accessor, so flag-dependent helpers are tested against real
// parsed flags.
func parse(t *testing.T, args ...string) *urf.Command {
	t.Helper()
	var got *urf.Command
	app := &urf.Command{
		Name:   name,
		Flags:  spec.Flags,
		Action: func(_ context.Context, c *urf.Command) error { got = c; return nil },
	}
	if err := app.Run(context.Background(), args); err != nil {
		t.Fatalf("parse: %v", err)
	}
	return got
}

func invocation(t *testing.T, args ...string) clix.Invocation {
	return clix.Invocation{Args: parse(t, args...), Stdin: strings.NewReader(""), Fs: afero.NewMemMapFs()}
}

func TestOptions(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"none", []string{name, "s"}, 0},
		{"loop", []string{name, "-n", "s"}, 1},
		{"print", []string{name, "-p", "s"}, 2},
		{"autosplit", []string{name, "-a", "s"}, 1},
		{"all", []string{name, "-n", "-p", "-a", "s"}, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(options(parse(t, tc.args...))); got != tc.want {
				t.Fatalf("options len=%d, want %d", got, tc.want)
			}
		})
	}
}

func TestBuild_MissingScriptIsUsageError(t *testing.T) {
	src, filter, err := build(invocation(t, name))
	if !errors.Is(err, ErrMissingScript) {
		t.Fatalf("err=%v, want ErrMissingScript", err)
	}
	if src != nil || filter != nil {
		t.Fatalf("src=%v filter=%v, want both nil on error", src, filter)
	}
	if err.Error() != string(ErrMissingScript) {
		t.Fatalf("message=%q, want %q", err.Error(), string(ErrMissingScript))
	}
}

func TestBuild_PassesScriptToPerl(t *testing.T) {
	src, filter, err := build(invocation(t, name, "print"))
	if err != nil || src == nil || filter == nil {
		t.Fatalf("build: src=%v filter=%v err=%v", src, filter, err)
	}
}

func Test_main(t *testing.T) {
	orig := runMain
	t.Cleanup(func() { runMain = orig })
	var gotName clix.Name
	runMain = func(s clix.Spec, _ clix.Version) { gotName = s.Name }
	main()
	if gotName != name {
		t.Fatalf("main used spec %q, want %s", gotName, name)
	}
}

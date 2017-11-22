package scp

import (
	"errors"

	"github.com/google/shlex"
	"github.com/pborman/getopt"
)

type Options struct {
	SourceMode           bool
	TargetMode           bool
	TargetIsDirectory    bool
	Verbose              bool
	PreserveTimesAndMode bool
	Recursive            bool
	Quiet                bool

	Sources []string
	Target  string
}

func ParseCommand(command string) ([]string, error) {
	args, err := shlex.Split(command)
	if err != nil {
		return []string{}, err
	}
	return args, err
}

func ParseFlags(args []string) (*Options, error) {
	cmd := args[0]

	if cmd != "scp" {
		return nil, errors.New("Usage: call scp")
	}

	opts := getopt.New()

	targetMode := opts.Bool('t', "", "Sets target mode for scp")
	opts.Lookup('t').SetOptional()

	sourceMode := opts.Bool('f', "", "Sets source mode for scp")
	opts.Lookup('f').SetOptional()

	targetIsDirectory := opts.Bool('d', "", "Indicates that the target is a directory")
	opts.Lookup('d').SetOptional()

	verbose := opts.Bool('v', "", "Indicates that the command should be run in verbose mode")
	opts.Lookup('v').SetOptional()

	preserveTimesAndMode := opts.Bool('p', "", "Indicates that scp should preserve timestamps and mode of files/directories transferred")
	opts.Lookup('p').SetOptional()

	recursive := opts.Bool('r', "", "Indicates a recursive transfer, must be set if source is a directory")
	opts.Lookup('r').SetOptional()

	// showprogress option is not used but can be provided
	quiet := opts.Bool('q', "", "Indicates that the user wishes to run in quiet mode")
	opts.Lookup('q').SetOptional()

	err := opts.Getopt(args, nil)
	if err != nil {
		return nil, err
	}

	if *targetMode == *sourceMode {
		return nil, errors.New("Must specify either target mode(-t) or source mode(-f) at a time")
	}

	var sources []string
	var target string

	if *sourceMode {
		if len(opts.Args()) < 1 {
			return nil, errors.New("Must specify at least one source in source mode")
		}

		sources = opts.Args()
	}

	if *targetMode {
		if len(opts.Args()) != 1 {
			return nil, errors.New("Must specify one target in target mode")
		}

		target = opts.Args()[0]
	}

	return &Options{
		TargetMode:           *targetMode,
		SourceMode:           *sourceMode,
		TargetIsDirectory:    *targetIsDirectory,
		Verbose:              *verbose,
		PreserveTimesAndMode: *preserveTimesAndMode,
		Recursive:            *recursive,
		Quiet:                *quiet,
		Sources:              sources,
		Target:               target,
	}, nil
}

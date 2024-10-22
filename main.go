package godnf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/CREDOProject/sharedutils/shell"
)

var execCommander = shell.New

// Dnf represents the DNF client.
type Dnf struct {
	binaryPath string
}

// Options represents the configuration options for the running command.
type Options struct {
	Verbose      bool
	DryRun       bool
	Output       io.Writer
	NotAssumeYes bool
	DestDir      string
}

// Returns a new godnf value, which represents an initialized DNF client.
func New(binaryPath string) *Dnf {
	if binaryPath == "" {
		if binaryPath, err := DetectdnfBinary(); err == nil {
			return &Dnf{binaryPath}
		} else {
			return nil
		}
	}
	return &Dnf{binaryPath}
}

var (
	errPackageNameNotSpecified = errors.New("packageName was not specified.")
)

// Install a dnf package from its packageName.
func (a *Dnf) Install(packageName string, opt *Options) error {
	_, err := a.runner(
		&runnerParams{
			argumentBuilder: func() ([]string, error) {
				if strings.TrimSpace(packageName) == "" {
					return nil, fmt.Errorf("Install: %v", errPackageNameNotSpecified)
				}
				return []string{"install", packageName}, nil
			},
			parser: func(string) ([]Package, error) {
				return nil, nil
			},
			opt: opt,
		})
	return err
}

// Update a packages from is packageName. If packageName is empty, updates all
// the packages in the system.
func (a *Dnf) Update(packageName string, opt *Options) error {
	_, err := a.runner(&runnerParams{
		argumentBuilder: func() ([]string, error) {
			if strings.TrimSpace(packageName) == "" {
				return []string{"update"}, nil
			}
			return []string{"update", packageName}, nil
		},
		parser: func(string) ([]Package, error) {
			return nil, nil
		},
		opt: opt,
	})
	return err
}

// Obtains a list of dependencies from a packageName.
func (a *Dnf) Depends(packageName string, opt *Options) ([]Package, error) {
	return a.runner(&runnerParams{
		argumentBuilder: func() ([]string, error) {
			if strings.TrimSpace(packageName) == "" {
				return nil, fmt.Errorf("Depends: %v", errPackageNameNotSpecified)
			}
			return []string{"repoquery", "--deplist", packageName}, nil
		},
		parser: func(output string) ([]Package, error) {
			dependencies := []Package{}
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" {
					// Stop when there's two versions of the same package.
					return dependencies, nil
				}
				if strings.HasPrefix(trimmed, "provider:") {
					parts := strings.Split(trimmed, ": ")
					if len(parts) == 2 {
						dependencies = append(dependencies, Package{
							Name: parts[1],
						})
					}
				}
			}
			return dependencies, nil
		},
		opt: opt,
	})
}

// Remove a package from its packageName.
func (a *Dnf) Remove(packageName string, opt *Options) error {
	_, err := a.runner(&runnerParams{
		argumentBuilder: func() ([]string, error) {
			if strings.TrimSpace(packageName) == "" {
				return nil, fmt.Errorf("Remove: %v", errPackageNameNotSpecified)
			}
			return []string{"remove", packageName}, nil
		},
		parser: func(string) ([]Package, error) {
			return nil, nil
		},
		opt: opt,
	})
	return err
}

// Search a package from its packageName.
func (a *Dnf) Search(packageName string, opt *Options) error {
	_, err := a.runner(&runnerParams{
		argumentBuilder: func() ([]string, error) {
			if strings.TrimSpace(packageName) == "" {
				return nil, fmt.Errorf("Remove: %v", errPackageNameNotSpecified)
			}
			return []string{}, nil
		},
		parser: func(string) ([]Package, error) {
			return nil, nil
		},
		opt: opt,
	})
	return err
}

// List all installed packages.
func (a *Dnf) List(opt *Options) error {
	_, err := a.runner(&runnerParams{
		argumentBuilder: func() ([]string, error) {
			return []string{"list", "installed"}, nil
		},
		parser: func(string) ([]Package, error) {
			return nil, nil
		},
		opt: opt,
	})
	return err
}

type runnerParams struct {
	argumentBuilder func() ([]string, error)
	parser          func(string) ([]Package, error)
	opt             *Options
}

// runner runs a guest command with opt *Options.
func (a *Dnf) runner(params *runnerParams) ([]Package, error) {
	arguments, err := params.argumentBuilder()
	if err != nil {
		return nil, fmt.Errorf("runner: %v", err)
	}
	arguments = append(arguments, processOptions(params.opt)...)
	command := execCommander().Command(a.binaryPath, arguments...)

	var buffer bytes.Buffer
	command.Stdout = &buffer
	if params.opt.Output != nil {
		command.Stdout = io.MultiWriter(command.Stdout, params.opt.Output)
		command.Stderr = io.MultiWriter(command.Stderr, params.opt.Output)
	}
	err = command.Run()
	if err != nil {
		return nil, err
	}
	parsed, err := params.parser(buffer.String())
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

// processOptions returns a slice of command-line flags to be passed to dnf.
func processOptions(opt *Options) []string {
	args := []string{}
	if opt.DryRun {
		args = append(args, "--setopt", "tsflags=test")
	}
	if opt.Verbose {
		args = append(args, "--verbose")
	}
	if !opt.NotAssumeYes {
		args = append(args, "--assumeyes")
	}
	if opt.DestDir != "" {
		args = append(args, "--destdir", opt.DestDir)
	}
	return args
}

// Package represents a DNF package.
type Package struct {
	Name    string
	Version string
	Path    string
}

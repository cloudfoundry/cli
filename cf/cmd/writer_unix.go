//go:build !windows
// +build !windows

package cmd

import "os"

var Writer = os.Stdout

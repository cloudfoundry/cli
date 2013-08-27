// Copied from https://code.google.com/p/gopass/

// +build darwin freebsd linux netbsd openbsd

package terminal

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	sttyArg0   = "/bin/stty"
	exec_cwdir = ""
)

// Tells the terminal to turn echo off.
var sttyArgvEOff []string = []string{"stty", "-echo"}

// Tells the terminal to turn echo on.
var sttyArgvEOn []string = []string{"stty", "echo"}

var ws syscall.WaitStatus = 0

func (ui TerminalUI) AskForPassword(prompt string, args ...interface{}) (passwd string) {
	sig := make(chan os.Signal, 10)
	brk := make(chan bool)

	// Display the prompt.
	fmt.Println("")
	fmt.Printf(prompt+" ", args...)

	// File descriptors for stdin, stdout, and stderr.
	fd := []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()}

	// Setup notifications of termination signals to channel sig, create a process to
	// watch for these signals so we can turn back on echo if need be.
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT,
		syscall.SIGTERM)
	go catchSignal(fd, sig, brk)

	// Turn off the terminal echo.
	pid, err := echoOff(fd)
	if err != nil {
		return
	}

	// Turn on the terminal echo and stop listening for signals.
	defer close(brk)
	defer echoOn(fd)

	rd := bufio.NewReader(os.Stdin)
	syscall.Wait4(pid, &ws, 0, nil)

	line, err := rd.ReadString('\n')
	if err == nil {
		passwd = strings.TrimSpace(line)
	}

	// Carraige return after the user input.
	fmt.Println("")

	return
}

func echoOff(fd []uintptr) (int, error) {
	pid, err := syscall.ForkExec(sttyArg0, sttyArgvEOff, &syscall.ProcAttr{Dir: exec_cwdir, Files: fd})

	if err != nil {
		return 0, fmt.Errorf("failed turning off console echo for password entry:\n\t%s", err)
	}

	return pid, nil
}

// echoOn turns back on the terminal echo.
func echoOn(fd []uintptr) {
	// Turn on the terminal echo.
	pid, e := syscall.ForkExec(sttyArg0, sttyArgvEOn, &syscall.ProcAttr{Dir: exec_cwdir, Files: fd})

	if e == nil {
		syscall.Wait4(pid, &ws, 0, nil)
	}
}

// catchSignal tries to catch SIGKILL, SIGQUIT and SIGINT so that we can turn terminal
// echo back on before the program ends.  Otherwise the user is left with echo off on
// their terminal.
func catchSignal(fd []uintptr, sig chan os.Signal, brk chan bool) {
	select {
	case <-sig:
		echoOn(fd)
		os.Exit(-1)
	case <-brk:
	}
}

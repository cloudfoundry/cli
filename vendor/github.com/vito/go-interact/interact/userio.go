package interact

import (
	"errors"
	"fmt"
	"io"

	"github.com/vito/go-interact/interact/terminal"
)

type userIO interface {
	WriteLine(line string) error

	ReadLine(prompt string) (string, error)
	ReadPassword(prompt string) (string, error)
}

type ttyUser struct {
	*terminal.Terminal
}

var ErrKeyboardInterrupt = errors.New("keyboard interrupt")

func newTTYUser(input io.Reader, output io.Writer) ttyUser {
	return ttyUser{
		Terminal: terminal.NewTerminal(readWriter{input, output}, ""),
	}
}

func (u ttyUser) WriteLine(line string) error {
	_, err := fmt.Fprintf(u.Terminal, "%s\r\n", line)
	return err
}

func (u ttyUser) ReadLine(prompt string) (string, error) {
	u.Terminal.SetPrompt(prompt)
	input, err := u.Terminal.ReadLine()
	if err == terminal.ErrKeyboardInterrupt {
		return input, ErrKeyboardInterrupt
	}

	return input, err
}

type nonTTYUser struct {
	io.Reader
	io.Writer
}

func newNonTTYUser(input io.Reader, output io.Writer) nonTTYUser {
	return nonTTYUser{
		Reader: input,
		Writer: output,
	}
}

func (u nonTTYUser) WriteLine(line string) error {
	_, err := fmt.Fprintf(u.Writer, "%s\n", line)
	return err
}

func (u nonTTYUser) ReadLine(prompt string) (string, error) {
	_, err := fmt.Fprintf(u.Writer, "%s", prompt)
	if err != nil {
		return "", err
	}

	line, err := u.readLine()
	if err != nil {
		return "", err
	}

	_, err = fmt.Fprintf(u.Writer, "%s\n", line)
	if err != nil {
		return "", err
	}

	return line, nil
}

func (u nonTTYUser) ReadPassword(prompt string) (string, error) {
	_, err := fmt.Fprintf(u.Writer, "%s", prompt)
	if err != nil {
		return "", err
	}

	line, err := u.readLine()
	if err != nil {
		return "", err
	}

	_, err = fmt.Fprintf(u.Writer, "\n")
	if err != nil {
		return "", err
	}

	return line, nil
}

func (u nonTTYUser) readLine() (string, error) {
	var line string

	for {
		chr := make([]byte, 1)
		n, err := u.Reader.Read(chr)

		if n == 1 {
			if chr[0] == '\n' {
				return line, nil
			}

			line += string(chr)
		}

		if err != nil {
			return "", err
		}
	}
}

type readWriter struct {
	io.Reader
	io.Writer
}

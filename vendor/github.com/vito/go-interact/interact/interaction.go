package interact

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"

	"golang.org/x/crypto/ssh/terminal"
)

// Interaction represents a single question to ask, optionally with a set of
// choices to limit the answer to.
type Interaction struct {
	Prompt  string
	Choices []Choice

	Input  io.Reader
	Output io.Writer
}

// NewInteraction constructs an interaction with the given prompt, limited to
// the given choices, if any.
//
// Defaults Input and Output to os.Stdin and os.Stderr, respectively.
func NewInteraction(prompt string, choices ...Choice) Interaction {
	return Interaction{
		Input:   os.Stdin,
		Output:  os.Stdout,
		Prompt:  prompt,
		Choices: choices,
	}
}

// Resolve prints the prompt, indicating the default value, and asks for the
// value to populate into the destination dst, which should be a pointer to a
// value to set.
//
// The default value is whatever value is currently held in dst, and will be
// shown in the prompt. Note that zero-values are valid defaults (e.g. false
// for a boolean prompt), so to disambiguate from having just allocated dst,
// and not intending its current zero-value to be the default, you must wrap it
// in a RequiredDestination.
//
// If the choices are limited, the default value will be inferred by finding
// the value held in dst within the set of choices. The number corresponding
// to the choice will be the default value shown to the user. If no default is
// found, Resolve will require the user to make a selection.
//
// The type of dst determines how the value is read. Currently supported types
// for the destination are int, string, bool, and any arbitrary value that is
// defined within the set of Choices.
//
// Valid input strings for bools are "y", "n", "Y", "N", "yes", and "no".
// Integer values are parsed in base-10. String values will not include any
// trailing linebreak.
func (interaction Interaction) Resolve(dst interface{}) error {
	prompt := interaction.prompt(dst)

	var user userIO
	if file, ok := interaction.Input.(*os.File); ok && terminal.IsTerminal(int(file.Fd())) {
		state, err := terminal.MakeRaw(int(file.Fd()))
		if err != nil {
			return err
		}

		defer terminal.Restore(int(file.Fd()), state)

		user = newTTYUser(interaction.Input, interaction.Output)
	} else {
		user = newNonTTYUser(interaction.Input, interaction.Output)
	}

	if len(interaction.Choices) == 0 {
		return interaction.resolveSingle(dst, user, prompt)
	}

	return interaction.resolveChoices(dst, user, prompt)
}

func (interaction Interaction) prompt(dst interface{}) string {
	if len(interaction.Choices) > 0 {
		num, present := interaction.choiceNumber(dst)
		if present {
			return fmt.Sprintf("%s (%d): ", interaction.Prompt, num)
		}

		return fmt.Sprintf("%s: ", interaction.Prompt)
	}

	switch v := dst.(type) {
	case RequiredDestination:
		switch v.Destination.(type) {
		case *bool:
			return fmt.Sprintf("%s [yn]: ", interaction.Prompt)
		default:
			return fmt.Sprintf("%s: ", interaction.Prompt)
		}
	case *int:
		return fmt.Sprintf("%s (%d): ", interaction.Prompt, *v)
	case *string:
		return fmt.Sprintf("%s (%s): ", interaction.Prompt, *v)
	case *bool:
		var indicator string
		if *v {
			indicator = "Yn"
		} else {
			indicator = "yN"
		}

		return fmt.Sprintf("%s [%s]: ", interaction.Prompt, indicator)
	case *Password:
		if len(*v) == 0 {
			return fmt.Sprintf("%s (): ", interaction.Prompt)
		}

		return fmt.Sprintf("%s (has default): ", interaction.Prompt)
	default:
		return fmt.Sprintf("%s (unknown): ", interaction.Prompt)
	}
}

func (interaction Interaction) choiceNumber(dst interface{}) (int, bool) {
	for i, c := range interaction.Choices {
		dstVal := reflect.ValueOf(dst).Elem()

		if c.Value == nil && dstVal.IsNil() {
			return i + 1, true
		}

		if reflect.DeepEqual(c.Value, dstVal.Interface()) {
			return i + 1, true
		}
	}

	return 0, false
}

func (interaction Interaction) resolveSingle(dst interface{}, user userIO, prompt string) error {
	for {
		_, retry, err := interaction.readInto(dst, user, prompt)
		if err == io.EOF {
			return err
		}

		if err != nil {
			if retry {
				user.WriteLine(fmt.Sprintf("invalid input (%s)", err))
				continue
			} else {
				return err
			}
		}

		break
	}

	return nil
}

func (interaction Interaction) resolveChoices(dst interface{}, user userIO, prompt string) error {
	dstVal := reflect.ValueOf(dst)

	for i, choice := range interaction.Choices {
		err := user.WriteLine(fmt.Sprintf("%d: %s", i+1, choice.Display))
		if err != nil {
			return err
		}
	}

	for {
		var retry bool
		var err error

		num, present := interaction.choiceNumber(dst)
		if present {
			_, retry, err = interaction.readInto(&num, user, prompt)
		} else {
			_, retry, err = interaction.readInto(Required(&num), user, prompt)
		}

		if err == io.EOF {
			return err
		}

		if err != nil {
			if retry {
				user.WriteLine(fmt.Sprintf("invalid selection (%s)", err))
				continue
			} else {
				return err
			}
		}

		if num == 0 || num > len(interaction.Choices) {
			user.WriteLine(fmt.Sprintf("invalid selection (must be 1-%d)", len(interaction.Choices)))
			continue
		}

		choice := interaction.Choices[num-1]

		if choice.Value == nil {
			dstVal.Elem().Set(reflect.Zero(dstVal.Type().Elem()))
		} else {
			choiceVal := reflect.ValueOf(choice.Value)

			if choiceVal.Type().AssignableTo(dstVal.Type().Elem()) {
				dstVal.Elem().Set(choiceVal)
			} else {
				return NotAssignableError{
					Value:       choiceVal.Type(),
					Destination: dstVal.Type().Elem(),
				}
			}
		}

		return nil
	}
}

func (interaction Interaction) readInto(dst interface{}, user userIO, prompt string) (bool, bool, error) {
	switch v := dst.(type) {
	case RequiredDestination:
		for {
			read, retry, err := interaction.readInto(v.Destination, user, prompt)
			if err != nil {
				return false, retry, err
			}

			if read {
				return true, false, nil
			}
		}

	case *int:
		line, err := user.ReadLine(prompt)
		if err != nil {
			return false, false, err
		}

		if len(line) == 0 {
			return false, false, nil
		}

		num, err := strconv.Atoi(line)
		if err != nil {
			return false, true, ErrNotANumber
		}

		*v = num

		return true, false, nil

	case *string:
		line, err := user.ReadLine(prompt)
		if err != nil {
			return false, false, err
		}

		if len(line) == 0 {
			return false, false, nil
		}

		*v = line

		return true, false, nil

	case *Password:
		pass, err := user.ReadPassword(prompt)
		if err != nil {
			return false, false, err
		}

		if len(pass) == 0 {
			return false, false, nil
		}

		*v = Password(pass)

		return true, false, nil

	case *bool:
		line, err := user.ReadLine(prompt)
		if err != nil {
			return false, false, err
		}

		if len(line) == 0 {
			return false, false, nil
		}

		switch line {
		case "y", "Y", "yes":
			*v = true
		case "n", "N", "no":
			*v = false
		default:
			return false, true, ErrNotBoolean
		}

		return true, false, nil
	}

	return false, false, fmt.Errorf("unknown destination type: %T", dst)
}

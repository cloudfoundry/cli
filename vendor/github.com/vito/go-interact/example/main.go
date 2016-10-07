package main

import (
	"fmt"
	"os"

	"github.com/vito/go-interact/interact"
)

func main() {
	// Well? [yN]:
	boolAnsFalseDefault := false
	err := interact.NewInteraction("Well?").Resolve(&boolAnsFalseDefault)
	if err != nil {
		fatal(err)
	}

	fmt.Println(boolAnsFalseDefault)

	// Well? [Yn]:
	boolAnsTrueDefault := true
	err = interact.NewInteraction("Well?").Resolve(&boolAnsTrueDefault)
	if err != nil {
		fatal(err)
	}

	fmt.Println(boolAnsTrueDefault)

	// Well? [yn]:
	var boolAnsNoDefault bool
	err = interact.NewInteraction("Well?").Resolve(interact.Required(&boolAnsNoDefault))
	if err != nil {
		fatal(err)
	}

	fmt.Println(boolAnsNoDefault)

	// Message (hello):
	strAnsDefault := "hello"
	err = interact.NewInteraction("Message").Resolve(&strAnsDefault)
	if err != nil {
		fatal(err)
	}

	fmt.Println(strAnsDefault)

	// Message ():
	var strAnsEmptyDefault string
	err = interact.NewInteraction("Message").Resolve(&strAnsEmptyDefault)
	if err != nil {
		fatal(err)
	}

	fmt.Println(strAnsEmptyDefault)

	// Message:
	var strAnsNoDefault string
	err = interact.NewInteraction("Message").Resolve(interact.Required(&strAnsNoDefault))
	if err != nil {
		fatal(err)
	}

	fmt.Println(strAnsNoDefault)

	numbers := []string{"uno", "dos", "tres"}

	// 1: One
	// 2: Two
	// 3: Three
	// Choose a number:
	var chosenFoo string
	err = interact.NewInteraction(
		"Choose a number",
		interact.Choice{Display: "One", Value: numbers[0]},
		interact.Choice{Display: "Two", Value: numbers[1]},
		interact.Choice{Display: "Three", Value: numbers[2]},
	).Resolve(&chosenFoo)
	if err != nil {
		fatal(err)
	}

	fmt.Println(chosenFoo)

	// 1: One
	// 2: Two
	// 3: Three
	// Choose a number (2):
	chosenFooWithDefault := "dos"
	err = interact.NewInteraction(
		"Choose a number",
		interact.Choice{Display: "One", Value: numbers[0]},
		interact.Choice{Display: "Two", Value: numbers[1]},
		interact.Choice{Display: "Three", Value: numbers[2]},
	).Resolve(&chosenFooWithDefault)
	if err != nil {
		fatal(err)
	}

	fmt.Println(chosenFooWithDefault)

	// 1. One
	// 2. Two
	// 3. Three
	// 4. none
	// Choose a number (4):
	var chosenFooOptional *string
	err = interact.NewInteraction(
		"Choose a number",
		interact.Choice{Display: "One", Value: &numbers[0]},
		interact.Choice{Display: "Two", Value: &numbers[1]},
		interact.Choice{Display: "Three", Value: &numbers[2]},
		interact.Choice{Display: "none", Value: nil},
	).Resolve(&chosenFooOptional)
	if err != nil {
		fatal(err)
	}

	fmt.Println(chosenFooOptional)

	// Username:
	var username string
	err = interact.NewInteraction("Username").Resolve(interact.Required(&username))
	if err != nil {
		fatal(err)
	}

	fmt.Println(username)

	// Password:
	var password interact.Password
	err = interact.NewInteraction("Password").Resolve(interact.Required(&password))
	if err != nil {
		fatal(err)
	}

	fmt.Println(password)

	// Interrupt:
	var ctrlCTest string
	err = interact.NewInteraction("Interrupt").Resolve(interact.Required(&ctrlCTest))
	if err != nil {
		fmt.Println(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

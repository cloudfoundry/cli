package commands_test

import (
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"

	"github.com/codegangsta/cli"
)

func TestBashAutocompleteFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}

	fakeUI := callBashAutocomplete([]string{"foo"}, reqFactory)
	assert.True(t, fakeUI.FailedWithUsage)
}

func TestBashAutocompleteRequirements(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}

	callBashAutocomplete([]string{}, reqFactory)
	assert.True(t, testhelpers.CommandDidPassRequirements)
}

func TestBashAutocompleteOutput(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}

	fakeUI := callBashAutocomplete([]string{}, reqFactory)
	assert.Contains(t, fakeUI.Outputs[0], "_cf()")
	assert.Contains(t, fakeUI.Outputs[len(fakeUI.Outputs)-1], "complete -F _cf ")
}

func callBashAutocomplete(args []string, reqFactory *testhelpers.FakeReqFactory) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("bash-autocomplete", args)

	cmd := NewBashAutocomplete(ui, []cli.Command{})
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}

package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type Stacks struct {
	ui         term.UI
	stacksRepo api.StackRepository
}

func NewStacks(ui term.UI, stacksRepo api.StackRepository) (s *Stacks) {
	s = new(Stacks)
	s.ui = ui
	s.stacksRepo = stacksRepo
	return
}

func (s *Stacks) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (s *Stacks) Run(c *cli.Context) {
	s.ui.Say("Getting stacks")

	stacks, err := s.stacksRepo.FindAll()
	if err != nil {
		s.ui.Failed(err.Error())
		return
	}

	s.ui.Ok()

	table := [][]string{
		[]string{"name", "description"},
	}

	for _, stack := range stacks {
		table = append(table, []string{
			stack.Name,
			stack.Description,
		})
	}

	s.ui.DisplayTable(table, nil)
}

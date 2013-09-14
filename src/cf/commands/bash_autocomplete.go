package commands

import (
	"bytes"
	"errors"
	"os"
	"text/template"

	"cf/requirements"
	"cf/terminal"

	"github.com/codegangsta/cli"
)

type BashAutocomplete struct {
	ui       terminal.UI
	commands []cli.Command
}

func NewBashAutocomplete(ui terminal.UI, commands []cli.Command) *BashAutocomplete {
	return &BashAutocomplete{
		ui:       ui,
		commands: commands,
	}
}

func (cmd *BashAutocomplete) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) > 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "bash-autocomplete")
		return
	}

	return
}

func (cmd *BashAutocomplete) Run(c *cli.Context) {
	// Very helpful article on autocompletion:
	// http://www.debian-administration.org/article/317/An_introduction_to_bash_completion_part_2
	completion := `_cf()
{
    local cur prev cmds
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    cmds="{{range .}}{{.Name}} {{end}} bash-autocomplete"

    case "${prev}" in
      {{programName}})
        COMPREPLY=( $(compgen -W "${cmds}" -- ${cur}) )
        return 0
        ;;
      help)
        COMPREPLY=( $(compgen -W "${cmds}" -- ${cur}) )
        return 0
        ;;
    esac
}
complete -F _cf {{programName}}`

	helpers := template.FuncMap{
		"programName": func() string {
			return os.Args[0]
		},
	}

	t := template.Must(template.New("bashCompletion").Funcs(helpers).Parse(completion))

	var output bytes.Buffer
	t.Execute(&output, cmd.commands)
	cmd.ui.Say(output.String())
}

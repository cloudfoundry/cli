package ssh

import (
	"fmt"
	"regexp"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

var windowsDisk = regexp.MustCompile(`^[A-Za-z]:\\`)

type SCPArgs struct {
	raw       []string
	recursive bool
}

func NewSCPArgs(rawArgs []string, recursive bool) SCPArgs {
	return SCPArgs{raw: rawArgs, recursive: recursive}
}

func (a SCPArgs) AllOrInstanceGroupOrInstanceSlug() (boshdir.AllOrInstanceGroupOrInstanceSlug, error) {
	for _, rawArg := range a.raw {
		pieces := strings.SplitN(rawArg, ":", 2)

		if len(pieces) == 2 {
			return boshdir.NewAllOrInstanceGroupOrInstanceSlugFromString(pieces[0])
		}
	}

	err := bosherr.Errorf("Missing remote host information in source/destination arguments")

	return boshdir.AllOrInstanceGroupOrInstanceSlug{}, err
}

func (a SCPArgs) ForHost(host boshdir.Host) []string {
	args := []string{}

	if a.recursive {
		args = append(args, "-r")
	}

	for _, rawArg := range a.raw {
		pieces := strings.SplitN(rawArg, ":", 2)

		if len(pieces) == 2 && !windowsDisk.MatchString(rawArg) {
			// Resolve named host to actual user@ip
			pieces[0] = fmt.Sprintf("%s@%s", host.Username, printableHost{host})
		}

		for i := range pieces {
			pieces[i] = strings.Replace(pieces[i], "((instance_id))", host.IndexOrID, -1)
		}

		args = append(args, strings.Join(pieces, ":"))
	}

	return args
}

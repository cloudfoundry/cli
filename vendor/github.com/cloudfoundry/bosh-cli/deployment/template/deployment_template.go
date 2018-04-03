package template

import (
	gosha512 "crypto/sha512"
	"fmt"

	"github.com/cppforlife/go-patch/patch"

	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
)

type DeploymentTemplate struct {
	template boshtpl.Template
}

func NewDeploymentTemplate(content []byte) DeploymentTemplate {
	return DeploymentTemplate{template: boshtpl.NewTemplate(content)}
}

func (t DeploymentTemplate) Evaluate(vars boshtpl.Variables, op patch.Op) (InterpolatedTemplate, error) {
	bytes, err := t.template.Evaluate(vars, op, boshtpl.EvaluateOpts{ExpectAllKeys: true})
	if err != nil {
		return InterpolatedTemplate{}, err
	}

	sha512 := gosha512.New()

	_, err = sha512.Write(bytes)
	if err != nil {
		panic("Error calculating sha512 of interpolated template")
	}

	shaSumString := fmt.Sprintf("%x", sha512.Sum(nil))

	return NewInterpolatedTemplate(bytes, shaSumString), nil
}

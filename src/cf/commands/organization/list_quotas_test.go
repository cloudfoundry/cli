package organization_test

import (
	"cf"
	. "cf/commands/organization"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestListQuotasRequirements(t *testing.T) {
	quotaRepo := &testapi.FakeQuotaRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callListQuotas(reqFactory, quotaRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callListQuotas(reqFactory, quotaRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestListQuotas(t *testing.T) {
	quotas := []cf.Quota{{Name: "quota-name", MemoryLimit: 1024}}
	quotaRepo := &testapi.FakeQuotaRepository{FindAllQuotas: quotas}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	ui := callListQuotas(reqFactory, quotaRepo)

	assert.Contains(t, ui.Outputs[0], "Getting quotas")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "name")
	assert.Contains(t, ui.Outputs[3], "memory limit")
	assert.Contains(t, ui.Outputs[4], "quota-name")
	assert.Contains(t, ui.Outputs[4], "1G")
}

func callListQuotas(reqFactory *testreq.FakeReqFactory, quotaRepo *testapi.FakeQuotaRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("quotas", []string{})
	cmd := NewListQuotas(fakeUI, quotaRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

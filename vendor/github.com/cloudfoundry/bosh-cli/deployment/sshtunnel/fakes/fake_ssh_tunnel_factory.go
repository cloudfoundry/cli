package fakes

import (
	bisshtunnel "github.com/cloudfoundry/bosh-cli/deployment/sshtunnel"
)

type FakeFactory struct {
	SSHTunnel           bisshtunnel.SSHTunnel
	NewSSHTunnelOptions bisshtunnel.Options
}

func NewFakeFactory() *FakeFactory {
	return &FakeFactory{}
}

func (f *FakeFactory) NewSSHTunnel(options bisshtunnel.Options) bisshtunnel.SSHTunnel {
	f.NewSSHTunnelOptions = options

	return f.SSHTunnel
}

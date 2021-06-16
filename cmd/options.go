package cmd

import "github.com/kubermatic-labs/registryman/pkg/globalregistry"

type cliOptions struct {
	forceDelete bool
}

var _ globalregistry.CanForceDelete = &cliOptions{}

func (o *cliOptions) ForceDeleteProjects() bool {
	return o.forceDelete
}

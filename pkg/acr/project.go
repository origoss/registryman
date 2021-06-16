/*
   Copyright 2021 The Kubermatic Kubernetes Platform contributors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package acr

import (
	"fmt"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type project struct {
	name string
	api  *projectAPI
}

func (p *project) GetName() string {
	return p.name
}

func (p *project) Delete() error {
	repos, err := p.getRepositories()
	if err != nil {
		return err
	}

	if len(repos) > 0 {
		switch opt := p.api.reg.GetOptions().(type) {
		case globalregistry.CanForceDelete:
			if f := opt.ForceDeleteProjects(); !f {
				return fmt.Errorf("%s: repositories are present, please delete them before deleting the project, %w", p.GetName(), globalregistry.RecoverableError)
			}
		}
		for _, repo := range repos {
			p.api.reg.logger.V(1).Info("deleting repository",
				"repositoryName", repo.GetName(),
			)
			err = p.deleteRepository(repo)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *project) AssignMember(member globalregistry.ProjectMember) (*globalregistry.ProjectMemberCredentials, error) {
	return nil, fmt.Errorf("method ACR.AssignMember not implemented: %w", globalregistry.RecoverableError)
}

func (p *project) UnassignMember(member globalregistry.ProjectMember) error {
	return globalregistry.RecoverableError
}

func (p *project) AssignReplicationRule(remoteReg globalregistry.RegistryConfig, trigger globalregistry.ReplicationTrigger, direction globalregistry.ReplicationDirection) (globalregistry.ReplicationRule, error) {
	return nil, globalregistry.RecoverableError
}

func (p *project) GetMembers() ([]globalregistry.ProjectMember, error) {
	p.api.reg.logger.V(-1).Info("ACR.GetMembers not implemented")
	return []globalregistry.ProjectMember{}, nil
}

func (p *project) GetReplicationRules(
	trigger *globalregistry.ReplicationTrigger,
	direction *globalregistry.ReplicationDirection) ([]globalregistry.ReplicationRule, error) {

	return nil, nil
}

func (p *project) getRepositories() ([]globalregistry.Repository, error) {
	return p.api.listProjectRepositories(p)
}

func (p *project) deleteRepository(r globalregistry.Repository) error {
	return p.api.deleteProjectRepository(p, r)
}

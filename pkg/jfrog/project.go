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

package jfrog

import (
	"fmt"
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

// interface guard
var _ globalregistry.Project = &project{}
var _ globalregistry.ProjectWithRepositories = &project{}
var _ globalregistry.ProjectWithMembers = &project{}

var _ globalregistry.MemberManipulatorProject = &project{}
var _ globalregistry.DestructibleProject = &project{}

func (p *project) GetName() string {
	return p.name
}

// Delete removes the project from registry
func (p *project) Delete() error {
	repos, err := p.GetRepositories()
	if err != nil {
		return err
	}

	if len(repos) > 0 {
		switch opt := p.registry.GetOptions().(type) {
		case globalregistry.CanForceDelete:
			if f := opt.ForceDeleteProjects(); !f {
				return fmt.Errorf("%s: repositories are present, please delete them before deleting the project, %w", p.GetName(), globalregistry.ErrRecoverableError)
			}
		}

	}
	return p.registry.delete(p.GetName())
}

func (p *project) AssignMember(member globalregistry.ProjectMember) (*globalregistry.ProjectMemberCredentials, error) {
	role, err := roleFromString(member.GetRole())
	if err != nil {
		return nil, err
	}
	permissionReqBody, err := p.registry.getPermission(p.registry.GetDockerRegistryName() + "_" + p.GetName())
	if err != nil {
		return nil, err
	}

	if permissionReqBody.Principals.Users == nil {
		permissionReqBody.Principals.Users = make(map[string][]string)
	}

	permissionReqBody.Principals.Users[member.GetName()] = strings.Split(role.String(), ",")

	err = p.registry.createPermission(p.GetName(), permissionReqBody)
	return nil, err

}

func (p *project) GetMembers() ([]globalregistry.ProjectMember, error) {
	members, err := p.registry.getMembers(p)
	if err != nil {
		return nil, err
	}
	projectMembers := make([]globalregistry.ProjectMember, len(members))
	for i, m := range members {
		projectMembers[i] = m.toProjectMember()
	}
	return projectMembers, nil
}

func (p *project) UnassignMember(member globalregistry.ProjectMember) error {

	var m *projectMember
	members, err := p.registry.getMembers(p)
	if err != nil {
		return err
	}
	for _, memb := range members {
		if memb.GetName() == member.GetName() {
			m = &memb
			break
		}
	}
	if m == nil {
		return fmt.Errorf("user member not found")
	}

	permissionReqBody, err := p.registry.getPermission(p.registry.GetDockerRegistryName() + "_" + p.GetName())
	if err != nil {
		return err
	}
	delete(permissionReqBody.Principals.Users, m.GetName())

	err = p.registry.createPermission(p.GetName(), permissionReqBody)
	return err
}

func (p *project) GetRepositories() ([]string, error) {
	repos, err := p.registry.listFolders(p.GetName())
	if err != nil {
		return nil, err
	}
	return repos, nil
}

// GetUsedStorage implements the globalregistry.Project interface.
// func (p *project) GetUsedStorage() (int, error) {
// 	return p.registry.getUsedStorage(p)
// }

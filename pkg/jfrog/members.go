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
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const (
	userType  = "User"
	robotType = "Robot"
	groupType = "Group"
)

type projectMember struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

var _ globalregistry.ProjectMember = &projectMember{}

func (m *projectMember) GetName() string {
	return m.Name
}

func (m *projectMember) GetType() string {
	return userType
}

func (m *projectMember) GetRole() string {
	return roleFromList(m.Roles)
}

func (m *projectMember) toProjectMember() globalregistry.ProjectMember {
	p := &projectMember{
		Name:  m.Name,
		Roles: m.Roles,
	}
	return p

}

func (r *registry) getMembers(p *project) ([]projectMember, error) {
	projectMembers, err := p.registry.getPermission(r.GetDockerRegistryName() + "_" + p.GetName())
	if err != nil {
		return nil, err
	}

	projectMembersResult := make([]projectMember, len(projectMembers.Principals.Users))

	c := 0
	for user, roles := range projectMembers.Principals.Users {
		projectMembersResult[c] = projectMember{
			Name:  user,
			Roles: roles,
		}
		c++
	}

	return projectMembersResult, err
}

// func (r *registry) getMembers(p *project) ([]projectMember, error) {
// 	url := *r.parsedUrl
// 	url.Path = fmt.Sprintf("%s/%s/%s", storagePath, r.GetDockerRegistryName(), p.name)
// 	url.RawQuery = "permissions"
// 	r.logger.V(1).Info("creating new request", "url", url.String())
// 	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header["Content-Type"] = []string{"application/json"}
// 	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

// 	resp, err := r.do(req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer resp.Body.Close()

// 	projectMembers := &projectMemberRequestBody{}

// 	err = json.NewDecoder(resp.Body).Decode(&projectMembers)
// 	if err != nil {
// 		r.logger.Error(err, "json decoding failed")
// 		b := bytes.NewBuffer(nil)
// 		_, err := b.ReadFrom(resp.Body)
// 		if err != nil {
// 			panic(err)
// 		}
// 		r.logger.Info(b.String())
// 		fmt.Printf("body: %+v\n", b.String())
// 	}
// 	// TODO: add groups +len(projectMembers.Principals.Groups
// 	projectMembersResult := make([]projectMember, len(projectMembers.Principals.Users))

// 	c := 0
// 	for user, roles := range projectMembers.Principals.Users {
// 		projectMembersResult[c] = projectMember{
// 			Name:  user,
// 			Roles: roles,
// 		}
// 		c++
// 	}

// 	return projectMembersResult, err
// }

// func (r *registry) createProjectMember(projectID int, projectMember *projectMemberRequestBody) (int, error) {
// 	url := *r.parsedUrl
// 	url.Path = fmt.Sprintf("%s/%d/members", path, projectID)
// 	reqBodyBuf := bytes.NewBuffer(nil)
// 	err := json.NewEncoder(reqBodyBuf).Encode(projectMember)
// 	if err != nil {
// 		return 0, err
// 	}
// 	req, err := http.NewRequest(http.MethodPost, url.String(), reqBodyBuf)
// 	if err != nil {
// 		return 0, err
// 	}

// 	req.Header["Content-Type"] = []string{"application/json"}
// 	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

// 	resp, err := r.do(req)
// 	if err != nil {
// 		return 0, err
// 	}

// 	defer resp.Body.Close()
// 	switch resp.StatusCode {
// 	case 409:
// 		return 0, fmt.Errorf("project member cannot be added: %w", globalregistry.ErrAlreadyExists)
// 	case 500:
// 		switch {
// 		case projectMember.MemberUser != nil:
// 			name := projectMember.MemberUser.Username
// 			return 0, fmt.Errorf("internal server error, invalid name? (%s)", name)
// 		case projectMember.MemberGroup != nil:
// 			name := projectMember.MemberGroup.LdapGroupDn
// 			return 0, fmt.Errorf("internal server error, invalid DN? (%s)", name)
// 		default:
// 			panic("projectMember is neither user nor group")
// 		}
// 	}

// 	memberID, err := strconv.Atoi(strings.TrimPrefix(
// 		resp.Header.Get("Location"),
// 		fmt.Sprintf("%s/%d/members/", path, projectID)))
// 	if err != nil {
// 		r.logger.Error(err, "cannot parse member ID from response Location header",
// 			"location-header", resp.Header.Get("Location"))
// 		return 0, err
// 	}

// 	return memberID, nil
// }

// func (r *registry) deleteProjectMember(projectID int, memberId int) error {
// 	url := *r.parsedUrl
// 	url.Path = fmt.Sprintf("%s/%d/members/%d", path, projectID, memberId)
// 	r.logger.V(1).Info("creating new request", "url", url.String())
// 	req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
// 	if err != nil {
// 		return err
// 	}
// 	r.logger.V(1).Info("sending HTTP request", "req-uri", req.URL)

// 	req.Header["Content-Type"] = []string{"application/json"}
// 	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

// 	resp, err := r.do(req)
// 	if err != nil {
// 		return err
// 	}

// 	defer resp.Body.Close()

// 	return nil
// }

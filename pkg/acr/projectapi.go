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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type projectAPI struct {
	reg *registry
}

func newProjectAPI(reg *registry) (*projectAPI, error) {
	return &projectAPI{
		reg: reg,
	}, nil
}

func (p *projectAPI) Create(name string) (globalregistry.Project, error) {
	return nil, fmt.Errorf("method ACR.projectAPI.Create not implemented: %w", globalregistry.RecoverableError)
}

func (p *projectAPI) GetByName(name string) (globalregistry.Project, error) {
	projects, err := p.List()
	if err != nil {
		return nil, err
	}
	for _, project := range projects {
		if project.GetName() == name {
			return project, nil
		}
	}
	return nil, nil
}

type bytesBody struct {
	*bytes.Buffer
}

func (bb bytesBody) Close() error { return nil }

func (s *registry) do(req *http.Request) (*http.Response, error) {
	resp, err := s.Client.Do(req)
	if err != nil {
		s.logger.Error(err, "http.Client cannot Do",
			"req-url", req.URL,
		)
		return nil, err
	}

	buf := bytesBody{
		Buffer: new(bytes.Buffer),
	}
	n, err := buf.ReadFrom(resp.Body)
	if err != nil {
		s.logger.Error(err, "cannot read HTTP response body")
		return nil, err
	}
	resp.Body = buf

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		s.logger.V(-1).Info("HTTP response status code is not OK",
			"status-code", resp.StatusCode,
			"resp-body-size", n,
			"req-url", req.URL,
		)
		s.logger.V(1).Info(buf.String())
	}
	return resp, nil
}

func (p *projectAPI) List() ([]globalregistry.Project, error) {
	p.reg.parsedUrl.Path = path
	req, err := http.NewRequest(http.MethodGet, p.reg.parsedUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	resp, err := p.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	projectData := &repositories{}

	err = json.NewDecoder(resp.Body).Decode(projectData)
	if err != nil {
		p.reg.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		p.reg.logger.Info(b.String())
	}

	pStatus := p.collectProjectNamesFromRepos(projectData.Repositories)

	return pStatus, err
}

func (p *projectAPI) collectProjectNamesFromRepos(repoNames []string) []globalregistry.Project {
	projectNames := make(map[string]struct{})

	for _, pData := range repoNames {
		pName := strings.Split(pData, "/")[0]
		projectNames[pName] = struct{}{}
	}
	pStatus := make([]globalregistry.Project, len(projectNames))

	i := 0
	for projectName := range projectNames {
		pStatus[i] = &project{
			api:  p,
			name: projectName,
		}
		i++
	}
	return pStatus
}

type projectRepository struct {
	name string
	proj *project
}

var _ globalregistry.Repository = &projectRepository{}

func (pr *projectRepository) GetName() string {
	return pr.name
}

func (pr *projectRepository) Delete() error {
	return pr.proj.api.deleteProjectRepository(
		pr.proj,
		pr)
}

func (p *projectAPI) deleteProjectRepository(proj *project, repo globalregistry.Repository) error {
	p.reg.logger.V(1).Info("deleting ACR repository",
		"repositoryName", repo.GetName(),
	)
	url := *p.reg.parsedUrl
	url.Path = fmt.Sprintf("/acr/v1/%s", repo.GetName())
	req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	_, err = p.reg.do(req)
	return err
}

func (p *projectAPI) listProjectRepositories(proj *project) ([]globalregistry.Repository, error) {
	p.reg.logger.V(1).Info("listing project repositories",
		"projectName", proj.GetName(),
	)
	url := *p.reg.parsedUrl
	url.Path = path
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	resp, err := p.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	projectData := &repositories{}

	err = json.NewDecoder(resp.Body).Decode(projectData)
	if err != nil {
		p.reg.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		p.reg.logger.Info(b.String())
	}

	repos := make([]globalregistry.Repository, 0)
	for _, pData := range projectData.Repositories {
		pName := strings.Split(pData, "/")[0]
		if pName == proj.GetName() {
			repo := &projectRepository{
				name: pData,
				proj: proj,
			}
			repos = append(repos, repo)
		}
	}
	return repos, err
}

//func (p *projectAPI) delete(name string) error {
//	return fmt.Errorf("not implemented")
//}

//func (p *projectAPI) listProjectRepositories(proj *project) ([]globalregistry.Repository, error) {
//	return nil, fmt.Errorf("not implemented")
//}

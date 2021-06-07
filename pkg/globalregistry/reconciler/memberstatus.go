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
package reconciler

import (
	"bytes"
	"context"
	"fmt"

	"os"

	"encoding/base64"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"path/filepath"
)

type MemberStatus struct {
	Name string
	Type string
	Role string
}

func (ms *MemberStatus) GetName() string {
	return ms.Name
}

func (ms *MemberStatus) GetType() string {
	return ms.Type
}

func (ms *MemberStatus) GetRole() string {
	return ms.Role
}

type memberAddAction struct {
	MemberStatus
	projectName string
}

var _ Action = &memberAddAction{}

func (ma *memberAddAction) String() string {
	return fmt.Sprintf("adding member %s to %s",
		ma.Name, ma.projectName)
}

type persistMemberCredentials struct {
	globalregistry.ProjectMemberCredentials
	action   *memberAddAction
	registry globalregistry.Registry
}

var _ SideEffect = &persistMemberCredentials{}

func (pmc *persistMemberCredentials) Perform(ctx context.Context) error {
	path := ctx.Value(SideEffectPath)
	if path == nil {
		return fmt.Errorf("Context shall contain SideEffectPath")
	}
	serializer := ctx.Value(SideEffectSerializer)
	if path == nil {
		return fmt.Errorf("Context shall contain SideEffectSerializer")
	}
	ser := serializer.(*json.Serializer)
	buf := bytes.NewBuffer(nil)
	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	_, err := fmt.Fprintf(encoder, "%s:%s",
		pmc.Username, pmc.Password,
	)
	if err != nil {
		return err
	}
	dockerConfigJson := fmt.Sprintf("{\"auths\": {\"%s\": {\"auth\": \"%s\"}}}",
		pmc.registry.GetAPIEndpoint(), buf.String(),
	)
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Immutable: nil,
		StringData: map[string]string{
			".dockerconfigjson": dockerConfigJson,
		},
		Type: "kubernetes.io/dockerconfigjson",
	}
	secret.SetName(pmc.action.Name)
	secret.SetAnnotations(map[string]string{
		"globalregistry.org/project-name":  pmc.action.projectName,
		"globalregistry.org/registry-name": pmc.registry.GetName(),
	})

	filename := filepath.Join(path.(string), fmt.Sprintf("%s_%s_%s_creds.yaml",
		pmc.registry.GetName(),
		pmc.action.projectName,
		pmc.action.Name,
	))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = ser.Encode(secret, f)
	if err != nil {
		return err
	}
	return nil
}

func (ma *memberAddAction) Perform(reg globalregistry.Registry) (SideEffect, error) {
	project, err := reg.ProjectAPI().GetByName(ma.projectName)
	if err != nil {
		return nilEffect, err
	}
	creds, err := project.AssignMember(ma)
	if err != nil {
		return nilEffect, err
	}
	if creds != nil {
		return &persistMemberCredentials{
			ProjectMemberCredentials: *creds,
			action:                   ma,
			registry:                 reg,
		}, nil
	}
	return nilEffect, nil
}

type removeMemberCredentials struct {
	action   *memberRemoveAction
	registry globalregistry.Registry
}

var _ SideEffect = &removeMemberCredentials{}

func (rmc *removeMemberCredentials) Perform(ctx context.Context) error {
	path := ctx.Value(SideEffectPath)
	if path == nil {
		return fmt.Errorf("Context shall contain SideEffectPath")
	}

	filename := filepath.Join(path.(string), fmt.Sprintf("%s_%s_%s_creds.yaml",
		rmc.registry.GetName(),
		rmc.action.projectName,
		rmc.action.Name,
	))
	return os.Remove(filename)
}

type memberRemoveAction struct {
	MemberStatus
	projectName string
}

var _ Action = &memberRemoveAction{}

func (ma *memberRemoveAction) String() string {
	return fmt.Sprintf("removing member %s from %s",
		ma.Name, ma.projectName)
}

func (ma *memberRemoveAction) Perform(reg globalregistry.Registry) (SideEffect, error) {
	project, err := reg.ProjectAPI().GetByName(ma.projectName)
	if err != nil {
		return nilEffect, err
	}
	err = project.UnassignMember(ma)
	if err != nil {
		return nilEffect, err
	}
	if ma.Type == "Robot" {
		return &removeMemberCredentials{
			action:   ma,
			registry: reg,
		}, nil
	}
	return nilEffect, nil
}

func CompareMemberStatuses(projectName string, actual, expected []MemberStatus) []Action {
	actualDiff := []MemberStatus{}
	expectedDiff := []MemberStatus{}
ActLoop:
	for _, act := range actual {
		for _, exp := range expected {
			if act == exp {
				continue ActLoop
			}
		}
		// act was not found among expected members
		actualDiff = append(actualDiff, act)
	}
ExpLoop:
	for _, exp := range expected {
		for _, act := range actual {
			if act == exp {
				continue ExpLoop
			}
		}
		// exp was not found among actual members
		expectedDiff = append(expectedDiff, exp)
	}
	actions := make([]Action, 0)

	// actualDiff contains the members which are there but are not needed
	for _, act := range actualDiff {
		actions = append(actions, &memberRemoveAction{
			act,
			projectName,
		})
	}

	// expectedClone contains the members which are missing and thus they
	// shall be created
	for _, exp := range expectedDiff {
		actions = append(actions, &memberAddAction{
			exp,
			projectName,
		})
	}

	return actions
}

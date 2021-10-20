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

package registry

import (
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type calculatedReplication int

const (
	noReplication calculatedReplication = iota
	pullReplication
	pushReplication
)

type registryCapabilities struct {
	isGlobal bool
	globalregistry.ReplicationCapabilities
}

func calculateReplicationRule(local, remote registryCapabilities) calculatedReplication {
	if local.isGlobal && remote.isGlobal {
		panic("both local and remote are global")
	}
	if !local.isGlobal && !remote.isGlobal {
		return noReplication
	}
	if remote.isGlobal && remote.CanPush() {
		return noReplication
	}
	if local.isGlobal && local.CanPush() {
		return pushReplication
	}
	if !local.isGlobal && local.CanPull() {
		return pullReplication
	}
	return noReplication
}

type replicationRule struct {
	calculatedReplication
	project               *project
	remote                *Registry
	replicationAnnotation string
}

var _ globalregistry.ReplicationRule = &replicationRule{}

func (rule *replicationRule) GetProjectName() string {
	return rule.project.GetName()
}

func (rule *replicationRule) GetName() string {
	panic("not implemented")
}

var fallbackPullTrigger = "cron */10 * * * *"

func (rule *replicationRule) Trigger() string {
	triggerWords := strings.SplitN(rule.project.Spec.Trigger, " ", 2)
	triggerWord := ""
	if len(triggerWords) > 0 {
		triggerWord = triggerWords[0]
	}

	switch rule.calculatedReplication {
	case noReplication:
		panic("noReplication not handled")

		// In case of push replication we always configure event-based
		// replication triger
	case pushReplication:
		switch triggerWord {
		case "cron", "manual":
			return rule.project.Spec.Trigger
		default:
			return "event_based"
		}

		// In case of pull replication we always configure manual
		// replication triger
	case pullReplication:
		switch triggerWord {
		case "cron", "manual":
			return rule.project.Spec.Trigger
		default:
			return fallbackPullTrigger
		}
	default:
		panic("unhandled case")
	}

}

func (rule *replicationRule) Direction() string {
	switch rule.calculatedReplication {
	case noReplication:
		panic("noReplication not handled")
	case pushReplication:
		return "Push"
	case pullReplication:
		return "Pull"
	default:
		panic("unhandled case")
	}

}

func (rule *replicationRule) RemoteRegistry() globalregistry.Registry {
	return rule.remote
}

func (rule *replicationRule) Type() globalregistry.ReplicationType {
	triggerWords := strings.SplitN(rule.Trigger(), " ", 2)
	triggerWord := ""
	if len(triggerWords) > 0 {
		triggerWord = triggerWords[0]
	}

	switch triggerWord {
	case "manual", "event_based":
		return globalregistry.RegistryReplication
	case "cron":
		switch rule.replicationAnnotation {
		case "registry":
			return globalregistry.RegistryReplication
		default:
			return globalregistry.SkopeoReplication
		}
	default:
		return globalregistry.SkopeoReplication
	}
}

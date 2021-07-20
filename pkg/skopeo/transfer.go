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

package skopeo

import (
	"fmt"
	"os/exec"

	"github.com/go-logr/logr"
)

const (
	skopeCommandName           = "skopeo"
	syncCommand                = "sync"
	sourceFlag                 = "--src"
	destinationFlag            = "--dest"
	dockerTransport            = "docker"
	directoryTransport         = "dir"
	scopedFlag                 = "--scoped"
	souceCredentialsFlag       = "--src-creds"
	destinationCredentialsFlag = "--dest-creds"
)

type transfer struct {
	username string
	password string
}

// New creates a new transfer struct.
func New(username, password string) *transfer {
	return &transfer{
		username: username,
		password: password,
	}
}

// Export exports Docker repositories from a source repository to a destination path.
func (t *transfer) Export(source, destination string, logger logr.Logger) ([]byte, error) {
	logger.Info("exporting images started")

	skopeoCommand := exec.Command(
		skopeCommandName,
		syncCommand,
		sourceFlag,
		dockerTransport,
		destinationFlag,
		directoryTransport,
		scopedFlag,
		souceCredentialsFlag,
		fmt.Sprintf("%s:%s", t.username, t.password),
		source,
		destination,
	)

	return skopeoCommand.CombinedOutput()
}

// Import imports Docker repositories from a source path to a destination repository.
// func (t *transfer) Import(source, destination string, logger logr.Logger) error {
// 	logger.Info("importing images started")

// 	err := syncImages(&transferData{
// 		sourcePath:           source,
// 		destinationPath:      destination,
// 		sourceCtx:            t.dirCtx,
// 		destinationCtx:       t.dockerCtx,
// 		sourceTransport:      directory.Transport.Name(),
// 		destinationTransport: docker.Transport.Name(),
// 		scoped:               false,
// 	})

// 	if err != nil {
// 		return fmt.Errorf("syncing images failed: %w", err)
// 	}

// 	return nil
// }

// Sync synchronizes Docker repositories from a source repository to a destination repository.
// func (t *transfer) Sync(sourceRepo, destinationRepo string, logger logr.Logger) error {
// 	logger.Info("syncing images started")
// 	err := syncImages(&transferData{
// 		sourcePath:           sourceRepo,
// 		destinationPath:      destinationRepo,
// 		sourceCtx:            t.dockerCtx,
// 		destinationCtx:       t.dockerCtx,
// 		sourceTransport:      docker.Transport.Name(),
// 		destinationTransport: docker.Transport.Name(),
// 		scoped:               false,
// 	})

// 	if err != nil {
// 		return fmt.Errorf("syncing images failed: %w", err)
// 	}

// 	return nil
// }

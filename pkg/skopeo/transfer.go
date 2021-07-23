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
	"os"
	"os/exec"

	"github.com/go-logr/logr"

	_ "embed"
)

const (
	commandPath = "registryman-skopeo"

	syncCommand                = "sync"
	sourceTransportFlag        = "--src"
	destinationTransportFlag   = "--dest"
	dockerTransport            = "docker"
	directoryTransport         = "dir"
	scopedFlag                 = "--scoped"
	sourceCredentialsFlag      = "--src-creds"
	destinationCredentialsFlag = "--dest-creds"
)

//go:embed skopeo
var skopeoBinary []byte

type transfer struct {
	username string
	password string
}

// TODO: create job with kubernetes lib + common interface
// New creates a new transfer struct.
func New(username, password string) (*transfer, error) {
	err := os.WriteFile(commandPath, skopeoBinary, 0711)
	if err != nil {
		return nil, err
	}
	return &transfer{
		username: username,
		password: password,
	}, nil
}

// Export exports Docker repositories from a source repository to a destination path.
func (t *transfer) Export(source, destination string, logger logr.Logger) error {
	logger.Info("exporting images started")

	skopeoCommand := exec.Command(
		fmt.Sprintf("./%s", commandPath),
		syncCommand,
		sourceTransportFlag,
		dockerTransport,
		destinationTransportFlag,
		directoryTransport,
		scopedFlag,
		sourceCredentialsFlag,
		fmt.Sprintf("%s:%s", t.username, t.password),
		source,
		destination,
	)

	// TODO: remove this in prod!
	logger.Info(skopeoCommand.String())

	skopeoCommand.Stderr = os.Stderr
	skopeoCommand.Stdout = os.Stdout

	return skopeoCommand.Run()
}

// Import imports Docker repositories from a source path to a destination repository.
func (t *transfer) Import(source, destination string, logger logr.Logger) error {
	logger.Info("importing images started")

	skopeoCommand := exec.Command(
		fmt.Sprintf("./%s", commandPath),
		syncCommand,
		sourceTransportFlag,
		directoryTransport,
		destinationTransportFlag,
		dockerTransport,
		destinationCredentialsFlag,
		fmt.Sprintf("%s:%s", t.username, t.password),
		source,
		destination,
	)

	// TODO: remove this in prod!
	logger.Info(skopeoCommand.String())

	skopeoCommand.Stderr = os.Stderr
	skopeoCommand.Stdout = os.Stdout

	return skopeoCommand.Run()
}

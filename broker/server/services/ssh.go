/*
 * Copyright 2018, CS Systemes d'Information, http://www.c-s.fr
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package services

import (
	"fmt"
	"strings"
	"time"

	//"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"

	"github.com/CS-SI/SafeScale/providers"
	"github.com/CS-SI/SafeScale/providers/metadata"
	"github.com/CS-SI/SafeScale/providers/model"
	"github.com/CS-SI/SafeScale/providers/model/enums/HostExtension"
	"github.com/CS-SI/SafeScale/utils/retry"
	"github.com/CS-SI/SafeScale/utils/retry/Verdict"

	"github.com/CS-SI/SafeScale/system"
)

const protocolSeparator = ":"

//go:generate mockgen -destination=../mocks/mock_sshapi.go -package=mocks github.com/CS-SI/SafeScale/broker/server/services SSHAPI

// SSHAPI defines ssh management API
type SSHAPI interface {
	// Connect(name string) error
	Run(hostname, cmd string) (int, string, string, error)
	Copy(from string, to string) (int, string, string, error)
	GetConfig(interface{}) (*system.SSHConfig, error)
}

// SSHService SSH service
type SSHService struct {
	provider    *providers.Service
	hostService HostAPI
}

// NewSSHService ...
func NewSSHService(api *providers.Service) *SSHService {
	return newSSHService(api, nil)
}

// newSSHService creates a SSH service
func newSSHService(api *providers.Service, hostService HostAPI) *SSHService {
	return &SSHService{
		provider: api,
	}
}

// GetConfig creates SSHConfig to connect to an host
func (svc *SSHService) GetConfig(hostParam interface{}) (*system.SSHConfig, error) {
	var host *model.Host

	switch hostParam.(type) {
	case string:
		mh, err := metadata.LoadHost(svc.provider, hostParam.(string))
		if err != nil {
			return nil, err
		}
		host = mh.Get()
	case *model.Host:
		host = hostParam.(*model.Host)
	default:
		panic("param must be a string or a *model.Host!")
	}

	sshConfig := system.SSHConfig{
		PrivateKey: host.PrivateKey,
		Port:       22,
		Host:       host.GetAccessIP(),
		User:       model.DefaultUser,
	}
	heNetworkV1 := model.HostExtensionNetworkV1{}
	err := host.Extensions.Get(HostExtension.NetworkV1, &heNetworkV1)
	if err != nil {
		return nil, err
	}
	if heNetworkV1.DefaultGatewayID != "" {
		mgw, err := metadata.LoadHost(svc.provider, heNetworkV1.DefaultGatewayID)
		if err != nil {
			return nil, err
		}
		gw := mgw.Get()
		GatewayConfig := system.SSHConfig{
			PrivateKey: gw.PrivateKey,
			Port:       22,
			Host:       gw.GetAccessIP(),
			User:       model.DefaultUser,
		}
		sshConfig.GatewayConfig = &GatewayConfig
	}
	return &sshConfig, nil
}

// WaitServerReady waits for remote SSH server to be ready. After timeout, fails
func (svc *SSHService) WaitServerReady(hostParam interface{}, timeout time.Duration) error {
	var err error
	sshSvc := NewSSHService(svc.provider)
	ssh, err := sshSvc.GetConfig(hostParam)
	if err != nil {
		return fmt.Errorf("failed to read SSH config: %s", err.Error())
	}
	return ssh.WaitServerReady(timeout)
}

// Run tries to execute command 'cmd' on the host
func (svc *SSHService) Run(hostName, cmd string) (int, string, string, error) {
	var stdOut, stdErr string
	var retCode int
	var err error

	host, err := svc.hostService.Get(hostName)
	if err != nil {
		return 0, "", "", fmt.Errorf("no host found with name or id '%s'", hostName)
	}

	// retrieve ssh config to perform some commands
	ssh, err := svc.GetConfig(host)

	if err != nil {
		return 0, "", "", err
	}

	err = retry.WhileUnsuccessfulDelay1SecondWithNotify(
		func() error {
			retCode, stdOut, stdErr, err = svc.run(ssh, cmd)
			return err
		},
		2*time.Minute,
		func(t retry.Try, v Verdict.Enum) {
			if v == Verdict.Retry {
				log.Printf("Remote SSH service on host '%s' isn't ready, retrying...\n", hostName)
			}
		},
	)

	return retCode, stdOut, stdErr, err
}

// run executes command on the host
func (svc *SSHService) run(ssh *system.SSHConfig, cmd string) (int, string, string, error) {
	// Create the command
	sshCmd, err := ssh.Command(cmd)
	if err != nil {
		return 0, "", "", err
	}
	return sshCmd.Run()
}

func extracthostName(in string) (string, error) {
	parts := strings.Split(in, protocolSeparator)
	if len(parts) == 1 {
		return "", nil
	}
	if len(parts) > 2 {
		return "", fmt.Errorf("too many parts in path")
	}
	hostName := strings.TrimSpace(parts[0])
	for _, protocol := range []string{"file", "http", "https", "ftp"} {
		if strings.ToLower(hostName) == protocol {
			return "", fmt.Errorf("no protocol expected. Only host name")
		}
	}

	return hostName, nil
}

func extractPath(in string) (string, error) {
	parts := strings.Split(in, protocolSeparator)
	if len(parts) == 1 {
		return in, nil
	}
	if len(parts) > 2 {
		return "", fmt.Errorf("too many parts in path")
	}
	_, err := extracthostName(in)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(parts[1]), nil
}

// Copy copy file/directory
func (svc *SSHService) Copy(from, to string) (int, string, string, error) {
	hostName := ""
	var upload bool
	var localPath, remotePath string
	// Try extract host
	hostFrom, err := extracthostName(from)
	if err != nil {
		return 0, "", "", err
	}
	hostTo, err := extracthostName(to)
	if err != nil {
		return 0, "", "", err
	}

	// Host checks
	if hostFrom != "" && hostTo != "" {
		return 0, "", "", fmt.Errorf("copy between 2 hosts is not supported yet")
	}
	if hostFrom == "" && hostTo == "" {
		return 0, "", "", fmt.Errorf("no host name specified neither in from nor to")
	}

	fromPath, err := extractPath(from)
	if err != nil {
		return 0, "", "", err
	}
	toPath, err := extractPath(to)
	if err != nil {
		return 0, "", "", err
	}

	if hostFrom != "" {
		hostName = hostFrom
		remotePath = fromPath
		localPath = toPath
		upload = false
	} else {
		hostName = hostTo
		remotePath = toPath
		localPath = fromPath
		upload = true
	}

	host, err := svc.hostService.Get(hostName)
	if err != nil {
		return 0, "", "", fmt.Errorf("no host found with name or id '%s'", hostName)
	}

	// retrieve ssh config to perform some commands
	ssh, err := svc.GetConfig(host.ID)
	if err != nil {
		return 0, "", "", err
	}

	return ssh.Copy(remotePath, localPath, upload)
}

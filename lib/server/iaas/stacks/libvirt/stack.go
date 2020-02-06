//+build libvirt

package local

import (
	"fmt"
	"time"

	"github.com/libvirt/libvirt-go"

	"github.com/CS-SI/SafeScale/lib/server/resources/abstracts"
	"github.com/CS-SI/SafeScale/lib/utils/scerr"
	"github.com/CS-SI/SafeScale/lib/server/iaas/stacks"
)

type Stack struct {
	LibvirtService *libvirt.Connect
	LibvirtConfig  *stacks.LocalConfiguration
	Config         *stacks.ConfigurationOptions
	AuthOptions    *stacks.AuthenticationOptions
}

func (s Stack) WaitHostReady(hostParam interface{}, timeout time.Duration) (*abstracts.Host, error) {
	return nil, scerr.NotImplementedError("WaitHostReady not implemented yet!") // FIXME Technical debt
}

// Build Create and initialize a ClientAPI
func New(auth stacks.AuthenticationOptions, localCfg stacks.LocalConfiguration, cfg stacks.ConfigurationOptions) (*Stack, error) {
	stack := &Stack{
		Config:        &cfg,
		LibvirtConfig: &localCfg,
		AuthOptions:   &auth,
	}

	libvirtConnection, err := libvirt.NewConnect(stack.LibvirtConfig.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt : %s", err.Error())
	}
	stack.LibvirtService = libvirtConnection

	if stack.LibvirtConfig.LibvirtStorage != "" {
		err := stack.CreatePoolIfUnexistant(stack.LibvirtConfig.LibvirtStorage)
		if err != nil {
			return nil, fmt.Errorf("unable to create StoragePool : %s", err.Error())
		}
	}

	return stack, nil
}

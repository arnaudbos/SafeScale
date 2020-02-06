//+build !libvirt

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

package local

import (
	"fmt"
	"time"

	"github.com/CS-SI/SafeScale/lib/server/iaas"
	"github.com/CS-SI/SafeScale/lib/server/iaas/providers"
	providerapi "github.com/CS-SI/SafeScale/lib/server/iaas/providers/api"
	"github.com/CS-SI/SafeScale/lib/server/iaas/userdata"
	"github.com/CS-SI/SafeScale/lib/server/resources/abstracts"
	"github.com/CS-SI/SafeScale/lib/server/resources/enums/hoststate"
)

var errorStr = "libvirt Driver is not enabled, use the libvirt option while compiling (make libvirt all)"

// provider is the implementation of the local driver regarding to the api.Provider
type provider struct {
	tenantParameters map[string]interface{}
}

//AuthOptions fields are the union of those recognized by each identity implementation and provider.
type AuthOptions struct {
}

// CfgOptions configuration options
type CfgOptions struct {
}

func (provider *provider) WaitHostReady(hostParam interface{}, timeout time.Duration) (*abstracts.Host, error) {
	return nil, fmt.Errorf(errorStr)
}

func (provider *provider) Build(params map[string]interface{}) (providerapi.Provider, error) {
	provider.tenantParameters = map[string]interface{}{}
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetAuthenticationOptions() (providers.Config, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetConfigurationOptions() (providers.Config, error) {
	return nil, fmt.Errorf(errorStr)
}

func (provider *provider) ListAvailabilityZones() (map[string]bool, error) {
	return nil, fmt.Errorf(errorStr)
}

// ListRegions returns a list with the regions available
func (provider *provider) ListRegions() ([]string, error) {
	return nil, fmt.Errorf(errorStr)
}

func (provider *provider) ListImages(all bool) ([]abstracts.Image, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetImage(id string) (*abstracts.Image, error) {
	return nil, fmt.Errorf(errorStr)
}

func (provider *provider) GetTemplate(id string) (*abstracts.HostTemplate, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) ListTemplates(all bool) ([]abstracts.HostTemplate, error) {
	return nil, fmt.Errorf(errorStr)
}

func (provider *provider) CreateKeyPair(name string) (*abstracts.KeyPair, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetKeyPair(id string) (*abstracts.KeyPair, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) ListKeyPairs() ([]abstracts.KeyPair, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) DeleteKeyPair(id string) error {
	return fmt.Errorf(errorStr)
}

func (provider *provider) CreateNetwork(req abstracts.NetworkRequest) (*abstracts.Network, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetNetwork(id string) (*abstracts.Network, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetNetworkByName(name string) (*abstracts.Network, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) ListNetworks() ([]*abstracts.Network, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) DeleteNetwork(id string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) CreateGateway(req abstracts.GatewayRequest) (*abstracts.Host, *userdata.Content, error) {
	return nil, nil, fmt.Errorf(errorStr)
}
func (provider *provider) DeleteGateway(string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) CreateVIP(networkID string, description string) (*abstracts.VIP, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) AddPublicIPToVIP(vip *abstracts.VIP) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) BindHostToVIP(vip *abstracts.VIP, hostID string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) UnbindHostFromVIP(vip *abstracts.VIP, hostID string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) DeleteVIP(vip *abstracts.VIP) error {
	return fmt.Errorf(errorStr)
}

func (provider *provider) CreateHost(request abstracts.HostRequest) (*abstracts.Host, *userdata.Content, error) {
	return nil, nil, fmt.Errorf(errorStr)
}
func (provider *provider) ResizeHost(id string, request abstracts.SizingRequirements) (*abstracts.Host, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) InspectHost(interface{}) (*abstracts.Host, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetHostByName(string) (*abstracts.Host, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetHostState(interface{}) (hoststate.Enum, error) {
	return hoststate.ERROR, fmt.Errorf(errorStr)
}
func (provider *provider) ListHosts() ([]*abstracts.Host, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) DeleteHost(id string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) StartHost(id string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) StopHost(id string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) RebootHost(id string) error {
	return fmt.Errorf(errorStr)
}

func (provider *provider) CreateVolume(request abstracts.VolumeRequest) (*abstracts.Volume, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) GetVolume(id string) (*abstracts.Volume, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) ListVolumes() ([]abstracts.Volume, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) DeleteVolume(id string) error {
	return fmt.Errorf(errorStr)
}

func (provider *provider) CreateVolumeAttachment(request abstracts.VolumeAttachmentRequest) (string, error) {
	return "", fmt.Errorf(errorStr)
}
func (provider *provider) GetVolumeAttachment(serverID, id string) (*abstracts.VolumeAttachment, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) ListVolumeAttachments(serverID string) ([]abstracts.VolumeAttachment, error) {
	return nil, fmt.Errorf(errorStr)
}
func (provider *provider) DeleteVolumeAttachment(serverID, id string) error {
	return fmt.Errorf(errorStr)
}
func (provider *provider) GetName() string {
	return "local_disabled"
}
func (provider *provider) GetTenantParameters() map[string]interface{} {
	return nil
}

// GetCapabilities returns the capabilities of the provider
func (provider *provider) GetCapabilities() providers.Capabilities {
	return providers.Capabilities{}
}

func init() {
	// log.Debug("Registering fake local provider")
	iaas.Register("local", &provider{})
}

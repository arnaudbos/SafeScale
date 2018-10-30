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

package openstack

import (
	"github.com/CS-SI/SafeScale/iaas/provider"
	"github.com/CS-SI/SafeScale/iaas/resource/enums/VolumeSpeed"
)

// AuthenticationOptions fields are the union of those recognized by each identity implementation and provider
type AuthenticationOptions struct {
	// IdentityEndpoint specifies the HTTP endpoint that is required to work with
	// the Identity API of the appropriate version. While it's ultimately needed by
	// all of the identity services, it will often be populated by a provider-level
	// function.
	IdentityEndpoint string

	// Username is required if using Identity V2 API. Consult with your provider's
	// control panel to discover your account's username. In Identity V3, either
	// UserID or a combination of Username and DomainID or DomainName are needed.
	Username, UserID string

	// Exactly one of Password or APIKey is required for the Identity V2 and V3
	// APIs. Consult with your provider's control panel to discover your account's
	// preferred method of authentication.
	Password, APIKey string

	// At most one of DomainID and DomainName must be provided if using Username
	// with Identity V3. Otherwise, either are optional.
	DomainID, DomainName string

	// The TenantID and TenantName fields are optional for the Identity V2 API.
	// Some providers allow you to specify a TenantName instead of the TenantId.
	// Some require both. Your provider's authentication policies will determine
	// how these fields influence authentication.
	TenantID, TenantName string

	// AllowReauth should be set to true if you grant permission for Gophercloud to
	// cache your credentials in memory, and to allow Gophercloud to attempt to
	// re-authenticate automatically if/when your token expires.  If you set it to
	// false, it will not cache these settings, but re-authentication will not be
	// possible. This setting defaults to false.
	//
	// NOTE: The reauth function will try to re-authenticate endlessly if left unchecked.
	// The way to limit the number of attempts is to provide a custom HTTP client to the provider client
	// and provide a transport that implements the RoundTripper interface and stores the number of failed retries.
	// For an example of this, see here: https://github.com/gophercloud/rack/blob/1.0.0/auth/clients.go#L311
	AllowReauth bool

	// TokenID allows users to authenticate (possibly as another user) with an
	// authentication token ID.
	TokenID string

	//Openstack region (data center) where the infrstructure will be created
	Region string

	//FloatingIPPool name of the floating IP pool
	//Necessary only if UseFloatingIP is true
	FloatingIPPool string
}

// GetAuthOpts returns the auth options
func (s *Stack) GetAuthOpts() (providers.Config, error) {
	cfg := providers.ConfigMap{}

	cfg.Set("TenantName", s.AuthOpts.TenantName)
	cfg.Set("Login", s.AuthOpts.Username)
	cfg.Set("Password", s.AuthOpts.Password)
	cfg.Set("AuthUrl", s.AuthOpts.IdentityEndpoint)
	cfg.Set("Region", s.AuthOpts.Region)
	return cfg, nil
}

// ConfigurationOptions are the stack configuration options
type ConfigurationOptions struct {
	// Name of the provider (external) network
	ProviderNetwork string

	// DNSList list of DNS
	DNSList []string

	// UseFloatingIP indicates if floating IP are used (optional)
	UseFloatingIP bool

	// UseLayer3Networking indicates if layer 3 networking features (router) can be used
	// if UseFloatingIP is true UseLayer3Networking must be true
	UseLayer3Networking bool

	// AutoHostNetworkInterfaces indicates if network interfaces are configured automatically by the provider or needs a post configuration
	AutoHostNetworkInterfaces bool

	// VolumeSpeeds map volume types with volume speeds
	VolumeSpeeds map[string]VolumeSpeed.Enum

	// S3Protocol protocol used to mount object storage (ex: swiftks or s3)
	S3Protocol string

	// MetadataBucketName contains the name of the bucket storing metadata
	MetadataBucketName string
}

// GetCfgOpts return configuration parameters
func (s *Stack) GetCfgOpts() (providers.Config, error) {
	cfg := providers.ConfigMap{}

	cfg.Set("DNSList", s.CfgOpts.DNSList)
	cfg.Set("S3Protocol", s.CfgOpts.S3Protocol)
	cfg.Set("AutoHostNetworkInterfaces", s.CfgOpts.AutoHostNetworkInterfaces)
	cfg.Set("UseLayer3Networking", s.CfgOpts.UseLayer3Networking)
	cfg.Set("MetadataBucket", s.CfgOpts.MetadataBucketName)

	return cfg, nil
}

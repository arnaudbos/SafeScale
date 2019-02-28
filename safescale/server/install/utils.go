/*
 * Copyright 2018-2019, CS Systemes d'Information, http://www.c-s.fr
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

package install

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/viper"

	pb "github.com/CS-SI/SafeScale/safescale"
	"github.com/CS-SI/SafeScale/safescale/client"
	"github.com/CS-SI/SafeScale/system"
	"github.com/CS-SI/SafeScale/utils/retry"
)

const (
	featureScriptTemplateContent = `
rm -f /var/tmp/feature.{{.reserved_Name}}.{{.reserved_Action}}_{{.reserved_Step}}.log
exec 1<&-
exec 2<&-
exec 1<>/var/tmp/feature.{{.reserved_Name}}.{{.reserved_Action}}_{{.reserved_Step}}.log
exec 2>&1

{{ .reserved_BashLibrary }}

{{ .reserved_Content }}
`
)

var (
	featureScriptTemplate *template.Template
)

// parseTargets validates targets on the cluster from the feature specification
// Without error, returns 'master target', 'private node target' and 'public node target'
func parseTargets(specs *viper.Viper) (string, string, string, error) {
	if !specs.IsSet("feature.target.cluster") {
		return "", "", "", fmt.Errorf("feature isn't suitable for a cluster")
	}

	master := strings.ToLower(strings.TrimSpace(specs.GetString("feature.target.cluster.master")))
	switch master {
	case "":
		fallthrough
	case "false":
		fallthrough
	case "no":
		fallthrough
	case "none":
		fallthrough
	case "0":
		master = "0"
	case "any":
		fallthrough
	case "one":
		fallthrough
	case "1":
		master = "1"
	case "all":
		fallthrough
	case "*":
		master = "*"
	default:
		return "", "", "", fmt.Errorf("invalid value '%s' for field 'feature.target.cluster.master'", master)
	}

	privnode := strings.ToLower(strings.TrimSpace(specs.GetString("feature.target.cluster.node.private")))
	switch privnode {
	case "false":
		fallthrough
	case "no":
		fallthrough
	case "none":
		privnode = "0"
	case "any":
		fallthrough
	case "one":
		fallthrough
	case "1":
		privnode = "1"
	case "":
		fallthrough
	case "all":
		fallthrough
	case "*":
		privnode = "*"
	default:
		return "", "", "", fmt.Errorf("invalid value '%s' for field 'feature.target.cluster.node.private'", privnode)
	}

	pubnode := strings.ToLower(strings.TrimSpace(specs.GetString("feature.target.cluster.node.public")))
	switch pubnode {
	case "":
		fallthrough
	case "false":
		fallthrough
	case "no":
		fallthrough
	case "none":
		fallthrough
	case "0":
		pubnode = "0"
	case "any":
		fallthrough
	case "one":
		fallthrough
	case "1":
		pubnode = "1"
	case "all":
		fallthrough
	case "*":
		pubnode = "*"
	default:
		return "", "", "", fmt.Errorf("invalid value '%s' for field 'feature.target.cluster.node.public'", pubnode)
	}

	if master == "0" && privnode == "0" && pubnode == "0" {
		return "", "", "", fmt.Errorf("invalid 'feature.target.cluster': no target designated")
	}
	return master, privnode, pubnode, nil
}

// UploadStringToRemoteFile creates a file 'filename' on remote 'host' with the content 'content'
func UploadStringToRemoteFile(content string, host *pb.Host, filename string, owner, group, rights string) error {
	if content == "" {
		panic("content is empty!")
	}
	if host == nil {
		panic("host is nil!")
	}
	if filename == "" {
		panic("filename is empty!")
	}
	f, err := system.CreateTempFileFromString(content, 0600)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %s", err.Error())
	}
	to := fmt.Sprintf("%s:%s", host.Name, filename)
	safescale := client.New().Ssh
	retryErr := retry.WhileUnsuccessful(
		func() error {
			var retcode int
			retcode, _, _, err = safescale.Copy(f.Name(), to, 15*time.Second, client.DefaultExecutionTimeout)
			if err != nil {
				return err
			}
			if retcode != 0 {
				// If retcode == 1 (general copy error), retry. It may be a temporary network incident
				if retcode == 1 {
					// File may exist on target, try to remote it
					_, _, _, err = safescale.Run(host.Name, fmt.Sprintf("sudo rm -f %s", filename), 15*time.Second, client.DefaultExecutionTimeout)
					return fmt.Errorf("file may exist on remote with inappropriate access rights, deleted it and retrying")
				}
				if system.IsSCPRetryable(retcode) {
					err = fmt.Errorf("failed to copy temporary file to '%s' (retcode: %d=%s)", to, retcode, system.SCPErrorString(retcode))
				}
				return nil
			}
			return nil
		},
		1*time.Second,
		2*time.Minute,
	)
	_ = os.Remove(f.Name())
	if retryErr != nil {
		switch retryErr.(type) {
		case retry.ErrTimeout:
			return fmt.Errorf("timeout trying to copy temporary file to '%s': %s", to, retryErr.Error())
		}
		return err
	}

	cmd := ""
	if owner != "" {
		cmd += "sudo chown " + owner + " " + filename
	}
	if group != "" {
		if cmd != "" {
			cmd += " && "
		}
		cmd += "sudo chgrp " + group + " " + filename
	}
	if rights != "" {
		if cmd != "" {
			cmd += " && "
		}
		cmd += "sudo chmod " + rights + " " + filename
	}
	retryErr = retry.WhileUnsuccessful(
		func() error {
			var retcode int
			retcode, _, _, err = safescale.Run(host.Name, cmd, 15*time.Second, client.DefaultExecutionTimeout)
			if err != nil {
				return err
			}
			if retcode != 0 {
				err = fmt.Errorf("failed to change rights of file '%s' (retcode=%d)", to, retcode)
				return nil
			}
			return nil
		},
		2*time.Second,
		1*time.Minute,
	)
	if retryErr != nil {
		switch retryErr.(type) {
		case retry.ErrTimeout:
			return fmt.Errorf("timeout trying to change rights of file '%s' on host '%s': %s", filename, host.Name, err.Error())
		default:
			return fmt.Errorf("failed to change rights of file '%s' on host '%s': %s", filename, host.Name, retryErr.Error())
		}
	}
	return nil
}

// normalizeScript envelops the script with log redirection to /var/tmp/feature.<name>.<action>.log
// and ensures BashLibrary are there
func normalizeScript(params map[string]interface{}) (string, error) {
	var err error

	if featureScriptTemplate == nil {
		// parse then execute the template
		featureScriptTemplate, err = template.New("normalize_script").Parse(featureScriptTemplateContent)
		if err != nil {
			return "", fmt.Errorf("error parsing bash template: %s", err.Error())
		}
	}

	// Configures BashLibrary template var
	bashLibrary, err := system.GetBashLibrary()
	if err != nil {
		return "", err
	}
	params["reserved_BashLibrary"] = bashLibrary

	dataBuffer := bytes.NewBufferString("")
	err = featureScriptTemplate.Execute(dataBuffer, params)
	if err != nil {
		return "", err
	}
	return dataBuffer.String(), nil
}

func replaceVariablesInString(text string, v Variables) (string, error) {
	tmpl, err := template.New("text").Parse(text)
	if err != nil {
		return "", fmt.Errorf("failed to parse: %s", err.Error())
	}
	dataBuffer := bytes.NewBufferString("")
	err = tmpl.Execute(dataBuffer, v)
	if err != nil {
		return "", fmt.Errorf("failed to replace variables: %s", err.Error())
	}
	return dataBuffer.String(), nil
}

func findConcernedHosts(list []string, c *Feature) (string, error) {
	// No metadata yet for features, first host is designated concerned host
	if len(list) > 0 {
		return list[0], nil
	}
	return "", fmt.Errorf("no hosts")
	//for _, h := range list {
	//}
}

// determineContext ...
func determineContext(t Target) (hT *HostTarget, cT *ClusterTarget, nT *NodeTarget) {
	hT = nil
	cT = nil
	nT = nil

	var ok bool

	hT, ok = t.(*HostTarget)
	if !ok {
		cT, ok = t.(*ClusterTarget)
		if !ok {
			nT, _ = t.(*NodeTarget)
		}
	}
	return
}

// Check if required parameters defined in specification file have been set in 'v'
func checkParameters(f *Feature, v Variables) error {
	if f.specs.IsSet("feature.parameters") {
		params := f.specs.GetStringSlice("feature.parameters")
		for _, k := range params {
			if _, ok := v[k]; !ok {
				return fmt.Errorf("missing value for parameter '%s'", k)
			}
		}
	}
	return nil
}

// setImplicitParameters configures parameters that are implicitely defined, based on target
func setImplicitParameters(t Target, v Variables) {
	hT, cT, nT := determineContext(t)
	if cT != nil {
		cluster := cT.cluster
		identity := cluster.GetIdentity()
		v["ClusterName"] = identity.Name
		v["ClusterComplexity"] = strings.ToLower(identity.Complexity.String())
		v["ClusterFlavor"] = strings.ToLower(identity.Flavor.String())
		networkCfg := cluster.GetNetworkConfig()
		v["GatewayIP"] = networkCfg.GatewayIP
		v["PublicIP"] = networkCfg.PublicIP
		v["MasterIDs"] = cluster.ListMasterIDs()
		v["MasterIPs"] = cluster.ListMasterIPs()
		if _, ok := v["Username"]; !ok {
			v["Username"] = "cladm"
			v["Password"] = identity.AdminPassword
		}
		if _, ok := v["CIDR"]; !ok {
			v["CIDR"] = networkCfg.CIDR
		}
	} else {
		var host *pb.Host
		if nT != nil {
			host = nT.HostTarget.host
		}
		if hT != nil {
			host = hT.host
		}
		v["Hostname"] = host.Name
		v["HostIP"] = host.PrivateIP
		gw := gatewayFromHost(host)
		if gw != nil {
			v["GatewayIP"] = gw.PrivateIP
			v["PublicIP"] = gw.PublicIP
		} else {
			v["PublicIP"] = host.PublicIP
		}
		if _, ok := v["Username"]; !ok {
			v["Username"] = "gpac"
		}
	}
}

func gatewayFromHost(host *pb.Host) *pb.Host {
	gwID := host.GetGatewayID()
	// If host has no gateway, host is gateway
	if gwID == "" {
		return host
	}
	gw, err := client.New().Host.Inspect(gwID, client.DefaultExecutionTimeout)
	if err != nil {
		return nil
	}
	return gw
}
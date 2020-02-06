/*
 * Copyright 2018-2020, CS Systemes d'Information, http://www.c-s.fr
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

package feature

import (
	"io/ioutil"
	"strings"

	mapset "github.com/deckarep/golang-set"

	"github.com/CS-SI/SafeScale/lib/server/iaas"
	"github.com/CS-SI/SafeScale/lib/server/resources/operations/features"
	srvutils "github.com/CS-SI/SafeScale/lib/server/utils"
	"github.com/CS-SI/SafeScale/lib/utils"
	"github.com/CS-SI/SafeScale/lib/utils/concurrency"
)

// List lists all features suitable for hosts
func List() ([]interface{}, error) {
	cfgFiles := mapset.NewSet()

	captured := mapset.NewSet()

	var paths []string
	paths = append(paths, utils.AbsPathify("$HOME/.safescale/features"))
	paths = append(paths, utils.AbsPathify("$HOME/.config/safescale/features"))
	paths = append(paths, utils.AbsPathify("/etc/safescale/features"))

	for _, path := range paths {
		files, err := ioutil.ReadDir(path)
		if err == nil {
			for _, f := range files {
				if isCfgFile := strings.HasSuffix(strings.ToLower(f.Name()), ".yml"); isCfgFile == true {
					cfgFiles.Add(strings.Replace(strings.ToLower(f.Name()), ".yml", "", 1))
				}
			}
		}
	}
	for _, feat := range features.GetAllEmbeddedMap() {
		yamlKey := "feature.suitableFor.host"

		if !captured.Contains(feat.GetName()) {
			ok := false
			if feat.GetSpecs().IsSet(yamlKey) {
				value := strings.ToLower(feat.GetSpecs().GetString(yamlKey))
				ok = value == "ok" || value == "yes" || value == "true" || value == "1"
			}
			if ok {
				cfgFiles.Add(feat.GetFilename())
			}

			captured.Add(feat.GetName())
		}
	}

	return cfgFiles.ToSlice(), nil
}

// New searches for a spec file name 'name' and initializes a new Feature object
// with its content
func New(task concurrency.Task, svc iaas.Service, name string) (Feature, error) {
	if name == "" {
		return nil, utils.InvalidParameterError("name", "can't be empty string!")
	}
	assumeEmbed := false
	if task == nil {
		assumeEmbed = true
	}
	if svc == nil {
		assumeEmbed = true
	}

	if assumeEmbed {
		// Failed to find a spec file on filesystem, trying with embedded ones
		feat, err = features.NewEmbedded(name)
		if err != nil {
			return utils.NotFoundError(err.Error()), nil
		}
	} else {
		feat, err := features.New(task, svc, name)
		if err != nil {
			if _, ok := err.(utils.ErrNotFound); !ok {
				return nil, srvutils.ThrowErr(err)
			}

			// Failed to find a spec file on filesystem, trying with embedded ones
			feat, err = features.NewEmbedded(name)
			if err != nil {
				return nil, utils.NotFoundError(err.Error())
			}
		}
	}
	return feat, nil
}

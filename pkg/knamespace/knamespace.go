// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at

//   http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package knamespace

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type NamespaceConfig struct {
	Name        string            `yaml:"name"`
	Mode        string            `yaml:"mode"`
	Annotations map[string]string `yaml:"annotations"`
	Labels      map[string]string `yaml:"labels"`
}

type NamespacesConfig struct {
	DefaultConfig NamespaceConfig   `yaml:"defaultNamespaceSettings"`
	Namespaces    []NamespaceConfig `yaml:"namespaces"`
}

// Gets the config for a specific Knamespace
func (n NamespacesConfig) GetConfig(namespaceName string) (*NamespaceConfig, error) {
	for _, namespaceConfig := range n.Namespaces {
		if namespaceConfig.Name == namespaceName {

			// Apply Defaults to the retrieved namespaceConfig
			if namespaceConfig.Annotations == nil {
				namespaceConfig.Annotations = n.DefaultConfig.Annotations
			}

			if namespaceConfig.Labels == nil {
				namespaceConfig.Labels = n.DefaultConfig.Labels
			}

			if namespaceConfig.Mode == "" {
				namespaceConfig.Mode = n.DefaultConfig.Mode
			}

			return &namespaceConfig, nil
		}
	}
	return nil, fmt.Errorf(fmt.Sprintf("namespace %s not found in configuration", namespaceName))
}

// Return Knamespacer Defaults
func (n NamespacesConfig) GetDefault() (*NamespaceConfig, error) {
	return &n.DefaultConfig, nil
}

// Return Namespaces config from Namespaces config file
func GetNamespacesConfig(namespacesConfigFileName string) (*NamespacesConfig, error) {

	contents, err := readConfigFile(namespacesConfigFileName)
	if err != nil {
		return nil, err
	}

	data, err := parseConfigFileContents(contents)
	if err != nil {
		return nil, err
	}

	log.Debugf("Defaults: %#v", data.DefaultConfig)
	log.Debugf("Namespaces: %#v", data.Namespaces)

	return data, nil
}

// Read config file, return contents
func readConfigFile(namespacesConfigFileName string) ([]byte, error) {
	log.Debugf("Reading Config File %s:", namespacesConfigFileName)
	contents, err := os.ReadFile(namespacesConfigFileName)
	if err != nil {
		log.Errorf("Error reading Knamespacer conig file %s : %s", namespacesConfigFileName, err)
		return nil, err
	}
	log.Debugf(string(contents))
	return contents, err
}

// Unmarshal contents of config file into NamespacesesConfig
func parseConfigFileContents(contents []byte) (*NamespacesConfig, error) {
	data := &NamespacesConfig{}
	err := yaml.UnmarshalStrict(contents, data)
	if err != nil {
		log.Errorf("Error parsing Knamespacer Configuration: %s", err)
		return nil, err
	}
	return data, nil
}

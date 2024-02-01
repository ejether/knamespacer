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
			return &namespaceConfig, nil
		}
	}
	return nil, fmt.Errorf(fmt.Sprintf("namespace %s not found in configuration", namespaceName))
}

// Get Knamespacer Defaults
func (n NamespacesConfig) GetDefault() (*NamespaceConfig, error) {
	return &n.DefaultConfig, nil
}

// Read and return namespaces configuration. Exit if we can't.
func GetNamespacesConfig(namespacesConfigFileName string) *NamespacesConfig {
	data := &NamespacesConfig{}
	log.Infof("Reading Config File %s:", namespacesConfigFileName)
	contents, err := os.ReadFile(namespacesConfigFileName)
	if err != nil {
		log.Fatalf("Error reading Knamespacer Configuration File %s: %s", namespacesConfigFileName, err)
	}
	log.Debugf(string(contents))

	err = yaml.UnmarshalStrict(contents, data)
	if err != nil {
		log.Fatalf("Error parsing Configuration: %s", err)
	}
	log.Debugf("Defaults: %#v", data.DefaultConfig)
	log.Debugf("Namespaces: %#v", data.Namespaces)
	return data
}

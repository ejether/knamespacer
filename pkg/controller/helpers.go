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

package controller

import (
	"github.com/ejether/knamespacer/pkg/knamespace"
	"github.com/ejether/knamespacer/pkg/kube"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// Process cluster namespace and modify metadata if specified
func processNamespace(k8s *kube.K8sClient, namespaceName string, namespacesConfig *knamespace.NamespacesConfig) error {
	log.Info("Processing Cluster Namespace: ", namespaceName)

	namespaceConfig, err := namespacesConfig.GetConfig(namespaceName)
	if err != nil {
		log.Infof("No Knamespacer config specified for %s. Skipping.", namespaceName)
		return nil
	}

	namespace, err := k8s.GetClusterNamespace(namespaceName)
	if err != nil {
		log.Infof("Unable to fetch cluster namespace '%s' for modification: %s", namespaceName, err)
		return err
	}

	ModifyNamespaceMetadata(namespace, namespaceConfig)

	log.Debugf("Updated Namespace Meta: %#v", namespace.ObjectMeta)

	err = k8s.UpdateNamespace(namespace)
	if err != nil {
		log.Errorf("Failed to update namespace %s: %s", namespace, err)
		return nil
	}

	return nil
}

// Updates Annotation and Label Metadata on the specified namespace according to the NamespaceConfig
func ModifyNamespaceMetadata(namespace *corev1.Namespace, namespaceConfig *knamespace.NamespaceConfig) {
	// ModifyNamespaceMetadata(namespace, namespaceConfig)
	log.Infof("Updating Namespace %s in %s mode", namespace.Name, namespaceConfig.Mode)
	log.Debugf("Initial Namespace Meta: %#v", namespace.ObjectMeta)

	switch namespaceConfig.Mode {
	case "sync":
		log.Debug("Syncing...")
		namespace.Annotations = syncNamespaceMeta(namespace.Annotations, namespaceConfig.Annotations)
		namespace.Labels = syncNamespaceMeta(namespace.Labels, namespaceConfig.Labels)
	case "upsert":
		log.Debug("Upserting...")
		namespace.Annotations = upsertNamespaceMeta(namespace.Annotations, namespaceConfig.Annotations)
		namespace.Labels = upsertNamespaceMeta(namespace.Labels, namespaceConfig.Labels)
	case "insert":
		log.Debug("Inserting...")
		namespace.Annotations = insertNamespaceMeta(namespace.Annotations, namespaceConfig.Annotations)
		namespace.Labels = insertNamespaceMeta(namespace.Labels, namespaceConfig.Labels)
	}
	log.Debugf("New Namespace Meta: %#v", namespace.ObjectMeta)
}

// Used to sync Annotations or Labels on a Namespace. Sync wholesale replaces the meta type so this just returns the new config
// metaObject passed in so all 'mode' functions have the same signature.
func syncNamespaceMeta(_ map[string]string, config map[string]string) map[string]string {
	return config
}

// Use to upsert Annotation or Labels on a Namespace. Upsert replaces any keys that are present with new values and adds new key:values.
// Any key:values present on the namespace meta object but not in the config are ignored
func upsertNamespaceMeta(metaObject map[string]string, config map[string]string) map[string]string {
	if metaObject == nil {
		metaObject = make(map[string]string)
	}

	for key, value := range config {
		metaObject[key] = value
	}
	return metaObject
}

// Used to insert Annotations or labels on a Namespace. Insert _only_ adds new key:values and ignores any that are already present even
// if they are specified in the config
func insertNamespaceMeta(metaObject map[string]string, config map[string]string) map[string]string {
	for key, value := range config {

		if metaObject == nil {
			metaObject = make(map[string]string)
		}

		if metaObject[key] == "" {
			metaObject[key] = value
		}
	}
	return metaObject
}

// Determine which Knamespaces don't exist in the cluster and create them
func createMissingNamespaces(k8s *kube.K8sClient, namespacesConfig *knamespace.NamespacesConfig) error {
	nsList, err := k8s.ListClusterNameSpaces()
	if err != nil {
		return err
	}
	var namespacesToCreate []string
	for _, configNamespace := range namespacesConfig.Namespaces {
		createNamespace := true
		for _, ns := range nsList.Items {
			if ns.Name == configNamespace.Name {
				createNamespace = false
			}
		}
		if createNamespace {
			namespacesToCreate = append(namespacesToCreate, configNamespace.Name)
		}
	}
	log.Infof("Creating configured name spaces that do not exist in cluster: %s", namespacesToCreate)
	return k8s.CreateNamespaces(namespacesToCreate)
}

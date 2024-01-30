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
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	toolsWatch "k8s.io/client-go/tools/watch"

	"knamespacer/pkg/knamespace"
	"knamespacer/pkg/kube"
)

func Controller(namespacesConfigFile string) {

	namespacesConfig := knamespace.GetNamespacesConfig(namespacesConfigFile)

	if err := createMissingNamespaces(namespacesConfig); err != nil {
		log.Errorf("Unable to create some namespaces configure, but not already present: %s", err)
	}
	var wg sync.WaitGroup
	go WatchNamespaces(namespacesConfig)
	wg.Add(1)
	wg.Wait()
}

// Watch for changes to cluster namespaces
func WatchNamespaces(namespacesConfig *knamespace.NamespacesConfig) {
	clientset := kube.GetClientSet()

	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		timeOut := int64(60)
		return clientset.CoreV1().Namespaces().Watch(context.Background(), metav1.ListOptions{TimeoutSeconds: &timeOut})
	}

	watcher, _ := toolsWatch.NewRetryWatcher("1", &cache.ListWatch{WatchFunc: watchFunc})

	for event := range watcher.ResultChan() {
		item := event.Object.(*corev1.Namespace)
		namespaceName := item.GetName()
		log.Infof("Caught event %s on namespace: %s. Processing...", event.Type, namespaceName)

		err := processNamespace(namespaceName, namespacesConfig)
		if err != nil {
			log.Errorf("Encounter %s while processing %s. Skipping....", err, namespaceName)
		}
	}
}

// Process cluster namespace and modify metadata if specified
func processNamespace(namespaceName string, namespacesConfig *knamespace.NamespacesConfig) error {
	log.Info("Processing Cluster Namespace : ", namespaceName)

	namespaceConfig, err := namespacesConfig.GetConfig(namespaceName)

	if err != nil {
		log.Infof("No Knamespacer config specified for %s. Skipping.", namespaceName)
		return nil
	}

	namespace, err := kube.GetClusterNamespace(namespaceName)
	if err != nil {
		log.Infof("Unable to fetch cluster namespace '%s' for modification: %s", namespaceName, err)
		return err
	}

	newNamespaceAnnotations := namespaceConfig.Annotations //, err := //generateAnnotations(namespace, namespaceConfig)
	newNamespaceLabels := namespaceConfig.Annotations      //, err := generateLabels(namespace, namespaceConfig)

	err = kube.ModifyNamespaceMetadata(namespace, newNamespaceAnnotations, newNamespaceLabels)
	if err != nil {
		log.Errorf("Failed to update namespace %s: %s", namespace, err)
		return nil
	}

	return nil
}

// Determine which Knamespaces don't exist in the cluster and create them
func createMissingNamespaces(namespacesConfig *knamespace.NamespacesConfig) error {
	nsList := kube.ListClusterNameSpaces()
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
	log.Infof("Creating configured name spaces that do not exits in cluster: %s", namespacesToCreate)
	return kube.CreateNamespaces(namespacesToCreate)

}

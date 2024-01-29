// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package kube

import (
	"context"
	"errors"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Get k8s client.
func GetClientSet() *kubernetes.Clientset {

	log.Debug("Get kubernetes config.")
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Debug("In-cluster config not found. Using local config.")
		// Not in cluster? Let's try locally
		kubehome := filepath.Join(homedir.HomeDir(), ".kube", "config")
		cfg, err = clientcmd.BuildConfigFromFlags("", kubehome)
		if err != nil {
			log.Fatalf("Error loading local kubernetes configuration: %s", err)
		}
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error creating kubernetes client: %s", err)
	}

	return clientset
}

// List namespaces currently in cluster. Exit if we can't.
func ListClusterNameSpaces() *corev1.NamespaceList {
	clientset := GetClientSet()
	nsList, err := clientset.CoreV1().
		Namespaces().
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing Cluster Namespaces: %s", err)
	}

	return nsList
}

// Creates a namespace if it does not exist
func CreateNamespaces(namespaceNames []string) error {
	didGetError := false
	for _, nsName := range namespaceNames {
		if err := createNamespace(nsName); err != nil {
			log.Errorf("Unable to create namespace %s: %s", nsName, err)
			didGetError = true
		}
	}
	if didGetError {
		return errors.New("Failed to create some namespaces")
	}

	return nil

}

// Creates a Namespace
func createNamespace(namespaceName string) error {
	clientset := GetClientSet()
	nsName := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	_, err := clientset.CoreV1().Namespaces().Create(context.Background(), nsName, metav1.CreateOptions{})
	return err
}

// Modify the Metadata of the specified Namespace
func ModifyNamespaceMetadata(namespace *corev1.Namespace, annotations map[string]string, labels map[string]string) error {

	namespace.ObjectMeta.Annotations = annotations
	namespace.ObjectMeta.Labels = labels

	clientset := GetClientSet()
	_, err := clientset.CoreV1().Namespaces().Update(context.TODO(), namespace, metav1.UpdateOptions{})
	if err != nil {
		log.Errorf("Could not update namespace %s: %s", namespace.Name, err)
		return err
	}
	return nil
}

// Retrieve corev1.Namespace from cluster
func GetClusterNamespace(namespaceName string) (*corev1.Namespace, error) {
	clientset := GetClientSet()
	namespace, err := clientset.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
	return namespace, err
}

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

package utils

import (
	"os"
	"testing"

	"github.com/ejether/knamespacer/pkg/kube"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func SetupTestEnvironment() (*kube.K8sClient, func(), error) {
	env := &envtest.Environment{}
	cfg, err := env.Start()
	if err != nil {
		return nil, nil, err
	}
	testClient := &kube.K8sClient{
		K8s: kube.GetClient(cfg),
	}

	// linter is mad add the ignored error when defer stopFn()
	return testClient, func() {
		_ = env.Stop()
	}, nil
}

func SetupTestEnvironmentWithKubeconfig() (string, func(), error) {
	env := &envtest.Environment{}
	_, err := env.Start()
	if err != nil {
		return "", nil, err
	}

	user, err := env.ControlPlane.AddUser(envtest.User{
		Name:   "envtest-admin",
		Groups: []string{"system:masters"},
	}, nil)
	if err != nil {
		log.Error(err, "unable to provision admin user, continuing on without it")
		return "", nil, err
	}

	kubeconfigFile, err := os.CreateTemp("", "scratch-env-kubeconfig-")
	if err != nil {
		log.Error(err, "unable to create kubeconfig file, continuing on without it")
		return "", nil, err
	}

	kubeConfig, err := user.KubeConfig()
	if err != nil {
		log.Error(err, "unable to create kubeconfig")
	}

	if _, err := kubeconfigFile.Write(kubeConfig); err != nil {
		log.Error(err, "unable to save kubeconfig")
		return "", nil, err
	}

	log.Info("Wrote kubeconfig")

	// linter is mad add the ignored error when defer stopFn()
	return kubeconfigFile.Name(), func() {
		_ = env.Stop()
	}, nil
}

func InArray(ns corev1.Namespace, arr []corev1.Namespace) bool {
	for _, v := range arr {
		if ns.Name == v.Name {
			return true
		}
	}
	return false
}

func TestNamespaces() []corev1.Namespace {
	testNamespaces := []corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "alpha",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "beta",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gamma",
			},
		},
	}

	return testNamespaces
}

func CheckExpectedNamespaces(t *testing.T, expected []corev1.Namespace, actual corev1.NamespaceList) {
	comp := func(a map[string]string, b map[string]string) bool {
		for ak, av := range a {
			if b[ak] == "" || b[ak] != av {
				return false
			}
		}
		return true
	}

	for _, expectedNamespace := range expected {
		var actualNs corev1.Namespace
		found := false

		for _, ns := range actual.Items {
			if expectedNamespace.Name == ns.Name {
				actualNs = ns
				found = true
			}
		}

		if !found {
			t.Errorf("Expected Namespace Not Discovered In Cluster (%v)\n", expectedNamespace.Name)
			return
		}

		if !comp(expectedNamespace.Annotations, actualNs.Annotations) {
			t.Errorf("Expected Annotations Not Matched (%v / %v) \n", expectedNamespace.Annotations, actualNs.Annotations)
		}

		if !comp(expectedNamespace.Labels, actualNs.Labels) {
			t.Errorf("Expected Labels Not Matched (%v / %v) \n", expectedNamespace.Labels, actualNs.Labels)
		}

	}
}

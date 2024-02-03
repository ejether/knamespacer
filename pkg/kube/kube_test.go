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

package kube

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestEnvTest(t *testing.T) {
	loggerOpts := &logzap.Options{
		Development: true, // a sane default
		ZapOpts:     []zap.Option{zap.AddCaller()},
	}

	ctrl.SetLogger(logzap.New(logzap.UseFlagOptions(loggerOpts)))
	log := ctrl.Log.WithName("TestEnvTest")

	log.Info("Starting...")
	env := &envtest.Environment{}
	os.Setenv("KUBEBUILDER_ATTACH_CONTROL_PLANE_OUTPUT", "true")
	log.Info("Starting apiserver & etcd")

	cfg, err := env.Start()
	if err != nil {
		log.Error(err, "unable to start the test environment")
		// shut down the environment in case we started it and failed while
		// installing CRDs or provisioning users.
		if err := env.Stop(); err != nil {
			log.Error(err, "unable to stop the test environment after an error (this might be expected, but just though you should know)")
		}
	}

	log.Info("apiserver running", "host", cfg.Host)

	// NB(directxman12): this group is unfortunately named, but various
	// kubernetes versions require us to use it to get "admin" access.
	user, err := env.ControlPlane.AddUser(envtest.User{
		Name:   "envtest-admin",
		Groups: []string{"system:masters"},
	}, nil)
	if err != nil {
		log.Error(err, "unable to provision admin user, continuing on without it")
	}

	kubeconfigFile, err := os.CreateTemp("", "scratch-env-kubeconfig-")
	if err != nil {
		log.Error(err, "unable to create kubeconfig file, continuing on without it")
	}
	defer os.Remove(kubeconfigFile.Name())

	{
		log := log.WithValues("path", kubeconfigFile.Name())
		log.V(1).Info("Writing kubeconfig")

		kubeConfig, err := user.KubeConfig()
		if err != nil {
			log.Error(err, "unable to create kubeconfig")
		}

		if _, err := kubeconfigFile.Write(kubeConfig); err != nil {
			log.Error(err, "unable to save kubeconfig")
		}

		log.Info("Wrote kubeconfig")
	}

	if opts := env.WebhookInstallOptions; opts.LocalServingPort != 0 {
		log.Info("webhooks configured for", "host", opts.LocalServingHost, "port", opts.LocalServingPort, "dir", opts.LocalServingCertDir)
	}

	ctx := ctrl.SetupSignalHandler()
	<-ctx.Done()

	log.Info("Shutting down apiserver & etcd")
	err = env.Stop()
	if err != nil {
		log.Error(err, "unable to stop the test environment")
	}

	log.Info("Shutdown successful")

}

func TestListClusterNameSpaces(t *testing.T) {
	fakeK8s, err := getFakeK8s()
	assert.Nil(t, err)

	fakeClient := K8sClient{
		k8s: fakeK8s,
	}

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

	for _, v := range testNamespaces {
		err := fakeClient.k8s.Create(context.TODO(), &v)
		assert.Nil(t, err)
	}

	nss, err := fakeClient.ListClusterNameSpaces()
	assert.Nil(t, err)

	if len(nss.Items) != len(testNamespaces) {
		assert.Nil(t, errors.New("returned namespaces do not match length of test namespaces"))
	}

	for _, v := range testNamespaces {
		if !inArray(v, nss.Items) {
			assert.Nil(t, errors.New("returned namespaces do not match test namespaces"))
		}
		// spew.Dump(v)
	}
	// spew.Dump(nss)

}

func getFakeK8s(initObjs ...client.Object) (client.WithWatch, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	// ...
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(initObjs...).Build(), nil
}

func inArray(ns corev1.Namespace, arr []corev1.Namespace) bool {
	for _, v := range arr {
		if ns.Name == v.Name {
			return true
		}
	}
	return false
}

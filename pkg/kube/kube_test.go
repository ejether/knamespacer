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

package kube_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ejether/knamespacer/pkg/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	logzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	testNamespaces = []corev1.Namespace{
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
)

func init() {
	loggerOpts := &logzap.Options{
		Development: true, // a sane default
		ZapOpts:     []zap.Option{zap.AddCaller()},
	}

	ctrl.SetLogger(logzap.New(logzap.UseFlagOptions(loggerOpts)))
}

func TestListClusterNameSpaces(t *testing.T) {
	testClient, stopFn, err := utils.SetupTestEnvironment()
	defer stopFn()
	assert.Nil(t, err)

	for _, v := range testNamespaces {
		err := testClient.K8s.Create(context.TODO(), &v)
		assert.Nil(t, err)
	}

	nss, err := testClient.ListClusterNameSpaces()
	assert.Nil(t, err)

	for _, v := range testNamespaces {
		if !utils.InArray(v, nss.Items) {
			assert.Nil(t, errors.New("returned namespaces do not match test namespaces"))
		}
	}
}

func TestCreateNamespaces(t *testing.T) {
	testClient, stopFn, err := utils.SetupTestEnvironment()
	defer stopFn()
	assert.Nil(t, err)

	names := []string{}
	for _, v := range testNamespaces {
		names = append(names, v.Name)
	}

	err = testClient.CreateNamespaces(names)
	assert.Nil(t, err)

	nss := corev1.NamespaceList{}
	err = testClient.K8s.List(context.Background(), &nss)
	assert.Nil(t, err)

	for _, v := range testNamespaces {
		if !utils.InArray(v, nss.Items) {
			assert.Nil(t, errors.New("returned namespaces do not match test namespaces"))
		}
	}
}

func TestCreateNamespace(t *testing.T) {
	testClient, stopFn, err := utils.SetupTestEnvironment()
	defer stopFn()
	assert.Nil(t, err)

	for _, v := range testNamespaces {
		err = testClient.CreateNamespace(v.Name)
		assert.Nil(t, err)
	}

	nss := corev1.NamespaceList{}
	err = testClient.K8s.List(context.Background(), &nss)
	assert.Nil(t, err)

	for _, v := range testNamespaces {
		if !utils.InArray(v, nss.Items) {
			assert.Nil(t, errors.New("returned namespaces do not match test namespaces"))
		}
	}
}

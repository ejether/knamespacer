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
	"fmt"

	"github.com/ejether/knamespacer/pkg/knamespace"
	"github.com/ejether/knamespacer/pkg/kube"

	log "github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KnamespacerController struct {
	client.Client
	Scheme          *runtime.Scheme
	NamespaceConfig *knamespace.NamespacesConfig
	StartUp         bool
}

func (r *KnamespacerController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	namespaceName := req.Name
	log.Infof("Reconiling: %v", namespaceName)

	k8s := &kube.K8sClient{
		K8s: r.Client,
	}

	if r.StartUp {
		err := createMissingNamespaces(k8s, r.NamespaceConfig)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("encounter %w while creating %s. Skipping", err, namespaceName)
		}
		r.StartUp = false
	}

	err := processNamespace(k8s, namespaceName, r.NamespaceConfig)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("encounter %w while processing %s. Skipping", err, namespaceName)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KnamespacerController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}

package controller

import (
	"context"
	"fmt"

	"github.com/ejether/knamespacer/pkg/knamespace"
	"github.com/ejether/knamespacer/pkg/kube"

	"github.com/davecgh/go-spew/spew"
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
	log.Infof("Reconiling: %v", spew.Sdump(namespaceName, req))

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

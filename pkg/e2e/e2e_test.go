package e2e

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	logzap "sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/davecgh/go-spew/spew"
	"github.com/ejether/knamespacer/cmd"
	"github.com/ejether/knamespacer/pkg/kube"
	"github.com/ejether/knamespacer/pkg/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	expectedNamespaces = []corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "one",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "two",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "three",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "four",
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

func TestEndToEnd(t *testing.T) {
	log := ctrl.Log.WithName("TestEndToEnd")
	log.Info("Starting TestEndToEnd Test Function")

	kubeconfig, stopFn, err := utils.SetupTestEnvironmentWithKubeconfig()
	defer stopFn()
	assert.Nil(t, err)
	defer os.Remove(kubeconfig)

	log.Info(kubeconfig)
	log.Info(spew.Sdump(os.ReadFile(kubeconfig)))

	err = os.Setenv("KUBECONFIG", kubeconfig)
	assert.Nil(t, err)

	k8s := kube.NewK8sClient()

	cmd.RootCmd.SetArgs([]string{fmt.Sprintf("--config=%s", "./config.yaml")})
	go func() {
		cmd.Execute()
	}()

	time.Sleep(1 * time.Second)

	nss, err := k8s.ListClusterNameSpaces()
	assert.Nil(t, err)

	for _, v := range nss.Items {
		log.Info(v.Name)
	}

	for _, v := range expectedNamespaces {
		if !utils.InArray(v, nss.Items) {
			assert.Nil(t, errors.New("returned namespaces do not match test namespaces"))
		}
	}

}

package e2e

import (
	"fmt"
	"os"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	logzap "sigs.k8s.io/controller-runtime/pkg/log/zap"

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
				Annotations: map[string]string{
					"foo": "one",
				},
				Labels: map[string]string{
					"bar": "one",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "two",
				Annotations: map[string]string{
					"foo": "two",
				},
				Labels: map[string]string{
					"bar": "two",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "three",
				Annotations: map[string]string{
					"this": "should",
					"foo":  "three",
				},
				Labels: map[string]string{
					"add": "new",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "four",
				Annotations: map[string]string{
					"default": "annotation",
				},
				Labels: map[string]string{
					"default": "label",
				},
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

	assert.NotEmpty(t, kubeconfig)
	kubeconfigBytes, err := os.ReadFile(kubeconfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, kubeconfigBytes)

	err = os.Setenv("KUBECONFIG", kubeconfig)
	assert.Nil(t, err)

	k8s := kube.NewK8sClient()

	cmd.RootCmd.SetArgs([]string{fmt.Sprintf("--config=%s", "./config.yaml")})
	go cmd.Execute()

	time.Sleep(1 * time.Second)

	nss, err := k8s.ListClusterNameSpaces()
	assert.Nil(t, err)

	utils.CheckExpectedNamespaces(t, expectedNamespaces, *nss)
}

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

package cmd

import (
	"os"

	"github.com/ejether/knamespacer/pkg/controller"
	"github.com/ejether/knamespacer/pkg/knamespace"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	log "github.com/sirupsen/logrus"

	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	ctrl "sigs.k8s.io/controller-runtime"
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "knamespacer",
	Short: "Controller for your kubernetes namespaces",
	Long:  `Controller for your kubernetes namespaces`,
	Run:   Run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var configFile string
var debug bool

func init() {
	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Yaml file with Namespaces to configure")
	err := RootCmd.MarkPersistentFlagRequired("config")
	if err != nil {
		log.Fatal(err)
	}

	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode")
}

func Run(cmd *cobra.Command, args []string) {
	log.SetOutput(os.Stdout)
	// Set debug logging.
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		log.Error(err, "unable to add client-go scheme")
		os.Exit(1)
	}

	// Starting a manager, which handles the connection to the API as well as caching
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:   ":8080",
			SecureServing: false,
		},
		WebhookServer:          webhook.NewServer(webhook.Options{}),
		HealthProbeBindAddress: ":8081",
		LeaderElection:         false,
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Retrieve the config file once
	var nspcCfg *knamespace.NamespacesConfig
	if nspcCfg, err = knamespace.GetNamespacesConfig(configFile); err != nil {
		log.Error(err, "unable to retrieve")
		os.Exit(1)
	}

	// Register the controller
	if err = (&controller.KnamespacerController{
		StartUp:         true,
		NamespaceConfig: nspcCfg,

		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "ImageBuild")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	// Manager Starts the Controller
	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}

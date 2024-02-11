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
	"sync"

	"github.com/ejether/knamespacer/pkg/controller"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "knamespacer",
	Short: "Controller for your kubernetes namespaces",
	Long:  `Controller for your kubernetes namespaces`,
	Run: func(cmd *cobra.Command, args []string) {

		log.SetOutput(os.Stdout)
		// Set debug logging.
		if debug {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}

		var wg sync.WaitGroup
		go controller.Controller(configFile)
		wg.Add(1)
		wg.Wait()
	},
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

/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"sync"

	"knamespacer/pkg/controller"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
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
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var configFile string
var debug bool

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Yaml file with Namespaces to configure")
	rootCmd.MarkPersistentFlagRequired("config")

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode")

}

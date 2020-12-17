package cmd

/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fantai/ftab/pkg/httpfile"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile, testFile string
var outputFormat string
var conns, requests int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ftab",
	Short: "A http(s) benchmark tool with variable replacement",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {

		fp, err := os.Open(testFile)
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer fp.Close()

		file, err := httpfile.ParseReader(fp)
		if err != nil {
			return fmt.Errorf("parse file: %w", err)
		}
		defer file.Release()

		if requests > 1 {
			r := httpfile.ReportStat(httpfile.Bench(file, conns, requests))
			r.Currency = conns

			switch outputFormat {
			case "plain":
				httpfile.PlainOutput(&r, os.Stdout)
			case "json":
				text, err := json.MarshalIndent(&r, "", "  ")
				if err != nil {
					return fmt.Errorf("marshall json: %w", err)
				}
				fmt.Println(string(text))
			default:
				httpfile.HumanOutput(&r, os.Stdout)
			}
		} else {
			traceInfo := httpfile.Execute(file)
			fmt.Println(traceInfo)
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	log, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(log)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ftab.yaml)")

	rootCmd.Flags().StringVarP(&outputFormat, "output", "m", "human", "result output format[plain, human, json]")
	rootCmd.Flags().StringVarP(&testFile, "in", "i", "test.http", "the http file to bench")
	rootCmd.Flags().IntVarP(&conns, "connections", "c", 1, "connection in this bench ")
	rootCmd.Flags().IntVarP(&requests, "requests", "n", 1, "total requests in this bench ")

	viper.BindPFlags(rootCmd.Flags())

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".ftab" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".ftab")
	}

	viper.SetEnvPrefix("FTAB")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

}

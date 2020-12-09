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
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/fantai/ftab/internal/echo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var echoAddr string

// echoCmd represents the echo command
var echoCmd = &cobra.Command{
	Use:   "echo",
	Short: "a echo server",
	Long:  `a ECHO Server send what read to client`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("start server at http://%s\n", echoAddr)
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go echo.Start(echoAddr, wg)

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		wg.Done()
	},
}

func init() {
	rootCmd.AddCommand(echoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// echoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// echoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	echoCmd.Flags().StringVarP(&echoAddr, "echo.addr", "a", "127.0.0.1:6601", "the server address")
	echoCmd.Flags().BoolP("echo.verbose", "v", false, "verbose mode")

	viper.BindPFlags(echoCmd.Flags())

}

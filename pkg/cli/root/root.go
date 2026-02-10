/*
 * Copyright 2023 Aurelia Schittler
 *
 * Licensed under the EUPL, Version 1.2 or – as soon they
   will be approved by the European Commission - subsequent
   versions of the EUPL (the "Licence");
 * You may not use this work except in compliance with the
   Licence.
 * You may obtain a copy of the Licence at:
 *
 * https://joinup.ec.europa.eu/software/page/eupl5
 *
 * Unless required by applicable law or agreed to in
   writing, software distributed under the Licence is
   distributed on an "AS IS" basis,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
   express or implied.
 * See the Licence for the specific language governing
   permissions and limitations under the Licence.
*/

package root

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"atfutil/pkg/cli/allocate"
	"atfutil/pkg/cli/binary/version"
	"atfutil/pkg/cli/cidr"
	"atfutil/pkg/cli/list"
	"atfutil/pkg/cli/release"
	"atfutil/pkg/cli/render"
	"atfutil/pkg/cli/validate"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "atfutil",
	Short: "atfutil can validate and render atf (allocation table format) yaml files",
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.atfutil.yaml)")
	rootCmd.PersistentFlags().StringP("file", "f", "-", "input file")
	rootCmd.PersistentFlags().StringP("output", "o", "-", "output file")

	// Bind flags to viper
	viper.BindPFlag("file", rootCmd.PersistentFlags().Lookup("file"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))

	// Add subcommands
	rootCmd.AddCommand(validate.Command())
	rootCmd.AddCommand(render.Command())
	rootCmd.AddCommand(allocate.Command())
	rootCmd.AddCommand(release.Command())
	rootCmd.AddCommand(cidr.Command())
	rootCmd.AddCommand(list.Command())
	rootCmd.AddCommand(version.Command())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".atfutil")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// Command returns the root command
func Command() *cobra.Command {
	return rootCmd
}

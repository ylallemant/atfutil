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

package touch

import (
	"fmt"
	"net"
	"os"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"atfutil/pkg/atf"
)

var touchCmd = &cobra.Command{
	Use:   "touch",
	Short: "check if a file exists and create it if missing",
	Long:  "check if an ATF file exists and create it with an initial superblock if it is missing",
	Run:   runTouch,
}

func init() {
	touchCmd.Flags().StringP("name", "n", "", "name of the superblock")
	touchCmd.Flags().String("cidr", "", "CIDR of the superblock")

	touchCmd.MarkFlagRequired("name")
	touchCmd.MarkFlagRequired("cidr")
}

// Command returns the touch command
func Command() *cobra.Command {
	return touchCmd
}

func runTouch(cmd *cobra.Command, args []string) {
	filename := viper.GetString("file")

	if filename == "" || filename == "-" {
		quitWithError(errors.New("must specify a file path with --file"))
	}

	name, _ := cmd.Flags().GetString("name")
	cidrStr, _ := cmd.Flags().GetString("cidr")

	// Check if file exists
	if _, err := os.Stat(filename); err == nil {
		// File exists, nothing to do
		fmt.Printf("file %s already exists\n", filename)
		os.Exit(0)
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		quitWithError(errors.Wrap(err, "failed to check file"))
	}

	// File does not exist, create it
	_, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		quitWithError(errors.Wrap(err, "invalid CIDR"))
	}

	atfFile := &atf.File{
		Name:        &name,
		Superblock:  &atf.IPNet{IPNet: ipNet},
		Allocations: []*atf.Allocation{},
	}

	outBytes, err := yaml.Marshal(atfFile)
	if err != nil {
		quitWithError(errors.Wrap(err, "failed to marshal ATF file"))
	}

	outFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		quitWithError(errors.Wrap(err, "failed to create file"))
	}
	defer outFile.Close()

	_, err = outFile.Write(outBytes)
	if err != nil {
		quitWithError(errors.Wrap(err, "failed to write file"))
	}

	fmt.Printf("created %s\n", filename)
	os.Exit(0)
}

func quitWithError(err error) {
	fmt.Fprintf(os.Stderr, "err: %s\n", err.Error())
	os.Exit(1)
}

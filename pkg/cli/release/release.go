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

package release

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"atfutil/pkg/atf"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "release a block and remove it from the file",
	Long:  "release a block and remove it from the file. The block is identified by its CIDR or ID.",
	Run:   runRelease,
}

func init() {
	releaseCmd.Flags().StringP("cidr", "c", "", "CIDR of the block to release")
	releaseCmd.Flags().StringP("id", "I", "", "ID of the block to release")
	releaseCmd.Flags().Bool("in-place", false, "modify the input file in place")
}

func Command() *cobra.Command {
	return releaseCmd
}

func runRelease(cmd *cobra.Command, args []string) {
	outputFilename := viper.GetString("output")
	inputFilename := viper.GetString("file")
	inPlace, _ := cmd.Flags().GetBool("in-place")

	if outputFilename != "-" && inPlace {
		quitWithError(errors.New("cannot use --output and --in-place at the same time"))
	}

	releaseCIDR, _ := cmd.Flags().GetString("cidr")
	releaseID, _ := cmd.Flags().GetString("id")

	if releaseCIDR == "" && releaseID == "" {
		quitWithError(errors.New("must specify either --cidr or --id"))
	}

	inFile, err := getInputFile(inputFilename)
	if err != nil {
		quitWithError(err)
	}
	defer inFile.Close()

	atfFile, err := loadAtfFromFile(inFile)
	if err != nil {
		quitWithError(err)
	}

	found := false
	newAllocations := make([]*atf.Allocation, 0, len(atfFile.Allocations))
	for _, alloc := range atfFile.Allocations {
		match := false
		if releaseCIDR != "" && alloc.Network != nil {
			_, cidrNet, err := net.ParseCIDR(releaseCIDR)
			if err != nil {
				quitWithError(errors.Wrap(err, "invalid CIDR format"))
			}
			if alloc.Network.String() == cidrNet.String() {
				match = true
			}
		}
		if releaseID != "" && alloc.Ident == releaseID {
			match = true
		}

		if match {
			found = true
			fmt.Fprintf(os.Stderr, "Released block: %s (%s)\n", alloc.Ident, alloc.Network.String())
		} else {
			newAllocations = append(newAllocations, alloc)
		}
	}

	if !found {
		if releaseCIDR != "" {
			quitWithError(errors.Errorf("no allocation found with CIDR %s", releaseCIDR))
		} else {
			quitWithError(errors.Errorf("no allocation found with ID %s", releaseID))
		}
	}

	atfFile.Allocations = newAllocations

	outBytes, err := yaml.Marshal(atfFile)
	if err != nil {
		quitWithError(err)
	}

	if outputFilename == "-" && inPlace {
		outputFilename = inputFilename
	}

	outFile, err := getOutputFile(outputFilename)
	if err != nil {
		quitWithError(err)
	}
	defer outFile.Close()

	_, err = outFile.Write(outBytes)
	if err != nil {
		quitWithError(err)
	}

	os.Exit(0)
}

func quitWithError(err error) {
	fmt.Fprintf(os.Stderr, "err: %s\n", err.Error())
	os.Exit(1)
}

func getInputFile(inputFilename string) (*os.File, error) {
	var inputFile *os.File
	if inputFilename == "" {
		return nil, errors.New("need an input filename")
	}
	if inputFilename == "-" {
		inputFile = os.Stdin
	} else {
		file, err := os.OpenFile(inputFilename, os.O_RDONLY, 0)
		if err != nil {
			return nil, err
		}
		inputFile = file
	}
	return inputFile, nil
}

func getOutputFile(outputFilename string) (*os.File, error) {
	var outputFile *os.File
	if outputFilename == "" {
		return nil, errors.New("need an output filename")
	}
	if outputFilename == "-" {
		outputFile = os.Stdout
	} else {
		file, err := os.OpenFile(outputFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return nil, err
		}
		outputFile = file
	}
	return outputFile, nil
}

func loadAtfFromFile(inputFile *os.File) (*atf.File, error) {
	data, err := io.ReadAll(inputFile)
	if err != nil {
		return nil, err
	}
	atfFile := new(atf.File)
	err = yaml.Unmarshal(data, atfFile)
	if err != nil {
		return nil, err
	}
	if atfFile.Superblock == nil {
		return nil, errors.New("file missing superblock")
	}
	for i, alloc := range atfFile.Allocations {
		if alloc.Network == nil {
			return nil, errors.Errorf("file missing network in allocation [%d]", i)
		}
	}
	return atfFile, nil
}

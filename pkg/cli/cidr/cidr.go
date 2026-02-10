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

package cidr

import (
	"fmt"
	"io"
	"os"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"atfutil/pkg/atf"
	"atfutil/pkg/netcalc"
	"atfutil/pkg/netpool"
)

var cidrCmd = &cobra.Command{
	Use:   "cidr",
	Short: "show the superblock CIDR or from a given block",
	Long:  "show the superblock CIDR or the CIDR from a given block identified by ID",
	Run:   runCIDR,
}

func init() {
	cidrCmd.Flags().StringP("id", "i", "", "ID of the block to show CIDR for")
	cidrCmd.Flags().BoolP("allocate", "a", false, "if the block does not exist, allocate it")
	cidrCmd.Flags().IntP("size", "s", -1, "size of the network to allocate (required with --allocate)")
}

func Command() *cobra.Command {
	return cidrCmd
}

func runCIDR(cmd *cobra.Command, args []string) {
	inputFilename := viper.GetString("file")
	outputFilename := viper.GetString("output")

	blockID, _ := cmd.Flags().GetString("id")
	allocateIfMissing, _ := cmd.Flags().GetBool("allocate")
	allocSize, _ := cmd.Flags().GetInt("size")

	inFile, err := getInputFile(inputFilename)
	if err != nil {
		quitWithError(err)
	}
	defer inFile.Close()

	atfFile, err := loadAtfFromFile(inFile)
	if err != nil {
		quitWithError(err)
	}

	if blockID == "" {
		fmt.Println(atfFile.Superblock.String())
		os.Exit(0)
	}

	var foundAlloc *atf.Allocation
	for _, alloc := range atfFile.Allocations {
		if alloc.Ident == blockID {
			foundAlloc = alloc
			break
		}
		for _, subAlloc := range alloc.SubAlloc {
			if subAlloc.Ident == blockID {
				foundAlloc = subAlloc
				break
			}
		}
		if foundAlloc != nil {
			break
		}
	}

	if foundAlloc != nil {
		fmt.Println(foundAlloc.Network.String())
		os.Exit(0)
	}

	if !allocateIfMissing {
		quitWithError(errors.Errorf("no allocation found with ID %s", blockID))
	}

	if allocSize == -1 {
		quitWithError(errors.New("must specify --size when using --allocate"))
	}

	pool, err := netpool.FromAtf(atfFile)
	if err != nil {
		quitWithError(err)
	}

	superAllocSize, _ := atfFile.Superblock.Mask.Size()

	if allocSize > netcalc.AWS_MIN_SUBNET_SIZE || allocSize <= superAllocSize {
		quitWithError(errors.Errorf("requested block size is out of range (%d < block < %d)", superAllocSize, netcalc.AWS_MIN_SUBNET_SIZE))
	}

	net, err := pool.Pool.Alloc(allocSize)
	if err != nil {
		quitWithError(err)
	}

	newAlloc := &atf.Allocation{
		Ident:   blockID,
		Network: &atf.IPNet{IPNet: net},
	}

	atfFile.Allocations = append(atfFile.Allocations, newAlloc)

	fmt.Println(net.String())

	outBytes, err := yaml.Marshal(atfFile)
	if err != nil {
		quitWithError(err)
	}

	// Default output to input file
	if outputFilename == "-" {
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

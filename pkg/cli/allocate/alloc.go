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

package allocate

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"atfutil/pkg/atf"
	"atfutil/pkg/netcalc"
	"atfutil/pkg/netpool"
)

var allocCmd = &cobra.Command{
	Use:   "allocate",
	Short: "allocate a new subnet",
	Long:  "allocate a new subnet, the smallest fitting free slice is automatically found and allocated to keep your IP space fragmentation low",
	Run:   runAlloc,
}

func init() {
	// Alloc-specific flags
	allocCmd.Flags().IntP("size", "s", -1, "size of the network to allocate")
	allocCmd.Flags().StringP("id", "I", "", "ID for the allocated block")
	allocCmd.Flags().StringP("parent", "p", "", "ID of the parent block to allocate from")
	allocCmd.Flags().StringP("description", "d", "", "description for the newly allocated subnet")
	allocCmd.Flags().Bool("in-place", false, "modify the input file in place")

	// Bind flags to viper
	viper.BindPFlag("size", allocCmd.Flags().Lookup("size"))
	viper.BindPFlag("id", allocCmd.Flags().Lookup("id"))
	viper.BindPFlag("parent", allocCmd.Flags().Lookup("parent"))
	viper.BindPFlag("description", allocCmd.Flags().Lookup("description"))
	viper.BindPFlag("in-place", allocCmd.Flags().Lookup("in-place"))
}

// Command returns the alloc command
func Command() *cobra.Command {
	return allocCmd
}

func runAlloc(cmd *cobra.Command, args []string) {
	outputFilename := viper.GetString("output-file")
	inputFilename := viper.GetString("input-file")
	inPlace, _ := cmd.Flags().GetBool("in-place")

	// output filename is set and in-place is set
	if outputFilename != "-" && inPlace {
		quitWithError(errors.New("cannot use --output-file and --in-place at the same time"))
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

	pool, err := netpool.FromAtf(atfFile)
	if err != nil {
		quitWithError(err)
	}

	superAllocSize, _ := atfFile.Superblock.Mask.Size()

	allocSize, _ := cmd.Flags().GetInt("size")
	allocID, _ := cmd.Flags().GetString("id")
	allocParent, _ := cmd.Flags().GetString("parent")
	allocDesc, _ := cmd.Flags().GetString("description")

	if allocSize > netcalc.AWS_MIN_SUBNET_SIZE || allocSize <= superAllocSize {
		quitWithError(errors.Errorf("requested block size is out of range (%d < block < %d)", superAllocSize, netcalc.AWS_MIN_SUBNET_SIZE))
	}

	// TODO: allow allocating from a suballocation with a flag
	net, err := pool.Pool.Alloc(allocSize)
	if err != nil {
		quitWithError(err)
	}

	newAlloc := &atf.Allocation{
		Ident:       allocID,
		Network:     &atf.IPNet{IPNet: net},
		Description: allocDesc,
	}

	// If parent is specified, add as sub-allocation
	if allocParent != "" {
		found := false
		for _, alloc := range atfFile.Allocations {
			if alloc.Ident == allocParent {
				alloc.SubAlloc = append(alloc.SubAlloc, newAlloc)
				found = true
				break
			}
		}
		if !found {
			quitWithError(errors.Errorf("parent allocation with ID %s not found", allocParent))
		}
	} else {
		atfFile.Allocations = append(atfFile.Allocations, newAlloc)
	}

	outBytes, err := yaml.Marshal(atfFile)
	if err != nil {
		quitWithError(err)
	}

	// output filename is not set and in-place is set
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
	data, err := ioutil.ReadAll(inputFile)
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

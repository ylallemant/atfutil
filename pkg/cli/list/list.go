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

package list

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"atfutil/pkg/atf"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all allocated blocks",
	Long:  "list all allocated blocks in the ATF file",
	Run:   runList,
}

func init() {
}

func Command() *cobra.Command {
	return listCmd
}

func runList(cmd *cobra.Command, args []string) {
	inputFilename := viper.GetString("file")

	inFile, err := getInputFile(inputFilename)
	if err != nil {
		quitWithError(err)
	}
	defer inFile.Close()

	atfFile, err := loadAtfFromFile(inFile)
	if err != nil {
		quitWithError(err)
	}

	fmt.Printf("Superblock: %s\n", atfFile.Superblock.String())
	if atfFile.Name != nil {
		fmt.Printf("Name: %s\n", *atfFile.Name)
	}
	fmt.Println()

	fmt.Println("Allocations:")
	printAllocations(atfFile.Allocations, "")

	os.Exit(0)
}

func printAllocations(allocations []*atf.Allocation, indent string) {
	for _, alloc := range allocations {
		desc := ""
		if alloc.Description != "" {
			desc = fmt.Sprintf(" - %s", alloc.Description)
		}
		reserved := ""
		if alloc.IsReserved {
			reserved = " [reserved]"
		}
		fmt.Printf("%s%s\t%s%s%s\n", indent, alloc.Network.String(), alloc.Ident, reserved, desc)

		if len(alloc.SubAlloc) > 0 {
			printAllocations(alloc.SubAlloc, indent+"  ")
		}
	}
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

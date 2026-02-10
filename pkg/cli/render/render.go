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

package render

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"atfutil/pkg/atf"
	"atfutil/pkg/netpool"
	renderPkg "atfutil/pkg/render"
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "render an atf.yaml to a human readable format",
	Run:   runRender,
}

func init() {
	// Render-specific flags
	renderCmd.Flags().BoolP("all-blocks", "a", false, "include free blocks when rendering")
	renderCmd.Flags().StringP("render-format", "f", "markdown", "render format (markdown)")

	// Bind flags to viper
	viper.BindPFlag("all-blocks", renderCmd.Flags().Lookup("all-blocks"))
	viper.BindPFlag("render-format", renderCmd.Flags().Lookup("render-format"))
}

// Command returns the render command
func Command() *cobra.Command {
	return renderCmd
}

func runRender(cmd *cobra.Command, args []string) {
	outBuffer := &bytes.Buffer{}

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

	pool, err := netpool.FromAtf(atfFile)
	if err != nil {
		quitWithError(err)
	}

	renderFree, _ := cmd.Flags().GetBool("all-blocks")
	renderFormat, _ := cmd.Flags().GetString("render-format")

	switch renderFormat {
	case "markdown":
		renderPkg.RenderPoolToMarkdown(outBuffer, pool, renderFree)
	default:
		quitWithError(errors.New("unknown render format"))
	}

	outputFilename := viper.GetString("output")
	outFile, err := getOutputFile(outputFilename)
	if err != nil {
		quitWithError(err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, outBuffer)
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

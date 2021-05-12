// Copyright 2006-2021 VMware, Inc.
// SPDX-License-Identifier: MIT
package license

import (
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const sourceFileExt = "go"
const existingLic = `/*
Copyright`

type License struct {
	projectLicense    []byte
	sourceFileLicense []byte
	sourceFiles       []string
}

var (
	projectLicenseFilename    string
	sourceFileLicenseFilename string
)

var LicenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Add license info to project",
	Long: `The license command will add a LICENSE file in the root of the project
as well as licensing text at the beginning of every source code file.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		var lic License

		if len(projectLicenseFilename) != 0 {
			pLicense, err := ioutil.ReadFile(projectLicenseFilename)
			if err != nil {
				return err
			}
			lic.projectLicense = pLicense
		}

		if len(sourceFileLicenseFilename) != 0 {
			sLicense, err := ioutil.ReadFile(sourceFileLicenseFilename)
			if err != nil {
				return err
			}
			lic.sourceFileLicense = sLicense
		}

		return lic.updateFiles()
	},
}

// updateFiles adds license content
func (l *License) updateFiles() error {

	changes := false

	if len(l.projectLicense) != 0 {
		licenseF, err := os.Create("LICENSE")
		if err != nil {
			return err
		}
		defer licenseF.Close()
		licenseF.WriteString(string(l.projectLicense))
		changes = true
	}

	if len(l.sourceFileLicense) != 0 {

		l.getSourceFiles(sourceFileExt)

		for _, sourceFile := range l.sourceFiles {
			sourceFileContent, err := ioutil.ReadFile(sourceFile)
			if err != nil {
				return err
			}

			// remove any existing license
			strippedContent, err := stripLicense(sourceFileContent)
			if err != nil {
				return err
			}

			licensedContent := string(l.sourceFileLicense) + strippedContent

			f, err := os.OpenFile(sourceFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}

			defer f.Close()

			f.WriteString(licensedContent)
		}
		changes = true
	}

	if !changes {
		return errors.New("No project or source code license files provided - no changes made")
	}

	return nil
}

// getSourceFiles finds all source code files
func (l *License) getSourceFiles(fileExt string) error {

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if len(info.Name()) > 4 {
			if info.Name()[len(info.Name())-3:] == ".go" {
				l.sourceFiles = append(l.sourceFiles, path)
			}
		}
		return nil
	})

	return nil
}

// stripLicense removes any existing license from source code files
func stripLicense(sourceFileContent []byte) (string, error) {

	var sourceFileContentOut string
	endOfLicenseFound := false

	if strings.HasPrefix(string(sourceFileContent), existingLic) {
		scanner := bufio.NewScanner(strings.NewReader(string(sourceFileContent)))
		for scanner.Scan() {
			if endOfLicenseFound {
				sourceFileContentOut = sourceFileContentOut + "\n" + string(scanner.Text())
			} else if scanner.Text() == "*/" {
				endOfLicenseFound = true
			}
		}
	} else {
		return string(sourceFileContent), nil
	}

	return sourceFileContentOut, nil
}

func init() {
	LicenseCmd.Flags().StringVarP(&projectLicenseFilename, "project-license", "p", "", "path to project license file")
	LicenseCmd.Flags().StringVarP(&sourceFileLicenseFilename, "source-code-license", "s", "", "path to file with source code license text")
}

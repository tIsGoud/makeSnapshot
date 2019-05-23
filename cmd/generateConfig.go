// Copyright © 2019 Albert W. Alberts <a.w.alberts@tisgoud.nl>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var sampleConfigFile string

// generateConfigCmd represents the generateConfig command
var generateConfigCmd = &cobra.Command{
	Use:     "generateConfig",
	Short:   "Generate a sample configuration file for makeSnapshot",
	Long:    `The generateConfig option creates an empty default configuration file in the application directory.`,
	Example: `makeSnapshot generateConfig`,
	Run: func(cmd *cobra.Command, args []string) {
		writeSampleConfigFile(sampleConfigFile)
	},
}

func init() {
	rootCmd.AddCommand(generateConfigCmd)

	generateConfigCmd.Flags().StringVarP(&sampleConfigFile, "sampleConfig", "s", defaultConfigName+".yaml", "config file to create")
}

func writeSampleConfigFile(configFile string) {
	if !fileExists(configFile) {
		file, err := os.Create(configFile)
		if err != nil {
			log.Printf("Error: %s", err)
		}
		file.WriteString("---\n")
		file.WriteString("baseURL: \"https://your.base.url\"\n")
		file.WriteString("tenant: \"your tenant name\"\n")
		file.WriteString("domain: \"your domain name\"\n")
		file.WriteString("username: \"your username without domain\"\n")
		file.WriteString("password: \"your password\"\n")
		file.WriteString("...\n")
		file.Sync()
		file.Close()
		log.Printf("Created config file %q", configFile)
	} else {
		log.Printf("Error: Unable to create %q, file or directory already exists", configFile)
	}
}

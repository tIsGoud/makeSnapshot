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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Commandline flag variables
var (
	configFile   string
	domain       string
	dryRun       bool
	keepExisting bool
	machineName  string
	trace        bool
)

// Internal variables
var (
	version           = "0.9.2"
	defaultConfigName = "makeSnapshot"
	userAgent         = "makeSnapShot " + version // Useragent with version number is used in the HTTP requests
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "makeSnapshot",
	Short: "Create a snapshot of a virtual machine",
	Long: `
makeSnapshot is a CLI tool to create a snapshot of a virtual machine on the KPN vRA platform.
Only one snapshot per VM is allowed. The default behaviour is to overwrite the existing snapshot.

Several API calls are needed to creates a snapshot of the virtual machine.

Tracing can be turned on to provide information on the progress.

After the snapshot request is send the status of the request is checked every 10 seconds.
The time between request and the final status can take half-a-minute or more.

Required parameters like the baseURL, tenant, domain, and credentials are read from a 'yaml' config file
---
baseURL: "https://base-platformURL"
tenant: "tenantName"
domain: "login domain"
userName: "userName"
password: "password"
...

When the snapshot is created the app exits with status code 0.
Otherwise the app exits with status code 1.
The exit code is not displayed but can be checked with 'echo $?'

DISCLAIMER:
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

Written by A.W. Alberts - Copyright © 2019 'tIsGoud
`,
	Version: version,
	Example: `  With tracing information and a non-default config file:
  makeSnapshot -c [config.yaml] -m [virtual machine name] -t

  Without tracing and with the default configuration file:
  makeSnapshot -m [virtual machine name]

  Note: The virtual machine name is case sensitive!`,

	Run: func(cmd *cobra.Command, args []string) {
		validateConfig()

		traceInfo(`Creating snapshot of virtual machine "` + machineName + `" for tenant "` + viper.GetString("tenant") + `"`)

		// Step 1 - Get bearer token (POST {baseURL}/identity/api/tokens)
		bearerToken := getBearerToken()

		// Step 2 - Get VirtualMachine Resource id  (GET {baseURL}/catalog-service/api/consumer/resources?page=1&limit=5000)
		virtualMachineID := getVirtualMachineResourceID(bearerToken, machineName)

		// Step 3 - Get snapshot resource resource action id (GET {baseURL}/catalog-service/api/consumer/resources/{machineID}/actions/)
		snapshotActionID := getSnapshotResourceActionID(bearerToken, virtualMachineID)

		// Step 4 - Get resource action template (GET {baseURL}/catalog-service/api/consumer/resources/{vmID}/actions/{snapshotActionID}/requests/template)
		getResourceActionTemplate() // Fake call, but could be a future enhancement to use the template to populate a struct and use the struct in Step 5.

		// On dry-run skip the snapshot request
		if !dryRun {
			// Step 5 - Send snapshot request (POST {baseURL}/catalog-service/api/consumer/resources/{vmID}/actions/{actionID}/requests/)
			requestStatusURL := sendSnapshotRequest(bearerToken, virtualMachineID, snapshotActionID)

			// Step 6 - Get request result state (GET {baseURL}/catalog-service/api/consumer/{requestStatusURL})
			getRequestResultState(bearerToken, requestStatusURL)
		} else {

			traceInfo("Step 5 - Skipped because of dry-run")
			traceInfo("Step 6 - Skipped because of dry-run")
		}

		// Silly message at the end of the program
		traceInfo("Bye from makeSnapshot")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file to create in the app directory (default "+defaultConfigName+".yaml)")
	rootCmd.Flags().StringVarP(&domain, "domain", "d", "", "login domain (overrides the domain value in the config file)")
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "r", false, "dry-run the application, running full initialization and pre-snapshot calls only")
	rootCmd.Flags().BoolVarP(&keepExisting, "keepExisting", "k", false, "do not overwrite a possible existing snapshot")
	rootCmd.Flags().StringVarP(&machineName, "machineName", "m", "", "name of the virtual machine to snapshot, case sensitive and required")
	rootCmd.Flags().BoolVarP(&trace, "trace", "t", false, "show tracing information")
	rootCmd.MarkFlagRequired("machineName")
	viper.BindPFlag("domain", rootCmd.Flags().Lookup("domain"))
}

// initConfig reads in config file
func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		// Search config in application directory with name {defaultConfigName} (without extension)
		viper.AddConfigPath(".")
		viper.SetConfigName(defaultConfigName)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		traceInfo("Using config file:" + viper.ConfigFileUsed())
	}
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func validateConfig() {
	if fileExists(viper.ConfigFileUsed()) {
		exitOnEmptyString("baseURL", viper.GetString("baseURL"))
		exitOnEmptyString("tenant", viper.GetString("tenant"))
		exitOnEmptyString("domain", viper.GetString("domain"))
		exitOnEmptyString("userName", viper.GetString("userName"))
		exitOnEmptyString("password", viper.GetString("password"))
	} else {
		log.Fatalf("Error: Unable to find configfile %q", viper.ConfigFileUsed())
	}
}

// Step 1 - Get bearer token (POST {baseURL}/identity/api/tokens)
func getBearerToken() string {

	traceInfo("Step 1 - Get bearer token")

	// Only once needed to get the bearer token
	var requestVars GetBearerTokenRequest
	requestVars.Username = viper.GetString("userName") + "@" + viper.GetString("domain")
	requestVars.Password = viper.GetString("password")
	requestVars.Tenant = viper.GetString("tenant")

	jsonValue, _ := json.Marshal(requestVars)

	// Create client
	client := &http.Client{}

	// Create request
	req, _ := http.NewRequest("POST", viper.GetString("baseURL")+"/identity/api/tokens", bytes.NewBuffer(jsonValue))

	// Headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	// Fetch Request and handle possible connection errors
	resp, err := client.Do(req)
	logFatalError(err)

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Handle HTTP response status != 200
	if resp.StatusCode != 200 {
		re := regexp.MustCompile(`("systemMessage":")(.*)(","moreInfoUrl)`)
		matches := re.FindStringSubmatch(string(respBody))
		log.Fatalf("Error: Unexpected HTTP response status code %d, %s", resp.StatusCode, matches[2])
	}

	var gbtResponse GetBearerTokenResponse
	err = json.Unmarshal(respBody, &gbtResponse)
	logFatalError(err)

	// Return the API bearerToken, doing nothing smart like caching based on the expiration date
	exitOnEmptyString("bearerToken", gbtResponse.ID)

	// Return the "full" token
	return "Bearer " + gbtResponse.ID
}

// Step 2 - Get VirtualMachine Resource id (GET {baseURL}/catalog-service/api/consumer/resources?page=1&limit=5000)
func getVirtualMachineResourceID(token, machine string) string {

	traceInfo("Step 2 - Get virtual machine resource ID for " + machine)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("GET", viper.GetString("baseURL")+"/catalog-service/api/consumer/resources?page=1&limit=5000", nil)

	// Headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", token)
	req.Header.Set("User-Agent", userAgent)

	parseFormErr := req.ParseForm()
	if parseFormErr != nil {
		log.Println(parseFormErr)
	}

	// Fetch Request and handle possible connection errors
	resp, err := client.Do(req)
	logFatalError(err)

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Handle HTTP response status != 200
	if resp.StatusCode != 200 {
		re := regexp.MustCompile(`<h1>(.*)</h1>`)
		matches := re.FindStringSubmatch(string(respBody))
		log.Fatalf("Error: %s", matches[1])
	}

	// RegEx tested on https://regex101.com/
	re := regexp.MustCompile(`"@type":"CatalogResource","id":"(?P<id>.{36})","iconId":"Infrastructure.CatalogItem.Machine.Virtual.vSphere","resourceTypeRef":{"id":"Infrastructure.Virtual","label":"Virtual Machine"},"name":".{3}(?P<name>` + machine + `)","description"`)
	matches := re.FindStringSubmatch(string(respBody))
	if matches == nil {
		log.Fatalf("Error: Unable to find Catalog Resource id for virtual machine %q", machine)
	} else {
		// Match found but only spaces (highly unlikely)
		exitOnEmptyString("machineID", matches[1])
	}
	return matches[1]
}

// Step 3 - Get snapshot resource resource action id (GET {baseURL}/catalog-service/api/consumer/resources/{vmID}/actions/)
func getSnapshotResourceActionID(token, vmID string) string {

	traceInfo("Step 3 - Get snapshot resource action ID for " + machineName)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("GET", viper.GetString("baseURL")+"/catalog-service/api/consumer/resources/"+vmID+"/actions/", nil)

	// Headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", token)
	req.Header.Set("User-Agent", userAgent)

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Handle HTTP response status != 200
	if resp.StatusCode != 200 {
		re := regexp.MustCompile(`<h1>(.*)</h1>`)
		matches := re.FindStringSubmatch(string(respBody))
		log.Fatalf("Error: %s", matches[1])
	}

	// RegEx tested on https://regex101.com/
	re := regexp.MustCompile(`"name":"Create VM Snapshot".*?"ACTION","id":"(?P<id>.*?)",`)
	matches := re.FindStringSubmatch(string(respBody))
	if matches == nil {
		log.Fatalf("Error: Unable to find Create Snapshot Action id")
	} else {
		// Match found but only spaces (highly unlikely)
		exitOnEmptyString("Create Snapshot Action ID", matches[1])
	}
	return matches[1]
}

// Step 4 - Get resource action template (GET {baseURL}/catalog-service/api/consumer/resources/{vmID}/actions/{actionID}/requests/template)
func getResourceActionTemplate() {
	traceInfo("Step 4 - Get resource action template")

	// // Create client
	// client := &http.Client{}

	// // Create request
	// req, err := http.NewRequest("GET", "https://vpc.kpnvdc.nl/catalog-service/api/consumer/resources/2a415ba9-81f5-4bff-b35f-bccfd5587165/actions/fcf490d5-a7e9-4640-be83-ac74d4484c91/requests/template", nil)

	// // Headers
	// req.Header.Add("Content-Type", "application/json")
	// req.Header.Add("Accept", "application/json")
	// req.Header.Add("Authorization", "Bearer MTU1NjUzNzA3NDIxNToxYTM0Y2Q2ZjBjY2RhMWQyYmJiZTp0ZW5hbnQ6QVBJTWFya2V0cGxhY2V1c2VybmFtZTphcGltYXJrZXRwbGFjZUB2cGMuY2xvdWRubGV4cGlyYXRpb246MTU1NjU2NTg3NDAwMDpkMmIwYTQ2OGEwYzEwYTNkMDhlYjg0OGNiNmYwOTJhMGFkZDVkZTE1NmU0NTMzZDY0OTBlZTkwMWU0ZTMwNmY2NGM5MjhjODBkYWJjNGFmZjNlZmJmM2ZhNGUxZjMxYWI4MjgyNjRjZTQ5OGJjYzkyYTcxZDUyNGMwNzk0NDlkYw==")
	// req.Header.Set("User-Agent", userAgent)

	// // Fetch Request
	// resp, err := client.Do(req)

	// if err != nil {
	// 	log.Println("Failure : ", err)
	// }

	// // Read Response Body
	// respBody, _ := ioutil.ReadAll(resp.Body)

	// // Display Results
	// log.Println("response Status : ", resp.Status)
	// log.Println("response Headers : ", resp.Header)
	// log.Println("response Body : ", string(respBody))
}

// Step 5 - Send snapshot request (POST {baseURL}/catalog-service/api/consumer/resources/{vmID}/actions/{actionID}/requests/)
func sendSnapshotRequest(token, vmID, snapshotActionID string) string {

	traceInfo("Step 5 - Send snapshot request for " + machineName)

	var json []byte

	// Ugly but working json string, could be improved by converting it into types (un- and marshalling)
	// Default behaviour is to remove the existing snapshot ("provider-deleteExisting")
	if keepExisting {
		json = []byte(`{"type": "com.vmware.vcac.catalog.domain.request.CatalogResourceRequest","data": {"provider-existingSnapshotName": null,"provider-deleteExisting": false,"provider-description": "Snapshotdescription","provider-name": "Snapshot name","provider-__ASD_PRESENTATION_INSTANCE": null,"provider-__asd_tenantRef": "` + viper.GetString("tenant") + `"},"description": "makeSnapshot call"}`)
	} else {
		json = []byte(`{"type": "com.vmware.vcac.catalog.domain.request.CatalogResourceRequest","data": {"provider-existingSnapshotName": null,"provider-deleteExisting": true,"provider-description": "Snapshotdescription","provider-name": "Snapshot name","provider-__ASD_PRESENTATION_INSTANCE": null,"provider-__asd_tenantRef": "` + viper.GetString("tenant") + `"},"description": "makeSnapshot call"}`)
	}
	body := bytes.NewBuffer(json)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", viper.GetString("baseURL")+"/catalog-service/api/consumer/resources/"+vmID+"/actions/"+snapshotActionID+"/requests/", body)

	// Headers
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	req.Header.Add("Accept", "application/json;charset=UTF-8")
	req.Header.Add("Authorization", token)
	req.Header.Set("User-Agent", userAgent)

	// Fetch Request
	resp, err := client.Do(req)
	logFatalError(err)

	// Handle HTTP response status != 200
	if resp.StatusCode != 201 {
		re := regexp.MustCompile(`<h1>(.*)</h1>`)
		// Read Response Body
		respBody, _ := ioutil.ReadAll(resp.Body)
		matches := re.FindStringSubmatch(string(respBody))
		log.Fatalf("Error: %s", matches[1])
	}

	exitOnEmptyString("Resource Action Request URL", resp.Header.Get("Location"))

	return resp.Header.Get("Location")
}

// Step 6 - Get request result state (GET {baseURL}/catalog-service/api/consumer/requests/{requestStatusURL})
func getRequestResultState(token, requestStatusURL string) {

	traceInfo("Step 6 - Get snapshot request status...")

	// Create client
	client := &http.Client{}

	// Create request
	req, _ := http.NewRequest("GET", requestStatusURL, nil)

	// Headers
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	req.Header.Add("Accept", "application/json;charset=UTF-8")
	req.Header.Add("Authorization", token)

	// RegEx tested on https://regex101.com/
	re := regexp.MustCompile(`"stateName":"(?P<state>.*?)"`)

	for {
		// Give the system some time before polling the request status
		time.Sleep(10 * time.Second)

		// Fetch Request
		resp, err := client.Do(req)

		if err != nil {
			log.Println("Failure : ", err)
		}

		// Read Response Body
		respBody, _ := ioutil.ReadAll(resp.Body)

		matches := re.FindStringSubmatch(string(respBody))
		traceInfo("Step 6 - Snapshot request status: " + matches[1])

		if matches[1] == "Failed" {
			log.Fatalf("Error: Snapshot request failed, check the vRA portal for more info")
		}
		if matches[1] == "Successful" {
			break
		}
	}
}

// Print trace info when the trace flag is set on the commandline
func traceInfo(info string) {
	if trace {
		log.Println(info)
	}
}

func exitOnEmptyString(stringName, stringValue string) {
	if len(strings.TrimSpace(stringValue)) == 0 {
		log.Fatalf("Error: zero-length string `%s`", stringName)
	}
}

// Log the error and exit
func logFatalError(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

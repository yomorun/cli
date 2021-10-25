package ga

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
	"github.com/yomorun/cli/pkg/log"
)

const apiURL = "https://www.google-analytics.com/collect"
const trackingID = "UA-47208480-12" // yomo.run

// Send the event to Google Analytics.
func Send(cliVersion string, cmd string, subCMD string) {
	_, err := http.PostForm(apiURL, url.Values{
		"v":   {"1"},                      // Protocol Version
		"tid": {trackingID},               // Tracking ID
		"ds":  {"yomo-cli"},               // Data Source
		"cid": {getClientID()},            // Client ID
		"ua":  {getUserAgent(cliVersion)}, // User Agent
		"t":   {"event"},                  // Hit type
		"ec":  {"yomo-cli"},               // Event Category
		"ea":  {cmd},                      // Event Action
		"el":  {subCMD},                   // Event Label
		"cn":  {"yomo-cli"},               // Campaign Name
		"cs":  {"yomo-cli"},               // Campaign Source
		"cm":  {"yomo-cli"},               // Campaign Medium
		"ck":  {cmd},                      // Campaign Keyword
		"cc":  {subCMD},                   // Campaign Content
	})

	if err != nil {
		log.FailureStatusEvent(os.Stdout, "[GA] send the event failed: %s", err.Error())
	}
}

// getClientID gets/creates an unique client ID.
func getClientID() string {
	// get home directory.
	dir, err := os.UserHomeDir()
	if err != nil {
		log.FailureStatusEvent(os.Stdout, "[GA] get the home directory failed: %s", err.Error())
		return generateNewClientID()
	}

	// the file location for Client ID.
	cidFilePath := filepath.Join(dir, "yomo-cid.txt")
	// read Client ID in the specified file.
	b, err := ioutil.ReadFile(cidFilePath)
	if err == nil {
		// the file exists, return the Client ID in this file.
		return string(b)
	}

	// generate a new Client ID if not exists.
	newCID := generateNewClientID()

	// create a text file to store the ClientID.
	f, err := os.Create(cidFilePath)
	if err != nil {
		log.FailureStatusEvent(os.Stdout, "[GA] create the file in home directory failed: %s", err.Error())
		return newCID
	}

	defer f.Close()

	_, err = f.Write([]byte(newCID))
	if err != nil {
		log.FailureStatusEvent(os.Stdout, "[GA] write the CID into the file in home directory failed: %s", err.Error())
	}

	return newCID
}

// generateNewClientID generates a random UUID (version 4) as ClientID.
func generateNewClientID() string {
	return uuid.New().String()
}

// getUserAgent returns the User-Agent for YoMo CLI.
func getUserAgent(cliVersion string) string {
	return fmt.Sprintf("yomo-cli/%s (%s)", cliVersion, runtime.GOOS)
}

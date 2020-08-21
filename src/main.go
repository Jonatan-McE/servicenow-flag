package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	LuxID             string `short:"l" long:"luxaforapiid" description:"Luxafor API key ID (comma seperated ID's are supported)" value-name:"<luxid>"`
	SNUsername        string `short:"u" long:"username" description:"ServiceNow account username " value-name:"<username>"`
	SNPassword        string `short:"p" long:"password" description:"ServiceNow account password " value-name:"<password>"`
	SNAssignmentGroup string `short:"a" long:"assignmentgroup" description:"ServiceNow account assignment-group " value-name:"<assignmentgroup>"`
	SNBaseURL         string `short:"c" long:"customurl" description:"ServiceNow custom base-url (e.g https://servicenow.com) " value-name:"<customurl>"`
	Low               int    `long:"low" description:"Low value for led color " value-name:"<low>"`
	High              int    `long:"high" description:"High value for led color " value-name:"<high>"`
	Verbose           bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
}

var luxaforBaseURL string = "https://api.luxafor.com/webhook/v1/actions/"
var serviceNowBaseURL string = "https://servicenow.com/api/now/table/task?"
var verbose bool

func main() {
	var luxID string = os.Getenv("SNF_LUXID")
	var snUser string = os.Getenv("SNF_SNUSER")
	var snPass string = os.Getenv("SNF_SNPASS")
	var snAssigngroup string = os.Getenv("SNF_SNASSIGNGROUP")
	var low int
	var high int
	var err error
	if low, err = strconv.Atoi(os.Getenv("SNF_LOW")); err != nil {
		low = 1
	}
	if high, err = strconv.Atoi(os.Getenv("SNF_HIGH")); err != nil {
		high = 2
	}
	if os.Getenv("SNF_SNBASEURL") != "" {
		serviceNowBaseURL = (os.Getenv("SNF_SNBASEURL") + "/api/now/table/task?")
	}
	if os.Getenv("SNF_VERBOSE") != "" {
		verbose, _ = strconv.ParseBool(os.Getenv("SNF_VERBOSE"))
	}
	// Parse the commandline options and argoments
	_, err = flags.Parse(&opts)
	if err != nil {
		return
	}

	// Override enviroment variables with command line argoments
	if opts.LuxID != "" {
		luxID = opts.LuxID
	}
	if opts.SNUsername != "" {
		snUser = opts.SNUsername
	}
	if opts.SNPassword != "" {
		snPass = opts.SNPassword
	}
	if opts.SNAssignmentGroup != "" {
		snAssigngroup = opts.SNAssignmentGroup
	}
	if opts.SNBaseURL != "" {
		serviceNowBaseURL = (opts.SNBaseURL + "/api/now/table/task?")
	}
	if opts.Low != 0 {
		low = opts.Low
	}
	if opts.High != 0 {
		high = opts.High
	}

	luxIDs := strings.Split(luxID, ",")

	// Check for missing variables
	if len(luxIDs) > 0 {
		if verbose {
			log(fmt.Sprintf("Luxafor API ID is set (%s)", luxIDs))
		}
	} else {
		log("Luxafor API ID is NOT set!")
		os.Exit(0)
	}
	if snUser != "" {
		if verbose {
			log("ServiceNow username is set")
		}
	} else {
		log("ServiceNow username is NOT set!")
		os.Exit(0)
	}
	if snPass != "" {
		if verbose {
			log("ServiceNow password is set")
		}
	} else {
		log("ServiceNow password is NOT set!")
		os.Exit(0)
	}
	if snAssigngroup != "" {
		if verbose {
			log(fmt.Sprintf("ServiceNow assignment group is set (%s)", snAssigngroup))
		}
	} else {
		log("ServiceNow assignment group is NOT set!")
		os.Exit(0)
	}
	if opts.SNBaseURL != "" || os.Getenv("SNBASEURL") != "" {
		if opts.Verbose {
			log(fmt.Sprintf("Custom ServiceNow base-url is set (%s)", serviceNowBaseURL))
		}
	}
	if low != 0 {
		if verbose {
			log(fmt.Sprintf("Min value is set to (%v)", low))
		}
	}
	if high != 0 {
		if verbose {
			log(fmt.Sprintf("Max value is set (%v)", high))
		}
	}

	log("Application starting...")

	// Reset flag status at startup
	for _, id := range luxIDs {
		err = queryLuxafor(id, "000000", "solid_color")
		if err != nil {
			log(fmt.Sprintf("Failed to updatge flag with error: %s", err))
		}
	}

	// Loop status-check every 5 min
	var previousNum int
	var numOfCycles int
	for {
		var flagColor string

		if numOfCycles != 0 {
			time.Sleep(300 * time.Second)
		}
		numOfCycles++

		currentNum, err := queryServiceNow(snUser, snPass, snAssigngroup)
		if err != nil {
			log(fmt.Sprintf("API call failed with error: %s", err))
			continue
		}

		if currentNum > 0 && currentNum <= low {
			flagColor = "00ff00" // Green
		} else if currentNum > low && currentNum <= high {
			flagColor = "0000ff" // Blue
		} else if currentNum > high {
			flagColor = "ff0000" // Red
		} else {
			flagColor = "000000" // Off
		}

		if currentNum != previousNum || numOfCycles > 11 {
			if currentNum != previousNum {
				for _, id := range luxIDs {
					err = queryLuxafor(id, flagColor, "blink")
					if err != nil {
						log(fmt.Sprintf("Failed to updatge flag with error: %s", err))
						return
					}
				}
				// Add a 3 sec delay to space out multiple POST requeests
				time.Sleep(3 * time.Second)
				for _, id := range luxIDs {
					err = queryLuxafor(id, flagColor, "solid_color")
					if err != nil {
						log(fmt.Sprintf("Failed to updatge flag with error: %s", err))
						return
					}
				}
				log(fmt.Sprintf("%d tickets in queue (Flag update)", currentNum))
			} else {
				for _, id := range luxIDs {
					err = queryLuxafor(id, flagColor, "solid_color")
					if err != nil {
						log(fmt.Sprintf("Failed to updatge flag with error: %s", err))
						return
					}
				}
				log(fmt.Sprintf("%d tickets in queue (Periodic flage update)", currentNum))
			}
			previousNum = currentNum
			numOfCycles = 0
		} else {
			log(fmt.Sprintf("%d tickets in queue (No update)", currentNum))
		}
	}

}

func queryServiceNow(user string, pass string, assignGroup string) (int, error) {
	type SNQueue struct {
		Result []struct {
			ID string `json:"number"`
		} `json:"result"`
	}
	var snQueue SNQueue

	// Custom ServiceNow API Options
	sysparms := []string{
		"sysparm_fields=number",
		"sysparm_exclude_reference_link=true",
		"sysparm_limit=200",
	}
	sysparmQuery := []string{
		"active=true",
		"assigned_to=",
		"assignment_group=",
	}

	// Create request url string
	url := serviceNowBaseURL
	for _, s := range sysparms {
		url = (url + "&" + s)
	}
	url = (url + "&sysparm_query=")
	for _, q := range sysparmQuery {
		url = (url + "^" + q)
	}
	url = (url + assignGroup)

	// Create and send HTTP GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(user, pass)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log(fmt.Sprintf("Data retrival failed with %s", resp.Status))
		return 0, nil
	}

	if verbose {
		log(fmt.Sprintf("Rest API GET  completed sucsessful: %s (ServiceNow)", resp.Status))
	}

	// Process returned data
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	json.Unmarshal([]byte(data), &snQueue)

	// Return number of results
	return len(snQueue.Result), nil
}

func queryLuxafor(id string, color string, action string) error {
	json := []byte(fmt.Sprintf("{\"userId\":\"%s\",\"actionFields\":{\"color\":\"custom\",\"custom_color\":\"%s\"}}", id, color))

	url := (luxaforBaseURL + action)

	// Create and send HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log(fmt.Sprintf("API call failed with %s", resp.Status))
		return nil
	}

	if verbose {
		log(fmt.Sprintf("Rest API POST completed sucsessful: %s (Luxafor Flag - %s)", resp.Status, action))
	}

	return nil
}

func log(message string) {
	t := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("%s - %s\n", t, message)
}

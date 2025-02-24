package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const MAX_CONCURRENT_JOBS = 2

type result struct {
	sumValue      int
	multiplyValue int
}

var mutex = &sync.Mutex{}

var statuscount = map[string]int{}

var default_payload_url = "https://raw.githubusercontent.com/Proviesec/directory-payload-list/main/directory-full-list.txt"
var fav_payload_url = "https://raw.githubusercontent.com/Proviesec/directory-files-payload-lists/main/directory-proviesec-favorite-list.txt"

var url string
var dirlist string
var generate_payload string
var generate_payload_length int
var showStatus string
var redirect string
var recursionLevel int
var bypass string
var bypassTooManyRequests string
var concurrency int
var output string
var onlydomains string
var requestAddHeader string
var requestAddAgent string
var checkBackslash string
var helpmode string
var filterWrongStatus200 string
var filterWrongSubdomain string
var filterStatusCode string
var filterTestLength string
var filterPossible404 string
var filterContentType string
var filterContentTypeList []string
var filterMatchWord string
var filterStatusCodeList []string
var filterStatusNot string
var filterStatusNotList []string
var filterLength string
var filterLengthList []string
var filterMaxLength int
var filterMinLength int
var filterLengthNot string
var filterLengthNotList []string
var lengthcount = map[int]int{}
var testlength int
var configfile string

type Config struct {
	URL                   string `json:"url"`
	Dirlist               string `json:"dirlist"`
	GeneratePayload       bool   `json:"generate_payload"`
	ShowStatus            bool   `json:"showStatus"`
	FilterTestLength      bool   `json:"filterTestLength"`
	FilterPossible404     bool   `json:"filterPossible404"`
	OnlyDomains           bool   `json:"onlydomains"`
	Redirect              bool   `json:"redirect"`
	RecursionLevel        int    `json:"recursionLevel"`
	Bypass                bool   `json:"bypass"`
	CheckBackslash        bool   `json:"checkBackslash"`
	BypassTooManyRequests bool   `json:"bypassTooManyRequests"`
	Concurrency           int    `json:"concurrency"`
	Output                string `json:"output"`
	Helpmode              bool   `json:"helpmode"`
	FilterWrongStatus200  bool   `json:"filterWrongStatus200"`
	FilterWrongSubdomain  bool   `json:"filterWrongSubdomain"`
}

func init() {

	// get config file from the command line
	flag.StringVar(&configfile, "configfile", "", "Config file")
	flag.StringVar(&configfile, "cf", "", "Config file")

	// get url parameter from name "url" in the command line
	flag.StringVar(&url, "url", "", "URL")
	flag.StringVar(&url, "u", "", "URL")

	// get directoryList parameter from name "directoryList" in the command line
	flag.StringVar(&dirlist, "dirlist", "", "Directory List")
	flag.StringVar(&dirlist, "d", "default", "Directory List")

	//get generate_payload parameter from name "generate_payload" in the command line
	flag.StringVar(&generate_payload, "generate_payload", "", "Generate Payload")
	flag.StringVar(&generate_payload, "g", "", "Generate Payload")

	// get status parameter from the command lline
	flag.StringVar(&showStatus, "showStatus", "false", "show status")
	flag.StringVar(&showStatus, "s", "false", "show status")

	// get testlength parameter from the command lline
	flag.StringVar(&filterTestLength, "filterTestLength", "false", "filterTestLength")
	flag.StringVar(&filterTestLength, "t", "false", "filterTestLength")

	// get filterpossible404 parameter from the command lline
	flag.StringVar(&filterPossible404, "filterPossible404", "false", "filterPossible404")
	flag.StringVar(&filterPossible404, "p404", "false", "filterPossible404")

	// get onlydomains parameter from the command lline
	flag.StringVar(&onlydomains, "onlydomains", "false", "only domains")
	flag.StringVar(&onlydomains, "od", "false", "only domains")

	// get concurrency parameter from the command line
	flag.StringVar(&redirect, "redirect", "true", "redirect")
	flag.StringVar(&redirect, "r", "true", "redirect")

	// get recursionLevel parameter from the command line
	flag.IntVar(&recursionLevel, "recursionLevel", 0, "recursionLevel")
	flag.IntVar(&recursionLevel, "rl", 0, "recursionLevel")

	// get bypass parameter from the command line
	flag.StringVar(&bypass, "bypass", "false", "bypass")
	flag.StringVar(&bypass, "b", "false", "bypass")

	// get checkBackslash parameter from the command line
	flag.StringVar(&checkBackslash, "checkBackslash", "false", "checkBackslash")
	flag.StringVar(&checkBackslash, "cb", "false", "checkBackslash")

	// get bypassTooManyRequests parameter from the command line
	flag.StringVar(&bypassTooManyRequests, "bypassTooManyRequests", "false", "bypassTooManyRequests")
	flag.StringVar(&bypassTooManyRequests, "btr", "false", "bypassTooManyRequests")

	// get concurrency parameter from the command line
	flag.IntVar(&concurrency, "concurrency", 1, "concurrency")
	flag.IntVar(&concurrency, "c", 1, "concurrency")

	// get output parameter from the command line
	flag.StringVar(&output, "output", "", "output")
	flag.StringVar(&output, "o", "", "output")

	// get helpmode parameter from the command line
	flag.StringVar(&helpmode, "helpmode", "false", "helpmode")
	flag.StringVar(&helpmode, "h", "false", "helpmode")

	// get filterWrongStatus200 parameter from the command line
	flag.StringVar(&filterWrongStatus200, "filterWrongStatus200", "false", "filterWrongStatus200")
	flag.StringVar(&filterWrongStatus200, "fws", "false", "filterWrongStatus200")

	//get filterWrongSubdomain parameter from the command line
	flag.StringVar(&filterWrongSubdomain, "filterWrongSubdomain", "false", "filterWrongSubdomain")
	flag.StringVar(&filterWrongSubdomain, "fwd", "false", "filterWrongSubdomain")

	// get filterContentType parameter from the command line
	flag.StringVar(&filterContentType, "filterContentType", "", "filterContentType")
	flag.StringVar(&filterContentType, "f", "", "filterContentType")

	// get filterMatchWord parameter from the command line
	flag.StringVar(&filterMatchWord, "filterMatchWord", "", "filterMatchWord")
	flag.StringVar(&filterMatchWord, "fm", "", "filterMatchWord")

	// get list of show status codes from the command line
	flag.StringVar(&filterStatusCode, "filterStatusCode", "", "Show only this status code")
	flag.StringVar(&filterStatusCode, "fsc", "", "Show only this status code")

	// get list of status code to be filtered from the command line
	flag.StringVar(&filterStatusNot, "filterStatusCodeNot", "", "Show not this status code")
	flag.StringVar(&filterStatusNot, "fscn", "", "Show not this status code")

	// get  length from the command line
	flag.StringVar(&filterLength, "filterLength", "-1", "Show only response length")
	flag.StringVar(&filterLength, "fl", "-1", "Show only response length")

	// get not show length from the command line
	flag.StringVar(&filterLengthNot, "filterLengthNot", "-1", "Don´t show this response length")
	flag.StringVar(&filterLengthNot, "fln", "-1", "Don´t show this response length")

	// get filterMaxLength from the command line
	flag.IntVar(&filterMaxLength, "filterMaxLength", -1, "Show only response length less than")
	flag.IntVar(&filterMaxLength, "fml", -1, "Show only response length less than")

	// get filterMinLength from the command line
	flag.IntVar(&filterMinLength, "filterMinLength", -1, "Show only response length greater than")
	flag.IntVar(&filterMinLength, "fmg", -1, "Show only response length greater than")

	// get add header from the command line
	flag.StringVar(&requestAddHeader, "requestAddHeader", "", "Add header to request")
	flag.StringVar(&requestAddHeader, "rah", "", "Add header to request")

	// get add agent from the command line
	flag.StringVar(&requestAddAgent, "requestAddAgent", "", "Add agent to request")
	flag.StringVar(&requestAddAgent, "raa", "", "Add agent to request")

	flag.Parse()

	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Println(err)
	}

	if url == "" {
		url = config.URL
	}

	if dirlist == "default" {
		if config.Dirlist == "" {
			dirlist = config.Dirlist
		}
	}

	filterStatusCodeList = strings.Split(filterStatusCode, ",")
	filterStatusNotList = strings.Split(filterStatusNot, ",")
	filterContentTypeList = strings.Split(filterContentType, ",")
	generate_payload_length, _ = strconv.Atoi(generate_payload)

	if generate_payload_length < 0 || generate_payload_length > 20000 {
		generate_payload_length = 20000
	}

	filterLengthList = strings.Split(filterLength, ",")
	filterLengthNotList = strings.Split(filterLengthNot, ",")

	// If the identified URL has neither http or https infront of it. Create both and scan them.
	if !strings.Contains(url, "http://") && !strings.Contains(url, "https://") {
		url = "https://" + url
	}
	// if the URL does not end with a /, add it.
	if !strings.HasSuffix(url, "/") && checkBackslash != "true" {
		url = url + "/"
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func GetResponseDetails(response *http.Response) (string, int, string, int, int) {
	// Get the response body as a string
	dataInBytes, _ := ioutil.ReadAll(response.Body)
	pageContent := string(dataInBytes)

	pageTitle := ""
	pageBody := ""
	// Find a substr
	titleStartIndex := strings.Index(pageContent, "<title>")
	if titleStartIndex == -1 {
		pageTitle = "No title element found"
	} else {
		// <title> = length = 7
		titleStartIndex += 7

		// Find the index of the closing tag
		titleEndIndex := strings.Index(pageContent, "</title>")
		if titleEndIndex == -1 {
			pageTitle = "No closing tag for title found."
		} else {
			pageTitle = "Page title:" + string([]byte(pageContent[titleStartIndex:titleEndIndex]))
		}
	}

	matchWord := ""
	if filterMatchWord != "" {
		// if filterMatchWord in pageContent then output the 10 postions before and after the matchWord
		if strings.Contains(pageContent, filterMatchWord) {
			matchWordIndex := strings.Index(pageContent, filterMatchWord)
			if matchWordIndex > 10 {
				matchWord = pageContent[matchWordIndex-10 : matchWordIndex+10]
			} else {
				matchWord = pageContent[0 : matchWordIndex+10]
			}
		}
	}
	bodyStartIndex := strings.Index(pageContent, "<body>")
	if bodyStartIndex != -1 {
		// <body> = length = 6
		bodyStartIndex += 6

		// Find the index of the closing tag
		bodyEndIndex := strings.Index(pageContent, "</body>")
		if bodyEndIndex != -1 {
			pageBody = string([]byte(pageContent[bodyStartIndex:bodyEndIndex]))
		}
	}
	// Match non-space character sequences.
	re := regexp.MustCompile(`[\S]+`)

	// Find all matches and return count.
	resultsCount := re.FindAllString(pageContent, -1)

	return pageTitle, len(pageContent), matchWord, len(resultsCount), len(pageBody)
}

func lineCounter(r io.Reader) (int, error) {
	count := 0
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		count++
	}

	return count, scanner.Err()
}

func createPayload(length int) []byte {

	// Create a byte slice of the specified length

	payload := make([]byte, length)

	// Fill the slice with the sequence of characters "a", "b", "c", etc.

	for i := 0; i < length; i++ {

		payload[i] = byte('a' + (i % 26))

	}

	// Return the payload

	return payload

}

func NextAlias(last string) string {
	if last == "" {
		return "a"
	} else if last[len(last)-1] == 'z' {
		return last[:len(last)-1] + "aa"
	} else {
		return last[:len(last)-1] + string(last[len(last)-1]+1)
	}
}

func urlFuzzScanner(directoryList []string) {
	// open the text file directoryList and read the lines in it
	// if directoryList is default or fav then use one of the default lists
	if directoryList[0] == "default" || directoryList[0] == "fav" {

		directoryListUrl := default_payload_url
		if directoryList[0] == "fav" {
			directoryListUrl = fav_payload_url
		}

		dir_resp, err := http.Get(directoryListUrl)

		if err != nil {
			log.Fatal(err)
		}
		defer dir_resp.Body.Close()
		if dir_resp.StatusCode == 200 {
			// get the response body as a string
			dataInBytes, _ := ioutil.ReadAll(dir_resp.Body)
			pageContent := string(dataInBytes)
			payload_create, err := os.Create("default_payload.txt") // Truncates if file already exists, be careful!
			if err != nil {
				log.Fatalf("failed creating file: %s", err)
			}
			defer payload_create.Close()
			_, err = payload_create.WriteString(pageContent)
		}
		directoryList[0] = "default_payload.txt"
	}
	if generate_payload == "true" {
		payload_create, err := os.Create("random_payload.txt") // Truncates if file already exists, be careful!
		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		}
		last := ""
		for i := 0; i < generate_payload_length; i++ {
			next := NextAlias(last)
			_, err = payload_create.WriteString(next + "")
			last = next
		}
		directoryList[0] = "random_payload.txt"
	}

	file, err := os.Open(directoryList[0])
	if err != nil {
		fmt.Fprint(os.Stdout, "\r"+err.Error()+strings.Repeat(" ", 100)+"\n")
	}
	defer file.Close()
	// read the lines in the text file

	file_lines, err := os.OpenFile(directoryList[0], os.O_RDONLY, 0444)
	if err != nil {
		fmt.Fprint(os.Stdout, "\r"+err.Error()+strings.Repeat(" ", 100)+"\n")
	}
	defer file_lines.Close()
	count_lines, _ := lineCounter(file_lines)
	if concurrency <= 0 {
		concurrency = MAX_CONCURRENT_JOBS
	}

	concurrent := make(chan int, concurrency)

	scanner := bufio.NewScanner(file)
	count := 0

	if output == "" {
		output = "output.txt"
	} else {
		output = output + ".txt"
	}

	file_create, err := os.Create(output) // Truncates if file already exists, be careful!
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer file_create.Close()

	if filterTestLength == "true" {

		random_string := "213804asdad32432sdf"
		url_test := ""
		if strings.Contains(url, "#PSFUZZ#") {
			url_test = strings.Replace(url, "#PSFUZZ#", random_string, 1)
		} else {
			url_test = url + random_string
		}
		resp_test, err := sendRequest(url_test, "")
		if err != nil {
			// Handle the error here and continue the execution of the program
			fmt.Println("Error occurred while sending request:", err)
			return
		}
		_, testlength, _, _, _ = GetResponseDetails(resp_test)
	}

	for scanner.Scan() {
		percent := (count * 100 / count_lines)
		fill := strings.Repeat("x", percent) + strings.Repeat("-", 100-percent)
		_, _ = fmt.Fprint(os.Stdout, "\r[")
		_, _ = fmt.Fprintf(os.Stdout, "%s]", fill)
		p := int(count * 100 / (count_lines + 1))
		_, _ = fmt.Fprintf(os.Stdout, "\t%d %%", p)

		word := scanner.Text()
		// check if the line is empty
		if word == "" {
			continue
		}

		concurrent <- 1
		count++

		go func(count int, url string, showStatus string) {
			// find the wildcard in the url
			if strings.Contains(url, "#PSFUZZ#") {
				url = strings.Replace(url, "#PSFUZZ#", word, 1)
			} else {
				url = url + word
			}
			// write the result to the file
			testUrl(url, showStatus, file_create, false, requestAddHeader, bypass)
			<-concurrent
		}(count, url, showStatus)
	}
	return
}

func sendRequest(url string, requestHeader string) (*http.Response, error) {

	if requestHeader != "" {
		requestHeader = requestAddHeader
	}
	// create a new http client
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	// create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprint(os.Stdout, "\r"+err.Error()+strings.Repeat(" ", 100)+"\n")
		return nil, err
	}
	// set the user agent
	if requestAddAgent != "" {
		req.Header.Set("User-Agent", requestAddAgent)
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36 (zz99)")
	}

	// if requestHeader is not empty then add the headers to the request
	if requestHeader != "" {
		headers := strings.Split(requestHeader, ",")
		for _, header := range headers {
			header = strings.TrimSpace(header)
			headerSplit := strings.Split(header, ":")
			if len(headerSplit) == 2 {
				req.Header.Set(headerSplit[0], headerSplit[1])
			}
		}
	}

	// define the request with a timeout of 2 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// make the request
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		fmt.Fprint(os.Stdout, "\r"+err.Error()+strings.Repeat(" ", 100)+"\n")
		return nil, err
	}
	return resp, nil
}

func testUrl(url string, showStatus string, file_create *os.File, redirected bool, requestHeader string, bypassResponse string) {
	resp, err := sendRequest(url, requestHeader)

	if err != nil {
		// Handle the error here and continue the execution of the program
		fmt.Println("Error occurred while sending request:", err)
		return
	}

	mutex.Lock()
	statuscount[resp.Status] = statuscount[resp.Status] + 1
	mutex.Unlock()

	responseAnalyse(resp, url, showStatus, file_create, redirected, bypassResponse)
}

func responseAnalyse(resp *http.Response, url string, showStatus string, file_create *os.File, redirected bool, bypassResponse string) {
	// create output string variable
	var outputString string
	if checkStatus(strconv.Itoa(resp.StatusCode)) && checkContentType(resp.Header.Get("Content-Type")) {
		title, length, matchWord, countWords, _ := GetResponseDetails(resp)
		if ((filterMatchWord != "" && matchWord != "") || filterMatchWord == "") && ((contains(filterLengthList, strconv.Itoa(length)) || contains(filterLengthList, "-1")) && (!contains(filterLengthNotList, strconv.Itoa(length)) || contains(filterLengthNotList, "-1")) || checkLength(strconv.Itoa(length))) {

			// save the length of the response
			mutex.Lock()
			lengthcount[length] = lengthcount[length] + 1
			mutex.Unlock()

			if filterWrongStatus200 == "true" {
				if strings.Contains(title, "Technical subdomain") || strings.Contains(title, "Page Not Available") || strings.Contains(title, "Access Gateway") || strings.Contains(title, "Not Found") || strings.Contains(title, "ERROR") || strings.Contains(title, "Error") || strings.Contains(title, "Forbidden") || strings.Contains(title, "Bad Request") || strings.Contains(title, "Internal Server Error") || strings.Contains(title, "Bad Gateway") || length <= 1 {
					return
				}
			}
			if filterWrongSubdomain == "true" {
				if length == 21 || length == 180 || strings.Contains(title, "Origin DNS error") || resp.StatusCode == http.StatusPermanentRedirect {
					return
				}
			}
			if filterTestLength == "true" {
				if length == testlength {
					return
				}
			}
			if strings.Contains(title, "404") {
				if filterPossible404 == "true" {
					return
				}
				title = title + " -- possibile a 404"
			}
			if filterMatchWord != "" && matchWord != "" {
				title = title + " -- MATCH: " + matchWord + " --"
			}
			if redirected {
				outputString = "redirected to "
			}
			outputString = outputString + url + " - " + resp.Status + " " + strings.Repeat(" ", 100) + "\n" + title + " Length:" + strconv.Itoa(length) + " Words:" + strconv.Itoa(countWords) + "\n"

			// convert resp.ContentLength to string
			fmt.Fprint(os.Stdout, "\r"+outputString)
			if redirected {
				fmt.Fprint(os.Stdout, "redirected to ")
			}
			if redirect == "true" && (resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently) { //status code 302
				redirUrl, _ := resp.Location()
				testUrl(redirUrl.String(), showStatus, file_create, true, requestAddHeader, bypassResponse)
			}
			if onlydomains == "true" {
				// if url string start with http:// or https:// then remove it
				url_string := url
				if strings.HasPrefix(url_string, "http://") {
					url_string = strings.Replace(url_string, "http://", "", 1)
				}
				if strings.HasPrefix(url_string, "https://") {
					url_string = strings.Replace(url_string, "https://", "", 1)
				}
				outputString = url_string + "\n"
			}
		}
		if (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized) && bypassResponse == "true" {
			bypassStatusCode40x(url, showStatus, file_create)
		}
		// if the status code is 429 then wait 5 seconds and retry
		if bypassTooManyRequests == "true" && resp.StatusCode == http.StatusTooManyRequests {
			time.Sleep(5 * time.Second)
			testUrl(url, showStatus, file_create, redirected, requestAddHeader, bypassResponse)
		}
	}
	_, _ = file_create.WriteString(outputString)
}

func bypassStatusCode40x(url string, showStatus string, file_create *os.File) {

	arrayPath := [13]string{"/*", "//.", "/%2e/", "/%2f/", "/./", "/", "/*/", "/..;/", "/..%3B/", "////", "/%20", "%00", "#test"}
	for _, element := range arrayPath {
		time.Sleep(time.Second)
		testUrl(url+element, showStatus, file_create, false, "", "false")
	}
	arrayExtensions := [10]string{".yml", ".php", ".html", ".zip", ".txt", ".yaml", ".wadl", ".htm", ".asp", ".aspx"}
	for _, element := range arrayExtensions {
		time.Sleep(time.Second)
		testUrl(url+element, showStatus, file_create, false, "", "false")
	}
	arrayHeader := [6]string{"X-Custom-IP-Authorization127.0.0.1", "Host:Localhost", "X-Forwarded-For:127.0.0.1:80", "X-Forwarded-For:http://127.0.0.1", "X-Custom-IP-Authorization:127.0.0.1", "Content-Length:0"}
	for _, element := range arrayHeader {
		time.Sleep(time.Second)
		testUrl(url, showStatus, file_create, false, element, "false")
	}
}
func checkContentType(contentType string) bool {
	if contains(filterContentTypeList, contentType) || contains(filterContentTypeList, "") {
		return true
	}
	return false
}

func checkStatus(s string) bool {
	if (contains(filterStatusCodeList, s) || showStatus == "true") && !contains(filterStatusNotList, s) {
		return true
	}
	for _, v := range filterStatusNotList {
		// check if in v is the string "-" Example 200-250 and compare the two numbers
		if strings.Contains(v, "-") {
			// split the string in two parts
			splitted := strings.Split(v, "-")
			// check if the length is in the range
			if splitted[0] <= s && s <= splitted[1] {
				return false
			}
		}
	}
	for _, v := range filterStatusCodeList {
		// check if in v is the string "-" Example 400-405 and compare the two numbers
		if strings.Contains(v, "-") {
			// split the string in two parts
			splitted := strings.Split(v, "-")
			// check if the length is in the range
			if splitted[0] <= s && s <= splitted[1] {
				return true
			}
		}
	}
	return false
}

func loadConfig(file string) (*Config, error) {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	return &config, err
}

func checkLength(s string) bool {
	// if filterMaxLength is bigger than 0 then check if the length is smaller than the filterMaxLength
	if filterMaxLength > 0 {
		// convert the string to int
		i, _ := strconv.Atoi(s)
		if i > filterMaxLength {
			return false
		}
	}
	// if filterMinLength is bigger than 0 then check if the length is bigger than the filterMinLength
	if filterMinLength > 0 {
		i, _ := strconv.Atoi(s)
		if filterMinLength > i {
			return false
		}
	}
	for _, v := range filterLengthNotList {
		// check if in v is the string "-" Example 10-200 and compare the two numbers
		if strings.Contains(v, "-") {
			// split the string in two parts
			splitted := strings.Split(v, "-")
			// check if the length is in the range
			if splitted[0] <= s && s <= splitted[1] {
				return false
			}
		}
	}
	for _, v := range filterLengthList {
		// check if in v is the string "-" Example 10-200 and compare the two numbers
		if strings.Contains(v, "-") {
			// split the string in two parts
			splitted := strings.Split(v, "-")
			// check if the length is in the range
			if splitted[0] <= s && s <= splitted[1] {
				return true
			}
		}
	}
	return false
}

func main() {
	fmt.Fprint(os.Stdout, "PSFuzz - Payload Scanner\n")
	fmt.Fprint(os.Stdout, "Version: 0.8.0\n")
	fmt.Fprint(os.Stdout, "Author: Proviesec\n")
	// ouput ascii art
	fmt.Fprint(os.Stdout, `                                                                                                                   
%%%%%%%%%%%   %%%%%%%%%%%%   %%%%%%%%%%  %%%%    %%%% %%%%%%%%%%%% %%%%%%%%%%%%                 
%%%%%%%%%%%%  %%%%%%%%%%%    %%%%        %%%%    %%%%        %%%%         %%%%                   
        %%%%   %%%%          %%%%        %%%%    %%%%       %%           %%                        
        %%%%      % %%%      %%%%%%%%%%  %%%%    %%%%   %%%          %%%                                      
%%%%%%%%%            %%%%%   %%%%        %%%%    %%%%  %%%%         %%%%                                          
%%%%          %%%%%%%%%%%%%  %%%%        %%%%%%%#%%%% %%%%%%%%%%%% %%%%%%%%%%%%                                   
%%%            %%%%%%%%%%     %%           %%%%%%        %%%%%%%%    %%%%%%%%                       

	`)

	directoryList := strings.Split(dirlist, ",")

	// check the directory list, if the found in the url
	urlFuzzScanner(directoryList)
	fmt.Fprint(os.Stdout, "\n")
	fmt.Println(statuscount) // map[string]int
	fmt.Println(lengthcount)
}

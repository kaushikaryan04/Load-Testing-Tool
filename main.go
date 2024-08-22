package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {

	reader := bufio.NewReader(os.Stdin)
	var err error
	fmt.Println("Enter Url")
	var url string
	_, err = fmt.Scan(&url)
	if err != nil {
		fmt.Println("err while parsing url", err)
		return
	}

	fmt.Println("What kind of request should be tested for POST or GET")
	RequestMethod, _ := reader.ReadString('\n')
	RequestMethod = strings.ToLower(RequestMethod)
	fmt.Println("url", url)
	fmt.Println("Enter duration for each user in minutes")
	var durationString string
	_, err = fmt.Scan(&durationString)
	TimeDuration, err := time.ParseDuration(durationString + "m")
	if err != nil {
		fmt.Println("Enter an Integer for minute")
		return
	}

	fmt.Println("Number of users to simulate")
	var Users int
	_, err = fmt.Scan(&Users)
	if err != nil {
		fmt.Println("error in reading no. of users")
		return
	}

	Parameters := CollectParameters()
	res := make(chan *result)
	var wg sync.WaitGroup
	for i := 0; i < Users; i++ {
		wg.Add(1)
		fmt.Println("request no", i)
		if RequestMethod == "post" {
			ParametersString := ConvertParamsToString(Parameters)
			go MakeRequest(url, ParametersString, RequestMethod, TimeDuration, res, &wg)
		} else {
			go MakeRequest(FormatURLForGet(url, Parameters), "", "get", TimeDuration, res, &wg)
		}
	}
	go func() {
		wg.Wait()
		close(res)
	}()
	var totalRequests int
	var totalErrors int
	var totalDuration time.Duration
	for r := range res {
		totalRequests++
		if r.error != nil {
			totalErrors++
		}
		totalDuration += r.ResponseTime

	}
	fmt.Printf("Total requests: %d\n", totalRequests)
	fmt.Printf("Total errors: %d\n", totalErrors)
	fmt.Printf("Average response time: %v\n", totalDuration/time.Duration(totalRequests))
	fmt.Printf("Requests per second: %.2f\n", float64(totalRequests)/TimeDuration.Seconds())

}

type result struct {
	ResponseTime time.Duration
	error        error
}

func MakeRequest(url string, Parameters string, RequestMethod string, duration time.Duration, results chan<- *result, wg *sync.WaitGroup) {
	defer wg.Done()
	EndDuration := time.Now().Add(duration)

	for time.Now().Before(EndDuration) {
		start := time.Now()
		var err error
		if RequestMethod == "post" {
			_, err = http.Post(url, "application/json", bytes.NewBuffer([]byte(Parameters)))
			if err != nil {
				fmt.Println("error from server", err)
			}
		} else {
			_, err = http.Get(url)
			if err != nil {
				fmt.Println("error from server", err)
			}
		}
		results <- &result{
			ResponseTime: time.Since(start),
			error:        err,
		}

		fmt.Println("Success from server")
	}
}
func FormatURLForGet(url string, ParameterMap map[string]string) string {
	if len(ParameterMap) == 0 {
		return url
	}
	var UrlString strings.Builder
	UrlString.WriteString(url)

	for key, value := range ParameterMap {
		UrlString.WriteString(fmt.Sprintf("?%s=%s", key, value))
	}

	return UrlString.String()

}

func ConvertParamsToString(ParameterMap map[string]string) string {

	var ParameterString strings.Builder
	ParameterString.WriteString(`{`)
	for key, value := range ParameterMap {
		temp := fmt.Sprintf(`"%s":"%s",`, key, value)

		ParameterString.WriteString(temp)
	}
	ParameterString.WriteString(`}`)
	return ParameterString.String()

}

func CollectParameters() map[string]string {

	ParameterMaps := make(map[string]string)

	fmt.Println("Enter Parameters in the format of key=value pair if any parameters are required for the request \n Enter exit if done")
	Scanner := bufio.NewScanner(os.Stdin)
	for {
		Scanner.Scan()
		input := Scanner.Text()

		if strings.ToLower(input) == "exit" {
			break
		}
		pair := strings.Split(input, "=")
		if len(pair) != 2 {
			fmt.Println("Invalid pair type make sure to split the key value using = ")
			continue
		}
		ParameterMaps[strings.TrimSpace(pair[0])] = strings.TrimSpace(pair[1])
	}

	return ParameterMaps
}

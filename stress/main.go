package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	vegeta "github.com/tsenart/vegeta/lib"
)

const (
	devBaseURL          = "https://personalization-dev.rtl-di.nl/internal"
	devPublicationPoint = "stress_test"
	devCampaign         = "stress"
	devSignalType       = "id"
)

func makePostRequest(endpoint string, message interface{}) (int, error) {
	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	url := devBaseURL + endpoint
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return -1, err
	}

	// debug only
	if b, err := ioutil.ReadAll(resp.Body); err == nil {
		fmt.Println(string(b))
	}

	return resp.StatusCode, nil
}

func writeKeyToFile(f *os.File, key string, items []string) {
	w := bufio.NewWriter(f)

	fmt.Fprint(w, key, ";", strings.Join(items[:], ","))
	fmt.Fprint(w, "\n")

	w.Flush()
}

func upload() {
	start := time.Now()

	min := 100000
	max := 250000

	// body := make(map[string]interface{})

	f, err := os.OpenFile("./items.csv", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// add data to model
	for i := 0; i < 1600000; i++ {

		k := uuid.New()
		sig := k.String()

		// generate random items
		var recommendedItems []string
		for j := 0; j < 25; j++ {
			val := rand.Intn(max-min+1) + min
			recommendedItems = append(recommendedItems, strconv.Itoa(val))
		}

		// body["publicationPoint"] = devPublicationPoint
		// body["campaign"] = devCampaign
		// body["signal"] = sig
		// body["recommendations"] = recommendedItems

		// sc, err := makePostRequest("/streaming", body)
		// if err != nil {
		// 	panic(err)
		// }

		// store Key for stress test later
		writeKeyToFile(f, sig, recommendedItems)

		// if sc != http.StatusCreated {
		// 	panic(errors.New("status code is " + strconv.Itoa(sc)))
		// }
	}

	elapsed := time.Since(start)
	log.Printf("Uploading took %s", elapsed)
}

func stress() {

	body := `
	{
		"publicationPoint": "stress_test",
		"campaign": "stress",
		"signals": [
		  {
			"id": "639811c3-650d-41a6-a995-922817b33586"
		  }
		]
	  }
	`

	rate := vegeta.Rate{Freq: 800, Per: time.Second}
	duration := 5 * time.Second
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "POST",
		URL:    "https://personalization-dev.rtl-di.nl/public/recommend",
		Body:   []byte(body),
	})
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
		metrics.Add(res)
	}
	metrics.Close()

	fmt.Printf("Max latency: %s\n", metrics.Latencies.Max)
	fmt.Printf("Success rate: %f\n", metrics.Success*100)
	fmt.Printf("Status code: %v\n", metrics.StatusCodes)
}

func main() {
	upload()
}

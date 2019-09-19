package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/rtlnl/data-personalization-api/models"
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
		log.Fatal().Msg(err.Error())
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

func writeKeyToFile(f *os.File, entry *models.SingleEntry) {
	w := bufio.NewWriter(f)

	b, _ := json.Marshal(entry)

	fmt.Fprint(w, string(b))
	fmt.Fprint(w, "\n")

	w.Flush()
}

func upload() {
	start := time.Now()

	min := 100000
	max := 250000

	f, err := os.OpenFile("./data/items_test.jsonl", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	defer f.Close()

	// add data to model
	for i := 0; i < 5; i++ {

		k := uuid.New()
		sig := k.String()

		entry := models.SingleEntry{}

		// generate random items
		var recommendedItems []models.ItemScore
		for j := 0; j < 25; j++ {
			val := rand.Intn(max-min+1) + min
			score := rand.Float64()

			is := models.ItemScore{}
			is["item"] = strconv.Itoa(val)
			is["score"] = fmt.Sprintf("%f", score)

			recommendedItems = append(recommendedItems, is)
		}

		entry.SignalID = sig
		entry.Recommended = recommendedItems

		// store Key for stress test later
		writeKeyToFile(f, &entry)
	}

	elapsed := time.Since(start)
	log.Info().Msgf("Uploading took %s", elapsed)
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

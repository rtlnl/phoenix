package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
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

func writeKeyToFile(k string, mu *sync.Mutex) {
	// debug only
	// fmt.Println(k)

	mu.Lock()
	defer mu.Unlock()

	f, err := os.OpenFile("./keys.txt", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprint(w, "\n")
	fmt.Fprint(w, k)

	err = w.Flush() // Don't forget to flush!
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	start := time.Now()

	min := 0
	max := 50
	totRoutines := 2

	body := make(map[string]interface{})

	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(totRoutines)

	// add data to model
	for i := 0; i < totRoutines; i++ {
		go func(i int) {
			defer wg.Done()

			var recommendedItems []string
			k := uuid.New()
			sig := k.String()

			// generate random items
			for j := 0; j < 5; j++ {
				val := rand.Intn(max-min+1) + min
				recommendedItems = append(recommendedItems, "item_"+strconv.Itoa(val))
			}

			body["publicationPoint"] = devPublicationPoint
			body["campaign"] = devCampaign
			body["signal"] = sig
			body["recommendations"] = recommendedItems

			sc, err := makePostRequest("/streaming", body)
			if err != nil {
				panic(err)
			}

			// store Key for stress test later
			writeKeyToFile(sig, &mu)

			if sc != http.StatusCreated {
				panic(errors.New("status code is " + strconv.Itoa(sc)))
			}
		}(i)
	}

	wg.Wait()

	elapsed := time.Since(start)
	log.Printf("Uploading took %s", elapsed)
}

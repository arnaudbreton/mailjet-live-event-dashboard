/**
 * This file provided by Facebook is for non-commercial testing and evaluation purposes only.
 * Facebook reserves all rights not expressly granted.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * FACEBOOK BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
 * ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

type eventPayload struct {
	Event string `json:"event"`
	MessageID int `json:"MessageID"`
}

type eventItem struct {
	EventType string
	MessageID int
}

type messagePayload struct {
	ApiKey string
	ApiSecret string
	FromEmail string
	Recipient string
	Body string
	Subject string
}

type mailjetAPIMessagePayload struct {
	FromEmail string
	Subject string
	Recipient string
	Body string `json:"Text-part"`
}

type mailjetConfig struct {
	BaseUrl string
}

const dataFile = "./events.json"
const configFilePath = "./config.json"

var eventMutex = new(sync.Mutex)

var config = mailjetConfig{}

// Handle events
func handleevents(w http.ResponseWriter, r *http.Request) {
	// Since multiple requests could come in at once, ensure we have a lock
	// around all file operations
	eventMutex.Lock()
	defer eventMutex.Unlock()

	// Stat the file, so we can find its current permissions
	fi, err := os.Stat(dataFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to stat the data file (%s): %s", dataFile, err), http.StatusInternalServerError)
		return
	}

	// Read the events from the file.
	eventData, err := ioutil.ReadFile(dataFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to read the data file (%s): %s", dataFile, err), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "POST":
		// Decode the JSON data
		events := make([]eventItem, 0)
		if err := json.Unmarshal(eventData, &events); err != nil {
			http.Error(w, fmt.Sprintf("Unable to Unmarshal events from data file (%s): %s", dataFile, err), http.StatusInternalServerError)
			return
		}

		response, _ := ioutil.ReadAll(r.Body)
		log.Println("New event payload received", string(response))
		// log.Println("Response", string(response))
		eventPayload := eventPayload{}
		json.Unmarshal(response, &eventPayload)
		// log.Printf("Unmarshal %+v\n", event)

		// Add a new event to the in memory slice of events
		eventItem := eventItem{
			EventType: eventPayload.Event,
			MessageID: eventPayload.MessageID,
		}
		events = append(events, eventItem)
		// fmt.Printf("%+v\n", events)

		// Marshal the events to indented json.
		eventData, err = json.Marshal(events)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to marshal events to json: %s", err), http.StatusInternalServerError)
			return
		}

		// Write out the events to the file, preserving permissions
		err := ioutil.WriteFile(dataFile, eventData, fi.Mode())
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to write events to data file (%s): %s", dataFile, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		io.Copy(w, bytes.NewReader(eventData))

	case "GET":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		// stream the contents of the file to the response
		io.Copy(w, bytes.NewReader(eventData))

	default:
		// Don't know the method, so error
		http.Error(w, fmt.Sprintf("Unsupported method: %s", r.Method), http.StatusMethodNotAllowed)
	}
}

// Handle messages
func handleMessages(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":
		response, _ := ioutil.ReadAll(r.Body)
		log.Println("New message payload received", string(response))

		messagePayload := messagePayload{}
		json.Unmarshal(response, &messagePayload)

		payload := mailjetAPIMessagePayload{
			FromEmail: messagePayload.FromEmail,
			Recipient: messagePayload.Recipient,
			Subject: messagePayload.Subject,
			Body: messagePayload.Body,
		}
		payloadMarshalled, err := json.Marshal(payload)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error when marshalling payload : %s", err), http.StatusInternalServerError)
			return
		}

		client := &http.Client{}
		req, _ := http.NewRequest("POST", config.BaseUrl + "/v3/send/message", bytes.NewReader(payloadMarshalled))
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(messagePayload.ApiKey, messagePayload.ApiSecret)

		mailjetResponse, err := client.Do(req)
		if  err != nil {
			http.Error(w, fmt.Sprintf("Unable to POST the message to Mailjet : %s", err), http.StatusInternalServerError)
			return
		}
		log.Println("Payload POST-ed to Mailjet Send API", mailjetResponse)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		// io.Copy(w, bytes.NewReader(messagePayload))

	default:
		// Don't know the method, so error
		http.Error(w, fmt.Sprintf("Unsupported method: %s", r.Method), http.StatusMethodNotAllowed)
	}
}

func main() {
	port := ""
	if len(os.Args) == 2 {
		port = os.Args[1]
	}
	if port == "" {
		port = "3000"
	}

	// Read the events from the file.
	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to read the config file (%s): %s", configFilePath, err), http.StatusInternalServerError)
		return
	}
	json.Unmarshal(configFile, &config)
	log.Println("Read config", config)

	http.HandleFunc("/events.json", handleevents)
	http.HandleFunc("/messages", handleMessages)

	http.Handle("/", http.FileServer(http.Dir("./public")))
	log.Println("Server started: http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

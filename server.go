package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"github.com/gorilla/mux"
)

type eventPayload interface{}

type mailjetAPIEvent struct {
	Event string `json:"event"`
}

type eventItem struct {
	EventType string
	Payload eventPayload
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
	Body string `json:"Html-part"`
}

type eventSetupPayload struct {
	ApiKey string
	ApiSecret string
	EventType string
	CallbackUrl string
}

type mailjetAPIEventCallbackUrlPayload struct {
	EventType string
	Url string
}

type mailjetConfig struct {
	BaseUrl string `json:"base_url"`
	MaxEventsCount int `json:"max_events_count"`
	Default map[string]string `json:"default"`
}

type apiError struct {
	ErrorMessage string
}

const dataFileBaseName = "events_%s.json"
const configFilePath = "./config.json"
const eventCallbackUrlBaseUrl = "/v3/REST/eventcallbackurl"

var eventMutex = new(sync.Mutex)

var config = mailjetConfig{}

var TraceLogger *log.Logger
var ErrorLogger *log.Logger

func handleError(w http.ResponseWriter, message string, status int) {
	ErrorLogger.Println(status, message)

	e := apiError{ErrorMessage: message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(e)
}

// Handle events
func handleEvents(w http.ResponseWriter, r *http.Request) {
	// Since multiple requests could come in at once, ensure we have a lock
	// around all file operations
	eventMutex.Lock()
	defer eventMutex.Unlock()

	vars := mux.Vars(r)
	apiKey := vars["apikey"]
	if apiKey == "" {
		handleError(w, "An API Key must be provided", http.StatusBadRequest)
		return
	}

	dataFileSession := fmt.Sprintf(dataFileBaseName, apiKey)

	// Stat the file, so we can find its current permissions
	var fi os.FileInfo
	var errStat error
	fi, errStat = os.Stat(dataFileSession)
	if errStat != nil {
		f, err := os.Create(dataFileSession)
		if err != nil {
			handleError(w, "Error when creating session file", http.StatusInternalServerError)
		}

		fi, _ = f.Stat()
		f.WriteString("[]")
		f.Close()
	}

	// Read the events from the file.
	eventData, err := ioutil.ReadFile(dataFileSession)
	if err != nil {
		handleError(w, fmt.Sprintf("Unable to read the data file (%s): %s", dataFileSession, err), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "POST":
		// Decode the JSON data
		events := make([]eventItem, 0)
		if err := json.Unmarshal(eventData, &events); err != nil {
			handleError(w, fmt.Sprintf("Unable to Unmarshal events from data file (%s): %s", dataFileSession, err), http.StatusInternalServerError)
			return
		}

		response, _ := ioutil.ReadAll(r.Body)
		TraceLogger.Println("New event payload received", string(response))

		// Add a new event to the in memory slice of events
		var mjEvent mailjetAPIEvent
		json.Unmarshal(response, &mjEvent)

		var mjEventPayload eventPayload
		json.Unmarshal(response, &mjEventPayload)
		newEventItem := eventItem{
			EventType: mjEvent.Event,
			Payload: mjEventPayload,
		}
		events = append([]eventItem{newEventItem}, events...)
		if config.MaxEventsCount > 0 && len(events) > config.MaxEventsCount {
			events = events[:config.MaxEventsCount]
		}

		// Marshal the events to indented json.
		eventData, err = json.Marshal(events)
		if err != nil {
			handleError(w, fmt.Sprintf("Unable to marshal events to json: %s", err), http.StatusInternalServerError)
			return
		}

		// Write out the events to the file, preserving permissions
		err := ioutil.WriteFile(dataFileSession, eventData, fi.Mode())
		if err != nil {
			handleError(w, fmt.Sprintf("Unable to write events to data file (%s): %s", dataFileSession, err), http.StatusInternalServerError)
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
		handleError(w, fmt.Sprintf("Unsupported method: %s", r.Method), http.StatusMethodNotAllowed)
	}
}

// Handle messages
func handleMessages(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":
		reqBody, _ := ioutil.ReadAll(r.Body)
		TraceLogger.Println("New message payload received", string(reqBody))

		messagePayload := messagePayload{}
		json.Unmarshal(reqBody, &messagePayload)

		if messagePayload.ApiKey == "" {
			handleError(w, "API key is mandatory", http.StatusBadRequest)
			return
		}

		if messagePayload.ApiSecret == "" {
			handleError(w, "API secret is mandatory", http.StatusBadRequest)
			return
		}

		if messagePayload.FromEmail == "" {
			handleError(w, "FromEmail is mandatory", http.StatusBadRequest)
			return
		}

		if messagePayload.Recipient == "" {
			messagePayload.Recipient = messagePayload.FromEmail
		}

		if messagePayload.Subject == "" {
			handleError(w, "Subject is mandatory", http.StatusBadRequest)
			return
		}

		if messagePayload.Body == "" {
			handleError(w, "Body is mandatory", http.StatusBadRequest)
			return
		}

		payload := mailjetAPIMessagePayload{
			FromEmail: messagePayload.FromEmail,
			Recipient: messagePayload.Recipient,
			Subject: messagePayload.Subject,
			Body: messagePayload.Body,
		}
		payloadMarshalled, err := json.Marshal(payload)
		if err != nil {
			handleError(w, fmt.Sprintf("Error when marshalling payload : %s", err), http.StatusInternalServerError)
			return
		}

		client := &http.Client{}
		req, _ := http.NewRequest("POST", config.BaseUrl + "/v3/send/message", bytes.NewReader(payloadMarshalled))
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(messagePayload.ApiKey, messagePayload.ApiSecret)

		mailjetResponse, err := client.Do(req)
		if  err != nil {
			handleError(w, err.Error(), mailjetResponse.StatusCode)
			return
		}
		TraceLogger.Println("Payload POST-ed to Mailjet Send API", mailjetResponse)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		io.Copy(w, bytes.NewReader(reqBody))

	default:
		// Don't know the method, so error
		handleError(w, fmt.Sprintf("Unsupported method: %s", r.Method), http.StatusMethodNotAllowed)
	}
}

// Handle messages
func handleEventSetup(w http.ResponseWriter, r *http.Request) {
	
	switch r.Method {
	case "POST":
		reqMessage, _ := ioutil.ReadAll(r.Body)
		TraceLogger.Println("New event setup payload received", string(reqMessage))

		p := eventSetupPayload{}
		json.Unmarshal(reqMessage, &p)

		if p.ApiKey == "" {
			handleError(w, "API key is mandatory", http.StatusBadRequest)
			return
		}

		if p.ApiSecret == "" {
			handleError(w, "API secret is mandatory", http.StatusBadRequest)
			return
		}

		if p.EventType == "" {
			handleError(w, "EventType is mandatory", http.StatusBadRequest)
			return
		}

		if p.CallbackUrl == "" {
			handleError(w, "CallbackUrl is mandatory", http.StatusBadRequest)
			return
		}

		mjPayload := mailjetAPIEventCallbackUrlPayload{
			EventType: p.EventType,
			Url: p.CallbackUrl,
		}
		payloadMarshalled, err := json.Marshal(mjPayload)
		if err != nil {
			handleError(w, fmt.Sprintf("Error when marshalling payload : %s", err), http.StatusInternalServerError)
			return
		}

		client := &http.Client{}

		baseEventUrl, _ := url.Parse(fmt.Sprintf("%s/%s", config.BaseUrl, eventCallbackUrlBaseUrl))
		eventUrl, err := url.Parse(fmt.Sprintf("%s/%s", baseEventUrl, fmt.Sprintf("%s|%t", p.EventType, false)))
		if err != nil {
			TraceLogger.Println("Error while building event url", err)
			handleError(w, fmt.Sprintf("Error while building event url : %s", err), http.StatusInternalServerError)
			return
		}

		getReq, _ := http.NewRequest("GET", eventUrl.String(), nil)
		getReq.SetBasicAuth(p.ApiKey, p.ApiSecret) 
		getResponse, err := client.Do(getReq)

		TraceLogger.Println("Mailjet API GET response to", eventUrl.String(), getResponse)
		if getResponse.StatusCode == 401 {
			handleError(w, fmt.Sprintf("Unauthorized"), http.StatusUnauthorized)
			return
		} else if getResponse.StatusCode == 404 {
			postReq, _ := http.NewRequest("POST", baseEventUrl.String(), bytes.NewReader(payloadMarshalled))
			postReq.Header.Set("Content-Type", "application/json")
			postReq.SetBasicAuth(p.ApiKey, p.ApiSecret)

			postResponse, err := client.Do(postReq)
			if  err != nil {
				handleError(w, fmt.Sprintf("Unable to create the eventcallbackurl at Mailjet : %s", err), http.StatusInternalServerError)
				return
			}
			TraceLogger.Println("Mailjet API POST response to", baseEventUrl.String(), postResponse)
		} else {
			putReq, _ := http.NewRequest("PUT", eventUrl.String(), bytes.NewReader(payloadMarshalled))
			putReq.SetBasicAuth(p.ApiKey, p.ApiSecret) 
			putReq.Header.Set("Content-Type", "application/json")
			putResponse, err := client.Do(putReq)
			if  err != nil {
				handleError(w, fmt.Sprintf("Unable to update the eventcallbackurl to Mailjet : %s", err), http.StatusInternalServerError)
				return
			}

			TraceLogger.Println("Mailjet API PUT response to", eventUrl, putResponse)
		}

		w.Header().Set("Content-Type", "application/json")
		io.Copy(w, bytes.NewReader(reqMessage))

	default:
		// Don't know the method, so error
		handleError(w, fmt.Sprintf("Unsupported method: %s", r.Method), http.StatusMethodNotAllowed)
	}
}

// Handle messages
func handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		configJson, _ := json.Marshal(&config)

		w.Header().Set("Content-Type", "application/json")
		io.Copy(w, bytes.NewReader(configJson))
	default:
		// Don't know the method, so error
		handleError(w, fmt.Sprintf("Unsupported method: %s", r.Method), http.StatusMethodNotAllowed)
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

	TraceLogger = log.New(os.Stdout,
        "TRACE: ",
        log.Ldate|log.Ltime|log.Lshortfile)

	ErrorLogger = log.New(os.Stderr,
        "ERROR: ",
        log.Ldate|log.Ltime|log.Lshortfile)

	// Read the events from the file.
	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to read the config file (%s): %s", configFilePath, err), http.StatusInternalServerError)
		return
	}
	json.Unmarshal(configFile, &config)
	TraceLogger.Println(fmt.Sprintf("Read config %+v", config))

	r := mux.NewRouter()
    r.HandleFunc("/config", handleConfig)
    r.HandleFunc("/apikey/{apikey}/events", handleEvents)
	r.HandleFunc("/events/setup", handleEventSetup)
	r.HandleFunc("/messages", handleMessages)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))
    http.Handle("/", r)

	TraceLogger.Println("Server started: http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

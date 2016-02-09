package main

import (
	"io"
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	log "github.com/Sirupsen/logrus"
)

type Handlers struct {
	mapiAuth string
	matHostHeader string
	mapiHost string
	matHost string
}

func pingHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "pong")
}

func (h Handlers) buildInfoHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "build-info")
}

func (h Handlers) contentPreviewHandler(w http.ResponseWriter, r *http.Request) {

	log.WithFields(log.Fields{
		"requestUri" : r.RequestURI,
		"transactionId" : r.Header.Get("X-Request-Id"),
	}).Info("received request");

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	//TODO logging
	if uuid == "" {
		log.Fatal("missing UUID")
	}

	success, mapiResp := h.getNativeContent(uuid, w, r)
	if !success { return }
	success, matResp := h.getTransformedContent(uuid, *mapiResp, w, r)
	if(!success) { return }
	io.Copy(w, matResp.Body)
}

func ( h Handlers) getNativeContent(uuid string, w http.ResponseWriter, r *http.Request) (ok bool, mapiResp *http.Response)  {
	methode := fmt.Sprintf("http://%s%s%s", h.mapiHost, mapiPath, uuid)

	//TODO logging
	log.Printf("sending to MAPI at " + methode);

	client := &http.Client{}
	mapiReq, err := http.NewRequest("GET", methode, nil)
	mapiReq.Header.Set("Authorization", "Basic " + h.mapiAuth)
	mapiResp, err = client.Do(mapiReq)


	if err !=nil {
		//TODO logging
		log.Fatal(err)
	}

	//TODO logging
	fmt.Printf("mapi the status code %v", mapiResp.StatusCode)


	//TODO logging when error handling desided
	if mapiResp.StatusCode !=200 {
		if mapiResp.StatusCode == 404 {
			w.WriteHeader(http.StatusNotFound);
			return
		}
		//TODO break this down
		w.WriteHeader(http.StatusBadGateway)
		return false, nil
	}
	return true, mapiResp
}

func ( h Handlers) getTransformedContent(uuid string, mapiResp http.Response, w http.ResponseWriter, r *http.Request) (ok bool, matResp *http.Response)  {

	// order of writing a response
	//header
	//responseCode
	//body

	matUrl := fmt.Sprintf("http://%s%s%s", h.matHost, matPath, uuid)
	log.Printf("sending to MAT at " + matUrl);
	client2 := &http.Client{}
	matReq, _ := http.NewRequest("POST", matUrl, mapiResp.Body)
	matReq.Header.Add("Host", h.matHostHeader)
	matReq.Header.Add("Content-Type", "application/json")
	matResp, _ = client2.Do(matReq)

	if matResp.StatusCode !=200 {
		//TODO break this down
		fmt.Printf("---the status code %v", matResp.StatusCode)
		w.WriteHeader(http.StatusInternalServerError)
		return false, nil
	}

	fmt.Printf("mat the status code %v", matResp.StatusCode)
	return true, matResp
}

func (h Handlers) mapiCheck()  fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Articvle Preview Service will not work",
		Name:             "Methode Api Availablilty Check",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: "Checks that Methode API Service is reachable. Article Preview Service requests native content from Methode API service.",
		Checker:          func() (string, error) {
			return checkMehtodeApiAvailablity(h.mapiHost, h.mapiAuth)
		},
	}
}

package main

import (
	"io"
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/Sirupsen/logrus"
)
const 	TransactionIdHeader = "X-Request-Id"
const mapiPath = "/eom-file/"
const matPath ="/content-transform/"

var logger = log.New()

type Handlers struct {
	mapiAuth      string
	matHostHeader string
	mapiUri       string
	matUri        string
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func (h Handlers) buildInfoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "build-info")
}

func (h Handlers) contentPreviewHandler(w http.ResponseWriter, r *http.Request) {
	logger.Formatter = new(log.JSONFormatter)
	logger.WithFields(log.Fields{
		"requestUri" : r.RequestURI,
		"transactionId" : r.Header.Get(TransactionIdHeader),
	}).Info("received request");

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	success, mapiResp := h.getNativeContent(uuid, w, r)
	if !success { return }
	success, matResp := h.getTransformedContent(uuid, *mapiResp, w, r)
	if(!success) { return }
	io.Copy(w, matResp.Body)
	matResp.Body.Close()
}

func ( h Handlers) getNativeContent(uuid string, w http.ResponseWriter, r *http.Request) (ok bool, mapiResp *http.Response) {
	methode := fmt.Sprintf("%s%s%s", h.mapiUri, mapiPath, uuid)

	logger.Formatter = new(log.JSONFormatter)
	logger.WithFields(log.Fields{
		"requestUri" : methode,
		"transactionId" : r.Header.Get(TransactionIdHeader),
	}).Info("Request to MAPI");
	mapiReq, err := http.NewRequest("GET", methode, nil)
	mapiReq.Header.Set("Authorization", "Basic " + h.mapiAuth)
	mapiResp, err = client.Do(mapiReq)

	//this happens when hostname cannot be resolved or host is not accessible
	if err !=nil {
		log.WithFields(log.Fields{"error" : err, "transactionId" : r.Header.Get(TransactionIdHeader)}).Warnf("Cannot reach MAPI host")
		w.WriteHeader(http.StatusServiceUnavailable)
		return false, nil
	}
	if mapiResp.StatusCode !=200 {
		w.WriteHeader(http.StatusNotFound);
		log.WithFields(log.Fields{"MapiStatus" : mapiResp.StatusCode, "transactionId" : r.Header.Get(TransactionIdHeader)}).Warnf("Request to MAPI failed")
		return false, nil
	}

	logger.WithFields(log.Fields{"status" : mapiResp.Status, "transactionId" : mapiResp.Header.Get(TransactionIdHeader)}).Info("Response from MAPI")
	return true, mapiResp
}

func ( h Handlers) getTransformedContent(uuid string, mapiResp http.Response, w http.ResponseWriter, r *http.Request) (ok bool, matResp *http.Response) {

	// order of writing a response
	//header
	//responseCode
	//body

	matUrl := fmt.Sprintf("%s%s%s", h.matUri, matPath, uuid)
	log.Printf("sending to MAT at "+matUrl);
	matReq, _ := http.NewRequest("POST", matUrl, mapiResp.Body)
	matReq.Host = h.matHostHeader
	matReq.Header.Set("Content-Type", "application/json")
	matResp, _ = client.Do(matReq)

	if matResp.StatusCode != 200 {
		//TODO break this down
		fmt.Printf("---the status code %v\n", matResp.StatusCode)
		w.WriteHeader(http.StatusInternalServerError)
		return false, nil
	}

	fmt.Printf("mat the status code %v\n", matResp.StatusCode)
	return true, matResp
}

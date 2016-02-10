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
	if(!success) {
		mapiResp.Body.Close()
		return
	}
	io.Copy(w, matResp.Body)
	matResp.Body.Close()
}

func ( h Handlers) getNativeContent(uuid string, w http.ResponseWriter, r *http.Request) (ok bool, resp *http.Response) {
	logger.Formatter = new(log.JSONFormatter)
	requestUrl := fmt.Sprintf("%s%s%s", h.mapiUri, mapiPath, uuid)

	logger.WithFields(log.Fields{
		"requestUri" : requestUrl,
		"transactionId" : r.Header.Get(TransactionIdHeader),
	}).Info("Request to MAPI");

	req, err := http.NewRequest("GET", requestUrl, nil)
	req.Header.Set("Authorization", "Basic " + h.mapiAuth)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)

	//this happens when hostname cannot be resolved or host is not accessible
	if err !=nil {
		log.WithFields(log.Fields{"error" : err, "transactionId" : r.Header.Get(TransactionIdHeader)}).Warnf("Cannot reach MAPI host")
		w.WriteHeader(http.StatusServiceUnavailable)
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusNotFound);
		log.WithFields(log.Fields{"MAPI.Status" : resp.StatusCode, "transactionId" : r.Header.Get(TransactionIdHeader)}).Warnf("Request to MAPI failed")
		return false, nil
	}

	logger.WithFields(log.Fields{"status" : resp.Status, "transactionId" : resp.Header.Get(TransactionIdHeader)}).Info("Response from MAPI")
	return true, resp
}

func ( h Handlers) getTransformedContent(uuid string, mapiResp http.Response, w http.ResponseWriter, r *http.Request) (ok bool, resp *http.Response) {
	logger.Formatter = new(log.JSONFormatter)
	requestUrl := fmt.Sprintf("%s%s%s", h.matUri, matPath, uuid)

	logger.WithFields(log.Fields{
		"requestUri" : requestUrl,
		"transactionId" : r.Header.Get(TransactionIdHeader),
	}).Info("Request to MAT");

	req, err := http.NewRequest("POST", requestUrl, mapiResp.Body)
	req.Host = h.matHostHeader
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)

	//this happens when hostname cannot be resolved or host is not accessible
	if err !=nil {
		log.WithFields(log.Fields{"error" : err, "transactionId" : r.Header.Get(TransactionIdHeader)}).Warnf("Cannot reach MAT host")
		w.WriteHeader(http.StatusServiceUnavailable)
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusNotFound);
		log.WithFields(log.Fields{"MAT.Status" : resp.StatusCode, "transactionId" : r.Header.Get(TransactionIdHeader)}).Warnf("Request to MAT failed")
		return false, nil
	}

	logger.WithFields(log.Fields{"status" : resp.Status, "transactionId" : mapiResp.Header.Get(TransactionIdHeader)}).Info("Response from MAT")
	return true, resp
}

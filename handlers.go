package main

import (
	"io"
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/Sirupsen/logrus"
	tid "github.com/Financial-Times/transactionid-utils-go"
)

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
	transactionId := tid.GetTransactionIDFromRequest(r)
	logger.Formatter = new(log.JSONFormatter)
	logger.WithFields(log.Fields{
		"requestUri" : r.RequestURI,
		"transactionId" : transactionId,
	}).Info("request started");

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	success, mapiResp := h.getNativeContent(uuid, transactionId, w)
	if !success { return }
	success, matResp := h.getTransformedContent(uuid, *mapiResp, w)
	if(!success) {
		mapiResp.Body.Close()
		return
	}
	w.Header().Set(tid.TransactionIDHeader, matResp.Header.Get(tid.TransactionIDHeader))
	io.Copy(w, matResp.Body)
	matResp.Body.Close()
}

func ( h Handlers) getNativeContent(uuid string, transactionId string, w http.ResponseWriter) (ok bool, resp *http.Response) {
	logger.Formatter = new(log.JSONFormatter)
	requestUrl := fmt.Sprintf("%s%s%s", h.mapiUri, mapiPath, uuid)

	logger.WithFields(log.Fields{
		"requestUri" : requestUrl,
		"transactionId" : transactionId,
	}).Info("Request to MAPI");

	req, err := http.NewRequest("GET", requestUrl, nil)

	req.Header.Set(tid.TransactionIDHeader, transactionId)
	req.Header.Set("Authorization", "Basic " + h.mapiAuth)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)

	//this happens when hostname cannot be resolved or host is not accessible
	if err !=nil {
		log.WithFields(log.Fields{"error" : err, "transactionId" : req.Header.Get(tid.TransactionIDHeader)}).Warnf("Cannot reach MAPI host")
		w.WriteHeader(http.StatusServiceUnavailable)
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusNotFound);
		log.WithFields(log.Fields{"MAPI.Status" : resp.StatusCode, "transactionId" : resp.Header.Get(tid.TransactionIDHeader)}).Warnf("Request to MAPI failed")
		return false, nil
	}
	logger.WithFields(log.Fields{"status" : resp.Status, "transactionId" : resp.Header.Get(tid.TransactionIDHeader)}).Info("Response from MAPI")
	return true, resp
}

func ( h Handlers) getTransformedContent(uuid string, mapiResp http.Response, w http.ResponseWriter) (ok bool, resp *http.Response) {
	logger.Formatter = new(log.JSONFormatter)
	requestUrl := fmt.Sprintf("%s%s%s", h.matUri, matPath, uuid)

	logger.WithFields(log.Fields{
		"requestUri" : requestUrl,
		"transactionId" : mapiResp.Header.Get(tid.TransactionIDHeader),
	}).Info("Request to MAT");

	req, err := http.NewRequest("POST", requestUrl, mapiResp.Body)
	req.Host = h.matHostHeader
	req.Header.Set(tid.TransactionIDHeader, mapiResp.Header.Get(tid.TransactionIDHeader))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)

	//this happens when hostname cannot be resolved or host is not accessible
	if err !=nil {
		log.WithFields(log.Fields{"error" : err, "transactionId" : req.Header.Get(tid.TransactionIDHeader)}).Warnf("Cannot reach MAT host")
		w.WriteHeader(http.StatusServiceUnavailable)
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusNotFound);
		log.WithFields(log.Fields{"MAT.Status" : resp.StatusCode, "transactionId" : req.Header.Get(tid.TransactionIDHeader)}).Warnf("Request to MAT failed")
		return false, nil
	}

	logger.WithFields(log.Fields{"status" : resp.Status, "transactionId" : mapiResp.Header.Get(tid.TransactionIDHeader)}).Info("Response from MAT")
	return true, resp
}

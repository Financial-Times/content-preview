package main

import (
	"io"
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/Sirupsen/logrus"
	tid "github.com/Financial-Times/transactionid-utils-go"
	"golang.org/x/net/context"
)

const mapiPath = "/eom-file/"
const matPath ="/content-transform/"
const uuidKey = "uuid"

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
		"transactionId" : tid.GetTransactionIDFromRequest(r),
	}).Info("request started");

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	ctx := tid.TransactionAwareContext(context.Background(), r.Header.Get(tid.TransactionIDHeader))
	ctx = context.WithValue(ctx, uuidKey, uuid)

	success, mapiResp := h.getNativeContent(ctx, w)
	if !success { return }
	success, matResp := h.getTransformedContent(ctx, *mapiResp, w)
	if(!success) {
		mapiResp.Body.Close()
		return
	}
	io.Copy(w, matResp.Body)
	matResp.Body.Close()
}

func ( h Handlers) getNativeContent(ctx context.Context, w http.ResponseWriter) (ok bool, resp *http.Response) {
	logger.Formatter = new(log.JSONFormatter)
	uuid := ctx.Value(uuidKey).(string)
	requestUrl := fmt.Sprintf("%s%s%s", h.mapiUri, mapiPath, uuid)


	transactionId, _ := tid.GetTransactionIDFromContext(ctx)
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

func ( h Handlers) getTransformedContent(ctx context.Context, mapiResp http.Response, w http.ResponseWriter) (ok bool, resp *http.Response) {
	logger.Formatter = new(log.JSONFormatter)
	uuid := ctx.Value(uuidKey).(string)
	requestUrl := fmt.Sprintf("%s%s%s?preview=true", h.matUri, matPath, uuid)
	transactionId, _ := tid.GetTransactionIDFromContext(ctx)

	//TODO we need to assert that mapiResp.Header.Get(tid.TransactionIDHeader) ==  transactionId
	//to ensure that we are logging exactly what is actually passed around in the headers
	logger.WithFields(log.Fields{
		"requestUri" : requestUrl,
		"transactionId" : transactionId,
	}).Info("Request to MAT");

	req, err := http.NewRequest("POST", requestUrl, mapiResp.Body)
	req.Host = h.matHostHeader
	req.Header.Set(tid.TransactionIDHeader, transactionId)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)

	//this happens when hostname cannot be resolved or host is not accessible
	if err !=nil {
		log.WithFields(log.Fields{"error" : err, "transactionId" : transactionId }).Warnf("Cannot reach MAT host")
		w.WriteHeader(http.StatusServiceUnavailable)
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusNotFound);
		log.WithFields(log.Fields{"MAT.Status" : resp.StatusCode, "transactionId" : transactionId }).Warnf("Request to MAT failed")
		return false, nil
	}

	logger.WithFields(log.Fields{"status" : resp.Status, "requestUrl": req.URL.String(), "transactionId" : transactionId}).Info("Response from MAT")
	return true, resp
}

package main

import (
	"fmt"
	tid "github.com/Financial-Times/transactionid-utils-go"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"io"
	"net/http"
)

const uuidKey = "uuid"

type ContentHandler struct {
	serviceConfig *ServiceConfig
	log           *AppLogger
	metrics       *Metrics
}

func (h ContentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	h.log.TransactionStartedEvent(r.RequestURI, tid.GetTransactionIDFromRequest(r), uuid)
	ctx := tid.TransactionAwareContext(context.Background(), r.Header.Get(tid.TransactionIDHeader))
	ctx = context.WithValue(ctx, uuidKey, uuid)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	success, nativeContentSourceAppResponse := h.getNativeContent(ctx, w)

	if !success {
		return
	}
	success, transformAppResponse := h.getTransformedContent(ctx, *nativeContentSourceAppResponse, w)
	if !success {
		nativeContentSourceAppResponse.Body.Close()
		return
	}
	io.Copy(w, transformAppResponse.Body)
	transformAppResponse.Body.Close()
	h.metrics.recordResponseEvent()
}

func (h ContentHandler) getNativeContent(ctx context.Context, w http.ResponseWriter) (ok bool, resp *http.Response) {
	uuid := ctx.Value(uuidKey).(string)
	requestUrl := fmt.Sprintf("%s%s", h.serviceConfig.nativeContentAppUri, uuid)
	transactionId, _ := tid.GetTransactionIDFromContext(ctx)
	h.log.RequestEvent(h.serviceConfig.sourceAppName, requestUrl, transactionId, uuid)
	req, err := http.NewRequest("GET", requestUrl, nil)
	req.Header.Set(tid.TransactionIDHeader, transactionId)
	req.Header.Set("Authorization", "Basic "+h.serviceConfig.nativeContentAppAuth)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)

	return h.handleResponse(req, resp, err, w, uuid)
}

func (h ContentHandler) getTransformedContent(ctx context.Context, nativeContentSourceAppResponse http.Response, w http.ResponseWriter) (ok bool, resp *http.Response) {
	uuid := ctx.Value(uuidKey).(string)
	requestUrl := fmt.Sprintf("%s%s?preview=true", h.serviceConfig.transformAppUri, uuid)
	transactionId, _ := tid.GetTransactionIDFromContext(ctx)

	//TODO we need to assert that resp.Header.Get(tid.TransactionIDHeader) ==  transactionId
	//to ensure that we are logging exactly what is actually passed around in the headers
	h.log.RequestEvent(h.serviceConfig.transformAppName, requestUrl, transactionId, uuid)

	req, err := http.NewRequest("POST", requestUrl, nativeContentSourceAppResponse.Body)
	req.Host = h.serviceConfig.transformAppHostHeader
	req.Header.Set(tid.TransactionIDHeader, transactionId)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)

	return h.handleResponse(req, resp, err, w, uuid)

}

func (h ContentHandler) handleResponse(req *http.Request, extResp *http.Response, err error, w http.ResponseWriter, uuid string) (ok bool, resp *http.Response) {
	//this happens when hostname cannot be resolved or host is not accessible
	if err != nil {
		h.handleError(w, err, h.serviceConfig.transformAppName, req.URL.String(), req.Header.Get(tid.TransactionIDHeader), uuid)
		return false, nil
	}
	switch extResp.StatusCode {
	case http.StatusOK:
		h.log.ResponseEvent(h.serviceConfig.transformAppName, req.URL.String(), extResp, uuid)
		return true, extResp
	case http.StatusNotFound:
		h.handleNotFound(w, extResp, h.serviceConfig.transformAppName, req.URL.String(), uuid)
		return false, nil
	default:
		h.handleFailedRequest(w, extResp, h.serviceConfig.transformAppName, req.URL.String(), uuid)
		return false, nil
	}
}

func (h ContentHandler) handleError(w http.ResponseWriter, err error, serviceName string, url string, transactionId string, uuid string) {
	w.WriteHeader(http.StatusServiceUnavailable)
	h.log.ErrorEvent(serviceName, url, transactionId, err, uuid)
	h.metrics.recordErrorEvent()
}

func (h ContentHandler) handleFailedRequest(w http.ResponseWriter, resp *http.Response, serviceName string, url string, uuid string) {
	w.WriteHeader(http.StatusServiceUnavailable)
	h.log.RequestFailedEvent(serviceName, url, resp, uuid)
	h.metrics.recordRequestFailedEvent()
}

func (h ContentHandler) handleNotFound(w http.ResponseWriter, resp *http.Response, serviceName string, url string, uuid string) {
	w.WriteHeader(http.StatusNotFound)
	h.log.RequestFailedEvent(serviceName, url, resp, uuid)
	h.metrics.recordRequestFailedEvent()
}

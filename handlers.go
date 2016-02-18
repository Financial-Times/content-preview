package main

import (
	"fmt"
	"github.com/gorilla/mux"
	tid "github.com/Financial-Times/transactionid-utils-go"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"github.com/rcrowley/go-metrics"
)

const uuidKey = "uuid"

type Handlers struct {
	serviceConfig *ServiceConfig
	log           *AppLogger
}


func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func buildInfoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "build-info")
}

type ContentHandler struct {
	serviceConfig *ServiceConfig
	log           *AppLogger
}
func (h ContentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	h.log.TransactionStartedEvent(r.RequestURI, tid.GetTransactionIDFromRequest(r), uuid)

	ctx := tid.TransactionAwareContext(context.Background(), r.Header.Get(tid.TransactionIDHeader))
	ctx = context.WithValue(ctx, uuidKey, uuid)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	success, nativeContentSourceAppResponse := h.getNativeContent(ctx, w)
	if !success { return }
	success, transformAppResponse := h.getTransformedContent(ctx, *nativeContentSourceAppResponse, w)
	if(!success) {
		nativeContentSourceAppResponse.Body.Close()
		return
	}

	io.Copy(w, transformAppResponse.Body)
	transformAppResponse.Body.Close()
}

func ( h ContentHandler) getNativeContent(ctx context.Context, w http.ResponseWriter) (ok bool, resp *http.Response) {
	uuid := ctx.Value(uuidKey).(string)
	requestUrl := fmt.Sprintf("%s%s", h.serviceConfig.nativeContentAppUri, uuid)

	transactionId, _ := tid.GetTransactionIDFromContext(ctx)
	h.log.RequestEvent(h.serviceConfig.sourceAppName, requestUrl, transactionId, uuid)

	req, err := http.NewRequest("GET", requestUrl, nil)

	req.Header.Set(tid.TransactionIDHeader, transactionId)
	req.Header.Set("Authorization", "Basic " + h.serviceConfig.nativeContentAppAuth)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)

	//this happens when hostname cannot be resolved or host is not accessible
	if err !=nil {
		h.log.ErrorEvent(h.serviceConfig.sourceAppName, req.URL.String(), req.Header.Get(tid.TransactionIDHeader), err, uuid)
		w.WriteHeader(http.StatusServiceUnavailable)
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusNotFound);
		h.log.RequestFailedEvent(h.serviceConfig.sourceAppName, req.URL.String(), resp, uuid)
		return false, nil
	}
	h.log.ResponseEvent(h.serviceConfig.sourceAppName, req.URL.String(), resp, uuid)
	return true, resp
}

func ( h ContentHandler) getTransformedContent(ctx context.Context, nativeContentSourceAppResponse http.Response, w http.ResponseWriter) (ok bool, resp *http.Response) {
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

	//this happens when hostname cannot be resolved or host is not accessible
	if err !=nil {
		h.log.ErrorEvent(h.serviceConfig.transformAppName, req.URL.String(), req.Header.Get(tid.TransactionIDHeader), err, uuid)
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {

		w.WriteHeader(http.StatusNotFound);
		h.log.RequestFailedEvent(h.serviceConfig.transformAppName, req.URL.String(), resp, uuid)
		return false, nil
	}

	h.log.ResponseEvent(h.serviceConfig.transformAppName, req.URL.String(), resp, uuid)
	return true, resp
}

//Metrics
func metricsHttpEndpoint(w http.ResponseWriter, r *http.Request) {
	metrics.WriteOnce(metrics.DefaultRegistry, w)
}





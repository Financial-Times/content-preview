package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	tid "github.com/Financial-Times/transactionid-utils-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var contentPreviewService *httptest.Server
var methodeApiMock *httptest.Server
var methodeArticleTransformerMock *httptest.Server

func startMethodeApiMock(status string) {
	r := mux.NewRouter()
	if status == "happy" {
		r.Path("/eom-file/{uuid}").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(methodeApiHandlerMock)})
	} else {
		r.Path("/eom-file/{uuid}").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(unhappyHandler)})
	}
	methodeApiMock = httptest.NewServer(r)
}

func methodeApiHandlerMock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if r.Header.Get(tid.TransactionIDHeader) != "" {
		w.Header().Set(tid.TransactionIDHeader, r.Header.Get(tid.TransactionIDHeader))
	} else {
		w.Header().Set(tid.TransactionIDHeader, "tid_w58gqvazux")
	}

	if r.Header.Get("Authorization") == "Basic default" && uuid == "d7db73ec-cf53-11e5-92a1-c5e23ef99c77" {
		file, err := os.Open("test-resources/methode-api-output.json")
		if err != nil {
			return
		}
		defer file.Close()
		io.Copy(w, file)
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("{\"message\":\"404 Not Found - null\"}")
	}

}

func unhappyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func startMethodeArticleTransformerMock(status string) {
	r := mux.NewRouter()
	if status == "happy" {
		r.Path("/content-transform/{uuid}").Queries("preview", "true").Handler(handlers.MethodHandler{"POST": http.HandlerFunc(methodeArticleTransformerHandlerMock)})
	} else {
		r.Path("/content-transform/{uuid}").Handler(handlers.MethodHandler{"POST": http.HandlerFunc(unhappyHandler)})
	}

	methodeArticleTransformerMock = httptest.NewServer(r)
}

func methodeArticleTransformerHandlerMock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if r.Header.Get(tid.TransactionIDHeader) != "" && isEqualToMethodeApiOutput(r.Body) {
		w.Header().Set(tid.TransactionIDHeader, r.Header.Get(tid.TransactionIDHeader))
	} else {
		w.Header().Set(tid.TransactionIDHeader, "tid_w58gqvazux")
	}

	if uuid == "d7db73ec-cf53-11e5-92a1-c5e23ef99c77" {
		file, err := os.Open("test-resources/methode-article-transformer-output.json")
		if err != nil {
			return
		}
		defer file.Close()
		io.Copy(w, file)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

}

func isEqualToMethodeApiOutput(r io.Reader) bool {
	file, err := os.Open("test-resources/methode-api-output.json")
	if err != nil {
		return false
	}
	defer file.Close()
	methodeApiOutput := getStringFromReader(file)
	readerOutput := getStringFromReader(r)
	return methodeApiOutput == readerOutput
}

func stopServices() {
	methodeApiMock.Close()
	methodeArticleTransformerMock.Close()
	contentPreviewService.Close()
}

func startContentPreviewService() {

	methodeApiUrl := methodeApiMock.URL + "/eom-file/"
	nativeContentAppHealthUri := methodeApiMock.URL + "/build-info"
	methodArticleTransformerUrl := methodeArticleTransformerMock.URL + "/content-transform/"
	transformAppHealthUrl := methodeArticleTransformerMock.URL + "/build-info"

	sc := ServiceConfig{
		"content-preview",
		"8084",
		"default",
		"methode-article-transformer",
		methodeApiUrl,
		methodArticleTransformerUrl,
		nativeContentAppHealthUri,
		transformAppHealthUrl,
		"Native Content Service",
		"Native Content Transformer Service",
	}

	appLogger := NewAppLogger()
	metricsHandler := NewMetrics()
	contentHandler := ContentHandler{&sc, appLogger, &metricsHandler}

	h := setupServiceHandler(sc, metricsHandler, contentHandler)

	contentPreviewService = httptest.NewServer(h)
}

func TestShouldReturn200AndTrasformerOutput(t *testing.T) {
	startMethodeApiMock("happy")
	startMethodeArticleTransformerMock("happy")
	startContentPreviewService()
	defer stopServices()
	resp, err := http.Get(contentPreviewService.URL + "/content-preview/d7db73ec-cf53-11e5-92a1-c5e23ef99c77")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Response status should be 200")

	file, _ := os.Open("test-resources/methode-article-transformer-output.json")
	defer file.Close()

	expectedOutput := getStringFromReader(file)
	actualOutput := getStringFromReader(resp.Body)

	assert.Equal(t, expectedOutput, actualOutput, "Response body shoud be equal to transformer response body")
}

func getStringFromReader(r io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	return buf.String()
}

func TestShouldReturn404(t *testing.T) {
	startMethodeApiMock("happy")
	startMethodeArticleTransformerMock("happy")
	startContentPreviewService()
	defer stopServices()

	resp, err := http.Get(contentPreviewService.URL + "/content-preview/158ab514-f989-4abc-b42b-b3edf0811899")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Response status should be 404")

	contentPreviewService.Close()

}

func TestShouldReturn503whenMethodeApiIsNotHappy(t *testing.T) {
	startMethodeApiMock("unhappy")
	startMethodeArticleTransformerMock("happy")
	startContentPreviewService()
	defer stopServices()

	resp, err := http.Get(contentPreviewService.URL + "/content-preview/d7db73ec-cf53-11e5-92a1-c5e23ef99c77")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode, "Response status should be 503")
}

func TestShouldReturn503whenMethodeApiIsNotAvailable(t *testing.T) {
	startMethodeArticleTransformerMock("happy")
	startContentPreviewService()
	defer stopServices()

	resp, err := http.Get(contentPreviewService.URL + "/content-preview/d7db73ec-cf53-11e5-92a1-c5e23ef99c77")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode, "Response status should be 503")
}

func TestShouldReturn503whenTransformerIsNotHappy(t *testing.T) {
	startMethodeApiMock("happy")
	startMethodeArticleTransformerMock("unhappy")
	startContentPreviewService()
	defer stopServices()

	resp, err := http.Get(contentPreviewService.URL + "/content-preview/d7db73ec-cf53-11e5-92a1-c5e23ef99c77")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode, "Response status should be 503")
}

func TestShouldReturn503whenTransformerNotAvailable(t *testing.T) {
	startMethodeApiMock("happy")
	startContentPreviewService()
	defer stopServices()

	resp, err := http.Get(contentPreviewService.URL + "/content-preview/d7db73ec-cf53-11e5-92a1-c5e23ef99c77")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode, "Response status should be 503")
}

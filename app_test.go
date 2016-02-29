package main

import (
	"encoding/json"
	tid "github.com/Financial-Times/transactionid-utils-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var methodeApiMock *httptest.Server
var methodeArticleTransformerMock *httptest.Server

var somePort = rand.Int()%10000 + 40000

func setupServiceMocks() {
	setupMethodeApiMock()
	setupMethodeArticleTransformerMock()
}

func setupMethodeApiMock() {
	r := mux.NewRouter()
	r.Path("/eom-file/{uuid}").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(methodeApiHandlerMock)})
	methodeApiMock = httptest.NewUnstartedServer(r)
	methodeApiMock.Config.Addr = ":" + strconv.Itoa(somePort)
	methodeApiMock.Start()
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

func setupMethodeArticleTransformerMock() {
	r := mux.NewRouter()
	r.Path("/content-transform/{uuid}").Queries("preview", "true").Handler(handlers.MethodHandler{"POST": http.HandlerFunc(methodeArticleTransformerHandlerMock)})
	methodeArticleTransformerMock = httptest.NewUnstartedServer(r)
	methodeArticleTransformerMock.Config.Addr = ":" + strconv.Itoa(somePort+1)
	methodeArticleTransformerMock.Start()
}

func methodeArticleTransformerHandlerMock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if r.Header.Get(tid.TransactionIDHeader) != "" {
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

func stopServiceMocks() {
	methodeApiMock.Close()
	methodeArticleTransformerMock.Close()
}

func TestSomething(t *testing.T) {
	setupServiceMocks()
	assert.Equal(t, "pong", "pong")
	stopServiceMocks()
}

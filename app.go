package main

import (
	"os"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	log "github.com/Sirupsen/logrus"
	"time"
)

const serviceName = "content-preview"
const serviceDescription = "A RESTful API for retrieving and transforming content to preview data"


var timeout = time.Duration(5 * time.Second)
var client = &http.Client{Timeout: timeout}

func main() {

	log.SetLevel(log.InfoLevel)
	log.Infof("%s service started with args %s", serviceName, os.Args)

	app := cli.App(serviceName, serviceDescription)
	appPort := app.StringOpt("app-port", "8084", "Default port for app")
	mapiAuth := app.StringOpt("mapi-auth", "default", "Basic auth for MAPI")
	mapiUri := app.StringOpt("mapi-uri", "http://methode-api-uk-p.svc.ft.com", "Host and port for MAPI")
	matHostHeader := app.StringOpt("mat-host-header", "methode-article-transformer", "Hostheader for MAT")
	matUri := app.StringOpt("mat-uri", "http://ftapp05951-lvpr-uk-int:8080", "Host and port for MAT")

	app.Action = func() {
		r := mux.NewRouter()
		handler := Handlers{*mapiAuth, *matHostHeader, *mapiUri, *matUri}
		r.HandleFunc("/content-preview/{uuid}", handler.contentPreviewHandler)
		r.HandleFunc("/build-info", handler.buildInfoHandler)
		r.HandleFunc("/__health", fthealth.Handler(serviceName, serviceDescription, handler.mapiCheck(), handler.matCheck()))
		r.HandleFunc("/ping", pingHandler)
		http.Handle("/", r)
		log.Fatal(http.ListenAndServe(":"+*appPort, nil))
	}
	log.WithFields(log.Fields{
		"mapiUri" : *mapiUri,
		"matUri"  : *matUri,
	}).Infof("%s service started on localhost:%s with configuration", serviceName, *appPort)
	app.Run(os.Args)

}

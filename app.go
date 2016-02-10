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
const mapiPath = "/eom-file/"
const matPath ="/content-transform/"

var timeout = time.Duration(5 * time.Second)
var client = &http.Client{Timeout: timeout}

func main() {

	log.SetLevel(log.InfoLevel)
	log.Infof("%s service started with args %s", serviceName, os.Args)

	app := cli.App(serviceName, serviceDescription)
	appPort := app.StringOpt("app-port", "8084", "Default port for app")
	mapiAuth := app.StringOpt("mapi-auth", "default", "Basic auth for MAPI")
	mapiHostAndPort := app.StringOpt("mapi-host", "methode-api-uk-p.svc.ft.com", "Host and port for MAPI")
	matHostHeader := app.StringOpt("mat-host-header", "methode-article-transformer", "Hostheader for MAT")
	matHostAndPort := app.StringOpt("mat-host", "ftapp05951-lvpr-uk-int:8080", "Host and port for MAT")

	app.Action = func() {
		r := mux.NewRouter()
		handler := Handlers{*mapiAuth, *matHostHeader, *mapiHostAndPort, *matHostAndPort}
		r.HandleFunc("/content-preview/{uuid}", handler.contentPreviewHandler)
		r.HandleFunc("/build-info", handler.buildInfoHandler)
		r.HandleFunc("/__health", fthealth.Handler(serviceName, serviceDescription, handler.mapiCheck(), handler.matCheck()))
		r.HandleFunc("/ping", pingHandler)
		http.Handle("/", r)
		log.Fatal(http.ListenAndServe(":"+*appPort, nil))
	}
	log.WithFields(log.Fields{
		"mapiHostAndPort" : *mapiHostAndPort,
		"matHostAndPort"  : *matHostAndPort,
	}).Infof("%s service started on localhost:%s with configuration", serviceName, *appPort)
	app.Run(os.Args)

}

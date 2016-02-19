package main

import (
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/rcrowley/go-metrics"
	"net/http"
	"os"
	"time"
)

const serviceName = "content-preview"
const serviceDescription = "A RESTful API for retrieving and transforming content to preview data"

var timeout = time.Duration(5 * time.Second)
var client = &http.Client{Timeout: timeout}

func main() {
	log.SetLevel(log.InfoLevel)

	app := cli.App(serviceName, serviceDescription)
	appPort := app.StringOpt("app-port", "8084", "Default port for Content Preview app")
	nativeContentAppAuth := app.StringOpt("source-app-auth", "default", "Basic auth for MAPI")
	nativeContentAppUri := app.StringOpt("source-app-uri", "http://methode-api-uk-p.svc.ft.com/eom-file/", "URI of the Native Content Source Application endpoint")
	nativeContentAppHealthUri := app.StringOpt("source-app-health-uri", "http://methode-api-uk-p.svc.ft.com/build-info", "URI of the Native Content Source Application health endpoint")
	transformAppHostHeader := app.StringOpt("transform-app-host-header", "methode-article-transformer", "Transform Application Host Header")
	transformAppUri := app.StringOpt("transform-app-uri", "http://ftapp05951-lvpr-uk-int:8080/content-transform/", "URI of the Transform Application endpoint")
	transformAppHealthUri := app.StringOpt("transform-app-health-uri", "http://methode-article-transformer-01-pr-uk-int.svc.ft.com/build-info", "URI of the Transform Application health endpoint")

	app.Action = func() {
		r := mux.NewRouter()
		handler := Handlers{*nativeContentAppAuth, *transformAppHostHeader, *nativeContentAppUri, *transformAppUri, *nativeContentAppHealthUri, *transformAppHealthUri}
		r.Path("/content-preview/{uuid}").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(handler.contentPreviewHandler)})
		r.Path("/build-info").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(handler.buildInfoHandler)})
		r.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(fthealth.Handler(serviceName, serviceDescription, handler.mapiCheck(), handler.matCheck()))})
		r.Path("/__ping").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(handler.pingHandler)})

		log.WithFields(log.Fields{
			"source-app-uri":           *nativeContentAppUri,
			"transform-app-uri":        *transformAppUri,
			"source-app-health-uri":    *nativeContentAppHealthUri,
			"transform-app-health-uri": *transformAppHealthUri,
		}).Infof("%s service started on localhost:%s with configuration", serviceName, *appPort)

		err := http.ListenAndServe(":"+*appPort,
			httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry,
				httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), r)))

		if err != nil {
			log.Fatalf("Unable to start server: %v", err)
		}
	}

	app.Run(os.Args)

}

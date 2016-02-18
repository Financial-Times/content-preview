package main

import (
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/Sirupsen/logrus"
	"time"
	"net/http"
	"os"
)

const serviceDescription = "A RESTful API for retrieving and transforming content to preview data"

var timeout = time.Duration(5 * time.Second)
var client = &http.Client{Timeout: timeout}

func main() {

	app := cli.App("content-preview", serviceDescription)
	serviceName :=app.StringOpt("app-name", "content-preview", "The name of this service")
	appPort := app.StringOpt("app-port", "8084", "Default port for Content Preview app")
	nativeContentAppAuth := app.StringOpt("source-app-auth", "default", "Basic auth for MAPI")
	nativeContentAppUri := app.StringOpt("source-app-uri", "http://methode-api-uk-p.svc.ft.com/eom-file/", "URI of the Native Content Source Application endpoint")
	nativeContentAppHealthUri := app.StringOpt("source-app-health-uri", "http://methode-api-uk-p.svc.ft.com/build-info", "URI of the Native Content Source Application health endpoint")
	transformAppHostHeader := app.StringOpt("transform-app-host-header", "methode-article-transformer", "Transform Application Host Header")
	transformAppUri := app.StringOpt("transform-app-uri", "http://ftapp05951-lvpr-uk-int:8080/content-transform/", "URI of the Transform Application endpoint")
	transformAppHealthUri := app.StringOpt("transform-app-health-uri", "http://methode-article-transformer-01-pr-uk-int.svc.ft.com/build-info", "URI of the Transform Application health endpoint")
	sourceAppName := app.StringOpt("source-app-name", "Native Content Service", "Service name of the source application")
	transformAppName := app.StringOpt("transform-app-name", "Native Content Transformer Service", "Service name of the content transformer application")

	app.Action = func() {
		sc := ServiceConfig {*serviceName, *appPort, *nativeContentAppAuth,
			*transformAppHostHeader, *nativeContentAppUri, *transformAppUri, *nativeContentAppHealthUri, *transformAppHealthUri, *sourceAppName, *transformAppName}
		appLogger := NewAppLogger()
		metricsHandler := NewMetrics()
		contentHandler := ContentHandler{&sc, appLogger, &metricsHandler}

		r := mux.NewRouter()
		r.Path("/content-preview/{uuid}").Handler(handlers.MethodHandler{"GET": httphandlers.HTTPMetricsHandler(metricsHandler.registry,
httphandlers.TransactionAwareRequestLoggingHandler(logrus.StandardLogger(), contentHandler))})
		r.Path("/build-info").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(buildInfoHandler)})
		r.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(fthealth.Handler(*serviceName, serviceDescription, sc.nativeContentSourceCheck(), sc.transformerServiceCheck()))})
		r.Path("/__ping").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(pingHandler)})

		r.HandleFunc("/__metrics", metricsHttpEndpoint)

		appLogger.ServiceStartedEvent(*serviceName, sc.asMap())

		err := http.ListenAndServe( ":"+*appPort, r)
		if err != nil {
			logrus.Fatalf("Unable to start server: %v", err)
		}
	}
	app.Run(os.Args)
}

type ServiceConfig struct {
	serviceName 			  string
	appPort					  string
	nativeContentAppAuth      string
	transformAppHostHeader    string
	nativeContentAppUri       string
	transformAppUri           string
	nativeContentAppHealthUri string
	transformAppHealthUri     string
	sourceAppName			  string
	transformAppName 	      string
}

func (sc ServiceConfig) asMap () map[string] interface{} {

	return map[string] interface{} {
		"service-name" : sc.serviceName,
		"service-port" : sc.appPort,
		"source-app-name" : sc.sourceAppName,
		"source-app-uri" : sc.nativeContentAppUri,
		"transform-app-name" : sc.transformAppName,
		"transform-app-uri"  : sc.transformAppUri,
		"source-app-health-uri" : sc.nativeContentAppHealthUri,
		"transform-app-health-uri" : sc.transformAppHealthUri,
	}
}

package main

import (
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	oldhttphandlers "github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"net/http"
	"os"
	"time"
)

const serviceDescription = "A RESTful API for retrieving and transforming content to preview data"

var timeout = time.Duration(10 * time.Second)
var client = &http.Client{Timeout: timeout}

func main() {

	app := cli.App("content-preview", serviceDescription)
	serviceName := app.StringOpt("app-name", "content-preview", "The name of this service")
	appPort := app.StringOpt("app-port", "8084", "Default port for Content Preview app")
	nativeContentAppAuth := app.StringOpt("source-app-auth", "default", "Basic auth for MAPI")
	nativeContentAppUri := app.StringOpt("source-app-uri", "http://methode-api-uk-p.svc.ft.com/eom-file/", "URI of the Native Content Source Application endpoint")
	nativeContentAppHealthUri := app.StringOpt("source-app-health-uri", "http://methode-api-uk-p.svc.ft.com/build-info", "URI of the Native Content Source Application health endpoint")
	transformAppHostHeader := app.StringOpt("transform-app-host-header", "methode-article-transformer", "Transform Application Host Header")
	transformAppUri := app.StringOpt("transform-app-uri", "http://methode-article-transformer-01-iw-uk-p.svc.ft.com/content-transform/", "URI of the Transform Application endpoint")
	transformAppHealthUri := app.StringOpt("transform-app-health-uri", "http://methode-article-transformer-01-iw-uk-p.svc.ft.com/build-info", "URI of the Transform Application health endpoint")
	sourceAppName := app.StringOpt("source-app-name", "Native Content Service", "Service name of the source application")
	transformAppName := app.StringOpt("transform-app-name", "Native Content Transformer Service", "Service name of the content transformer application")
	sourceAppPanicGuide := app.String(cli.StringOpt{
		Name:   "source-app-panic-guide",
		Value:  "https://sites.google.com/a/ft.com/dynamic-publishing-team/content-preview-panic-guide",
		Desc:   "Native Content Source application panic guide url for healthcheck. Default panic guide is for content preview.",
		EnvVar: "SOURCE_APP_PANIC_GUIDE",
	})
	transformAppPanicGuide := app.String(cli.StringOpt{
		Name:   "transform-app-panic-guide",
		Value:  "https://sites.google.com/a/ft.com/dynamic-publishing-team/content-preview-panic-guide",
		Desc:   "Transform application panic guide url for healthcheck. Default panic guide is for content preview.",
		EnvVar: "TRANSFORM_APP_PANIC_GUIDE",
	})

	graphiteTCPAddress := app.String(cli.StringOpt{
		Name:   "graphite-tcp-address",
		Value:  "",
		Desc:   "Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally)",
		EnvVar: "GRAPHITE_TCP_ADDRESS",
	})
	graphitePrefix := app.String(cli.StringOpt{
		Name:   "graphite-prefix",
		Value:  "coco.services.$ENV.content-preview.0",
		Desc:   "Prefix to use. Should start with content, include the environment, and the host name. e.g. coco.pre-prod.sections-rw-neo4j.1",
		EnvVar: "GRAPHITE_PREFIX",
	})

	logMetrics := app.Bool(cli.BoolOpt{
		Name:   "log-metrics",
		Value:  false,
		Desc:   "Whether to log metrics. Set to true if running locally and you want metrics output",
		EnvVar: "LOG_METRICS",
	})

	app.Action = func() {
		sc := ServiceConfig{
			*serviceName,
			*appPort,
			*nativeContentAppAuth,
			*transformAppHostHeader,
			*nativeContentAppUri,
			*transformAppUri,
			*nativeContentAppHealthUri,
			*transformAppHealthUri,
			*sourceAppName,
			*transformAppName,
			*sourceAppPanicGuide,
			*transformAppPanicGuide,
			*graphiteTCPAddress,
			*graphitePrefix,
		}
		appLogger := NewAppLogger()
		metricsHandler := NewMetrics()
		contentHandler := ContentHandler{&sc, appLogger, &metricsHandler}

		h := setupServiceHandler(sc, metricsHandler, contentHandler)

		appLogger.ServiceStartedEvent(*serviceName, sc.asMap())
		metricsHandler.OutputMetricsIfRequired(*graphiteTCPAddress, *graphitePrefix, *logMetrics)

		err := http.ListenAndServe(":"+*appPort, h)

		if err != nil {
			logrus.Fatalf("Unable to start server: %v", err)
		}
	}
	app.Run(os.Args)
}

func setupServiceHandler(sc ServiceConfig, metricsHandler Metrics, contentHandler ContentHandler) *mux.Router {
	r := mux.NewRouter()
	r.Path("/content-preview/{uuid}").Handler(handlers.MethodHandler{"GET": oldhttphandlers.HTTPMetricsHandler(metricsHandler.registry,
		oldhttphandlers.TransactionAwareRequestLoggingHandler(logrus.StandardLogger(), contentHandler))})

	r.Path(httphandlers.BuildInfoPath).HandlerFunc(httphandlers.BuildInfoHandler)
	r.Path(httphandlers.PingPath).HandlerFunc(httphandlers.PingHandler)

	r.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(fthealth.Handler(sc.serviceName, serviceDescription, sc.nativeContentSourceCheck(), sc.transformerServiceCheck()))})
	r.Path("/__metrics").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(metricsHttpEndpoint)})

	return r
}

type ServiceConfig struct {
	serviceName               string
	appPort                   string
	nativeContentAppAuth      string
	transformAppHostHeader    string
	nativeContentAppUri       string
	transformAppUri           string
	nativeContentAppHealthUri string
	transformAppHealthUri     string
	sourceAppName             string
	transformAppName          string
	sourceAppPanicGuide       string
	transformAppPanicGuide    string
	graphiteTCPAddress        string
	graphitePrefix            string
}

func (sc ServiceConfig) asMap() map[string]interface{} {

	return map[string]interface{}{
		"service-name":              sc.serviceName,
		"service-port":              sc.appPort,
		"source-app-name":           sc.sourceAppName,
		"source-app-uri":            sc.nativeContentAppUri,
		"transform-app-name":        sc.transformAppName,
		"transform-app-uri":         sc.transformAppUri,
		"source-app-health-uri":     sc.nativeContentAppHealthUri,
		"transform-app-health-uri":  sc.transformAppHealthUri,
		"source-app-panic-guide":    sc.sourceAppPanicGuide,
		"transform-app-panic-guide": sc.transformAppPanicGuide,
		"graphite-tcp-address":      sc.graphiteTCPAddress,
		"graphite-prefix":           sc.graphitePrefix,
	}
}

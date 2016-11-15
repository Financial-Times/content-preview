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

var globalLogger *logrus.Logger

func main() {

	app := cli.App("content-preview", serviceDescription)
	serviceName := app.StringOpt("app-name", "content-preview", "The name of this service")
	appPort := app.String(cli.StringOpt{
		Name:   "app-port",
		Value:  "8084",
		Desc:   "Default port for Content Preview app",
		EnvVar: "APP_PORT",
	})
	sourceAppAuth := app.String(cli.StringOpt{
		Name:   "source-app-auth",
		Value:  "default",
		Desc:   "Basic auth for MAPI",
		EnvVar: "SOURCE_APP_AUTH",
	})
	sourceAppUri := app.String(cli.StringOpt{
		Name:   "source-app-uri",
		Value:  "http://methode-api-uk-p.svc.ft.com/eom-file/",
		Desc:   "URI of the Native Content Source Application endpoint",
		EnvVar: "SOURCE_APP_URI",
	})
	sourceAppAppHealthUri := app.String(cli.StringOpt{
		Name:   "source-app-health-uri",
		Value:  "http://methode-api-uk-p.svc.ft.com/build-info",
		Desc:   "URI of the Native Content Source Application health endpoint",
		EnvVar: "SOURCE_APP_HEALTH_URI",
	})
	transformAppHostHeader := app.String(cli.StringOpt{
		Name:   "transform-app-host-header",
		Value:  "methode-article-transformer",
		Desc:   "Transform Application Host Header",
		EnvVar: "TRANSFORM_APP_HOST_HEADER",
	})
	transformAppUri := app.String(cli.StringOpt{
		Name:   "transform-app-uri",
		Value:  "http://methode-article-transformer-01-iw-uk-p.svc.ft.com/content-transform/",
		Desc:   "URI of the Transform Application endpoint",
		EnvVar: "TRANSFORM_APP_URI",
	})
	transformAppHealthUri := app.String(cli.StringOpt{
		Name:   "transform-app-health-uri",
		Value:  "http://methode-article-transformer-01-iw-uk-p.svc.ft.com/build-info",
		Desc:   "URI of the Transform Application health endpoint",
		EnvVar: "TRANSFORM_APP_HEALTH_URI",
	})
	sourceAppName := app.String(cli.StringOpt{
		Name:   "source-app-name",
		Value:  "Native Content Service",
		Desc:   "Service name of the source application",
		EnvVar: "SOURCE_APP_NAME",
	})
	transformAppName := app.String(cli.StringOpt{
		Name:   "transform-app-name",
		Value:  "Native Content Transformer Service",
		Desc:   "Service name of the content transformer application",
		EnvVar: "TRANSFORM_APP_NAME",
	})
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
			*sourceAppAuth,
			*transformAppHostHeader,
			*sourceAppUri,
			*transformAppUri,
			*sourceAppAppHealthUri,
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
		globalLogger = appLogger.log
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
	serviceName            string
	appPort                string
	sourceAppAuth          string
	transformAppHostHeader string
	sourceAppAppUri        string
	transformAppUri        string
	sourceAppHealthUri     string
	transformAppHealthUri  string
	sourceAppName          string
	transformAppName       string
	sourceAppPanicGuide    string
	transformAppPanicGuide string
	graphiteTCPAddress     string
	graphitePrefix         string
}

func (sc ServiceConfig) asMap() map[string]interface{} {
	return map[string]interface{}{
		"service-name":              sc.serviceName,
		"service-port":              sc.appPort,
		"source-app-name":           sc.sourceAppName,
		"source-app-uri":            sc.sourceAppAppUri,
		"transform-app-name":        sc.transformAppName,
		"transform-app-uri":         sc.transformAppUri,
		"source-app-health-uri":     sc.sourceAppHealthUri,
		"transform-app-health-uri":  sc.transformAppHealthUri,
		"source-app-panic-guide":    sc.sourceAppPanicGuide,
		"transform-app-panic-guide": sc.transformAppPanicGuide,
		"graphite-tcp-address":      sc.graphiteTCPAddress,
		"graphite-prefix":           sc.graphitePrefix,
	}
}

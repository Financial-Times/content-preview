package main

import (
	"net/http"
	"os"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	oldhttphandlers "github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/sirupsen/logrus"
)

const serviceDescription = "A RESTful API for retrieving and transforming content to preview data"

var timeout = time.Duration(10 * time.Second)
var client = &http.Client{Timeout: timeout}

func main() {
	app := cli.App("content-preview", serviceDescription)
	appSystemCode := app.String(cli.StringOpt{
		Name:   "app-system-code",
		Value:  "content-preview",
		Desc:   "The system code of this service",
		EnvVar: "APP_SYSTEM_CODE",
	})
	appName := app.String(cli.StringOpt{
		Name:   "app-name",
		Value:  "Content Preview",
		Desc:   "The name of this service",
		EnvVar: "APP_NAME",
	})
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
	sourceAppHealthUri := app.String(cli.StringOpt{
		Name:   "source-app-health-uri",
		Value:  "http://methode-api-uk-p.svc.ft.com/build-info",
		Desc:   "URI of the Native Content Source Application health endpoint",
		EnvVar: "SOURCE_APP_HEALTH_URI",
	})
	transformAppUri := app.String(cli.StringOpt{
		Name:   "transform-app-uri",
		Value:  "http://localhost:8080/__methode-article-transformer/map/",
		Desc:   "URI of the Transform Application endpoint",
		EnvVar: "TRANSFORM_APP_URI",
	})
	transformAppHealthUri := app.String(cli.StringOpt{
		Name:   "transform-app-health-uri",
		Value:  "http://localhost:8080/__methode-article-transformer/build-info",
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
		Value:  "https://dewey.ft.com/up-mapi.html",
		Desc:   "Native Content Source application panic guide url for healthcheck. Default panic guide is for Methode API.",
		EnvVar: "SOURCE_APP_PANIC_GUIDE",
	})
	transformAppPanicGuide := app.String(cli.StringOpt{
		Name:   "transform-app-panic-guide",
		Value:  "https://dewey.ft.com/up-mam.html",
		Desc:   "Transform application panic guide url for healthcheck. Default panic guide is for Methode Article Mapper.",
		EnvVar: "TRANSFORM_APP_PANIC_GUIDE",
	})
	businessImpact := app.String(cli.StringOpt{
		Name:   "business-impact",
		Value:  "Editorial users won't be able to preview articles",
		Desc:   "Describe the business impact the dependent services would produce if one is broken.",
		EnvVar: "BUSINESS_IMPACT",
	})
	app.Action = func() {
		sc := ServiceConfig{
			appSystemCode:          *appSystemCode,
			appName:                *appName,
			appPort:                *appPort,
			sourceAppName:          *sourceAppName,
			sourceAppAuth:          *sourceAppAuth,
			sourceAppUri:           *sourceAppUri,
			sourceAppHealthUri:     *sourceAppHealthUri,
			sourceAppPanicGuide:    *sourceAppPanicGuide,
			transformAppName:       *transformAppName,
			transformAppUri:        *transformAppUri,
			transformAppHealthUri:  *transformAppHealthUri,
			transformAppPanicGuide: *transformAppPanicGuide,
			businessImpact:         *businessImpact,
		}

		appLogger := NewAppLogger()
		metricsHandler := NewMetrics()
		contentHandler := ContentHandler{&sc, appLogger, &metricsHandler}
		h := setupServiceHandler(sc, metricsHandler, contentHandler)
		appLogger.ServiceStartedEvent(*appSystemCode, sc.asMap())
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

	gtgHandler := httphandlers.NewGoodToGoHandler(sc.gtgCheck)
	r.Path(httphandlers.GTGPath).HandlerFunc(gtgHandler)

	hc := fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{SystemCode: sc.appSystemCode, Description: serviceDescription, Name: sc.appName, Checks: []fthealth.Check{sc.nativeContentSourceCheck(), sc.transformerServiceCheck()}},
		Timeout:     10 * time.Second,
	}
	r.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(fthealth.Handler(&hc))})

	r.Path("/__metrics").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(metricsHttpEndpoint)})
	return r
}

type ServiceConfig struct {
	appSystemCode          string
	appName                string
	appPort                string
	sourceAppName          string
	sourceAppAuth          string
	sourceAppUri           string
	sourceAppHealthUri     string
	sourceAppPanicGuide    string
	transformAppName       string
	transformAppUri        string
	transformAppHealthUri  string
	transformAppPanicGuide string
	businessImpact         string
}

func (sc ServiceConfig) asMap() map[string]interface{} {
	return map[string]interface{}{
		"app-system-code":           sc.appSystemCode,
		"app-name":                  sc.appName,
		"app-port":                  sc.appPort,
		"source-app-name":           sc.sourceAppName,
		"source-app-uri":            sc.sourceAppUri,
		"transform-app-name":        sc.transformAppName,
		"transform-app-uri":         sc.transformAppUri,
		"source-app-health-uri":     sc.sourceAppHealthUri,
		"transform-app-health-uri":  sc.transformAppHealthUri,
		"source-app-panic-guide":    sc.sourceAppPanicGuide,
		"transform-app-panic-guide": sc.transformAppPanicGuide,
		"business-impact":           sc.businessImpact,
	}
}

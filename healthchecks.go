package main

import (
	"errors"
	"fmt"
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/service-status-go/gtg"
	"net/http"
)

func (sc *ServiceConfig) nativeContentSourceCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Editorial users won't be able to preview articles",
		Name:             sc.sourceAppName + " Availabililty Check",
		PanicGuide:       sc.sourceAppPanicGuide,
		Severity:         1,
		TechnicalSummary: "Checks that " + sc.sourceAppName + " Service is reachable. Article Preview Service requests native content from " + sc.sourceAppName + " service.",
		Checker: func() (string, error) {
			return checkServiceAvailability(sc.sourceAppName, sc.sourceAppHealthUri, sc.sourceAppAuth)
		},
	}
}

func (sc *ServiceConfig) transformerServiceCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Editorial users won't be able to preview articles",
		Name:             sc.transformAppName + " Availabililty Check",
		PanicGuide:       sc.transformAppPanicGuide,
		Severity:         1,
		TechnicalSummary: "Checks that " + sc.transformAppName + " Service is reachable. Article Preview Service relies on " + sc.transformAppName + " service to process content.",
		Checker: func() (string, error) {
			return checkServiceAvailability(sc.transformAppName, sc.transformAppHealthUri, "")
		},
	}
}

func (sc *ServiceConfig) gtgCheck() gtg.Status {
	msg, err := checkServiceAvailability(sc.sourceAppName, sc.sourceAppHealthUri, sc.sourceAppAuth)
	if err != nil {
		return gtg.Status{GoodToGo:false, Message:msg}
	}

	msg, err = checkServiceAvailability(sc.transformAppName, sc.transformAppHealthUri, "")
	if err != nil {
		return gtg.Status{GoodToGo:false, Message:msg}
	}

	return gtg.Status{GoodToGo:true}
}

func checkServiceAvailability(serviceName string, healthUri string, auth string) (string, error) {
	req, err := http.NewRequest("GET", healthUri, nil)
	if auth != "" {
		req.Header.Set("Authorization", "Basic " + auth)
	}
	resp, err := client.Do(req)
	if err != nil {
		msg := fmt.Sprintf("%s service is unreachable: %v", serviceName, err)
		return msg, errors.New(msg)
	}
	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("%s service is not responding with OK. status=%d", serviceName, resp.StatusCode)
		return msg, errors.New(msg)
	}
	return "Ok", nil
}

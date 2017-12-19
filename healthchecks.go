package main

import (
	"errors"
	"fmt"
	"net/http"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/service-status-go/gtg"
)

func (sc *ServiceConfig) nativeContentSourceCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   sc.businessImpact,
		Name:             sc.sourceAppName,
		PanicGuide:       sc.sourceAppPanicGuide,
		Severity:         1,
		TechnicalSummary: "Checks that " + sc.sourceAppName + " is reachable. " + sc.appName + " requests native content from " + sc.sourceAppName,
		Checker: func() (string, error) {
			return checkServiceAvailability(sc.sourceAppName, sc.sourceAppHealthUri, sc.sourceAppAuth)
		},
	}
}

func (sc *ServiceConfig) transformerServiceCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   sc.businessImpact,
		Name:             sc.transformAppName,
		PanicGuide:       sc.transformAppPanicGuide,
		Severity:         1,
		TechnicalSummary: "Checks that " + sc.transformAppName + " is reachable. " + sc.appName + " relies on " + sc.transformAppName + " to process content",
		Checker: func() (string, error) {
			return checkServiceAvailability(sc.transformAppName, sc.transformAppHealthUri, "")
		},
	}
}

func (sc *ServiceConfig) gtgCheck() gtg.Status {
	sourceCheck := func() gtg.Status {
		msg, err := checkServiceAvailability(sc.sourceAppName, sc.sourceAppHealthUri, sc.sourceAppAuth)
		if err != nil {
			return gtg.Status{GoodToGo: false, Message: msg}
		}
		return gtg.Status{GoodToGo: true}
	}
	transformerCheck := func() gtg.Status {
		msg, err := checkServiceAvailability(sc.transformAppName, sc.transformAppHealthUri, "")
		if err != nil {
			return gtg.Status{GoodToGo: false, Message: msg}
		}

		return gtg.Status{GoodToGo: true}
	}

	return gtg.FailFastParallelCheck([]gtg.StatusChecker{
		sourceCheck,
		transformerCheck,
	})()
}

func checkServiceAvailability(serviceName string, healthUri string, auth string) (string, error) {
	req, err := http.NewRequest("GET", healthUri, nil)
	if auth != "" {
		req.Header.Set("Authorization", "Basic "+auth)
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

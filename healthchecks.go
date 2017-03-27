package main

import (
	"errors"
	"fmt"
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"net/http"
)

func (sc *ServiceConfig) nativeContentSourceCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   sc.businessImpact,
		Name:             sc.sourceAppName,
		PanicGuide:       sc.sourceAppPanicGuide,
		Severity:         1,
		TechnicalSummary: "Checks that " + sc.sourceAppName + " is reachable. " + sc.appName + " requests native content from " + sc.sourceAppName,
		Checker: func() (string, error) {
			return checkServiceAvailability(sc.sourceAppName, sc.sourceAppHealthUri, sc.sourceAppAuth, "")
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
			return checkServiceAvailability(sc.transformAppName, sc.transformAppHealthUri, "", sc.transformAppHostHeader)
		},
	}
}

func checkServiceAvailability(serviceName string, healthUri string, auth string, hostHeader string) (string, error) {
	req, err := http.NewRequest("GET", healthUri, nil)
	if auth != "" {
		req.Header.Set("Authorization", "Basic "+auth)
	}
	if hostHeader != "" {
		req.Host = hostHeader
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

package main

import (
	"fmt"
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
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
			return checkServiceAvailability(sc.sourceAppName, sc.nativeContentAppHealthUri, sc.nativeContentAppAuth, "")
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
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("%s service is unreachable", serviceName), fmt.Errorf("%s service is unreachable", serviceName)
	}
	return "Ok", nil
}

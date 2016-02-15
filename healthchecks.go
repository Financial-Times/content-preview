package main
import (
	"net/http"
	"fmt"
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
)

func (h Handlers) mapiCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Editorial users won't be able to preview articles",
		Name:             "Methode Api Availabililty Check",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: "Checks that Methode API Service is reachable. Article Preview Service requests native content from Methode API service.",
		Checker:          func() (string, error) {
			return checkServiceAvailability("Methode API", h.mapiUri, h.mapiAuth)
		},
	}
}

func (h Handlers) matCheck() fthealth.Check {
	return fthealth.Check {
		BusinessImpact:   "Editorial users won't be able to preview articles",
		Name:             "Methode Article Transformer Availabililty Check",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: "Checks that Methode Article Transformer Service is reachable. Article Preview Service relies on Methode Article Transformer service to process content.",
		Checker:          func() (string, error) {
			return checkServiceAvailability("Methode Article Transformer", h.matUri, "")
		},
	}
}

func checkServiceAvailability(serviceName string, host string, auth string) (string, error) {
	url := fmt.Sprintf("%s/build-info", host)
	req, err := http.NewRequest("GET", url, nil)
	if auth != "" {
	req.Header.Set("Authorization", "Basic " + auth)
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("%s service is unreachable", serviceName), fmt.Errorf("%s service is unreachable", serviceName)
	}
	return "Ok", nil
}

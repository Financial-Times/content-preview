package main
import (
	"net/http"
	"fmt"
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
)

func (h Handlers) mapiCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Articvle Preview Service will not work",
		Name:             "Methode Api Availablilty Check",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: "Checks that Methode API Service is reachable. Article Preview Service requests native content from Methode API service.",
		Checker:          func() (string, error) {
			return checkServiceAvailablity("Methode API", h.mapiHost, h.mapiAuth)
		},
	}
}

func (h Handlers) matCheck() fthealth.Check {
	return fthealth.Check {
		BusinessImpact:   "Article Peview service will not work",
		Name:             "Mehtode Article Transformer Availablilty Check",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: "Checks that Methode Article Transformer Service is reachable. Article Preview Service relies on Methode Article Transformer service to process content.",
		Checker:          func() (string, error) {
			return checkServiceAvailablity("Methode Article Transformer", h.matHost, "")
		},
	}
}

func checkServiceAvailablity(serviceName string, host string, auth string) (string, error) {
	url := fmt.Sprintf("http://%s/build-info", host)
	req, err := http.NewRequest("GET", url, nil)
	if auth != "" {
	req.Header.Set("Authorization", "Basic " + auth)
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return fmt.Sprintf("%s is unreachable", serviceName), err
	}
	return "Ok", nil
}

func checkMethodeArticleTransformerAvailablity(host string) (string, error){
	url := fmt.Sprintf("http://%s/build-info", host)
	mapReq, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(mapReq)
	if resp.StatusCode == 200 {
		return "Ok", nil
	}
	return fmt.Sprintf("Methode Article Trransformer is unreachable"), err
}
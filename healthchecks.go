package main
import (
	"net/http"
	"fmt"
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
)

func (h Handlers) matCheck() fthealth.Check {
	return fthealth.Check {
		BusinessImpact:   "Article Peview service will not work",
		Name:             "Mehtode Article Transformer Availablilty Check",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: "Checks that Methode Article Transformer Service is reachable. Article Preview Service relies on Methode Article Transformer service to process content.",
		Checker:          func() (string, error) {
			return checkMethodeArticleTransformerAvailablity(h.matHost)
		},
	}
}

func checkMehtodeApiAvailablity(host string, mapiAuth string) (string, error) {
	url := fmt.Sprintf("http://%s/build-info", host)
	client := &http.Client{}
	mapReq, err := http.NewRequest("GET", url, nil)
	mapReq.Header.Set("Authorization", "Basic " + mapiAuth)
	resp, err := client.Do(mapReq)
	if resp.StatusCode == 200 {
		return "Ok", nil
	}
	return fmt.Sprintf("Methode API respnded with code %s", resp.Status), err
}

func checkMethodeArticleTransformerAvailablity(host string) (string, error){
	url := fmt.Sprintf("http://%s/build-info", host)
	client := &http.Client{}
	mapReq, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(mapReq)
	if resp.StatusCode == 200 {
		return "Ok", nil
	}
	return fmt.Sprintf("Methode Article Trransformer respnded with code %s", resp.Status), err
}
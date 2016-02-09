package main

import (
	"io"
	"net/http"
	"os"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	log "github.com/Sirupsen/logrus"
)

const serviceName = "content-preview"
const serviceDescription = "A RESTful API for retrieving and transforming content to preview data"
const mapiPath = "/eom-file/"
const matPath ="/content-transform/"

func main() {
	log.SetLevel(log.InfoLevel)
	log.Infof("%s service started with args %s", serviceName, os.Args)

	app := cli.App(serviceName, serviceDescription)
	appPort := app.StringOpt("app-port", "8084", "Default port for app")
	mapiAuth := app.StringOpt("mapi-auth", "default", "Basic auth for MAPI")
	mapiHostAndPort := app.StringOpt("mapi-host", "methode-api-uk-p.svc.ft.com", "Host and port for MAPI")
	matHostHeader := app.StringOpt("mat-host-header", "methode-article-transformer", "Hostheader for MAT")
	matHostAndPort := app.StringOpt("mat-host", "ftapp05951-lvpr-uk-int:8080", "Host and port for MAT")

	app.Action = func() {
		r := mux.NewRouter()
		handler := Handlers{*mapiAuth, *matHostHeader, *mapiHostAndPort, *matHostAndPort}
		r.HandleFunc("/content-preview/{uuid}", handler.contentPreviewHandler)
		r.HandleFunc("/build-info", handler.buildInfoHandler)
		r.HandleFunc("/__health", fthealth.Handler(serviceName, serviceDescription, handler.mapiCheck(), handler.matCheck()))
		r.HandleFunc("/ping", pingHandler)
		http.Handle("/", r)
		log.Fatal(http.ListenAndServe(":"+*appPort, nil))

	}
	log.Infof("%s service started on port %s", serviceName, *appPort)
	app.Run(os.Args)

}

type Handlers struct {
	mapiAuth string
	matHostHeader string
	mapiHost string
	matHost string
}

func pingHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "pong")
}

func (h Handlers) buildInfoHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "build-info")
}

var client = &http.Client{}

func (h Handlers) contentPreviewHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request")

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		log.Fatal("missing UUID")
	}

	methode := fmt.Sprintf("http://%s%s%s", h.mapiHost, mapiPath, uuid)

	log.Printf("sending to MAPI at %v\n" + methode)

	mapReq, err := http.NewRequest("GET", methode, nil)
	mapReq.Header.Set("Authorization", "Basic "+h.mapiAuth)
	mapiResp, err := client.Do(mapReq)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("mapi the status code %v\n", mapiResp.StatusCode)
	if mapiResp.StatusCode != 200 {
		if mapiResp.StatusCode == 404 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		//TODO break this down
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	// order of writing a response
	//header
	//responseCode
	//body
	matUrl := fmt.Sprintf("http://%s%s%s", h.matHost, matPath, uuid)
	log.Printf("sending to MAT at " + matUrl);
	client := &http.Client{}
	matReq, err := http.NewRequest("POST", matUrl, mapiResp.Body)
	matReq.Host = h.matHostHeader
	matReq.Header.Set("Content-Type", "application/json")

	q := matReq.URL.Query()
	q.Add("preview", "true")
	matReq.URL.RawQuery = q.Encode()

	matResp, err := client.Do(matReq)

	log.Printf("mat the status code %v\n", matResp.StatusCode)

	if matResp.StatusCode != 200 {
		//TODO break this down
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.Copy(w, matResp.Body)
}

func (h Handlers) mapiCheck()  fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Articvle Preview Service will not work",
		Name:             "Methode Api Availablilty Check",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: "Checks that Methode API Service is reachable. Article Preview Service requests native content from Methode API service.",
		Checker:          func() (string, error) {
			return checkMehtodeApiAvailablity(h.mapiHost, h.mapiAuth)
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

package main

import (
	"os"
	"io"
	"log"
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"

)

func main() {
	app := cli.App("content-preview", "A RESTful API for retrieving and transforming content preview data")
	mapiAuth := app.StringOpt("mapi-auth", "default", "Basic auth for MAPI")
	mapiUri := app.StringOpt("mapi-uri", "http://methode-api-uk-p.svc.ft.com/eom-file/", "Host and path for MAPI")
	matUri := app.StringOpt("mat-uri", "http://ftapp05951-lvpr-uk-int:8080/content-transform/", "Host and path for MAT")

	app.Action = func() {
		r := mux.NewRouter()
		handler := Handlers{*mapiAuth, *mapiUri, *matUri}
		r.HandleFunc("/content-preview/{uuid}", handler.contentPreviewHandler)
		http.Handle("/", r)

		log.Fatal(http.ListenAndServe(":8084", nil))
	}
	app.Run(os.Args)

}

type Handlers struct {
	mapiAuth string
	mapiUri string
	matUri string
}

func (h Handlers) contentPreviewHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request");

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		log.Fatal("missing UUID")
	}

	log.Printf(h.mapiAuth);
	log.Printf(h.mapiUri);
	methode := h.mapiUri + uuid


	client := &http.Client{}
	mapReq, _ := http.NewRequest("GET", methode, nil)
	mapReq.Header.Set("Authorization", h.mapiAuth)
	mapiResp, err := client.Do(mapReq)



	if err !=nil {
		log.Fatal(err)
	}
	if mapiResp.StatusCode !=200 {
		if mapiResp.StatusCode == 404 {
			w.WriteHeader(http.StatusNotFound);
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

	matUrl := h.matUri + uuid
	matResp, err := http.Post(matUrl, "application/json", mapiResp.Body)

	if matResp.StatusCode !=200 {
		//TODO break this down
		fmt.Printf("---the status code %v", matResp.StatusCode)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Printf("the status code %v", matResp.StatusCode)
	io.Copy(w, matResp.Body)
}


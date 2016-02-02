package main

import (
	"io"
	"log"
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
)

func main() {
	server()

}

func server() {
	r := mux.NewRouter()
	r.HandleFunc("/content-preview/{uuid}", contentPreviewHandler)
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8084", nil))
}

func contentPreviewHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request");

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		log.Fatal("missing UUID")
	}

	methode := "http://methode-api-uk-p.svc.ft.com/eom-file/" + uuid
	mapiResp, err := http.Get(methode)

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

	matUrl := "http://ftapp05951-lvpr-uk-int:8080/content-transform/" + uuid
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


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
	r.HandleFunc("/content-preview/{uuid}", fooHandler)
	http.Handle("/", r)

	//http.HandleFunc("/content-preview", fooHandler)

	log.Fatal(http.ListenAndServe(":8084", nil))
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request");

	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		log.Fatal("missing UUID")
	}

	methode := "http://methode-api-uk-p.svc.ft.com/eom-file/" + uuid
	body, err := curl(methode)
	if err !=nil {
		log.Fatal(err)
	}
	matUrl := "http://ftapp05951-lvpr-uk-int:8080/content-transform/" + uuid
	resp, err := http.Post(matUrl, "application/json", body)

	if err !=nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		//TODO
		fmt.Printf("---the status code %v", resp.StatusCode)

	}

	fmt.Printf("the status code %v", resp.StatusCode)
	io.Copy(w, resp.Body)
}


func curl(url string) (respBody io.ReadCloser, err error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatalf("Unexpected status code %d", resp.StatusCode)
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return resp.Body, err
}

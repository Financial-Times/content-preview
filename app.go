package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	resp, err := http.Get("http://methode-article-transformer-01-pr-uk-p.svc.ft.com/content/b7b871f6-8a89-11e4-8e24-00144feabdc0")
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatalf("Unexpected status code %d", resp.StatusCode)
	}

	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}

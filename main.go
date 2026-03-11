package main

import (
	"net/http"
)

type Address struct {
	URL string
}

func shorten(w http.ResponseWriter, r *http.Request) {
	if (r.Method != "POST") {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}


}

func main() {

}

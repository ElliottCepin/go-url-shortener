package main

import (
	"net/http"
)

func shorten(w http.ResponseWriter, r *http.Request) {
	if (r.Method != "POST") {
		w.WriteHeader(http.StatusMethodNotAllowed) // I'd rather just write 405, but whatever	
	}
}

func main() {

}

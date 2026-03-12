package main

import (
	"net/http"
	"math/rand"
	"fmt"
	"encoding/json"
)

type Address struct {
	URL string
}

var alphabet string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

func shorten(w http.ResponseWriter, r *http.Request) {
	if (r.Method != "POST") {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var addr Address
	err := json.NewDecoder(r.Body).Decode(&addr)

	if (err != nil) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	
	var code string = ""
	for i := 0; i < 10; i++ {
		index := rand.Int() % len(alphabet)
		code += alphabet[index:index+1]
	}
	
	fmt.Fprintf(w, "%v", code)

}

func main() {

}

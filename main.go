package main

import (
	"net/http"
	"math/rand"
	"fmt"
	"encoding/json"
	"sync"
)

type Shortcode struct {
	URL string
	Code string
}

type Address struct {
	URL string
}

var alphabet string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"
var shortcodes []Shortcode = make([]Shortcode, 0, 0)
var codeLock sync.Mutex 

// generates a random shortcode or 
func generateShortcode(shortcode string, address string) string {
	if (len(shortcode) > 10) {
		valid := true
		codeLock.Lock()
		for i := 0; i < len(shortcodes); i++ {
			if (shortcodes[i].Code == shortcode) {
				valid = false
				break
			}
		}
		if (valid) {
		shortcodes = append(shortcodes, Shortcode{address, shortcode})
			codeLock.Unlock()
			return shortcode
		}
		codeLock.Unlock()
		
	}
	

	var code string = ""
	for i := 0; i < 12; i++ {
		index := rand.Int() % len(alphabet)
		code += alphabet[index:index+1]
	}

	return generateShortcode(code, address)
}

func shorten(w http.ResponseWriter, r *http.Request) {
	if (r.Method != "POST") {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var addr Address
	err := json.NewDecoder(r.Body).Decode(&addr)

	if (err != nil) {
		http.Error(w, "JSON Invalid", http.StatusBadRequest)
		return
	}
	
	code := generateShortcode("", addr.URL)


	http.HandleFunc("/" + code, func(w2 http.ResponseWriter, r2 *http.Request) {
		if (r2.Method != "GET") {
			w2.WriteHeader(http.StatusMethodNotAllowed)
			return
		} 

		w2.Header().Set("Location", addr.URL)
		w2.WriteHeader(http.StatusMovedPermanently)

	})
	fmt.Fprintf(w, "%v", code)


}

func serve() error {
	http.ListenAndServe(":8080", nil)

	return nil
}
func main() {

}

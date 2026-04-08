package main

import (
	"net/http"
	"math/rand"
	"fmt"
	"encoding/json"
	"sync"
	"time"
	"io"
)

type Shortcode struct {
	URL string
	Code string
	Clicks int
	CreatedAt time.Time
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
		shortcodes = append(shortcodes, Shortcode{address, shortcode, 0, time.Now()})
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
func wrapShorten(mux *http.ServeMux) (http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request) {
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


	mux.HandleFunc("/" + code, func(w2 http.ResponseWriter, r2 *http.Request) {
		if (r2.Method != "GET") {
			w2.WriteHeader(http.StatusMethodNotAllowed)
			return
		} 

		found := false	
		for i := 0; i<len(shortcodes); i++ {
			if shortcodes[i].URL == addr.URL {
				shortcodes[i].Clicks += 1
				found = true
				break	
			}
		}
		
		if !found {
			w2.WriteHeader(http.StatusInternalServerError)
			return	
		}
		

		w2.Header().Set("Location", addr.URL)
		w2.WriteHeader(http.StatusMovedPermanently)
		

	})
	mux.HandleFunc("/stats/" + code, func(w2 http.ResponseWriter, r2 *http.Request) {
		if (r2.Method != "GET") {
			w2.WriteHeader(http.StatusMethodNotAllowed)
			return
		} 
		
		found := false
		var index int
		for i := 0; i<len(shortcodes); i++ {
			if shortcodes[i].URL == addr.URL {
				found = true
				index = i
				break	
			}
		}
	
		if !found {
			w2.WriteHeader(http.StatusBadRequest)
			return	
		}
		

		reader, err := json.Marshal(shortcodes[index])
		w2.Header().Set("Content-Type", "application/json")

		if (err != nil) {
			w2.WriteHeader(http.StatusInternalServerError)
			return
		}

		w2.Write(reader)

		

	})
	fmt.Fprintf(w, "%v", code)

	}
}

func logger(next *http.HandlerFunc, writer io.Writer) (http.HandlerFunc){
	return func (w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w,r)
	}
}


func serve(mux *http.ServeMux, writer io.Writer) error {
	shorten := wrapShorten(mux)
	mux.Handle("/shorten", logger(&shorten, writer))
	http.ListenAndServe(":8080", mux)

	return nil
}
func main() {

}

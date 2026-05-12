package main

import (
	"net/http"
	"net"
	"math/rand"
	"fmt"
	"encoding/json"
	"sync"
	"time"
	"io"
	"os"
	"syscall"
	"os/signal"
	"context"
	"log"
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

var userLock sync.Mutex
var users map[string]int = make(map[string]int)
var usertime map[string]time.Time = make(map[string]time.Time)
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
func wrapShorten(mux *http.ServeMux, writer io.Writer) (http.HandlerFunc) {
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


	mux.HandleFunc("/" + code, rateLimiter(logger(func(w2 http.ResponseWriter, r2 *http.Request) {
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
		

	}, writer)))
	mux.HandleFunc("/stats/" + code, rateLimiter(logger(func(w2 http.ResponseWriter, r2 *http.Request) {
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

		

	}, writer)))
	fmt.Fprintf(w, "%v", code)

	}
}

func logger(next http.HandlerFunc, writer io.Writer) (http.HandlerFunc){
	return func (w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if (err != nil) {
			fmt.Fprintf(writer, "Error logging ip: %v", err)	
			fmt.Fprintf(writer, "method: %s", r.Method)
		} else { 
			fmt.Fprintf(writer, "method: %s, ip: %s", r.Method, ip)
		}
		next.ServeHTTP(w,r)
	}
}


func rateLimiter(next http.HandlerFunc) (http.HandlerFunc){
	return func (w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if (err != nil) {	
			next.ServeHTTP(w,r)
		} else {
			userLock.Lock()
			defer userLock.Unlock()
			strikes, ok := users[ip]
			utime, ok := usertime[ip]

			if (ok) {
				if ((time.Now().Unix() - utime.Unix()) > 60 ) {
					users[ip] = 1
					usertime[ip] = time.Now()
				} else {
					users[ip] += 1
					if (strikes > 9) {
						http.Error(w, "Rate limited", http.StatusForbidden)
						return
					} 
				}
			} else {
				users[ip] = 1
				usertime[ip] = time.Now()
			}

			next.ServeHTTP(w,r)
		}
	}
}
func serve(mux *http.ServeMux, writer io.Writer) error {
	shorten := wrapShorten(mux, writer)
	logged := logger(shorten, writer)
	limited := rateLimiter(logged)
	mux.Handle("/shorten", limited)
	
	srv := &http.Server{Addr: ":8080", Handler: mux,}
	
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Forced shutdown %v", err)
	}
	return nil
}
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    		http.ServeFile(w, r, "index.html")
	})
	serve(mux, os.Stdout)
	
}

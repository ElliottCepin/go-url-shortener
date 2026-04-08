package main

import (
	"net/http/httptest"
	"net/http"
	"testing"
	"encoding/json"
	"bytes"
	"regexp"
	"strings"
	"time"
	"io"
)


func TestLoggerStats(t *testing.T) {
	// the basic idea is to get the logger to write to io.Writer
	// then, we can test different cases
	// First, however, we have to actually know what our logger looks like.
	
	// STEPS:
	// - set up mux
	// - handle endpoints wrapped with logger
	// - logger should take an io.Writer
	// - figure out what io.write to use for testing
	// - test logger outputs
	// - - figure out what ip gets used if you do go test

	var b bytes.Buffer

	root := http.NewServeMux()
	
	go func () {
		if err := serve(root, &b); err != nil {
			t.Errorf("Unexpected Error: %v", err)
		}
	}()
	
	address := Address{"https://elliottcepin.dev/"}
	jsonEncoded, err := json.Marshal(address)
	
	
	if (err != nil) {
		t.Errorf("Unexpected Error: %v", err)
	}

	reader := bytes.NewReader(jsonEncoded)

	req, err := http.NewRequest("GET", "http://localhost:8080/shorten", reader)

	if err != nil {
		t.Errorf("Unexpected error with request: %v", err)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	_, err = client.Do(req)
	
	if err != nil {
		t.Errorf("Unexpected error enacting request: %v", err)
	}
	
	logged := b.String()
	if (logged != "this got logged") {
		t.Errorf("Expected to log 'this got logged', instead logged: %s", logged)
	}
	
}

func TestStats(t *testing.T) {
	mux := http.NewServeMux()
	address := Address{"https://elliottcepin.dev/"}
	var b bytes.Buffer

	go func() {
		if err := serve(mux, &b); err != nil {
			t.Errorf("Unexpected error %v", err)
		}
	}()

	jsonEncoded, err := json.Marshal(address)
	
	
	if (err != nil) {
		t.Errorf("Unexpected Error: %v", err)
	}

	reader := bytes.NewReader(jsonEncoded)
	
	req, err := http.NewRequest("POST", "http://localhost:8080/shorten", reader)

	if (err != nil) {
		t.Errorf("Unexpected Error: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Ripped from stack overflow: Etienne Bruines
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	res, err := client.Do(req)

	body, err := io.ReadAll(res.Body)

	if err != nil {
		t.Errorf("Unexpected error reading request body: %v", err)
	}

	defer res.Body.Close()
	code := string(body)


	
	req, err = http.NewRequest("GET", "http://localhost:8080/" + code, nil)

	if err != nil {
		t.Errorf("Unexpected error with request: %v", err)
	}


	res, err = client.Do(req)

	if (err != nil) {
		t.Errorf("Unexpected Error: %v", err)
	}
	
	res, err = http.Get("http://localhost:8080/stats/" + code)
	
	var shortcode Shortcode

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&shortcode)

	if (err != nil) {
		t.Errorf("Unexpected decoder error %v", err)
	}

	if (shortcode.Clicks != 1) {
		t.Errorf("Expected shortcode clicks to be 1, got: %v", shortcode.Clicks)
	}

	if (shortcode.CreatedAt.Day() != time.Now().Day()) {
		t.Errorf("Expected shortcode creation day to be %v, got: %v", shortcode.CreatedAt.Day(), time.Now().Day())
	}


}

func TestRedirect(t *testing.T) {
	mux := http.NewServeMux()
	var b bytes.Buffer
	address := Address{"https://elliottcepin.dev/"}
	jsonEncoded, err := json.Marshal(address)
	
	
	if (err != nil) {
		t.Errorf("Unexpected Error: %v", err)
	}

	reader := bytes.NewReader(jsonEncoded)

	req, err := http.NewRequest("POST", "http://localhost:8080/shorten", reader)

	if err != nil {
		t.Errorf("Unexpected error with request setup: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	go func() {
		if err := serve(mux, &b); err != nil {
			t.Errorf("Unexpected error %v", err)
		}
	}()

	// Ripped from stack overflow: Etienne Bruines
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	res, err := client.Do(req)

	if err != nil {
		t.Errorf("Unexpected error enacting request: %v", err)
	}
	
	body, err := io.ReadAll(res.Body)
	
	if err != nil {
		t.Errorf("Unexpected error reading request body: %v", err)
	}

	defer res.Body.Close()

	code := string(body)
	
	req, err = http.NewRequest("GET", "http://localhost:8080/" + code, nil)

	if err != nil {
		t.Errorf("Unexpected error with request: %v", err)
	}


	res, err = client.Do(req)

	if (err != nil) {
		t.Errorf("Unexpected Error: %v", err)
	}

	location, err := res.Location()
	rescode := res.StatusCode

	if (err != nil) {
		t.Errorf("Unexpeccted error retrieving location: %v", err)
	}

	if (rescode != 301) {
		t.Errorf("Expected response code 301, got %v", rescode)
	}

	if (location.String() != "https://elliottcepin.dev/" ) {
		t.Errorf("Expected redirect to 'https://elliottcepin.dev/', got %v", location.String())
	}

}

func TestGenerateShortcodeDuplicateEntries(t *testing.T) {
	code1 := generateShortcode("Chickencoop", "https://hyper.link/0")
	code2 := generateShortcode("Chickencoop", "https://hyper.link/1")

	if (code1 == code2) {
		t.Errorf("returned shortcodes are identical: %s, %s", code1, code2)
	}
}

func TestShortenMalformedRequest(t *testing.T) {
	mux := http.NewServeMux()
	shorten := wrapShorten(mux)
	reader := strings.NewReader("{\"URL\": \"http\")")

	req := httptest.NewRequest("POST", "/shorten", reader)
	rec := httptest.NewRecorder()
	
	shorten(rec, req)
	
	status := rec.Code

	if (status != http.StatusBadRequest) {
		t.Errorf("Expected error code %v. Got error code %v", http.StatusBadRequest, status)
	}
}	

func TestShortenReturnsCode(t *testing.T) {
	mux := http.NewServeMux()
	shorten := wrapShorten(mux)
	address := Address{"https://elliottcepin.dev/"}
	jsonEncoded, err := json.Marshal(address)
	
	if (err != nil) {
		t.Errorf("Unexpected Error: %v", err)
	}

	reader := bytes.NewReader(jsonEncoded)
	
	req := httptest.NewRequest("POST", "/shorten", reader)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	shorten(rec, req)
	code := rec.Body.String()

	matched, _ := regexp.MatchString("^([a-z]|[A-Z]|[0-9])*$", code)
	if (!matched) {
		t.Errorf("%v is an invalid code: failed regex", rec.Body.String())
	}

	if (len(code) < 10) {
		t.Errorf("%v is an invalid code: length < 10", rec.Body.String())
	}

	


}

func TestShortenGet(t *testing.T) {
	mux := http.NewServeMux()
	shorten := wrapShorten(mux)
	req := httptest.NewRequest("GET", "/shorten", nil)
	rec := httptest.NewRecorder()

	shorten(rec, req)

	if (rec.Code != 405) {
		t.Errorf("Expected code 405. Got %v", rec)
	}
}


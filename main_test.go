package main

import (
	"net/http/httptest"
	"net/http"
	"testing"
	"encoding/json"
	"bytes"
	"regexp"
	"strings"
)

func TestGenerateShortcodeDuplicateEntries(t *testing.T) {
	code1 := generateShortcode("Chickencoop", "https://hyper.link/0")
	code2 := generateShortcode("Chickencoop", "https://hyper.link/1")

	if (code1 == code2) {
		t.Errorf("returned shortcodes are identical: %s, %s", code1, code2)
	}
}

func TestShortenMalformedRequest(t *testing.T) {
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
	req := httptest.NewRequest("GET", "/shorten", nil)
	rec := httptest.NewRecorder()

	shorten(rec, req)

	if (rec.Code != 405) {
		t.Errorf("Expected code 405. Got %v", rec)
	}
}


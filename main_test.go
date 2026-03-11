package main

import (
	"net/http/httptest"
	"testing"
	"encoding/json"
	"bytes"
	"regexp"
)

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

	matched, _ := regexp.MatchString("^([a-z]|[A-Z]|[0-9])*$", rec.Body.String())
	if (matched) {
		t.Errorf("%v is an invalid code: failed regex", rec.Body.String())
	}

	if (len(rec.Body.String()) < 10) {
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


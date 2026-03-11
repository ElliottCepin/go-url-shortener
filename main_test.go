package main

import (
	"net/http/httptest"
	"testing"
)

func TestShortenPost(t *testing.T) {
	req := httptest.NewRequest("POST", "/shorten", nil)
	rec := httptest.NewRecorder()

	shorten(rec, req)

	if (rec.Code != 200) {
		t.Errorf("Expected code 200. Got %v", rec)
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


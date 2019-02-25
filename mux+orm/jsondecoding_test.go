package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-chi/render"
)

var jsonBytes = []byte(`{"Name":"Alice","Body":"Hello","Time":1294706395881547000}`)
var reader = strings.NewReader(string(jsonBytes))

type JsonStruct struct {
	Name string `json="name`
	Body string `json="body`
	Time int64  `json="time`
}

func BenchmarkJsonDecoder(b *testing.B) {
	var s JsonStruct

	for i := 0; i < b.N; i++ {
		_ = json.NewDecoder(reader).Decode(&s)
	}
}

func BenchmarkJsonUnmarshal(b *testing.B) {
	var s JsonStruct

	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(jsonBytes, &s)
	}
}

func BenchmarkChiRenderJsonDecoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		render.DecodeJSON(reader, &b)
	}
}

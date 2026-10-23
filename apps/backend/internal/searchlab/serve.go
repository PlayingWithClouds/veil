package searchlab

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
)

//go:embed label.html
var labelPage []byte

// labelData is what the labelling page loads: the queries, every list with
// its results, and the grades given so far.
type labelData struct {
	Queries []Query                   `json:"queries"`
	Lists   []List                    `json:"lists"`
	Grades  map[string]map[string]int `json:"grades"`
}

// gradeRequest sets or clears (grade −1) one result's grade.
type gradeRequest struct {
	Query string `json:"query"`
	Key   string `json:"key"`
	Grade int    `json:"grade"`
}

// Serve runs the labelling page on address; every grade is written to
// labels.tsv right away.
func Serve(address, directory string, config *Config, snapshot *Snapshot, labels *Labels) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.Write(labelPage)
	})
	mux.HandleFunc("GET /api/data", func(writer http.ResponseWriter, request *http.Request) {
		writeJSON(writer, labelData{Queries: config.Queries, Lists: snapshot.Lists, Grades: labels.all()})
	})
	mux.HandleFunc("POST /api/grade", func(writer http.ResponseWriter, request *http.Request) {
		var grade gradeRequest
		if err := json.NewDecoder(request.Body).Decode(&grade); err != nil || grade.Grade > GradePerfect {
			http.Error(writer, "bad grade", http.StatusBadRequest)
			return
		}
		labels.Set(grade.Query, grade.Key, grade.Grade)
		if err := SaveLabels(directory, snapshot, labels); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	})
	log.Printf("labelling page on http://%s", address)
	return http.ListenAndServe(address, mux)
}

// all returns a copy of every grade.
func (labels *Labels) all() map[string]map[string]int {
	labels.mu.Lock()
	defer labels.mu.Unlock()
	out := map[string]map[string]int{}
	for queryID, grades := range labels.grades {
		out[queryID] = map[string]int{}
		for key, grade := range grades {
			out[queryID][key] = grade
		}
	}
	return out
}

// writeJSON sends value as JSON.
func writeJSON(writer http.ResponseWriter, value any) {
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		log.Printf("searchlab: write response: %v", err)
	}
}

package server

import (
	"encoding/base64"
	"encoding/json"
	v11 "flix-indexer/flix/v1_1"
	"fmt"
	"log"
	"net/http"
)

type HttpServer struct {
	Logger  *log.Logger
	Indexer *v11.TemplateIndexer
	Port    int
}

type ListResponse struct {
	Data []v11.Template `json:"data"`
}

func (s *HttpServer) Setup() {
	http.HandleFunc("/v1.1/templates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		response := ListResponse{
			Data: []v11.Template{},
		}

		cadenceBase64 := r.URL.Query().Get("cadence_base64")
		if cadenceBase64 != "" {
			cadenceSource, err := base64.StdEncoding.DecodeString(cadenceBase64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			}

			match, err := s.Indexer.GetBySource(cadenceSource)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			if match != nil {
				response.Data = []v11.Template{*match}
			}

		} else {
			response.Data = s.Indexer.List()
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			s.Logger.Printf("error serializing response: %e", err)
		}

		_, err = w.Write(jsonResponse)
		if err != nil {
			s.Logger.Printf("error writing response: %e", err)
		}
	})

	http.HandleFunc("/v1.1/templates/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		response := s.Indexer.GetByID(r.PathValue("id"))
		if response == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			s.Logger.Printf("error serializing response: %e", err)
		}

		_, err = w.Write(jsonResponse)
		if err != nil {
			s.Logger.Printf("error writing response: %e", err)
		}
	})
}

func (s *HttpServer) Start() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil)
}

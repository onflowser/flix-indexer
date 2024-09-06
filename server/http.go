package server

import (
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
		response, err := json.Marshal(ListResponse{
			Data: s.Indexer.List(),
		})

		if err != nil {
			s.Logger.Printf("error serializing response: %e", err)
		}

		w.Header().Add("Content-Type", "application/json")

		_, err = w.Write(response)
		if err != nil {
			s.Logger.Printf("error writing response: %e", err)
		}
	})
}

func (s *HttpServer) Start() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil)
}

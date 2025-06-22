package api

import (
	"net/http"
)

func (s *Server) GetTags(w http.ResponseWriter, r *http.Request, params GetTagsParams) {
}

func (s *Server) GetTag(w http.ResponseWriter, r *http.Request, name Tag) {
}

func (s *Server) PutTag(w http.ResponseWriter, r *http.Request, name Tag) {
}

func (s *Server) DeleteTag(w http.ResponseWriter, r *http.Request, name Tag){
}

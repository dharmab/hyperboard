package api

import (
	"net/http"
)

func (s *Server) GetPosts(w http.ResponseWriter, r *http.Request, params GetPostsParams) {
}

func (s *Server) GetPost(w http.ResponseWriter, r *http.Request, id Id) {
}

func (s *Server) PutPost(w http.ResponseWriter, r *http.Request, id Id) {
}

func (s *Server) DeletePost(w http.ResponseWriter, r *http.Request, id Id) {
}

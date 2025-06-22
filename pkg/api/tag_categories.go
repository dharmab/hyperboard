package api

import (
	"net/http"
)

func (s *Server) GetTagCategories(w http.ResponseWriter, r *http.Request, params GetTagCategoriesParams) {
}

func (s *Server) GetTagCategory(w http.ResponseWriter, r *http.Request, name TagCategory) {
}

func (s *Server) PutTagCategory(w http.ResponseWriter, r *http.Request, name TagCategory) {
}

func (s *Server) DeleteTagCategory(w http.ResponseWriter, r *http.Request, name TagCategory) {
}

package main

import (
	"encoding/json"
	"net/http"
)

type createUserRequest struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Attribute attribute `json:"attribute"`
}

type attribute struct {
	Birthday string `json:"birthday,omitempty"`
	Gender   string `json:"gender,omitempty"`
}

type createUserResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type userHandler struct {
}

type userProtoHandler struct {
}

func (h *userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if v := r.URL.Query().Get("token"); v != "12345" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var request createUserRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Some process...

	response := createUserResponse{
		ID:   11241988,
		Name: request.Name,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", " ")
	encoder.Encode(&response)
}

func (h *userProtoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	response := &UserProtoResponse{
		Id:     169743,
		Name:   "Immortan Joe",
		Active: true,
		Setting: &UserProtoResponse_Setting{
			Email: "immortan@madmax.com",
		},
	}
	buf, err := response.Marshal()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/protobuf")
	w.Write(buf)
}

package server

import (
	"safelyyou-assessment/devices"
)

import "net/http"

type Server struct {
	store *devices.Store
}

func New(store *devices.Store) *Server {

}

func (s *Server) Run(addr string) error {

}

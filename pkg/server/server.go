package server

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/rahul0tripathi/framecoiner/pkg/log"
)

type Server struct {
	app     *echo.Echo
	notify  chan error
	address string
}

type Router interface {
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

func New(address string, logger log.Logger) *Server {
	app := echo.New()

	return &Server{
		app:     app,
		notify:  make(chan error),
		address: address,
	}

}

func (s *Server) Router() Router {
	return s.app
}

func (s *Server) Start() {
	go func() {
		s.notify <- s.app.Start(s.address)
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown(context.Background())
}

func ResponseJSON(c echo.Context, status int, response interface{}) error {
	return c.JSON(status, response)
}

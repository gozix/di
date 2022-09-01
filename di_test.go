// Copyright 2022 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package di_test

import (
	"errors"
	"io"
	"net/http"
)

type (
	Controller interface {
		Register(server *http.ServeMux)
	}

	BarController struct {
		_ int
	}

	BazController struct {
		_ int
	}

	CycledController struct {
		_ int
	}

	FlakyController struct {
		_ int
	}
)

func (c *BarController) Register(srv *http.ServeMux) {
	srv.HandleFunc("/bar", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "Bar")
	})
}

func (c *BazController) Register(srv *http.ServeMux) {
	srv.HandleFunc("/baz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "Baz")
	})
}

func (c *CycledController) Register(srv *http.ServeMux) {
	srv.HandleFunc("/cycled", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "cycled")
	})
}

func (c *FlakyController) Register(srv *http.ServeMux) {
	srv.HandleFunc("/flaky", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "flaky")
	})
}

func NewServer(mux *http.ServeMux) *http.Server {
	return &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
}

func NewServerMux(controllers []Controller) (*http.ServeMux, func() error) {
	var s = http.NewServeMux()
	for _, c := range controllers {
		c.Register(s)
	}

	return s, func() error {
		return nil
	}
}

func NewBarController() *BarController {
	return &BarController{}
}

func NewBazController() *BazController {
	return &BazController{}
}

func NewFlakyController() (*FlakyController, error) {
	return nil, errors.New("always fail")
}

func NewCycledController(cycled *CycledController) *CycledController {
	return cycled
}

package celeritas

import (
	"fmt"
	"net/http"
	"time"
)

func (c *Celeritas) ListenAndServe() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", c.config.port),
		ErrorLog:     c.ErrorLog,
		Handler:      c.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	if c.DB.Pool != nil {
		defer c.DB.Pool.Close()
	}

	if redisPool != nil {
		defer redisPool.Close()
	}

	if badgerConn != nil {
		defer badgerConn.Close()
	}

	go c.listenRPC()
	c.InfoLog.Printf("Listening on port %s", c.config.port)
	return srv.ListenAndServe()
}

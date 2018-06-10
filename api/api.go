// Package api contains the general framework for writing handlers in the
// CheeseGull API.
package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	raven "github.com/getsentry/raven-go"
	"github.com/julienschmidt/httprouter"
	"github.com/paulbellamy/ratecounter"

	"github.com/Gigamons/cheesegull/downloader"
	"github.com/Gigamons/cheesegull/housekeeper"
	"github.com/Gigamons/cheesegull/logger"
)

// Context is the information that is passed to all request handlers in relation
// to the request, and how to answer it.
type Context struct {
	Request  *http.Request
	DB       *sql.DB
	SearchDB *sql.DB
	House    *housekeeper.House
	DLClient *downloader.Client
	writer   http.ResponseWriter
	params   httprouter.Params
}

var RSLR = ratecounter.NewAvgRateCounter(time.Second * 60)
var BMSLR = ratecounter.NewAvgRateCounter(time.Second * 60)

// Write writes content to the response body.
func (c *Context) Write(b []byte) (int, error) {
	return c.writer.Write(b)
}

// ReadHeader reads a header from the request.
func (c *Context) ReadHeader(s string) string {
	return c.Request.Header.Get(s)
}

// WriteHeader sets a header in the response.
func (c *Context) WriteHeader(key, value string) {
	c.writer.Header().Set(key, value)
}

// Code sets the response's code.
func (c *Context) Code(i int) {
	c.writer.WriteHeader(i)
}

// Param retrieves a parameter in the URL's path.
func (c *Context) Param(s string) string {
	return c.params.ByName(s)
}

func (c *Context) IncreaseR() {
	RSLR.Incr(1)
}

func (c *Context) IncreaseBM() {
	BMSLR.Incr(1)
}

// WriteJSON writes JSON to the response.
func (c *Context) WriteJSON(code int, v interface{}) error {
	c.WriteHeader("Content-Type", "application/json; charset=utf-8")
	c.Code(code)
	return json.NewEncoder(c.writer).Encode(v)
}

var envSentryDSN = os.Getenv("SENTRY_DSN")

// Err attempts to log an error to Sentry, as well as stdout.
func (c *Context) Err(err error) {
	if err == nil {
		return
	}
	if envSentryDSN != "" {
		raven.CaptureError(err, nil, raven.NewHttp(c.Request))
	}
	log.Println(err)
}

type handlerPath struct {
	method, path string
	f            func(c *Context)
}

var handlers []handlerPath

// GET registers a handler for a GET request.
func GET(path string, f func(c *Context)) {
	handlers = append(handlers, handlerPath{"GET", path, f})
}

// POST registers a handler for a POST request.
func POST(path string, f func(c *Context)) {
	handlers = append(handlers, handlerPath{"POST", path, f})
}

// CreateHandler creates a new http.Handler using the handlers registered
// through GET and POST.
func CreateHandler(db, searchDB *sql.DB, house *housekeeper.House, dlc *downloader.Client) http.Handler {
	r := httprouter.New()
	for _, h := range handlers {
		// Create local copy that we know won't change as the loop proceeds.
		h := h
		r.Handle(h.method, h.path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			start := time.Now()
			ctx := &Context{
				Request:  r,
				DB:       db,
				SearchDB: searchDB,
				House:    house,
				DLClient: dlc,
				writer:   w,
				params:   p,
			}
			defer func() {
				err := recover()
				if err == nil {
					return
				}
				switch err := err.(type) {
				case error:
					ctx.Err(err)
				case stringer:
					ctx.Err(errors.New(err.String()))
				case string:
					ctx.Err(errors.New(err))
				default:
					log.Println("PANIC", err)
				}
				debug.PrintStack()
			}()
			h.f(ctx)
			ctx.IncreaseR()
			logger.Request(" %-10s %-4s %s",
				time.Since(start).String(),
				r.Method,
				r.URL.Path,
			)
		})
	}
	return r
}

func DDOG() {
	c, err := statsd.New("127.0.0.1:8125")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()
	c.Namespace = "Cheesegull"
	go func() {
		for {
			err = c.Count("BMPM", BMSLR.Hits(), nil, BMSLR.Rate())
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(time.Second * 19)
		}
	}()
	go func() {
		for {
			err = c.Count("RSLR", RSLR.Hits(), nil, RSLR.Rate())
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(time.Second * 19)
		}
	}()
	for {
		time.Sleep(time.Second * 19)
	}
}

type stringer interface {
	String() string
}

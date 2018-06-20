package api

import (
	"expvar"
	"io/ioutil"
	"os"

	"github.com/Gigamons/cheesegull/config"
)

// Version is set by main and it is given to requests at /
var Version = "v2.DEV"

func index(c *Context) {
	c.WriteHeader("Content-Type", "text/html; charset=utf-8")
	if _, err := os.Stat("index.html"); os.IsNotExist(err) {
		c.Write([]byte(config.Parse().Server.Website))
	} else {
		f, err := ioutil.ReadFile("index.html")
		if err != nil {
			c.Write([]byte(config.Parse().Server.Website))
			return
		}
		c.Write(f)
	}
}

var _evh = expvar.Handler()

func expvarHandler(c *Context) {
	_evh.ServeHTTP(c.writer, c.Request)
}

func init() {
	GET("/", index)
	GET("/expvar", expvarHandler)
}

package main

import (
	"fmt"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qservice/core"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime"
	"strconv"
)

func QRecovery(f func(c *gin.Context, err interface{})) gin.HandlerFunc {
	return QRecoveryWithWriter(f)
}

func QRecoveryWithWriter(f func(c *gin.Context, err interface{})) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				qlog.Log(qlog.ERROR, err)
				//httprequest, _ := httputil.DumpRequest(c.Request, false)
				//goErr := errors.Wrap(err, 3)
				//reset := string([]byte{27, 91, 48, 109})
				//logger.Printf("[Nice Recovery] panic recovered:\n\n%s%s\n\n%s%s", httprequest, goErr.Error(), goErr.Stack(), reset)
				f(c, err)
			}
		}()
		c.Next() // execute all the handlers
	}
}

func initWeb() {

	gin.SetMode("release")

	var engine = gin.Default()
	engine.GET("/call", func(context *gin.Context) {
		var data = context.Query("data")
		var address = context.Query("address")
		var timeout = context.Query("timeout")

		var request = core.NewMessage(address, data, 15*1000, nil, nil)

		if len(timeout) > 0 {
			var err error
			request.Timeout, err = strconv.ParseInt(timeout, 10, 64)
			if err != nil {
				context.String(500, fmt.Sprintf("%v", err))
				return
			}
		}

		var response, _ = overseer.Post(request)

		if response.IsError() {
			var reply = fmt.Sprintf("%v", response.ReplyErr)
			context.String(500, reply)
		} else {
			var reply = fmt.Sprintf("%v", response.ReplyData)
			context.String(200, reply)
		}

	})

	engine.Use(QRecovery(func(c *gin.Context, err interface{}) {
		var bytes = make([]byte, 8192)
		var stack = runtime.Stack(bytes, false)
		var stackString = string(bytes[:stack])
		c.String(500, stackString)
	}))

	go func() {
		var server = &http.Server{
			Addr:    ":5555",
			Handler: engine,
		}
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

}

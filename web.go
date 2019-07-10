package main

import (
	"fmt"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qservice/core"
	"github.com/camsiabor/qservice/impl/zookeeper"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
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

	initRoute(engine)

	engine.Use(QRecovery(func(c *gin.Context, err interface{}) {

		c.String(500, qref.StackStringErr(err, 0))
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

func initRoute(engine *gin.Engine) {

	engine.GET("/call", call)
	engine.GET("/zkiter", zkiter)

}

func call(context *gin.Context) {
	var data = context.Query("data")
	var address = context.Query("address")
	var timeout = context.Query("timeout")

	var request = core.NewMessage(address, data, time.Duration(15)*time.Second)

	if len(timeout) > 0 {
		var itimeout, err = strconv.ParseInt(timeout, 10, 64)
		if err != nil {
			context.String(500, fmt.Sprintf("%v", err))
			return
		}
		request.Timeout = time.Duration(itimeout) * time.Second
	}

	var response, err = overseer.Post(request)

	if err != nil {
		var reply = fmt.Sprintf("%v", err)
		context.String(500, reply)
		return
	}

	if response.IsError() {
		var reply = fmt.Sprintf("%v", response.ReplyErr)
		context.String(500, reply)
		return
	}

	var reply = fmt.Sprintf("%v", response.ReplyData)
	context.String(200, reply)

}

func zkiter(context *gin.Context) {
	var endpoint = context.Query("endpoint")
	var path = context.Query("path")
	var depth, _ = strconv.Atoi(context.Query("depth"))
	var id = context.Query("id")

	if len(id) == 0 {
		id = endpoint
	}

	var conn, err = GetZookeeper(id, endpoint)
	if err != nil {
		context.String(500, qref.StackStringErr(err, 0))
		return
	}
	if depth < 0 {
		depth = 0
	}
	var reply = ""
	var all = []string{path}
	_ = zookeeper.Iterate(conn, path, path, depth, func(current string, parent string, root string, depth int) bool {
		reply = reply + "\n" + current
		all = append(all, current)
		return true
	})
	context.String(200, reply)
}

func zkcreate(context *gin.Context) {

}

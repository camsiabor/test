package main

import (
	"fmt"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/core"
	"github.com/gin-gonic/gin"
	"github.com/samuel/go-zookeeper/zk"
	"net/http"
	"strconv"
	"strings"
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

	engine.GET("/zk/iter", zkiter)
	engine.GET("/zk/create", zkcreate)
	engine.GET("/zk/delete", zkiter)
	engine.GET("/zk/put", zkiter)
	engine.GET("/zk/get", zkiter)

}

func call(context *gin.Context) {
	var data = context.Query("data")
	var address = context.Query("address")
	var timeout = context.Query("timeout")
	var local = util.AsBool(context.Query("local"), false)

	var request = core.NewMessage(address, data, time.Duration(15)*time.Second)

	if len(timeout) > 0 {
		var itimeout, err = strconv.ParseInt(timeout, 10, 64)
		if err != nil {
			context.String(500, fmt.Sprintf("%v", err))
			return
		}
		request.Timeout = time.Duration(itimeout) * time.Second
	}

	var err error
	var response *core.Message
	if local {
		response, err = localOverseer.Post(request)
	} else {
		response, err = clusterOverseer.Post(request)
	}

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
	var id = context.Query("id")
	var path = context.Query("path")
	var endpoint = context.Query("endpoint")
	var depth, _ = strconv.Atoi(context.Query("depth"))

	if len(id) == 0 {
		id = endpoint
	}

	var conn, err = ZkConn(id, endpoint)
	if err != nil {
		context.String(500, qref.StackStringErr(err, 0))
		return
	}
	if depth < 0 {
		depth = 0
	}
	var builder strings.Builder
	_ = ZkIterate(conn, path, path, depth, func(conn *zk.Conn, current string, parent string, root string, depth int) bool {
		builder.WriteString("\n")
		builder.WriteString(current)
		var _, stat, err = conn.Get(current)
		if err == nil {
			var info = fmt.Sprintf("   %d", stat.DataLength)
			builder.WriteString(info)
		}
		return true
	})
	context.String(200, builder.String())
}

func zkcreate(context *gin.Context) {
	var id = context.Query("id")
	var path = context.Query("path")
	var data = context.Query("data")
	var ephemeral, _ = strconv.Atoi(context.Query("ephe"))
	var sequential, _ = strconv.Atoi(context.Query("seq"))
	var conn, _ = ZkConn(id, "")

	if conn == nil {
		context.String(500, "id not found "+id)
		return
	}
	var flag int32
	if ephemeral >= 1 {
		flag = flag | zk.FlagEphemeral
	}
	if sequential >= 1 {
		flag = flag | zk.FlagSequence
	}
	var cpath, err = conn.Create(path, []byte(data), flag, zk.WorldACL(zk.PermAll))
	if err == nil {
		context.String(200, "created "+cpath)
	} else {
		context.String(500, "created fail "+path+" : "+err.Error())
	}

}

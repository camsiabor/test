package main

import (
	"encoding/json"
	"fmt"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/core"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/samuel/go-zookeeper/zk"
	"net/http"
	"path/filepath"
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

	engine.GET("/call", callp)

	engine.GET("/zk/iter", zkiter)
	engine.GET("/zk/create", zkcreate)
	engine.GET("/zk/delete", zkiter)
	engine.GET("/zk/put", zkiter)
	engine.GET("/zk/get", zkiter)

	var root = "../src/github.com/camsiabor/test/web"
	root, _ = filepath.Abs(root)

	engine.Static("/js", root+"/js")
	engine.Static("/css", root+"/css")
	engine.Static("/img", root+"/img")
	engine.Static("/res", root+"/res")
	engine.Static("/svg", root+"/svg")
	engine.Static("/tmp", root+"/tmp")
	engine.Static("/h", root+"/")

	engine.GET("/ws", wsconnect)
}

var upgrader = websocket.Upgrader{}

func wsconnect(context *gin.Context) {
	var conn, err = upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		context.String(500, qref.StackStringErr(err, 0))
		return
	}
	go wsread(conn)
}

func wsread(conn *websocket.Conn) {
	defer conn.Close()
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		err, data = wshandle(data)
		err = conn.WriteMessage(messageType, data)
		if err != nil {
			break
		}
	}
}

func wshandle(data []byte) (err error, ret []byte) {
	var code = 404
	var consume int64
	var result interface{} = "handler not found"
	var response = map[string]interface{}{}

	defer func() {

		var pan = recover()
		if pan == nil {
			err = util.AsError(pan)
		}
		if err != nil {
			result = qref.StackStringErr(err, 0)
		}
		response["code"] = code
		response["result"] = result
		response["consume"] = consume

		ret, err = json.Marshal(response)
		if err != nil {
			response["result"] = "marshal response result fail : " + err.Error()
			ret, _ = json.Marshal(response)
		}

	}()

	var request map[string]interface{}
	err = json.Unmarshal(data, &request)

	var start = time.Now().UnixNano()
	var id = util.GetInt64(request, 0, "id")
	response["id"] = id
	if err == nil {
		var action = util.GetStr(request, "", "action")
		if action == "call" {
			code = 200
			var address = util.GetStr(request, "", "method")
			var params = util.GetMap(request, false, "params")
			var timeout = util.GetInt64(request, 15000, "timeout")
			var local = util.GetBool(request, false, "local")
			result, err = call(address, params, timeout, local)
		}
	}
	var end = time.Now().UnixNano()
	consume = (end - start) / 1000 / 1000
	return
}

func call(address string, data interface{}, timeout int64, local bool) (interface{}, error) {

	var request = core.NewMessage(address, data, time.Duration(timeout)*time.Millisecond)
	var err error
	var response *core.Message
	if local {
		response, err = localOverseer.Post(request)
	} else {
		response, err = clusterOverseer.Post(request)
	}
	if err != nil {
		return nil, err
	}
	if response.IsError() {
		return nil, fmt.Errorf("%v", response.ReplyErr)
	}
	return response.ReplyData, nil
}

func callp(context *gin.Context) {
	var data = context.Query("data")
	var address = context.Query("address")
	var timeout = util.AsInt64(context.Query("timeout"), 15000)
	var local = util.AsBool(context.Query("local"), false)

	var ret, err = call(address, data, timeout, local)
	if err != nil {
		var reply = fmt.Sprintf("%v", err)
		context.String(500, reply)
		return
	}
	var reply = fmt.Sprintf("%v", ret)
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

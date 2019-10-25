package httpt

import (
	"encoding/json"
	"fmt"
	"github.com/camsiabor/qcom/qchan"
	"github.com/camsiabor/qcom/qerr"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/qtiny"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"path/filepath"
	"sync"
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

func InitWeb(config map[string]interface{}) {

	var address = util.GetStr(config, ":5555", "endpoint")

	gin.SetMode("release")

	var engine = gin.Default()

	initRoute(engine, config)

	engine.Use(QRecovery(func(c *gin.Context, err interface{}) {
		c.String(500, qerr.StackString(0, 1024, util.AsStr(err, "")))
	}))

	go func() {
		var server = &http.Server{
			Addr:    address,
			Handler: engine,
		}
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

}

func initRoute(engine *gin.Engine, config map[string]interface{}) {

	engine.GET("/call", callp)

	var root = util.GetStr(config, "../../src/github.com/camsiabor/test/web", "root")
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

var upgrader = &websocket.Upgrader{}

func wsconnect(context *gin.Context) {
	var conn, err = upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		context.String(500, qerr.StackString(0, 1024, err.Error()))
		return
	}
	go wsread(conn)
}

func wsread(conn *websocket.Conn) {
	var mutex sync.Mutex
	defer conn.Close()
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("websocket read message error %v", err.Error())
			break
		}
		go func() {
			_, data = wshandle(data)
			if data != nil {
				func() {
					mutex.Lock()
					defer mutex.Unlock()
					_ = conn.WriteMessage(messageType, data)
				}()
			}
		}()
	}
}

func wshandle(data []byte) (err error, ret []byte) {

	var code = 404
	var consume int64
	var action string
	var name string
	var cut *qerr.StackCut
	var result interface{} = "handler not found"
	var response = map[string]interface{}{}

	var start int64
	defer func() {
		var end = time.Now().UnixNano()
		consume = (end - start) / 1000 / 1000

		var pan = recover()
		if pan != nil {
			err = util.AsError(pan)
		}
		if err != nil {
			result = action + " @ " + name + "\n"
			if cut == nil {
				cut = qerr.StackCutting(1, 1024)
			}
			result = fmt.Sprintf("%v @ %v \n%v \n%v:%v %v\n%v", action, name, err.Error(), cut.File, cut.Line, cut.Func, string(cut.Stack))
			// qerr.StackStringErr(1, 1024, "%v : %v %v\n", action, name, err.Error())
			//result = fmt.Sprintf("%v @ %v \n%v \n%v:%v %v", action, name, err.Error(), cut.File, cut.Line, cut.Func)
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

	var id = util.GetInt64(request, 0, "id")

	var timeout = util.GetInt64(request, 15*1000, "timeout")
	if timeout <= 3*1000 {
		timeout = 3 * 1000
	}
	response["id"] = id

	result, cut, err = qchan.Wait(time.Duration(timeout)*time.Millisecond, true, func() (i interface{}, e error) {

		start = time.Now().UnixNano()
		if err == nil {
			action = util.GetStr(request, "", "action")
			if action == "call" {
				code = 200
				var node = util.GetStr(request, "", "node")
				var method = util.GetStr(request, "", "method")
				var params = util.GetMap(request, false, "params")
				var local = util.GetBool(request, false, "local")
				var remote = util.GetBool(request, false, "remote")
				var gatekey = util.GetStr(request, "", "gatekey")
				var flag qtiny.MessageFlag
				if local {
					flag = flag | qtiny.MessageFlagLocalOnly
				}
				if remote {
					flag = flag | qtiny.MessageFlagRemoteOnly
				}
				name = method
				result, err = call(node, method, params, gatekey, timeout, flag)
			} else if action == "script" {
				var script = util.GetStr(request, "", "script")
				var scriptType = util.GetStr(request, "anko", "type")
				var params = util.GetMap(request, false, "params")
				var method = util.GetStr(request, "", "method")
				if len(method) == 0 {
					method = util.GetStr(request, "", "name")
				}
				name = method
				result, err = exec(scriptType, script, name, params)
			}
		}

		if err != nil {
			code = 500
			result = err.Error()
		}
		return result, err

	}, nil)

	return err, ret
}

func exec(scriptType string, script string, name string, params interface{}) (interface{}, error) {
	/*
		var env = vm.NewEnv()
		_ = env.Define("params", params)
		packages.DefineImport(env)
		env.SetName(name)
		return env.Execute(script)

	*/
	panic("not implement")
}

func call(node string, method string, data interface{}, gatekey string, timeout int64, flag qtiny.MessageFlag) (string, error) {

	if data != nil {
		var bytes, err = json.Marshal(data)
		if err == nil {
			data = string(bytes)
		}
	}

	var tina = qtiny.GetTina()
	var request = qtiny.NewMessage(method, data, time.Duration(timeout)*time.Millisecond)
	request.Receiver = node
	request.LocalFlag = flag

	var microroller = tina.GetMicroroller()
	var response, err = microroller.Post(gatekey, request)
	if err != nil {
		return "", err
	}
	if response.IsError() {
		return "", fmt.Errorf("[%v]\n\n%v\n\n%v", response.Sender, response.ReplyErr, response.ReplyTrace)
	} else {
		return fmt.Sprintf("[%v]\n\n%v", response.Sender, response.ReplyData), nil
	}
}

func callp(context *gin.Context) {
	var data = context.Query("data")
	var node = context.Query("node")
	var method = context.Query("address")
	var gatekey = context.Query("gatekey")
	var timeout = util.AsInt64(context.Query("timeout"), 15000)
	var local = util.AsBool(context.Query("local"), false)
	var remote = util.AsBool(context.Query("remote"), false)

	var flag qtiny.MessageFlag
	if local {
		flag = flag | qtiny.MessageFlagLocalOnly
	}
	if remote {
		flag = flag | qtiny.MessageFlagRemoteOnly
	}

	var reply, err = call(node, method, data, gatekey, timeout, flag)
	if err == nil {
		context.String(200, reply)
	} else {
		context.String(500, err.Error())
	}
}

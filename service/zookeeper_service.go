package service

import (
	"fmt"
	"github.com/camsiabor/go-zookeeper/zk"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/impl/gateway/zookeeper"
	"github.com/camsiabor/qservice/qtiny"
	"log"

	"strings"
	"time"
)

func zkGetParams(message *qtiny.Message) (request map[string]interface{}, id string, endpoint string, path string) {
	request = util.AsMap(message.Data, false)
	id = util.GetStr(request, "", "id")
	endpoint = util.GetStr(request, "127.0.0.1:2181", "endpoint")
	path = util.GetStr(request, "/", "path")
	return
}

func ZookeeperTiny() *qtiny.TinyGuide {

	var guide = &qtiny.TinyGuide{}

	guide.Start = func(event qtiny.TinyGuideEvent, tiny qtiny.TinyKind, guide qtiny.TinyGuideKind, config map[string]interface{}, future *qtiny.Future, err error) {

		if err != nil {
			log.Printf("zookeeper tiny guide start error %v", err)
			return
		}

		_ = tiny.NanoLocalRegister(qtiny.NewNano("qam.zk.conn", 0, nil, func(message *qtiny.Message) {
			var _, id, endpoint, _ = zkGetParams(message)
			var _, err = ZooWatcherGet(id, endpoint)
			if err == nil {
				_ = message.Reply(0, id+" connected")
			} else {
				_ = message.Error(500, err.Error())
			}
		}))

		_ = tiny.NanoLocalRegister(qtiny.NewNano("qam.zk.close", 0, nil, func(message *qtiny.Message) {
			var _, id, _, _ = zkGetParams(message)
			var watcher, _ = ZooWatcherGet(id, "")
			if watcher != nil {
				_ = watcher.Stop(nil)
			}
			_ = message.Reply(0, id+" closed")
		}))

		_ = tiny.NanoLocalRegister(qtiny.NewNano("qam.zk.iter", 0, nil, func(message *qtiny.Message) {
			var request, id, endpoint, path = zkGetParams(message)
			var depth = util.GetInt(request, 3, "depth")
			var filter = util.GetStr(request, "", "filter")
			var timeout = util.GetInt64(request, 5, "timeout")
			if len(id) == 0 {
				id = endpoint
			}

			var watcher, err = ZooWatcherGet(id, endpoint)
			if err != nil {
				_ = message.Error(404, err.Error())
				return
			}
			if depth < 0 {
				depth = 0
			}
			var connectChannel = watcher.WaitForConnected()
			if connectChannel != nil {
				var chosen, connected, recvok = util.Timeout(connectChannel, time.Duration(timeout)*time.Second)
				if chosen < 0 {
					_ = message.Error(300, "connect to zookeeper timeout : "+id)
					ZooWatcherClose(id)
					return
				}
				if !connected.Bool() || !recvok {
					_ = message.Error(500, "instance closed : "+id)
					ZooWatcherClose(id)
					return
				}
			}
			var conn = watcher.GetConn()
			var builder strings.Builder
			_ = zookeeper.ZkIterate(conn, path, path, depth, func(conn *zk.Conn, current string, parent string, root string, depth int) bool {
				if len(filter) > 0 && !strings.Contains(current, filter) {
					return true
				}
				builder.WriteString("\n")
				builder.WriteString(current)
				var _, stat, err = conn.Get(current)
				if err == nil {
					var info = fmt.Sprintf("   %d", stat.DataLength)
					builder.WriteString(info)
				}
				return true
			})
			_ = message.Reply(0, builder.String())
		}))

		var watches = map[string]<-chan zk.Event{}
		_ = tiny.NanoLocalRegister(qtiny.NewNano("qam.zk.watch", 0, nil, func(message *qtiny.Message) {
			var _, id, endpoint, path = zkGetParams(message)
			var watcher, _ = ZooWatcherGet(id, endpoint)
			var conn = watcher.GetConn()
			var key = id + "@" + path

			if watches[key] != nil {
				_ = message.Reply(0, "already watching "+path)
				return
			}

			var _, _, ch, err = conn.ChildrenW(path)

			if err != nil {
				_ = message.Error(500, err.Error())
				return
			}
			watches[key] = ch

			go func() {
				for {
					var event, ok = <-ch
					if !ok {
						break
					}
					var stype = ""
					var serror = ""
					switch event.Type {
					case zk.EventNodeCreated:
						stype = "create"
					case zk.EventNodeDeleted:
						stype = "delete"
					case zk.EventNodeDataChanged:
						stype = "data change"
					case zk.EventNodeChildrenChanged:
						stype = "children change"
					}
					if event.Err != nil {
						serror = event.Err.Error()
					}
					fmt.Printf("%s %s [%s] %s\n", event.Path, stype, event.Server, serror)
				}
				fmt.Println("watching " + path + " end!")
			}()

			_ = message.Reply(0, "watching "+path)
		}))
	}

	return guide

}

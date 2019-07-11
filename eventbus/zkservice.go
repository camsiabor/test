package eventbus

import (
	"fmt"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/core"
	"github.com/camsiabor/qservice/impl/zookeeper"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
)

func getParams(message *core.Message) (request map[string]interface{}, id string, endpoint string, path string) {
	request = util.AsMap(message.Data, false)
	id = util.GetStr(request, "", "id")
	endpoint = util.GetStr(request, "127.0.0.1:2181", "endpoint")
	path = util.GetStr(request, "/", "path")
	return
}

func InitZkTService() {
	var overseer = GetOverseer(true)

	_ = overseer.ServiceRegister("qam.zk.conn", nil, func(message *core.Message) {
		var _, id, endpoint, _ = getParams(message)
		var _, err = zookeeper.ZooWatcherGet(id, endpoint)
		if err == nil {
			_ = message.Reply(0, id+" connected")
		} else {
			_ = message.Error(500, err.Error())
		}
	})

	_ = overseer.ServiceRegister("qam.zk.close", nil, func(message *core.Message) {
		var _, id, _, _ = getParams(message)
		var watcher, _ = zookeeper.ZooWatcherGet(id, "")
		if watcher != nil {
			_ = watcher.Stop(nil)
		}
		_ = message.Reply(0, id+" closed")
	})

	_ = overseer.ServiceRegister("qam.zk.iter", nil, func(message *core.Message) {
		var request, id, endpoint, path = getParams(message)
		var depth = util.GetInt(request, 0, "depth")
		var filter = util.GetStr(request, "", "filter")

		if len(id) == 0 {
			id = endpoint
		}

		var watcher, err = zookeeper.ZooWatcherGet(id, endpoint)
		if err != nil {
			_ = message.Error(404, err.Error())
			return
		}
		if depth < 0 {
			depth = 0
		}
		watcher.WaitForConnected()
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
	})

	var watches = map[string]<-chan zk.Event{}
	_ = overseer.ServiceRegister("qam.zk.watch", nil, func(message *core.Message) {
		var _, id, endpoint, path = getParams(message)
		var watcher, _ = zookeeper.ZooWatcherGet(id, endpoint)
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
	})

}

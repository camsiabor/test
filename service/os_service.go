package service

import (
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/qtiny"
	"log"
	"os"
	"os/exec"
)

func OsTinyGuide() *qtiny.TinyGuide {

	var guide = &qtiny.TinyGuide{}

	guide.Start = func(event qtiny.TinyGuideEvent, tiny qtiny.TinyKind, guide qtiny.TinyGuideKind, config map[string]interface{}, future *qtiny.Future, err error) {

		if err != nil {
			log.Printf("os tiny guide start error %v", err)
			return
		}

		_ = tiny.NanoLocalRegister(qtiny.NewNano("qam.os.exec", 0, nil, func(message *qtiny.Message) {

			var data = util.AsMap(message.Data, false)
			var cmdline = util.GetStr(data, "", "cmd")
			if len(cmdline) == 0 {
				_ = message.Error(500, "no command line")
				return
			}
			var cmd = exec.Command(cmdline)
			cmd.Env = append(os.Environ(),
				"FOO=duplicate_value", // ignored
				"FOO=actual_value",    // this value is used
			)
			var sync = util.GetBool(data, false, "sync")
			var err error
			if sync {
				err = cmd.Run()
			} else {
				err = cmd.Start()
			}
			if err == nil {
				_ = message.Reply(0, "[success] "+cmdline)
			} else {
				_ = message.Error(500, "[fail] "+cmdline+" | "+err.Error())
			}

		}))

	}

	return guide

}

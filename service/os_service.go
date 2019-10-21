package service

import (
	"bytes"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qservice/qtiny"
	"log"
	"os"
	"os/exec"
)

func start(_ qtiny.TinyGuideEvent, tiny qtiny.TinyKind, _ qtiny.TinyGuideKind, _ map[string]interface{}, _ *qtiny.Future, err error) {

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
		var stdout, stderr *bytes.Buffer
		if sync {
			stdout = &bytes.Buffer{}
			stderr = &bytes.Buffer{}
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			err = cmd.Run()
		} else {
			err = cmd.Start()
		}

		if sync {
			if stderr != nil && stderr.Len() > 0 {
				_ = message.Error(500, string(stderr.Bytes()))
				return
			} else if err == nil && stdout != nil {
				_ = message.Reply(0, string(stdout.Bytes()))
				return
			}
		}

		if err == nil {
			_ = message.Reply(0, "[success] ["+tiny.GetTina().GetNodeId()+"] "+cmdline)
		} else {
			_ = message.Error(500, "[fail] ["+tiny.GetTina().GetNodeId()+"] "+cmdline+" | "+err.Error())
		}

	}))

}

func OsTinyGuide() *qtiny.TinyGuide {

	var guide = &qtiny.TinyGuide{}

	guide.Start = start
	return guide

}

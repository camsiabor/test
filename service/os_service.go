package service

import (
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
			_ = message.Error(500, "i am wrong")

			cmd := exec.Command("prog")
			cmd.Env = append(os.Environ(),
				"FOO=duplicate_value", // ignored
				"FOO=actual_value",    // this value is used
			)
			if err := cmd.Run(); err != nil {
				log.Fatal(err)
			}

		}))

	}

	return guide

}

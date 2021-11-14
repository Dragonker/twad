package main

import (
	"os"

	"github.com/Dragonker/twad/base"
	"github.com/Dragonker/twad/rofimode"
	"github.com/Dragonker/twad/tui"
)

func main() {
	args := os.Args[1:]

	base.Config()

	for _, v := range args {
		switch v {
		case "--rofi":
			rofimode.RunRofiMode("rofi")
			return
		case "--dmenu":
			rofimode.RunRofiMode("dmenu")
			return
		}

	}
	//cfg.GetInstance().Configured = false
	tui.Draw()
}

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/zeindevs/urlclip"
)

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("urlclip")
	systray.SetTooltip("URL Reader from Clipboard")

	systray.AddMenuItem("[urlclip] URL Clipboard Monitor", "").Disable()
	systray.AddSeparator()
	mActive := systray.AddMenuItemCheckbox("Disable", "Monitor Clipboard", true)
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the program")

	active := true
	quitCh := make(chan os.Signal, 1)

	go urlclip.Run(quitCh, &active)

	quitTrayCh := make(chan os.Signal, 1)
	signal.Notify(quitTrayCh, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-mActive.ClickedCh:
			if mActive.Checked() {
				mActive.Uncheck()
				mActive.SetTitle("Enable")
			} else {
				mActive.Check()
				mActive.SetTitle("Disable")
			}
			active = mActive.Checked()
			log.Println("active", mActive.Checked())
		case <-mQuit.ClickedCh:
			quitCh <- os.Interrupt
			systray.Quit()
		case <-quitTrayCh:
			quitCh <- os.Interrupt
			systray.Quit()
		}
	}
}

func onExit() {
	log.Println("systray terminated")
}

func main() {
	systray.Run(onReady, onExit)
}

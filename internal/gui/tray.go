package gui

import (
	"context"
	_ "embed"
	"log"

	"fyne.io/systray"
)

//go:embed tray_icon.png
var iconData []byte

type TrayManager struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewTrayManager() *TrayManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &TrayManager{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (t *TrayManager) Run(onReady func(), onExit func()) {
	systray.Run(func() {
		t.setupTray()
		if onReady != nil {
			onReady()
		}
	}, func() {
		if onExit != nil {
			onExit()
		}
	})
}

func (t *TrayManager) setupTray() {
	// Set icon
	systray.SetIcon(iconData)
	systray.SetTitle("pushbulleter")
	systray.SetTooltip("pushbulleter")

	// Add menu items
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				log.Println("Quit clicked")
				t.cancel()
				systray.Quit()
				return

			case <-t.ctx.Done():
				systray.Quit()
				return
			}
		}
	}()
}

func (t *TrayManager) Stop() {
	t.cancel()
}

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
	systray.SetTitle("Pushbullet Client")
	systray.SetTooltip("Pushbullet Client - Connected")

	// Add menu items
	mShow := systray.AddMenuItem("Show Events", "Show event window")
	systray.AddSeparator()
	mSettings := systray.AddMenuItem("Settings", "Open settings")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				log.Println("Show events clicked")
				// TODO: Implement show events window

			case <-mSettings.ClickedCh:
				log.Println("Settings clicked")
				// TODO: Implement settings window

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

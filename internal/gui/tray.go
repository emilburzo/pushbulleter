package gui

import (
	"context"
	"log"

	"fyne.io/systray"
)

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
	// Set icon (you'll need to embed an icon file)
	systray.SetIcon(getIconData())
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

// getIconData returns embedded icon data
// For now, return a simple placeholder
func getIconData() []byte {
	// This is a simple 16x16 PNG icon data (placeholder)
	// In a real implementation, you'd embed a proper icon file
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x91, 0x68, 0x36, 0x00, 0x00, 0x00,
		0x19, 0x74, 0x45, 0x58, 0x74, 0x53, 0x6F, 0x66, 0x74, 0x77, 0x61, 0x72,
		0x65, 0x00, 0x41, 0x64, 0x6F, 0x62, 0x65, 0x20, 0x49, 0x6D, 0x61, 0x67,
		0x65, 0x52, 0x65, 0x61, 0x64, 0x79, 0x71, 0xC9, 0x65, 0x3C, 0x00, 0x00,
		0x00, 0x2E, 0x49, 0x44, 0x41, 0x54, 0x78, 0xDA, 0x62, 0xFC, 0x3F, 0x95,
		0x9F, 0x01, 0x37, 0x60, 0x62, 0xC0, 0x0B, 0x46, 0xAA, 0x34, 0x40, 0x80,
		0x01, 0x00, 0x06, 0x50, 0x35, 0x30, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45,
		0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
}

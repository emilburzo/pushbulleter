# Pushbulleter for Linux/XFCE

A modern Go-based Pushbullet client specifically designed for Linux with XFCE desktop environment. Features optimized desktop notifications, system tray integration, and end-to-end encryption support.

## Features

- **XFCE-optimized notifications**: Enhanced desktop notifications with proper urgency levels, extended display times, and sound alerts
- **System tray integration**: Runs quietly in the background with a system tray icon
- **End-to-end encryption**: Full support for Pushbullet's E2E encryption
- **XDG compliance**: Follows Linux desktop standards for configuration and autostart
- **Native Linux integration**: Uses notify-send for notifications, follows XFCE conventions
- **Autostart support**: Automatic startup on login with proper desktop entry

## Installation

### Prerequisites

- Go 1.21 or later
- Linux with XFCE desktop environment
- libnotify-bin package for desktop notifications
- xfce4-notifyd (XFCE notification daemon)

### Build from source

```bash
git clone <repository-url>
cd pushbulleter
go mod download
go build -o pushbulleter cmd/pushbulleter/main.go
```

### Install

```bash
# Install to system
sudo cp pushbulleter /usr/local/bin/

# Or install to user bin
mkdir -p ~/.local/bin
cp pushbulleter ~/.local/bin/
```

Make sure `~/.local/bin` is in your PATH if using user installation.

## Configuration

The application uses XDG-compliant configuration. The config file is located at:
`$XDG_CONFIG_HOME/pushbulleter/config.yaml` (usually `~/.config/pushbulleter/config.yaml`)

### Initial setup

1. Get your API key from [Pushbullet Account Settings](https://www.pushbullet.com/#settings/account)
2. Run the application once to generate the default config file
3. Edit the config file and add your API key:

```yaml
api_key: "your_api_key_here"
e2e_enabled: false
e2e_key: ""
notifications:
  enabled: true
  show_mirrors: true
  show_sms: true
  show_calls: true
  filters: []
gui:
  show_tray_icon: true
  start_minimized: false
autostart: false
```

### End-to-end encryption

To enable E2E encryption:

1. Set `e2e_enabled: true` in the config
2. Set your encryption password in `e2e_key`
3. Restart the application

## Usage

### GUI mode (default)

```bash
pushbulleter
```

This starts the application with a system tray icon.

### Daemon mode

```bash
pushbulleter -daemon
```

This runs the application in the background without GUI.

### Custom config file

```bash
pushbulleter -config /path/to/config.yaml
```

### Autostart

To enable automatic startup on login, set `autostart: true` in the config file. This will create a desktop entry in `~/.config/autostart/`.

## Notifications

The client shows XFCE-optimized desktop notifications for:

- **Incoming calls** (📞) - Critical urgency, 25 second display, phone ring sound
- **SMS messages** (💬) - Critical urgency, 18 second display, message sound
- **Mirrored notifications** from Android devices - Normal urgency, 10 second display
- **Notes and links** sent to your devices - Normal urgency, 10 second display
- **File shares** (📎) - Normal urgency, 10 second display

All notifications use appropriate icons and categories for better XFCE integration. You can customize which notifications to show in the config file.

### Notification Requirements

- `libnotify-bin` - provides the `notify-send` command
- `xfce4-notifyd` - XFCE notification daemon (usually pre-installed)

Install missing packages:
```bash
sudo apt install libnotify-bin xfce4-notifyd
```

## System Tray

Right-click the tray icon to access:
- Show recent events
- Settings (planned)
- Quit application

## Development

### Project structure

```
├── cmd/pushbulleter/    # Main application entry point
├── internal/
│   ├── app/                  # Application logic
│   ├── config/               # Configuration management
│   ├── gui/                  # GUI components (tray, windows)
│   ├── notifications/        # Notification handling
│   └── pushbullet/          # Pushbullet API client
├── go.mod
└── README.md
```

### Dependencies

- `github.com/getlantern/systray` - System tray integration
- `github.com/gorilla/websocket` - WebSocket client for real-time stream
- `golang.org/x/crypto` - Cryptographic functions for E2E encryption
- `gopkg.in/yaml.v3` - YAML configuration parsing

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Troubleshooting

### No notifications appearing

1. Check if required packages are installed: `sudo apt install libnotify-bin xfce4-notifyd`
2. Test notifications: `notify-send "Test" "This is a test notification"`
3. Restart XFCE notification daemon: `killall xfce4-notifyd && xfce4-notifyd &`
4. Check XFCE notification settings: Settings → Notifications
5. Check the application logs for errors

### Connection issues

1. Verify your API key is correct
2. Check your internet connection
3. Look for firewall issues blocking WebSocket connections

### Autostart not working

1. Check if the desktop entry was created: `ls ~/.config/autostart/`
2. Verify the executable path in the desktop entry
3. Check your desktop environment's autostart settings

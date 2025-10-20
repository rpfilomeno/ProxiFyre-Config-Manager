# ProxiFyre Configuration Manager

A modern Windows GUI application for managing [ProxiFyre](https://github.com/wiresock/proxifyre) SOCKS5 proxy configurations with built-in service restart capabilities.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.16-blue.svg)

<img width="880" height="608" alt="Main Window" src="https://github.com/user-attachments/assets/d18b396b-6fca-4e9e-9794-c455c6b329a2" />
<img width="253" height="152" alt="SysTray" src="https://github.com/user-attachments/assets/f3581825-4827-443e-b335-1512a12b2bf7" />



## Features

- ✨ **Intuitive GUI** - Easy-to-use graphical interface for configuration management
- 🔄 **Service Control** - Restart ProxiFyre service directly from the app
- 📝 **Multiple Proxies** - Manage unlimited proxy configurations
- 🎯 **Process Exclusions** - Specify applications to bypass the proxy (v2.1.1+)
- ✅ **Real-time Validation** - Ensures configuration integrity
- 💾 **Auto-save** - Changes are preserved when switching between proxies
- 🔐 **Password Support** - Secure handling of proxy credentials



## Prerequisites

### For Users (Running the Application)
- Windows 10/11
- [ProxiFyre](https://github.com/wiresock/proxifyre) installed
- Administrator privileges (required for service restart)

### For Developers (Building from Source)
- Go 1.16 or later
- Fyne v2 dependencies

## Installation

### Option 1: Download Pre-built Binary (Recommended)
1. Download the latest release from the [Releases](../../releases) page
2. Extract `ProxiFyreConfig.exe` to your ProxiFyre installation directory
3. Run as Administrator

### Option 2: Build from Source

#### Windows

1. **Install Go**
   - Download from [golang.org](https://golang.org/dl/)
   - Verify installation: `go version`

2. **Install Dependencies**
   ```bash
   # Install Fyne dependencies for Windows
   go install fyne.io/fyne/v2/cmd/fyne@latest
   ```

3. **Clone and Build**
   ```bash
   git clone <repository-url>
   cd proxifyre-config-manager

   # Initialize Go module
   go mod init proxifyre-config
   go mod tidy

   # Build the application
   go build -o ProxiFyreConfig.exe

   # Or build with Windows manifest for proper admin elevation
   go build -ldflags="-H windowsgui" -o ProxiFyreConfig.exe
   ```

4. **Place the executable** in the same directory as `ProxiFyre.exe`

## Usage

### Starting the Application

1. Navigate to your ProxiFyre installation directory
2. Right-click `ProxiFyreConfig.exe` and select **Run as Administrator**
3. The application will automatically load `app-config.json` from the current directory

### Managing Proxies

#### Adding a New Proxy
1. Click **Add Proxy** button
2. Enter application names (one per line):
   - Use partial names: `firefox` matches `firefox.exe`
   - Use full paths: `C:\Program Files\WindowsApps\ROBLOX`
3. Configure SOCKS5 endpoint: `proxy.example.com:1080`
4. Add credentials (optional): username and password
5. Select protocols: TCP and/or UDP

#### Editing Existing Proxies
1. Select a proxy from the list
2. Modify any settings in the right panel
3. Changes are auto-saved when switching to another proxy

#### Removing a Proxy
1. Select the proxy you want to remove
2. Click **Remove Proxy**
3. Confirm deletion

### Managing Exclusions

Exclusions allow you to specify applications that should **bypass** the proxy:

1. Scroll to the **Excluded Applications** section
2. Enter application names or paths (one per line):
   ```
   firefox
   edge
   C:\Program Files\LocalApp\NotProxiedApp.exe
   ```
3. Save configuration

### Saving Configuration

#### Save Only
- Click **Save Configuration**
- Changes are written to `app-config.json`
- Service continues running with old configuration

#### Save and Restart Service
- Click **Save & Restart Service**
- Configuration is saved
- ProxiFyre service is automatically restarted
- New configuration takes effect immediately

### Reloading Configuration

- Click **Reload from File** to discard unsaved changes
- Useful for reverting to the last saved state

## Configuration File Format

The application manages `app-config.json` in the following format:

```json
{
  "logLevel": "Error",
  "proxies": [
    {
      "appNames": ["chrome", "firefox"],
      "socks5ProxyEndpoint": "proxy.example.com:1080",
      "username": "myuser",
      "password": "mypass",
      "supportedProtocols": ["TCP", "UDP"]
    }
  ],
  "excludes": [
    "edge",
    "localservice.exe"
  ]
}
```

### Configuration Options

| Field | Type | Description |
|-------|------|-------------|
| `logLevel` | string | Logging verbosity: `Error`, `Warning`, `Info`, `Debug`, `All` |
| `proxies` | array | List of proxy configurations |
| `appNames` | array | Application names or paths to proxy |
| `socks5ProxyEndpoint` | string | Proxy server address and port |
| `username` | string | Optional proxy authentication username |
| `password` | string | Optional proxy authentication password |
| `supportedProtocols` | array | Protocols to proxy: `TCP`, `UDP`, or both |
| `excludes` | array | Applications to bypass the proxy |

## Troubleshooting

### Application won't start
- Ensure you're running as Administrator
- Check that `app-config.json` exists (it will be created if missing)
- Verify ProxiFyre is installed in the same directory

### Service restart fails
- Confirm `ProxiFyre.exe` exists in the same directory
- Verify you have Administrator privileges
- Check if ProxiFyre service is installed: `ProxiFyre.exe install`

### Configuration not taking effect
- Use **Save & Restart Service** instead of just **Save Configuration**
- Verify the service is running: `sc query ProxiFyre`
- Check ProxiFyre logs in `/logs` directory

### Proxy not working for specific applications
- Ensure application names are correct (case-insensitive)
- For UWP apps, use the full path to the WindowsApps folder
- Check if the application is in the exclusions list

## Advanced Usage

### Running from Different Directory

If you want to manage a config file in a different location:

```bash
# Edit the configPath variable in main.go before building
var configPath = "C:\\Path\\To\\app-config.json"
```

### Multiple ProxiFyre Instances

You can manage multiple ProxiFyre configurations by:
1. Creating separate folders for each configuration
2. Placing a copy of `ProxiFyreConfig.exe` in each folder
3. Each instance will manage its own `app-config.json`

## Development

### Project Structure

```
proxifyre-config-manager/
├── main.go           # Main application code
├── go.mod            # Go module dependencies
├── go.sum            # Dependency checksums
├── ProxyFyre.png     # App Icon
└── README.md         # This file
```

### Building for Release

```bash
# Build optimized binary
go build -ldflags="-s -w -H windowsgui" -o ProxiFyreConfig.exe

# Create release package
mkdir release
copy ProxiFyreConfig.exe release\
copy README.md release\
```

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with [Fyne](https://fyne.io/) - A cross-platform GUI toolkit for Go
- Created for [ProxiFyre](https://github.com/wiresock/proxifyre) by Wiresock

## Support

For issues related to:
- **This GUI application**: Open an issue in this repository
- **ProxiFyre itself**: Visit the [ProxiFyre repository](https://github.com/wiresock/proxifyre)

## Changelog

### v1.0.0 (Initial Release)
- ✨ Full configuration management GUI
- 🔄 Service restart integration
- 📝 Multiple proxy support
- 🎯 Process exclusion management (v2.1.1+ feature)
- 💾 Real-time configuration validation
- 🔐 Secure password handling

---

**Note**: This is an unofficial GUI tool for ProxiFyre. For the official ProxiFyre documentation and support, please visit the [official repository](https://github.com/wiresock/proxifyre).

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Config structures matching ProxiFyre configuration
type ProxyConfig struct {
	AppNames            []string `json:"appNames"`
	Socks5ProxyEndpoint string   `json:"socks5ProxyEndpoint"`
	Username            string   `json:"username,omitempty"`
	Password            string   `json:"password,omitempty"`
	SupportedProtocols  []string `json:"supportedProtocols"`
}

type AppConfig struct {
	LogLevel string        `json:"logLevel"`
	Proxies  []ProxyConfig `json:"proxies"`
	Excludes []string      `json:"excludes"`
}

var configPath = "app-config.json"
var embeddedIcon []byte

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("ProxiFyre Configuration Manager")
	myWindow.Resize(fyne.NewSize(900, 700))

	iconResource, err := fyne.LoadResourceFromPath("ProxiFyre.png")
	if err != nil {
		// Handle error, e.g., print to console or use a default icon
		panic(err)
	}

	myApp.SetIcon(iconResource)
	myWindow.SetIcon(iconResource)

	if desk, ok := myApp.(desktop.App); ok {

		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Show Window", func() {
				// Logic to show or restore the main window
				myWindow.Show()
			}),
			fyne.NewMenuItem("Install Service", func() {
				myWindow.Show()
				dialog.ShowConfirm("Install as Service", "This will install ProxiFyre as a service. Continue?", func(confirmed bool) {
					if confirmed {
						if err := installService(); err != nil {
							dialog.ShowError(fmt.Errorf("Failed to install service: %v", err), myWindow)
						} else {
							dialog.ShowInformation("Success", "Service installed successfully!", myWindow)
						}
					}
				}, myWindow)
			}),
			fyne.NewMenuItem("Uninstall Service", func() {
				myWindow.Show()
				dialog.ShowConfirm("Uninstall as Service", "This will uninstall ProxiFyre as a service. Continue?", func(confirmed bool) {
					if confirmed {
						if err := uninstallService(); err != nil {
							dialog.ShowError(fmt.Errorf("Failed to uninstall service: %v", err), myWindow)
						} else {
							dialog.ShowInformation("Success", "Service uninstalled successfully!", myWindow)
						}
					}
				}, myWindow)
			}),

			fyne.NewMenuItem("Quit", func() {
				myWindow.Show()
				dialog.ShowConfirm("Quit Manager", "This will quit ProxiFyre Manager. Continue?", func(confirmed bool) {
					if confirmed {
						myApp.Quit() // Quits the entire Fyne application
					}
				}, myWindow)

			}),
		)
		desk.SetSystemTrayMenu(m)
	}

	myWindow.SetCloseIntercept(func() {
		myWindow.Hide()
	})

	// Load existing config
	config := loadConfig()

	// Create UI elements
	logLevelSelect := widget.NewSelect(
		[]string{"Error", "Warning", "Info", "Debug", "All"},
		func(value string) {
			config.LogLevel = value
		},
	)
	logLevelSelect.SetSelected(config.LogLevel)

	// Proxy list
	proxyList := widget.NewList(
		func() int { return len(config.Proxies) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Proxy")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(fmt.Sprintf("Proxy %d: %s", id+1, config.Proxies[id].Socks5ProxyEndpoint))
		},
	)

	var selectedProxyIndex int = -1
	var proxyEditor *fyne.Container

	// Proxy editor
	appNamesEntry := widget.NewMultiLineEntry()
	appNamesEntry.SetPlaceHolder("One app per line (e.g., firefox, chrome)")
	appNamesEntry.Wrapping = fyne.TextWrapWord
	appNamesEntry.SetMinRowsVisible(3)

	endpointEntry := widget.NewEntry()
	endpointEntry.SetPlaceHolder("proxy.example.com:1080")

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Optional username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Optional password")

	tcpCheck := widget.NewCheck("TCP", nil)
	udpCheck := widget.NewCheck("UDP", nil)

	excludesEntry := widget.NewMultiLineEntry()
	excludesEntry.SetPlaceHolder("One app per line to exclude from proxy")
	excludesEntry.Wrapping = fyne.TextWrapWord
	excludesEntry.SetMinRowsVisible(3)
	if len(config.Excludes) > 0 {
		excludesEntry.SetText(strings.Join(config.Excludes, "\n"))
	}

	updateProxyEditor := func(index int) {
		if index < 0 || index >= len(config.Proxies) {
			appNamesEntry.SetText("")
			endpointEntry.SetText("")
			usernameEntry.SetText("")
			passwordEntry.SetText("")
			tcpCheck.SetChecked(false)
			udpCheck.SetChecked(false)
			return
		}

		proxy := config.Proxies[index]
		appNamesEntry.SetText(strings.Join(proxy.AppNames, "\n"))
		endpointEntry.SetText(proxy.Socks5ProxyEndpoint)
		usernameEntry.SetText(proxy.Username)
		passwordEntry.SetText(proxy.Password)

		tcpCheck.SetChecked(false)
		udpCheck.SetChecked(false)
		for _, proto := range proxy.SupportedProtocols {
			if strings.ToUpper(proto) == "TCP" {
				tcpCheck.SetChecked(true)
			}
			if strings.ToUpper(proto) == "UDP" {
				udpCheck.SetChecked(true)
			}
		}
	}

	saveProxyChanges := func() {
		if selectedProxyIndex < 0 || selectedProxyIndex >= len(config.Proxies) {
			return
		}

		appNames := strings.Split(appNamesEntry.Text, "\n")
		var cleanedAppNames []string
		for _, name := range appNames {
			trimmed := strings.TrimSpace(name)
			if trimmed != "" {
				cleanedAppNames = append(cleanedAppNames, trimmed)
			}
		}

		var protocols []string
		if tcpCheck.Checked {
			protocols = append(protocols, "TCP")
		}
		if udpCheck.Checked {
			protocols = append(protocols, "UDP")
		}

		config.Proxies[selectedProxyIndex] = ProxyConfig{
			AppNames:            cleanedAppNames,
			Socks5ProxyEndpoint: endpointEntry.Text,
			Username:            usernameEntry.Text,
			Password:            passwordEntry.Text,
			SupportedProtocols:  protocols,
		}

		proxyList.Refresh()
	}

	proxyList.OnSelected = func(id widget.ListItemID) {
		saveProxyChanges()
		selectedProxyIndex = id
		updateProxyEditor(id)
	}

	// Buttons
	addProxyBtn := widget.NewButton("Add Proxy", func() {
		saveProxyChanges()
		newProxy := ProxyConfig{
			AppNames:            []string{},
			Socks5ProxyEndpoint: "",
			SupportedProtocols:  []string{"TCP"},
		}
		config.Proxies = append(config.Proxies, newProxy)
		selectedProxyIndex = len(config.Proxies) - 1
		proxyList.Refresh()
		proxyList.Select(selectedProxyIndex)
	})

	removeProxyBtn := widget.NewButton("Remove Proxy", func() {
		if selectedProxyIndex < 0 || selectedProxyIndex >= len(config.Proxies) {
			dialog.ShowInformation("Error", "Please select a proxy to remove", myWindow)
			return
		}

		dialog.ShowConfirm("Confirm Delete", "Are you sure you want to delete this proxy?", func(confirmed bool) {
			if confirmed {
				config.Proxies = append(config.Proxies[:selectedProxyIndex], config.Proxies[selectedProxyIndex+1:]...)
				selectedProxyIndex = -1
				proxyList.Refresh()
				updateProxyEditor(-1)
			}
		}, myWindow)
	})

	saveBtn := widget.NewButton("Save Configuration", func() {
		saveProxyChanges()

		// Update excludes
		excludeLines := strings.Split(excludesEntry.Text, "\n")
		var cleanedExcludes []string
		for _, line := range excludeLines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				cleanedExcludes = append(cleanedExcludes, trimmed)
			}
		}
		config.Excludes = cleanedExcludes

		if err := saveConfig(config); err != nil {
			dialog.ShowError(fmt.Errorf("Failed to save configuration: %v", err), myWindow)
			return
		}

		dialog.ShowInformation("Success", "Configuration saved successfully!", myWindow)
	})

	saveAndRestartBtn := widget.NewButton("Save & Restart Service", func() {
		saveProxyChanges()

		// Update excludes
		excludeLines := strings.Split(excludesEntry.Text, "\n")
		var cleanedExcludes []string
		for _, line := range excludeLines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				cleanedExcludes = append(cleanedExcludes, trimmed)
			}
		}
		config.Excludes = cleanedExcludes

		if err := saveConfig(config); err != nil {
			dialog.ShowError(fmt.Errorf("Failed to save configuration: %v", err), myWindow)
			return
		}

		dialog.ShowConfirm("Restart Service", "Configuration saved. Restart ProxiFyre service now?", func(confirmed bool) {
			if confirmed {
				if err := restartService(); err != nil {
					dialog.ShowError(fmt.Errorf("Failed to restart service: %v", err), myWindow)
				} else {
					dialog.ShowInformation("Success", "Service restarted successfully!", myWindow)
				}
			}
		}, myWindow)
	})

	exitBtn := widget.NewButton("Exit Manager", func() {
		dialog.ShowConfirm("Confirm Exit", "Are you sure you want to exit this manager?", func(confirmed bool) {
			if confirmed {
				myApp.Quit()
			}
		}, myWindow)
	})

	loadBtn := widget.NewButton("Reload from File", func() {
		dialog.ShowConfirm("Reload Configuration", "This will discard unsaved changes. Continue?", func(confirmed bool) {
			if confirmed {
				config = loadConfig()
				logLevelSelect.SetSelected(config.LogLevel)
				selectedProxyIndex = -1
				proxyList.Refresh()
				updateProxyEditor(-1)
				if len(config.Excludes) > 0 {
					excludesEntry.SetText(strings.Join(config.Excludes, "\n"))
				} else {
					excludesEntry.SetText("")
				}
				dialog.ShowInformation("Success", "Configuration reloaded!", myWindow)
			}
		}, myWindow)
	})

	// Layout
	proxyEditor = container.NewVBox(
		widget.NewLabel("Proxy Configuration"),
		widget.NewForm(
			widget.NewFormItem("Application Names", appNamesEntry),
			widget.NewFormItem("SOCKS5 Endpoint", endpointEntry),
			widget.NewFormItem("Username", usernameEntry),
			widget.NewFormItem("Password", passwordEntry),
		),
		widget.NewLabel("Supported Protocols:"),
		container.NewHBox(tcpCheck, udpCheck),
	)

	leftPanel := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Proxy List"),
			container.NewHBox(addProxyBtn, removeProxyBtn),
		),
		nil,
		nil,
		nil,
		proxyList,
	)

	rightPanel := container.NewVBox(
		proxyEditor,
		widget.NewSeparator(),
		widget.NewLabel("Global Excluded Applications"),
		excludesEntry,
	)

	split := container.NewHSplit(leftPanel, container.NewScroll(rightPanel))
	split.SetOffset(0.3)

	topBar := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Log Level", logLevelSelect),
		),
		widget.NewSeparator(),
	)

	bottomBar := container.NewHBox(
		widget.NewLabel(""), // spacer
		loadBtn,
		saveBtn,
		saveAndRestartBtn,
		layout.NewSpacer(), //fill spacer
		exitBtn,
		widget.NewLabel(""), // spacer
	)

	content := container.NewBorder(
		topBar,
		bottomBar,
		nil,
		nil,
		split,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func loadConfig() AppConfig {
	config := AppConfig{
		LogLevel: "Error",
		Proxies:  []ProxyConfig{},
		Excludes: []string{},
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Warning: Could not read config file: %v\n", err)
		}
		return config
	}

	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("Warning: Could not parse config file: %v\n", err)
		return config
	}

	return config
}

func saveConfig(config AppConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func restartService() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %v", err)
	}

	proxifyreDir := filepath.Dir(exePath)
	proxifyreExe := filepath.Join(proxifyreDir, "ProxiFyre.exe")

	// Check if ProxiFyre.exe exists
	if _, err := os.Stat(proxifyreExe); os.IsNotExist(err) {
		return fmt.Errorf("ProxiFyre.exe not found in %s", proxifyreDir)
	}

	// Stop the service
	stopCmd := exec.Command(proxifyreExe, "stop")
	if err := stopCmd.Run(); err != nil {
		return fmt.Errorf("failed to stop service: %v", err)
	}

	// Start the service
	startCmd := exec.Command(proxifyreExe, "start")
	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("failed to start service: %v", err)
	}

	return nil
}

func installService() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %v", err)
	}

	proxifyreDir := filepath.Dir(exePath)
	proxifyreExe := filepath.Join(proxifyreDir, "ProxiFyre.exe")

	// Check if ProxiFyre.exe exists
	if _, err := os.Stat(proxifyreExe); os.IsNotExist(err) {
		return fmt.Errorf("ProxiFyre.exe not found in %s", proxifyreDir)
	}

	// Stop the service
	installCmd := exec.Command(proxifyreExe, "install")
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install service: %v", err)
	}

	// Start the service
	startCmd := exec.Command(proxifyreExe, "start")
	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("failed to start service: %v", err)
	}

	return nil
}

func uninstallService() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %v", err)
	}

	proxifyreDir := filepath.Dir(exePath)
	proxifyreExe := filepath.Join(proxifyreDir, "ProxiFyre.exe")

	// Check if ProxiFyre.exe exists
	if _, err := os.Stat(proxifyreExe); os.IsNotExist(err) {
		return fmt.Errorf("ProxiFyre.exe not found in %s", proxifyreDir)
	}

	// Stop the service
	stopCmd := exec.Command(proxifyreExe, "stop")
	if err := stopCmd.Run(); err != nil {
		return fmt.Errorf("failed to stop service: %v", err)
	}

	// Stop the service
	uninstallCmd := exec.Command(proxifyreExe, "uninstall")
	if err := uninstallCmd.Run(); err != nil {
		return fmt.Errorf("failed to uninstall service: %v", err)
	}

	return nil
}

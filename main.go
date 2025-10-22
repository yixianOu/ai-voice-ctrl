package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create custom logger to suppress runtime errors
	customLogger := logger.NewDefaultLogger()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "ai-voice-ctrl",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Logger: customLogger,
		Linux: &linux.Options{
			// Use WebKit2GTK for better compatibility
			WebviewGpuPolicy: linux.WebviewGpuPolicyOnDemand,
			// Enable input method for better text editing
			WindowIsTranslucent: false,
		},
	})
	if err != nil {
		log.Fatal("Error:", err.Error())
	}
}

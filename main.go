package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/mattn/go-ieproxy"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/xmdhs/player-go/cors"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.Proxy = ieproxy.GetProxyFunc()

	app := &App{
		t: t,
	}
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	err = wails.Run(&options.App{
		Title:      "player",
		Width:      800,
		Height:     600,
		Assets:     &assets,
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewUserDataPath: path.Join(pwd, "data"),
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

type App struct {
	ctx context.Context
	t   *http.Transport
}

func (b *App) startup(ctx context.Context) {
	b.ctx = ctx
}

func (b *App) shutdown(ctx context.Context) {}

func (b *App) CorsServer() int {
	return cors.Server(b.ctx, b.t)
}

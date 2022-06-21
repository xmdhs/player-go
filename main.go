//go:build !browser

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
	"github.com/xmdhs/player-go/api"
	"github.com/xmdhs/player-go/cors"
	"go.etcd.io/bbolt"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.Proxy = ieproxy.GetProxyFunc()

	db, err := bbolt.Open("player.db", 0600, nil)
	if err != nil {
		panic(err)
	}
	app := &App{
		t:  t,
		db: db,
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
	db  *bbolt.DB
}

func (b *App) startup(ctx context.Context) {
	b.ctx = ctx
}

func (b *App) shutdown(ctx context.Context) {}

func (b *App) CorsServer() int {
	return cors.Server(b.ctx, b.t)
}

func (b *App) ApiServer() int {
	return api.Server(b.ctx, b.db, b.t)
}

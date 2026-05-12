package main

import (
	"embed"
	"fmt"
	"siren/core"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Atualizado para o diretório de build do Nuxt 3
//go:embed all:frontend/.output/public
var assets embed.FS

//go:embed build/appicon.png
var iconData []byte

func main() {
	// Inicializa o núcleo do Siren (Persistência e Motor de Áudio)
	store, err := core.NewStore()
	if err != nil {
		fmt.Printf("Erro crítico ao carregar configurações: %v\n", err)
		return
	}

	engine := core.NewEngine()
	manager := core.NewManager(store, engine)

	// Cria uma instância da estrutura do service com o orquestrador injetado
	appService := NewApp(manager)

	// Inicializa a aplicação Wails v3
	app := application.New(application.Options{
		Name:        "Siren",
		Description: "Siren Audio Tunnel Daemon",
		Services: []application.Service{
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// Configuração da Janela Principal (v3 API corrigida)
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Siren",
		Width:  1024,
		Height: 768,
	})

	// Configuração do System Tray Nativo (v3 API corrigida)
	tray := app.SystemTray.New()
	tray.SetIcon(iconData)

	// Criação do Menu do Tray
	trayMenu := app.NewMenu()
	trayMenu.Add("Mostrar Siren").OnClick(func(ctx *application.Context) {
		window.Show()
		window.Focus()
	})
	trayMenu.Add("Desconectar Túnel").OnClick(func(ctx *application.Context) {
		appService.StopTunnel()
	})
	trayMenu.AddSeparator()
	trayMenu.Add("Sair").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	tray.SetMenu(trayMenu)

	// Executa a aplicação
	err = app.Run()
	if err != nil {
		fmt.Printf("Erro na execução do Siren: %v\n", err)
	}
}
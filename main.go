package main

import (
	"embed"
	"fmt"
	"siren/core"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// Atualizado para o diretório de build do Nuxt 3
//go:embed all:frontend/.output/public
var assets embed.FS

func main() {
	// Inicializa o núcleo do Siren (Persistência e Motor de Áudio)
	store, err := core.NewStore()
	if err != nil {
		fmt.Printf("Erro crítico ao carregar configurações: %v\n", err)
		return
	}

	engine := core.NewEngine()
	manager := core.NewManager(store, engine)

	// Cria uma instância da estrutura da aplicação com o orquestrador injetado
	app := NewApp(manager)

	// Cria a aplicação com as opções do Wails
	err = wails.Run(&options.App{
		Title:             "Siren",
		Width:             1024,
		Height:            768,
		HideWindowOnClose: true, // Mantém o app rodando oculto ao fechar
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app, // Vincula os métodos de App.go para o Javascript
		},
	})

	if err != nil {
		fmt.Println("Erro na inicialização da GUI:", err.Error())
	}
}
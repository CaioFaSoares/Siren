package main

import (
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// Atualizado para o diretório de build do Nuxt 3
//go:embed all:frontend/.output/public
var assets embed.FS

func main() {
	// Cria uma instância da estrutura da aplicação
	app := NewApp()

	// Cria a aplicação com as opções do Wails
	err := wails.Run(&options.App{
		Title:  "Siren",
		Width:  1024,
		Height: 768,
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
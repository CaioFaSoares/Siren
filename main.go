package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// Atualizado para o diretório de build do Nuxt 3
//go:embed all:frontend/.output/public
var assets embed.FS

func main() {
	// Comando Raiz
	rootCmd := &cobra.Command{
		Use:   "siren",
		Short: "Siren - Ferramenta de Tunneling de Áudio",
		Long:  `Siren é uma aplicação híbrida para tunneling de áudio, operando via GUI ou CLI.`,
		// A ação padrão se o usuário rodar apenas `siren` (sem argumentos) é abrir a interface gráfica
		Run: func(cmd *cobra.Command, args []string) {
			iniciarGUI()
		},
	}

	// Subcomando CLI (Ex: `siren tunnel`)
	tunnelCmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Inicia o túnel de áudio via terminal (sem GUI)",
		Run: func(cmd *cobra.Command, args []string) {
			// Futura integração com internal/audio e Viper
			fmt.Println("🎧 Iniciando Siren Audio Tunnel no modo CLI (Headless)...")
			fmt.Println("Pressione Ctrl+C para encerrar.")
			
			// Esse select impede que o programa feche imediatamente.
			// Útil para processos como servidores ou escutas de áudio contínuas.
			select {} 
		},
	}

	// Anexa o subcomando ao comando raiz
	rootCmd.AddCommand(tunnelCmd)

	// Executa o avaliador de comandos do Cobra
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Erro fatal ao iniciar:", err)
		os.Exit(1)
	}
}

// A inicialização do Wails isolada na sua própria função
func iniciarGUI() {
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
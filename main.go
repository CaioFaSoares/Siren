package main

import (
	"embed"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"siren/core"

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

	// Agrupador de comandos de túnel
	tunnelCmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Gerencia o túnel de áudio via terminal (Headless)",
	}

	// Subcomando: Start
	var localNodeOverride string
	var remoteNodeOverride string
	var tunnelMode string
	var verboseFlag bool

	startCmd := &cobra.Command{
		Use:   "start [device_id]",
		Short: "Inicia o túnel de áudio para um dispositivo específico",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.Verbose = verboseFlag
			deviceID := args[0]

			// Inicialização dos componentes do Core
			store, err := core.NewStore()
			if err != nil {
				fmt.Printf("❌ Erro ao inicializar armazenamento: %v\n", err)
				return
			}

			engine := core.NewEngine()
			manager := core.NewManager(store, engine)

			// Buscar a configuração de túnel
			config, err := store.GetTunnelByDeviceID(deviceID)
			if err == nil {
				if localNodeOverride != "" {
					config.LocalNodeID = localNodeOverride
					fmt.Printf("🎯 Usando node local específico: %s\n", localNodeOverride)
				}
				if remoteNodeOverride != "" {
					config.RemoteNodeID = remoteNodeOverride
					fmt.Printf("🎯 Usando node remoto específico: %s\n", remoteNodeOverride)
				}
				if tunnelMode != "" {
					config.Mode = core.TunnelMode(tunnelMode)
					fmt.Printf("🔄 Modo de operação: %s\n", tunnelMode)
				}
			}

			fmt.Printf("🎧 Iniciando túnel Siren para o dispositivo: %s\n", deviceID)
			
			// Acionar a engine diretamente com o config modificado se necessário
			device, _ := store.GetDeviceByID(deviceID)
			if err := engine.Start(config, device.IP); err != nil {
				fmt.Printf("❌ Erro fatal: %v\n", err)
				return
			}

			fmt.Println("🚀 Túnel estabelecido com sucesso!")
			fmt.Println("📌 Pressione Ctrl+C para encerrar o processo.")

			// Canal para capturar sinais de interrupção (Ctrl+C)
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			// Bloqueia aqui até receber o sinal
			<-sigChan

			fmt.Println("\n🛑 Sinal de interrupção recebido. Limpando recursos...")
			if err := manager.StopCurrentTunnel(); err != nil {
				fmt.Printf("⚠️ Erro ao encerrar motor de áudio: %v\n", err)
			}
			fmt.Println("✅ Siren finalizado. Até logo!")
		},
	}

	// Subcomando: Stop (Informativo por enquanto)
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Instruções para encerrar o túnel",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ℹ️  No modo CLI atual, o túnel deve ser encerrado com Ctrl+C no terminal de origem.")
		},
	}

	// Agrupador de comandos de dispositivos
	deviceCmd := &cobra.Command{
		Use:   "device",
		Short: "Gerencia o inventário de dispositivos",
	}

	// Subcomando: Device List
	deviceListCmd := &cobra.Command{
		Use:   "list",
		Short: "Lista todos os dispositivos cadastrados",
		Run: func(cmd *cobra.Command, args []string) {
			store, _ := core.NewStore()
			engine := core.NewEngine()
			manager := core.NewManager(store, engine)

			devices := manager.GetAvailableDevices()
			if len(devices) == 0 {
				fmt.Println("📭 Nenhum dispositivo cadastrado.")
				return
			}

			fmt.Println("📋 Dispositivos Cadastrados:")
			fmt.Printf("%-10s %-20s %-15s %-10s\n", "ID", "NOME", "IP", "PLATAFORMA")
			fmt.Println(strings.Repeat("-", 60))
			for _, d := range devices {
				fmt.Printf("%-10s %-20s %-15s %-10s\n", d.ID, d.Name, d.IP, d.Platform)
			}
		},
	}

	// Subcomando: Device Add
	deviceAddCmd := &cobra.Command{
		Use:   "add [nome] [ip] [plataforma]",
		Short: "Cadastra um novo dispositivo",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			ip := args[1]
			platform := args[2]

			store, _ := core.NewStore()
			engine := core.NewEngine()
			manager := core.NewManager(store, engine)

			id, err := manager.AddDevice(name, ip, platform)
			if err != nil {
				fmt.Printf("❌ Erro: %v\n", err)
				return
			}

			fmt.Printf("✅ Dispositivo '%s' adicionado com sucesso!\n", name)
			fmt.Printf("🔑 ID Gerado: %s\n", id)
		},
	}

	// Subcomando: Device Remove
	deviceRemoveCmd := &cobra.Command{
		Use:   "remove [id]",
		Short: "Remove um dispositivo pelo ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]

			store, _ := core.NewStore()
			engine := core.NewEngine()
			manager := core.NewManager(store, engine)

			if err := manager.RemoveDevice(id); err != nil {
				fmt.Printf("❌ Erro ao remover: %v\n", err)
				return
			}

			fmt.Printf("🗑️  Dispositivo %s removido.\n", id)
		},
	}

	// Agrupador de comandos de nodes locais
	nodeCmd := &cobra.Command{
		Use:   "node",
		Short: "Gerencia nodes de áudio locais (microfones/alto-falantes)",
	}

	// Subcomando: Node List
	nodeListCmd := &cobra.Command{
		Use:   "list [source|sink]",
		Short: "Lista nodes de áudio locais",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			nodeType := args[0]
			store, _ := core.NewStore()
			engine := core.NewEngine()
			manager := core.NewManager(store, engine)

			nodes, err := manager.GetLocalNodes(nodeType)
			if err != nil {
				fmt.Printf("❌ Erro ao listar nodes: %v\n", err)
				return
			}

			if len(nodes) == 0 {
				fmt.Printf("📭 Nenhum node do tipo '%s' encontrado.\n", nodeType)
				return
			}

			fmt.Printf("📋 Nodes Locais (%s):\n", nodeType)
			fmt.Printf("%-10s %-40s %-10s\n", "ID", "NOME", "PADRÃO")
			fmt.Println(strings.Repeat("-", 60))
			for _, n := range nodes {
				isDefault := ""
				if n.IsDefault {
					isDefault = "✅"
				}
				fmt.Printf("%-10s %-40s %-10s\n", n.ID, n.Name, isDefault)
			}
		},
	}

	// Organiza a árvore de comandos
	nodeCmd.AddCommand(nodeListCmd)
	rootCmd.AddCommand(nodeCmd)

	deviceCmd.AddCommand(deviceListCmd, deviceAddCmd, deviceRemoveCmd)
	rootCmd.AddCommand(deviceCmd)

	startCmd.Flags().StringVarP(&localNodeOverride, "node", "n", "", "ID do node de áudio local a ser usado (Source)")
	startCmd.Flags().StringVarP(&remoteNodeOverride, "remote-node", "r", "", "ID do node de áudio remoto a ser usado (Sink)")
	startCmd.Flags().StringVarP(&tunnelMode, "mode", "m", "duplex", "Modo do túnel (sender, receiver, duplex)")
	startCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Habilita a verbosidade de logs (comandos rodados por baixo dos panos)")
	tunnelCmd.AddCommand(startCmd)
	tunnelCmd.AddCommand(stopCmd)
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
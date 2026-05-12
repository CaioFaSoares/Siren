package main

import (
	"context"
	"fmt"
	"os/exec"
	stdRuntime "runtime"
	"siren/core"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	Ctx     context.Context
	manager *core.Manager
}

// NewApp creates a new App application struct
func NewApp(manager *core.Manager) *App {
	return &App{
		manager: manager,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.Ctx = ctx
}

// GetDevices retorna a lista de dispositivos cadastrados no orquestrador
func (a *App) GetDevices() []core.Device {
	return a.manager.GetAvailableDevices()
}

// AddDevice cadastra um novo dispositivo
func (a *App) AddDevice(name, ip, platform string) error {
	_, err := a.manager.AddDevice(name, ip, platform)
	return err
}

// RemoveDevice remove um dispositivo pelo ID
func (a *App) RemoveDevice(id string) error {
	return a.manager.RemoveDevice(id)
}

// GetLocalInputs retorna a lista de dispositivos de entrada de áudio locais
func (a *App) GetLocalInputs() []core.AudioNode {
	nodes, err := a.manager.GetLocalNodes("source")
	if err != nil {
		fmt.Printf("Erro ao obter inputs locais: %v\n", err)
		return []core.AudioNode{}
	}
	return nodes
}

// GetLocalOutputs retorna a lista de dispositivos de saída de áudio locais
func (a *App) GetLocalOutputs() []core.AudioNode {
	nodes, err := a.manager.GetLocalNodes("sink")
	if err != nil {
		fmt.Printf("Erro ao obter outputs locais: %v\n", err)
		return []core.AudioNode{}
	}
	return nodes
}

// StartTunnel configura e inicia um túnel de áudio para um dispositivo remoto
func (a *App) StartTunnel(deviceID string, mode string, localNodeID string, remoteNodeID string) error {
	store := a.manager.GetStore()

	// 1. Obter a configuração atual do túnel
	config, err := store.GetTunnelByDeviceID(deviceID)
	if err != nil {
		return fmt.Errorf("configuração de túnel não encontrada: %w", err)
	}

	// 2. Atualizar parâmetros conforme selecionado na UI
	config.Mode = core.TunnelMode(mode)
	config.LocalNodeID = localNodeID
	config.RemoteNodeID = remoteNodeID

	// 3. Salvar configuração atualizada
	if err := store.SaveTunnel(config); err != nil {
		return fmt.Errorf("falha ao salvar configuração: %w", err)
	}

	// 4. Iniciar o motor de áudio
	if err := a.manager.StartTunnelToDevice(deviceID); err != nil {
		return err
	}

	// 5. Emitir evento de reatividade para o frontend
	wailsRuntime.EventsEmit(a.Ctx, "tunnel-status", true)

	return nil
}

// StopTunnel encerra o túnel de áudio ativo
func (a *App) StopTunnel() error {
	if err := a.manager.StopCurrentTunnel(); err != nil {
		return err
	}

	// Notificar o frontend
	wailsRuntime.EventsEmit(a.Ctx, "tunnel-status", false)

	return nil
}

// CheckSystemRequirements verifica se os binários necessários estão presentes no sistema
func (a *App) CheckSystemRequirements() map[string]bool {
	res := make(map[string]bool)

	if stdRuntime.GOOS == "linux" {
		_, err := exec.LookPath("pw-cli")
		res["pw-cli"] = err == nil
	} else if stdRuntime.GOOS == "darwin" {
		_, err := exec.LookPath("roc-send")
		res["roc-send"] = err == nil

		_, err = exec.LookPath("roc-recv")
		res["roc-recv"] = err == nil
	}

	return res
}


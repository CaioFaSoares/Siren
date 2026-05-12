package main

import (
	"context"
	"os/exec"
	"runtime"
	"siren/core"
)

// App struct
type App struct {
	ctx     context.Context
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
	a.ctx = ctx
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

// CheckSystemRequirements verifica se os binários necessários estão presentes no sistema
func (a *App) CheckSystemRequirements() map[string]bool {
	res := make(map[string]bool)

	if runtime.GOOS == "linux" {
		_, err := exec.LookPath("pw-cli")
		res["pw-cli"] = err == nil
	} else if runtime.GOOS == "darwin" {
		_, err := exec.LookPath("roc-send")
		res["roc-send"] = err == nil

		_, err = exec.LookPath("roc-recv")
		res["roc-recv"] = err == nil
	}

	return res
}

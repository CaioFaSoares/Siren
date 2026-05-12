package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Manager atua como o orquestrador central do Siren, conectando Store e Engine
type Manager struct {
	store  *Store
	engine AudioEngine
}

// NewManager cria uma nova instância do orquestrador
func NewManager(store *Store, engine AudioEngine) *Manager {
	return &Manager{
		store:  store,
		engine: engine,
	}
}

// StartTunnelToDevice inicia um túnel de áudio para um dispositivo específico
func (m *Manager) StartTunnelToDevice(deviceID string) error {
	// 1. Buscar o dispositivo no Store
	device, err := m.store.GetDeviceByID(deviceID)
	if err != nil {
		return fmt.Errorf("erro ao localizar dispositivo: %w", err)
	}

	// 2. Buscar a configuração de túnel associada
	config, err := m.store.GetTunnelByDeviceID(deviceID)
	if err != nil {
		return fmt.Errorf("erro ao localizar configuração de túnel: %w", err)
	}

	// TODO: No futuro, podemos atualizar o IP do config dinamicamente 
	// se o IP do device tiver mudado (ZeroTier/Local)

	// 3. Acionar a engine
	if err := m.engine.Start(config, device.IP); err != nil {
		return fmt.Errorf("erro ao iniciar motor de áudio para %s (%s): %w", device.Name, device.IP, err)
	}

	return nil
}

// StopCurrentTunnel encerra qualquer túnel que esteja ativo
func (m *Manager) StopCurrentTunnel() error {
	return m.engine.Stop()
}

// GetAvailableDevices retorna a lista de dispositivos configurados
func (m *Manager) GetAvailableDevices() []Device {
	return m.store.GetDevices()
}

// GetLocalNodes retorna a lista de hardware local (microfones/alto-falantes)
func (m *Manager) GetLocalNodes(nodeType string) ([]AudioNode, error) {
	if nodeType == "source" {
		return m.engine.GetInputs()
	}
	return m.engine.GetOutputs()
}

// AddDevice cadastra um novo dispositivo e cria uma configuração de túnel padrão
func (m *Manager) AddDevice(name, ip string, platform string) (string, error) {
	// Validar plataforma
	p := Platform(strings.ToLower(platform))
	if p != PlatformDarwin && p != PlatformLinux {
		return "", fmt.Errorf("plataforma inválida: %s (use 'linux' ou 'darwin')", platform)
	}

	// Gerar ID curto (8 caracteres)
	id := uuid.New().String()[:8]

	dev := Device{
		ID:       id,
		Name:     name,
		IP:       ip,
		Platform: p,
		LastSeen: time.Now().Format(time.RFC3339),
	}

	// Salvar o dispositivo
	if err := m.store.SaveDevice(dev); err != nil {
		return "", fmt.Errorf("erro ao salvar dispositivo: %w", err)
	}

	// Criar e salvar uma configuração de túnel padrão para este dispositivo
	config := TunnelConfig{
		RemoteDeviceID: id,
		Mode:           ModeDuplex,
		SourcePort:     10001,
		RepairPort:     10002,
		ControlPort:    10003,
		RxSourcePort:   10003,
		RxRepairPort:   10004,
		Active:         false,
	}

	if err := m.store.SaveTunnel(config); err != nil {
		return id, fmt.Errorf("dispositivo salvo, mas erro ao criar configuração de túnel: %w", err)
	}

	return id, nil
}

// RemoveDevice remove um dispositivo e sua configuração de túnel
func (m *Manager) RemoveDevice(id string) error {
	// TODO: No futuro, remover também a configuração de túnel associada
	return m.store.RemoveDevice(id)
}

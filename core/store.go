package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

// Store gerencia a persistência de dados do Siren usando Viper
type Store struct {
	v  *viper.Viper
	mu sync.RWMutex
}

// NewStore inicializa o gerenciador de configuração em ~/.config/siren/config.json
func NewStore() (*Store, error) {
	v := viper.New()

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter diretório home: %w", err)
	}

	configDir := filepath.Join(home, ".config", "siren")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de configuração: %w", err)
	}

	v.SetConfigName("config")
	v.SetConfigType("json")
	v.AddConfigPath(configDir)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Define valores padrão se o arquivo não existir
			v.Set("devices", []Device{})
			v.Set("tunnels", []TunnelConfig{})
			if err := v.SafeWriteConfig(); err != nil {
				return nil, fmt.Errorf("erro ao criar arquivo de configuração inicial: %w", err)
			}
		} else {
			return nil, fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
		}
	}

	return &Store{v: v}, nil
}

// getDevicesInternal lê dispositivos do viper sem travar o mutex (uso interno)
func (s *Store) getDevicesInternal() []Device {
	var devices []Device
	err := s.v.UnmarshalKey("devices", &devices)
	if err != nil {
		return []Device{}
	}
	return devices
}

// GetDevices retorna a lista de todos os dispositivos salvos
func (s *Store) GetDevices() []Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getDevicesInternal()
}

// SaveDevice adiciona ou atualiza um dispositivo no armazenamento
func (s *Store) SaveDevice(device Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	devices := s.getDevicesInternal()
	found := false
	for i, d := range devices {
		if d.ID == device.ID {
			devices[i] = device
			found = true
			break
		}
	}

	if !found {
		devices = append(devices, device)
	}

	s.v.Set("devices", devices)
	return s.v.WriteConfig()
}

// RemoveDevice remove um dispositivo pelo seu ID
func (s *Store) RemoveDevice(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	devices := s.getDevicesInternal()
	var newDevices []Device
	for _, d := range devices {
		if d.ID != id {
			newDevices = append(newDevices, d)
		}
	}

	s.v.Set("devices", newDevices)
	return s.v.WriteConfig()
}

// getTunnelsInternal lê túneis do viper sem travar o mutex (uso interno)
func (s *Store) getTunnelsInternal() []TunnelConfig {
	var tunnels []TunnelConfig
	err := s.v.UnmarshalKey("tunnels", &tunnels)
	if err != nil {
		return []TunnelConfig{}
	}
	return tunnels
}

// GetTunnels retorna a lista de todas as configurações de túnel
func (s *Store) GetTunnels() []TunnelConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getTunnelsInternal()
}

// SaveTunnel salva ou atualiza a configuração de um túnel
func (s *Store) SaveTunnel(config TunnelConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tunnels := s.getTunnelsInternal()
	found := false
	for i, t := range tunnels {
		if t.RemoteDeviceID == config.RemoteDeviceID {
			tunnels[i] = config
			found = true
			break
		}
	}

	if !found {
		tunnels = append(tunnels, config)
	}

	s.v.Set("tunnels", tunnels)
	return s.v.WriteConfig()
}

// GetDeviceByID busca um dispositivo pelo seu ID
func (s *Store) GetDeviceByID(id string) (Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := s.getDevicesInternal()
	for _, d := range devices {
		if d.ID == id {
			return d, nil
		}
	}
	return Device{}, fmt.Errorf("dispositivo com ID %s não encontrado", id)
}

// GetTunnelByDeviceID busca a configuração de túnel associada a um dispositivo remoto
func (s *Store) GetTunnelByDeviceID(deviceID string) (TunnelConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tunnels := s.getTunnelsInternal()
	for _, t := range tunnels {
		if t.RemoteDeviceID == deviceID {
			return t, nil
		}
	}
	return TunnelConfig{}, fmt.Errorf("configuração de túnel para o dispositivo %s não encontrada", deviceID)
}

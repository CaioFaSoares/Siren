package core

import "time"

type Platform string

const (
	PlatformDarwin Platform = "darwin"
	PlatformLinux  Platform = "linux"
)

type NodeType string

const (
	SourceNode NodeType = "source" // Input (Microfone)
	SinkNode   NodeType = "sink"   // Output (Alto-falante)
)

// Device representa um computador na rede (Local ou VPN ZeroTier/Tailscale)
type Device struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	IP       string    `json:"ip"`
	Platform Platform  `json:"platform"`
	IsLocal  bool      `json:"is_local"`
	LastSeen time.Time `json:"last_seen"`
}

// AudioNode representa uma interface de áudio física (microfone ou alto-falante)
type AudioNode struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Type      NodeType `json:"type"`
	IsDefault bool     `json:"is_default"`
}

// TunnelConfig contém as configurações específicas do protocolo ROC para o túnel
type TunnelConfig struct {
	RemoteDeviceID string `json:"remote_device_id"`
	LocalNodeID    string `json:"local_node_id"`
	RemoteNodeID   string `json:"remote_node_id"`

	// ROC Ports
	SourcePort  int `json:"source_port"`  // Default: 10001
	RepairPort  int `json:"repair_port"`  // Default: 10002
	ControlPort int `json:"control_port"` // Default: 10003

	Active bool `json:"active"`
}
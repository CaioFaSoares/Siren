package core


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
	ID       string    `json:"id" mapstructure:"id"`
	Name     string    `json:"name" mapstructure:"name"`
	IP       string    `json:"ip" mapstructure:"ip"`
	Platform Platform  `json:"platform" mapstructure:"platform"`
	IsLocal  bool      `json:"is_local" mapstructure:"is_local"`
	LastSeen string `json:"last_seen" mapstructure:"last_seen"`
}

// AudioNode representa uma interface de áudio física (microfone ou alto-falante)
type AudioNode struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Type      NodeType `json:"type"`
	IsDefault bool     `json:"is_default"`
}

type TunnelMode string

const (
	ModeSender   TunnelMode = "sender"   // Envia áudio local para remoto
	ModeReceiver TunnelMode = "receiver" // Recebe áudio remoto para local
	ModeDuplex   TunnelMode = "duplex"   // Envia e recebe simultaneamente
)

// TunnelConfig contém as configurações específicas do protocolo ROC para o túnel
type TunnelConfig struct {
	RemoteDeviceID string `json:"remote_device_id" mapstructure:"remote_device_id"`
	LocalNodeID    string `json:"local_node_id" mapstructure:"local_node_id"`
	RemoteNodeID   string `json:"remote_node_id" mapstructure:"remote_node_id"`
	Mode           TunnelMode `json:"mode" mapstructure:"mode"`

	// ROC Ports (TX - Transmissão)
	SourcePort  int `json:"source_port" mapstructure:"source_port"`  // Default: 10001
	RepairPort  int `json:"repair_port" mapstructure:"repair_port"`  // Default: 10002
	ControlPort int `json:"control_port" mapstructure:"control_port"` // Default: 10003

	// ROC Ports (RX - Recepção, usados para Duplex)
	RxSourcePort int `json:"rx_source_port" mapstructure:"rx_source_port"` // Default: 10003
	RxRepairPort int `json:"rx_repair_port" mapstructure:"rx_repair_port"` // Default: 10004

	Active bool `json:"active" mapstructure:"active"`
}
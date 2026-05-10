// internal/config/models.go
type Device struct {
    ID       string `json:"id"`
    Name     string `json:"name"`     // Ex: "MacBook Pro" ou "Yamori Linux"
    IP       string `json:"ip"`       // IP local ou da ZeroTier/Tailscale
    Platform string `json:"platform"` // "darwin" ou "linux"
}

type TunnelConfig struct {
    TargetDeviceID string `json:"target_device_id"`
    SourcePort     int    `json:"source_port"` // Padrão: 10001
    RepairPort     int    `json:"repair_port"` // Padrão: 10002
}
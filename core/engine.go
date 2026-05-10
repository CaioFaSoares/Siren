// internal/audio/engine.go
type Engine interface {
    Start(targetIP string, targetPlatform string) error
    Stop() error
    Status() string
}
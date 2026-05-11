package core

// AudioEngine define o contrato para os motores de áudio específicos de cada SO
type AudioEngine interface {
	// Start inicia o roteamento de áudio baseado na configuração do túnel
	Start(config TunnelConfig) error

	// Stop encerra o túnel ativo e limpa recursos
	Stop() error

	// GetInputs retorna os dispositivos de entrada (microfones) disponíveis
	GetInputs() ([]AudioNode, error)

	// GetOutputs retorna os dispositivos de saída (alto-falantes) disponíveis
	GetOutputs() ([]AudioNode, error)
}

// NewEngine é uma factory que retorna a implementação correta baseada no SO
func NewEngine() AudioEngine {
	return newOSSpecificEngine()
}
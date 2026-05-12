//go:build darwin

package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

type darwinEngine struct {
	cancel context.CancelFunc
	mu     sync.Mutex
}

func newOSSpecificEngine() AudioEngine {
	return &darwinEngine{}
}

func (e *darwinEngine) Start(config TunnelConfig, targetIP string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cancel != nil {
		return fmt.Errorf("um túnel já está ativo")
	}

	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel

	// Definir o URI de entrada. O ROC no macOS geralmente espera coreaudio://default
	// para o dispositivo padrão. Se falhar, precisaremos de um nome específico.
	inputURI := "coreaudio://default"
	if config.LocalNodeID != "" && config.LocalNodeID != "default" {
		inputURI = fmt.Sprintf("coreaudio://%s", config.LocalNodeID)
	}

	args := []string{
		"-vv", // Aumentar verbosidade para diagnóstico
		"-s", fmt.Sprintf("rtp+rs8m://%s:%d", targetIP, config.SourcePort),
		"-r", fmt.Sprintf("rs8m://%s:%d", targetIP, config.RepairPort),
		"-i", inputURI,
	}

	cmd := exec.CommandContext(ctx, "roc-send", args...)

	// Rodar em background
	go func() {
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil && ctx.Err() == nil {
			fmt.Printf("Erro no processo roc-send: %v (Stderr: %s)\n", err, stderr.String())
		}
	}()

	return nil
}

func (e *darwinEngine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cancel != nil {
		e.cancel()
		e.cancel = nil
	}
	return nil
}

func (e *darwinEngine) GetInputs() ([]AudioNode, error) {
	return e.listDevices("roc-send")
}

func (e *darwinEngine) GetOutputs() ([]AudioNode, error) {
	return e.listDevices("roc-recv")
}

func (e *darwinEngine) listDevices(binary string) ([]AudioNode, error) {
	// O ROC lista dispositivos usando --list-devices
	// roc-send -i coreaudio --list-devices
	cmd := exec.Command(binary, "-i", "coreaudio://default", "--list-devices")
	if binary == "roc-recv" {
		cmd = exec.Command(binary, "-o", "coreaudio://default", "--list-devices")
	}

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("erro ao listar dispositivos via %s: %w", binary, err)
	}

	var nodes []AudioNode
	scanner := bufio.NewScanner(&out)
	
	// Exemplo de saída esperada:
	//   * default (Default)
	//     45 (Built-in Microphone)
	
	re := regexp.MustCompile(`^\s*(\*?\s*)(\S+)\s+\((.+)\)`)
	nodeType := SourceNode
	if binary == "roc-recv" {
		nodeType = SinkNode
	}

	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindStringSubmatch(line)
		if len(match) > 3 {
			isDefault := strings.Contains(match[1], "*")
			id := match[2]
			name := match[3]

			nodes = append(nodes, AudioNode{
				ID:        id,
				Name:      name,
				Type:      nodeType,
				IsDefault: isDefault,
			})
		}
	}

	return nodes, nil
}

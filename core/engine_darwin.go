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

	// Tentar resolver o nome real do dispositivo padrão se LocalNodeID for "default" ou vazio
	inputURI := "coreaudio://default"
	if config.LocalNodeID != "" && config.LocalNodeID != "default" {
		inputURI = fmt.Sprintf("coreaudio://%s", config.LocalNodeID)
	} else {
		// Busca a lista de dispositivos para encontrar qual é o padrão
		nodes, err := e.listDevices("roc-send")
		if err == nil {
			for _, n := range nodes {
				if n.IsDefault {
					inputURI = fmt.Sprintf("coreaudio://%s", n.ID)
					break
				}
			}
		}
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
	// No macOS, o ROC não lista dispositivos de forma confiável via CLI.
	// Usamos o system_profiler que é nativo e garantido.
	cmd := exec.Command("system_profiler", "SPAudioDataType")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("erro ao executar system_profiler: %w", err)
	}

	var nodes []AudioNode
	scanner := bufio.NewScanner(&out)
	
	nodeType := SourceNode
	if binary == "roc-recv" {
		nodeType = SinkNode
	}

	/*
	Exemplo de saída:
        K66:
          Default Input Device: Yes
          Input Channels: 2
	*/

	var currentDeviceName string
	var isDefault bool
	var hasInput bool
	var hasOutput bool

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detectar nome do dispositivo (termina com :)
		if strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, "Source:") && !strings.Contains(trimmed, "Device:") {
			// Salvar dispositivo anterior
			if currentDeviceName != "" {
				if (nodeType == SourceNode && hasInput) || (nodeType == SinkNode && hasOutput) {
					nodes = append(nodes, AudioNode{
						ID:        currentDeviceName,
						Name:      currentDeviceName,
						Type:      nodeType,
						IsDefault: isDefault,
					})
				}
			}
			// Reset para o novo dispositivo
			currentDeviceName = strings.TrimSuffix(trimmed, ":")
			isDefault = false
			hasInput = false
			hasOutput = false
			continue
		}

		if strings.Contains(line, "Default Input Device: Yes") || strings.Contains(line, "Default Output Device: Yes") {
			isDefault = true
		}
		if strings.Contains(line, "Input Channels:") {
			hasInput = true
		}
		if strings.Contains(line, "Output Channels:") {
			hasOutput = true
		}
	}

	// Adicionar o último dispositivo
	if currentDeviceName != "" {
		if (nodeType == SourceNode && hasInput) || (nodeType == SinkNode && hasOutput) {
			nodes = append(nodes, AudioNode{
				ID:        currentDeviceName,
				Name:      currentDeviceName,
				Type:      nodeType,
				IsDefault: isDefault,
			})
		}
	}

	return nodes, nil
}

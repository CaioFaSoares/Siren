//go:build darwin

package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
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

	// Montar argumentos para roc-send (Transmissão: Mac -> Linux)
	if config.Mode == ModeSender || config.Mode == ModeDuplex {
		sendArgs := []string{
			"-vv",
			"-s", fmt.Sprintf("rtp+rs8m://%s:%d", targetIP, config.SourcePort),
			"-r", fmt.Sprintf("rs8m://%s:%d", targetIP, config.RepairPort),
		}

		// Se houver um ID específico, adiciona a flag -i. Caso contrário, omite para usar o padrão do ROC sem crash.
		if config.LocalNodeID != "" && config.LocalNodeID != "default" {
			sendArgs = append(sendArgs, "-i", fmt.Sprintf("coreaudio://%s", config.LocalNodeID))
		}

		sendCmd := exec.CommandContext(ctx, "roc-send", sendArgs...)
		
		if Verbose {
			fmt.Printf("🔍 [ROC Exec] %s\n", strings.Join(sendCmd.Args, " "))
			sendCmd.Stdout = os.Stdout
			sendCmd.Stderr = os.Stderr
		}
		
		go func() {
			if !Verbose {
				var stderr bytes.Buffer
				sendCmd.Stderr = &stderr
				if err := sendCmd.Run(); err != nil && ctx.Err() == nil {
					fmt.Printf("Erro no processo roc-send: %v (Stderr: %s)\n", err, stderr.String())
				}
			} else {
				if err := sendCmd.Run(); err != nil && ctx.Err() == nil {
					fmt.Printf("Erro no processo roc-send: %v\n", err)
				}
			}
		}()
	}

	// Montar argumentos para roc-recv (Recepção: Linux -> Mac)
	if config.Mode == ModeReceiver || config.Mode == ModeDuplex {
		recvArgs := []string{
			"-vv",
			"-s", fmt.Sprintf("rtp+rs8m://0.0.0.0:%d", config.RxSourcePort),
			"-r", fmt.Sprintf("rs8m://0.0.0.0:%d", config.RxRepairPort),
		}

		// Se houver um ID de output específico, adiciona a flag -o.
		if config.RemoteNodeID != "" && config.RemoteNodeID != "default" {
			recvArgs = append(recvArgs, "-o", fmt.Sprintf("coreaudio://%s", config.RemoteNodeID))
		}

		recvCmd := exec.CommandContext(ctx, "roc-recv", recvArgs...)

		if Verbose {
			fmt.Printf("🔍 [ROC Exec] %s\n", strings.Join(recvCmd.Args, " "))
			recvCmd.Stdout = os.Stdout
			recvCmd.Stderr = os.Stderr
		}

		go func() {
			if !Verbose {
				var stderr bytes.Buffer
				recvCmd.Stderr = &stderr
				if err := recvCmd.Run(); err != nil && ctx.Err() == nil {
					fmt.Printf("Erro no processo roc-recv: %v (Stderr: %s)\n", err, stderr.String())
				}
			} else {
				if err := recvCmd.Run(); err != nil && ctx.Err() == nil {
					fmt.Printf("Erro no processo roc-recv: %v\n", err)
				}
			}
		}()
	}

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

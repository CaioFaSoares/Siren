//go:build linux

package core

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

type linuxEngine struct {
	activeModuleID string
	mu             sync.Mutex
}

func newOSSpecificEngine() AudioEngine {
	return &linuxEngine{}
}

func (e *linuxEngine) Start(config TunnelConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.activeModuleID != "" {
		return fmt.Errorf("um túnel já está ativo (ID: %s)", e.activeModuleID)
	}

	// Comando para carregar o módulo ROC Sink (envio)
	// No PipeWire, carregamos libpipewire-module-roc-sink
	cmdArgs := []string{
		"load-module",
		"libpipewire-module-roc-sink",
		fmt.Sprintf("remote.ip=%s", "127.0.0.1"), // TODO: Usar IP do config quando disponível
		fmt.Sprintf("remote.source.port=%d", config.SourcePort),
		fmt.Sprintf("remote.repair.port=%d", config.RepairPort),
		fmt.Sprintf("remote.control.port=%d", config.ControlPort),
		"sink.name=Siren-Sink",
		"node.description=Siren Audio Tunnel",
	}

	cmd := exec.Command("pw-cli", cmdArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao carregar módulo PipeWire: %w (Saída: %s)", err, out.String())
	}

	// Regex para capturar o ID do módulo na saída do pw-cli
	// A saída costuma ser apenas o ID ou algo como "module: 123"
	re := regexp.MustCompile(`(\d+)`)
	match := re.FindStringSubmatch(out.String())
	if len(match) < 2 {
		return fmt.Errorf("não foi possível capturar o ID do módulo na saída: %s", out.String())
	}

	e.activeModuleID = match[1]
	return nil
}

func (e *linuxEngine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.activeModuleID == "" {
		return nil
	}

	cmd := exec.Command("pw-cli", "destroy", e.activeModuleID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao destruir módulo %s: %w", e.activeModuleID, err)
	}

	e.activeModuleID = ""
	return nil
}

func (e *linuxEngine) GetInputs() ([]AudioNode, error) {
	return e.listNodes("source")
}

func (e *linuxEngine) GetOutputs() ([]AudioNode, error) {
	return e.listNodes("sink")
}

// listNodes usa pw-cli para listar dispositivos de áudio
func (e *linuxEngine) listNodes(nodeType string) ([]AudioNode, error) {
	cmd := exec.Command("pw-cli", "ls", "Node")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("erro ao listar nodes: %w", err)
	}

	var nodes []AudioNode
	scanner := bufio.NewScanner(&out)
	
	var currentNode *AudioNode
	idRe := regexp.MustCompile(`id (\d+),`)
	nameRe := regexp.MustCompile(`node.name = "([^"]+)"`)
	descRe := regexp.MustCompile(`node.description = "([^"]+)"`)
	mediaClassRe := regexp.MustCompile(`media.class = "([^"]+)"`)

	for scanner.Scan() {
		line := scanner.Text()

		if match := idRe.FindStringSubmatch(line); len(match) > 1 {
			if currentNode != nil {
				nodes = append(nodes, *currentNode)
			}
			currentNode = &AudioNode{ID: match[1]}
		}

		if currentNode == nil {
			continue
		}

		if match := nameRe.FindStringSubmatch(line); len(match) > 1 {
			// Usamos o nome técnico como fallback se a descrição falhar
			if currentNode.Name == "" {
				currentNode.Name = match[1]
			}
		}

		if match := descRe.FindStringSubmatch(line); len(match) > 1 {
			currentNode.Name = match[1]
		}

		if match := mediaClassRe.FindStringSubmatch(line); len(match) > 1 {
			class := match[1]
			if strings.Contains(class, "Source") {
				currentNode.Type = SourceNode
			} else if strings.Contains(class, "Sink") {
				currentNode.Type = SinkNode
			}
		}
	}

	if currentNode != nil {
		nodes = append(nodes, *currentNode)
	}

	// Filtrar pelo tipo solicitado
	var filtered []AudioNode
	for _, n := range nodes {
		if (nodeType == "source" && n.Type == SourceNode) || (nodeType == "sink" && n.Type == SinkNode) {
			filtered = append(filtered, n)
		}
	}

	return filtered, nil
}

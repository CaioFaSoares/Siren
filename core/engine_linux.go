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
	activeModuleIDs []string
	mu             sync.Mutex
}

func newOSSpecificEngine() AudioEngine {
	return &linuxEngine{}
}

func (e *linuxEngine) Start(config TunnelConfig, targetIP string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.activeModuleIDs) > 0 {
		return fmt.Errorf("um túnel já está ativo (IDs: %v)", e.activeModuleIDs)
	}

	e.activeModuleIDs = []string{}

	loadModule := func(args []string) error {
		cmd := exec.Command("pw-cli", args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		if Verbose {
			fmt.Printf("🔍 [PipeWire Exec] %s\n", strings.Join(cmd.Args, " "))
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("erro ao carregar módulo %s: %w (Saída: %s)", args[1], err, out.String())
		}

		output := strings.TrimSpace(out.String())
		if output != "" {
			re := regexp.MustCompile(`(\d+)`)
			match := re.FindStringSubmatch(output)
			if len(match) >= 2 {
				e.activeModuleIDs = append(e.activeModuleIDs, match[1])
			}
		}
		return nil
	}

	// ModeSender ou Duplex: Linux -> Mac (Audio)
	if config.Mode == ModeSender || config.Mode == ModeDuplex {
		argsStr := fmt.Sprintf("remote.ip=%s remote.source.port=%d remote.repair.port=%d fec.code=rs8m sink.name=Siren_Audio", targetIP, config.RxSourcePort, config.RxRepairPort)
		if config.RemoteNodeID != "" && config.RemoteNodeID != "default" {
			argsStr += fmt.Sprintf(" node.target=%s", config.RemoteNodeID)
		}
		cmdArgs := []string{"load-module", "libpipewire-module-roc-sink", argsStr}

		if err := loadModule(cmdArgs); err != nil {
			return err
		}
	}

	// ModeReceiver ou Duplex: Mac -> Linux (Microfone)
	if config.Mode == ModeReceiver || config.Mode == ModeDuplex {
		argsStr := fmt.Sprintf("local.ip=0.0.0.0 local.source.port=%d local.repair.port=%d fec.code=rs8m source.name=Siren_Mic source.props={media.class=Audio/Source node.description=Siren_Incoming_Audio}", config.SourcePort, config.RepairPort)
		if config.LocalNodeID != "" && config.LocalNodeID != "default" {
			argsStr += fmt.Sprintf(" node.target=%s", config.LocalNodeID)
		}
		cmdArgs := []string{"load-module", "libpipewire-module-roc-source", argsStr}

		if err := loadModule(cmdArgs); err != nil {
			return err
		}
	}

	return nil
}

func (e *linuxEngine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var errors []string

	if len(e.activeModuleIDs) > 0 {
		for _, id := range e.activeModuleIDs {
			cmd := exec.Command("pw-cli", "destroy", id)
			if err := cmd.Run(); err != nil {
				errors = append(errors, fmt.Sprintf("erro ao destruir módulo ID %s: %v", id, err))
			}
		}
		e.activeModuleIDs = nil
	}

	// Fallback de segurança: Derrubar qualquer módulo órfão
	exec.Command("pkill", "-f", "libpipewire-module-roc-sink").Run()
	exec.Command("pkill", "-f", "libpipewire-module-roc-source").Run()

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}

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

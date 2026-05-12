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

	// Função interna que simula o script Bash: echo '...' | pw-cli
	loadModule := func(module string, argsStr string) error {
		pipeline := fmt.Sprintf("echo 'load-module %s %s' | pw-cli", module, argsStr)
		cmd := exec.Command("sh", "-c", pipeline)
		
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		if Verbose {
			fmt.Printf("🔍 [PipeWire Exec] %s\n", pipeline)
		}

		err := cmd.Run()
		output := strings.TrimSpace(out.String())

		if Verbose && output != "" {
			fmt.Printf("📄 [PipeWire Output] %s\n", output)
		}

		if err != nil {
			return fmt.Errorf("falha ao rodar shell: %w, saída: %s", err, output)
		}

		// O pw-cli sempre retorna o ID do módulo quando tem sucesso (ex: "123" ou "module: 123")
		re := regexp.MustCompile(`(\d+)`)
		match := re.FindStringSubmatch(output)
		if len(match) > 1 {
			e.activeModuleIDs = append(e.activeModuleIDs, match[1])
			return nil
		}

		// Se a regex não achou um ID, o módulo não foi carregado. 
		return fmt.Errorf("o PipeWire recusou o módulo. Saída: %s", output)
	}

	// 1. Iniciar Roteamento SENDER (Linux -> Mac)
	if config.Mode == ModeSender || config.Mode == ModeDuplex {
		argsStr := fmt.Sprintf("remote.ip=%s remote.source.port=%d remote.repair.port=%d fec.code=rs8m sink.name=Siren_Audio", targetIP, config.RxSourcePort, config.RxRepairPort)
		if config.RemoteNodeID != "" && config.RemoteNodeID != "default" {
			argsStr += fmt.Sprintf(" node.target=%s", config.RemoteNodeID)
		}
		if err := loadModule("libpipewire-module-roc-sink", argsStr); err != nil {
			e.Stop() // Limpa se o primeiro carregou mas o segundo falhar
			return fmt.Errorf("erro no envio (sink): %v", err)
		}
	}

	// 2. Iniciar Roteamento RECEIVER (Mac -> Linux)
	if config.Mode == ModeReceiver || config.Mode == ModeDuplex {
		// O uso das aspas duplas ao redor do source.props é vital para o parser do PipeWire
		argsStr := fmt.Sprintf(`local.ip=0.0.0.0 local.source.port=%d local.repair.port=%d fec.code=rs8m source.name=Siren_Mic source.props="{ media.class=Audio/Source node.description=Siren_Incoming_Audio }"`, config.SourcePort, config.RepairPort)
		if config.LocalNodeID != "" && config.LocalNodeID != "default" {
			argsStr += fmt.Sprintf(" node.target=%s", config.LocalNodeID)
		}
		if err := loadModule("libpipewire-module-roc-source", argsStr); err != nil {
			e.Stop()
			return fmt.Errorf("erro na recepção (source): %v", err)
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

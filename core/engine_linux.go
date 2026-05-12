//go:build linux

package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

type linuxEngine struct {
	cancel context.CancelFunc
	mu     sync.Mutex
}

func newOSSpecificEngine() AudioEngine {
	return &linuxEngine{}
}

func (e *linuxEngine) Start(config TunnelConfig, targetIP string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cancel != nil {
		return fmt.Errorf("um túnel já está ativo")
	}

	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel

	cmd := exec.CommandContext(ctx, "pw-cli")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		e.cancel()
		e.cancel = nil
		return fmt.Errorf("erro ao criar pipe para pw-cli: %v", err)
	}

	var out bytes.Buffer
	if Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &out
		cmd.Stderr = &out
	}

	if err := cmd.Start(); err != nil {
		e.cancel()
		e.cancel = nil
		return fmt.Errorf("erro ao iniciar processo pw-cli: %v", err)
	}

	// 1. Roteamento SENDER (Linux -> Mac)
	if config.Mode == ModeSender || config.Mode == ModeDuplex {
		argsStr := fmt.Sprintf("remote.ip=%s remote.source.port=%d remote.repair.port=%d fec.code=rs8m sink.name=Siren_Audio", targetIP, config.RxSourcePort, config.RxRepairPort)
		if config.RemoteNodeID != "" && config.RemoteNodeID != "default" {
			argsStr += fmt.Sprintf(" node.target=%s", config.RemoteNodeID)
		}
		cmdStr := fmt.Sprintf("load-module libpipewire-module-roc-sink %s\n", argsStr)
		if Verbose {
			fmt.Printf("🔍 [PipeWire Exec] Injetando no pw-cli:\n%s", cmdStr)
		}
		fmt.Fprintf(stdin, "%s", cmdStr)
	}

	// 2. Roteamento RECEIVER (Mac -> Linux)
	if config.Mode == ModeReceiver || config.Mode == ModeDuplex {
		argsStr := fmt.Sprintf(`local.ip=0.0.0.0 local.source.port=%d local.repair.port=%d fec.code=rs8m source.name=Siren_Mic source.props="{ media.class=Audio/Source node.description=Siren_Incoming_Audio }"`, config.SourcePort, config.RepairPort)
		if config.LocalNodeID != "" && config.LocalNodeID != "default" {
			argsStr += fmt.Sprintf(" node.target=%s", config.LocalNodeID)
		}
		cmdStr := fmt.Sprintf("load-module libpipewire-module-roc-source %s\n", argsStr)
		if Verbose {
			fmt.Printf("🔍 [PipeWire Exec] Injetando no pw-cli:\n%s", cmdStr)
		}
		fmt.Fprintf(stdin, "%s", cmdStr)
	}

	// IMPORTANTE: Não enviamos "quit" e não damos stdin.Close(). 
	// O pipe aberto é o que mantém o pw-cli rodando e os módulos ativos no sistema.

	// Monitória assíncrona
	go func() {
		err := cmd.Wait()
		if err != nil && ctx.Err() == nil {
			if !Verbose {
				fmt.Printf("Aviso: pw-cli encerrou inesperadamente: %v\nSaída: %s\n", err, out.String())
			} else {
				fmt.Printf("Aviso: pw-cli encerrou inesperadamente: %v\n", err)
			}
		}
	}()

	return nil
}

func (e *linuxEngine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cancel != nil {
		e.cancel() // O cancelamento mata o pw-cli, e o PipeWire remove os módulos automaticamente.
		e.cancel = nil
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

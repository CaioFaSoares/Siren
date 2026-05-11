# Siren - Master Plan

**Versão:** 1.0.0  
**Status:** Em Desenvolvimento  
**Autor:** Caio Soares & Siren AI  

## 1. Visão Geral
O **Siren** é uma solução de infraestrutura de áudio "any-to-any" projetada para rotear fluxos sonoros entre diferentes sistemas operacionais (foco inicial: macOS e Linux) com baixa latência, utilizando o protocolo ROC. O projeto se diferencia por ser um binário único que atua como CLI, Daemon (background) e GUI (Wails + Nuxt).

## 2. Pilha Tecnológica
- **Backend:** Go 1.26+ (Lógica de sistemas e CLI)
- **GUI Framework:** Wails v2 (WebView nativa)
- **CLI Framework:** Cobra (Interface de linha de comando)
- **Configuração:** Viper (Gerenciamento de estado e persistência)
- **Frontend:** Nuxt 3 + Nuxt UI (Tailwind CSS + Headless UI)
- **Infra/Build:** Nix (Flakes) para ambiente imutável e reprodutível.

## 3. Arquitetura do Sistema
O projeto adota uma arquitetura **Hexagonal (Ports and Adapters)** para garantir que a lógica de áudio seja independente da interface de controle.

### Estrutura de Diretórios
- `/core`: Domínio central. Contém os modelos de dados e a interface `AudioEngine`.
- `/frontend`: Interface rica em Nuxt.js.
- `/build`: Assets específicos de cada SO (ícones, manifestos).
- `main.go`: Ponto de entrada que roteia para CLI ou GUI.
- `app.go`: Bridge de comunicação Wails <-> Go.

## 4. Épicos de Desenvolvimento

### Épico 1: Core de Configuração e Estado
- **Objetivo:** Persistir dispositivos e preferências sem hardcoding.
- **Tasks:**
    - [ ] Implementar CRUD de `Device` no pacote `core`.
    - [ ] Integrar Viper para salvar `config.yaml` em pastas padrão do SO.
    - [ ] Criar lógica de detecção automática de IP (local/ZeroTier).

### Épico 2: Motores de Áudio (Engine)
- **Objetivo:** Implementar o tunneling real eliminando memory leaks.
- **Tasks:**
    - [ ] **Linux (PipeWire):** Implementar motor que captura ID do módulo e executa `pw-cli destroy` no Stop.
    - [ ] **macOS (Darwin):** Implementar motor que gerencia ciclos de vida dos processos `roc-send/recv` via `os/exec`.
    - [ ] **Hardware Selection:** Implementar listagem de sinks/sources físicos para roteamento dinâmico.

### Épico 3: Modo Daemon e Orquestração
- **Objetivo:** Permitir que o áudio rode sem janelas abertas.
- **Tasks:**
    - [ ] Implementar comando `siren tunnel start` via Cobra.
    - [ ] Adicionar suporte a System Tray no Wails para controle em background.
    - [ ] Implementar "Watchers" para hot-plugging de dispositivos de áudio.

### Épico 4: Ponte de Comunicação (Wails Bridge)
- **Objetivo:** Conectar o backend à UI de forma reativa.
- **Tasks:**
    - [ ] Expor métodos do `Manager` no `app.go`.
    - [ ] Implementar sistema de eventos Go -> JS para atualizações de status em tempo real.

### Épico 5: Interface Nuxt (UI/UX)
- **Objetivo:** Criar um painel de controle intuitivo.
- **Tasks:**
    - [ ] Dashboard de conexões ativas.
    - [ ] Gerenciador de inventário de dispositivos.
    - [ ] Seletores de Input/Output com visual nativo (Nuxt UI).

## 5. Roadmap Técnico (Sprints)

1. **Sprint 1 (Fundação):** Finalização do `core/models.go` e persistência com Viper.
2. **Sprint 2 (Motores):** Implementação do PipeWireEngine e DarwinEngine (Start/Stop limpo).
3. **Sprint 3 (CLI/Daemon):** Comandos Cobra funcionais e teste de tunneling via terminal.
4. **Sprint 4 (GUI):** Desenvolvimento da interface Nuxt e integração de eventos Wails.
5. **Sprint 5 (Polimento):** Tray icon, notificações de sistema e builds de produção.

## 6. Considerações Multiplataforma
- **Linux:** Exige `webkit2gtk-4.1` e `pipewire`. Build via `wails dev -tags webkit2_41`.
- **macOS:** Exige `roc-toolkit`. Build nativo via `wails dev`.
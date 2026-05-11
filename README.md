# 🧜‍♀️ Siren

**Siren** é uma solução de infraestrutura de áudio "any-to-any" projetada para rotear fluxos sonoros entre diferentes sistemas operacionais (macOS e Linux) com baixa latência, utilizando o protocolo **ROC**.

O projeto opera como uma aplicação híbrida: um binário único que atua como **CLI**, **Daemon** (background) e **GUI** (Wails + Nuxt 3).

## 🚀 Funcionalidades Atuais

- **Core Multiplataforma:** Motor de áudio específico para Linux (PipeWire) e macOS (CoreAudio/ROC).
- **CLI Robusta:** Gerenciamento de inventário e túneis diretamente pelo terminal.
- **Persistência:** Armazenamento automático de dispositivos e configurações em `~/.config/siren/config.json`.
- **Zero Zumbis:** Gerenciamento limpo de processos e módulos, garantindo liberação do hardware ao encerrar.

## 🛠️ Pré-requisitos

### Linux
- **PipeWire** (com módulo ROC instalado)
- `pw-cli` disponível no PATH

### macOS
- **ROC Toolkit** instalado (ex: `brew install roc-toolkit`)
- Binários `roc-send` e `roc-recv` disponíveis no PATH

## 💻 Como Usar (CLI)

### Gerenciar Dispositivos
```bash
# Adicionar um novo dispositivo
go run main.go app.go device add "Meu-Mac" "192.168.1.50" darwin

# Listar dispositivos cadastrados (para obter o ID)
go run main.go app.go device list

# Remover um dispositivo
go run main.go app.go device remove <id>
```

### Iniciar Túnel de Áudio
```bash
# Iniciar o roteamento para um dispositivo
go run main.go app.go tunnel start <id>

# Para parar, basta pressionar Ctrl+C no terminal
```

## 🏗️ Desenvolvimento

Este projeto utiliza **Wails v2** com **Nuxt 3**.

### Ambiente Nix (Recomendado)
Se você usa Nix, basta rodar:
```bash
nix develop
```

### Rodar Interface Gráfica (Modo Dev)
```bash
# Linux
wails dev -tags webkit2_41

# macOS
wails dev
```

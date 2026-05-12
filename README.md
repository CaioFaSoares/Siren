# 🧜‍♀️ Siren

**Siren** é uma ferramenta premium de roteamento de áudio "any-to-any", projetada para conectar fluxos sonoros entre **macOS** e **Linux** com ultra-baixa latência e correção de erros, utilizando o protocolo **ROC**.

Seja para usar o microfone de um Mac no seu setup Linux ou ouvir o áudio do seu servidor de mídia nos alto-falantes de outro computador, o Siren orquestra tudo de forma transparente.

## ✨ Destaques

- **Modo Duplex:** Envie e receba áudio simultaneamente (Bidirecional).
- **Engine Híbrida:** 
  - **Linux:** Integração profunda com **PipeWire** via `pw-cli`, garantindo que os módulos nunca sumam do sistema.
  - **macOS:** Orquestração de processos `roc-send` e `roc-recv` para máxima compatibilidade.
- **Abstração de Nodes:** Escolha exatamente qual microfone (Source) enviar e em qual alto-falante (Sink) receber.
- **CLI & GUI:** Funciona como um utilitário de terminal poderoso ou uma aplicação visual moderna (Wails + Nuxt 3).
- **Persistência Inteligente:** Suas configurações de dispositivos e túneis são salvas automaticamente via Viper.

## 🛠️ Pré-requisitos

### 🐧 Linux
- **PipeWire** (com o plugin ROC instalado: `pipewire-module-roc`)
- `pw-cli` disponível no seu PATH.

### 🍎 macOS
- **ROC Toolkit** instalado (via Homebrew: `brew install roc-toolkit`).
- Binários `roc-send` e `roc-recv` disponíveis no seu PATH.

## 💻 Uso via CLI

O Siren agora possui binários separados para CLI e GUI. 

### Gerenciar Inventário
```bash
# Se estiver usando o binário compilado:
./build/bin/siren-cli device list

# Adicionar um novo computador
./build/bin/siren-cli device add "MacBook-Pro" "192.168.1.50" darwin
```

### Roteamento de Áudio (Túnel)
```bash
# Iniciar túnel bidirecional (Duplex) para um dispositivo
./build/bin/siren-cli tunnel start <device_id> --mode duplex
```

## 🏗️ Desenvolvimento

Este projeto é um monorepo que separa a lógica de interface (GUI) da lógica de terminal (CLI).

### Ambiente Nix
```bash
nix develop
```

### Compilar CLI
```bash
go build -o build/bin/siren-cli ./cmd/siren-cli
```

### Rodar GUI em Modo Dev
```bash
# Compilar e rodar com hot-reload (Linux com suporte a WebKit)
wails dev -tags webkit2_41
```

### Compilar GUI (Produção)
```bash
wails build -tags webkit2_41
```

---
*Desenvolvido com ❤️ para audiófilos e adeptos de multisetup.*

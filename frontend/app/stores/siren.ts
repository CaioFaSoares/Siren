// Mock de segurança para SSR
if (typeof window === 'undefined') {
  (globalThis as any).window = {}
}

import { defineStore } from 'pinia'
import { ref, onMounted } from 'vue'
import { Events } from '@wailsio/runtime'
import * as App from '../../bindings/siren/app'
import { Device, AudioNode } from '../../bindings/siren/core/models'

export const useSirenStore = defineStore('siren', () => {
  // --- State ---
  const devices = ref<Device[]>([])
  const localInputs = ref<AudioNode[]>([])
  const localOutputs = ref<AudioNode[]>([])
  const isTunnelActive = ref(false)
  const activeTunnelDeviceID = ref('')
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // --- Actions ---

  /**
   * Busca a lista de dispositivos remotos cadastrados no backend
   */
  async function fetchDevices() {
    try {
      isLoading.value = true
      const result = await App.GetDevices()
      devices.value = result
    } catch (err) {
      error.value = 'Falha ao buscar dispositivos'
      console.error(err)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Busca os nós de áudio físicos do sistema local (PipeWire)
   */
  async function fetchHardwareNodes() {
    try {
      const [inputs, outputs] = await Promise.all([
        App.GetLocalInputs(),
        App.GetLocalOutputs()
      ])
      localInputs.value = inputs
      localOutputs.value = outputs
    } catch (err) {
      console.error('Falha ao buscar hardware nodes:', err)
    }
  }

  /**
   * Inicia um túnel de áudio para um dispositivo específico
   */
  async function connectTunnel(deviceID: string, mode: string, localNodeID: string, remoteNodeID: string) {
    try {
      isLoading.value = true
      await App.StartTunnel(deviceID, mode, localNodeID, remoteNodeID)
      activeTunnelDeviceID.value = deviceID
      // O state isTunnelActive será atualizado pelo evento tunnel-status vindo do Go
    } catch (err) {
      error.value = 'Erro ao iniciar túnel'
      console.error(err)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Cadastra um novo dispositivo no backend
   */
  async function addDevice(name: string, ip: string, platform: string) {
    try {
      isLoading.value = true
      await App.AddDevice(name, ip, platform)
      await fetchDevices() // Atualiza a lista após adicionar
    } catch (err) {
      error.value = 'Erro ao adicionar dispositivo'
      console.error(err)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Remove um dispositivo pelo ID
   */
  async function removeDevice(id: string) {
    try {
      isLoading.value = true
      await App.RemoveDevice(id)
      await fetchDevices()
    } catch (err) {
      error.value = 'Erro ao remover dispositivo'
      console.error(err)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Encerra o túnel de áudio ativo
   */
  async function disconnectTunnel() {
    try {
      isLoading.value = true
      await App.StopTunnel()
      activeTunnelDeviceID.value = ''
    } catch (err) {
      console.error('Erro ao parar túnel:', err)
    } finally {
      isLoading.value = false
    }
  }

  // --- Lifecycle & Events ---

  // Inicialização e Listeners
  onMounted(() => {
    // Sincronização inicial
    fetchDevices()
    fetchHardwareNodes()

    // Escutar eventos de status do daemon (reatividade real)
    Events.On('tunnel-status', (data: any) => {
      console.log('Evento tunnel-status recebido:', data)
      isTunnelActive.value = !!data
    })
  })

  return {
    // State
    devices,
    localInputs,
    localOutputs,
    isTunnelActive,
    activeTunnelDeviceID,
    isLoading,
    error,

    // Actions
    fetchDevices,
    fetchHardwareNodes,
    addDevice,
    removeDevice,
    connectTunnel,
    disconnectTunnel
  }
})

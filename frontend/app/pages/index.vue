<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { Events } from '@wailsio/runtime'
import { useSirenStore } from '~/stores/siren'

const siren = useSirenStore()
const toast = useToast()

// --- Handlers de Eventos ---

const onTunnelStatus = (data: any) => {
  console.log('Event [tunnel-status]:', data)
  
  // O valor vem do Go como bool
  const isActive = !!data
  
  if (isActive) {
    toast.add({
      title: 'Túnel Ativo',
      description: 'A transmissão de áudio foi estabelecida com sucesso.',
      icon: 'i-heroicons-check-circle',
      color: 'primary'
    })
  } else {
    // Se estava ativo e agora não está, notificamos a desconexão
    if (siren.isTunnelActive) {
      toast.add({
        title: 'Túnel Encerrado',
        description: 'O áudio foi desconectado.',
        icon: 'i-heroicons-information-circle',
        color: 'neutral'
      })
    }
    siren.activeTunnelDeviceID = ''
  }
  
  siren.isTunnelActive = isActive
}

// --- Lifecycle ---

onMounted(() => {
  // Sincronizar dados iniciais
  siren.fetchDevices()
  siren.fetchHardwareNodes()

  // Registrar listeners do Wails v3
  Events.On('tunnel-status', onTunnelStatus)
})

onUnmounted(() => {
  // Limpeza de listeners para evitar vazamento de memória
  Events.Off('tunnel-status', onTunnelStatus)
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 relative z-10">
    <!-- Header -->
    <header class="flex items-center justify-between mb-12">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-primary-500 rounded-2xl shadow-lg shadow-primary-500/20">
          <UIcon name="i-heroicons-musical-note-20-solid" class="w-8 h-8 text-white" />
        </div>
        <div>
          <h1 class="text-3xl font-black tracking-tight text-white">Siren</h1>
          <p class="text-sm text-gray-500 font-medium">Remote Audio Orchestrator</p>
        </div>
      </div>

      <div class="flex items-center gap-4">
        <UBadge v-if="siren.isTunnelActive" color="primary" variant="subtle" class="hidden sm:flex">
          Streaming Ativo
        </UBadge>
        <AddDeviceModal />
      </div>
    </header>

    <!-- Main Grid -->
    <main>
      <div v-if="siren.isLoading && siren.devices.length === 0" class="flex flex-col items-center justify-center py-24 space-y-4">
        <ULoadingIcon class="w-10 h-10 text-primary-500" />
        <p class="text-gray-500 animate-pulse">Sincronizando dispositivos...</p>
      </div>

      <div v-else-if="siren.devices.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
        <DeviceCard 
          v-for="device in siren.devices" 
          :key="device.id" 
          :device="device" 
        />
      </div>

      <!-- Empty State -->
      <div v-else class="flex flex-col items-center justify-center py-24 text-center">
        <div class="w-24 h-24 bg-gray-900 rounded-full flex items-center justify-center mb-6 border border-gray-800">
          <UIcon name="i-heroicons-computer-desktop" class="w-10 h-10 text-gray-700" />
        </div>
        <h2 class="text-xl font-bold text-white mb-2">Nenhum dispositivo encontrado</h2>
        <p class="text-gray-500 max-w-sm mb-8">
          Adicione seu primeiro computador remoto para começar a transmitir áudio de baixa latência.
        </p>
        <AddDeviceModal />
      </div>
    </main>

    <!-- Footer / Status Bar -->
    <footer class="mt-20 pt-8 border-t border-white/5 flex items-center justify-between text-xs text-gray-600 uppercase tracking-widest font-bold">
      <div class="flex items-center gap-4">
        <span>Status do Sistema: <span class="text-green-500">Online</span></span>
        <span class="w-1 h-1 bg-gray-800 rounded-full" />
        <span>Wails v3 Alpha</span>
      </div>
      <div>
        Siren v0.1.0
      </div>
    </footer>
  </div>
</template>

<style scoped>
/* Efeito de fade para o grid */
.grid {
  animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>

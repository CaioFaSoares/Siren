<script setup lang="ts">
import { ref, computed } from 'vue'
import { useSirenStore } from '~/stores/siren'
import type { Device } from '~/bindings/siren/core/models'

const props = defineProps<{
  device: Device
}>()

const siren = useSirenStore()

// --- Local State for Toggles ---
const sendMic = ref(true)
const recvMic = ref(false)
const sendAudio = ref(false)
const recvAudio = ref(true)

// IDs de Hardware Selecionados
const selectedSource = ref('')
const selectedSink = ref('')

// --- Computed ---
const isConnected = computed(() => {
  return siren.isTunnelActive && siren.activeTunnelDeviceID === props.device.id
})

const platformIcon = computed(() => {
  return props.device.platform === 'darwin' 
    ? 'i-simple-icons-apple' 
    : 'i-simple-icons-linux'
})

// Determina o modo baseado nos toggles
const calculatedMode = computed(() => {
  const isTX = sendMic.value || sendAudio.value
  const isRX = recvMic.value || recvAudio.value
  
  if (isTX && isRX) return 'duplex'
  if (isTX) return 'sender'
  if (isRX) return 'receiver'
  return 'none'
})

// --- Actions ---
async function handleConnection() {
  if (isConnected.value) {
    await siren.disconnectTunnel()
  } else {
    if (calculatedMode.value === 'none') {
      alert('Selecione pelo menos uma direção de áudio')
      return
    }
    
    await siren.connectTunnel(
      props.device.id,
      calculatedMode.value,
      selectedSource.value,
      selectedSink.value
    )
  }
}
</script>

<template>
  <UCard
    class="relative overflow-hidden transition-all duration-300 border-gray-800"
    :class="[
      isConnected 
        ? 'ring-2 ring-primary-500 bg-primary-950/20 shadow-lg shadow-primary-900/20' 
        : 'bg-gray-900/40 backdrop-blur-md hover:bg-gray-900/60'
    ]"
  >
    <!-- Header -->
    <template #header>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div 
            class="p-2 rounded-lg bg-gray-800 text-gray-400"
            :class="{ 'bg-primary-500/20 text-primary-400': isConnected }"
          >
            <UIcon :name="platformIcon" class="w-6 h-6" />
          </div>
          <div>
            <h3 class="font-bold text-white leading-tight">{{ device.name }}</h3>
            <p class="text-xs text-gray-500 font-mono">{{ device.ip }}</p>
          </div>
        </div>
        
        <UBadge 
          v-if="isConnected" 
          color="primary" 
          variant="subtle" 
          size="xs"
          class="animate-pulse"
        >
          ATIVO
        </UBadge>
        <UButton
          v-else
          icon="i-heroicons-trash"
          color="gray"
          variant="ghost"
          size="xs"
          @click="siren.removeDevice(device.id)"
        />
      </div>
    </template>

    <!-- Body -->
    <div class="space-y-6">
      <!-- Toggles Grid -->
      <div class="grid grid-cols-2 gap-4 p-3 rounded-xl bg-black/20 border border-white/5">
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-xs text-gray-400">Enviar Mic</span>
            <UToggle v-model="sendMic" size="sm" />
          </div>
          <div class="flex items-center justify-between">
            <span class="text-xs text-gray-400">Enviar Áudio</span>
            <UToggle v-model="sendAudio" size="sm" />
          </div>
        </div>
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-xs text-gray-400">Receber Mic</span>
            <UToggle v-model="recvMic" size="sm" />
          </div>
          <div class="flex items-center justify-between">
            <span class="text-xs text-gray-400">Receber Áudio</span>
            <UToggle v-model="recvAudio" size="sm" />
          </div>
        </div>
      </div>

      <!-- Hardware Selection -->
      <div class="grid grid-cols-1 gap-4">
        <HardwareSelector type="source" v-model="selectedSource" />
        <HardwareSelector type="sink" v-model="selectedSink" />
      </div>
    </div>

    <!-- Footer -->
    <template #footer>
      <UButton
        block
        size="lg"
        :color="isConnected ? 'red' : 'primary'"
        :variant="isConnected ? 'soft' : 'solid'"
        :icon="isConnected ? 'i-heroicons-stop-circle' : 'i-heroicons-play-circle'"
        :loading="siren.isLoading"
        @click="handleConnection"
      >
        {{ isConnected ? 'Desconectar' : 'Conectar Agora' }}
      </UButton>
    </template>
  </UCard>
</template>

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
    class="relative overflow-hidden transition-all duration-300 border-slate-800/50 hover:border-primary-500/30 hover:shadow-primary-500/10 group"
    :class="[
      isConnected 
        ? 'ring-1 ring-primary-500/50 bg-primary-950/20 shadow-lg shadow-primary-900/10' 
        : 'bg-slate-900/40 backdrop-blur-md'
    ]"
  >
    <!-- Header -->
    <template #header>
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div 
            class="p-2.5 rounded-xl bg-slate-800/50 text-slate-400 transition-colors group-hover:bg-slate-800"
            :class="{ 'bg-primary-500/20 text-primary-400': isConnected }"
          >
            <UIcon :name="platformIcon" class="w-5 h-5" />
          </div>
          <div>
            <h3 class="font-bold text-slate-100 leading-tight">{{ device.name }}</h3>
            <p class="text-[10px] text-slate-500 font-mono tracking-tighter uppercase">{{ device.ip }}</p>
          </div>
        </div>
        
        <div class="flex items-center gap-2">
          <UBadge 
            v-if="isConnected" 
            color="primary" 
            variant="subtle" 
            size="xs"
            class="animate-pulse font-bold tracking-widest text-[9px]"
          >
            LIVE
          </UBadge>
          <UButton
            v-else
            icon="i-heroicons-trash"
            color="neutral"
            variant="ghost"
            size="xs"
            class="opacity-0 group-hover:opacity-100 transition-opacity"
            @click="siren.removeDevice(device.id)"
          />
        </div>
      </div>
    </template>

    <!-- Body -->
    <div class="space-y-6">
      <!-- Toggles Grid -->
      <div class="grid grid-cols-2 gap-4 p-4 rounded-xl bg-slate-950/40 border border-slate-800/50">
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <span class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Enviar Mic</span>
            <UToggle v-model="sendMic" size="sm" color="primary" />
          </div>
          <div class="flex items-center justify-between">
            <span class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Enviar Som</span>
            <UToggle v-model="sendAudio" size="sm" color="primary" />
          </div>
        </div>
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <span class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Recv Mic</span>
            <UToggle v-model="recvMic" size="sm" color="primary" />
          </div>
          <div class="flex items-center justify-between">
            <span class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Recv Som</span>
            <UToggle v-model="recvAudio" size="sm" color="primary" />
          </div>
        </div>
      </div>

      <!-- Hardware Selection -->
      <div class="space-y-3">
        <HardwareSelector type="source" v-model="selectedSource" />
        <HardwareSelector type="sink" v-model="selectedSink" />
      </div>
    </div>

    <!-- Footer -->
    <template #footer>
      <UButton
        block
        size="lg"
        :color="isConnected ? 'error' : 'primary'"
        :variant="isConnected ? 'solid' : 'outline'"
        :icon="isConnected ? 'i-heroicons-stop-circle' : 'i-heroicons-play-circle'"
        :loading="siren.isLoading"
        :class="{ 'animate-pulse bg-rose-600 border-none': isConnected }"
        @click="handleConnection"
      >
        {{ isConnected ? 'Encerrar Túnel' : 'Iniciar Conexão' }}
      </UButton>
    </template>
  </UCard>
</template>

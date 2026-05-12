<script setup lang="ts">
import { ref } from 'vue'
import { useSirenStore } from '~/stores/siren'

const siren = useSirenStore()
const isOpen = ref(false)

const state = ref({
  name: '',
  ip: '',
  platform: 'linux'
})

const platforms = [
  { label: 'Linux', value: 'linux' },
  { label: 'macOS', value: 'darwin' }
]

async function onSubmit() {
  if (!state.value.name || !state.value.ip) return
  
  await siren.addDevice(state.value.name, state.value.ip, state.value.platform)
  
  // Limpar form e fechar
  state.value = { name: '', ip: '', platform: 'linux' }
  isOpen.value = false
}

defineExpose({
  open: () => (isOpen.value = true)
})
</script>

<template>
  <div>
    <UButton
      label="Adicionar Dispositivo"
      icon="i-heroicons-plus-circle"
      color="gray"
      variant="ghost"
      @click="isOpen = true"
    />

    <UModal v-model="isOpen">
      <UCard :ui="{ ring: '', divide: 'divide-y divide-gray-100 dark:divide-gray-800' }">
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-base font-semibold leading-6 text-gray-900 dark:text-white">
              Novo Computador
            </h3>
            <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark-20-solid" class="-my-1" @click="isOpen = false" />
          </div>
        </template>

        <UForm :state="state" class="space-y-4" @submit="onSubmit">
          <UFormGroup label="Nome do Dispositivo" name="name" required>
            <UInput v-model="state.name" placeholder="Ex: iMac-Pro" />
          </UFormGroup>

          <UFormGroup label="Endereço IP (ZeroTier/Tailscale)" name="ip" required>
            <UInput v-model="state.ip" placeholder="10.147.x.x" />
          </UFormGroup>

          <UFormGroup label="Sistema Operacional" name="platform">
            <USelect v-model="state.platform" :options="platforms" />
          </UFormGroup>

          <div class="flex justify-end gap-2 pt-4">
            <UButton label="Cancelar" color="gray" variant="ghost" @click="isOpen = false" />
            <UButton type="submit" label="Salvar Dispositivo" color="primary" :loading="siren.isLoading" />
          </div>
        </UForm>
      </UCard>
    </UModal>
  </div>
</template>

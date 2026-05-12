<script setup lang="ts">
import { computed } from 'vue'
import { useSirenStore } from '~/stores/siren'
import type { AudioNode } from '~/bindings/siren/core/models'

const props = defineProps<{
  type: 'source' | 'sink'
  modelValue: string
}>()

const emit = defineEmits(['update:modelValue'])

const siren = useSirenStore()

// Computado para selecionar a lista correta baseada no tipo
const options = computed(() => {
  const list = props.type === 'source' ? siren.localInputs : siren.localOutputs
  
  // Adiciona a opção padrão
  const base = [{ id: '', name: 'Padrão do Sistema', is_default: true }]
  return [...base, ...list]
})

// Valor selecionado (ID)
const selected = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

// Label para mostrar no dropdown
const selectedLabel = computed(() => {
  const item = options.value.find(o => o.id === selected.value)
  return item ? item.name : 'Selecionar hardware...'
})
</script>

<template>
  <div class="space-y-1">
    <span class="text-xs font-medium text-gray-500 uppercase tracking-wider">
      {{ type === 'source' ? 'Microfone (Input)' : 'Saída (Output)' }}
    </span>
    
    <USelectMenu
      v-model="selected"
      :options="options"
      value-attribute="id"
      option-attribute="name"
      class="w-full"
    >
      <template #default="{ open }">
        <UButton
          color="gray"
          variant="ghost"
          class="w-full justify-between bg-gray-100/50 dark:bg-gray-800/50"
          :icon="type === 'source' ? 'i-heroicons-microphone' : 'i-heroicons-speaker-wave'"
        >
          <span class="truncate">{{ selectedLabel }}</span>
          <UIcon
            name="i-heroicons-chevron-up-down-20-solid"
            class="w-5 h-5 transition-transform"
            :class="[open && 'transform rotate-180']"
          />
        </UButton>
      </template>

      <template #option="{ option }">
        <div class="flex items-center gap-2">
          <UIcon
            v-if="option.is_default"
            name="i-heroicons-star-20-solid"
            class="text-yellow-500"
          />
          <span>{{ option.name }}</span>
        </div>
      </template>
    </USelectMenu>
  </div>
</template>

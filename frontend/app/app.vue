<script setup lang="ts">
// Siren - Shell Principal
useHead({
  title: 'Siren - Audio Daemon',
  meta: [
    { name: 'viewport', content: 'width=device-width, initial-scale=1' }
  ],
  bodyAttrs: {
    class: 'bg-gray-950 selection:bg-primary-500/30'
  }
})
</script>

<template>
  <UApp>
    <div class="min-h-screen font-sans antialiased">
      <NuxtPage />
    </div>
    
    <UNotifications />
  </UApp>
</template>

<style>
/* Smooth transitions */
.page-enter-active,
.page-leave-active {
  transition: all 0.2s;
}
.page-enter-from,
.page-leave-to {
  opacity: 0;
  filter: blur(1rem);
}
</style>

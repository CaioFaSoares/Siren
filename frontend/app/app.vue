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
    <div class="min-h-screen bg-gray-950 text-white font-sans antialiased">
      <!-- Background Effects -->
      <div class="fixed inset-0 overflow-hidden pointer-events-none">
        <div class="absolute -top-[10%] -left-[10%] w-[40%] h-[40%] bg-primary-500/10 blur-[120px] rounded-full" />
        <div class="absolute top-[20%] -right-[10%] w-[30%] h-[30%] bg-blue-500/5 blur-[100px] rounded-full" />
      </div>

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

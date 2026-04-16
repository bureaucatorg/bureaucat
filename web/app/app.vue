<script setup lang="ts">
const { isAuthenticated } = useAuth();
</script>

<template>
  <div class="min-h-screen bg-background text-foreground">
    <a href="#main-content" class="skip-to-main">Skip to main content</a>
    <NuxtRouteAnnouncer />
    <Sidebar v-if="isAuthenticated" />
    <div :class="isAuthenticated && 'pl-12'">
      <NuxtPage />
    </div>
    <Sonner position="bottom-right" :expand="true" rich-colors />
  </div>
</template>

<template>
  <div class="min-h-screen bg-background text-foreground">
    <a href="#main-content" class="skip-to-main">Skip to main content</a>
    <NuxtRouteAnnouncer />
    <NuxtPage />
    <Sonner position="bottom-right" :expand="true" rich-colors />
  </div>
</template>

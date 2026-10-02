<script setup lang="ts">
withDefaults(defineProps<{ size?: number }>(), { size: 32 });
</script>

<template>
  <svg :width="size" :height="size" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path d="M5.5 14L2.5 3L13 10Z" class="fill-amber-500" />
    <path d="M26.5 14L29.5 3L19 10Z" class="fill-amber-500" />
    <path d="M7 12L4.5 5.5L11.5 9.5Z" class="fill-amber-600" />
    <path d="M25 12L27.5 5.5L20.5 9.5Z" class="fill-amber-600" />
    <circle cx="16" cy="18" r="11" class="fill-amber-500" />
    <circle cx="11.5" cy="16.5" r="3.8" class="stroke-zinc-950" stroke-width="1" fill="none" />
    <circle cx="20.5" cy="16.5" r="3.8" class="stroke-zinc-950" stroke-width="1" fill="none" />
    <path d="M15.3 16C15.5 15.2 16.5 15.2 16.7 16" class="stroke-zinc-950" stroke-width="0.8" fill="none" />
    <ellipse cx="11.5" cy="16.8" rx="1.8" ry="2" class="fill-zinc-950" />
    <ellipse cx="20.5" cy="16.8" rx="1.8" ry="2" class="fill-zinc-950" />
    <circle cx="11.8" cy="16.2" r="0.9" class="fill-amber-500" />
    <circle cx="20.8" cy="16.2" r="0.9" class="fill-amber-500" />
    <circle cx="12.3" cy="15.7" r="0.45" class="fill-zinc-50" />
    <circle cx="21.3" cy="15.7" r="0.45" class="fill-zinc-50" />
    <path d="M14.8 21L16 20L17.2 21Z" class="fill-amber-700" />
    <path d="M14 22.5C14.8 23.2 17.2 23.2 18 22.5" class="stroke-amber-800" stroke-width="0.7" stroke-linecap="round" fill="none" />
    <line x1="2" y1="17.5" x2="7.5" y2="19" class="stroke-amber-600" stroke-width="0.5" stroke-linecap="round" />
    <line x1="1.5" y1="20.5" x2="7.5" y2="20.5" class="stroke-amber-600" stroke-width="0.5" stroke-linecap="round" />
    <line x1="30" y1="17.5" x2="24.5" y2="19" class="stroke-amber-600" stroke-width="0.5" stroke-linecap="round" />
    <line x1="30.5" y1="20.5" x2="24.5" y2="20.5" class="stroke-amber-600" stroke-width="0.5" stroke-linecap="round" />
  </svg>
</template>

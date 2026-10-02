<script setup lang="ts">
import { ref, onMounted } from "vue";
import { Star, Sun, Moon } from "lucide-vue-next";
import CatLogo from "./CatLogo.vue";
import { appLink } from "../lib/config";

const dark = ref(false);
onMounted(() => (dark.value = document.documentElement.classList.contains("dark")));

function toggleTheme() {
  dark.value = !dark.value;
  document.documentElement.classList.toggle("dark", dark.value);
  try {
    localStorage.setItem("bureaucat-color-mode", dark.value ? "dark" : "light");
  } catch {}
}
</script>

<template>
  <header class="sticky top-0 z-40 border-b border-border/60 bg-paper/80 backdrop-blur-xl">
    <div class="mx-auto flex h-12 max-w-6xl items-center justify-between gap-3 px-4 md:px-6">
      <a href="/" aria-label="BureauCat home" class="flex items-center gap-2 font-semibold tracking-tight">
        <CatLogo :size="28" />
        <span class="text-lg">Bureau<span class="text-amber-500">Cat</span></span>
      </a>
      <nav class="flex items-center gap-1 sm:gap-2">
        <a
          href="https://github.com/bureaucatorg/bureaucat"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex h-8 items-center gap-1.5 rounded-md border border-border/70 px-2 text-xs text-muted-foreground transition-colors hover:text-foreground sm:px-2.5"
          aria-label="Star BureauCat on GitHub"
        >
          <Star class="size-3.5" />
          <span class="hidden sm:inline">Star on GitHub</span>
        </a>
        <button
          type="button"
          class="inline-flex size-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          :aria-label="dark ? 'Switch to light theme' : 'Switch to dark theme'"
          @click="toggleTheme"
        >
          <Moon v-if="dark" class="size-4" />
          <Sun v-else class="size-4" />
        </button>
        <a :href="appLink('/signin')" class="inline-flex h-8 items-center rounded-md px-3 text-sm font-medium transition-colors hover:bg-muted">
          Sign in
        </a>
      </nav>
    </div>
  </header>
</template>

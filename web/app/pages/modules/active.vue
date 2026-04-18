<script setup lang="ts">
import { ChevronLeft, Loader2, Layers } from "lucide-vue-next";

definePageMeta({ middleware: ["auth"] });

useSeoMeta({ title: "Active Modules" });

const { activeModules, loading, listActiveModules } = useModules();

onMounted(() => {
  listActiveModules();
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />
    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-8">
        <nav class="mb-4 flex items-center gap-2 text-sm text-muted-foreground">
          <ChevronLeft class="size-4" />
          <span class="font-semibold text-amber-600 dark:text-amber-500">
            Active Modules
          </span>
        </nav>

        <div class="mb-8 flex flex-col gap-2">
          <h1 class="flex items-center gap-3 text-3xl font-bold tracking-tight">
            <Layers class="size-8" />
            Active Modules
          </h1>
          <p class="text-muted-foreground">
            Every module currently in progress across projects you belong to.
          </p>
        </div>

        <div v-if="loading" class="flex items-center justify-center py-12">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <div
          v-else-if="activeModules.length === 0"
          class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
        >
          <div class="flex size-16 items-center justify-center rounded-full bg-muted">
            <Layers class="size-8 text-muted-foreground" />
          </div>
          <h3 class="mt-4 text-lg font-semibold">No active modules</h3>
          <p class="mt-2 max-w-sm text-center text-sm text-muted-foreground">
            Open a project and flip a module's status to
            <span class="font-medium">in progress</span> to see it surface here.
          </p>
        </div>

        <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <ModuleCard
            v-for="m in activeModules"
            :key="m.id"
            :module="m"
            show-project
          />
        </div>
      </div>
    </main>
  </div>
</template>

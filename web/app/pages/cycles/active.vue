<script setup lang="ts">
import { ChevronLeft, Loader2, Repeat } from "lucide-vue-next";

definePageMeta({ middleware: ["auth"] });

useSeoMeta({ title: "Active Cycles" });

const { activeCycles, loading, listActiveCycles } = useCycles();

onMounted(() => {
  listActiveCycles();
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
            Active Cycles
          </span>
        </nav>

        <div class="mb-8 flex flex-col gap-2">
          <h1 class="flex items-center gap-3 text-3xl font-bold tracking-tight">
            <Repeat class="size-8" />
            Active Cycles
          </h1>
          <p class="text-muted-foreground">
            Every cycle currently running across projects you belong to.
          </p>
        </div>

        <div v-if="loading" class="flex items-center justify-center py-12">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <div
          v-else-if="activeCycles.length === 0"
          class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
        >
          <div class="flex size-16 items-center justify-center rounded-full bg-muted">
            <Repeat class="size-8 text-muted-foreground" />
          </div>
          <h3 class="mt-4 text-lg font-semibold">No active cycles</h3>
          <p class="mt-2 max-w-sm text-center text-sm text-muted-foreground">
            Open a project and create a cycle that covers today's date to see it
            surface here.
          </p>
        </div>

        <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <CycleCard
            v-for="c in activeCycles"
            :key="c.id"
            :cycle="c"
            show-project
          />
        </div>
      </div>
    </main>
  </div>
</template>

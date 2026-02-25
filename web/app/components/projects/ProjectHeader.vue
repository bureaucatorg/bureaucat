<script setup lang="ts">
import { FolderKanban, Users, Settings, ChevronLeft } from "lucide-vue-next";
import type { Project } from "~/types";

const props = defineProps<{
  project: Project;
  memberCount?: number;
}>();

const isAdmin = computed(() => props.project.role === "admin");
</script>

<template>
  <div class="space-y-4">
    <nav class="flex items-center gap-2 text-sm text-muted-foreground">
      <NuxtLink to="/projects" class="flex items-center gap-1 hover:text-foreground">
        <ChevronLeft class="size-4" />
        Projects
      </NuxtLink>
      <span>/</span>
      <span class="font-medium text-foreground">
        {{ project.name }}
      </span>
    </nav>

    <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
      <div class="flex items-start gap-4">
        <div
          class="flex size-14 items-center justify-center rounded-xl bg-muted"
        >
          <FolderKanban class="size-7 text-muted-foreground" />
        </div>
        <div>
          <h1 class="text-2xl font-bold tracking-tight">{{ project.name }}</h1>
          <p
            v-if="project.description"
            class="mt-1 text-sm text-muted-foreground"
          >
            {{ project.description }}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <div
          v-if="memberCount"
          class="flex items-center gap-2 rounded-lg border px-3 py-1.5 text-sm"
        >
          <Users class="size-4 text-muted-foreground" />
          <span>{{ memberCount }} members</span>
        </div>
        <Badge variant="secondary" class="capitalize">{{ project.role }}</Badge>
      </div>
    </div>
  </div>
</template>

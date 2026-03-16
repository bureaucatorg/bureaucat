<script setup lang="ts">
import { FolderKanban, Users, Settings, ChevronLeft, Link, Check } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { Project } from "~/types";

const props = defineProps<{
  project: Project;
  memberCount?: number;
}>();

const isAdmin = computed(() => props.project.role === "admin");

const copied = ref(false);
function copyLink() {
  navigator.clipboard.writeText(window.location.href);
  copied.value = true;
  toast.success("Link copied");
  setTimeout(() => { copied.value = false; }, 2000);
}
</script>

<template>
  <div class="space-y-4">
    <nav class="flex items-center gap-2 text-sm text-muted-foreground">
      <ChevronLeft class="size-4" />
      <NuxtLink to="/projects" class="hover:text-foreground">
        Projects
      </NuxtLink>
      <span>/</span>
      <NuxtLink
        :to="`/projects/${project.project_key}`"
        class="font-semibold text-amber-600 hover:text-amber-700 dark:text-amber-500 dark:hover:text-amber-400"
      >
        {{ project.project_key }}
      </NuxtLink>
      <button
        class="ml-1 rounded-md p-1 text-muted-foreground/50 hover:text-muted-foreground"
        @click="copyLink"
      >
        <Check v-if="copied" class="size-3.5 text-emerald-500" />
        <Link v-else class="size-3.5" />
      </button>
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

<script setup lang="ts" generic="T extends TaskRow">
import { X } from "lucide-vue-next";

// Minimum shape the table needs from each task row.
interface TaskRow {
  id: string;
  task_number: number;
  task_id: string;
  title: string;
  state_name: string;
  state_color: string;
}

const props = defineProps<{
  tasks: T[];
  projectKey: string;
  isAdmin: boolean;
  removeLabel?: string;
}>();

const emit = defineEmits<{ remove: [taskId: string] }>();

const gridStyle = computed(() =>
  props.isAdmin
    ? "grid-template-columns: 140px minmax(0, 1fr) 90px 28px;"
    : "grid-template-columns: 140px minmax(0, 1fr) 90px;"
);
</script>

<template>
  <div class="overflow-hidden rounded-lg border bg-background">
    <div
      class="grid items-center gap-3 border-b bg-muted/40 px-4 py-2 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground"
      :style="gridStyle"
    >
      <span>State</span>
      <span>Title</span>
      <span>ID</span>
      <span v-if="isAdmin"></span>
    </div>

    <div class="max-h-[70vh] overflow-y-auto [scrollbar-gutter:stable]">
      <div
        v-for="task in tasks"
        :key="task.id"
        class="group grid items-center gap-3 border-b border-border/40 px-4 py-2.5 text-sm transition-colors last:border-0 hover:bg-muted/40"
        :style="gridStyle"
      >
        <span
          class="inline-flex w-fit max-w-full items-center truncate rounded px-1.5 py-0.5 font-mono text-[10px] font-medium uppercase tracking-wider"
          :style="{
            backgroundColor: (task.state_color || '#6B7280') + '22',
            color: task.state_color || '#6B7280',
          }"
        >
          {{ task.state_name }}
        </span>

        <NuxtLink
          :to="`/projects/${projectKey}/tasks/${task.task_number}`"
          class="min-w-0 truncate font-medium text-foreground hover:text-amber-600 hover:underline dark:hover:text-amber-500"
        >
          {{ task.title }}
        </NuxtLink>

        <span class="font-mono text-[11px] text-muted-foreground">
          {{ task.task_id }}
        </span>

        <button
          v-if="isAdmin"
          class="rounded p-1 text-muted-foreground opacity-0 transition-opacity hover:bg-muted hover:text-destructive focus-visible:opacity-100 group-hover:opacity-100"
          :aria-label="`${removeLabel || 'Remove'} ${task.title}`"
          @click.prevent="emit('remove', task.id)"
        >
          <X class="size-3.5" />
        </button>
      </div>
    </div>
  </div>
</template>

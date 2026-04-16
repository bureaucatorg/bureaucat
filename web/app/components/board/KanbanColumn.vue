<script setup lang="ts">
import type { Task } from "~/types";

const props = defineProps<{
  columnId: string;
  label: string;
  color: string;
  tasks: Task[];
  projectKey: string;
  isMember: boolean;
  /** Disable drop target (e.g. due_bucket grouping, where moves are ambiguous). */
  dropLocked?: boolean;
}>();

const emit = defineEmits<{
  /**
   * Drop fires when a task is dropped on this column. fromColumnId is the
   * source column (may equal columnId — parent should noop in that case).
   */
  drop: [task: Task, fromColumnId: string, toColumnId: string];
}>();

const isDragOver = ref(false);

function handleDrop(event: DragEvent) {
  event.preventDefault();
  isDragOver.value = false;
  if (!props.isMember || props.dropLocked) return;

  const payload = event.dataTransfer?.getData("application/json");
  if (!payload) return;
  try {
    const data = JSON.parse(payload) as { task: Task; fromColumnId: string };
    if (data?.task) {
      emit("drop", data.task, data.fromColumnId ?? "", props.columnId);
    }
  } catch {
    // ignore invalid payload
  }
}

function handleDragOver(event: DragEvent) {
  event.preventDefault();
  if (props.isMember && !props.dropLocked) {
    event.dataTransfer!.dropEffect = "move";
    isDragOver.value = true;
  } else if (event.dataTransfer) {
    event.dataTransfer.dropEffect = "none";
  }
}

function handleDragLeave() {
  isDragOver.value = false;
}
</script>

<template>
  <div
    class="flex min-w-56 flex-1 flex-col rounded-lg border bg-muted/30 transition-colors"
    :class="{
      'border-primary border-2 bg-primary/5': isDragOver,
      'cursor-not-allowed': dropLocked,
    }"
    @drop="handleDrop"
    @dragover="handleDragOver"
    @dragleave="handleDragLeave"
  >
    <div class="flex items-center gap-2 border-b px-3 py-2.5">
      <div class="size-2.5 rounded-full" :style="{ backgroundColor: color }" />
      <h3 class="truncate text-sm font-medium">{{ label }}</h3>
      <span class="ml-auto text-xs text-muted-foreground">{{ tasks.length }}</span>
    </div>
    <div class="flex flex-1 flex-col gap-2 p-2">
      <div
        v-if="tasks.length === 0"
        class="flex flex-1 items-center justify-center rounded-lg border border-dashed py-8 text-xs text-muted-foreground"
      >
        No tasks
      </div>
      <KanbanCard
        v-for="task in tasks"
        :key="task.id + ':' + columnId"
        :task="task"
        :project-key="projectKey"
        :is-member="isMember && !dropLocked"
        :column-id="columnId"
      />
    </div>
  </div>
</template>

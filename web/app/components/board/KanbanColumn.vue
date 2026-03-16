<script setup lang="ts">
import type { Task, ProjectState } from "~/types";

const props = defineProps<{
  type: string;
  label: string;
  color: string;
  states: ProjectState[];
  tasks: Task[];
  projectKey: string;
  isMember: boolean;
}>();

const emit = defineEmits<{
  move: [task: Task, newStateId: string];
}>();

// Default state for this type (for dropping)
const defaultState = computed(() => props.states[0]);
const isDragOver = ref(false);

function handleDrop(event: DragEvent) {
  event.preventDefault();
  isDragOver.value = false;
  if (!props.isMember || !defaultState.value) return;

  const taskData = event.dataTransfer?.getData("application/json");
  if (taskData) {
    try {
      const task = JSON.parse(taskData) as Task;
      if (task.state_type !== props.type) {
        emit("move", task, defaultState.value.id);
      }
    } catch {
      // Invalid data
    }
  }
}

function handleDragOver(event: DragEvent) {
  event.preventDefault();
  if (props.isMember) {
    event.dataTransfer!.dropEffect = "move";
    isDragOver.value = true;
  }
}

function handleDragLeave() {
  isDragOver.value = false;
}
</script>

<template>
  <div
    class="flex min-w-56 flex-1 flex-col rounded-lg border bg-muted/30 transition-colors"
    :class="{ 'border-primary border-2 bg-primary/5': isDragOver }"
    @drop="handleDrop"
    @dragover="handleDragOver"
    @dragleave="handleDragLeave"
  >
    <!-- Column header -->
    <div class="flex items-center gap-2 border-b px-3 py-2.5">
      <div
        class="size-2.5 rounded-full"
        :style="{ backgroundColor: color }"
      />
      <h3 class="text-sm font-medium">{{ label }}</h3>
      <span class="ml-auto text-xs text-muted-foreground">
        {{ tasks.length }}
      </span>
    </div>

    <!-- Tasks -->
    <div class="flex flex-1 flex-col gap-2 p-2">
      <div
        v-if="tasks.length === 0"
        class="flex flex-1 items-center justify-center rounded-lg border border-dashed py-8 text-xs text-muted-foreground"
      >
        No tasks
      </div>
      <KanbanCard
        v-for="task in tasks"
        :key="task.id"
        :task="task"
        :project-key="projectKey"
        :is-member="isMember"
      />
    </div>
  </div>
</template>

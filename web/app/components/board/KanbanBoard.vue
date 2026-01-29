<script setup lang="ts">
import { Loader2 } from "lucide-vue-next";
import type { Task, ProjectState } from "~/types";

const props = defineProps<{
  tasks: Task[];
  states: ProjectState[];
  projectKey: string;
  isMember: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const { updateTask } = useTasks();
const updating = ref(false);

// Group states by state_type for columns
const columns = computed(() => {
  const types = ["backlog", "unstarted", "started", "completed", "cancelled"];
  return types
    .map((type) => {
      const typeStates = props.states.filter((s) => s.state_type === type);
      const typeTasks = props.tasks.filter((t) => t.state_type === type);
      return {
        type,
        label: getTypeLabel(type),
        states: typeStates,
        tasks: typeTasks,
        color: getTypeColor(type),
      };
    })
    .filter((col) => col.states.length > 0);
});

function getTypeLabel(type: string): string {
  const labels: Record<string, string> = {
    backlog: "Backlog",
    unstarted: "To Do",
    started: "In Progress",
    completed: "Done",
    cancelled: "Cancelled",
  };
  return labels[type] || type;
}

function getTypeColor(type: string): string {
  const colors: Record<string, string> = {
    backlog: "#6B7280",
    unstarted: "#3B82F6",
    started: "#10B981",
    completed: "#22C55E",
    cancelled: "#9CA3AF",
  };
  return colors[type] || "#6B7280";
}

async function handleMoveTask(task: Task, newStateId: string) {
  if (!props.isMember || updating.value) return;

  updating.value = true;
  const result = await updateTask(props.projectKey, task.task_number, {
    state_id: newStateId,
  });
  updating.value = false;

  if (result.success) {
    emit("refresh");
  }
}
</script>

<template>
  <div class="relative">
    <!-- Updating overlay -->
    <div
      v-if="updating"
      class="absolute inset-0 z-10 flex items-center justify-center bg-background/50"
    >
      <Loader2 class="size-6 animate-spin" />
    </div>

    <!-- Board -->
    <div class="flex gap-4 overflow-x-auto pb-4">
      <KanbanColumn
        v-for="column in columns"
        :key="column.type"
        :type="column.type"
        :label="column.label"
        :color="column.color"
        :states="column.states"
        :tasks="column.tasks"
        :project-key="projectKey"
        :is-member="isMember"
        @move="handleMoveTask"
      />
    </div>
  </div>
</template>

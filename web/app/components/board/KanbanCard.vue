<script setup lang="ts">
import type { Task } from "~/types";
import { PRIORITY_LABELS } from "~/types";

const props = defineProps<{
  task: Task;
  projectKey: string;
  isMember: boolean;
}>();

const router = useRouter();
const priorityInfo = computed(() => PRIORITY_LABELS[props.task.priority] || PRIORITY_LABELS[0]);
const isDragging = ref(false);

function handleDragStart(event: DragEvent) {
  if (!props.isMember) {
    event.preventDefault();
    return;
  }
  isDragging.value = true;
  event.dataTransfer!.effectAllowed = "move";
  event.dataTransfer!.setData("application/json", JSON.stringify(props.task));
}

function handleDragEnd() {
  isDragging.value = false;
}

function handleClick() {
  if (!isDragging.value) {
    router.push(`/projects/${props.projectKey}/tasks/${props.task.task_number}`);
  }
}
</script>

<template>
  <div
    :draggable="isMember"
    class="group cursor-pointer rounded-lg border bg-background p-3 shadow-sm transition-all hover:border-amber-500/30 hover:shadow-md"
    :class="{ 'cursor-grab active:cursor-grabbing': isMember }"
    @dragstart="handleDragStart"
    @dragend="handleDragEnd"
    @click="handleClick"
  >
      <!-- Task ID and priority -->
      <div class="mb-2 flex items-center justify-between">
        <span class="font-mono text-xs text-muted-foreground">
          {{ task.task_id }}
        </span>
        <div
          v-if="task.priority > 0"
          class="size-2 rounded-full"
          :style="{ backgroundColor: priorityInfo.color }"
          :title="priorityInfo.label"
        />
      </div>

      <!-- Title -->
      <p class="line-clamp-2 text-sm font-medium">{{ task.title }}</p>

      <!-- Labels -->
      <div
        v-if="task.labels && task.labels.length > 0"
        class="mt-2 flex flex-wrap gap-1"
      >
        <span
          v-for="label in task.labels.slice(0, 3)"
          :key="label.id"
          class="rounded px-1.5 py-0.5 text-xs"
          :style="{
            backgroundColor: label.color + '20',
            color: label.color,
          }"
        >
          {{ label.name }}
        </span>
        <span
          v-if="task.labels.length > 3"
          class="text-xs text-muted-foreground"
        >
          +{{ task.labels.length - 3 }}
        </span>
      </div>

      <!-- Assignees -->
      <div
        v-if="task.assignees && task.assignees.length > 0"
        class="mt-2 flex items-center justify-end"
      >
        <div class="flex -space-x-1.5">
          <Avatar
            v-for="assignee in task.assignees.slice(0, 3)"
            :key="assignee.id"
            class="size-5 border border-background"
          >
            <AvatarFallback class="text-[10px]">
              {{ assignee.first_name[0] }}{{ assignee.last_name[0] }}
            </AvatarFallback>
          </Avatar>
          <div
            v-if="task.assignees.length > 3"
            class="flex size-5 items-center justify-center rounded-full border border-background bg-muted text-[10px]"
          >
            +{{ task.assignees.length - 3 }}
          </div>
        </div>
      </div>
    </div>
</template>

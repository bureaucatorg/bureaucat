<script setup lang="ts">
import { Circle, CircleDot, CheckCircle2, XCircle, Clock } from "lucide-vue-next";
import type { Task } from "~/types";
import { PRIORITY_LABELS } from "~/types";

const props = defineProps<{
  task: Task;
  projectKey: string;
}>();

const stateIcon = computed(() => {
  switch (props.task.state_type) {
    case "backlog":
      return Clock;
    case "unstarted":
      return Circle;
    case "started":
      return CircleDot;
    case "completed":
      return CheckCircle2;
    case "cancelled":
      return XCircle;
    default:
      return Circle;
  }
});

const priorityInfo = computed(() => PRIORITY_LABELS[props.task.priority] || PRIORITY_LABELS[0]);
</script>

<template>
  <NuxtLink :to="`/projects/${projectKey}/tasks/${task.task_number}`">
    <div
      class="group flex items-center gap-3 rounded-lg border border-border/50 bg-background/50 p-3 transition-all hover:border-amber-500/30 hover:bg-muted/50"
    >
      <component
        :is="stateIcon"
        class="size-4 shrink-0"
        :style="{ color: task.state_color }"
      />
      <div class="min-w-0 flex-1">
        <div class="flex items-baseline gap-6">
          <span class="shrink-0 font-mono text-sm text-muted-foreground">{{ task.task_id }}</span>
          <span class="truncate text-sm font-medium">{{ task.title }}</span>
        </div>
        <div class="mt-1 flex items-center gap-2">
          <span class="text-xs text-muted-foreground">{{ task.state_name }}</span>
          <div
            v-if="task.labels && task.labels.length > 0"
            class="flex items-center gap-1"
          >
            <span
              v-for="label in task.labels.slice(0, 2)"
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
              v-if="task.labels.length > 2"
              class="text-xs text-muted-foreground"
            >
              +{{ task.labels.length - 2 }}
            </span>
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <div
          v-if="task.assignees && task.assignees.length > 0"
          class="flex -space-x-1.5"
        >
          <Avatar
            v-for="assignee in task.assignees.slice(0, 3)"
            :key="assignee.id"
            class="size-6 border-2 border-background"
          >
            <AvatarFallback class="text-xs">
              {{ assignee.first_name[0] }}{{ assignee.last_name[0] }}
            </AvatarFallback>
          </Avatar>
          <div
            v-if="task.assignees.length > 3"
            class="flex size-6 items-center justify-center rounded-full border-2 border-background bg-muted text-xs"
          >
            +{{ task.assignees.length - 3 }}
          </div>
        </div>
        <div
          v-if="task.priority > 0"
          class="size-2 rounded-full"
          :style="{ backgroundColor: priorityInfo.color }"
          :title="priorityInfo.label"
        />
      </div>
    </div>
  </NuxtLink>
</template>

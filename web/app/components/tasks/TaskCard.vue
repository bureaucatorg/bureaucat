<script setup lang="ts">
import { Circle, CircleDot, CheckCircle2, XCircle, Clock, MessageSquare } from "lucide-vue-next";
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

// Collect all people involved: creator + assignees (deduplicated)
const involvedPeople = computed(() => {
  const people: { id: string; firstName: string; lastName: string }[] = [];
  const seen = new Set<string>();

  if (props.task.created_by && !seen.has(props.task.created_by)) {
    seen.add(props.task.created_by);
    people.push({
      id: props.task.created_by,
      firstName: props.task.creator_first_name || "",
      lastName: props.task.creator_last_name || "",
    });
  }

  if (props.task.assignees) {
    for (const a of props.task.assignees) {
      if (!seen.has(a.user_id)) {
        seen.add(a.user_id);
        people.push({
          id: a.user_id,
          firstName: a.first_name,
          lastName: a.last_name,
        });
      }
    }
  }

  return people;
});
</script>

<template>
  <NuxtLink :to="`/projects/${projectKey}/tasks/${task.task_number}`">
    <div
      class="task-row group grid items-center rounded-lg border border-border/50 bg-background/50 px-3 py-2.5 transition-all hover:border-amber-500/30 hover:bg-muted/50"
    >
      <!-- Col 1: Task ID -->
      <span class="font-mono text-sm text-muted-foreground truncate">{{ task.task_id }}</span>

      <!-- Col 2: Title -->
      <span class="truncate text-sm font-medium min-w-0">{{ task.title }}</span>

      <!-- Col 3: State badge -->
      <div class="flex items-center gap-1 rounded-md border bg-muted/50 px-1.5 py-0.5 w-fit justify-self-end">
        <component
          :is="stateIcon"
          class="size-3.5 shrink-0 stroke-[2.5]"
          :style="{ color: task.state_color }"
        />
        <span class="text-xs text-muted-foreground whitespace-nowrap">{{ task.state_name }}</span>
      </div>

      <!-- Col 4: Priority badge -->
      <div class="flex items-center gap-1 rounded-md border bg-muted/50 px-1.5 py-0.5 w-fit justify-self-end">
        <span
          class="size-2.5 shrink-0 rounded-full ring-1.5 ring-offset-1 ring-offset-background"
          :style="{ backgroundColor: priorityInfo.color, '--tw-ring-color': priorityInfo.color }"
        />
        <span class="text-xs text-muted-foreground whitespace-nowrap">{{ priorityInfo.label }}</span>
      </div>

      <!-- Col 5: Stacked avatars -->
      <div class="flex items-center justify-end">
        <div
          v-if="involvedPeople.length > 0"
          class="flex -space-x-1.5"
        >
          <Avatar
            v-for="person in involvedPeople.slice(0, 4)"
            :key="person.id"
            class="size-6 border-2 border-background"
            :title="`${person.firstName} ${person.lastName}`"
          >
            <AvatarFallback class="text-[10px]">
              {{ person.firstName?.[0] || "" }}{{ person.lastName?.[0] || "" }}
            </AvatarFallback>
          </Avatar>
          <Avatar
            v-if="involvedPeople.length > 4"
            class="size-6 border-2 border-background"
            :title="`${involvedPeople.length - 4} more`"
          >
            <AvatarFallback class="text-[10px] bg-muted">
              +{{ involvedPeople.length - 4 }}
            </AvatarFallback>
          </Avatar>
        </div>
      </div>

      <!-- Col 6: Comment count -->
      <div class="flex items-center justify-end">
        <div
          class="flex items-center gap-1 rounded-full bg-muted px-1.5 py-0.5"
          :title="`${task.comment_count} comment${task.comment_count !== 1 ? 's' : ''}`"
        >
          <MessageSquare class="size-3 text-muted-foreground" />
          <span class="text-xs font-medium text-muted-foreground">{{ task.comment_count }}</span>
        </div>
      </div>
    </div>
  </NuxtLink>
</template>

<style scoped>
.task-row {
  grid-template-columns: 6rem 1fr 10rem 7rem 6rem 3rem;
  column-gap: 0.375rem;
}
</style>

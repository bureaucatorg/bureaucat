<script setup lang="ts">
import type { Task, ProjectState } from "~/types";

const props = withDefaults(
  defineProps<{
    tasks: Task[];
    // Single-project usage (project page): all tasks share one project key,
    // one set of states, and one membership flag.
    projectKey?: string;
    states?: ProjectState[];
    isMember?: boolean;
    // Multi-project usage (dashboard): tasks span projects, so states and
    // membership are resolved per task via its project_key. These take
    // precedence over the single-value props above when provided.
    statesByProject?: Record<string, ProjectState[]>;
    isMemberByProject?: Record<string, boolean>;
  }>(),
  { states: () => [], isMember: false }
);

const emit = defineEmits<{
  updated: [];
}>();

function statesFor(task: Task): ProjectState[] {
  return props.statesByProject?.[task.project_key] ?? props.states;
}

function isMemberFor(task: Task): boolean {
  if (props.isMemberByProject) return props.isMemberByProject[task.project_key] ?? false;
  return props.isMember;
}
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-border/50 divide-y divide-border/50">
    <TaskCard
      v-for="task in tasks"
      :key="task.id"
      :task="task"
      :project-key="projectKey ?? task.project_key"
      :states="statesFor(task)"
      :is-member="isMemberFor(task)"
      @updated="emit('updated')"
    />
  </div>
</template>

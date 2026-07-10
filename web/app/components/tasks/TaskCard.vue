<script setup lang="ts">
import { Circle, CircleDot, CheckCircle2, XCircle, Clock, MessageSquare, Building2 } from "lucide-vue-next";
import type { Task, ProjectState } from "~/types";
import { PRIORITY_LABELS } from "~/types";

const props = withDefaults(
  defineProps<{
    task: Task;
    projectKey?: string;
    states?: ProjectState[];
    isMember?: boolean;
    // When true, a leading checkbox is shown and the row toggles selection
    // instead of navigating on the checkbox itself.
    selectable?: boolean;
    selected?: boolean;
    // Multi-project usage (dashboard): show the task's workspace as a leading
    // column. Redundant inside a single project, so off by default.
    showWorkspace?: boolean;
    workspaceName?: string;
  }>(),
  {
    states: () => [],
    isMember: false,
    selectable: false,
    selected: false,
    showWorkspace: false,
    workspaceName: "",
  }
);

const emit = defineEmits<{
  updated: [];
  toggleSelect: [];
}>();

// Fall back to the task's own project_key (e.g. on the dashboard, where tasks
// span multiple projects) when no explicit key is passed.
const resolvedKey = computed(() => props.projectKey ?? props.task.project_key);

const { updateTask } = useTasks();
const updatingState = ref(false);

// Inline state editing is only offered to members who have states to pick from.
const canEditState = computed(() => props.isMember && props.states.length > 0);

async function changeState(stateId: string) {
  if (stateId === props.task.state_id || updatingState.value) return;
  updatingState.value = true;
  const res = await updateTask(resolvedKey.value, props.task.task_number, { state_id: stateId });
  updatingState.value = false;
  if (res.success) emit("updated");
}

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
  const people: { id: string; firstName: string; lastName: string; avatarUrl?: string }[] = [];
  const seen = new Set<string>();

  if (props.task.created_by && !seen.has(props.task.created_by)) {
    seen.add(props.task.created_by);
    people.push({
      id: props.task.created_by,
      firstName: props.task.creator_first_name || "",
      lastName: props.task.creator_last_name || "",
      avatarUrl: props.task.creator_avatar_url,
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
          avatarUrl: a.avatar_url,
        });
      }
    }
  }

  return people;
});
</script>

<template>
  <NuxtLink :to="`/projects/${resolvedKey}/tasks/${task.task_number}`" class="block">
    <div
      class="task-row group grid items-center bg-background/50 px-3 py-2.5 transition-colors hover:bg-muted/50"
      :class="{ selectable, 'has-workspace': showWorkspace, 'bg-accent/40': selectable && selected }"
    >
      <!-- Col 0: Selection checkbox -->
      <div v-if="selectable" class="justify-self-center" @click.stop.prevent>
        <Checkbox :model-value="selected" @update:model-value="emit('toggleSelect')" />
      </div>

      <!-- Col: Workspace (multi-project lists only) -->
      <div v-if="showWorkspace" class="min-w-0">
        <span
          class="inline-flex max-w-full items-center gap-1 rounded-md border bg-muted/50 px-1.5 py-0.5"
          :title="workspaceName"
        >
          <Building2 class="size-3 shrink-0 text-muted-foreground" />
          <span class="truncate text-xs text-muted-foreground">{{ workspaceName || "—" }}</span>
        </span>
      </div>

      <!-- Col 1: Task ID -->
      <span class="font-mono text-sm text-muted-foreground truncate">{{ task.task_id }}</span>

      <!-- Col 2: Title -->
      <span class="truncate text-sm font-medium min-w-0">{{ task.title }}</span>

      <!-- Col 3: State badge (editable for members) -->
      <div class="justify-self-end" @click.stop.prevent>
        <TaskStateSelector
          v-if="canEditState"
          :states="states"
          :model-value="task.state_id"
          :disabled="updatingState"
          compact
          dense
          @update:model-value="changeState"
        />
        <div
          v-else
          class="flex items-center gap-1 rounded-md border bg-muted/50 px-1.5 py-0.5 w-fit"
        >
          <component
            :is="stateIcon"
            class="size-3.5 shrink-0 stroke-[2.5]"
            :style="{ color: task.state_color }"
          />
          <span class="text-xs text-muted-foreground whitespace-nowrap">{{ task.state_name }}</span>
        </div>
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
          <NuxtLink
            v-for="person in involvedPeople.slice(0, 4)"
            :key="person.id"
            :to="`/profile/${person.id}`"
            :title="`${person.firstName} ${person.lastName}`"
            class="hover:z-10"
            @click.stop
          >
            <Avatar class="size-6 border-2 border-background transition-transform hover:scale-110">
              <AvatarImage
                v-if="person.avatarUrl"
                :src="person.avatarUrl"
                :alt="`${person.firstName} ${person.lastName}`"
              />
              <AvatarFallback class="text-[10px]" :seed="person.id">
                {{ person.firstName?.[0] || "" }}{{ person.lastName?.[0] || "" }}
              </AvatarFallback>
            </Avatar>
          </NuxtLink>
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
          <span class="font-mono text-xs font-medium text-muted-foreground">{{ task.comment_count }}</span>
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
.task-row.selectable {
  grid-template-columns: 1.5rem 6rem 1fr 10rem 7rem 6rem 3rem;
}
.task-row.has-workspace {
  grid-template-columns: 8rem 6rem 1fr 10rem 7rem 6rem 3rem;
}
.task-row.has-workspace.selectable {
  grid-template-columns: 1.5rem 8rem 6rem 1fr 10rem 7rem 6rem 3rem;
}
</style>

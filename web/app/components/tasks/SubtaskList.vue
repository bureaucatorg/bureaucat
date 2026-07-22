<script setup lang="ts">
import { Maximize2 } from "lucide-vue-next";
import type { Subtask, ProjectState, ProjectMember, ProjectLabel } from "~/types";

const props = withDefaults(
  defineProps<{
    subtasks: Subtask[];
    projectKey: string;
    states?: ProjectState[];
    members?: ProjectMember[];
    labels?: ProjectLabel[];
    isMember?: boolean;
  }>(),
  { states: () => [], members: () => [], labels: () => [], isMember: false }
);

const emit = defineEmits<{ updated: [] }>();

const { updateTask } = useTasks();
const updatingId = ref<string | null>(null);

// Which subtask's detail card is currently open (null = none).
const openId = ref<string | null>(null);

// Inline state editing is only offered to members who have states to pick from.
const canEditState = computed(() => props.isMember && props.states.length > 0);

async function changeState(subtask: Subtask, stateId: string) {
  if (stateId === subtask.state_id || updatingId.value) return;
  updatingId.value = subtask.id;
  const res = await updateTask(props.projectKey, subtask.task_number, { state_id: stateId });
  updatingId.value = null;
  if (res.success) emit("updated");
}

// Collect all people involved: creator + assignees (deduplicated), matching the
// main task list's "users" column.
function involvedPeople(subtask: Subtask) {
  const people: { id: string; firstName: string; lastName: string; avatarUrl?: string }[] = [];
  const seen = new Set<string>();

  if (subtask.created_by && !seen.has(subtask.created_by)) {
    seen.add(subtask.created_by);
    people.push({
      id: subtask.created_by,
      firstName: subtask.creator_first_name || "",
      lastName: subtask.creator_last_name || "",
      avatarUrl: subtask.creator_avatar_url,
    });
  }

  for (const a of subtask.assignees ?? []) {
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

  return people;
}
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-border/50 divide-y divide-border/50">
    <Popover
      v-for="subtask in subtasks"
      :key="subtask.id"
      :open="openId === subtask.id"
      @update:open="(v: boolean) => (openId = v ? subtask.id : null)"
    >
      <PopoverTrigger as-child>
        <button
          type="button"
          class="block w-full text-left"
        >
          <div
            class="subtask-row group grid items-center bg-background/50 px-3 py-2.5 transition-colors hover:bg-muted/50"
          >
            <!-- Col 1: Task ID (opens the full page) -->
            <NuxtLink
              :to="`/projects/${projectKey}/tasks/${subtask.task_number}`"
              :title="`Open ${subtask.task_id}`"
              class="-my-2.5 flex w-fit items-center gap-1.5 rounded-md px-2 py-2.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
              @click.stop
            >
              <Maximize2 class="size-3.5 shrink-0 opacity-50" />
              <span class="font-mono text-sm truncate">{{ subtask.task_id }}</span>
            </NuxtLink>

            <!-- Col 2: Title -->
            <span class="truncate text-sm font-medium min-w-0">{{ subtask.title }}</span>

            <!-- Col 3: State badge (editable for members) -->
            <div class="justify-self-end" @click.stop.prevent>
              <TaskStateSelector
                v-if="canEditState"
                :states="states"
                :model-value="subtask.state_id"
                :disabled="updatingId === subtask.id"
                compact
                dense
                @update:model-value="(id: string) => changeState(subtask, id)"
              />
              <div
                v-else
                class="flex items-center gap-1 rounded-md border bg-muted/50 px-1.5 py-0.5 w-fit"
              >
                <span
                  class="size-2.5 shrink-0 rounded-full"
                  :style="{ backgroundColor: subtask.state_color }"
                />
                <span class="text-xs text-muted-foreground whitespace-nowrap">{{ subtask.state_name }}</span>
              </div>
            </div>

            <!-- Col 4: Stacked avatars -->
            <div class="flex items-center justify-end">
              <div v-if="involvedPeople(subtask).length > 0" class="flex -space-x-1.5">
                <NuxtLink
                  v-for="person in involvedPeople(subtask).slice(0, 4)"
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
                  v-if="involvedPeople(subtask).length > 4"
                  class="size-6 border-2 border-background"
                  :title="`${involvedPeople(subtask).length - 4} more`"
                >
                  <AvatarFallback class="text-[10px] bg-muted">
                    +{{ involvedPeople(subtask).length - 4 }}
                  </AvatarFallback>
                </Avatar>
              </div>
            </div>
          </div>
        </button>
      </PopoverTrigger>

      <PopoverContent
        align="start"
        :collision-padding="16"
        class="w-[48rem] max-w-[calc(100vw-2rem)] max-h-[min(75vh,var(--reka-popover-content-available-height))] overflow-y-auto shadow-lg"
      >
        <SubtaskDetailCard
          v-if="openId === subtask.id"
          :project-key="projectKey"
          :task-number="subtask.task_number"
          :states="states"
          :members="members"
          :project-labels="labels"
          :is-member="isMember"
          @updated="emit('updated')"
        />
      </PopoverContent>
    </Popover>
  </div>
</template>

<style scoped>
.subtask-row {
  grid-template-columns: auto 1fr 10rem 5rem;
  column-gap: 0.375rem;
}
</style>

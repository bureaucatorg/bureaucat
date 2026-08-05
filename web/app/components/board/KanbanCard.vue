<script setup lang="ts">
import { Maximize2 } from "lucide-vue-next";
import type { Task, ProjectState, ProjectMember, ProjectLabel } from "~/types";
import { PRIORITY_LABELS } from "~/types";

const props = withDefaults(
  defineProps<{
    task: Task;
    projectKey: string;
    isMember: boolean;
    /** Dragging is disabled under groupings where a move is ambiguous. */
    canDrag?: boolean;
    /** The id of the column this card is currently rendered inside. */
    columnId?: string;
    states?: ProjectState[];
    members?: ProjectMember[];
    labels?: ProjectLabel[];
  }>(),
  { canDrag: undefined, states: () => [], members: () => [], labels: () => [] }
);

const draggable = computed(() => props.canDrag ?? props.isMember);

const emit = defineEmits<{ refresh: [] }>();

const priorityInfo = computed(() => PRIORITY_LABELS[props.task.priority] || PRIORITY_LABELS[0]);
const isDragging = ref(false);
const detailOpen = ref(false);

function handleDragStart(event: DragEvent) {
  if (!draggable.value) {
    event.preventDefault();
    return;
  }
  isDragging.value = true;
  event.dataTransfer!.effectAllowed = "move";
  event.dataTransfer!.setData(
    "application/json",
    JSON.stringify({ task: props.task, fromColumnId: props.columnId ?? "" })
  );
}

function handleDragEnd() {
  isDragging.value = false;
}

function toggleDetail(open: boolean) {
  // A drag ends with a click on some browsers; never treat that as a card tap.
  if (isDragging.value) return;
  detailOpen.value = open;
}
</script>

<template>
  <Popover :open="detailOpen" @update:open="toggleDetail">
    <PopoverTrigger as-child>
      <div
        :draggable="draggable"
        role="button"
        tabindex="0"
        :aria-label="`Task ${task.task_id}: ${task.title}`"
        class="group cursor-pointer rounded-lg border bg-background p-3 text-left shadow-sm transition-all hover:border-amber-500/30 hover:shadow-md focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
        :class="{ 'cursor-grab active:cursor-grabbing': draggable }"
        @dragstart="handleDragStart"
        @dragend="handleDragEnd"
        @keydown.enter="toggleDetail(true)"
        @keydown.space.prevent="toggleDetail(true)"
      >
        <!-- Task ID and priority -->
        <div class="mb-2 flex items-center justify-between">
          <NuxtLink
            :to="`/projects/${projectKey}/tasks/${task.task_number}`"
            :title="`Open ${task.task_id}`"
            class="-mx-1 -my-0.5 flex items-center gap-1.5 rounded-md px-1 py-0.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            @click.stop
          >
            <Maximize2 class="size-3 shrink-0 opacity-50" />
            <span class="font-mono text-xs">{{ task.task_id }}</span>
          </NuxtLink>
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
            <NuxtLink
              v-for="assignee in task.assignees.slice(0, 3)"
              :key="assignee.id"
              :to="`/profile/${assignee.user_id}`"
              :title="`${assignee.first_name} ${assignee.last_name}`"
              class="hover:z-10"
              @click.stop
            >
              <Avatar class="size-5 border border-background transition-transform hover:scale-110">
                <AvatarFallback class="text-[10px]" :seed="assignee.user_id">
                  {{ assignee.first_name[0] }}{{ assignee.last_name[0] }}
                </AvatarFallback>
              </Avatar>
            </NuxtLink>
            <div
              v-if="task.assignees.length > 3"
              class="flex size-5 items-center justify-center rounded-full border border-background bg-muted text-[10px]"
            >
              +{{ task.assignees.length - 3 }}
            </div>
          </div>
        </div>
      </div>
    </PopoverTrigger>

    <PopoverContent
      align="start"
      :collision-padding="16"
      class="w-[48rem] max-w-[calc(100vw-2rem)] max-h-[min(75vh,var(--reka-popover-content-available-height))] overflow-y-auto shadow-lg"
    >
      <TaskDetailCard
        v-if="detailOpen"
        :project-key="projectKey"
        :task-number="task.task_number"
        :states="states"
        :members="members"
        :project-labels="labels"
        :is-member="isMember"
        @updated="emit('refresh')"
      />
    </PopoverContent>
  </Popover>
</template>

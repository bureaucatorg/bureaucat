<script setup lang="ts">
import {
  Maximize2,
  Loader2,
  Calendar as CalendarIcon,
  Clock,
  Circle,
  CircleDot,
  CheckCircle2,
  XCircle,
} from "lucide-vue-next";
import { marked } from "marked";
import type { Task } from "~/types";
import { PRIORITY_LABELS } from "~/types";

const props = defineProps<{
  projectKey: string;
  taskNumber: number;
}>();

const { getAuthHeader } = useAuth();

const task = ref<Task | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);

const stateIconMap: Record<string, typeof Circle> = {
  backlog: Clock,
  unstarted: Circle,
  started: CircleDot,
  completed: CheckCircle2,
  cancelled: XCircle,
};

const priority = computed(() => {
  const p = task.value?.priority ?? 0;
  return PRIORITY_LABELS[p] || PRIORITY_LABELS[0];
});

const renderedDescription = computed(() => {
  const desc = task.value?.description;
  if (!desc) return "";
  return desc.startsWith("<") ? desc : (marked(desc) as string);
});

const taskLink = computed(() => `/projects/${props.projectKey}/tasks/${props.taskNumber}`);

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

// Fetch full task details directly (not via the shared store) so opening the
// card from a parent task's page doesn't clobber the parent bound to the view.
async function loadTask() {
  loading.value = true;
  error.value = null;
  try {
    const response = await fetch(
      `/api/v1/projects/${props.projectKey}/tasks/${props.taskNumber}`,
      { headers: getAuthHeader() }
    );
    if (!response.ok) {
      const body = await response.json().catch(() => ({}));
      error.value = body.message || "Failed to load subtask";
      return;
    }
    task.value = await response.json();
  } catch {
    error.value = "Network error";
  } finally {
    loading.value = false;
  }
}

onMounted(loadTask);
</script>

<template>
  <div>
    <div v-if="loading" class="flex items-center justify-center py-12">
      <Loader2 class="size-5 animate-spin text-muted-foreground" />
    </div>

    <div v-else-if="error" class="py-12 text-center text-sm text-destructive">
      {{ error }}
    </div>

    <template v-else-if="task">
      <div class="mb-4 flex items-start justify-between gap-3">
        <div class="min-w-0 space-y-0.5">
          <span class="font-mono text-xs text-muted-foreground">{{ task.task_id }}</span>
          <h2 class="text-lg font-semibold leading-tight">{{ task.title }}</h2>
        </div>
        <Button variant="outline" size="sm" class="shrink-0 gap-1.5" as-child>
          <NuxtLink :to="taskLink">
            <Maximize2 class="size-3.5" />
            Open
          </NuxtLink>
        </Button>
      </div>

      <div class="space-y-4">
        <!-- Meta badges -->
        <div class="flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground">
          <div class="flex items-center gap-1 rounded-md border bg-muted/50 px-1.5 py-0.5">
            <component
              :is="stateIconMap[task.state_type] || Circle"
              class="size-3.5 stroke-[2.5]"
              :style="{ color: task.state_color }"
            />
            <span>{{ task.state_name }}</span>
          </div>
          <div class="flex items-center gap-1 rounded-md border bg-muted/50 px-1.5 py-0.5">
            <span
              class="size-2.5 rounded-full ring-1.5 ring-offset-1 ring-offset-background"
              :style="{ backgroundColor: priority.color, '--tw-ring-color': priority.color }"
            />
            <span>{{ priority.label }}</span>
          </div>
          <NuxtLink
            :to="`/profile/${task.created_by}`"
            class="flex items-center gap-1 rounded-md border bg-muted/50 py-0.5 pl-0.5 pr-1.5 hover:bg-muted transition-colors"
          >
            <Avatar class="size-4">
              <AvatarImage v-if="task.creator_avatar_url" :src="task.creator_avatar_url" />
              <AvatarFallback class="text-[10px]" :seed="task.created_by">
                {{ task.creator_first_name?.[0] }}{{ task.creator_last_name?.[0] }}
              </AvatarFallback>
            </Avatar>
            <span>{{ task.creator_first_name }} {{ task.creator_last_name }}</span>
          </NuxtLink>
          <span>created on {{ formatDate(task.created_at) }}</span>
        </div>

        <!-- Description -->
        <div>
          <h3 class="mb-1.5 text-sm font-medium text-muted-foreground">Description</h3>
          <div
            v-if="task.description"
            class="prose prose-sm max-w-none dark:prose-invert"
            v-html="renderedDescription"
          />
          <p v-else class="text-sm italic text-muted-foreground">No description</p>
        </div>

        <!-- Assignees -->
        <div v-if="task.assignees && task.assignees.length">
          <h3 class="mb-1.5 text-sm font-medium text-muted-foreground">Assignees</h3>
          <div class="flex flex-wrap gap-1.5">
            <NuxtLink
              v-for="a in task.assignees"
              :key="a.user_id"
              :to="`/profile/${a.user_id}`"
              class="flex items-center gap-1.5 rounded-md border bg-muted/50 py-0.5 pl-0.5 pr-2 text-sm hover:bg-muted transition-colors"
            >
              <Avatar class="size-5">
                <AvatarImage v-if="a.avatar_url" :src="a.avatar_url" />
                <AvatarFallback class="text-[10px]" :seed="a.user_id">
                  {{ a.first_name?.[0] }}{{ a.last_name?.[0] }}
                </AvatarFallback>
              </Avatar>
              <span>{{ a.first_name }} {{ a.last_name }}</span>
            </NuxtLink>
          </div>
        </div>

        <!-- Labels -->
        <div v-if="task.labels && task.labels.length">
          <h3 class="mb-1.5 text-sm font-medium text-muted-foreground">Labels</h3>
          <div class="flex flex-wrap gap-1.5">
            <span
              v-for="l in task.labels"
              :key="l.id"
              class="flex items-center gap-1.5 rounded-md border bg-muted/50 px-2 py-0.5 text-xs"
            >
              <span class="size-2 rounded-full" :style="{ backgroundColor: l.color }" />
              {{ l.name }}
            </span>
          </div>
        </div>

        <!-- Dates -->
        <div class="flex flex-wrap gap-x-6 gap-y-1.5 text-xs text-muted-foreground">
          <div v-if="task.start_date" class="flex items-center gap-1.5">
            <CalendarIcon class="size-3.5" />
            <span>Start {{ formatDateTime(task.start_date) }}</span>
          </div>
          <div v-if="task.due_date" class="flex items-center gap-1.5">
            <CalendarIcon class="size-3.5" />
            <span>Due {{ formatDateTime(task.due_date) }}</span>
          </div>
          <div class="flex items-center gap-1.5">
            <Clock class="size-3.5" />
            <span>Updated {{ formatDate(task.updated_at) }}</span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import {
  ChevronLeft,
  Loader2,
  Plus,
  Repeat,
  Trash2,
  CalendarDays,
  X,
  ChevronDown,
} from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { CycleAssigneeSummary } from "~/types";

definePageMeta({ middleware: ["auth"] });

const route = useRoute();
const router = useRouter();

const projectKey = computed(() => route.params.key as string);
const cycleId = computed(() => route.params.cycleId as string);

const {
  currentCycle,
  tasks,
  metrics,
  siblings,
  getCycle,
  listCycleTasks,
  getCycleMetrics,
  deleteCycle,
  removeTaskFromCycle,
  listSiblings,
  clearCurrent,
} = useCycles();
const { currentProject, getProject } = useProjects();

const isAdmin = computed(() => currentProject.value?.role === "admin");

const loading = ref(true);
const error = ref<string | null>(null);
const showAddTask = ref(false);
const showDeleteConfirm = ref(false);
const deleting = ref(false);
const assigneeFilter = ref<string | null>(null);
const switcherOpen = ref(false);

useHead({
  title: computed(
    () => `${currentCycle.value?.title ?? "Cycle"} · ${projectKey.value}`
  ),
});

const statusChip: Record<string, string> = {
  upcoming: "border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-300",
  active: "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
  completed: "border-muted-foreground/30 bg-muted text-muted-foreground",
};

function formatDate(d: string): string {
  if (!d) return "";
  const dt = new Date(d + "T00:00:00");
  return dt.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
}

const tableGridStyle = computed(() =>
  isAdmin.value
    ? "grid-template-columns: 140px minmax(0, 1fr) 90px 28px;"
    : "grid-template-columns: 140px minmax(0, 1fr) 90px;"
);

const progressPct = computed(() => {
  const m = metrics.value;
  if (!m || m.total === 0) return 0;
  return Math.round((m.completed / m.total) * 100);
});

const filteredAssigneeName = computed(() => {
  if (!assigneeFilter.value || !metrics.value) return null;
  const a = metrics.value.assignees.find(
    (a: CycleAssigneeSummary) => a.user_id === assigneeFilter.value
  );
  return a ? `${a.first_name} ${a.last_name}`.trim() || a.username : null;
});

async function loadAll() {
  loading.value = true;
  error.value = null;
  const [c, m, t, s] = await Promise.all([
    getCycle(projectKey.value, cycleId.value),
    getCycleMetrics(projectKey.value, cycleId.value),
    listCycleTasks(projectKey.value, cycleId.value, assigneeFilter.value),
    listSiblings(projectKey.value),
  ]);
  if (!c.success) error.value = c.error || "Failed to load cycle";
  if (!m.success && !error.value) error.value = m.error || "Failed to load metrics";
  if (!t.success && !error.value) error.value = t.error || "Failed to load tasks";
  void s;
  loading.value = false;
}

async function reloadTasksAndMetrics() {
  await Promise.all([
    listCycleTasks(projectKey.value, cycleId.value, assigneeFilter.value),
    getCycleMetrics(projectKey.value, cycleId.value),
  ]);
}

function setAssigneeFilter(userId: string | null) {
  assigneeFilter.value = userId;
  listCycleTasks(projectKey.value, cycleId.value, userId);
}

async function handleDelete() {
  deleting.value = true;
  const result = await deleteCycle(projectKey.value, cycleId.value);
  deleting.value = false;
  if (result.success) {
    toast.success("Cycle deleted");
    router.push(`/projects/${projectKey.value}?tab=cycles`);
  } else {
    toast.error(result.error || "Failed to delete cycle");
  }
}

async function handleRemoveTask(taskId: string) {
  const result = await removeTaskFromCycle(projectKey.value, cycleId.value, taskId);
  if (result.success) {
    toast.success("Task removed from cycle");
    reloadTasksAndMetrics();
  } else {
    toast.error(result.error || "Failed to remove task");
  }
}

onMounted(async () => {
  if (!currentProject.value || currentProject.value.project_key !== projectKey.value) {
    await getProject(projectKey.value);
  }
  await loadAll();
});

onBeforeUnmount(() => {
  clearCurrent();
});

watch(cycleId, async () => {
  assigneeFilter.value = null;
  await loadAll();
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />
    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-8">
        <nav class="mb-6 flex items-center gap-2 text-sm text-muted-foreground">
          <ChevronLeft class="size-4" />
          <NuxtLink to="/projects" class="hover:text-foreground">Projects</NuxtLink>
          <span>/</span>
          <NuxtLink
            :to="`/projects/${projectKey}`"
            class="hover:text-foreground"
          >
            {{ currentProject?.name ?? projectKey }}
          </NuxtLink>
          <span>/</span>
          <NuxtLink
            :to="`/projects/${projectKey}?tab=cycles`"
            class="hover:text-foreground"
          >
            Cycles
          </NuxtLink>
          <span>/</span>
          <span class="font-semibold text-amber-600 dark:text-amber-500">
            {{ currentCycle?.title ?? "…" }}
          </span>
        </nav>

        <div v-if="loading" class="flex items-center justify-center py-12">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <div
          v-else-if="error"
          class="rounded-md bg-destructive/10 p-4 text-sm text-destructive"
        >
          {{ error }}
        </div>

        <template v-else-if="currentCycle">
          <!-- Header -->
          <header class="mb-6 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div class="min-w-0">
              <div class="flex items-center gap-3">
                <div
                  class="flex size-10 items-center justify-center rounded-lg bg-muted"
                >
                  <Repeat class="size-5 text-amber-600 dark:text-amber-500" />
                </div>
                <h1 class="truncate text-2xl font-bold tracking-tight sm:text-3xl">
                  {{ currentCycle.title }}
                </h1>
                <span
                  :class="[
                    'rounded-md border px-2 py-0.5 text-[11px] font-medium uppercase tracking-wide',
                    statusChip[currentCycle.status] || statusChip.upcoming,
                  ]"
                >
                  {{ currentCycle.status }}
                </span>
              </div>
              <p class="mt-2 flex items-center gap-1.5 text-sm text-muted-foreground">
                <CalendarDays class="size-3.5" />
                {{ formatDate(currentCycle.start_date) }} →
                {{ formatDate(currentCycle.end_date) }}
              </p>
              <p
                v-if="currentCycle.description"
                class="mt-3 max-w-2xl whitespace-pre-wrap text-sm text-muted-foreground"
              >
                {{ currentCycle.description }}
              </p>
            </div>

            <div class="flex shrink-0 items-center gap-2">
              <!-- Cycle switcher -->
              <DropdownMenu v-model:open="switcherOpen">
                <DropdownMenuTrigger as-child>
                  <Button variant="outline" size="sm">
                    Switch cycle
                    <ChevronDown class="ml-1 size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" class="w-64 max-h-80 overflow-y-auto">
                  <DropdownMenuItem
                    v-for="s in siblings"
                    :key="s.id"
                    :disabled="s.id === cycleId"
                    @click="
                      s.id !== cycleId &&
                        router.push(`/projects/${projectKey}/cycles/${s.id}`)
                    "
                  >
                    <div class="flex min-w-0 flex-1 flex-col">
                      <span class="truncate">{{ s.title }}</span>
                      <span class="text-[10px] uppercase tracking-wide text-muted-foreground">
                        {{ s.status }} · {{ s.start_date }} → {{ s.end_date }}
                      </span>
                    </div>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>

              <Button
                v-if="isAdmin"
                variant="outline"
                size="sm"
                class="text-destructive hover:text-destructive"
                @click="showDeleteConfirm = true"
              >
                <Trash2 class="mr-1.5 size-4" />
                Delete
              </Button>
            </div>
          </header>

          <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
            <!-- LEFT: Task list -->
            <section class="min-w-0">
              <div class="mb-4 flex items-center justify-between">
                <h2 class="text-lg font-semibold">Tasks</h2>
                <Button v-if="isAdmin" size="sm" @click="showAddTask = true">
                  <Plus class="mr-1.5 size-4" />
                  Add Task
                </Button>
              </div>

              <div
                v-if="assigneeFilter"
                class="mb-3 inline-flex items-center gap-2 rounded-md border bg-muted/50 px-2 py-1 text-xs"
              >
                Filtering by
                <span class="font-medium">{{ filteredAssigneeName }}</span>
                <button
                  class="rounded p-0.5 hover:bg-muted"
                  aria-label="Clear filter"
                  @click="setAssigneeFilter(null)"
                >
                  <X class="size-3" />
                </button>
              </div>

              <div
                v-if="tasks.length === 0"
                class="flex flex-col items-center rounded-lg border border-dashed py-12 text-center"
              >
                <Repeat class="size-6 text-muted-foreground" />
                <p class="mt-3 text-sm text-muted-foreground">
                  {{ assigneeFilter ? "No tasks for this assignee." : "No tasks in this cycle yet." }}
                </p>
                <Button v-if="isAdmin && !assigneeFilter" class="mt-4" size="sm" @click="showAddTask = true">
                  <Plus class="mr-1.5 size-4" /> Add Task
                </Button>
              </div>

              <div v-else class="overflow-hidden rounded-lg border bg-background">
                <div
                  class="grid items-center gap-3 border-b bg-muted/40 px-4 py-2 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground"
                  :style="tableGridStyle"
                >
                  <span>State</span>
                  <span>Title</span>
                  <span>ID</span>
                  <span v-if="isAdmin"></span>
                </div>

                <div class="max-h-[70vh] overflow-y-auto [scrollbar-gutter:stable]">
                  <div
                    v-for="task in tasks"
                    :key="task.id"
                    class="group grid items-center gap-3 border-b border-border/40 px-4 py-2.5 text-sm transition-colors last:border-0 hover:bg-muted/40"
                    :style="tableGridStyle"
                  >
                    <span
                      class="inline-flex w-fit max-w-full items-center truncate rounded px-1.5 py-0.5 font-mono text-[10px] font-medium uppercase tracking-wider"
                      :style="{
                        backgroundColor: (task.state_color || '#6B7280') + '22',
                        color: task.state_color || '#6B7280',
                      }"
                    >
                      {{ task.state_name }}
                    </span>

                    <NuxtLink
                      :to="`/projects/${projectKey}/tasks/${task.task_number}`"
                      class="min-w-0 truncate font-medium text-foreground hover:text-amber-600 hover:underline dark:hover:text-amber-500"
                    >
                      {{ task.title }}
                    </NuxtLink>

                    <span class="font-mono text-[11px] text-muted-foreground">
                      {{ task.task_id }}
                    </span>

                    <button
                      v-if="isAdmin"
                      class="rounded p-1 text-muted-foreground opacity-0 transition-opacity hover:bg-muted hover:text-destructive focus-visible:opacity-100 group-hover:opacity-100"
                      :aria-label="`Remove ${task.title} from cycle`"
                      @click.prevent="handleRemoveTask(task.id)"
                    >
                      <X class="size-3.5" />
                    </button>
                  </div>
                </div>
              </div>
            </section>

            <!-- RIGHT: Overview -->
            <aside class="space-y-6">
              <!-- Progress -->
              <section class="rounded-lg border p-4">
                <h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  Progress
                </h3>
                <div class="flex items-baseline gap-2">
                  <span class="text-3xl font-bold tabular-nums">{{ progressPct }}%</span>
                  <span class="text-sm text-muted-foreground">
                    {{ metrics?.completed ?? 0 }} / {{ metrics?.total ?? 0 }} done
                  </span>
                </div>
                <div class="mt-3 h-2 w-full overflow-hidden rounded-full bg-muted">
                  <div
                    class="h-full rounded-full bg-amber-500 transition-all"
                    :style="{ width: progressPct + '%' }"
                  />
                </div>
                <dl class="mt-4 grid grid-cols-2 gap-2 text-xs">
                  <div class="flex justify-between">
                    <dt class="text-muted-foreground">Todo</dt>
                    <dd class="font-medium tabular-nums">{{ metrics?.todo ?? 0 }}</dd>
                  </div>
                  <div class="flex justify-between">
                    <dt class="text-muted-foreground">In progress</dt>
                    <dd class="font-medium tabular-nums">{{ metrics?.in_progress ?? 0 }}</dd>
                  </div>
                  <div class="flex justify-between">
                    <dt class="text-muted-foreground">Done</dt>
                    <dd class="font-medium tabular-nums">{{ metrics?.completed ?? 0 }}</dd>
                  </div>
                  <div class="flex justify-between">
                    <dt class="text-muted-foreground">Cancelled</dt>
                    <dd class="font-medium tabular-nums">{{ metrics?.cancelled ?? 0 }}</dd>
                  </div>
                </dl>
              </section>

              <!-- State breakdown -->
              <section
                v-if="metrics && metrics.state_breakdown.length"
                class="rounded-lg border p-4"
              >
                <h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  By state
                </h3>
                <ul class="space-y-1.5">
                  <li
                    v-for="b in metrics.state_breakdown"
                    :key="b.state_id"
                    class="flex items-center justify-between gap-2 text-sm"
                  >
                    <span class="flex items-center gap-2 truncate">
                      <span
                        class="size-2 rounded-full"
                        :style="{ backgroundColor: b.state_color || '#6B7280' }"
                      />
                      <span class="truncate">{{ b.state_name }}</span>
                    </span>
                    <span class="font-medium tabular-nums text-muted-foreground">
                      {{ b.count }}
                    </span>
                  </li>
                </ul>
              </section>

              <!-- Assignees -->
              <section
                v-if="metrics && metrics.assignees.length"
                class="rounded-lg border p-4"
              >
                <h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  Assignees
                </h3>
                <ul class="space-y-1">
                  <li v-for="a in metrics.assignees" :key="a.user_id">
                    <button
                      type="button"
                      class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors"
                      :class="
                        assigneeFilter === a.user_id
                          ? 'bg-amber-500/10 text-amber-700 dark:text-amber-400'
                          : 'hover:bg-muted'
                      "
                      @click="
                        setAssigneeFilter(
                          assigneeFilter === a.user_id ? null : a.user_id
                        )
                      "
                    >
                      <Avatar class="size-6">
                        <AvatarImage
                          v-if="a.avatar_url"
                          :src="a.avatar_url"
                          :alt="a.first_name"
                        />
                        <AvatarFallback class="text-[9px]" :seed="a.user_id">
                          {{ (a.first_name[0] || "") + (a.last_name[0] || "") }}
                        </AvatarFallback>
                      </Avatar>
                      <span class="min-w-0 flex-1 truncate">
                        {{ `${a.first_name} ${a.last_name}`.trim() || a.username }}
                      </span>
                      <span class="font-medium tabular-nums text-muted-foreground">
                        {{ a.task_count }}
                      </span>
                    </button>
                  </li>
                </ul>
              </section>
            </aside>
          </div>

          <AddTaskToCycleDialog
            v-model:open="showAddTask"
            :project-key="projectKey"
            :cycle-id="cycleId"
            @added="reloadTasksAndMetrics"
          />

          <Dialog v-model:open="showDeleteConfirm">
            <DialogContent class="sm:max-w-md">
              <DialogHeader>
                <DialogTitle>Delete cycle?</DialogTitle>
                <DialogDescription>
                  Tasks assigned to this cycle will be unassigned but not deleted.
                  This can't be undone.
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  :disabled="deleting"
                  @click="showDeleteConfirm = false"
                >
                  Cancel
                </Button>
                <Button variant="destructive" :disabled="deleting" @click="handleDelete">
                  <Loader2 v-if="deleting" class="mr-2 size-4 animate-spin" />
                  Delete
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </template>
      </div>
    </main>
  </div>
</template>

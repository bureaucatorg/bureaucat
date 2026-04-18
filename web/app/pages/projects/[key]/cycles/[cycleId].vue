<script setup lang="ts">
import {
  ChevronLeft,
  Loader2,
  Plus,
  Pencil,
  Repeat,
  Trash2,
  CalendarDays,
  X,
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
  getCycle,
  listCycleTasks,
  getCycleMetrics,
  deleteCycle,
  removeTaskFromCycle,
  listUnassignedTasks,
  addTasksToCycle,
  clearCurrent,
} = useCycles();

async function loadCyclePickerTasks(search: string, limit: number) {
  return await listUnassignedTasks(projectKey.value, search, limit);
}
async function addTasksToCurrentCycle(taskIds: string[]) {
  return await addTasksToCycle(projectKey.value, cycleId.value, taskIds);
}
const { currentProject, getProject } = useProjects();

const isAdmin = computed(() => currentProject.value?.role === "admin");

const loading = ref(true);
const error = ref<string | null>(null);
const showAddTask = ref(false);
const showEdit = ref(false);
const showDeleteConfirm = ref(false);
const deleting = ref(false);
const assigneeFilter = ref<string | null>(null);

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
  const [c, m, t] = await Promise.all([
    getCycle(projectKey.value, cycleId.value),
    getCycleMetrics(projectKey.value, cycleId.value),
    listCycleTasks(projectKey.value, cycleId.value, assigneeFilter.value),
  ]);
  if (!c.success) error.value = c.error || "Failed to load cycle";
  if (!m.success && !error.value) error.value = m.error || "Failed to load metrics";
  if (!t.success && !error.value) error.value = t.error || "Failed to load tasks";
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
              <Button
                v-if="isAdmin"
                variant="outline"
                size="sm"
                @click="showEdit = true"
              >
                <Pencil class="mr-1.5 size-4" />
                Edit
              </Button>
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

              <CollectionTaskTable
                v-else
                :tasks="tasks"
                :project-key="projectKey"
                :is-admin="isAdmin"
                remove-label="Remove from cycle:"
                @remove="handleRemoveTask"
              />
            </section>

            <!-- RIGHT: Overview -->
            <aside class="space-y-6">
              <ProgressCard :metrics="metrics" />

              <StateBreakdownCard
                v-if="metrics"
                :buckets="metrics.state_breakdown"
              />

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

          <AddTasksDialog
            v-model:open="showAddTask"
            :project-key="projectKey"
            :collection-id="cycleId"
            title="Add tasks to cycle"
            description="Pick tasks that aren't yet in a cycle, or create a brand new one."
            empty-hint="No unassigned tasks found."
            :load-tasks="loadCyclePickerTasks"
            :add-tasks="addTasksToCurrentCycle"
            @added="reloadTasksAndMetrics"
          />

          <CreateCycleDialog
            v-model:open="showEdit"
            :project-key="projectKey"
            :cycle="currentCycle"
            @saved="() => getCycle(projectKey, cycleId)"
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

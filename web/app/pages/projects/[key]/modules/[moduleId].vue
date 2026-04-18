<script setup lang="ts">
import {
  ChevronLeft,
  Loader2,
  Plus,
  Layers,
  Trash2,
  Copy,
  Pencil,
  CalendarDays,
} from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { ModuleStatus } from "~/types";
import { MODULE_STATUSES } from "~/types";

definePageMeta({ middleware: ["auth"] });

const route = useRoute();
const router = useRouter();

const projectKey = computed(() => route.params.key as string);
const moduleId = computed(() => route.params.moduleId as string);

const {
  currentModule,
  tasks,
  metrics,
  members,
  getModule,
  updateModule,
  deleteModule,
  listModuleTasks,
  listPickerTasks,
  addTasksToModule,
  removeTaskFromModule,
  listModuleMembers,
  getModuleMetrics,
  clearCurrent,
} = useModules();
const { currentProject, getProject } = useProjects();

const isAdmin = computed(() => currentProject.value?.role === "admin");

const loading = ref(true);
const error = ref<string | null>(null);
const showAddTask = ref(false);
const showEdit = ref(false);
const showDuplicate = ref(false);
const showDeleteConfirm = ref(false);
const deleting = ref(false);

useHead({
  title: computed(
    () => `${currentModule.value?.title ?? "Module"} · ${projectKey.value}`
  ),
});

const statusChip: Record<string, string> = {
  backlog:     "border-muted-foreground/30 bg-muted text-muted-foreground",
  planned:     "border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-300",
  in_progress: "border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-500",
  paused:      "border-orange-500/30 bg-orange-500/10 text-orange-700 dark:text-orange-300",
  completed:   "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
  cancelled:   "border-rose-500/30 bg-rose-500/10 text-rose-700 dark:text-rose-300",
};

function formatDate(d?: string): string {
  if (!d) return "";
  const dt = new Date(d + "T00:00:00");
  return dt.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
}

function formatRange(a?: string, b?: string): string {
  if (!a && !b) return "No dates set";
  if (a && !b) return `From ${formatDate(a)}`;
  if (!a && b) return `Until ${formatDate(b)}`;
  return `${formatDate(a)} → ${formatDate(b)}`;
}


async function loadAll() {
  loading.value = true;
  error.value = null;
  const [m, met, t, mem] = await Promise.all([
    getModule(projectKey.value, moduleId.value),
    getModuleMetrics(projectKey.value, moduleId.value),
    listModuleTasks(projectKey.value, moduleId.value),
    listModuleMembers(projectKey.value, moduleId.value),
  ]);
  if (!m.success) error.value = m.error || "Failed to load module";
  if (!met.success && !error.value) error.value = met.error || "Failed to load metrics";
  if (!t.success && !error.value) error.value = t.error || "Failed to load tasks";
  void mem;
  loading.value = false;
}

async function reloadTasksAndMetrics() {
  await Promise.all([
    listModuleTasks(projectKey.value, moduleId.value),
    getModuleMetrics(projectKey.value, moduleId.value),
    listModuleMembers(projectKey.value, moduleId.value),
  ]);
}

async function reloadModule() {
  await Promise.all([
    getModule(projectKey.value, moduleId.value),
    listModuleMembers(projectKey.value, moduleId.value),
  ]);
}

async function changeStatus(status: ModuleStatus) {
  if (!currentModule.value || currentModule.value.status === status) return;
  const result = await updateModule(projectKey.value, moduleId.value, { status });
  if (result.success) {
    toast.success(`Status set to ${status.replace("_", " ")}`);
  } else {
    toast.error(result.error || "Failed to update status");
  }
}

async function handleDelete() {
  deleting.value = true;
  const result = await deleteModule(projectKey.value, moduleId.value);
  deleting.value = false;
  if (result.success) {
    toast.success("Module deleted");
    router.push(`/projects/${projectKey.value}?tab=modules`);
  } else {
    toast.error(result.error || "Failed to delete module");
  }
}

async function handleRemoveTask(taskId: string) {
  const result = await removeTaskFromModule(projectKey.value, moduleId.value, taskId);
  if (result.success) {
    toast.success("Task removed from module");
    reloadTasksAndMetrics();
  } else {
    toast.error(result.error || "Failed to remove task");
  }
}

async function loadPickerTasks(search: string, limit: number) {
  return await listPickerTasks(projectKey.value, moduleId.value, search, limit);
}
async function addTasksToCurrentModule(taskIds: string[]) {
  return await addTasksToModule(projectKey.value, moduleId.value, taskIds);
}

function onDuplicated(newModule: { id: string }) {
  router.push(`/projects/${projectKey.value}/modules/${newModule.id}`);
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

watch(moduleId, async () => {
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
          <NuxtLink :to="`/projects/${projectKey}`" class="hover:text-foreground">
            {{ currentProject?.name ?? projectKey }}
          </NuxtLink>
          <span>/</span>
          <NuxtLink
            :to="`/projects/${projectKey}?tab=modules`"
            class="hover:text-foreground"
          >
            Modules
          </NuxtLink>
          <span>/</span>
          <span class="font-semibold text-amber-600 dark:text-amber-500">
            {{ currentModule?.title ?? "…" }}
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

        <template v-else-if="currentModule">
          <header class="mb-6 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div class="min-w-0">
              <div class="flex items-center gap-3">
                <div class="flex size-10 items-center justify-center rounded-lg bg-muted">
                  <Layers class="size-5 text-amber-600 dark:text-amber-500" />
                </div>
                <h1 class="truncate text-2xl font-bold tracking-tight sm:text-3xl">
                  {{ currentModule.title }}
                </h1>
              </div>

              <div class="mt-3 flex flex-wrap items-center gap-2">
                <!-- Status pill / dropdown -->
                <DropdownMenu v-if="isAdmin">
                  <DropdownMenuTrigger as-child>
                    <button
                      type="button"
                      :class="[
                        'rounded-md border px-2 py-0.5 text-[11px] font-medium uppercase tracking-wide hover:opacity-80',
                        statusChip[currentModule.status] || statusChip.backlog,
                      ]"
                    >
                      {{ currentModule.status.replace("_", " ") }}
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="start">
                    <DropdownMenuItem
                      v-for="s in MODULE_STATUSES"
                      :key="s"
                      :disabled="currentModule.status === s"
                      @click="changeStatus(s)"
                    >
                      {{ s.replace("_", " ") }}
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
                <span
                  v-else
                  :class="[
                    'rounded-md border px-2 py-0.5 text-[11px] font-medium uppercase tracking-wide',
                    statusChip[currentModule.status] || statusChip.backlog,
                  ]"
                >
                  {{ currentModule.status.replace("_", " ") }}
                </span>

                <span class="flex items-center gap-1.5 text-sm text-muted-foreground">
                  <CalendarDays class="size-3.5" />
                  {{ formatRange(currentModule.start_date, currentModule.end_date) }}
                </span>

                <span
                  v-if="currentModule.lead"
                  class="flex items-center gap-1.5 text-sm text-muted-foreground"
                >
                  <span class="text-muted-foreground">Lead:</span>
                  <Avatar class="size-5">
                    <AvatarImage
                      v-if="currentModule.lead.avatar_url"
                      :src="currentModule.lead.avatar_url"
                    />
                    <AvatarFallback class="text-[9px]" :seed="currentModule.lead.user_id">
                      {{ (currentModule.lead.first_name[0] || "") + (currentModule.lead.last_name[0] || "") }}
                    </AvatarFallback>
                  </Avatar>
                  <span class="font-medium text-foreground">
                    {{ currentModule.lead.first_name }} {{ currentModule.lead.last_name }}
                  </span>
                </span>
              </div>

              <p
                v-if="currentModule.description"
                class="mt-3 max-w-2xl whitespace-pre-wrap text-sm text-muted-foreground"
              >
                {{ currentModule.description }}
              </p>
            </div>

            <div class="flex shrink-0 flex-wrap items-center gap-2">
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
                @click="showDuplicate = true"
              >
                <Copy class="mr-1.5 size-4" />
                Duplicate
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
                v-if="tasks.length === 0"
                class="rounded-lg border border-dashed py-10 text-center text-sm text-muted-foreground"
              >
                No tasks linked yet.
                <div v-if="isAdmin" class="mt-3">
                  <Button size="sm" @click="showAddTask = true">
                    <Plus class="mr-1.5 size-4" />
                    Add Task
                  </Button>
                </div>
              </div>

              <CollectionTaskTable
                v-else
                :tasks="tasks"
                :project-key="projectKey"
                :is-admin="isAdmin"
                remove-label="Remove from module:"
                @remove="handleRemoveTask"
              />
            </section>

            <!-- RIGHT: Metrics + Members -->
            <aside class="space-y-6">
              <ProgressCard :metrics="metrics" />

              <StateBreakdownCard
                v-if="metrics"
                :buckets="metrics.state_breakdown"
              />

              <!-- Members card (bordered, matches cycles aesthetic) -->
              <section class="rounded-lg border p-4">
                <ModuleMembersPanel
                  :project-key="projectKey"
                  :module-id="moduleId"
                  :lead-id="currentModule.lead?.user_id"
                  :members="members"
                  :is-admin="isAdmin"
                  @changed="reloadModule"
                />
              </section>
            </aside>
          </div>

          <AddTasksDialog
            v-model:open="showAddTask"
            :project-key="projectKey"
            :collection-id="moduleId"
            title="Add tasks to module"
            description="Pick tasks to link, or create a brand new one. Assignees are auto-added as module members."
            empty-hint="No eligible tasks found."
            :load-tasks="loadPickerTasks"
            :add-tasks="addTasksToCurrentModule"
            @added="reloadTasksAndMetrics"
          />

          <CreateModuleDialog
            v-model:open="showEdit"
            :project-key="projectKey"
            :module="currentModule"
            @saved="reloadModule"
          />

          <DuplicateModuleDialog
            v-model:open="showDuplicate"
            :project-key="projectKey"
            :source="currentModule"
            :source-tasks="tasks"
            @duplicated="onDuplicated"
          />

          <Dialog v-model:open="showDeleteConfirm">
            <DialogContent class="sm:max-w-md">
              <DialogHeader>
                <DialogTitle>Delete module?</DialogTitle>
                <DialogDescription>
                  Tasks linked to this module will be unlinked but not deleted.
                  This can't be undone.
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <Button variant="outline" :disabled="deleting" @click="showDeleteConfirm = false">
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

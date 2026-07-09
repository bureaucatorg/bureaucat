<script setup lang="ts">
import {
  ListTodo,
  Kanban,
  FileText,
  Repeat,
  Layers,
  Users,
  Settings,
  Plus,
  Loader2,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  Eye,
  Save,
  FolderInput,
  X,
  Lock,
} from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { FilterTree, ProjectView, MoveTasksResponse } from "~/types";

definePageMeta({
  middleware: ["auth"],
});

const route = useRoute();
const router = useRouter();
const projectKey = computed(() => route.params.key as string);

// Valid tab values
const validTabs = ["tasks", "board", "pages", "cycles", "modules", "views", "members", "settings"] as const;
type TabValue = (typeof validTabs)[number];

const activeTab = computed({
  get: () => {
    const tab = route.query.tab as string;
    return validTabs.includes(tab as TabValue) ? tab : "tasks";
  },
  set: (value: string) => {
    router.replace({
      query: { ...route.query, tab: value === "tasks" ? undefined : value },
    });
  },
});

const {
  currentProject,
  members,
  states,
  labels,
  templates,
  getProject,
  listMembers,
  listStates,
  listLabels,
  listTemplates,
} = useProjects();

useHead({
  title: computed(() => currentProject.value?.name ?? projectKey.value),
});

const {
  tasks,
  loading: tasksLoading,
  total: totalTasks,
  page: tasksPage,
  totalPages: tasksTotalPages,
  listTasks,
} = useTasks();

const { user } = useAuth();
const currentUserId = computed(() => user.value?.id);

const {
  views,
  listViews,
  getView,
} = useViews();

const {
  tree,
  setTree,
  clearTreeAndView,
  clearAll,
  sortBy,
  sortDir,
  groupBy,
  activeViewSlug,
  setActiveView,
  searchQuery,
  effectiveTree,
  hydrateFromUrl,
  encodeTree,
} = useFilterTree();

const loading = ref(true);
const error = ref<string | null>(null);
const showCreateTask = ref(false);
const showAddMember = ref(false);
const showSaveView = ref(false);
const renameViewTarget = ref<ProjectView | null>(null);

const isAdmin = computed(() => currentProject.value?.role === "admin");
const isMember = computed(
  () => currentProject.value?.role === "admin" || currentProject.value?.role === "member"
);
const isDisabled = computed(() => currentProject.value?.disabled ?? false);
// A member can mutate the project only while it is enabled.
const canWrite = computed(() => isMember.value && !isDisabled.value);

// Bulk task selection / move.
const selectedTasks = ref<Set<number>>(new Set());
const showBulkMove = ref(false);

function toggleTaskSelection(taskNumber: number) {
  if (selectedTasks.value.has(taskNumber)) selectedTasks.value.delete(taskNumber);
  else selectedTasks.value.add(taskNumber);
  selectedTasks.value = new Set(selectedTasks.value);
}

function clearSelection() {
  selectedTasks.value = new Set();
}

// True when every task on the current page is selected.
const allSelected = computed(
  () => tasks.value.length > 0 && tasks.value.every((t) => selectedTasks.value.has(t.task_number))
);

function toggleSelectAll() {
  if (allSelected.value) {
    selectedTasks.value = new Set();
  } else {
    selectedTasks.value = new Set(tasks.value.map((t) => t.task_number));
  }
}

async function handleBulkMoved(payload: { targetKey: string; result?: MoveTasksResponse }) {
  const result = payload.result;
  if (result) {
    if (result.failed > 0) {
      toast.warning(`Moved ${result.moved} task${result.moved === 1 ? "" : "s"}, ${result.failed} failed`);
    } else {
      toast.success(`Moved ${result.moved} task${result.moved === 1 ? "" : "s"}`);
    }
  }
  selectedTasks.value = new Set();
  await loadTasks(tasksPage.value);
}

const currentPageFromUrl = computed(() => {
  const p = parseInt(route.query.page as string, 10);
  return Number.isFinite(p) && p > 0 ? p : 1;
});

function setPageInUrl(page: number) {
  router.replace({
    query: { ...route.query, page: page > 1 ? String(page) : undefined },
  });
}

async function loadProject() {
  loading.value = true;
  error.value = null;

  const result = await getProject(projectKey.value);
  if (!result.success) {
    error.value = result.error || "Failed to load project";
    loading.value = false;
    return;
  }

  await Promise.all([
    listMembers(projectKey.value),
    listStates(projectKey.value),
    listLabels(projectKey.value),
    listTemplates(projectKey.value),
    listViews(projectKey.value),
  ]);

  // If the URL referenced a saved view but carried no ?f=, hydrate the filters
  // from the stored view so the chip row and group-by match what's running.
  if (activeViewSlug.value && tree.value.children.length === 0) {
    const res = await getView(projectKey.value, activeViewSlug.value);
    if (res.success && res.data) {
      setTree(res.data.filter_tree);
      sortBy.value = res.data.sort_by;
      sortDir.value = res.data.sort_dir;
      groupBy.value = res.data.group_by;
    } else {
      // View disappeared or became inaccessible — drop the stale slug.
      setActiveView(null);
    }
  }

  await loadTasks(currentPageFromUrl.value);
  loading.value = false;
}

async function loadTasks(page = 1) {
  // The server falls back to a saved view's filter (?view=) only when no
  // explicit tree (?f=) is sent. If the user has emptied the filter — e.g. by
  // removing the last chip — we must NOT pass the view slug, or the server
  // would re-hydrate the view's filter and the "clear" would appear to do
  // nothing. Only hand over the slug while an actual filter is present.
  const hasFilter = effectiveTree.value.children.length > 0;
  await listTasks(projectKey.value, page, 20, {
    tree: effectiveTree.value,
    sortBy: sortBy.value,
    sortDir: sortDir.value,
    viewSlug: hasFilter ? activeViewSlug.value ?? undefined : undefined,
  });
}

async function handleTaskCreated() {
  setPageInUrl(1);
  await loadTasks(1);
}

async function handleMemberAdded() {
  await listMembers(projectKey.value);
}

async function handleSettingsRefresh() {
  await Promise.all([
    getProject(projectKey.value),
    listStates(projectKey.value),
    listLabels(projectKey.value),
    listTemplates(projectKey.value),
  ]);
}

function prevPage() {
  if (tasksPage.value > 1) {
    clearSelection();
    setPageInUrl(tasksPage.value - 1);
  }
}
function nextPage() {
  if (tasksPage.value < tasksTotalPages.value) {
    clearSelection();
    setPageInUrl(tasksPage.value + 1);
  }
}

async function applyView(slug: string) {
  const res = await getView(projectKey.value, slug);
  if (!res.success || !res.data) return;
  const v: ProjectView = res.data;

  // Build all query params in one go to avoid race conditions from
  // multiple router.replace() calls overwriting each other.
  const q: Record<string, string | undefined> = { ...route.query };

  // View slug
  q.view = v.slug;

  // Filter tree
  if (v.filter_tree && v.filter_tree.children.length > 0) {
    q.f = encodeTree(v.filter_tree);
  } else {
    delete q.f;
  }

  // Sort
  q.sort_by = v.sort_by === "created_at" ? undefined : v.sort_by;
  q.sort_dir = v.sort_dir === "desc" ? undefined : v.sort_dir;

  // Group by
  q.group_by = v.group_by === "state" ? undefined : v.group_by;

  // Switch to the view's default tab
  const targetTab = v.default_tab || "tasks";
  q.tab = targetTab === "tasks" ? undefined : targetTab;

  // Reset page
  delete q.page;

  await router.replace({ query: q });

  // Sync local tree state after the route has updated
  setTree(v.filter_tree, { resetPage: false });
}

function openRenameView(view: ProjectView) {
  renameViewTarget.value = view;
  showSaveView.value = true;
}

function resetFilters() {
  clearAll();
}

function handleTreeUpdate(next: FilterTree) {
  // Removing the last chip means "no filter" — also drop the saved-view slug so
  // the server doesn't fall back to the view's filter (and it doesn't return on
  // the next visit). Any remaining chips keep the view association (drift).
  if (next.children.length === 0) {
    clearTreeAndView();
    return;
  }
  setTree(next);
}

// React to URL page changes (including browser back/forward)
watch(currentPageFromUrl, (newPage) => {
  if (!loading.value && newPage !== tasksPage.value) {
    loadTasks(newPage);
  }
});

// Reload tasks when effective filter, sort, or active view changes.
// Page reset is handled by the individual URL writers (setTree, searchQuery,
// sortBy/sortDir) in a single router.replace each — issuing another replace
// here would race and clobber the just-written ?f= (dropping the filter from
// the URL, so it would vanish on browser back).
watch(
  [effectiveTree, sortBy, sortDir, activeViewSlug],
  () => {
    if (!loading.value) {
      loadTasks(1);
    }
  },
  { deep: true }
);

const existingMemberIds = computed(() => members.value.map((m) => m.user_id));

onMounted(async () => {
  await hydrateFromUrl();
  loadProject();
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-8">
        <div v-if="loading" class="flex items-center justify-center py-20">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <div
          v-else-if="error"
          class="flex flex-col items-center justify-center py-20"
        >
          <p class="text-lg text-destructive">{{ error }}</p>
          <Button class="mt-4" variant="outline" @click="loadProject">
            Try Again
          </Button>
        </div>

        <template v-else-if="currentProject">
          <ProjectHeader
            :project="currentProject"
            :member-count="members.length"
          />

          <div
            v-if="isDisabled"
            class="mt-4 flex items-center gap-2 rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-400"
          >
            <Lock class="size-4 shrink-0" />
            <span>
              This project is disabled and read-only. No changes can be made until an admin re-enables it in Settings.
            </span>
          </div>

          <Tabs v-model="activeTab" class="mt-6">
            <div class="flex items-center justify-between gap-3">
              <TabsList>
                <TabsTrigger value="tasks" class="gap-2">
                  <ListTodo class="size-4" />
                  Tasks
                </TabsTrigger>
                <TabsTrigger value="board" class="gap-2">
                  <Kanban class="size-4" />
                  Board
                </TabsTrigger>
                <TabsTrigger value="pages" class="gap-2">
                  <FileText class="size-4" />
                  Pages
                </TabsTrigger>
                <TabsTrigger value="cycles" class="gap-2">
                  <Repeat class="size-4" />
                  Cycles
                </TabsTrigger>
                <TabsTrigger value="modules" class="gap-2">
                  <Layers class="size-4" />
                  Modules
                </TabsTrigger>
                <TabsTrigger value="members" class="gap-2">
                  <Users class="size-4" />
                  Members
                </TabsTrigger>
                <TabsTrigger value="views" class="gap-2">
                  <Eye class="size-4" />
                  Views
                  <span
                    v-if="views.length > 0"
                    class="ml-0.5 rounded-full bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground"
                  >
                    {{ views.length }}
                  </span>
                </TabsTrigger>
                <TabsTrigger v-if="isAdmin" value="settings" class="gap-2">
                  <Settings class="size-4" />
                  Settings
                </TabsTrigger>
              </TabsList>

              <Button
                v-if="activeTab === 'tasks' || activeTab === 'board'"
                variant="outline"
                size="sm"
                class="gap-1.5"
                @click="renameViewTarget = null; showSaveView = true"
              >
                <Save class="size-3.5" />
                {{ activeViewSlug ? "Save as view" : "Save view" }}
              </Button>
            </div>

            <!-- Shared filter bar for tasks + board tabs -->
            <div
              v-if="activeTab === 'tasks' || activeTab === 'board'"
              class="mt-6 space-y-3"
            >
              <div class="flex items-start gap-3">
                <FilterBar
                  class="flex-1"
                  :tree="tree"
                  :search-query="searchQuery"
                  :sort-by="sortBy"
                  :sort-dir="sortDir"
                  :group-by="groupBy"
                  :states="states"
                  :labels="labels"
                  :members="members"
                  :show-group-by="activeTab === 'board'"
                  @update:tree="handleTreeUpdate"
                  @update:search-query="(v) => (searchQuery = v)"
                  @update:sort-by="(v) => (sortBy = v)"
                  @update:sort-dir="(v) => (sortDir = v)"
                  @update:group-by="(v) => (groupBy = v)"
                  @reset="resetFilters"
                />
                <div v-if="canWrite" class="flex items-center">
                  <Button class="rounded-r-none" @click="showCreateTask = true">
                    <Plus class="mr-2 size-4" />
                    Create Task
                  </Button>
                  <Button
                    class="rounded-l-none border-l border-primary-foreground/20 px-2"
                    aria-label="Create task in full page"
                    as-child
                  >
                    <a :href="`/projects/${projectKey}/tasks/new`" target="_blank">
                      <ExternalLink class="size-4" />
                    </a>
                  </Button>
                </div>
              </div>
            </div>

            <!-- Tasks Tab -->
            <TabsContent value="tasks" class="mt-6 space-y-4">
              <div v-if="tasksLoading" class="flex items-center justify-center py-12">
                <Loader2 class="size-6 animate-spin text-muted-foreground" />
              </div>

              <div
                v-else-if="tasks.length === 0"
                class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
              >
                <ListTodo class="size-8 text-muted-foreground" />
                <h3 class="mt-4 font-semibold">No tasks match</h3>
                <p class="mt-1 text-sm text-muted-foreground">
                  Try clearing filters or create a new task.
                </p>
                <Button v-if="canWrite" class="mt-4" as-child>
                  <NuxtLink :to="`/projects/${projectKey}/tasks/new`">
                    <Plus class="mr-2 size-4" />
                    Create Task
                  </NuxtLink>
                </Button>
              </div>

              <template v-else>
                <div
                  v-if="canWrite"
                  class="flex items-center justify-between rounded-md border bg-muted/40 px-3 py-2"
                >
                  <div class="flex items-center gap-3">
                    <Button variant="outline" size="sm" @click="toggleSelectAll">
                      {{ allSelected ? "Deselect all" : "Select all" }}
                    </Button>
                    <span v-if="selectedTasks.size > 0" class="text-sm font-medium">
                      {{ selectedTasks.size }} selected
                    </span>
                  </div>
                  <div v-if="selectedTasks.size > 0" class="flex items-center gap-2">
                    <Button size="sm" @click="showBulkMove = true">
                      <FolderInput class="mr-2 size-4" />
                      Move
                    </Button>
                    <Button variant="ghost" size="sm" @click="clearSelection">
                      <X class="mr-1 size-4" />
                      Clear
                    </Button>
                  </div>
                </div>

                <TaskList
                  :tasks="tasks"
                  :project-key="projectKey"
                  :states="states"
                  :is-member="canWrite"
                  :selectable="canWrite"
                  :selected="selectedTasks"
                  @updated="() => loadTasks(tasksPage)"
                  @toggle-select="toggleTaskSelection"
                />

                <div
                  v-if="tasksTotalPages > 1"
                  class="flex items-center justify-between border-t pt-4"
                >
                  <p class="text-sm text-muted-foreground">
                    Showing {{ tasks.length }} of {{ totalTasks }} tasks
                  </p>
                  <div class="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      aria-label="Previous page"
                      :disabled="tasksPage === 1"
                      @click="prevPage"
                    >
                      <ChevronLeft class="size-4" />
                    </Button>
                    <span class="text-sm">
                      Page {{ tasksPage }} of {{ tasksTotalPages }}
                    </span>
                    <Button
                      variant="outline"
                      size="sm"
                      aria-label="Next page"
                      :disabled="tasksPage >= tasksTotalPages"
                      @click="nextPage"
                    >
                      <ChevronRight class="size-4" />
                    </Button>
                  </div>
                </div>
              </template>
            </TabsContent>

            <!-- Board Tab -->
            <TabsContent value="board" class="-mx-32 mt-6 px-2">
              <KanbanBoard
                :tasks="tasks"
                :states="states"
                :members="members"
                :labels="labels"
                :project-key="projectKey"
                :is-member="canWrite"
                :group-by="groupBy"
                :current-user-id="currentUserId"
                @refresh="() => loadTasks(tasksPage)"
              />
            </TabsContent>

            <!-- Pages Tab -->
            <TabsContent value="pages" class="mt-6">
              <ProjectPagesTab :project-key="projectKey" :can-write="canWrite" />
            </TabsContent>

            <!-- Cycles Tab -->
            <TabsContent value="cycles" class="mt-6">
              <ProjectCyclesTab :project-key="projectKey" :is-admin="isAdmin" />
            </TabsContent>

            <!-- Modules Tab -->
            <TabsContent value="modules" class="mt-6">
              <ProjectModulesTab :project-key="projectKey" :is-admin="isAdmin" />
            </TabsContent>

            <!-- Members Tab -->
            <TabsContent value="members" class="mt-6 space-y-4">
              <div class="flex items-center justify-between">
                <div>
                  <h2 class="text-lg font-semibold">Project Members</h2>
                  <p class="text-sm text-muted-foreground">
                    {{ members.length }} member{{ members.length !== 1 ? "s" : "" }}
                  </p>
                </div>
                <Button v-if="isAdmin" @click="showAddMember = true">
                  <Plus class="mr-2 size-4" />
                  Add Member
                </Button>
              </div>

              <Card>
                <CardContent class="p-0">
                  <MemberList
                    :members="members"
                    :project-key="projectKey"
                    :current-user-role="currentProject.role"
                    @refresh="listMembers(projectKey)"
                  />
                </CardContent>
              </Card>
            </TabsContent>

            <!-- Views Tab -->
            <TabsContent value="views" class="mt-6 space-y-4">
              <div class="flex items-center justify-between">
                <div>
                  <h2 class="text-lg font-semibold">Saved Views</h2>
                  <p class="text-sm text-muted-foreground">
                    Filter combinations you or the team return to often.
                  </p>
                </div>
                <Button
                  v-if="isMember"
                  variant="outline"
                  @click="renameViewTarget = null; showSaveView = true"
                >
                  <Plus class="mr-2 size-4" />
                  Save current filters
                </Button>
              </div>

              <ViewsList
                :project-key="projectKey"
                :views="views"
                :active-slug="activeViewSlug"
                :current-user-id="currentUserId"
                :is-admin="isAdmin"
                @apply:view="applyView"
                @rename:view="openRenameView"
                @refresh="listViews(projectKey)"
              />
            </TabsContent>

            <!-- Settings Tab -->
            <TabsContent v-if="isAdmin" value="settings" class="mt-6 space-y-8">
              <ProjectSettings
                :project="currentProject"
                :is-admin="isAdmin"
                @refresh="handleSettingsRefresh"
              />

              <Separator />

              <StatesManager
                :states="states"
                :project-key="projectKey"
                :is-admin="isAdmin"
                @refresh="listStates(projectKey)"
              />

              <Separator />

              <LabelsManager
                :labels="labels"
                :project-key="projectKey"
                :is-admin="isAdmin"
                @refresh="listLabels(projectKey)"
              />

              <Separator />

              <TemplatesManager
                :templates="templates"
                :project-key="projectKey"
                :is-admin="isAdmin"
                @refresh="listTemplates(projectKey)"
              />

              <Separator />

              <ProjectDangerZone :project="currentProject" />
            </TabsContent>
          </Tabs>
        </template>

        <CreateTaskDialog
          v-model:open="showCreateTask"
          project-selector
          all-workspaces
          :initial-project-key="projectKey"
          @created="handleTaskCreated"
        />

        <MoveTaskDialog
          v-model:open="showBulkMove"
          :project-key="projectKey"
          :task-numbers="Array.from(selectedTasks)"
          @moved="handleBulkMoved"
        />

        <AddMemberDialog
          v-model:open="showAddMember"
          :project-key="projectKey"
          :existing-member-ids="existingMemberIds"
          @added="handleMemberAdded"
        />

        <SaveViewDialog
          :open="showSaveView"
          :project-key="projectKey"
          :initial="renameViewTarget ? {
            slug: renameViewTarget.slug,
            name: renameViewTarget.name,
            description: renameViewTarget.description,
            visibility: renameViewTarget.visibility,
            default_tab: renameViewTarget.default_tab,
          } : undefined"
          :current-tree="tree"
          :current-group-by="groupBy"
          :current-sort-by="sortBy"
          :current-sort-dir="sortDir"
          @update:open="(v) => { showSaveView = v; if (!v) renameViewTarget = null; }"
          @saved="(slug) => { listViews(projectKey); if (!renameViewTarget) applyView(slug); }"
        />
      </div>
    </main>
  </div>
</template>

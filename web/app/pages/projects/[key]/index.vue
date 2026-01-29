<script setup lang="ts">
import {
  ListTodo,
  Kanban,
  Users,
  Settings,
  Plus,
  Loader2,
  ChevronLeft,
  ChevronRight,
} from "lucide-vue-next";
import type { TaskFilters } from "~/types";

definePageMeta({
  middleware: ["auth"],
});

const route = useRoute();
const projectKey = computed(() => route.params.key as string);

const {
  currentProject,
  members,
  states,
  labels,
  getProject,
  listMembers,
  listStates,
  listLabels,
} = useProjects();

const {
  tasks,
  loading: tasksLoading,
  total: totalTasks,
  page: tasksPage,
  totalPages: tasksTotalPages,
  listTasks,
} = useTasks();

const loading = ref(true);
const error = ref<string | null>(null);
const activeTab = ref("tasks");
const taskFilters = ref<TaskFilters>({});
const showCreateTask = ref(false);
const showAddMember = ref(false);

const isAdmin = computed(() => currentProject.value?.role === "admin");
const isMember = computed(
  () => currentProject.value?.role === "admin" || currentProject.value?.role === "member"
);

async function loadProject() {
  loading.value = true;
  error.value = null;

  const result = await getProject(projectKey.value);
  if (!result.success) {
    error.value = result.error || "Failed to load project";
    loading.value = false;
    return;
  }

  // Load project data in parallel
  await Promise.all([
    listMembers(projectKey.value),
    listStates(projectKey.value),
    listLabels(projectKey.value),
    loadTasks(),
  ]);

  loading.value = false;
}

async function loadTasks(page = 1) {
  await listTasks(projectKey.value, page, 20, taskFilters.value);
}

async function handleFilterChange(filters: TaskFilters) {
  taskFilters.value = filters;
  await loadTasks(1);
}

async function handleTaskCreated() {
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
  ]);
}

function prevPage() {
  if (tasksPage.value > 1) {
    loadTasks(tasksPage.value - 1);
  }
}

function nextPage() {
  if (tasksPage.value < tasksTotalPages.value) {
    loadTasks(tasksPage.value + 1);
  }
}

const existingMemberIds = computed(() => members.value.map((m) => m.user_id));

onMounted(() => {
  loadProject();
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-8">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-20">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <!-- Error -->
        <div
          v-else-if="error"
          class="flex flex-col items-center justify-center py-20"
        >
          <p class="text-lg text-destructive">{{ error }}</p>
          <Button class="mt-4" variant="outline" @click="loadProject">
            Try Again
          </Button>
        </div>

        <!-- Project content -->
        <template v-else-if="currentProject">
          <!-- Header -->
          <ProjectHeader
            :project="currentProject"
            :member-count="members.length"
          />

          <!-- Tabs -->
          <Tabs v-model="activeTab" class="mt-8">
            <TabsList>
              <TabsTrigger value="tasks" class="gap-2">
                <ListTodo class="size-4" />
                Tasks
              </TabsTrigger>
              <TabsTrigger value="board" class="gap-2">
                <Kanban class="size-4" />
                Board
              </TabsTrigger>
              <TabsTrigger value="members" class="gap-2">
                <Users class="size-4" />
                Members
              </TabsTrigger>
              <TabsTrigger v-if="isAdmin" value="settings" class="gap-2">
                <Settings class="size-4" />
                Settings
              </TabsTrigger>
            </TabsList>

            <!-- Tasks Tab -->
            <TabsContent value="tasks" class="mt-6 space-y-4">
              <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <TaskFilters
                  :states="states"
                  :filters="taskFilters"
                  @update:filters="handleFilterChange"
                />
                <Button v-if="isMember" @click="showCreateTask = true">
                  <Plus class="mr-2 size-4" />
                  Create Task
                </Button>
              </div>

              <!-- Tasks loading -->
              <div v-if="tasksLoading" class="flex items-center justify-center py-12">
                <Loader2 class="size-6 animate-spin text-muted-foreground" />
              </div>

              <!-- Tasks empty -->
              <div
                v-else-if="tasks.length === 0"
                class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
              >
                <ListTodo class="size-8 text-muted-foreground" />
                <h3 class="mt-4 font-semibold">No tasks yet</h3>
                <p class="mt-1 text-sm text-muted-foreground">
                  Create your first task to get started
                </p>
                <Button v-if="isMember" class="mt-4" @click="showCreateTask = true">
                  <Plus class="mr-2 size-4" />
                  Create Task
                </Button>
              </div>

              <!-- Tasks list -->
              <template v-else>
                <div class="space-y-2">
                  <TaskCard
                    v-for="task in tasks"
                    :key="task.id"
                    :task="task"
                    :project-key="projectKey"
                  />
                </div>

                <!-- Pagination -->
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
            <TabsContent value="board" class="mt-6">
              <KanbanBoard
                :tasks="tasks"
                :states="states"
                :project-key="projectKey"
                :is-member="isMember"
                @refresh="loadTasks"
              />
            </TabsContent>

            <!-- Members Tab -->
            <TabsContent value="members" class="mt-6 space-y-4">
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-lg font-semibold">Project Members</h3>
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
            </TabsContent>
          </Tabs>
        </template>

        <!-- Create Task Dialog -->
        <CreateTaskDialog
          v-model:open="showCreateTask"
          :project-key="projectKey"
          :states="states"
          :labels="labels"
          :members="members"
          @created="handleTaskCreated"
        />

        <!-- Add Member Dialog -->
        <AddMemberDialog
          v-model:open="showAddMember"
          :project-key="projectKey"
          :existing-member-ids="existingMemberIds"
          @added="handleMemberAdded"
        />
      </div>
    </main>
  </div>
</template>

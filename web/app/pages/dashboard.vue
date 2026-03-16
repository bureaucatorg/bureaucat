<script setup lang="ts">
import {
  FolderKanban,
  ListTodo,
  Plus,
  ArrowRight,
  Loader2,
  Search,
  Circle,
  CircleDot,
  CheckCircle2,
  XCircle,
  Clock,
} from "lucide-vue-next";
import { PRIORITY_LABELS } from "~/types";

definePageMeta({
  middleware: ["auth"],
});

useSeoMeta({ title: "Dashboard" });

const { user, getAuthHeader } = useAuth();
const { projects, loading: projectsLoading, listProjects } = useProjects();

const showCreateDialog = ref(false);
const searchQuery = ref("");
let debounceTimer: ReturnType<typeof setTimeout> | null = null;

function fetchProjects() {
  listProjects(1, 100, searchQuery.value);
}

watch(searchQuery, () => {
  if (debounceTimer) clearTimeout(debounceTimer);
  debounceTimer = setTimeout(fetchProjects, 300);
});

async function handleCreated() {
  fetchProjects();
}

// My Tasks
interface MyTask {
  id: string;
  project_key: string;
  task_number: number;
  task_id: string;
  title: string;
  state_name: string;
  state_type: string;
  state_color: string;
  priority: number;
}

interface MyTasksResponse {
  tasks: MyTask[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

const myTasks = ref<MyTask[]>([]);
const myTasksTotal = ref(0);
const myTasksLoading = ref(false);

async function fetchMyTasks() {
  myTasksLoading.value = true;
  try {
    const response = await fetch("/api/v1/me/tasks?per_page=20", {
      headers: getAuthHeader(),
    });
    if (response.ok) {
      const data: MyTasksResponse = await response.json();
      myTasks.value = data.tasks || [];
      myTasksTotal.value = data.total;
    }
  } catch {
    // silently fail
  } finally {
    myTasksLoading.value = false;
  }
}

function getStateIcon(stateType: string) {
  switch (stateType) {
    case "backlog": return Clock;
    case "unstarted": return Circle;
    case "started": return CircleDot;
    case "completed": return CheckCircle2;
    case "cancelled": return XCircle;
    default: return Circle;
  }
}

onMounted(() => {
  fetchProjects();
  fetchMyTasks();
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-12">
        <!-- Welcome Section -->
        <div class="mb-8">
          <h1 class="text-3xl font-bold tracking-tight">
            Welcome back, {{ user?.first_name }}!
          </h1>
          <p class="mt-2 text-muted-foreground">
            Here's an overview of your projects and tasks
          </p>
        </div>

        <!-- Assigned to You -->
        <div class="mb-10">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="flex items-center gap-2 text-lg font-semibold">
              <ListTodo class="size-5" />
              Assigned to You
              <span v-if="myTasksTotal > 0" class="text-sm font-normal text-muted-foreground">
                ({{ myTasksTotal }})
              </span>
            </h2>
          </div>

          <div v-if="myTasksLoading" class="flex items-center justify-center py-8">
            <Loader2 class="size-6 animate-spin text-muted-foreground" />
          </div>

          <div
            v-else-if="myTasks.length === 0"
            class="rounded-lg border border-dashed py-8 text-center text-sm text-muted-foreground"
          >
            No tasks assigned to you
          </div>

          <div v-else class="space-y-1.5">
            <NuxtLink
              v-for="task in myTasks"
              :key="task.id"
              :to="`/projects/${task.project_key}/tasks/${task.task_number}`"
              class="flex items-center gap-3 rounded-lg border px-4 py-3 transition-colors hover:bg-muted/50"
            >
              <component
                :is="getStateIcon(task.state_type)"
                class="size-4 shrink-0"
                :style="{ color: task.state_color || undefined }"
              />
              <span class="min-w-0 flex-1 truncate text-sm">{{ task.title }}</span>
              <span
                v-if="task.priority > 0"
                class="size-2 shrink-0 rounded-full"
                :style="{ backgroundColor: PRIORITY_LABELS[task.priority]?.color }"
              />
              <span class="shrink-0 text-xs font-medium text-muted-foreground">
                {{ task.task_id }}
              </span>
            </NuxtLink>
          </div>
        </div>

        <!-- Your Projects Section -->
        <div>
          <div class="mb-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <h2 class="text-lg font-semibold">Your Projects</h2>
            <div class="flex items-center gap-3">
              <div class="relative">
                <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  v-model="searchQuery"
                  placeholder="Search projects..."
                  class="h-9 w-56 pl-9"
                />
              </div>
              <Button variant="ghost" size="sm" as-child>
                <NuxtLink to="/projects" class="flex items-center gap-1">
                  View all
                  <ArrowRight class="size-4" />
                </NuxtLink>
              </Button>
              <Button size="sm" @click="showCreateDialog = true">
                <Plus class="mr-1.5 size-4" />
                Create Project
              </Button>
            </div>
          </div>

          <!-- Loading -->
          <div v-if="projectsLoading" class="flex items-center justify-center py-12">
            <Loader2 class="size-8 animate-spin text-muted-foreground" />
          </div>

          <!-- Empty state -->
          <Card
            v-else-if="projects.length === 0 && !searchQuery"
            class="flex flex-col items-center justify-center border-dashed py-12"
          >
            <div class="flex size-14 items-center justify-center rounded-full bg-muted">
              <FolderKanban class="size-7 text-muted-foreground" />
            </div>
            <h3 class="mt-4 font-semibold">No projects yet</h3>
            <p class="mt-1 text-sm text-muted-foreground">
              Create your first project to get started
            </p>
            <Button class="mt-4" size="sm" @click="showCreateDialog = true">
              <Plus class="mr-1.5 size-4" />
              Create Project
            </Button>
          </Card>

          <!-- No search results -->
          <div
            v-else-if="projects.length === 0 && searchQuery"
            class="flex flex-col items-center justify-center rounded-lg border border-dashed py-12"
          >
            <Search class="size-8 text-muted-foreground" />
            <h3 class="mt-4 font-semibold">No projects found</h3>
            <p class="mt-1 text-sm text-muted-foreground">
              Try a different search term
            </p>
          </div>

          <!-- Projects grid -->
          <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <ProjectCard
              v-for="project in projects"
              :key="project.id"
              :project="project"
            />
          </div>
        </div>

        <CreateProjectDialog
          v-model:open="showCreateDialog"
          @created="handleCreated"
        />
      </div>
    </main>
  </div>
</template>

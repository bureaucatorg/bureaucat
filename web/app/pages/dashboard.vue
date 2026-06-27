<script setup lang="ts">
import {
  FolderKanban,
  ListTodo,
  Plus,
  ArrowRight,
  Loader2,
  Search,
  Lightbulb,
} from "lucide-vue-next";
import type { Project, Task, ProjectState } from "~/types";

definePageMeta({
  middleware: ["auth"],
});

useSeoMeta({ title: "Dashboard" });

const { user, getAuthHeader } = useAuth();
const { projects, loading: projectsLoading, listProjects } = useProjects();

const showCreateDialog = ref(false);
const showCreateTask = ref(false);
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
  fetchAllProjects();
}

// Full (unfiltered) project list for the Create Task dialog's selector, kept
// independent of the project search box above.
const allProjects = ref<Project[]>([]);

async function fetchAllProjects() {
  // Direct fetch (not the shared useProjects store) so it doesn't clobber the
  // search-filtered grid above.
  try {
    const response = await fetch("/api/v1/projects?page=1&per_page=100", {
      headers: getAuthHeader(),
    });
    if (response.ok) {
      const data = await response.json();
      allProjects.value = data.projects || [];
    }
  } catch {
    // silently fail — the dialog just shows no projects to pick
  }
}

async function handleTaskCreated() {
  fetchMyTasks();
}

// My Tasks
interface MyTaskAssignee {
  id: string;
  user_id: string;
  username: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
}

interface MyTask {
  id: string;
  project_key: string;
  task_number: number;
  task_id: string;
  title: string;
  state_id: string;
  state_name: string;
  state_type: string;
  state_color: string;
  priority: number;
  assignees: MyTaskAssignee[];
  comment_count: number;
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

// Adapt the dashboard's lightweight MyTask shape to the full Task shape so we
// can render the shared TaskList/TaskCard component. Fields the dashboard API
// doesn't return (creator, timestamps) are left empty — TaskCard treats them
// as absent (assignees-only avatars).
const myTasksAsTask = computed<Task[]>(() =>
  myTasks.value.map((t) => ({
    id: t.id,
    project_key: t.project_key,
    task_number: t.task_number,
    task_id: t.task_id,
    title: t.title,
    state_id: t.state_id,
    state_name: t.state_name,
    state_type: t.state_type,
    state_color: t.state_color,
    priority: t.priority,
    created_by: "",
    creator_username: "",
    creator_first_name: "",
    creator_last_name: "",
    assignees: t.assignees.map((a) => ({
      id: a.id,
      user_id: a.user_id,
      username: a.username,
      email: "",
      first_name: a.first_name,
      last_name: a.last_name,
      avatar_url: a.avatar_url,
    })),
    comment_count: t.comment_count,
    created_at: "",
    updated_at: "",
  }))
);

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
      fetchStatesForMyTasks();
    }
  } catch {
    // silently fail
  } finally {
    myTasksLoading.value = false;
  }
}

// Per-project state lists, loaded lazily for the projects that own the user's
// tasks. TaskCard needs these to offer the inline state selector.
const statesByProject = ref<Record<string, ProjectState[]>>({});

// Whether the user can edit a project's tasks (admins and members can), keyed
// by project_key. Derived from the full project list we already fetch.
const isMemberByProject = computed<Record<string, boolean>>(() => {
  const map: Record<string, boolean> = {};
  for (const p of allProjects.value) {
    map[p.project_key] = p.role === "admin" || p.role === "member";
  }
  return map;
});

async function fetchStatesForMyTasks() {
  const keys = [...new Set(myTasks.value.map((t) => t.project_key))];
  const missing = keys.filter((k) => !statesByProject.value[k]);
  await Promise.all(
    missing.map(async (key) => {
      try {
        const res = await fetch(`/api/v1/projects/${key}/states`, {
          headers: getAuthHeader(),
        });
        if (res.ok) {
          statesByProject.value[key] = await res.json();
        }
      } catch {
        // ignore — that project's rows simply stay read-only
      }
    })
  );
}

// Tips
const { ssoProviders, fetchSSOProviders } = useSettings();

const tips: { id: string; show: () => boolean }[] = [
  {
    id: "avatar-sso",
    show: () => !user.value?.avatar_url,
  },
];

const currentTip = ref<string | null>(null);

onMounted(async () => {
  fetchProjects();
  fetchAllProjects();
  fetchMyTasks();
  await fetchSSOProviders();

  const applicable = tips.filter((t) => t.show());
  if (applicable.length > 0) {
    currentTip.value = applicable[Math.floor(Math.random() * applicable.length)].id;
  }
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex-1">
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

        <!-- Tip -->
        <div
          v-if="currentTip"
          class="mb-8 flex items-center gap-3 rounded-lg bg-amber-500/10 px-4 py-3 text-sm text-amber-700 dark:text-amber-400"
        >
          <Lightbulb class="size-4 shrink-0" />
          <span v-if="currentTip === 'avatar-sso'">
            <span class="font-semibold">Tip:</span> You can set your profile picture on your SSO provider
            (e.g. <a
              v-if="ssoProviders.zitadel && ssoProviders.zitadel_url"
              :href="`${ssoProviders.zitadel_url}/ui/console/users/me?id=general`"
              target="_blank"
              rel="noopener noreferrer"
              class="underline underline-offset-2 hover:text-amber-900 dark:hover:text-amber-300"
            >Zitadel</a><span v-else>Zitadel</span>)
            to make sure it shows up on your avatar across Bureaucat.
          </span>
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
            <Button size="sm" :disabled="allProjects.length === 0" @click="showCreateTask = true">
              <Plus class="mr-1.5 size-4" />
              Create Task
            </Button>
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

          <TaskList
            v-else
            :tasks="myTasksAsTask"
            :states-by-project="statesByProject"
            :is-member-by-project="isMemberByProject"
            @updated="fetchMyTasks"
          />
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

        <CreateTaskDialog
          v-model:open="showCreateTask"
          :projects="allProjects"
          @created="handleTaskCreated"
        />
      </div>
    </main>
  </div>
</template>

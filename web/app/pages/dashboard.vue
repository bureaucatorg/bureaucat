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
  MessageSquare,
  Lightbulb,
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

// Tips
const tips: { text: string; show: () => boolean }[] = [
  {
    text: "You can set your profile picture on your SSO provider (e.g. Zitadel) to make sure it shows up on your avatar across Bureaucat.",
    show: () => !user.value?.avatar_url,
  },
];

const currentTip = ref<string | null>(null);

onMounted(() => {
  fetchProjects();
  fetchMyTasks();

  const applicable = tips.filter((t) => t.show());
  if (applicable.length > 0) {
    currentTip.value = applicable[Math.floor(Math.random() * applicable.length)].text;
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
          <span><span class="font-semibold">Tip:</span> {{ currentTip }}</span>
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
            >
              <div class="dashboard-task-row group grid items-center rounded-lg border border-border/50 bg-background/50 px-3 py-2.5 transition-all hover:border-amber-500/30 hover:bg-muted/50">
                <span class="font-mono text-sm text-muted-foreground truncate">{{ task.task_id }}</span>
                <span class="truncate text-sm font-medium min-w-0">{{ task.title }}</span>
                <div class="flex items-center gap-1 rounded-md border bg-muted/50 px-1.5 py-0.5 w-fit justify-self-end">
                  <component
                    :is="getStateIcon(task.state_type)"
                    class="size-3.5 shrink-0 stroke-[2.5]"
                    :style="{ color: task.state_color || undefined }"
                  />
                  <span class="text-xs text-muted-foreground whitespace-nowrap">{{ task.state_name }}</span>
                </div>
                <div class="flex items-center gap-1 rounded-md border bg-muted/50 px-1.5 py-0.5 w-fit justify-self-end">
                  <span
                    class="size-2.5 shrink-0 rounded-full ring-1.5 ring-offset-1 ring-offset-background"
                    :style="{ backgroundColor: PRIORITY_LABELS[task.priority]?.color, '--tw-ring-color': PRIORITY_LABELS[task.priority]?.color }"
                  />
                  <span class="text-xs text-muted-foreground whitespace-nowrap">{{ PRIORITY_LABELS[task.priority]?.label }}</span>
                </div>
                <div class="flex items-center justify-end">
                  <div v-if="task.assignees?.length" class="flex -space-x-1.5">
                    <NuxtLink
                      v-for="person in task.assignees.slice(0, 4)"
                      :key="person.user_id"
                      :to="`/profile/${person.user_id}`"
                      :title="`${person.first_name} ${person.last_name}`"
                      class="hover:z-10"
                      @click.stop
                    >
                      <Avatar class="size-6 border-2 border-background transition-transform hover:scale-110">
                        <AvatarImage
                          v-if="person.avatar_url"
                          :src="person.avatar_url"
                          :alt="`${person.first_name} ${person.last_name}`"
                        />
                        <AvatarFallback class="text-[10px]">
                          {{ person.first_name?.[0] || "" }}{{ person.last_name?.[0] || "" }}
                        </AvatarFallback>
                      </Avatar>
                    </NuxtLink>
                    <Avatar
                      v-if="task.assignees.length > 4"
                      class="size-6 border-2 border-background"
                      :title="`${task.assignees.length - 4} more`"
                    >
                      <AvatarFallback class="text-[10px] bg-muted">
                        +{{ task.assignees.length - 4 }}
                      </AvatarFallback>
                    </Avatar>
                  </div>
                </div>
                <div class="flex items-center justify-end">
                  <div
                    class="flex items-center gap-1 rounded-full bg-muted px-1.5 py-0.5"
                    :title="`${task.comment_count} comment${task.comment_count !== 1 ? 's' : ''}`"
                  >
                    <MessageSquare class="size-3 text-muted-foreground" />
                    <span class="font-mono text-xs font-medium text-muted-foreground">{{ task.comment_count }}</span>
                  </div>
                </div>
              </div>
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

<style scoped>
.dashboard-task-row {
  grid-template-columns: 6rem 1fr 10rem 7rem 6rem 3rem;
  column-gap: 0.375rem;
}
</style>

<script setup lang="ts">
import {
  FolderKanban,
  Plus,
  ArrowRight,
  Loader2,
  Search,
} from "lucide-vue-next";

definePageMeta({
  middleware: ["auth"],
});

useSeoMeta({ title: "Dashboard" });

const { user } = useAuth();
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

onMounted(() => {
  fetchProjects();
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

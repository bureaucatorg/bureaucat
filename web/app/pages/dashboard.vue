<script setup lang="ts">
import {
  FolderKanban,
  ListTodo,
  CheckCircle2,
  Clock,
  Plus,
  ArrowRight,
  Loader2,
} from "lucide-vue-next";

definePageMeta({
  middleware: ["auth"],
});

useSeoMeta({ title: "Dashboard" });

const { user } = useAuth();
const { projects, loading: projectsLoading, listProjects, total: totalProjects } = useProjects();

const showCreateDialog = ref(false);

// Recent projects (max 6)
const recentProjects = computed(() => {
  return projects.value.slice(0, 6);
});

async function handleCreated() {
  await listProjects();
}

onMounted(() => {
  listProjects();
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
        <div class="mb-8">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="text-lg font-semibold">Your Projects</h2>
            <div class="flex items-center gap-2">
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
            v-else-if="projects.length === 0"
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

          <!-- Projects grid -->
          <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <ProjectCard
              v-for="project in recentProjects"
              :key="project.id"
              :project="project"
            />
          </div>
        </div>

        <!-- Quick Actions -->
        <div>
          <h2 class="mb-4 text-lg font-semibold">Quick Actions</h2>
          <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <Card
              class="group cursor-pointer border-border/50 bg-background/50 transition-all hover:border-amber-500/30 hover:shadow-lg hover:shadow-amber-500/5"
              @click="showCreateDialog = true"
            >
              <CardContent class="flex items-center gap-3 p-4">
                <div
                  class="flex size-10 items-center justify-center rounded-lg bg-muted transition-colors group-hover:bg-amber-500/10"
                >
                  <Plus
                    class="size-5 text-muted-foreground transition-colors group-hover:text-amber-600 dark:group-hover:text-amber-500"
                  />
                </div>
                <div>
                  <p class="font-medium">New Project</p>
                  <p class="text-xs text-muted-foreground">Create a workspace</p>
                </div>
              </CardContent>
            </Card>

            <NuxtLink to="/projects">
              <Card
                class="group h-full cursor-pointer border-border/50 bg-background/50 transition-all hover:border-amber-500/30 hover:shadow-lg hover:shadow-amber-500/5"
              >
                <CardContent class="flex items-center gap-3 p-4">
                  <div
                    class="flex size-10 items-center justify-center rounded-lg bg-muted transition-colors group-hover:bg-amber-500/10"
                  >
                    <FolderKanban
                      class="size-5 text-muted-foreground transition-colors group-hover:text-amber-600 dark:group-hover:text-amber-500"
                    />
                  </div>
                  <div>
                    <p class="font-medium">Browse Projects</p>
                    <p class="text-xs text-muted-foreground">View all projects</p>
                  </div>
                </CardContent>
              </Card>
            </NuxtLink>
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

<script setup lang="ts">
import { Plus, FolderKanban, Search, Loader2 } from "lucide-vue-next";

definePageMeta({
  middleware: ["auth"],
});

const { projects, loading, listProjects, total } = useProjects();

const showCreateDialog = ref(false);
const searchQuery = ref("");

const filteredProjects = computed(() => {
  if (!searchQuery.value) return projects.value;
  const query = searchQuery.value.toLowerCase();
  return projects.value.filter(
    (p) =>
      p.name.toLowerCase().includes(query) ||
      p.project_key.toLowerCase().includes(query) ||
      p.description?.toLowerCase().includes(query)
  );
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
        <div class="mb-8 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 class="flex items-center gap-3 text-3xl font-bold tracking-tight">
              <FolderKanban class="size-8" />
              Projects
            </h1>
            <p class="mt-2 text-muted-foreground">
              Manage your projects and track approvals
            </p>
          </div>
          <Button @click="showCreateDialog = true">
            <Plus class="mr-2 size-4" />
            Create Project
          </Button>
        </div>

        <!-- Search -->
        <div class="mb-6">
          <div class="relative max-w-sm">
            <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              v-model="searchQuery"
              placeholder="Search projects..."
              class="pl-9"
            />
          </div>
        </div>

        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-12">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <!-- Empty state -->
        <div
          v-else-if="projects.length === 0"
          class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
        >
          <div class="flex size-16 items-center justify-center rounded-full bg-muted">
            <FolderKanban class="size-8 text-muted-foreground" />
          </div>
          <h3 class="mt-4 text-lg font-semibold">No projects yet</h3>
          <p class="mt-2 text-sm text-muted-foreground">
            Create your first project to get started
          </p>
          <Button class="mt-4" @click="showCreateDialog = true">
            <Plus class="mr-2 size-4" />
            Create Project
          </Button>
        </div>

        <!-- Projects grid -->
        <div
          v-else-if="filteredProjects.length > 0"
          class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3"
        >
          <ProjectCard
            v-for="project in filteredProjects"
            :key="project.id"
            :project="project"
          />
        </div>

        <!-- No search results -->
        <div
          v-else
          class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
        >
          <Search class="size-8 text-muted-foreground" />
          <h3 class="mt-4 text-lg font-semibold">No projects found</h3>
          <p class="mt-2 text-sm text-muted-foreground">
            Try a different search term
          </p>
        </div>

        <CreateProjectDialog
          v-model:open="showCreateDialog"
          @created="handleCreated"
        />
      </div>
    </main>
  </div>
</template>

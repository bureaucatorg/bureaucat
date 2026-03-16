<script setup lang="ts">
import { Plus, FolderKanban, Search, Loader2, ChevronLeft, ChevronRight, Link, Check } from "lucide-vue-next";
import { toast } from "vue-sonner";

definePageMeta({
  middleware: ["auth"],
});

useSeoMeta({ title: "Projects" });

const copied = ref(false);
function copyLink() {
  navigator.clipboard.writeText(window.location.href);
  copied.value = true;
  toast.success("Link copied");
  setTimeout(() => { copied.value = false; }, 2000);
}

const { projects, loading, listProjects, total, page, totalPages } = useProjects();

const showCreateDialog = ref(false);
const searchQuery = ref("");
const perPage = 12;

let debounceTimer: ReturnType<typeof setTimeout> | null = null;

function fetchProjects(p = 1) {
  listProjects(p, perPage, searchQuery.value);
}

watch(searchQuery, () => {
  if (debounceTimer) clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    fetchProjects(1);
  }, 300);
});

async function handleCreated() {
  fetchProjects(1);
}

function goToPage(p: number) {
  if (p < 1 || p > totalPages.value) return;
  fetchProjects(p);
}

onMounted(() => {
  fetchProjects(1);
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-8">
        <nav class="mb-4 flex items-center gap-2 text-sm text-muted-foreground">
          <ChevronLeft class="size-4" />
          <span class="font-semibold text-amber-600 dark:text-amber-500">Projects</span>
          <button
            class="ml-1 rounded-md p-1 text-muted-foreground/50 hover:text-muted-foreground"
            @click="copyLink"
          >
            <Check v-if="copied" class="size-3.5 text-emerald-500" />
            <Link v-else class="size-3.5" />
          </button>
        </nav>

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

        <!-- Empty state (no projects at all, no search) -->
        <div
          v-else-if="projects.length === 0 && !searchQuery"
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

        <!-- No search results -->
        <div
          v-else-if="projects.length === 0 && searchQuery"
          class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
        >
          <Search class="size-8 text-muted-foreground" />
          <h3 class="mt-4 text-lg font-semibold">No projects found</h3>
          <p class="mt-2 text-sm text-muted-foreground">
            Try a different search term
          </p>
        </div>

        <!-- Projects grid -->
        <template v-else>
          <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <ProjectCard
              v-for="project in projects"
              :key="project.id"
              :project="project"
            />
          </div>

          <!-- Pagination -->
          <div
            v-if="totalPages > 1"
            class="mt-8 flex items-center justify-between"
          >
            <p class="text-sm text-muted-foreground">
              {{ total }} project{{ total === 1 ? '' : 's' }}
            </p>
            <div class="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                :disabled="page <= 1"
                @click="goToPage(page - 1)"
              >
                <ChevronLeft class="mr-1 size-4" />
                Prev
              </Button>
              <span class="text-sm text-muted-foreground">
                Page {{ page }} of {{ totalPages }}
              </span>
              <Button
                variant="outline"
                size="sm"
                :disabled="page >= totalPages"
                @click="goToPage(page + 1)"
              >
                Next
                <ChevronRight class="ml-1 size-4" />
              </Button>
            </div>
          </div>
        </template>

        <CreateProjectDialog
          v-model:open="showCreateDialog"
          @created="handleCreated"
        />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { Plus, Layers, Loader2, ChevronLeft, ChevronRight } from "lucide-vue-next";
import type { ModuleListFilters } from "~/types";

const props = defineProps<{
  projectKey: string;
  isAdmin: boolean;
}>();

const { modules, loading, total, page, totalPages, listModules } = useModules();

const showCreate = ref(false);
const perPage = 12;
const filters = ref<ModuleListFilters>({ sort_by: "created_at", sort_dir: "desc" });

function fetchPage(p = 1) {
  listModules(props.projectKey, p, perPage, filters.value);
}

function goToPage(p: number) {
  if (p < 1 || p > totalPages.value) return;
  fetchPage(p);
}

function onSaved() {
  fetchPage(1);
}

onMounted(() => fetchPage(1));

watch(
  () => props.projectKey,
  () => fetchPage(1)
);

watch(
  filters,
  () => fetchPage(1),
  { deep: true }
);
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-lg font-semibold">Modules</h2>
        <p class="text-sm text-muted-foreground">
          Reusable sub-projects. Group tasks, assign a lead, and duplicate
          for repeating workflows.
        </p>
      </div>
      <Button v-if="isAdmin" @click="showCreate = true">
        <Plus class="mr-2 size-4" />
        New Module
      </Button>
    </div>

    <ModuleFiltersBar v-model="filters" :project-key="projectKey" />

    <div v-if="loading" class="flex items-center justify-center py-12">
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <div
      v-else-if="modules.length === 0"
      class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
    >
      <div class="flex size-16 items-center justify-center rounded-full bg-muted">
        <Layers class="size-8 text-muted-foreground" />
      </div>
      <h3 class="mt-4 text-lg font-semibold">No modules yet</h3>
      <p class="mt-2 max-w-sm text-center text-sm text-muted-foreground">
        Create a module to group tasks into a reusable sub-project with its own
        lead and members.
      </p>
      <Button v-if="isAdmin" class="mt-4" @click="showCreate = true">
        <Plus class="mr-2 size-4" />
        New Module
      </Button>
    </div>

    <template v-else>
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <ModuleCard
          v-for="m in modules"
          :key="m.id"
          :module="m"
          :to="`/projects/${projectKey}/modules/${m.id}`"
        />
      </div>

      <div
        v-if="totalPages > 1"
        class="flex items-center justify-between border-t pt-4"
      >
        <p class="text-sm text-muted-foreground">
          {{ total }} module{{ total === 1 ? "" : "s" }}
        </p>
        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            :disabled="page <= 1"
            @click="goToPage(page - 1)"
          >
            <ChevronLeft class="size-4" />
          </Button>
          <span class="text-sm">Page {{ page }} of {{ totalPages }}</span>
          <Button
            variant="outline"
            size="sm"
            :disabled="page >= totalPages"
            @click="goToPage(page + 1)"
          >
            <ChevronRight class="size-4" />
          </Button>
        </div>
      </div>
    </template>

    <CreateModuleDialog
      v-model:open="showCreate"
      :project-key="projectKey"
      @saved="onSaved"
    />
  </div>
</template>

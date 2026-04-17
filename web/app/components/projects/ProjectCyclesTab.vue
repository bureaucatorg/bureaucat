<script setup lang="ts">
import { Plus, Repeat, Loader2, ChevronLeft, ChevronRight } from "lucide-vue-next";

const props = defineProps<{
  projectKey: string;
  isAdmin: boolean;
}>();

const { cycles, loading, total, page, totalPages, listCycles } = useCycles();

const showCreate = ref(false);
const perPage = 12;

function fetchPage(p = 1) {
  listCycles(props.projectKey, p, perPage);
}

function goToPage(p: number) {
  if (p < 1 || p > totalPages.value) return;
  fetchPage(p);
}

function onCreated() {
  fetchPage(1);
}

onMounted(() => {
  fetchPage(1);
});

watch(
  () => props.projectKey,
  () => fetchPage(1)
);
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-lg font-semibold">Cycles</h2>
        <p class="text-sm text-muted-foreground">
          Time-boxed batches of work. Group tasks into a window with clear start
          and end dates.
        </p>
      </div>
      <Button v-if="isAdmin" @click="showCreate = true">
        <Plus class="mr-2 size-4" />
        New Cycle
      </Button>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <Loader2 class="size-6 animate-spin text-muted-foreground" />
    </div>

    <div
      v-else-if="cycles.length === 0"
      class="flex flex-col items-center justify-center rounded-lg border border-dashed py-16"
    >
      <div class="flex size-16 items-center justify-center rounded-full bg-muted">
        <Repeat class="size-8 text-muted-foreground" />
      </div>
      <h3 class="mt-4 text-lg font-semibold">No cycles yet</h3>
      <p class="mt-2 max-w-sm text-center text-sm text-muted-foreground">
        Create a cycle to plan a batch of work across a time window.
      </p>
      <Button v-if="isAdmin" class="mt-4" @click="showCreate = true">
        <Plus class="mr-2 size-4" />
        New Cycle
      </Button>
    </div>

    <template v-else>
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <CycleCard
          v-for="c in cycles"
          :key="c.id"
          :cycle="c"
          :to="`/projects/${projectKey}/cycles/${c.id}`"
        />
      </div>

      <div
        v-if="totalPages > 1"
        class="flex items-center justify-between border-t pt-4"
      >
        <p class="text-sm text-muted-foreground">
          {{ total }} cycle{{ total === 1 ? "" : "s" }}
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
          <span class="text-sm">
            Page {{ page }} of {{ totalPages }}
          </span>
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

    <CreateCycleDialog
      v-model:open="showCreate"
      :project-key="projectKey"
      @created="onCreated"
    />
  </div>
</template>

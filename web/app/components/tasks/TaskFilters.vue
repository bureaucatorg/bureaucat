<script setup lang="ts">
import { Search, X, Filter } from "lucide-vue-next";
import type { ProjectState, TaskFilters } from "~/types";
import { PRIORITY_LABELS } from "~/types";

const props = defineProps<{
  states: ProjectState[];
  filters: TaskFilters;
}>();

const emit = defineEmits<{
  "update:filters": [filters: TaskFilters];
}>();

const searchQuery = ref(props.filters.q || "");
const selectedStateId = ref(props.filters.state_id || "");
const selectedPriority = ref(props.filters.priority?.toString() || "");

const priorities = [
  { value: "", label: "All priorities" },
  { value: "4", label: "Urgent" },
  { value: "3", label: "High" },
  { value: "2", label: "Medium" },
  { value: "1", label: "Low" },
  { value: "0", label: "No priority" },
];

function updateFilters() {
  const filters: TaskFilters = {};
  if (searchQuery.value) filters.q = searchQuery.value;
  if (selectedStateId.value) filters.state_id = selectedStateId.value;
  if (selectedPriority.value) filters.priority = parseInt(selectedPriority.value);
  emit("update:filters", filters);
}

function clearFilters() {
  searchQuery.value = "";
  selectedStateId.value = "";
  selectedPriority.value = "";
  emit("update:filters", {});
}

const hasActiveFilters = computed(() => {
  return searchQuery.value || selectedStateId.value || selectedPriority.value;
});

watch([searchQuery, selectedStateId, selectedPriority], () => {
  updateFilters();
});
</script>

<template>
  <div class="flex flex-wrap items-center gap-3">
    <div class="relative flex-1 sm:max-w-xs">
      <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        v-model="searchQuery"
        placeholder="Search tasks..."
        class="pl-9"
      />
    </div>

    <NativeSelect v-model="selectedStateId" class="w-auto">
      <option value="">All states</option>
      <option v-for="state in states" :key="state.id" :value="state.id">
        {{ state.name }}
      </option>
    </NativeSelect>

    <NativeSelect v-model="selectedPriority" class="w-auto">
      <option v-for="p in priorities" :key="p.value" :value="p.value">
        {{ p.label }}
      </option>
    </NativeSelect>

    <Button
      v-if="hasActiveFilters"
      variant="ghost"
      size="sm"
      @click="clearFilters"
    >
      <X class="mr-1.5 size-4" />
      Clear
    </Button>
  </div>
</template>

<script setup lang="ts">
import { Plus, X, Loader2 } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { TaskLabel, ProjectLabel } from "~/types";

const props = defineProps<{
  taskLabels: TaskLabel[];
  projectKey: string;
  taskNum: number;
  projectLabels: ProjectLabel[];
  isMember: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const { addLabel, removeLabel } = useTasks();

const loading = ref<string | null>(null);
const showPopover = ref(false);

// Labels not already on the task
const availableLabels = computed(() => {
  const usedIds = new Set(props.taskLabels.map((l) => l.id));
  return props.projectLabels.filter((l) => !usedIds.has(l.id));
});

async function handleAdd(labelId: string) {
  loading.value = labelId;
  const result = await addLabel(props.projectKey, props.taskNum, labelId);
  loading.value = null;

  if (result.success) {
    toast.success("Label added");
    showPopover.value = false;
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to add label");
  }
}

async function handleRemove(labelId: string) {
  loading.value = labelId;
  const result = await removeLabel(props.projectKey, props.taskNum, labelId);
  loading.value = null;

  if (result.success) {
    toast.success("Label removed");
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to remove label");
  }
}
</script>

<template>
  <div class="space-y-2">
    <p class="text-xs text-muted-foreground">Labels</p>

    <div class="flex flex-wrap items-center gap-2">
      <div
        v-for="label in taskLabels"
        :key="label.id"
        class="group relative rounded-md px-2.5 py-1"
        :style="{
          backgroundColor: label.color + '20',
          color: label.color,
        }"
      >
        <span class="text-sm font-medium">{{ label.name }}</span>
        <button
          v-if="isMember"
          type="button"
          :aria-label="`Remove ${label.name}`"
          class="absolute -top-1.5 -right-1.5 flex size-4 items-center justify-center rounded-full bg-foreground text-background opacity-0 shadow-sm transition-opacity group-hover:opacity-100 focus:opacity-100 focus-visible:ring-2 focus-visible:ring-ring outline-none"
          :disabled="loading === label.id"
          @click="handleRemove(label.id)"
        >
          <Loader2
            v-if="loading === label.id"
            class="size-2.5 animate-spin"
          />
          <X v-else class="size-2.5" />
        </button>
      </div>

      <!-- Add button -->
      <SearchableSelect
        v-if="isMember && availableLabels.length > 0"
        v-model:open="showPopover"
        :items="availableLabels"
        :get-search-text="(l) => l.name"
        :get-key="(l) => l.id"
        placeholder="Search labels..."
        empty-text="No labels found"
        content-class="w-48"
        @select="(l) => handleAdd(l.id)"
      >
        <template #trigger>
          <Button variant="outline" size="sm" class="h-7 gap-1.5">
            <Plus class="size-3.5" />
            Add
          </Button>
        </template>
        <template #option="{ item: label }">
          <div
            class="size-3 shrink-0 rounded-full"
            :style="{ backgroundColor: label.color }"
          />
          {{ label.name }}
        </template>
      </SearchableSelect>

      <!-- Empty state -->
      <span
        v-if="taskLabels.length === 0 && !isMember"
        class="text-sm text-muted-foreground"
      >
        No labels
      </span>
    </div>
  </div>
</template>

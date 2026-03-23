<script setup lang="ts">
import { Plus, X, Loader2, Search } from "lucide-vue-next";
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
const searchQuery = ref("");

// Labels not already on the task
const availableLabels = computed(() => {
  const usedIds = new Set(props.taskLabels.map((l) => l.id));
  return props.projectLabels.filter((l) => !usedIds.has(l.id));
});

const filteredLabels = computed(() => {
  const q = searchQuery.value.toLowerCase().trim();
  if (!q) return availableLabels.value;
  return availableLabels.value.filter((l) => l.name.toLowerCase().includes(q));
});

watch(showPopover, (open) => {
  if (!open) searchQuery.value = "";
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
      <Popover v-if="isMember && availableLabels.length > 0" v-model:open="showPopover">
        <PopoverTrigger as-child>
          <Button variant="outline" size="sm" class="h-7 gap-1.5">
            <Plus class="size-3.5" />
            Add
          </Button>
        </PopoverTrigger>
        <PopoverContent align="start" class="w-48 p-0">
          <div class="border-b px-3 py-2">
            <div class="relative">
              <Search class="absolute left-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
              <Input
                v-model="searchQuery"
                placeholder="Search labels..."
                class="h-8 pl-7 text-sm"
              />
            </div>
          </div>
          <div class="max-h-48 overflow-y-auto">
            <div class="py-1">
              <button
                v-for="label in filteredLabels"
                :key="label.id"
                type="button"
                class="flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-accent disabled:opacity-50"
                :disabled="loading === label.id"
                @click="handleAdd(label.id)"
              >
                <div
                  class="size-3 shrink-0 rounded-full"
                  :style="{ backgroundColor: label.color }"
                />
                {{ label.name }}
              </button>
              <p
                v-if="filteredLabels.length === 0"
                class="px-3 py-2 text-center text-sm text-muted-foreground"
              >
                No labels found
              </p>
            </div>
          </div>
        </PopoverContent>
      </Popover>

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

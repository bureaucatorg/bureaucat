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
const showDropdown = ref(false);

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
    showDropdown.value = false;
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
    <p class="text-sm font-medium text-muted-foreground">Labels</p>

    <div class="flex flex-wrap items-center gap-2">
      <div
        v-for="label in taskLabels"
        :key="label.id"
        class="group flex items-center gap-1.5 rounded-full px-2.5 py-1"
        :style="{
          backgroundColor: label.color + '20',
          color: label.color,
        }"
      >
        <span class="text-sm font-medium">{{ label.name }}</span>
        <button
          v-if="isMember"
          type="button"
          class="rounded-full p-0.5 opacity-0 transition-opacity hover:bg-black/10 group-hover:opacity-100"
          :disabled="loading === label.id"
          @click="handleRemove(label.id)"
        >
          <Loader2
            v-if="loading === label.id"
            class="size-3 animate-spin"
          />
          <X v-else class="size-3" />
        </button>
      </div>

      <!-- Add button -->
      <DropdownMenu v-if="isMember && availableLabels.length > 0" v-model:open="showDropdown">
        <DropdownMenuTrigger as-child>
          <Button variant="outline" size="sm" class="h-7 gap-1.5">
            <Plus class="size-3.5" />
            Add
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" class="w-48">
          <DropdownMenuLabel>Add label</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            v-for="label in availableLabels"
            :key="label.id"
            :disabled="loading === label.id"
            @click="handleAdd(label.id)"
          >
            <div
              class="mr-2 size-3 rounded-full"
              :style="{ backgroundColor: label.color }"
            />
            {{ label.name }}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

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

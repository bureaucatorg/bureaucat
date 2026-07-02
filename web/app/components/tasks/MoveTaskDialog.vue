<script setup lang="ts">
import { Loader2, Search, Check } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { Project, MoveTasksResponse } from "~/types";

// Move one task (taskNum) or several (taskNumbers) to another project. The
// destination list is limited to projects the current user can see (the API
// only returns projects they're a member of) and excludes the source project.
const props = defineProps<{
  projectKey: string;
  // Single-move mode.
  taskNum?: number;
  // Bulk-move mode.
  taskNumbers?: number[];
}>();

const open = defineModel<boolean>("open", { default: false });
const emit = defineEmits<{
  // Fired after a successful move. `newTaskNumber` is only set in single mode.
  moved: [payload: { targetKey: string; newTaskNumber?: number; result?: MoveTasksResponse }];
}>();

const { listProjects } = useProjects();
const { moveTask, moveTasks } = useTasks();

const projects = ref<Project[]>([]);
const projectsLoading = ref(false);
const search = ref("");
const selectedKey = ref<string | null>(null);
const submitting = ref(false);
const error = ref<string | null>(null);

const count = computed(() =>
  props.taskNumbers ? props.taskNumbers.length : 1
);

const filtered = computed(() => {
  const q = search.value.toLowerCase().trim();
  const list = projects.value.filter((p) => p.project_key !== props.projectKey);
  if (!q) return list;
  return list.filter(
    (p) =>
      p.name.toLowerCase().includes(q) ||
      p.project_key.toLowerCase().includes(q)
  );
});

async function loadProjects() {
  projectsLoading.value = true;
  const res = await listProjects(1, 100);
  projectsLoading.value = false;
  if (res.success) projects.value = res.data?.projects || [];
}

watch(open, (isOpen) => {
  if (!isOpen) return;
  error.value = null;
  search.value = "";
  selectedKey.value = null;
  loadProjects();
});

async function handleMove() {
  if (!selectedKey.value || submitting.value) return;
  const targetKey = selectedKey.value;
  submitting.value = true;
  error.value = null;

  if (props.taskNumbers) {
    const res = await moveTasks(props.projectKey, props.taskNumbers, targetKey);
    submitting.value = false;
    if (!res.success || !res.data) {
      error.value = res.error || "Failed to move tasks";
      return;
    }
    open.value = false;
    emit("moved", { targetKey, result: res.data });
  } else if (props.taskNum !== undefined) {
    const res = await moveTask(props.projectKey, props.taskNum, targetKey);
    submitting.value = false;
    if (!res.success || !res.data) {
      error.value = res.error || "Failed to move task";
      return;
    }
    open.value = false;
    emit("moved", { targetKey, newTaskNumber: res.data.task_number });
  } else {
    submitting.value = false;
    error.value = "Nothing to move";
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>
          Move {{ count }} task{{ count === 1 ? "" : "s" }}
        </DialogTitle>
        <DialogDescription>
          Pick a destination project. The task gets a new ID in that project;
          its state and labels are matched by name, and cycle/module links are
          cleared.
        </DialogDescription>
      </DialogHeader>

      <div
        v-if="error"
        class="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
      >
        {{ error }}
      </div>

      <div class="space-y-3">
        <div class="relative">
          <Search
            class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
          />
          <Input v-model="search" placeholder="Search projects..." class="pl-9" />
        </div>

        <div class="max-h-72 overflow-y-auto rounded-md border [scrollbar-gutter:stable]">
          <div
            v-if="projectsLoading"
            class="flex items-center justify-center py-10 text-sm text-muted-foreground"
          >
            <Loader2 class="mr-2 size-4 animate-spin" /> Loading…
          </div>
          <div
            v-else-if="filtered.length === 0"
            class="py-10 text-center text-sm text-muted-foreground"
          >
            No other projects available.
          </div>
          <button
            v-for="p in filtered"
            v-else
            :key="p.id"
            type="button"
            class="flex w-full items-center gap-3 border-b border-border/40 px-3 py-2 text-left last:border-0 hover:bg-muted/40"
            :class="selectedKey === p.project_key ? 'bg-accent' : ''"
            @click="selectedKey = p.project_key"
          >
            <span
              class="inline-flex w-fit shrink-0 items-center rounded px-1.5 py-0.5 font-mono text-[10px] font-medium uppercase tracking-wider"
            >
              {{ p.project_key }}
            </span>
            <span class="min-w-0 flex-1 truncate text-sm">{{ p.name }}</span>
            <Check
              v-if="selectedKey === p.project_key"
              class="size-4 shrink-0 text-primary"
            />
          </button>
        </div>
      </div>

      <DialogFooter>
        <Button
          type="button"
          variant="outline"
          :disabled="submitting"
          @click="open = false"
        >
          Cancel
        </Button>
        <Button :disabled="submitting || !selectedKey" @click="handleMove">
          <Loader2 v-if="submitting" class="mr-2 size-4 animate-spin" />
          Move {{ count === 1 ? "task" : count + " tasks" }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

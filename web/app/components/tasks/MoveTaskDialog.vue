<script setup lang="ts">
import { Loader2, Search, Check, FolderOpen } from "lucide-vue-next";
import { watchDebounced } from "@vueuse/core";
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

// The source project is never a valid destination; drop it client-side since
// the server has no reason to know which project we're moving from.
const filtered = computed(() =>
  projects.value.filter((p) => p.project_key !== props.projectKey)
);

// Search runs on the backend so instances with more than a page of projects
// still surface matches. We cap the page size at a comfortable list length —
// users refine via the search box rather than scrolling hundreds of rows.
async function loadProjects(query = "") {
  projectsLoading.value = true;
  // Destinations span every workspace the user belongs to (workspaceScoped:
  // false), so a task can be moved out of its current workspace.
  const res = await listProjects(1, 50, query.trim(), false);
  projectsLoading.value = false;
  if (res.success) projects.value = res.data?.projects || [];
}

// Debounce so we don't fire a request on every keystroke.
watchDebounced(
  search,
  (q) => {
    if (open.value) loadProjects(q);
  },
  { debounce: 300 }
);

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
          <Input
            v-model="search"
            placeholder="Search projects..."
            class="pl-9 pr-9"
          />
          <Loader2
            v-if="projectsLoading"
            class="absolute right-3 top-1/2 size-4 -translate-y-1/2 animate-spin text-muted-foreground"
          />
        </div>

        <ScrollArea class="h-72 rounded-md border">
          <div
            v-if="projectsLoading && filtered.length === 0"
            class="flex flex-col items-center justify-center gap-2 py-16 text-sm text-muted-foreground"
          >
            <Loader2 class="size-5 animate-spin" />
            Loading projects…
          </div>
          <div
            v-else-if="filtered.length === 0"
            class="flex flex-col items-center justify-center gap-2 py-16 text-center text-sm text-muted-foreground"
          >
            <FolderOpen class="size-5" />
            {{ search.trim() ? "No projects match your search." : "No other projects available." }}
          </div>
          <div v-else class="p-1.5">
            <button
              v-for="p in filtered"
              :key="p.id"
              type="button"
              class="flex w-full items-center gap-3 rounded-md px-2 py-2 text-left transition-colors hover:bg-muted focus-visible:bg-muted focus-visible:outline-none"
              :class="selectedKey === p.project_key ? 'bg-accent' : ''"
              @click="selectedKey = p.project_key"
            >
              <span
                class="inline-flex h-6 w-16 shrink-0 items-center justify-center rounded-md bg-muted font-mono text-[10px] font-semibold uppercase tracking-wider text-muted-foreground"
                :title="p.project_key"
              >
                <span class="truncate px-1">{{ p.project_key }}</span>
              </span>
              <span class="min-w-0 flex-1 truncate text-sm">{{ p.name }}</span>
              <Check
                v-if="selectedKey === p.project_key"
                class="size-4 shrink-0 text-primary"
              />
            </button>
          </div>
        </ScrollArea>
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

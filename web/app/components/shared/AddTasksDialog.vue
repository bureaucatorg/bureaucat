<script setup lang="ts" generic="T extends PickerTask">
import { Loader2, Search } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { ProjectState } from "~/types";

// Minimum shape the picker needs from each task row.
interface PickerTask {
  id: string;
  title: string;
  task_id: string;
  state_name: string;
  state_color: string;
}

type PickResult<R> = { success: boolean; data?: R; error?: string };
type AckResult = { success: boolean; error?: string };

const props = defineProps<{
  projectKey: string;
  collectionId: string;
  title?: string;
  description?: string;
  emptyHint?: string;
  // Data loader: return picker tasks matching `search`.
  loadTasks: (search: string, limit: number) => Promise<PickResult<T[]>>;
  // Bulk add: attach existing tasks to the collection.
  addTasks: (taskIds: string[]) => Promise<AckResult>;
  // Create-and-add flow. Called with the new task's id after creation succeeds.
  // Kept separate from addTasks so consumers can run side-effects (e.g. toasts)
  // between the two steps if they like.
}>();

const open = defineModel<boolean>("open", { default: false });
const emit = defineEmits<{ added: [] }>();

const { createTask } = useTasks();
const { states, listStates } = useProjects();

const tab = ref<"existing" | "new">("existing");
const loading = ref(false);
const error = ref<string | null>(null);

const search = ref("");
const existingTasks = ref<T[]>([]);
const existingLoading = ref(false);
const selectedIds = ref<Set<string>>(new Set());
let searchDebounce: ReturnType<typeof setTimeout> | null = null;

const newForm = ref({ title: "", state_id: "" as string });

async function loadExisting() {
  existingLoading.value = true;
  const result = await props.loadTasks(search.value, 100);
  existingLoading.value = false;
  if (result.success) existingTasks.value = result.data || [];
}

watch(open, async (isOpen) => {
  if (!isOpen) return;
  error.value = null;
  selectedIds.value = new Set();
  search.value = "";
  tab.value = "existing";
  newForm.value = { title: "", state_id: "" };
  if (!states.value.length) {
    await listStates(props.projectKey);
  }
  const defaultState = states.value.find((s: ProjectState) => s.state_type === "unstarted");
  newForm.value.state_id = defaultState?.id || states.value[0]?.id || "";
  await loadExisting();
});

watch(search, () => {
  if (searchDebounce) clearTimeout(searchDebounce);
  searchDebounce = setTimeout(() => {
    loadExisting();
  }, 250);
});

function toggleSelect(id: string) {
  if (selectedIds.value.has(id)) selectedIds.value.delete(id);
  else selectedIds.value.add(id);
  selectedIds.value = new Set(selectedIds.value);
}

async function handleAddExisting() {
  if (selectedIds.value.size === 0) return;
  loading.value = true;
  error.value = null;
  const ids = Array.from(selectedIds.value);
  const result = await props.addTasks(ids);
  loading.value = false;
  if (result.success) {
    toast.success(`Added ${ids.length} task${ids.length === 1 ? "" : "s"}`);
    open.value = false;
    emit("added");
  } else {
    error.value = result.error || "Failed to add tasks";
  }
}

async function handleCreateNew() {
  if (!newForm.value.title.trim()) return;
  loading.value = true;
  error.value = null;
  const created = await createTask(props.projectKey, {
    title: newForm.value.title.trim(),
    state_id: newForm.value.state_id || undefined,
  });
  if (!created.success || !created.data) {
    loading.value = false;
    error.value = created.error || "Failed to create task";
    return;
  }
  const added = await props.addTasks([created.data.id]);
  loading.value = false;
  if (added.success) {
    toast.success("Task created and added");
    open.value = false;
    emit("added");
  } else {
    error.value = added.error || "Task created but failed to attach";
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="sm:max-w-3xl">
      <DialogHeader>
        <DialogTitle>{{ title || "Add tasks" }}</DialogTitle>
        <DialogDescription>
          {{ description || "Pick existing tasks or create a brand new one." }}
        </DialogDescription>
      </DialogHeader>

      <div
        v-if="error"
        class="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
      >
        {{ error }}
      </div>

      <Tabs v-model="tab" class="w-full min-w-0">
        <TabsList class="grid w-full grid-cols-2">
          <TabsTrigger value="existing">Existing</TabsTrigger>
          <TabsTrigger value="new">New</TabsTrigger>
        </TabsList>

        <TabsContent value="existing" class="mt-3 min-w-0 space-y-3">
          <div class="relative">
            <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input v-model="search" placeholder="Search tasks..." class="pl-9" />
          </div>

          <div class="overflow-hidden rounded-md border">
            <div
              class="grid items-center gap-3 border-b bg-muted/40 px-3 py-2 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground"
              style="grid-template-columns: 20px 160px minmax(0, 1fr) 80px;"
            >
              <span></span>
              <span>State</span>
              <span>Title</span>
              <span>ID</span>
            </div>
            <div class="max-h-96 overflow-y-auto [scrollbar-gutter:stable]">
              <div
                v-if="existingLoading"
                class="flex items-center justify-center py-10 text-sm text-muted-foreground"
              >
                <Loader2 class="mr-2 size-4 animate-spin" /> Loading…
              </div>
              <div
                v-else-if="existingTasks.length === 0"
                class="py-10 text-center text-sm text-muted-foreground"
              >
                {{ emptyHint || "No tasks found." }}
              </div>
              <label
                v-for="task in existingTasks"
                v-else
                :key="task.id"
                class="grid cursor-pointer items-center gap-3 border-b border-border/40 px-3 py-2 last:border-0 hover:bg-muted/40"
                style="grid-template-columns: 20px 160px minmax(0, 1fr) 80px;"
              >
                <Checkbox
                  :model-value="selectedIds.has(task.id)"
                  @update:model-value="toggleSelect(task.id)"
                />
                <span
                  class="inline-flex w-fit max-w-full items-center truncate rounded px-1.5 py-0.5 font-mono text-[10px] font-medium uppercase tracking-wider"
                  :style="{
                    backgroundColor: (task.state_color || '#6B7280') + '22',
                    color: task.state_color || '#6B7280',
                  }"
                >
                  {{ task.state_name }}
                </span>
                <span class="min-w-0 truncate text-sm">{{ task.title }}</span>
                <span class="font-mono text-[11px] text-muted-foreground">
                  {{ task.task_id }}
                </span>
              </label>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="new" class="mt-3 min-w-0 space-y-3">
          <div class="space-y-2">
            <Label for="shared_new_task_title">Title</Label>
            <Input
              id="shared_new_task_title"
              v-model="newForm.title"
              placeholder="Ship the new feature"
              required
              :disabled="loading"
            />
          </div>
          <div class="space-y-2">
            <Label for="shared_new_task_state">Initial state</Label>
            <select
              id="shared_new_task_state"
              v-model="newForm.state_id"
              class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="loading"
            >
              <option v-for="s in states" :key="s.id" :value="s.id">
                {{ s.name }}
              </option>
            </select>
          </div>
        </TabsContent>
      </Tabs>

      <DialogFooter>
        <Button type="button" variant="outline" :disabled="loading" @click="open = false">
          Cancel
        </Button>
        <Button
          v-if="tab === 'existing'"
          :disabled="loading || selectedIds.size === 0"
          @click="handleAddExisting"
        >
          <Loader2 v-if="loading" class="mr-2 size-4 animate-spin" />
          Add {{ selectedIds.size || "" }} task{{ selectedIds.size === 1 ? "" : "s" }}
        </Button>
        <Button
          v-else
          :disabled="loading || !newForm.title.trim()"
          @click="handleCreateNew"
        >
          <Loader2 v-if="loading" class="mr-2 size-4 animate-spin" />
          Create &amp; Add
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

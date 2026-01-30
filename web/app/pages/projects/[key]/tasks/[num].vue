<script setup lang="ts">
import {
  ChevronLeft,
  ChevronDown,
  Loader2,
  Pencil,
  Trash2,
  Check,
  X,
  Calendar,
  Clock,
} from "lucide-vue-next";
import { toast } from "vue-sonner";
import { marked } from "marked";
import { PRIORITY_LABELS } from "~/types";

marked.setOptions({
  breaks: true,
  gfm: true,
});

definePageMeta({
  middleware: ["auth"],
});

const route = useRoute();
const router = useRouter();

const projectKey = computed(() => route.params.key as string);
const taskNum = computed(() => parseInt(route.params.num as string));

const {
  currentProject,
  members,
  states,
  labels: projectLabels,
  getProject,
  listMembers,
  listStates,
  listLabels,
} = useProjects();

const { currentTask, getTask, updateTask, deleteTask } = useTasks();
const { comments, loading: commentsLoading, listComments } = useComments();
const { activities, loading: activitiesLoading, listActivity } = useActivity();

const loading = ref(true);
const error = ref<string | null>(null);
const editingTitle = ref(false);
const editingDescription = ref(false);
const editTitle = ref("");
const editDescription = ref("");
const updating = ref(false);
const deleting = ref(false);
const showDeleteConfirm = ref(false);

const isAdmin = computed(() => currentProject.value?.role === "admin");
const isMember = computed(
  () => currentProject.value?.role === "admin" || currentProject.value?.role === "member"
);

const priorityOptions = Object.entries(PRIORITY_LABELS).map(([value, info]) => ({
  value: parseInt(value),
  label: info.label,
  color: info.color,
}));

const currentPriority = computed(() => {
  const p = currentTask.value?.priority ?? 0;
  return PRIORITY_LABELS[p] || PRIORITY_LABELS[0];
});

async function loadData() {
  loading.value = true;
  error.value = null;

  // Load project data if not already loaded
  if (!currentProject.value || currentProject.value.project_key !== projectKey.value) {
    const projectResult = await getProject(projectKey.value);
    if (!projectResult.success) {
      error.value = projectResult.error || "Failed to load project";
      loading.value = false;
      return;
    }
  }

  // Load task
  const taskResult = await getTask(projectKey.value, taskNum.value);
  if (!taskResult.success) {
    error.value = taskResult.error || "Task not found";
    loading.value = false;
    return;
  }

  // Load supporting data in parallel
  await Promise.all([
    listMembers(projectKey.value),
    listStates(projectKey.value),
    listLabels(projectKey.value),
    listComments(projectKey.value, taskNum.value),
    listActivity(projectKey.value, taskNum.value),
  ]);

  loading.value = false;
}

function startEditTitle() {
  editTitle.value = currentTask.value?.title || "";
  editingTitle.value = true;
}

function cancelEditTitle() {
  editingTitle.value = false;
  editTitle.value = "";
}

async function saveTitle() {
  if (!editTitle.value.trim() || editTitle.value === currentTask.value?.title) {
    cancelEditTitle();
    return;
  }

  updating.value = true;
  const result = await updateTask(projectKey.value, taskNum.value, {
    title: editTitle.value,
  });
  updating.value = false;

  if (result.success) {
    toast.success("Title updated");
    cancelEditTitle();
  } else {
    toast.error(result.error || "Failed to update title");
  }
}

function startEditDescription() {
  editDescription.value = currentTask.value?.description || "";
  editingDescription.value = true;
}

function cancelEditDescription() {
  editingDescription.value = false;
  editDescription.value = "";
}

async function saveDescription() {
  if (editDescription.value === (currentTask.value?.description || "")) {
    cancelEditDescription();
    return;
  }

  updating.value = true;
  const result = await updateTask(projectKey.value, taskNum.value, {
    description: editDescription.value || undefined,
  });
  updating.value = false;

  if (result.success) {
    toast.success("Description updated");
    cancelEditDescription();
  } else {
    toast.error(result.error || "Failed to update description");
  }
}

async function handleStateChange(stateId: string) {
  updating.value = true;
  const result = await updateTask(projectKey.value, taskNum.value, {
    state_id: stateId,
  });
  updating.value = false;

  if (result.success) {
    toast.success("State updated");
    await listActivity(projectKey.value, taskNum.value);
  } else {
    toast.error(result.error || "Failed to update state");
  }
}

async function handlePriorityChange(priority: number) {
  updating.value = true;
  const result = await updateTask(projectKey.value, taskNum.value, {
    priority,
  });
  updating.value = false;

  if (result.success) {
    toast.success("Priority updated");
    await listActivity(projectKey.value, taskNum.value);
  } else {
    toast.error(result.error || "Failed to update priority");
  }
}

async function handleDelete() {
  deleting.value = true;
  const result = await deleteTask(projectKey.value, taskNum.value);
  deleting.value = false;

  if (result.success) {
    toast.success("Task deleted");
    router.push(`/projects/${projectKey.value}`);
  } else {
    toast.error(result.error || "Failed to delete task");
  }
}

async function refreshTask() {
  await Promise.all([
    getTask(projectKey.value, taskNum.value),
    listActivity(projectKey.value, taskNum.value),
  ]);
}

async function refreshComments() {
  await listComments(projectKey.value, taskNum.value);
  await listActivity(projectKey.value, taskNum.value);
}

const renderedDescription = computed(() => {
  if (!currentTask.value?.description) return "";
  return marked(currentTask.value.description) as string;
});

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

onMounted(() => {
  loadData();
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-8">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-20">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <!-- Error -->
        <div
          v-else-if="error"
          class="flex flex-col items-center justify-center py-20"
        >
          <p class="text-lg text-destructive">{{ error }}</p>
          <Button class="mt-4" variant="outline" as-child>
            <NuxtLink :to="`/projects/${projectKey}`">
              Back to Project
            </NuxtLink>
          </Button>
        </div>

        <!-- Task content -->
        <template v-else-if="currentTask">
          <!-- Breadcrumb -->
          <nav class="mb-6 flex items-center gap-2 text-sm text-muted-foreground">
            <NuxtLink
              :to="`/projects/${projectKey}`"
              class="flex items-center gap-1 hover:text-foreground"
            >
              <ChevronLeft class="size-4" />
              {{ currentProject?.name }}
            </NuxtLink>
            <span>/</span>
            <span class="font-semibold text-amber-600 dark:text-amber-500">
              {{ currentTask.task_id }}
            </span>
          </nav>

          <div class="flex flex-col gap-8 md:flex-row">
            <!-- Main content -->
            <div class="min-w-0 flex-1 space-y-6">
              <!-- Title -->
              <div>
                <div v-if="editingTitle" class="space-y-2">
                  <Input
                    v-model="editTitle"
                    class="text-xl font-bold"
                    :disabled="updating"
                    @keydown.enter="saveTitle"
                    @keydown.escape="cancelEditTitle"
                  />
                  <div class="flex gap-2">
                    <Button size="sm" :disabled="updating" @click="saveTitle">
                      <Loader2 v-if="updating" class="mr-1.5 size-3 animate-spin" />
                      <Check v-else class="mr-1.5 size-3" />
                      Save
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      :disabled="updating"
                      @click="cancelEditTitle"
                    >
                      <X class="mr-1.5 size-3" />
                      Cancel
                    </Button>
                  </div>
                </div>
                <div v-else class="group flex items-start gap-2">
                  <h1 class="text-2xl font-bold">{{ currentTask.title }}</h1>
                  <Button
                    v-if="isMember"
                    variant="ghost"
                    size="icon"
                    class="mt-1 size-7 opacity-0 transition-opacity group-hover:opacity-100"
                    @click="startEditTitle"
                  >
                    <Pencil class="size-3.5" />
                  </Button>
                </div>
              </div>

              <!-- Description -->
              <div>
                <h3 class="mb-2 text-sm font-medium text-muted-foreground">
                  Description
                </h3>
                <div v-if="editingDescription" class="space-y-2">
                  <Textarea
                    v-model="editDescription"
                    rows="4"
                    :disabled="updating"
                    placeholder="Add a description..."
                  />
                  <div class="flex gap-2">
                    <Button size="sm" :disabled="updating" @click="saveDescription">
                      <Loader2 v-if="updating" class="mr-1.5 size-3 animate-spin" />
                      <Check v-else class="mr-1.5 size-3" />
                      Save
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      :disabled="updating"
                      @click="cancelEditDescription"
                    >
                      <X class="mr-1.5 size-3" />
                      Cancel
                    </Button>
                  </div>
                </div>
                <div v-else class="group relative">
                  <div v-if="currentTask.description">
                    <div
                      class="prose prose-sm max-w-none dark:prose-invert"
                      v-html="renderedDescription"
                    />
                    <Button
                      v-if="isMember"
                      variant="ghost"
                      size="icon"
                      class="absolute top-0 right-0 size-7 opacity-0 transition-opacity group-hover:opacity-100"
                      @click="startEditDescription"
                    >
                      <Pencil class="size-3.5" />
                    </Button>
                  </div>
                  <button
                    v-else-if="isMember"
                    type="button"
                    class="w-full rounded-lg border border-dashed p-4 text-left text-sm text-muted-foreground hover:border-solid hover:bg-muted/50"
                    @click="startEditDescription"
                  >
                    Add a description...
                  </button>
                  <p v-else class="text-sm italic text-muted-foreground">
                    No description
                  </p>
                </div>
              </div>

              <Separator />

              <!-- Activity & Comments Combined -->
              <TaskActivityFeed
                :activities="activities"
                :comments="comments"
                :project-key="projectKey"
                :task-num="taskNum"
                :activities-loading="activitiesLoading"
                :comments-loading="commentsLoading"
                :is-member="isMember"
                @refresh-comments="refreshComments"
                @refresh-activity="listActivity(projectKey, taskNum)"
              />

              <!-- Comment Form -->
              <CommentForm
                v-if="isMember"
                :project-key="projectKey"
                :task-num="taskNum"
                @created="refreshComments"
              />
            </div>

            <!-- Sidebar -->
            <div class="w-full border-border pl-8 md:w-64 md:shrink-0 md:border-l">
              <div class="divide-y divide-border">
                <!-- State -->
                <div class="flex items-center justify-between py-3">
                  <p class="text-xs text-muted-foreground">State</p>
                  <TaskStateSelector
                    :states="states"
                    :model-value="currentTask.state_id"
                    :disabled="!isMember || updating"
                    compact
                    @update:model-value="handleStateChange"
                  />
                </div>

                <!-- Priority -->
                <div class="flex items-center justify-between py-3">
                  <p class="text-xs text-muted-foreground">Priority</p>
                  <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                      <Button
                        variant="ghost"
                        class="h-auto gap-1.5 px-0 py-0 font-medium hover:bg-transparent"
                        :disabled="!isMember || updating"
                      >
                        <span
                          class="size-2 rounded-full"
                          :style="{ backgroundColor: currentPriority.color }"
                        />
                        {{ currentPriority.label }}
                        <ChevronDown class="size-3.5 opacity-50" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" class="w-40">
                      <DropdownMenuItem
                        v-for="p in priorityOptions"
                        :key="p.value"
                        @click="handlePriorityChange(p.value)"
                      >
                        <span
                          class="mr-2 size-2 rounded-full"
                          :style="{ backgroundColor: p.color }"
                        />
                        {{ p.label }}
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>

                <!-- Assignees -->
                <div class="py-3">
                  <TaskAssignees
                    :assignees="currentTask.assignees || []"
                    :project-key="projectKey"
                    :task-num="taskNum"
                    :members="members"
                    :is-member="isMember"
                    @refresh="refreshTask"
                  />
                </div>

                <!-- Labels -->
                <div class="py-3">
                  <TaskLabels
                    :task-labels="currentTask.labels || []"
                    :project-key="projectKey"
                    :task-num="taskNum"
                    :project-labels="projectLabels"
                    :is-member="isMember"
                    @refresh="refreshTask"
                  />
                </div>

                <!-- Dates -->
                <div class="space-y-1.5 py-3 text-xs text-muted-foreground">
                  <div class="flex items-center gap-2">
                    <Calendar class="size-3.5" />
                    <span>Created {{ formatDate(currentTask.created_at) }}</span>
                  </div>
                  <div class="flex items-center gap-2">
                    <Clock class="size-3.5" />
                    <span>Updated {{ formatDate(currentTask.updated_at) }}</span>
                  </div>
                </div>
              </div>

              <!-- Delete -->
              <div v-if="isAdmin" class="mt-2 border-t border-border pt-3">
                <Button
                  variant="ghost"
                  size="sm"
                  class="w-full justify-start gap-2 text-destructive hover:bg-destructive/10 hover:text-destructive"
                  @click="showDeleteConfirm = true"
                >
                  <Trash2 class="size-3.5" />
                  Delete task
                </Button>
              </div>
            </div>
          </div>
        </template>

        <!-- Delete confirmation -->
        <Dialog v-model:open="showDeleteConfirm">
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Delete Task</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete
                <strong>{{ currentTask?.task_id }}</strong>?
                This action cannot be undone.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button
                variant="outline"
                :disabled="deleting"
                @click="showDeleteConfirm = false"
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                :disabled="deleting"
                @click="handleDelete"
              >
                <Loader2 v-if="deleting" class="mr-2 size-4 animate-spin" />
                Delete
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </main>
  </div>
</template>

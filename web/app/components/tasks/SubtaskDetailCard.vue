<script setup lang="ts">
import {
  Maximize2,
  Loader2,
  Calendar as CalendarIcon,
  Clock,
  Pencil,
  Check,
  X,
  ChevronDown,
  MessageSquare,
  Send,
  ArrowUpDown,
} from "lucide-vue-next";
import { toast } from "vue-sonner";
import { marked } from "marked";
import { CalendarDate, type DateValue } from "@internationalized/date";
import type { Task, ProjectState, ProjectMember, ProjectLabel, Comment } from "~/types";
import { PRIORITY_LABELS } from "~/types";

const props = withDefaults(
  defineProps<{
    projectKey: string;
    taskNumber: number;
    states?: ProjectState[];
    members?: ProjectMember[];
    projectLabels?: ProjectLabel[];
    isMember?: boolean;
  }>(),
  { states: () => [], members: () => [], projectLabels: () => [], isMember: false }
);

// Bubbles up whenever a field changes so the parent subtask list can refresh
// the affected row (state badge, title, avatars, ...).
const emit = defineEmits<{ updated: [] }>();

const { getAuthHeader, user } = useAuth();
const { updateTask } = useTasks();

const task = ref<Task | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);
const updating = ref(false);

const priorityOptions = Object.entries(PRIORITY_LABELS).map(([value, info]) => ({
  value: parseInt(value),
  label: info.label,
  color: info.color,
}));

const priority = computed(() => {
  const p = task.value?.priority ?? 0;
  return PRIORITY_LABELS[p] || PRIORITY_LABELS[0];
});

const renderedDescription = computed(() => {
  const desc = task.value?.description;
  if (!desc) return "";
  return desc.startsWith("<") ? desc : (marked(desc) as string);
});

const taskLink = computed(() => `/projects/${props.projectKey}/tasks/${props.taskNumber}`);

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

// Fetch full task details directly (not via the shared store) so opening the
// card from a parent task's page doesn't clobber the parent bound to the view.
async function loadTask() {
  loading.value = true;
  error.value = null;
  try {
    const response = await fetch(
      `/api/v1/projects/${props.projectKey}/tasks/${props.taskNumber}`,
      { headers: getAuthHeader() }
    );
    if (!response.ok) {
      const body = await response.json().catch(() => ({}));
      error.value = body.message || "Failed to load subtask";
      return;
    }
    task.value = await response.json();
  } catch {
    error.value = "Network error";
  } finally {
    loading.value = false;
  }
}

// Apply a PATCH, sync the local copy from the response, and let the parent know.
async function patchTask(
  data: import("~/types").UpdateTaskRequest,
  successMsg: string
): Promise<boolean> {
  updating.value = true;
  const result = await updateTask(props.projectKey, props.taskNumber, data);
  updating.value = false;
  if (result.success) {
    if (result.data) task.value = result.data;
    toast.success(successMsg);
    emit("updated");
    return true;
  }
  toast.error(result.error || "Failed to update");
  return false;
}

// --- Title ---
const editingTitle = ref(false);
const editTitle = ref("");

function startEditTitle() {
  editTitle.value = task.value?.title || "";
  editingTitle.value = true;
}
function cancelEditTitle() {
  editingTitle.value = false;
  editTitle.value = "";
}
async function saveTitle() {
  if (!editTitle.value.trim() || editTitle.value === task.value?.title) {
    cancelEditTitle();
    return;
  }
  if (await patchTask({ title: editTitle.value }, "Title updated")) cancelEditTitle();
}

// --- Description ---
const editingDescription = ref(false);
const editDescription = ref("");

function startEditDescription() {
  const desc = task.value?.description || "";
  editDescription.value = desc.startsWith("<") ? desc : (marked(desc) as string);
  editingDescription.value = true;
}
function cancelEditDescription() {
  editingDescription.value = false;
  editDescription.value = "";
}
async function saveDescription() {
  const isEmpty = !editDescription.value || editDescription.value === "<p></p>";
  const current = task.value?.description || "";
  if (editDescription.value === current) {
    cancelEditDescription();
    return;
  }
  if (
    await patchTask(
      { description: isEmpty ? undefined : editDescription.value },
      "Description updated"
    )
  ) {
    cancelEditDescription();
  }
}

// --- State & Priority ---
async function handleStateChange(stateId: string) {
  if (stateId === task.value?.state_id) return;
  await patchTask({ state_id: stateId }, "State updated");
}
async function handlePriorityChange(p: number) {
  if (p === task.value?.priority) return;
  await patchTask({ priority: p }, "Priority updated");
}

// --- Dates ---
const startDateOpen = ref(false);
const dueDateOpen = ref(false);
const startDateDraft = ref<DateValue | undefined>();
const startTimeDraft = ref("09:00");
const dueDateDraft = ref<DateValue | undefined>();
const dueTimeDraft = ref("17:00");

function isoToCalendarDate(iso: string | undefined): DateValue | undefined {
  if (!iso) return undefined;
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return undefined;
  return new CalendarDate(d.getFullYear(), d.getMonth() + 1, d.getDate());
}
function isoToTimeInput(iso: string | undefined, fallback: string): string {
  if (!iso) return fallback;
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return fallback;
  return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}
function combineDateTime(date: DateValue, time: string): string {
  const [hh, mm] = time.split(":").map((n) => parseInt(n, 10));
  const d = new Date(date.year, date.month - 1, date.day, hh || 0, mm || 0, 0, 0);
  return d.toISOString();
}

const startMaxValue = computed(() => isoToCalendarDate(task.value?.due_date));
const dueMinValue = computed(() => isoToCalendarDate(task.value?.start_date));

watch(startDateOpen, (open) => {
  if (!open) return;
  startDateDraft.value = isoToCalendarDate(task.value?.start_date);
  startTimeDraft.value = isoToTimeInput(task.value?.start_date, "09:00");
});
watch(dueDateOpen, (open) => {
  if (!open) return;
  dueDateDraft.value = isoToCalendarDate(task.value?.due_date);
  dueTimeDraft.value = isoToTimeInput(task.value?.due_date, "17:00");
});

async function updateDate(field: "start_date" | "due_date", value: string | null) {
  const label = field === "start_date" ? "Start date" : "Due date";
  await patchTask({ [field]: value }, value === null ? `${label} cleared` : `${label} updated`);
}
async function saveStartDate() {
  if (!startDateDraft.value) return;
  const iso = combineDateTime(startDateDraft.value, startTimeDraft.value);
  const due = task.value?.due_date;
  if (due && new Date(iso) > new Date(due)) {
    toast.error("Start date cannot be after due date");
    return;
  }
  startDateOpen.value = false;
  await updateDate("start_date", iso);
}
async function saveDueDate() {
  if (!dueDateDraft.value) return;
  const iso = combineDateTime(dueDateDraft.value, dueTimeDraft.value);
  const start = task.value?.start_date;
  if (start && new Date(iso) < new Date(start)) {
    toast.error("Due date cannot be before start date");
    return;
  }
  dueDateOpen.value = false;
  await updateDate("due_date", iso);
}
async function clearStartDate() {
  startDateOpen.value = false;
  await updateDate("start_date", null);
}
async function clearDueDate() {
  dueDateOpen.value = false;
  await updateDate("due_date", null);
}

// Assignees / labels mutate server-side via their own child components; reload
// the local copy and notify the parent so both stay in sync.
async function onRelationChanged() {
  await loadTask();
  emit("updated");
}

// --- Comments ---
// Managed locally (direct fetch), NOT via useComments(): that composable holds a
// module-level shared comment list bound to the parent task's activity feed, so
// loading/pushing the subtask's comments there would clobber the parent's feed.
const comments = ref<Comment[]>([]);
const commentsLoading = ref(false);
const commentContent = ref("");
const submittingComment = ref(false);

// Sort preference is shared with the task page's activity feed for consistency.
const SORT_KEY = "bureaucat-activity-sort";
const newestFirst = ref(
  typeof localStorage !== "undefined" ? localStorage.getItem(SORT_KEY) !== "oldest" : true
);

function toggleCommentSort() {
  newestFirst.value = !newestFirst.value;
  if (typeof localStorage !== "undefined") {
    localStorage.setItem(SORT_KEY, newestFirst.value ? "newest" : "oldest");
  }
}

const sortedComments = computed(() =>
  [...comments.value].sort((a, b) => {
    const diff = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
    return newestFirst.value ? -diff : diff;
  })
);

const commentsEndpoint = computed(
  () => `/api/v1/projects/${props.projectKey}/tasks/${props.taskNumber}/comments`
);

const isCommentEmpty = computed(() => {
  const text = commentContent.value.replace(/<[^>]*>/g, "").replace(/&nbsp;/g, " ").trim();
  return text.length === 0;
});

function canEditComment(comment: Comment): boolean {
  return props.isMember && comment.created_by === user.value?.id;
}

async function loadComments() {
  commentsLoading.value = true;
  try {
    const res = await fetch(commentsEndpoint.value, { headers: getAuthHeader() });
    if (res.ok) comments.value = await res.json();
  } catch {
    // Non-fatal: the rest of the card still renders.
  } finally {
    commentsLoading.value = false;
  }
}

async function submitComment() {
  if (isCommentEmpty.value) return;
  const trimmed = trimHtmlContent(commentContent.value);
  submittingComment.value = true;
  try {
    const res = await fetch(commentsEndpoint.value, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...getAuthHeader() },
      body: JSON.stringify({ content: trimmed }),
    });
    if (res.ok) {
      commentContent.value = "";
      await loadComments();
    } else {
      const err = await res.json().catch(() => ({}));
      toast.error(err.message || "Failed to add comment");
    }
  } catch {
    toast.error("Network error");
  } finally {
    submittingComment.value = false;
  }
}

function handleCommentKeydown(event: KeyboardEvent) {
  if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
    event.preventDefault();
    submitComment();
  }
}

onMounted(() => {
  loadTask();
  loadComments();
});
</script>

<template>
  <div>
    <div v-if="loading" class="flex items-center justify-center py-12">
      <Loader2 class="size-5 animate-spin text-muted-foreground" />
    </div>

    <div v-else-if="error" class="py-12 text-center text-sm text-destructive">
      {{ error }}
    </div>

    <template v-else-if="task">
      <div class="mb-4 flex items-start justify-between gap-3">
        <div class="min-w-0 flex-1 space-y-0.5">
          <span class="font-mono text-xs text-muted-foreground">{{ task.task_id }}</span>
          <!-- Title (editable) -->
          <div v-if="editingTitle" class="space-y-2">
            <Input
              v-model="editTitle"
              class="text-base font-semibold"
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
              <Button size="sm" variant="outline" :disabled="updating" @click="cancelEditTitle">
                <X class="mr-1.5 size-3" />
                Cancel
              </Button>
            </div>
          </div>
          <div v-else class="group flex items-start gap-1.5">
            <h2 class="text-lg font-semibold leading-tight">{{ task.title }}</h2>
            <Button
              v-if="isMember"
              variant="ghost"
              size="icon"
              aria-label="Edit title"
              class="mt-0.5 size-6 shrink-0 opacity-0 transition-opacity group-hover:opacity-100 focus:opacity-100"
              @click="startEditTitle"
            >
              <Pencil class="size-3.5" />
            </Button>
          </div>
        </div>
        <Button variant="outline" size="sm" class="shrink-0 gap-1.5" as-child>
          <NuxtLink :to="taskLink">
            <Maximize2 class="size-3.5" />
            Open
          </NuxtLink>
        </Button>
      </div>

      <div class="space-y-4">
        <!-- Creator / created meta -->
        <div class="flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground">
          <NuxtLink
            :to="`/profile/${task.created_by}`"
            class="flex items-center gap-1 rounded-md border bg-muted/50 py-0.5 pl-0.5 pr-1.5 hover:bg-muted transition-colors"
          >
            <Avatar class="size-4">
              <AvatarImage v-if="task.creator_avatar_url" :src="task.creator_avatar_url" />
              <AvatarFallback class="text-[10px]" :seed="task.created_by">
                {{ task.creator_first_name?.[0] }}{{ task.creator_last_name?.[0] }}
              </AvatarFallback>
            </Avatar>
            <span>{{ task.creator_first_name }} {{ task.creator_last_name }}</span>
          </NuxtLink>
          <span>created on {{ formatDate(task.created_at) }}</span>
          <span class="flex items-center gap-1">
            <Clock class="size-3.5" />
            updated {{ formatDate(task.updated_at) }}
          </span>
        </div>

        <!-- Description (editable) -->
        <div class="group">
          <div class="mb-1.5 flex items-center justify-between gap-2">
            <h3 class="text-sm font-medium text-muted-foreground">Description</h3>
            <Button
              v-if="isMember && !editingDescription && task.description"
              variant="ghost"
              size="icon"
              aria-label="Edit description"
              class="size-6 opacity-0 transition-opacity group-hover:opacity-100 focus:opacity-100"
              @click="startEditDescription"
            >
              <Pencil class="size-3.5" />
            </Button>
          </div>
          <div v-if="editingDescription" class="space-y-2">
            <TiptapEditor
              v-model="editDescription"
              :disabled="updating"
              :members="members"
            />
            <div class="flex gap-2">
              <Button size="sm" :disabled="updating" @click="saveDescription">
                <Loader2 v-if="updating" class="mr-1.5 size-3 animate-spin" />
                <Check v-else class="mr-1.5 size-3" />
                Save
              </Button>
              <Button size="sm" variant="outline" :disabled="updating" @click="cancelEditDescription">
                <X class="mr-1.5 size-3" />
                Cancel
              </Button>
            </div>
          </div>
          <template v-else>
            <div
              v-if="task.description"
              class="prose prose-sm max-w-none dark:prose-invert"
              v-html="renderedDescription"
            />
            <button
              v-else-if="isMember"
              type="button"
              class="w-full rounded-lg border border-dashed p-3 text-left text-sm text-muted-foreground hover:border-solid hover:bg-muted/50"
              @click="startEditDescription"
            >
              Add a description...
            </button>
            <p v-else class="text-sm italic text-muted-foreground">No description</p>
          </template>
        </div>

        <!-- Properties grid -->
        <div class="grid gap-x-6 gap-y-3 rounded-lg border bg-muted/20 p-3 sm:grid-cols-2">
          <!-- State -->
          <div class="flex items-center justify-between gap-2">
            <p class="text-xs text-muted-foreground">State</p>
            <TaskStateSelector
              :states="states"
              :model-value="task.state_id"
              :disabled="!isMember || updating || states.length === 0"
              compact
              dense
              @update:model-value="handleStateChange"
            />
          </div>

          <!-- Priority -->
          <div class="flex items-center justify-between gap-2">
            <p class="text-xs text-muted-foreground">Priority</p>
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button
                  variant="ghost"
                  class="h-auto gap-1.5 px-0 py-0 text-xs font-medium hover:bg-transparent"
                  :disabled="!isMember || updating"
                >
                  <span
                    class="size-2.5 rounded-full ring-1.5 ring-offset-1 ring-offset-background"
                    :style="{ backgroundColor: priority.color, '--tw-ring-color': priority.color }"
                  />
                  {{ priority.label }}
                  <ChevronDown class="size-3 opacity-50" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-40">
                <DropdownMenuItem
                  v-for="p in priorityOptions"
                  :key="p.value"
                  @click="handlePriorityChange(p.value)"
                >
                  <span class="mr-2 size-2 rounded-full" :style="{ backgroundColor: p.color }" />
                  {{ p.label }}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          <!-- Start date -->
          <div class="flex items-center justify-between gap-2">
            <p class="text-xs text-muted-foreground">Start date</p>
            <Popover v-model:open="startDateOpen">
              <PopoverTrigger as-child>
                <Button
                  variant="ghost"
                  class="h-auto gap-1.5 px-0 py-0 text-xs font-medium hover:bg-transparent"
                  :class="!task.start_date ? 'text-muted-foreground' : ''"
                  :disabled="!isMember || updating"
                >
                  <CalendarIcon class="size-3.5 opacity-70" />
                  <span>{{ task.start_date ? formatDateTime(task.start_date) : "Set date" }}</span>
                  <ChevronDown class="size-3 opacity-50" />
                </Button>
              </PopoverTrigger>
              <PopoverContent class="w-auto p-0" align="end">
                <Calendar v-model="startDateDraft" layout="month-and-year" :max-value="startMaxValue" />
                <div class="flex items-center gap-2 border-t px-3 py-2">
                  <CalendarIcon class="size-3.5 text-muted-foreground" />
                  <Input v-model="startTimeDraft" type="time" class="h-8 flex-1 text-sm" />
                  <Button size="sm" :disabled="!startDateDraft || updating" @click="saveStartDate">
                    Save
                  </Button>
                </div>
                <div v-if="task.start_date" class="border-t px-3 py-2">
                  <Button variant="ghost" size="sm" class="w-full" :disabled="updating" @click="clearStartDate">
                    <X class="mr-1.5 size-3.5" />
                    Clear
                  </Button>
                </div>
              </PopoverContent>
            </Popover>
          </div>

          <!-- Due date -->
          <div class="flex items-center justify-between gap-2">
            <p class="text-xs text-muted-foreground">Due date</p>
            <Popover v-model:open="dueDateOpen">
              <PopoverTrigger as-child>
                <Button
                  variant="ghost"
                  class="h-auto gap-1.5 px-0 py-0 text-xs font-medium hover:bg-transparent"
                  :class="!task.due_date ? 'text-muted-foreground' : ''"
                  :disabled="!isMember || updating"
                >
                  <CalendarIcon class="size-3.5 opacity-70" />
                  <span>{{ task.due_date ? formatDateTime(task.due_date) : "Set date" }}</span>
                  <ChevronDown class="size-3 opacity-50" />
                </Button>
              </PopoverTrigger>
              <PopoverContent class="w-auto p-0" align="end">
                <Calendar v-model="dueDateDraft" layout="month-and-year" :min-value="dueMinValue" />
                <div class="flex items-center gap-2 border-t px-3 py-2">
                  <CalendarIcon class="size-3.5 text-muted-foreground" />
                  <Input v-model="dueTimeDraft" type="time" class="h-8 flex-1 text-sm" />
                  <Button size="sm" :disabled="!dueDateDraft || updating" @click="saveDueDate">
                    Save
                  </Button>
                </div>
                <div v-if="task.due_date" class="border-t px-3 py-2">
                  <Button variant="ghost" size="sm" class="w-full" :disabled="updating" @click="clearDueDate">
                    <X class="mr-1.5 size-3.5" />
                    Clear
                  </Button>
                </div>
              </PopoverContent>
            </Popover>
          </div>
        </div>

        <!-- Assignees (editable) -->
        <TaskAssignees
          :assignees="task.assignees || []"
          :project-key="projectKey"
          :task-num="taskNumber"
          :members="members"
          :is-member="isMember"
          @refresh="onRelationChanged"
        />

        <!-- Labels (editable) -->
        <TaskLabels
          :task-labels="task.labels || []"
          :project-key="projectKey"
          :task-num="taskNumber"
          :project-labels="projectLabels"
          :is-member="isMember"
          @refresh="onRelationChanged"
        />

        <Separator />

        <!-- Comments -->
        <div class="space-y-4">
          <div class="flex items-center justify-between gap-2">
            <h3 class="flex items-center gap-2 text-sm font-semibold">
              <MessageSquare class="size-4" />
              Comments
              <span class="font-normal text-muted-foreground">({{ comments.length }})</span>
            </h3>
            <Button
              v-if="comments.length"
              variant="ghost"
              size="sm"
              @click="toggleCommentSort"
            >
              <ArrowUpDown class="mr-1.5 size-3.5" />
              {{ newestFirst ? "Newest first" : "Oldest first" }}
            </Button>
          </div>

          <div v-if="commentsLoading" class="flex items-center justify-center py-6">
            <Loader2 class="size-5 animate-spin text-muted-foreground" />
          </div>

          <div v-else-if="comments.length" class="space-y-4">
            <CommentItem
              v-for="comment in sortedComments"
              :key="comment.id"
              :comment="comment"
              :project-key="projectKey"
              :task-num="taskNumber"
              :can-edit="canEditComment(comment)"
              :members="members"
              @deleted="loadComments"
              @updated="loadComments"
            />
          </div>

          <p v-else class="rounded-lg border border-dashed py-6 text-center text-sm text-muted-foreground">
            No comments yet
          </p>

          <!-- Add comment -->
          <div v-if="isMember" class="flex gap-3">
            <Avatar class="size-8 shrink-0">
              <AvatarImage v-if="user?.avatar_url" :src="user.avatar_url" />
              <AvatarFallback class="text-xs" :seed="user?.id">
                {{ user?.first_name?.[0] }}{{ user?.last_name?.[0] }}
              </AvatarFallback>
            </Avatar>
            <form class="flex-1 space-y-2" @submit.prevent="submitComment" @keydown="handleCommentKeydown">
              <TiptapEditor
                v-model="commentContent"
                :disabled="submittingComment"
                :members="members"
                compact
              />
              <div class="flex justify-end">
                <Button
                  type="submit"
                  size="sm"
                  :disabled="submittingComment || isCommentEmpty"
                >
                  <Loader2 v-if="submittingComment" class="mr-1.5 size-3.5 animate-spin" />
                  <Send v-else class="mr-1.5 size-3.5" />
                  Comment
                </Button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

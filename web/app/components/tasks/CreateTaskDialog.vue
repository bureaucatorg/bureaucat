<script setup lang="ts">
import { Loader2, Check, ChevronsUpDown } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type {
  Project,
  ProjectState,
  ProjectLabel,
  ProjectMember,
  TaskTemplate,
} from "~/types";
import { mdToHtml } from "~/utils/markdown";

const props = withDefaults(
  defineProps<{
    // Locked-project mode (e.g. opened from /projects/[key]): pass the key plus
    // its metadata and the project selector stays hidden.
    projectKey?: string;
    states?: ProjectState[];
    labels?: ProjectLabel[];
    members?: ProjectMember[];
    templates?: TaskTemplate[];
    // Selector mode (e.g. opened from /dashboard or Shift+C): the dialog fetches
    // the workspace-scoped project list itself and shows a project picker. It
    // also fetches the chosen project's metadata on selection.
    projectSelector?: boolean;
    // Selector mode only: a project key to pre-select when the dialog opens
    // (e.g. the current project when Shift+C is pressed on /projects/[key]).
    initialProjectKey?: string;
    // Selector mode only: when true, the project picker lists projects across
    // every workspace instead of only the active one (mirrors the dashboard's
    // "All workspaces" toggle).
    allWorkspaces?: boolean;
    // Subtask mode (opened from a parent task's detail page): the created task
    // becomes a child of this (project-local) parent number. The dialog stays on
    // the current page instead of navigating to the new task.
    parentTaskNumber?: number;
  }>(),
  {
    states: () => [],
    labels: () => [],
    members: () => [],
    templates: () => [],
  }
);

const open = defineModel<boolean>("open", { default: false });

const emit = defineEmits<{
  created: [];
}>();

const { getAuthHeader } = useAuth();
const { currentWorkspace } = useWorkspaces();
const { createTask } = useTasks();
const { listStates, listLabels, listMembers, listTemplates } = useProjects();

// --- Project selection ---
// Selector mode is active when the caller opts into the project picker; the
// dialog then owns fetching the (workspace-scoped) list of projects to choose
// from.
const selectable = computed(() => props.projectSelector === true);
const selectedProjectKey = ref("");
const showProjectPopover = ref(false);
const metaLoading = ref(false);

// The workspace-scoped project list for the picker, fetched on open. Kept
// local (not the shared useProjects store) so it never clobbers a caller's
// own project grid.
const availableProjects = ref<Project[]>([]);

async function fetchProjects() {
  try {
    let url = "/api/v1/projects?page=1&per_page=100";
    if (!props.allWorkspaces && currentWorkspace.value) {
      url += `&workspace_id=${currentWorkspace.value.id}`;
    }
    const res = await fetch(url, { headers: getAuthHeader() });
    if (res.ok) {
      const data = await res.json();
      availableProjects.value = data.projects || [];
    }
  } catch {
    // silently fail — the picker just shows no projects to choose from
  }
}

// Metadata fetched on-demand when a project is chosen (selector mode only).
const fetchedStates = ref<ProjectState[]>([]);
const fetchedLabels = ref<ProjectLabel[]>([]);
const fetchedMembers = ref<ProjectMember[]>([]);
const fetchedTemplates = ref<TaskTemplate[]>([]);

const effectiveProjectKey = computed(() =>
  selectable.value ? selectedProjectKey.value : props.projectKey ?? ""
);
const effStates = computed(() => (selectable.value ? fetchedStates.value : props.states));
const effLabels = computed(() => (selectable.value ? fetchedLabels.value : props.labels));
const effMembers = computed(() => (selectable.value ? fetchedMembers.value : props.members));
const effTemplates = computed(() =>
  selectable.value ? fetchedTemplates.value : props.templates
);

const selectedProject = computed(
  () => availableProjects.value.find((p) => p.project_key === selectedProjectKey.value) ?? null
);

function projectSearchText(p: Project) {
  return `${p.name} ${p.project_key}`;
}

const loading = ref(false);
const error = ref<string | null>(null);
const selectedTemplateId = ref("");
const form = ref({
  title: "",
  description: "",
  state_id: "",
  priority: 0,
  assignees: [] as string[],
  labels: [] as string[],
});

const defaultState = computed(() => effStates.value.find((s) => s.is_default));

function resetForm() {
  form.value = {
    title: "",
    description: "",
    state_id: defaultState.value?.id || "",
    priority: 0,
    assignees: [],
    labels: [],
  };
  selectedTemplateId.value = "";
  error.value = null;
}

async function loadProjectMeta(key: string) {
  metaLoading.value = true;
  const [s, l, m, t] = await Promise.all([
    listStates(key),
    listLabels(key),
    listMembers(key),
    listTemplates(key),
  ]);
  fetchedStates.value = s.data ?? [];
  fetchedLabels.value = l.data ?? [];
  fetchedMembers.value = m.data ?? [];
  fetchedTemplates.value = t.data ?? [];
  metaLoading.value = false;
  // Reset project-dependent fields now that metadata is available.
  form.value.state_id = defaultState.value?.id || "";
  form.value.assignees = [];
  form.value.labels = [];
  selectedTemplateId.value = "";
}

function selectProject(key: string) {
  selectedProjectKey.value = key;
  showProjectPopover.value = false;
  loadProjectMeta(key);
}

watch(selectedTemplateId, (id) => {
  if (!id) return;
  const tmpl = effTemplates.value.find((t) => t.id === id);
  if (tmpl) {
    form.value.title = tmpl.title;
    // Templates are stored as HTML; convert any legacy markdown ones so the
    // Tiptap editor receives valid HTML rather than raw markdown text.
    form.value.description = mdToHtml(tmpl.description);
  }
});

watch(open, async (isOpen) => {
  if (isOpen) {
    if (selectable.value) {
      selectedProjectKey.value = "";
      fetchedStates.value = [];
      fetchedLabels.value = [];
      fetchedMembers.value = [];
      fetchedTemplates.value = [];
      await fetchProjects();
      // If the caller supplied a project (e.g. Shift+C on /projects/[key]) and
      // it's in the fetched list, pre-select it and skip the picker. Otherwise
      // drop the cursor straight into the project picker so the user can start
      // typing to search immediately. Deferred so it happens after the dialog's
      // own open/focus handling has settled.
      const preselect =
        props.initialProjectKey &&
        availableProjects.value.some((p) => p.project_key === props.initialProjectKey)
          ? props.initialProjectKey
          : "";
      nextTick(() => {
        if (preselect) {
          selectProject(preselect);
        } else {
          showProjectPopover.value = true;
        }
      });
    }
    resetForm();
  }
});

async function handleSubmit() {
  if (!effectiveProjectKey.value) {
    error.value = "Please select a project";
    return;
  }

  loading.value = true;
  error.value = null;

  const result = await createTask(effectiveProjectKey.value, {
    title: form.value.title,
    description: form.value.description || undefined,
    state_id: form.value.state_id || undefined,
    priority: form.value.priority,
    assignees: form.value.assignees.length > 0 ? form.value.assignees : undefined,
    labels: form.value.labels.length > 0 ? form.value.labels : undefined,
    parent_task_number: props.parentTaskNumber,
  });

  loading.value = false;

  if (result.success && result.data) {
    open.value = false;
    emit("created");
    // Subtask mode stays on the parent's page; standalone create navigates to
    // the new task.
    if (props.parentTaskNumber != null) {
      toast.success(`Subtask ${result.data.task_id} created`);
    } else {
      toast.success(`Task ${result.data.task_id} created`);
      await navigateTo(
        `/projects/${result.data.project_key}/tasks/${result.data.task_number}`
      );
    }
  } else {
    error.value = result.error || "Failed to create task";
  }
}

const isSubtaskMode = computed(() => props.parentTaskNumber != null);

const priorities = [
  { value: 0, label: "No priority" },
  { value: 1, label: "Low" },
  { value: 2, label: "Medium" },
  { value: 3, label: "High" },
  { value: 4, label: "Urgent" },
];

// shadcn/reka-ui Select works with string values only, and reserves the empty
// string for "no selection" — so these adapters bridge to the form's types.
const NO_TEMPLATE = "__none__";
const templateValue = computed({
  get: () => selectedTemplateId.value || NO_TEMPLATE,
  set: (v: string) => {
    selectedTemplateId.value = v === NO_TEMPLATE ? "" : v;
  },
});
const priorityValue = computed({
  get: () => String(form.value.priority),
  set: (v: string) => {
    form.value.priority = Number(v);
  },
});

const selectedAssignees = computed(() =>
  effMembers.value.filter((m) => form.value.assignees.includes(m.user_id))
);

// Members not yet picked — the pool offered in the assignee popover.
const availableAssignees = computed(() => {
  const selected = new Set(form.value.assignees);
  return effMembers.value.filter((m) => !selected.has(m.user_id));
});

const selectedLabels = computed(() =>
  effLabels.value.filter((l) => form.value.labels.includes(l.id))
);

// Labels not yet picked — the pool offered in the label popover.
const availableLabels = computed(() => {
  const selected = new Set(form.value.labels);
  return effLabels.value.filter((l) => !selected.has(l.id));
});

function memberSearchText(m: ProjectMember) {
  return `${m.first_name} ${m.last_name} ${m.username}`;
}

function addAssignee(userId: string) {
  if (!form.value.assignees.includes(userId)) {
    form.value.assignees.push(userId);
  }
}

function removeAssignee(userId: string) {
  form.value.assignees = form.value.assignees.filter((id) => id !== userId);
}

function addLabel(labelId: string) {
  if (!form.value.labels.includes(labelId)) {
    form.value.labels.push(labelId);
  }
}

function removeLabel(labelId: string) {
  form.value.labels = form.value.labels.filter((id) => id !== labelId);
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="sm:max-w-4xl max-h-[85vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>{{ isSubtaskMode ? "Create Subtask" : "Create New Task" }}</DialogTitle>
        <DialogDescription>
          <template v-if="isSubtaskMode">
            Add a subtask under {{ projectKey }}-{{ parentTaskNumber }}
          </template>
          <template v-else-if="selectable">
            {{ selectedProject ? `Add a new task to ${selectedProject.name}` : "Select a project to add a task to" }}
          </template>
          <template v-else>
            Add a new task to {{ projectKey }}
          </template>
        </DialogDescription>
      </DialogHeader>
      <form class="space-y-4" @submit.prevent="handleSubmit">
        <div
          v-if="error"
          class="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
        >
          {{ error }}
        </div>

        <!-- Project selector (selector mode only) -->
        <div v-if="selectable" class="space-y-2">
          <Label>Project</Label>
          <SearchableSelect
            v-model:open="showProjectPopover"
            :items="availableProjects"
            :get-search-text="projectSearchText"
            :get-key="(p) => p.id"
            placeholder="Search projects..."
            empty-text="No projects found"
            content-class="w-[var(--reka-popover-trigger-width)]"
            @select="(p) => selectProject(p.project_key)"
          >
            <template #trigger>
              <Button
                type="button"
                variant="outline"
                role="combobox"
                :disabled="loading"
                class="w-full justify-between font-normal"
              >
                <span class="flex min-w-0 items-center gap-2">
                  <template v-if="selectedProject">
                    <span class="truncate">{{ selectedProject.name }}</span>
                    <span class="shrink-0 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
                      {{ selectedProject.project_key }}
                    </span>
                  </template>
                  <span v-else class="text-muted-foreground">Select a project...</span>
                </span>
                <ChevronsUpDown class="size-4 shrink-0 opacity-50" />
              </Button>
            </template>
            <template #option="{ item: project }">
              <Check
                class="size-3.5 shrink-0 text-primary"
                :class="selectedProjectKey === project.project_key ? 'opacity-100' : 'opacity-0'"
              />
              <span class="min-w-0 flex-1 truncate">{{ project.name }}</span>
              <span class="shrink-0 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
                {{ project.project_key }}
              </span>
            </template>
          </SearchableSelect>
        </div>

        <!-- Loading project metadata -->
        <div v-if="metaLoading" class="flex items-center gap-2 py-2 text-sm text-muted-foreground">
          <Loader2 class="size-4 animate-spin" />
          Loading project details...
        </div>

        <!-- The rest of the form requires a project in selector mode. -->
        <template v-if="!selectable || (selectedProject && !metaLoading)">
          <div v-if="effTemplates.length > 0" class="space-y-2">
            <Label for="template">Template</Label>
            <Select v-model="templateValue" :disabled="loading">
              <SelectTrigger id="template" class="w-full">
                <SelectValue placeholder="No template" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem :value="NO_TEMPLATE">No template</SelectItem>
                <SelectItem v-for="tmpl in effTemplates" :key="tmpl.id" :value="tmpl.id">
                  {{ tmpl.name }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="space-y-2">
            <Label for="title">Title</Label>
            <Input
              id="title"
              v-model="form.title"
              placeholder="Task title"
              required
              :disabled="loading"
            />
          </div>

          <div class="space-y-2">
            <Label>Description</Label>
            <TiptapEditor
              v-model="form.description"
              :disabled="loading"
              :members="effMembers"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="state">State</Label>
              <Select v-model="form.state_id" :disabled="loading">
                <SelectTrigger id="state" class="w-full">
                  <SelectValue placeholder="Select a state" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="state in effStates" :key="state.id" :value="state.id">
                    {{ state.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div class="space-y-2">
              <Label for="priority">Priority</Label>
              <Select v-model="priorityValue" :disabled="loading">
                <SelectTrigger id="priority" class="w-full">
                  <SelectValue placeholder="Select priority" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="p in priorities" :key="p.value" :value="String(p.value)">
                    {{ p.label }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div v-if="effMembers.length > 0" class="space-y-2">
            <Label>Assignees</Label>
            <TokenSelect
              :selected="selectedAssignees"
              :available="availableAssignees"
              :get-key="(m) => m.user_id"
              :get-search-text="memberSearchText"
              :disabled="loading"
              placeholder="Add assignees..."
              empty-text="No members found"
              @add="(m) => addAssignee(m.user_id)"
              @remove="(m) => removeAssignee(m.user_id)"
            >
              <template #chip="{ item: member }">
                <Avatar class="size-5">
                  <AvatarFallback class="text-[10px]" :seed="member.user_id">
                    {{ member.first_name[0] }}{{ member.last_name[0] }}
                  </AvatarFallback>
                </Avatar>
                <span class="truncate">{{ member.first_name }} {{ member.last_name }}</span>
              </template>
              <template #option="{ item: member }">
                <Avatar class="size-6">
                  <AvatarFallback class="text-xs" :seed="member.user_id">
                    {{ member.first_name[0] }}{{ member.last_name[0] }}
                  </AvatarFallback>
                </Avatar>
                {{ member.first_name }} {{ member.last_name }}
              </template>
            </TokenSelect>
          </div>

          <div v-if="effLabels.length > 0" class="space-y-2">
            <Label>Labels</Label>
            <TokenSelect
              :selected="selectedLabels"
              :available="availableLabels"
              :get-key="(l) => l.id"
              :get-search-text="(l) => l.name"
              :chip-style="(l) => ({ backgroundColor: l.color + '20', color: l.color })"
              :chip-class="() => 'pl-2 pr-1 font-medium'"
              :disabled="loading"
              placeholder="Add labels..."
              empty-text="No labels found"
              @add="(l) => addLabel(l.id)"
              @remove="(l) => removeLabel(l.id)"
            >
              <template #chip="{ item: label }">
                <span class="truncate">{{ label.name }}</span>
              </template>
              <template #option="{ item: label }">
                <div
                  class="size-3 shrink-0 rounded-full"
                  :style="{ backgroundColor: label.color }"
                />
                {{ label.name }}
              </template>
            </TokenSelect>
          </div>
        </template>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            :disabled="loading"
            @click="open = false"
          >
            Cancel
          </Button>
          <Button type="submit" :disabled="loading || !form.title || !effectiveProjectKey">
            <Loader2 v-if="loading" class="mr-2 size-4 animate-spin" />
            {{ isSubtaskMode ? "Create Subtask" : "Create Task" }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

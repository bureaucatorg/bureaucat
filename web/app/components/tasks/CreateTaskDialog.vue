<script setup lang="ts">
import { Loader2, Plus, X, Search, Check, ChevronsUpDown } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type {
  Project,
  ProjectState,
  ProjectLabel,
  ProjectMember,
  TaskTemplate,
} from "~/types";

const props = withDefaults(
  defineProps<{
    // Locked-project mode (e.g. opened from /projects/[key]): pass the key plus
    // its metadata and the project selector stays hidden.
    projectKey?: string;
    states?: ProjectState[];
    labels?: ProjectLabel[];
    members?: ProjectMember[];
    templates?: TaskTemplate[];
    // Selector mode (e.g. opened from /dashboard): pass the list of projects to
    // choose from. The dialog fetches the chosen project's metadata itself.
    projects?: Project[];
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

const { createTask } = useTasks();
const { listStates, listLabels, listMembers, listTemplates } = useProjects();

// --- Project selection ---
// Selector mode is active whenever a `projects` list is supplied.
const selectable = computed(() => Array.isArray(props.projects));
const selectedProjectKey = ref("");
const showProjectPopover = ref(false);
const projectSearch = ref("");
const metaLoading = ref(false);

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
  () => props.projects?.find((p) => p.project_key === selectedProjectKey.value) ?? null
);

const filteredProjects = computed(() => {
  const list = props.projects ?? [];
  const q = projectSearch.value.toLowerCase().trim();
  if (!q) return list;
  return list.filter(
    (p) =>
      p.name.toLowerCase().includes(q) ||
      p.project_key.toLowerCase().includes(q)
  );
});

// Keyboard navigation for the project list.
const highlightedProjectIndex = ref(0);
const projectListRef = ref<HTMLElement | null>(null);

watch(showProjectPopover, (isOpen) => {
  if (!isOpen) {
    projectSearch.value = "";
  } else {
    highlightedProjectIndex.value = 0;
  }
});

// Reset/clamp the highlight as the filtered list changes (e.g. while typing).
watch(filteredProjects, (list) => {
  if (highlightedProjectIndex.value >= list.length) {
    highlightedProjectIndex.value = Math.max(0, list.length - 1);
  }
});

function scrollHighlightedProjectIntoView() {
  nextTick(() => {
    const el = projectListRef.value?.querySelector<HTMLElement>(
      `[data-index="${highlightedProjectIndex.value}"]`
    );
    el?.scrollIntoView({ block: "nearest" });
  });
}

function handleProjectKeydown(event: KeyboardEvent) {
  const count = filteredProjects.value.length;
  if (event.key === "ArrowDown") {
    event.preventDefault();
    if (count === 0) return;
    highlightedProjectIndex.value = (highlightedProjectIndex.value + 1) % count;
    scrollHighlightedProjectIntoView();
  } else if (event.key === "ArrowUp") {
    event.preventDefault();
    if (count === 0) return;
    highlightedProjectIndex.value =
      (highlightedProjectIndex.value - 1 + count) % count;
    scrollHighlightedProjectIntoView();
  } else if (event.key === "Enter") {
    event.preventDefault();
    const project = filteredProjects.value[highlightedProjectIndex.value];
    if (project) selectProject(project.project_key);
  } else if (event.key === "Escape") {
    showProjectPopover.value = false;
  }
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
    form.value.description = tmpl.description;
  }
});

watch(open, (isOpen) => {
  if (isOpen) {
    if (selectable.value) {
      selectedProjectKey.value = "";
      fetchedStates.value = [];
      fetchedLabels.value = [];
      fetchedMembers.value = [];
      fetchedTemplates.value = [];
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
  });

  loading.value = false;

  if (result.success && result.data) {
    toast.success(`Task ${result.data.task_id} created`);
    open.value = false;
    emit("created");
    await navigateTo(
      `/projects/${result.data.project_key}/tasks/${result.data.task_number}`
    );
  } else {
    error.value = result.error || "Failed to create task";
  }
}

const priorities = [
  { value: 0, label: "No priority" },
  { value: 1, label: "Low" },
  { value: 2, label: "Medium" },
  { value: 3, label: "High" },
  { value: 4, label: "Urgent" },
];

const assigneeSearch = ref("");
const labelSearch = ref("");
const showAssigneePopover = ref(false);
const showLabelPopover = ref(false);

const selectedAssignees = computed(() =>
  effMembers.value.filter((m) => form.value.assignees.includes(m.user_id))
);

const filteredAssigneeOptions = computed(() => {
  const selected = new Set(form.value.assignees);
  const available = effMembers.value.filter((m) => !selected.has(m.user_id));
  const q = assigneeSearch.value.toLowerCase().trim();
  if (!q) return available;
  return available.filter(
    (m) =>
      m.first_name.toLowerCase().includes(q) ||
      m.last_name.toLowerCase().includes(q) ||
      m.username.toLowerCase().includes(q)
  );
});

const selectedLabels = computed(() =>
  effLabels.value.filter((l) => form.value.labels.includes(l.id))
);

const filteredLabelOptions = computed(() => {
  const selected = new Set(form.value.labels);
  const available = effLabels.value.filter((l) => !selected.has(l.id));
  const q = labelSearch.value.toLowerCase().trim();
  if (!q) return available;
  return available.filter((l) => l.name.toLowerCase().includes(q));
});

watch(showAssigneePopover, (open) => {
  if (!open) assigneeSearch.value = "";
});

watch(showLabelPopover, (open) => {
  if (!open) labelSearch.value = "";
});

function addAssignee(userId: string) {
  if (!form.value.assignees.includes(userId)) {
    form.value.assignees.push(userId);
  }
  showAssigneePopover.value = false;
}

function removeAssignee(userId: string) {
  form.value.assignees = form.value.assignees.filter((id) => id !== userId);
}

function addLabel(labelId: string) {
  if (!form.value.labels.includes(labelId)) {
    form.value.labels.push(labelId);
  }
  showLabelPopover.value = false;
}

function removeLabel(labelId: string) {
  form.value.labels = form.value.labels.filter((id) => id !== labelId);
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="sm:max-w-4xl max-h-[85vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>Create New Task</DialogTitle>
        <DialogDescription>
          <template v-if="selectable">
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
          <Popover v-model:open="showProjectPopover">
            <PopoverTrigger as-child>
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
            </PopoverTrigger>
            <PopoverContent align="start" class="w-[var(--reka-popover-trigger-width)] p-0">
              <div class="border-b px-3 py-2">
                <div class="relative">
                  <Search class="absolute left-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    v-model="projectSearch"
                    placeholder="Search projects..."
                    class="h-8 pl-7 text-sm"
                    autofocus
                    @keydown="handleProjectKeydown"
                  />
                </div>
              </div>
              <div ref="projectListRef" class="max-h-60 overflow-y-auto p-1">
                <button
                  v-for="(project, idx) in filteredProjects"
                  :key="project.id"
                  type="button"
                  :data-index="idx"
                  class="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-sm"
                  :class="idx === highlightedProjectIndex ? 'bg-accent' : 'hover:bg-accent'"
                  @click="selectProject(project.project_key)"
                  @mouseenter="highlightedProjectIndex = idx"
                >
                  <Check
                    class="size-3.5 shrink-0 text-primary"
                    :class="selectedProjectKey === project.project_key ? 'opacity-100' : 'opacity-0'"
                  />
                  <span class="min-w-0 flex-1 truncate">{{ project.name }}</span>
                  <span class="shrink-0 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
                    {{ project.project_key }}
                  </span>
                </button>
                <p
                  v-if="filteredProjects.length === 0"
                  class="px-3 py-6 text-center text-sm text-muted-foreground"
                >
                  No projects found
                </p>
              </div>
            </PopoverContent>
          </Popover>
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
            <NativeSelect id="template" v-model="selectedTemplateId" :disabled="loading">
              <option value="">No template</option>
              <option v-for="tmpl in effTemplates" :key="tmpl.id" :value="tmpl.id">
                {{ tmpl.name }}
              </option>
            </NativeSelect>
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
              <NativeSelect id="state" v-model="form.state_id" :disabled="loading">
                <option v-for="state in effStates" :key="state.id" :value="state.id">
                  {{ state.name }}
                </option>
              </NativeSelect>
            </div>

            <div class="space-y-2">
              <Label for="priority">Priority</Label>
              <NativeSelect id="priority" v-model.number="form.priority" :disabled="loading">
                <option v-for="p in priorities" :key="p.value" :value="p.value">
                  {{ p.label }}
                </option>
              </NativeSelect>
            </div>
          </div>

          <div v-if="effMembers.length > 0" class="space-y-2">
            <Label>Assignees</Label>
            <div class="flex flex-wrap items-center gap-2">
              <div
                v-for="member in selectedAssignees"
                :key="member.user_id"
                class="flex items-center gap-1.5 rounded-md border bg-muted/50 py-1 pl-1 pr-1"
              >
                <Avatar class="size-5">
                  <AvatarFallback class="text-[10px]" :seed="member.user_id">
                    {{ member.first_name[0] }}{{ member.last_name[0] }}
                  </AvatarFallback>
                </Avatar>
                <span class="text-sm">{{ member.first_name }} {{ member.last_name }}</span>
                <button
                  type="button"
                  :aria-label="`Remove ${member.first_name} ${member.last_name}`"
                  class="ml-0.5 flex size-4 items-center justify-center rounded-sm hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring outline-none"
                  :disabled="loading"
                  @click="removeAssignee(member.user_id)"
                >
                  <X class="size-3" />
                </button>
              </div>
              <Popover v-model:open="showAssigneePopover">
                <PopoverTrigger as-child>
                  <Button type="button" variant="outline" size="sm" class="h-7 gap-1.5" :disabled="loading || filteredAssigneeOptions.length === 0 && selectedAssignees.length === effMembers.length">
                    <Plus class="size-3.5" />
                    Add
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="start" class="w-56 p-0">
                  <div class="border-b px-3 py-2">
                    <div class="relative">
                      <Search class="absolute left-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
                      <Input
                        v-model="assigneeSearch"
                        placeholder="Search members..."
                        class="h-8 pl-7 text-sm"
                      />
                    </div>
                  </div>
                  <div class="max-h-48 overflow-y-auto">
                    <div class="py-1">
                      <button
                        v-for="member in filteredAssigneeOptions"
                        :key="member.user_id"
                        type="button"
                        class="flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-accent"
                        @click="addAssignee(member.user_id)"
                      >
                        <Avatar class="size-6">
                          <AvatarFallback class="text-xs" :seed="member.user_id">
                            {{ member.first_name[0] }}{{ member.last_name[0] }}
                          </AvatarFallback>
                        </Avatar>
                        {{ member.first_name }} {{ member.last_name }}
                      </button>
                      <p
                        v-if="filteredAssigneeOptions.length === 0"
                        class="px-3 py-2 text-center text-sm text-muted-foreground"
                      >
                        No members found
                      </p>
                    </div>
                  </div>
                </PopoverContent>
              </Popover>
            </div>
          </div>

          <div v-if="effLabels.length > 0" class="space-y-2">
            <Label>Labels</Label>
            <div class="flex flex-wrap items-center gap-2">
              <div
                v-for="label in selectedLabels"
                :key="label.id"
                class="flex items-center gap-1.5 rounded-md px-2 py-1"
                :style="{
                  backgroundColor: label.color + '20',
                  color: label.color,
                }"
              >
                <span class="text-sm font-medium">{{ label.name }}</span>
                <button
                  type="button"
                  :aria-label="`Remove ${label.name}`"
                  class="flex size-4 items-center justify-center rounded-sm hover:opacity-70 focus-visible:ring-2 focus-visible:ring-ring outline-none"
                  :disabled="loading"
                  @click="removeLabel(label.id)"
                >
                  <X class="size-3" />
                </button>
              </div>
              <Popover v-model:open="showLabelPopover">
                <PopoverTrigger as-child>
                  <Button type="button" variant="outline" size="sm" class="h-7 gap-1.5" :disabled="loading || filteredLabelOptions.length === 0 && selectedLabels.length === effLabels.length">
                    <Plus class="size-3.5" />
                    Add
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="start" class="w-48 p-0">
                  <div class="border-b px-3 py-2">
                    <div class="relative">
                      <Search class="absolute left-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
                      <Input
                        v-model="labelSearch"
                        placeholder="Search labels..."
                        class="h-8 pl-7 text-sm"
                      />
                    </div>
                  </div>
                  <div class="max-h-48 overflow-y-auto">
                    <div class="py-1">
                      <button
                        v-for="label in filteredLabelOptions"
                        :key="label.id"
                        type="button"
                        class="flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-accent"
                        @click="addLabel(label.id)"
                      >
                        <div
                          class="size-3 shrink-0 rounded-full"
                          :style="{ backgroundColor: label.color }"
                        />
                        {{ label.name }}
                      </button>
                      <p
                        v-if="filteredLabelOptions.length === 0"
                        class="px-3 py-2 text-center text-sm text-muted-foreground"
                      >
                        No labels found
                      </p>
                    </div>
                  </div>
                </PopoverContent>
              </Popover>
            </div>
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
            Create Task
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

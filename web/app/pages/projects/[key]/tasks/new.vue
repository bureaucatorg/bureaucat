<script setup lang="ts">
import { ChevronLeft, Loader2, Plus, X, Search } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { ProjectState, ProjectLabel, ProjectMember, TaskTemplate } from "~/types";

definePageMeta({
  middleware: ["auth"],
});

const route = useRoute();
const router = useRouter();
const projectKey = computed(() => route.params.key as string);

useHead({
  title: computed(() => `New Task — ${projectKey.value}`),
});

const {
  currentProject,
  members,
  states,
  labels,
  templates,
  getProject,
  listMembers,
  listStates,
  listLabels,
  listTemplates,
} = useProjects();

const { createTask } = useTasks();

const pageLoading = ref(true);
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

const defaultState = computed(() => states.value.find((s) => s.is_default));

const isMember = computed(
  () => currentProject.value?.role === "admin" || currentProject.value?.role === "member"
);

async function loadProjectData() {
  pageLoading.value = true;

  if (!currentProject.value || currentProject.value.project_key !== projectKey.value) {
    const result = await getProject(projectKey.value);
    if (!result.success) {
      error.value = result.error || "Failed to load project";
      pageLoading.value = false;
      return;
    }
  }

  await Promise.all([
    listMembers(projectKey.value),
    listStates(projectKey.value),
    listLabels(projectKey.value),
    listTemplates(projectKey.value),
  ]);

  // Set default state after loading
  if (defaultState.value) {
    form.value.state_id = defaultState.value.id;
  }

  pageLoading.value = false;
}

watch(selectedTemplateId, (id) => {
  if (!id) return;
  const tmpl = templates.value.find((t) => t.id === id);
  if (tmpl) {
    form.value.title = tmpl.title;
    form.value.description = tmpl.description;
  }
});

async function handleSubmit() {
  loading.value = true;
  error.value = null;

  const result = await createTask(projectKey.value, {
    title: form.value.title,
    description: form.value.description || undefined,
    state_id: form.value.state_id || undefined,
    priority: form.value.priority,
    assignees: form.value.assignees.length > 0 ? form.value.assignees : undefined,
    labels: form.value.labels.length > 0 ? form.value.labels : undefined,
  });

  loading.value = false;

  if (result.success) {
    toast.success(`Task ${result.data?.task_id} created`);
    router.push(`/projects/${projectKey.value}/tasks/${result.data?.task_number}`);
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

// Assignees
const assigneeSearch = ref("");
const showAssigneePopover = ref(false);

const selectedAssignees = computed(() =>
  members.value.filter((m) => form.value.assignees.includes(m.user_id))
);

const filteredAssigneeOptions = computed(() => {
  const selected = new Set(form.value.assignees);
  const available = members.value.filter((m) => !selected.has(m.user_id));
  const q = assigneeSearch.value.toLowerCase().trim();
  if (!q) return available;
  return available.filter(
    (m) =>
      m.first_name.toLowerCase().includes(q) ||
      m.last_name.toLowerCase().includes(q) ||
      m.username.toLowerCase().includes(q)
  );
});

watch(showAssigneePopover, (open) => {
  if (!open) assigneeSearch.value = "";
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

// Labels
const labelSearch = ref("");
const showLabelPopover = ref(false);

const selectedLabels = computed(() =>
  labels.value.filter((l) => form.value.labels.includes(l.id))
);

const filteredLabelOptions = computed(() => {
  const selected = new Set(form.value.labels);
  const available = labels.value.filter((l) => !selected.has(l.id));
  const q = labelSearch.value.toLowerCase().trim();
  if (!q) return available;
  return available.filter((l) => l.name.toLowerCase().includes(q));
});

watch(showLabelPopover, (open) => {
  if (!open) labelSearch.value = "";
});

function addLabel(labelId: string) {
  if (!form.value.labels.includes(labelId)) {
    form.value.labels.push(labelId);
  }
  showLabelPopover.value = false;
}

function removeLabel(labelId: string) {
  form.value.labels = form.value.labels.filter((id) => id !== labelId);
}

onMounted(() => {
  loadProjectData();
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-3xl px-6 py-8">
        <!-- Loading -->
        <div v-if="pageLoading" class="flex items-center justify-center py-20">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <!-- Not a member -->
        <div
          v-else-if="!isMember"
          class="flex flex-col items-center justify-center py-20"
        >
          <p class="text-lg text-destructive">You don't have permission to create tasks</p>
          <Button class="mt-4" variant="outline" as-child>
            <NuxtLink :to="`/projects/${projectKey}`">Back to Project</NuxtLink>
          </Button>
        </div>

        <template v-else>
          <!-- Breadcrumb -->
          <nav class="mb-6 flex items-center gap-2 text-sm text-muted-foreground">
            <ChevronLeft class="size-4" />
            <NuxtLink :to="`/projects/${projectKey}`" class="hover:text-foreground">
              {{ projectKey }}
            </NuxtLink>
            <span>/</span>
            <span class="font-semibold text-foreground">New Task</span>
          </nav>

          <h1 class="mb-8 text-2xl font-bold tracking-tight">Create New Task</h1>

          <form class="space-y-6" @submit.prevent="handleSubmit">
            <div
              v-if="error"
              role="alert"
              class="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
            >
              {{ error }}
            </div>

            <!-- Template -->
            <div v-if="templates.length > 0" class="space-y-2">
              <Label for="template">Template</Label>
              <NativeSelect id="template" v-model="selectedTemplateId" :disabled="loading">
                <option value="">No template</option>
                <option v-for="tmpl in templates" :key="tmpl.id" :value="tmpl.id">
                  {{ tmpl.name }}
                </option>
              </NativeSelect>
            </div>

            <!-- Title -->
            <div class="space-y-2">
              <Label for="title">Title</Label>
              <Input
                id="title"
                v-model="form.title"
                placeholder="Task title"
                required
                :disabled="loading"
                class="text-lg"
              />
            </div>

            <!-- Description -->
            <div class="space-y-2">
              <Label>Description</Label>
              <TiptapEditor
                v-model="form.description"
                :disabled="loading"
              />
            </div>

            <!-- State & Priority -->
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <Label for="state">State</Label>
                <NativeSelect id="state" v-model="form.state_id" :disabled="loading">
                  <option v-for="state in states" :key="state.id" :value="state.id">
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

            <!-- Assignees -->
            <div v-if="members.length > 0" class="space-y-2">
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
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      class="h-7 gap-1.5"
                      :disabled="loading || (filteredAssigneeOptions.length === 0 && selectedAssignees.length === members.length)"
                    >
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

            <!-- Labels -->
            <div v-if="labels.length > 0" class="space-y-2">
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
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      class="h-7 gap-1.5"
                      :disabled="loading || (filteredLabelOptions.length === 0 && selectedLabels.length === labels.length)"
                    >
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

            <!-- Actions -->
            <div class="flex items-center justify-end gap-3 border-t pt-6">
              <Button type="button" variant="outline" :disabled="loading" as-child>
                <NuxtLink :to="`/projects/${projectKey}`">Cancel</NuxtLink>
              </Button>
              <Button type="submit" :disabled="loading || !form.title">
                <Loader2 v-if="loading" class="mr-2 size-4 animate-spin" />
                Create Task
              </Button>
            </div>
          </form>
        </template>
      </div>
    </main>
  </div>
</template>

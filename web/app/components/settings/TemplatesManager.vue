<script setup lang="ts">
import { Plus, Trash2, Pencil, Loader2 } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { CreateTemplateRequest, ProjectLabel, ProjectMember, TaskTemplate } from "~/types";

const props = defineProps<{
  templates: TaskTemplate[];
  labels: ProjectLabel[];
  members: ProjectMember[];
  projectKey: string;
  isAdmin: boolean;
  requireTemplate: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const { createTemplate, updateTemplate, deleteTemplate, updateProject } = useProjects();

const loading = ref(false);
const showCreateForm = ref(false);
const editingId = ref<string | null>(null);
const deletingId = ref<string | null>(null);
const savingSettings = ref(false);

async function toggleRequire(value: boolean) {
  savingSettings.value = true;
  const res = await updateProject(props.projectKey, { require_template: value });
  savingSettings.value = false;
  if (res.success) {
    toast.success(value ? "Templates are now required" : "Templates are now optional");
  } else {
    toast.error(res.error || "Failed to update template settings");
  }
}

async function handleCreate(data: CreateTemplateRequest) {
  loading.value = true;
  const result = await createTemplate(props.projectKey, data);
  loading.value = false;
  if (result.success) {
    toast.success("Template created");
    showCreateForm.value = false;
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to create template");
  }
}

async function handleUpdate(data: CreateTemplateRequest) {
  if (!editingId.value) return;
  loading.value = true;
  const result = await updateTemplate(props.projectKey, editingId.value, data);
  loading.value = false;
  if (result.success) {
    toast.success("Template updated");
    editingId.value = null;
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to update template");
  }
}

async function handleDelete(template: TaskTemplate) {
  deletingId.value = template.id;
  const result = await deleteTemplate(props.projectKey, template.id);
  deletingId.value = null;
  if (result.success) {
    toast.success("Template deleted");
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to delete template");
  }
}

function fieldSummary(t: TaskTemplate) {
  const n = t.fields?.length ?? 0;
  if (!n) return "";
  const req = t.fields.filter((f) => f.required).length;
  return `${n} field${n === 1 ? "" : "s"}${req ? `, ${req} required` : ""}`;
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h3 class="font-semibold">Templates</h3>
        <p class="text-sm text-muted-foreground">
          Structured forms people fill when creating tasks
        </p>
      </div>
      <Button
        v-if="isAdmin && !showCreateForm"
        size="sm"
        @click="showCreateForm = true"
      >
        <Plus class="mr-1.5 size-4" />
        Add Template
      </Button>
    </div>

    <div class="flex items-center justify-between gap-4 rounded-lg border px-3 py-2">
      <div>
        <p class="text-sm font-medium">Require a template</p>
        <p class="text-xs text-muted-foreground">
          New tasks must use one of these templates. Subtasks are exempt.
        </p>
      </div>
      <Switch
        :checked="requireTemplate"
        :disabled="!isAdmin || savingSettings || (!requireTemplate && templates.length === 0)"
        @update:checked="toggleRequire"
      />
    </div>

    <Card v-if="showCreateForm" class="border-dashed">
      <CardContent class="pt-6">
        <TemplateForm
          :labels="labels"
          :members="members"
          :loading="loading"
          submit-label="Create"
          @submit="handleCreate"
          @cancel="showCreateForm = false"
        />
      </CardContent>
    </Card>

    <div v-if="templates.length === 0" class="py-8 text-center text-sm text-muted-foreground">
      No templates yet. Create your first template to speed up task creation.
    </div>
    <div v-else class="space-y-2">
      <div
        v-for="template in templates"
        :key="template.id"
        class="rounded-lg border px-3 py-2"
      >
        <div v-if="editingId === template.id" class="py-2">
          <TemplateForm
            :template="template"
            :labels="labels"
            :members="members"
            :loading="loading"
            submit-label="Save"
            @submit="handleUpdate"
            @cancel="editingId = null"
          />
        </div>
        <div v-else class="flex items-center gap-3">
          <div class="min-w-0 flex-1">
            <p class="font-medium">{{ template.name }}</p>
            <p class="truncate text-sm text-muted-foreground">
              {{ [template.title, fieldSummary(template)].filter(Boolean).join(" · ") }}
            </p>
          </div>
          <Button
            v-if="isAdmin"
            variant="ghost"
            size="icon"
            aria-label="Edit template"
            class="size-8"
            @click="editingId = template.id"
          >
            <Pencil class="size-4" />
          </Button>
          <Button
            v-if="isAdmin"
            variant="ghost"
            size="icon"
            aria-label="Delete template"
            class="size-8 text-destructive hover:text-destructive"
            :disabled="deletingId === template.id"
            @click="handleDelete(template)"
          >
            <Loader2 v-if="deletingId === template.id" class="size-4 animate-spin" />
            <Trash2 v-else class="size-4" />
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

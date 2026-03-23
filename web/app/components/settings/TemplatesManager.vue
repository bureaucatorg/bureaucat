<script setup lang="ts">
import { Plus, Trash2, Pencil, Loader2, Check, X } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { TaskTemplate } from "~/types";

const props = defineProps<{
  templates: TaskTemplate[];
  projectKey: string;
  isAdmin: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const { createTemplate, updateTemplate, deleteTemplate } = useProjects();

const loading = ref(false);
const showCreateForm = ref(false);
const editingId = ref<string | null>(null);
const deletingId = ref<string | null>(null);

const newTemplate = ref({
  name: "",
  title: "",
  description: "",
});

const editForm = ref({
  name: "",
  title: "",
  description: "",
});

async function handleCreate() {
  if (!newTemplate.value.name.trim()) return;

  loading.value = true;
  const result = await createTemplate(props.projectKey, {
    name: newTemplate.value.name,
    title: newTemplate.value.title,
    description: newTemplate.value.description,
  });
  loading.value = false;

  if (result.success) {
    toast.success("Template created");
    newTemplate.value = { name: "", title: "", description: "" };
    showCreateForm.value = false;
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to create template");
  }
}

function startEdit(template: TaskTemplate) {
  editingId.value = template.id;
  editForm.value = {
    name: template.name,
    title: template.title,
    description: template.description,
  };
}

function cancelEdit() {
  editingId.value = null;
  editForm.value = { name: "", title: "", description: "" };
}

async function handleUpdate() {
  if (!editingId.value || !editForm.value.name.trim()) return;

  loading.value = true;
  const result = await updateTemplate(props.projectKey, editingId.value, {
    name: editForm.value.name,
    title: editForm.value.title,
    description: editForm.value.description,
  });
  loading.value = false;

  if (result.success) {
    toast.success("Template updated");
    cancelEdit();
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
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="font-semibold">Templates</h3>
        <p class="text-sm text-muted-foreground">
          Pre-fill task title and description when creating tasks
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

    <!-- Create form -->
    <Card v-if="showCreateForm" class="border-dashed">
      <CardContent class="pt-6">
        <form class="space-y-4" @submit.prevent="handleCreate">
          <div class="space-y-2">
            <Label>Name</Label>
            <Input
              v-model="newTemplate.name"
              placeholder="Template name"
              :disabled="loading"
            />
          </div>
          <div class="space-y-2">
            <Label>Title</Label>
            <Input
              v-model="newTemplate.title"
              placeholder="Pre-filled task title"
              :disabled="loading"
            />
          </div>
          <div class="space-y-2">
            <Label>Description</Label>
            <Textarea
              v-model="newTemplate.description"
              placeholder="Pre-filled task description"
              rows="3"
              :disabled="loading"
            />
          </div>
          <div class="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              :disabled="loading"
              @click="showCreateForm = false"
            >
              Cancel
            </Button>
            <Button type="submit" size="sm" :disabled="loading || !newTemplate.name">
              <Loader2 v-if="loading" class="mr-1.5 size-4 animate-spin" />
              Create
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>

    <!-- Templates list -->
    <div v-if="templates.length === 0" class="py-8 text-center text-sm text-muted-foreground">
      No templates yet. Create your first template to speed up task creation.
    </div>
    <div v-else class="space-y-2">
      <div
        v-for="template in templates"
        :key="template.id"
        class="rounded-lg border px-3 py-2"
      >
        <template v-if="editingId === template.id">
          <form class="space-y-3" @submit.prevent="handleUpdate">
            <div class="space-y-2">
              <Label>Name</Label>
              <Input
                v-model="editForm.name"
                class="h-8"
                :disabled="loading"
              />
            </div>
            <div class="space-y-2">
              <Label>Title</Label>
              <Input
                v-model="editForm.title"
                class="h-8"
                :disabled="loading"
              />
            </div>
            <div class="space-y-2">
              <Label>Description</Label>
              <Textarea
                v-model="editForm.description"
                rows="2"
                :disabled="loading"
              />
            </div>
            <div class="flex justify-end gap-1">
              <Button
                type="submit"
                variant="ghost"
                size="icon"
                aria-label="Save"
                class="size-8"
                :disabled="loading"
              >
                <Loader2 v-if="loading" class="size-4 animate-spin" />
                <Check v-else class="size-4" />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                aria-label="Cancel"
                class="size-8"
                :disabled="loading"
                @click="cancelEdit"
              >
                <X class="size-4" />
              </Button>
            </div>
          </form>
        </template>
        <template v-else>
          <div class="flex items-center gap-3">
            <div class="min-w-0 flex-1">
              <p class="font-medium">{{ template.name }}</p>
              <p v-if="template.title" class="truncate text-sm text-muted-foreground">
                {{ template.title }}
              </p>
            </div>
            <Button
              v-if="isAdmin"
              variant="ghost"
              size="icon"
              aria-label="Edit template"
              class="size-8"
              @click="startEdit(template)"
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
              <Loader2
                v-if="deletingId === template.id"
                class="size-4 animate-spin"
              />
              <Trash2 v-else class="size-4" />
            </Button>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

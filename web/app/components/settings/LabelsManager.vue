<script setup lang="ts">
import { Plus, Trash2, Pencil, Loader2, Check, X } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { ProjectLabel } from "~/types";

const props = defineProps<{
  labels: ProjectLabel[];
  projectKey: string;
  isAdmin: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const { createLabel, updateLabel, deleteLabel } = useProjects();

const loading = ref(false);
const showCreateForm = ref(false);
const editingId = ref<string | null>(null);
const deletingId = ref<string | null>(null);

const newLabel = ref({
  name: "",
  color: "#3B82F6",
});

const editForm = ref({
  name: "",
  color: "",
});

const presetColors = [
  "#EF4444", "#F97316", "#EAB308", "#22C55E", "#10B981",
  "#3B82F6", "#6366F1", "#8B5CF6", "#EC4899", "#6B7280",
];

async function handleCreate() {
  if (!newLabel.value.name.trim()) return;

  loading.value = true;
  const result = await createLabel(props.projectKey, {
    name: newLabel.value.name,
    color: newLabel.value.color,
  });
  loading.value = false;

  if (result.success) {
    toast.success("Label created");
    newLabel.value = { name: "", color: "#3B82F6" };
    showCreateForm.value = false;
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to create label");
  }
}

function startEdit(label: ProjectLabel) {
  editingId.value = label.id;
  editForm.value = { name: label.name, color: label.color };
}

function cancelEdit() {
  editingId.value = null;
  editForm.value = { name: "", color: "" };
}

async function handleUpdate() {
  if (!editingId.value || !editForm.value.name.trim()) return;

  loading.value = true;
  const result = await updateLabel(props.projectKey, editingId.value, {
    name: editForm.value.name,
    color: editForm.value.color,
  });
  loading.value = false;

  if (result.success) {
    toast.success("Label updated");
    cancelEdit();
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to update label");
  }
}

async function handleDelete(label: ProjectLabel) {
  deletingId.value = label.id;
  const result = await deleteLabel(props.projectKey, label.id);
  deletingId.value = null;

  if (result.success) {
    toast.success("Label deleted");
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to delete label");
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="font-semibold">Labels</h3>
        <p class="text-sm text-muted-foreground">
          Categorize tasks with custom labels
        </p>
      </div>
      <Button
        v-if="isAdmin && !showCreateForm"
        size="sm"
        @click="showCreateForm = true"
      >
        <Plus class="mr-1.5 size-4" />
        Add Label
      </Button>
    </div>

    <!-- Create form -->
    <Card v-if="showCreateForm" class="border-dashed">
      <CardContent class="pt-6">
        <form class="space-y-4" @submit.prevent="handleCreate">
          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label>Name</Label>
              <Input
                v-model="newLabel.name"
                placeholder="Label name"
                :disabled="loading"
              />
            </div>
            <div class="space-y-2">
              <Label>Color</Label>
              <div class="flex flex-wrap gap-1">
                <button
                  v-for="color in presetColors"
                  :key="color"
                  type="button"
                  :aria-label="`Select color ${color}`"
                  class="size-6 rounded border-2 transition-all focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
                  :class="{
                    'border-foreground scale-110': newLabel.color === color,
                    'border-transparent': newLabel.color !== color,
                  }"
                  :style="{ backgroundColor: color }"
                  @click="newLabel.color = color"
                />
              </div>
            </div>
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
            <Button type="submit" size="sm" :disabled="loading || !newLabel.name">
              <Loader2 v-if="loading" class="mr-1.5 size-4 animate-spin" />
              Create
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>

    <!-- Labels list -->
    <div v-if="labels.length === 0" class="py-8 text-center text-sm text-muted-foreground">
      No labels yet. Create your first label to categorize tasks.
    </div>
    <div v-else class="space-y-2">
      <div
        v-for="label in labels"
        :key="label.id"
        class="flex items-center gap-3 rounded-lg border px-3 py-2"
      >
        <template v-if="editingId === label.id">
          <Input
            v-model="editForm.name"
            class="h-8 flex-1"
            :disabled="loading"
          />
          <div class="flex gap-1">
            <button
              v-for="color in presetColors"
              :key="color"
              type="button"
              :aria-label="`Select color ${color}`"
              class="size-5 rounded transition-all focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 outline-none"
              :class="{
                'ring-2 ring-foreground ring-offset-1': editForm.color === color,
              }"
              :style="{ backgroundColor: color }"
              @click="editForm.color = color"
            />
          </div>
          <Button
            variant="ghost"
            size="icon"
            aria-label="Save"
            class="size-8"
            :disabled="loading"
            @click="handleUpdate"
          >
            <Loader2 v-if="loading" class="size-4 animate-spin" />
            <Check v-else class="size-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            aria-label="Cancel"
            class="size-8"
            :disabled="loading"
            @click="cancelEdit"
          >
            <X class="size-4" />
          </Button>
        </template>
        <template v-else>
          <span
            class="rounded px-2 py-0.5 text-sm font-medium"
            :style="{
              backgroundColor: label.color + '20',
              color: label.color,
            }"
          >
            {{ label.name }}
          </span>
          <span class="flex-1" />
          <Button
            v-if="isAdmin"
            variant="ghost"
            size="icon"
            aria-label="Edit label"
            class="size-8"
            @click="startEdit(label)"
          >
            <Pencil class="size-4" />
          </Button>
          <Button
            v-if="isAdmin"
            variant="ghost"
            size="icon"
            aria-label="Delete label"
            class="size-8 text-destructive hover:text-destructive"
            :disabled="deletingId === label.id"
            @click="handleDelete(label)"
          >
            <Loader2
              v-if="deletingId === label.id"
              class="size-4 animate-spin"
            />
            <Trash2 v-else class="size-4" />
          </Button>
        </template>
      </div>
    </div>
  </div>
</template>

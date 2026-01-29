<script setup lang="ts">
import { Loader2, Trash2 } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { Project } from "~/types";

const props = defineProps<{
  project: Project;
  isAdmin: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
  deleted: [];
}>();

const router = useRouter();
const { updateProject, deleteProject } = useProjects();

const loading = ref(false);
const deleteLoading = ref(false);
const showDeleteConfirm = ref(false);

const form = ref({
  name: props.project.name,
  description: props.project.description || "",
});

watch(
  () => props.project,
  (p) => {
    form.value = {
      name: p.name,
      description: p.description || "",
    };
  },
  { immediate: true }
);

async function handleSave() {
  loading.value = true;
  const result = await updateProject(props.project.project_key, {
    name: form.value.name,
    description: form.value.description || undefined,
  });
  loading.value = false;

  if (result.success) {
    toast.success("Project updated");
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to update project");
  }
}

async function handleDelete() {
  deleteLoading.value = true;
  const result = await deleteProject(props.project.project_key);
  deleteLoading.value = false;

  if (result.success) {
    toast.success("Project deleted");
    router.push("/projects");
    emit("deleted");
  } else {
    toast.error(result.error || "Failed to delete project");
  }
}

const hasChanges = computed(() => {
  return (
    form.value.name !== props.project.name ||
    form.value.description !== (props.project.description || "")
  );
});
</script>

<template>
  <div class="space-y-8">
    <!-- General settings -->
    <div class="space-y-4">
      <div>
        <h3 class="font-semibold">General</h3>
        <p class="text-sm text-muted-foreground">
          Basic project information
        </p>
      </div>

      <Card>
        <CardContent class="pt-6">
          <form class="space-y-4" @submit.prevent="handleSave">
            <div class="space-y-2">
              <Label for="project-key">Project Key</Label>
              <Input
                id="project-key"
                :value="project.project_key"
                disabled
                class="font-mono"
              />
              <p class="text-xs text-muted-foreground">
                Cannot be changed after creation
              </p>
            </div>

            <div class="space-y-2">
              <Label for="name">Name</Label>
              <Input
                id="name"
                v-model="form.name"
                :disabled="loading || !isAdmin"
              />
            </div>

            <div class="space-y-2">
              <Label for="description">Description</Label>
              <Textarea
                id="description"
                v-model="form.description"
                rows="3"
                :disabled="loading || !isAdmin"
              />
            </div>

            <div v-if="isAdmin" class="flex justify-end">
              <Button type="submit" :disabled="loading || !hasChanges">
                <Loader2 v-if="loading" class="mr-2 size-4 animate-spin" />
                Save Changes
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>

    <!-- Danger zone -->
    <div v-if="isAdmin" class="space-y-4">
      <div>
        <h3 class="font-semibold text-destructive">Danger Zone</h3>
        <p class="text-sm text-muted-foreground">
          Irreversible actions
        </p>
      </div>

      <Card class="border-destructive/50">
        <CardContent class="pt-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="font-medium">Delete Project</p>
              <p class="text-sm text-muted-foreground">
                Permanently delete this project and all its data
              </p>
            </div>
            <Button
              variant="destructive"
              :disabled="deleteLoading"
              @click="showDeleteConfirm = true"
            >
              <Trash2 class="mr-2 size-4" />
              Delete Project
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Delete confirmation dialog -->
    <Dialog v-model:open="showDeleteConfirm">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Project</DialogTitle>
          <DialogDescription>
            Are you sure you want to delete <strong>{{ project.name }}</strong>?
            This will permanently delete all tasks, comments, and activity logs.
            This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button
            variant="outline"
            :disabled="deleteLoading"
            @click="showDeleteConfirm = false"
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            :disabled="deleteLoading"
            @click="handleDelete"
          >
            <Loader2 v-if="deleteLoading" class="mr-2 size-4 animate-spin" />
            Delete Project
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

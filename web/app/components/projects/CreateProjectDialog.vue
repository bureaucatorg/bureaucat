<script setup lang="ts">
import { Loader2 } from "lucide-vue-next";
import { toast } from "vue-sonner";

const open = defineModel<boolean>("open", { default: false });

const emit = defineEmits<{
  created: [];
}>();

const { createProject } = useProjects();

const loading = ref(false);
const error = ref<string | null>(null);
const form = ref({
  project_key: "",
  name: "",
  description: "",
});

function resetForm() {
  form.value = {
    project_key: "",
    name: "",
    description: "",
  };
  error.value = null;
}

watch(open, (isOpen) => {
  if (isOpen) {
    resetForm();
  }
});

async function handleSubmit() {
  loading.value = true;
  error.value = null;

  const result = await createProject({
    project_key: form.value.project_key.toUpperCase(),
    name: form.value.name,
    description: form.value.description || undefined,
  });

  loading.value = false;

  if (result.success) {
    toast.success("Project created successfully");
    open.value = false;
    emit("created");
  } else {
    error.value = result.error || "Failed to create project";
  }
}

function validateProjectKey(e: Event) {
  const input = e.target as HTMLInputElement;
  // Only allow alphanumeric characters
  input.value = input.value.replace(/[^a-zA-Z0-9]/g, "").toUpperCase();
  form.value.project_key = input.value;
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>Create New Project</DialogTitle>
        <DialogDescription>
          Set up a new project to organize your tasks and workflows.
        </DialogDescription>
      </DialogHeader>
      <form class="space-y-4" @submit.prevent="handleSubmit">
        <div
          v-if="error"
          class="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
        >
          {{ error }}
        </div>
        <div class="space-y-2">
          <Label for="project_key">Project Key</Label>
          <Input
            id="project_key"
            :model-value="form.project_key"
            placeholder="DEVOP"
            maxlength="10"
            required
            :disabled="loading"
            class="font-mono uppercase"
            @input="validateProjectKey"
          />
          <p class="text-xs text-muted-foreground">
            2-10 alphanumeric characters. Used in task IDs (e.g., DEVOP-123)
          </p>
        </div>
        <div class="space-y-2">
          <Label for="name">Project Name</Label>
          <Input
            id="name"
            v-model="form.name"
            placeholder="DevOps Platform"
            required
            :disabled="loading"
          />
        </div>
        <div class="space-y-2">
          <Label for="description">Description</Label>
          <Textarea
            id="description"
            v-model="form.description"
            placeholder="Brief description of the project..."
            rows="3"
            :disabled="loading"
          />
        </div>
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            :disabled="loading"
            @click="open = false"
          >
            Cancel
          </Button>
          <Button type="submit" :disabled="loading || !form.project_key || !form.name">
            <Loader2 v-if="loading" class="mr-2 size-4 animate-spin" />
            Create Project
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

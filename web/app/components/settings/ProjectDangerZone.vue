<script setup lang="ts">
import { Loader2, Trash2 } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { Project } from "~/types";

const props = defineProps<{
  project: Project;
}>();

const router = useRouter();
const { deleteProject } = useProjects();

const deleteLoading = ref(false);
const showDeleteConfirm = ref(false);

async function handleDelete() {
  deleteLoading.value = true;
  const result = await deleteProject(props.project.project_key);
  deleteLoading.value = false;

  if (result.success) {
    toast.success("Project deleted");
    router.push("/projects");
  } else {
    toast.error(result.error || "Failed to delete project");
  }
}
</script>

<template>
  <div class="space-y-4">
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

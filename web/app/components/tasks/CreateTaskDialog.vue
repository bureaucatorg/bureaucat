<script setup lang="ts">
import { Loader2 } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { ProjectState, ProjectLabel, ProjectMember } from "~/types";

const props = defineProps<{
  projectKey: string;
  states: ProjectState[];
  labels: ProjectLabel[];
  members: ProjectMember[];
}>();

const open = defineModel<boolean>("open", { default: false });

const emit = defineEmits<{
  created: [];
}>();

const { createTask } = useTasks();

const loading = ref(false);
const error = ref<string | null>(null);
const form = ref({
  title: "",
  description: "",
  state_id: "",
  priority: 0,
  assignees: [] as string[],
  labels: [] as string[],
});

const defaultState = computed(() => props.states.find((s) => s.is_default));

function resetForm() {
  form.value = {
    title: "",
    description: "",
    state_id: defaultState.value?.id || "",
    priority: 0,
    assignees: [],
    labels: [],
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

  const result = await createTask(props.projectKey, {
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
    open.value = false;
    emit("created");
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

function toggleAssignee(userId: string) {
  const idx = form.value.assignees.indexOf(userId);
  if (idx === -1) {
    form.value.assignees.push(userId);
  } else {
    form.value.assignees.splice(idx, 1);
  }
}

function toggleLabel(labelId: string) {
  const idx = form.value.labels.indexOf(labelId);
  if (idx === -1) {
    form.value.labels.push(labelId);
  } else {
    form.value.labels.splice(idx, 1);
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>Create New Task</DialogTitle>
        <DialogDescription>
          Add a new task to {{ projectKey }}
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
          <Label for="description">Description</Label>
          <Textarea
            id="description"
            v-model="form.description"
            placeholder="Describe the task..."
            rows="3"
            :disabled="loading"
          />
        </div>

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

        <div v-if="members.length > 0" class="space-y-2">
          <Label>Assignees</Label>
          <div class="flex flex-wrap gap-2">
            <Button
              v-for="member in members"
              :key="member.user_id"
              type="button"
              size="sm"
              :variant="form.assignees.includes(member.user_id) ? 'default' : 'outline'"
              :disabled="loading"
              @click="toggleAssignee(member.user_id)"
            >
              {{ member.first_name }} {{ member.last_name }}
            </Button>
          </div>
        </div>

        <div v-if="labels.length > 0" class="space-y-2">
          <Label>Labels</Label>
          <div class="flex flex-wrap gap-2">
            <Button
              v-for="label in labels"
              :key="label.id"
              type="button"
              size="sm"
              :variant="form.labels.includes(label.id) ? 'default' : 'outline'"
              :disabled="loading"
              :style="{
                borderColor: form.labels.includes(label.id) ? label.color : undefined,
                backgroundColor: form.labels.includes(label.id) ? label.color : undefined,
              }"
              @click="toggleLabel(label.id)"
            >
              {{ label.name }}
            </Button>
          </div>
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
          <Button type="submit" :disabled="loading || !form.title">
            <Loader2 v-if="loading" class="mr-2 size-4 animate-spin" />
            Create Task
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

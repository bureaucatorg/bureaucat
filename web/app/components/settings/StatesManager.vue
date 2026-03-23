<script setup lang="ts">
import { Plus, Trash2, GripVertical, Loader2 } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { ProjectState, StateType } from "~/types";

const props = defineProps<{
  states: ProjectState[];
  projectKey: string;
  isAdmin: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const { createState, updateState, deleteState } = useProjects();

const loading = ref(false);
const showCreateForm = ref(false);
const editingId = ref<string | null>(null);
const deletingId = ref<string | null>(null);

const newState = ref({
  name: "",
  state_type: "unstarted" as StateType,
  color: "#3B82F6",
});

const stateTypes: { value: StateType; label: string }[] = [
  { value: "backlog", label: "Backlog" },
  { value: "unstarted", label: "Unstarted" },
  { value: "started", label: "Started" },
  { value: "completed", label: "Completed" },
  { value: "cancelled", label: "Cancelled" },
];

const presetColors = [
  "#6B7280", "#EF4444", "#F97316", "#EAB308", "#22C55E",
  "#10B981", "#3B82F6", "#8B5CF6", "#EC4899", "#14B8A6",
];

// Group states by type
const groupedStates = computed(() => {
  const groups: Record<string, ProjectState[]> = {};
  for (const type of stateTypes) {
    groups[type.value] = props.states
      .filter((s) => s.state_type === type.value)
      .sort((a, b) => a.position - b.position);
  }
  return groups;
});

async function handleCreate() {
  if (!newState.value.name.trim()) return;

  loading.value = true;
  const result = await createState(props.projectKey, {
    name: newState.value.name,
    state_type: newState.value.state_type,
    color: newState.value.color,
  });
  loading.value = false;

  if (result.success) {
    toast.success("State created");
    newState.value = { name: "", state_type: "unstarted", color: "#3B82F6" };
    showCreateForm.value = false;
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to create state");
  }
}

async function handleUpdate(state: ProjectState, updates: { name?: string; color?: string }) {
  loading.value = true;
  const result = await updateState(props.projectKey, state.id, updates);
  loading.value = false;

  if (result.success) {
    toast.success("State updated");
    editingId.value = null;
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to update state");
  }
}

async function handleDelete(state: ProjectState) {
  if (state.is_default) {
    toast.error("Cannot delete default state");
    return;
  }

  deletingId.value = state.id;
  const result = await deleteState(props.projectKey, state.id);
  deletingId.value = null;

  if (result.success) {
    toast.success("State deleted");
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to delete state");
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="font-semibold">Workflow States</h3>
        <p class="text-sm text-muted-foreground">
          Configure the states tasks can be in
        </p>
      </div>
      <Button
        v-if="isAdmin && !showCreateForm"
        size="sm"
        @click="showCreateForm = true"
      >
        <Plus class="mr-1.5 size-4" />
        Add State
      </Button>
    </div>

    <!-- Create form -->
    <Card v-if="showCreateForm" class="border-dashed">
      <CardContent class="pt-6">
        <form class="space-y-4" @submit.prevent="handleCreate">
          <div class="grid gap-4 sm:grid-cols-3">
            <div class="space-y-2">
              <Label>Name</Label>
              <Input
                v-model="newState.name"
                placeholder="State name"
                :disabled="loading"
              />
            </div>
            <div class="space-y-2">
              <Label>Type</Label>
              <NativeSelect v-model="newState.state_type" :disabled="loading">
                <option v-for="type in stateTypes" :key="type.value" :value="type.value">
                  {{ type.label }}
                </option>
              </NativeSelect>
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
                    'border-foreground scale-110': newState.color === color,
                    'border-transparent': newState.color !== color,
                  }"
                  :style="{ backgroundColor: color }"
                  @click="newState.color = color"
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
            <Button type="submit" size="sm" :disabled="loading || !newState.name">
              <Loader2 v-if="loading" class="mr-1.5 size-4 animate-spin" />
              Create
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>

    <!-- States by type -->
    <div v-for="type in stateTypes" :key="type.value" class="space-y-2">
      <h4 class="text-sm font-medium text-muted-foreground">{{ type.label }}</h4>
      <div v-if="groupedStates[type.value].length === 0" class="py-2 text-sm text-muted-foreground">
        No states in this group
      </div>
      <div v-else class="space-y-1">
        <div
          v-for="state in groupedStates[type.value]"
          :key="state.id"
          class="flex items-center gap-3 rounded-lg border px-3 py-2"
        >
          <div
            class="size-3 rounded-full"
            :style="{ backgroundColor: state.color }"
          />
          <span class="flex-1 text-sm font-medium">{{ state.name }}</span>
          <Badge v-if="state.is_default" variant="outline" class="text-xs">
            Default
          </Badge>
          <Button
            v-if="isAdmin && !state.is_default"
            variant="ghost"
            size="icon"
            aria-label="Delete state"
            class="size-8 text-destructive hover:text-destructive"
            :disabled="deletingId === state.id"
            @click="handleDelete(state)"
          >
            <Loader2
              v-if="deletingId === state.id"
              class="size-4 animate-spin"
            />
            <Trash2 v-else class="size-4" />
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

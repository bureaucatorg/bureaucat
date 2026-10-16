<script setup lang="ts">
import { Circle, CircleDot, CheckCircle2, XCircle, Clock, ChevronDown, Lock } from "lucide-vue-next";
import type { ProjectState, WorkflowTaskRef } from "~/types";

const props = withDefaults(
  defineProps<{
    states: ProjectState[];
    modelValue: string;
    disabled?: boolean;
    compact?: boolean;
    dense?: boolean;
    // When set, states the project workflow doesn't allow for this task are disabled.
    task?: WorkflowTaskRef;
  }>(),
  { compact: false, dense: false }
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const currentState = computed(() => props.states.find((s) => s.id === props.modelValue));

const { workflows, getWorkflow } = useWorkflow();
const { user } = useAuth();
watch(
  () => props.task?.project_key,
  (key) => {
    if (key) getWorkflow(key);
  },
  { immediate: true }
);
const allowed = computed(() =>
  props.task ? allowedTargetStateIds(workflows[props.task.project_key], props.task, user.value?.id) : null
);

function getStateIcon(stateType: string) {
  switch (stateType) {
    case "backlog":
      return Clock;
    case "unstarted":
      return Circle;
    case "started":
      return CircleDot;
    case "completed":
      return CheckCircle2;
    case "cancelled":
      return XCircle;
    default:
      return Circle;
  }
}

// Group states by type
const groupedStates = computed(() => {
  const groups: Record<string, ProjectState[]> = {
    backlog: [],
    unstarted: [],
    started: [],
    completed: [],
    cancelled: [],
  };
  for (const state of props.states) {
    if (groups[state.state_type]) {
      groups[state.state_type].push(state);
    }
  }
  return groups;
});
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button
        :variant="compact ? 'ghost' : 'outline'"
        :class="[
          compact ? 'h-auto gap-1.5 px-0 py-0 font-medium hover:bg-transparent' : 'justify-between',
          dense && 'gap-1 text-xs',
        ]"
        :disabled="disabled"
      >
        <span class="flex items-center" :class="dense ? 'gap-1' : 'gap-1.5'">
          <component
            :is="getStateIcon(currentState?.state_type || 'backlog')"
            :class="dense ? 'size-3.5 stroke-[2.5]' : compact ? 'size-5 stroke-[2.5]' : 'size-4'"
            :style="{ color: currentState?.color }"
          />
          {{ currentState?.name || "Select state" }}
        </span>
        <ChevronDown :class="dense ? 'size-3 opacity-50' : 'size-3.5 opacity-50'" />
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent class="w-56">
      <template v-for="(states, type) in groupedStates" :key="type">
        <template v-if="states.length > 0">
          <DropdownMenuLabel class="text-xs uppercase text-muted-foreground">
            {{ type }}
          </DropdownMenuLabel>
          <DropdownMenuItem
            v-for="state in states"
            :key="state.id"
            :disabled="!!allowed && !allowed.has(state.id)"
            @click="emit('update:modelValue', state.id)"
          >
            <component
              :is="getStateIcon(state.state_type)"
              class="mr-2 size-4"
              :style="{ color: state.color }"
            />
            {{ state.name }}
            <Lock
              v-if="allowed && !allowed.has(state.id)"
              class="ml-auto size-3.5 text-muted-foreground"
              aria-label="Not allowed by the project workflow"
            />
          </DropdownMenuItem>
          <DropdownMenuSeparator />
        </template>
      </template>
    </DropdownMenuContent>
  </DropdownMenu>
</template>

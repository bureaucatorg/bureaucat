<script setup lang="ts">
import { Circle, CircleDot, CheckCircle2, XCircle, Clock, ChevronDown } from "lucide-vue-next";
import type { ProjectState } from "~/types";

const props = withDefaults(
  defineProps<{
    states: ProjectState[];
    modelValue: string;
    disabled?: boolean;
    compact?: boolean;
  }>(),
  { compact: false }
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const currentState = computed(() => props.states.find((s) => s.id === props.modelValue));

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
        :class="compact ? 'h-auto gap-1.5 px-0 py-0 font-medium hover:bg-transparent' : 'justify-between'"
        :disabled="disabled"
      >
        <span class="flex items-center gap-1.5">
          <component
            :is="getStateIcon(currentState?.state_type || 'backlog')"
            :class="compact ? 'size-5 stroke-[2.5]' : 'size-4'"
            :style="{ color: currentState?.color }"
          />
          {{ currentState?.name || "Select state" }}
        </span>
        <ChevronDown class="size-3.5 opacity-50" />
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
            @click="emit('update:modelValue', state.id)"
          >
            <component
              :is="getStateIcon(state.state_type)"
              class="mr-2 size-4"
              :style="{ color: state.color }"
            />
            {{ state.name }}
          </DropdownMenuItem>
          <DropdownMenuSeparator />
        </template>
      </template>
    </DropdownMenuContent>
  </DropdownMenu>
</template>

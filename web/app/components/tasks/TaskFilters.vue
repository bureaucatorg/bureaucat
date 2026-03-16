<script setup lang="ts">
import {
  Search,
  X,
  ChevronDown,
  Circle,
  CircleDot,
  CheckCircle2,
  XCircle,
  Clock,
  User,
} from "lucide-vue-next";
import type { ProjectState, TaskFilters } from "~/types";
import type { ProjectMember } from "~/types";
import { PRIORITY_LABELS } from "~/types";

const props = defineProps<{
  states: ProjectState[];
  members: ProjectMember[];
  filters: TaskFilters;
}>();

const emit = defineEmits<{
  "update:filters": [filters: TaskFilters];
}>();

const searchQuery = ref(props.filters.q || "");
const selectedStateId = ref(props.filters.state_id || "");
const selectedPriority = ref(props.filters.priority?.toString() || "");
const selectedAssignee = ref(props.filters.assigned_to || "");

const priorities = [
  { value: "", label: "All priorities", color: "" },
  { value: "4", label: "Urgent", color: "#EF4444" },
  { value: "3", label: "High", color: "#F97316" },
  { value: "2", label: "Medium", color: "#EAB308" },
  { value: "1", label: "Low", color: "#3B82F6" },
  { value: "0", label: "No priority", color: "#6B7280" },
];

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

const currentState = computed(() => props.states.find((s) => s.id === selectedStateId.value));

const currentPriority = computed(() => {
  return priorities.find((p) => p.value === selectedPriority.value) || priorities[0];
});

function selectState(stateId: string) {
  selectedStateId.value = stateId;
}

function selectPriority(value: string) {
  selectedPriority.value = value;
}

const assigneeSearch = ref("");
const assigneeOpen = ref(false);

function selectAssignee(userId: string) {
  selectedAssignee.value = userId;
  assigneeSearch.value = "";
  assigneeOpen.value = false;
}

const currentAssignee = computed(() => {
  if (!selectedAssignee.value) return null;
  return props.members.find((m) => m.user_id === selectedAssignee.value) || null;
});

const filteredMembers = computed(() => {
  if (!assigneeSearch.value) return props.members;
  const q = assigneeSearch.value.toLowerCase();
  return props.members.filter(
    (m) =>
      m.first_name.toLowerCase().includes(q) ||
      m.last_name.toLowerCase().includes(q) ||
      m.username.toLowerCase().includes(q) ||
      m.email.toLowerCase().includes(q)
  );
});

function updateFilters() {
  const filters: TaskFilters = {};
  if (searchQuery.value) filters.q = searchQuery.value;
  if (selectedStateId.value) filters.state_id = selectedStateId.value;
  if (selectedPriority.value) filters.priority = parseInt(selectedPriority.value);
  if (selectedAssignee.value) filters.assigned_to = selectedAssignee.value;
  emit("update:filters", filters);
}

function clearFilters() {
  searchQuery.value = "";
  selectedStateId.value = "";
  selectedPriority.value = "";
  selectedAssignee.value = "";
  emit("update:filters", {});
}

const hasActiveFilters = computed(() => {
  return searchQuery.value || selectedStateId.value || selectedPriority.value || selectedAssignee.value;
});

watch([searchQuery, selectedStateId, selectedPriority, selectedAssignee], () => {
  updateFilters();
});
</script>

<template>
  <div class="flex flex-wrap items-center gap-3">
    <div class="relative flex-1 sm:max-w-xs">
      <Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        v-model="searchQuery"
        placeholder="Search tasks..."
        class="pl-9"
      />
    </div>

    <!-- State dropdown -->
    <DropdownMenu>
      <DropdownMenuTrigger as-child>
        <Button variant="outline" class="gap-1.5">
          <template v-if="currentState">
            <component
              :is="getStateIcon(currentState.state_type)"
              class="size-4"
              :style="{ color: currentState.color }"
            />
            {{ currentState.name }}
          </template>
          <template v-else>
            All states
          </template>
          <ChevronDown class="size-3.5 opacity-50" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent class="w-56">
        <DropdownMenuItem @click="selectState('')">
          All states
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <template v-for="(states, type) in groupedStates" :key="type">
          <template v-if="states.length > 0">
            <DropdownMenuLabel class="text-xs uppercase text-muted-foreground">
              {{ type }}
            </DropdownMenuLabel>
            <DropdownMenuItem
              v-for="state in states"
              :key="state.id"
              @click="selectState(state.id)"
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

    <!-- Priority dropdown -->
    <DropdownMenu>
      <DropdownMenuTrigger as-child>
        <Button variant="outline" class="gap-1.5">
          <span
            v-if="currentPriority.color"
            class="size-2 rounded-full"
            :style="{ backgroundColor: currentPriority.color }"
          />
          {{ currentPriority.label }}
          <ChevronDown class="size-3.5 opacity-50" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent class="w-40">
        <DropdownMenuItem
          v-for="p in priorities"
          :key="p.value"
          @click="selectPriority(p.value)"
        >
          <span
            v-if="p.color"
            class="mr-2 size-2 rounded-full"
            :style="{ backgroundColor: p.color }"
          />
          {{ p.label }}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>

    <!-- Assignee dropdown -->
    <Popover v-model:open="assigneeOpen">
      <PopoverTrigger as-child>
        <Button variant="outline" class="gap-1.5">
          <User class="size-4" />
          <template v-if="currentAssignee">
            {{ currentAssignee.first_name }} {{ currentAssignee.last_name }}
          </template>
          <template v-else>
            All assignees
          </template>
          <ChevronDown class="size-3.5 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent class="w-56 p-0" align="start">
        <div class="p-2">
          <div class="relative">
            <Search class="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              v-model="assigneeSearch"
              placeholder="Search members..."
              class="h-8 pl-8 text-sm"
            />
          </div>
        </div>
        <Separator />
        <div class="max-h-48 overflow-y-auto p-1">
          <button
            class="flex w-full items-center rounded-sm px-2 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground"
            @click="selectAssignee('')"
          >
            All assignees
          </button>
          <button
            v-for="member in filteredMembers"
            :key="member.user_id"
            class="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground"
            @click="selectAssignee(member.user_id)"
          >
            <Avatar class="size-5">
              <AvatarFallback class="text-[10px]">
                {{ member.first_name[0] }}{{ member.last_name[0] }}
              </AvatarFallback>
            </Avatar>
            {{ member.first_name }} {{ member.last_name }}
          </button>
          <p
            v-if="filteredMembers.length === 0"
            class="px-2 py-3 text-center text-xs text-muted-foreground"
          >
            No members found
          </p>
        </div>
      </PopoverContent>
    </Popover>

    <Button
      v-if="hasActiveFilters"
      variant="ghost"
      size="sm"
      @click="clearFilters"
    >
      <X class="mr-1.5 size-4" />
      Clear
    </Button>
  </div>
</template>

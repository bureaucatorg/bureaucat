<script setup lang="ts">
import { ChevronDown, Loader2, ArrowUpRight, Check } from "lucide-vue-next";
import { toast } from "vue-sonner";
import type { CycleSibling } from "~/types";

const props = defineProps<{
  projectKey: string;
  taskId: string;
  cycleId?: string;
  cycleTitle?: string;
  canEdit: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const { listSiblings, addTasksToCycle, removeTaskFromCycle } = useCycles();

const open = ref(false);
const cycles = ref<CycleSibling[]>([]);
const loadingCycles = ref(false);
const updating = ref(false);

const statusLabels: Record<string, string> = {
  active: "Active",
  upcoming: "Upcoming",
  completed: "Completed",
};

async function loadCycles() {
  loadingCycles.value = true;
  const result = await listSiblings(props.projectKey);
  if (result.success && result.data) {
    cycles.value = result.data;
  } else {
    toast.error(result.error || "Failed to load cycles");
  }
  loadingCycles.value = false;
}

watch(open, (isOpen) => {
  if (isOpen && !cycles.value.length) loadCycles();
});

async function selectCycle(cycle: CycleSibling) {
  if (cycle.id === props.cycleId) {
    open.value = false;
    return;
  }

  updating.value = true;
  // A task belongs to at most one cycle, so detach from the current one first.
  if (props.cycleId) {
    const removed = await removeTaskFromCycle(props.projectKey, props.cycleId, props.taskId);
    if (!removed.success) {
      updating.value = false;
      toast.error(removed.error || "Failed to update cycle");
      return;
    }
  }
  const result = await addTasksToCycle(props.projectKey, cycle.id, [props.taskId]);
  updating.value = false;
  open.value = false;

  if (result.success) {
    toast.success(`Added to ${cycle.title}`);
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to add task to cycle");
  }
}

async function clearCycle() {
  if (!props.cycleId) return;
  updating.value = true;
  const result = await removeTaskFromCycle(props.projectKey, props.cycleId, props.taskId);
  updating.value = false;
  open.value = false;

  if (result.success) {
    toast.success("Removed from cycle");
    emit("refresh");
  } else {
    toast.error(result.error || "Failed to remove task from cycle");
  }
}

function formatRange(cycle: CycleSibling): string {
  const fmt = (d: string) =>
    new Date(d).toLocaleDateString("en-US", { month: "short", day: "numeric" });
  return `${fmt(cycle.start_date)} – ${fmt(cycle.end_date)}`;
}
</script>

<template>
  <div class="flex items-center justify-between gap-2">
    <p class="shrink-0 text-xs text-muted-foreground">Cycle</p>

    <div class="flex min-w-0 items-center gap-1">
      <!-- Separate from the picker trigger so the title stays aligned with the
           other sidebar values. -->
      <NuxtLink
        v-if="canEdit && cycleId"
        :to="`/projects/${projectKey}/cycles/${cycleId}`"
        aria-label="Open cycle"
        class="shrink-0 text-muted-foreground/60 hover:text-foreground"
      >
        <ArrowUpRight class="size-3.5" />
      </NuxtLink>

      <DropdownMenu v-if="canEdit" v-model:open="open">
        <DropdownMenuTrigger as-child>
          <Button
            variant="ghost"
            class="h-auto min-w-0 shrink gap-1.5 px-0 py-0 font-medium hover:bg-transparent has-[>svg]:pl-0"
            :class="cycleId ? '' : 'text-muted-foreground'"
            :disabled="updating"
          >
            <Loader2 v-if="updating" class="size-3.5 animate-spin" />
            <span class="truncate" :title="cycleTitle">{{ cycleTitle || "None" }}</span>
            <ChevronDown class="size-3.5 shrink-0 opacity-50" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" class="w-56">
          <div v-if="loadingCycles" class="flex items-center justify-center py-3">
            <Loader2 class="size-4 animate-spin text-muted-foreground" />
          </div>
          <template v-else>
            <p
              v-if="!cycles.length"
              class="px-2 py-3 text-center text-xs text-muted-foreground"
            >
              No cycles in this project
            </p>
            <DropdownMenuItem
              v-for="cycle in cycles"
              :key="cycle.id"
              class="gap-2"
              @click="selectCycle(cycle)"
            >
              <Check
                class="size-3.5 shrink-0"
                :class="cycle.id === cycleId ? 'opacity-100' : 'opacity-0'"
              />
              <span class="min-w-0 flex-1">
                <span class="block truncate">{{ cycle.title }}</span>
                <span class="block text-xs text-muted-foreground">
                  {{ statusLabels[cycle.status] || cycle.status }} · {{ formatRange(cycle) }}
                </span>
              </span>
            </DropdownMenuItem>
            <template v-if="cycleId">
              <DropdownMenuSeparator />
              <DropdownMenuItem class="text-destructive" @click="clearCycle">
                Remove from cycle
              </DropdownMenuItem>
            </template>
          </template>
        </DropdownMenuContent>
      </DropdownMenu>

      <NuxtLink
        v-else-if="cycleId"
        :to="`/projects/${projectKey}/cycles/${cycleId}`"
        class="min-w-0 truncate text-sm font-medium hover:underline"
        :title="cycleTitle"
      >
        {{ cycleTitle }}
      </NuxtLink>
      <span v-else class="text-sm text-muted-foreground">None</span>
    </div>
  </div>
</template>

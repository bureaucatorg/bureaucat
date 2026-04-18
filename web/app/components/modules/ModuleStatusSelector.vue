<script setup lang="ts">
import {
  Circle,
  CircleDot,
  CheckCircle2,
  XCircle,
  Clock,
  PauseCircle,
  ChevronDown,
} from "lucide-vue-next";
import type { ModuleStatus } from "~/types";
import { MODULE_STATUSES } from "~/types";

withDefaults(
  defineProps<{
    disabled?: boolean;
    compact?: boolean;
  }>(),
  { compact: false }
);

const model = defineModel<ModuleStatus>({ required: true });

interface StatusMeta {
  icon: typeof Circle;
  color: string;
  label: string;
}

const STATUS_META: Record<ModuleStatus, StatusMeta> = {
  backlog:     { icon: Clock,         color: "#6B7280", label: "Backlog" },
  planned:     { icon: Circle,        color: "#0EA5E9", label: "Planned" },
  in_progress: { icon: CircleDot,     color: "#F59E0B", label: "In progress" },
  paused:      { icon: PauseCircle,   color: "#F97316", label: "Paused" },
  completed:   { icon: CheckCircle2,  color: "#10B981", label: "Completed" },
  cancelled:   { icon: XCircle,       color: "#F43F5E", label: "Cancelled" },
};

const current = computed(() => STATUS_META[model.value] ?? STATUS_META.backlog);
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button
        type="button"
        :variant="compact ? 'ghost' : 'outline'"
        :class="
          compact
            ? 'h-auto gap-1.5 px-0 py-0 font-medium hover:bg-transparent'
            : 'w-full justify-between'
        "
        :disabled="disabled"
      >
        <span class="flex items-center gap-1.5">
          <component
            :is="current.icon"
            :class="compact ? 'size-5 stroke-[2.5]' : 'size-4'"
            :style="{ color: current.color }"
          />
          {{ current.label }}
        </span>
        <ChevronDown class="size-3.5 opacity-50" />
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent class="w-56">
      <DropdownMenuItem
        v-for="s in MODULE_STATUSES"
        :key="s"
        @click="model = s"
      >
        <component
          :is="STATUS_META[s].icon"
          class="mr-2 size-4"
          :style="{ color: STATUS_META[s].color }"
        />
        {{ STATUS_META[s].label }}
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>

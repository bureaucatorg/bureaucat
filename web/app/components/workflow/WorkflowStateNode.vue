<script setup lang="ts">
import { Handle, Position } from "@vue-flow/core";
import { AlertTriangle, LogIn, UserMinus, UserPlus } from "lucide-vue-next";
import type { ProjectState, WorkflowStateConfig } from "~/types";

defineProps<{
  data: { state?: ProjectState; config: WorkflowStateConfig; warning?: string; selected: boolean };
  connectable: boolean;
}>();

const centerHandle = { top: "50%", left: "50%", transform: "translate(-50%, -50%)", opacity: 0 };
</script>

<template>
  <div
    class="group relative flex min-w-36 cursor-pointer flex-col gap-1 rounded-md border bg-card px-3 py-2 text-card-foreground shadow-sm transition-colors hover:border-foreground/30"
    :class="data.selected && 'border-foreground ring-1 ring-foreground'"
  >
    <div class="flex items-center gap-2">
      <span class="size-2.5 shrink-0 rounded-full" :style="{ backgroundColor: data.state?.color }" />
      <span class="truncate text-sm font-medium">{{ data.state?.name }}</span>
      <span v-if="data.warning" class="ml-auto" :title="data.warning">
        <AlertTriangle class="size-3.5 shrink-0 text-amber-500" />
      </span>
    </div>
    <div class="flex items-center gap-2 text-[11px] text-muted-foreground">
      <span class="capitalize">{{ data.state?.state_type }}</span>
      <span v-if="data.config.is_entry" class="flex items-center gap-0.5"><LogIn class="size-3" />Entry</span>
      <span v-if="data.config.add_assignees.length" class="flex items-center gap-0.5">
        <UserPlus class="size-3" />{{ data.config.add_assignees.length }}
      </span>
      <span v-if="data.config.remove_assignees.length" class="flex items-center gap-0.5">
        <UserMinus class="size-3" />{{ data.config.remove_assignees.length }}
      </span>
    </div>
    <Handle type="target" :position="Position.Left" :connectable="connectable" :style="centerHandle" />
    <Handle
      v-if="connectable"
      type="source"
      :position="Position.Right"
      class="!size-3 !border-2 !border-background !bg-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
    />
  </div>
</template>

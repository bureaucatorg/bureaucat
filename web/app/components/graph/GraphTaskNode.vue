<script setup lang="ts">
import { Handle, Position } from "@vue-flow/core";
import type { GraphTask } from "~/composables/useTaskGraph";

defineProps<{ data: GraphTask }>();

const centerHandle = { top: "50%", left: "50%", transform: "translate(-50%, -50%)", opacity: 0 };
</script>

<template>
  <div
    class="flex cursor-pointer items-center gap-1.5 rounded-md border bg-card px-2.5 py-1 font-mono text-xs font-medium text-card-foreground shadow-sm transition-colors hover:border-foreground/30"
  >
    <span class="size-2 shrink-0 rounded-full bg-muted-foreground" :style="data.state_color ? { backgroundColor: data.state_color } : undefined" />
    {{ data.project_key }}-{{ data.task_number }}
    <Handle type="target" :position="Position.Top" :connectable="false" :style="centerHandle" />
  </div>
</template>

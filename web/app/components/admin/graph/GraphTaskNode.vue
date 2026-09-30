<script setup lang="ts">
import { Handle, Position } from "@vue-flow/core";
import { FolderKanban } from "lucide-vue-next";
import { Tooltip, TooltipContent, TooltipTrigger } from "~/components/ui/tooltip";
import type { GraphTask } from "~/composables/useAdmin";

defineProps<{ data: GraphTask }>();

const centerHandle = { top: "50%", left: "50%", transform: "translate(-50%, -50%)", opacity: 0 };
</script>

<template>
  <Tooltip>
    <TooltipTrigger as-child>
      <div
        class="flex cursor-pointer items-center gap-1.5 rounded-md border bg-card px-2.5 py-1 font-mono text-xs font-medium text-card-foreground shadow-sm transition-colors hover:border-foreground/30"
      >
        <span class="size-2 shrink-0 rounded-full bg-muted-foreground" :style="data.state_color ? { backgroundColor: data.state_color } : undefined" />
        {{ data.project_key }}-{{ data.task_number }}
        <Handle type="target" :position="Position.Top" :connectable="false" :style="centerHandle" />
      </div>
    </TooltipTrigger>
    <TooltipContent
      side="top"
      :side-offset="10"
      class="w-80 rounded-lg border bg-popover p-0 text-sm text-popover-foreground shadow-lg [&_.rotate-45]:hidden"
    >
      <div class="space-y-2 p-3">
        <div class="flex items-center gap-2">
          <span class="font-mono text-xs text-muted-foreground">{{ data.project_key }}-{{ data.task_number }}</span>
          <span v-if="data.is_subtask" class="rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
            Subtask
          </span>
        </div>
        <p class="font-semibold leading-snug text-balance">{{ data.title }}</p>
      </div>
      <div class="flex items-center justify-between gap-3 border-t px-3 py-2.5 text-xs">
        <span
          class="inline-flex shrink-0 items-center gap-1.5 rounded bg-muted px-1.5 py-0.5 font-medium text-muted-foreground"
          :style="data.state_color ? { backgroundColor: data.state_color + '20', color: data.state_color } : undefined"
        >
          <span class="size-2 rounded-full bg-current" />
          {{ data.state_name }}
        </span>
        <span class="flex min-w-0 items-center gap-1.5 text-muted-foreground">
          <FolderKanban class="size-3.5 shrink-0" />
          <span class="truncate">{{ data.project_name }} · {{ data.workspace_name }}</span>
        </span>
      </div>
    </TooltipContent>
  </Tooltip>
</template>

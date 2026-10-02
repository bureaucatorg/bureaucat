<script setup lang="ts">
import { FolderKanban, ListTodo, Mail } from "lucide-vue-next";
import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";
import { Tooltip, TooltipContent, TooltipTrigger } from "~/components/ui/tooltip";
import type { HoveredGraphNode } from "~/composables/useTaskGraph";

defineProps<{
  node: HoveredGraphNode | null;
  reference: HTMLElement | null;
}>();
</script>

<template>
  <Tooltip :open="!!node && !!reference" disable-hoverable-content>
    <TooltipTrigger as-child :reference="reference ?? undefined">
      <span class="hidden" />
    </TooltipTrigger>
    <TooltipContent
      v-if="node?.type === 'user'"
      side="top"
      :side-offset="10"
      class="pointer-events-none w-72 rounded-lg border bg-popover p-0 text-sm text-popover-foreground shadow-lg [&_.rotate-45]:hidden!"
    >
      <div class="flex items-center gap-3 p-3">
        <Avatar class="size-10">
          <AvatarImage v-if="node.data.avatar_url" :src="node.data.avatar_url" />
          <AvatarFallback class="text-sm font-medium" :seed="node.data.id">
            {{ node.data.first_name[0] }}{{ node.data.last_name[0] }}
          </AvatarFallback>
        </Avatar>
        <div class="min-w-0">
          <p class="truncate font-semibold">{{ node.data.first_name }} {{ node.data.last_name }}</p>
          <p class="truncate text-xs text-muted-foreground">@{{ node.data.username }}</p>
        </div>
      </div>
      <div class="space-y-1.5 border-t px-3 py-2.5 text-xs text-muted-foreground">
        <p class="flex items-center gap-2">
          <Mail class="size-3.5 shrink-0" />
          <span class="truncate">{{ node.data.email }}</span>
        </p>
        <p class="flex items-center gap-2">
          <ListTodo class="size-3.5 shrink-0" />
          {{ node.data.task_count }} {{ node.data.task_count === 1 ? "task" : "tasks" }} in view
        </p>
      </div>
      <p class="border-t px-3 py-2 text-[11px] text-muted-foreground">
        Click to highlight · Double-click to open profile
      </p>
    </TooltipContent>
    <TooltipContent
      v-else-if="node?.type === 'task'"
      side="top"
      :side-offset="10"
      class="pointer-events-none w-80 rounded-lg border bg-popover p-0 text-sm text-popover-foreground shadow-lg [&_.rotate-45]:hidden!"
    >
      <div class="space-y-2 p-3">
        <div class="flex items-center gap-2">
          <span class="font-mono text-xs text-muted-foreground">{{ node.data.project_key }}-{{ node.data.task_number }}</span>
          <span v-if="node.data.is_subtask" class="rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
            Subtask
          </span>
        </div>
        <p class="font-semibold leading-snug text-balance">{{ node.data.title }}</p>
      </div>
      <div class="flex items-center justify-between gap-3 border-t px-3 py-2.5 text-xs">
        <span
          class="inline-flex shrink-0 items-center gap-1.5 rounded bg-muted px-1.5 py-0.5 font-medium text-muted-foreground"
          :style="node.data.state_color ? { backgroundColor: node.data.state_color + '20', color: node.data.state_color } : undefined"
        >
          <span class="size-2 rounded-full bg-current" />
          {{ node.data.state_name }}
        </span>
        <span class="flex min-w-0 items-center gap-1.5 text-muted-foreground">
          <FolderKanban class="size-3.5 shrink-0" />
          <span class="truncate">{{ node.data.project_name }} · {{ node.data.workspace_name }}</span>
        </span>
      </div>
      <p class="border-t px-3 py-2 text-[11px] text-muted-foreground">
        Click to highlight · Double-click to open task
      </p>
    </TooltipContent>
  </Tooltip>
</template>

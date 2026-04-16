<script setup lang="ts">
import { LayoutGrid, Check } from "lucide-vue-next";
import type { ViewGroupBy } from "~/types";

defineProps<{
  modelValue: ViewGroupBy;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: ViewGroupBy];
}>();

const OPTIONS: { id: ViewGroupBy; label: string; hint?: string }[] = [
  { id: "state_type", label: "Status category", hint: "Backlog / To Do / In Progress / Done" },
  { id: "state", label: "State" },
  { id: "priority", label: "Priority" },
  { id: "assignee", label: "Assignee" },
  { id: "label", label: "Label", hint: "Tasks appear in every column they're labelled with" },
  { id: "due_bucket", label: "Due date", hint: "Overdue / Today / This week / Later" },
];
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button variant="outline" size="sm" class="gap-1.5">
        <LayoutGrid class="size-3.5" />
        Group by:
        <span class="font-medium">
          {{ OPTIONS.find((o) => o.id === modelValue)?.label ?? modelValue }}
        </span>
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="start" class="w-60">
      <DropdownMenuLabel class="text-xs uppercase tracking-wider text-muted-foreground">
        Group board by
      </DropdownMenuLabel>
      <DropdownMenuItem
        v-for="opt in OPTIONS"
        :key="opt.id"
        class="flex items-start justify-between gap-2"
        @click="emit('update:modelValue', opt.id)"
      >
        <div class="min-w-0">
          <div class="text-sm">{{ opt.label }}</div>
          <div v-if="opt.hint" class="truncate text-xs text-muted-foreground">{{ opt.hint }}</div>
        </div>
        <Check v-if="modelValue === opt.id" class="mt-0.5 size-3.5 shrink-0 text-primary" />
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>

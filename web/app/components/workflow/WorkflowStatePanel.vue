<script setup lang="ts">
import { AlertTriangle } from "lucide-vue-next";
import type { ProjectMember, ProjectState, WorkflowStateConfig } from "~/types";

// config is the editor's draft object; edits here change the draft in place.
const props = defineProps<{
  config: WorkflowStateConfig;
  state?: ProjectState;
  members: ProjectMember[];
  warning?: string;
  editable: boolean;
}>();

const config = computed(() => props.config);
</script>

<template>
  <div class="space-y-5">
    <div class="space-y-1">
      <div class="flex items-center gap-2">
        <span class="size-2.5 rounded-full" :style="{ backgroundColor: state?.color }" />
        <h3 class="font-medium">{{ state?.name }}</h3>
      </div>
      <p class="text-xs capitalize text-muted-foreground">{{ state?.state_type }} state</p>
    </div>

    <p v-if="warning" class="flex items-start gap-2 text-xs text-amber-600 dark:text-amber-400">
      <AlertTriangle class="mt-0.5 size-3.5 shrink-0" />
      {{ warning }}
    </p>

    <label class="flex items-start gap-2 text-sm">
      <Checkbox v-model="config.is_entry" :disabled="!editable" class="mt-0.5" />
      <span>
        Entry state
        <span class="block text-xs text-muted-foreground">New tasks can start here.</span>
      </span>
    </label>

    <div class="space-y-2">
      <MemberSelector
        v-model="config.add_assignees"
        :members="members"
        multi
        label="Assign on entry"
        add-label="Add person"
        empty-label="Nobody"
        :disabled="!editable"
      />
      <p class="text-xs text-muted-foreground">Added alongside the task's current assignees.</p>
    </div>

    <div class="space-y-2">
      <MemberSelector
        v-model="config.remove_assignees"
        :members="members"
        multi
        label="Unassign on entry"
        add-label="Add person"
        empty-label="Nobody"
        :disabled="!editable"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ArrowRight, Trash2 } from "lucide-vue-next";
import type { ProjectMember, ProjectState, WorkflowTransitionConfig } from "~/types";

// transition is the editor's draft object; edits here change the draft in place.
const props = defineProps<{
  transition: WorkflowTransitionConfig;
  from?: ProjectState;
  to?: ProjectState;
  members: ProjectMember[];
  editable: boolean;
}>();

const emit = defineEmits<{ remove: [] }>();

const transition = computed(() => props.transition);

const open = computed(
  () => transition.value.mover_ids.length === 0 && !transition.value.allow_assignees && !transition.value.allow_creator
);
</script>

<template>
  <div class="space-y-5">
    <div class="space-y-1">
      <h3 class="flex flex-wrap items-center gap-1.5 font-medium">
        <span class="size-2.5 rounded-full" :style="{ backgroundColor: from?.color }" />
        {{ from?.name }}
        <ArrowRight class="size-3.5 text-muted-foreground" />
        <span class="size-2.5 rounded-full" :style="{ backgroundColor: to?.color }" />
        {{ to?.name }}
      </h3>
      <p class="text-xs text-muted-foreground">
        {{ open ? "Any project member can make this move." : "Only the people below can make this move." }}
        Project admins always can.
      </p>
    </div>

    <MemberSelector
      v-model="transition.mover_ids"
      :members="members"
      multi
      label="People"
      add-label="Add person"
      empty-label="No one specific"
      :disabled="!editable"
    />

    <div class="space-y-3">
      <label class="flex items-center gap-2 text-sm">
        <Checkbox v-model="transition.allow_assignees" :disabled="!editable" />
        The task's assignees
      </label>
      <label class="flex items-center gap-2 text-sm">
        <Checkbox v-model="transition.allow_creator" :disabled="!editable" />
        The task's creator
      </label>
    </div>

    <Button v-if="editable" variant="outline" size="sm" class="gap-1.5 text-destructive" @click="emit('remove')">
      <Trash2 class="size-3.5" />
      Remove transition
    </Button>
  </div>
</template>

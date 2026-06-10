<script setup lang="ts">
import { Plus, X } from "lucide-vue-next";
import type { ProjectMember } from "~/types";

// Presentational picker that mirrors the TaskAssignees UX (avatar pills +
// "+ Add" popover with search). Single-select mode emits a string (or null);
// multi-select mode emits a string[]. Works on local form state — no API calls.

const props = defineProps<{
  members: ProjectMember[];
  multi?: boolean;
  label?: string;
  addLabel?: string;
  emptyLabel?: string;
  disabled?: boolean;
}>();

const model = defineModel<string | string[] | null>({ default: null });

const showPopover = ref(false);

const selectedIds = computed<string[]>(() => {
  if (props.multi) return (model.value as string[] | null) ?? [];
  const v = model.value as string | null;
  return v ? [v] : [];
});

const membersById = computed(() => {
  const map = new Map<string, ProjectMember>();
  for (const m of props.members) map.set(m.user_id, m);
  return map;
});

const selectedMembers = computed<ProjectMember[]>(() =>
  selectedIds.value
    .map((id) => membersById.value.get(id))
    .filter((m): m is ProjectMember => !!m)
);

const availableMembers = computed(() => {
  const taken = new Set(selectedIds.value);
  return props.members.filter((m) => !taken.has(m.user_id));
});

// The pool offered in the popover: in multi mode, exclude already-picked
// members; in single mode, show everyone (the current pick is replaced).
const memberPool = computed(() =>
  props.multi ? availableMembers.value : props.members
);

function memberSearchText(m: ProjectMember) {
  return `${m.first_name} ${m.last_name} ${m.username}`;
}

function add(userId: string) {
  if (props.multi) {
    const current = (model.value as string[] | null) ?? [];
    if (!current.includes(userId)) {
      model.value = [...current, userId];
    }
  } else {
    model.value = userId;
  }
  showPopover.value = false;
}

function remove(userId: string) {
  if (props.multi) {
    const current = (model.value as string[] | null) ?? [];
    model.value = current.filter((id) => id !== userId);
  } else {
    model.value = null;
  }
}

const canShowAdd = computed(() => {
  if (props.disabled) return false;
  if (props.multi) return availableMembers.value.length > 0;
  // single-select: once a pick is made, the "+ Add" is replaced by the pill
  return selectedIds.value.length === 0;
});
</script>

<template>
  <div class="space-y-2">
    <Label v-if="label">{{ label }}</Label>

    <div class="flex flex-wrap items-center gap-2">
      <div
        v-for="m in selectedMembers"
        :key="m.user_id"
        class="group relative flex items-center gap-1.5 rounded-md border bg-muted/50 py-1 pl-1 pr-2.5"
      >
        <div class="flex items-center gap-1.5">
          <Avatar class="size-6">
            <AvatarImage v-if="m.avatar_url" :src="m.avatar_url" />
            <AvatarFallback class="text-xs" :seed="m.user_id">
              {{ m.first_name[0] }}{{ m.last_name[0] }}
            </AvatarFallback>
          </Avatar>
          <span class="text-sm">{{ m.first_name }} {{ m.last_name }}</span>
        </div>
        <button
          v-if="!disabled"
          type="button"
          :aria-label="`Remove ${m.first_name} ${m.last_name}`"
          class="absolute -top-1.5 -right-1.5 flex size-4 items-center justify-center rounded-full bg-foreground text-background opacity-0 shadow-sm transition-opacity group-hover:opacity-100 focus:opacity-100 focus-visible:ring-2 focus-visible:ring-ring outline-none"
          @click="remove(m.user_id)"
        >
          <X class="size-2.5" />
        </button>
      </div>

      <SearchableSelect
        v-if="canShowAdd"
        v-model:open="showPopover"
        :items="memberPool"
        :get-search-text="memberSearchText"
        :get-key="(m) => m.user_id"
        placeholder="Search members..."
        empty-text="No members found"
        @select="(m) => add(m.user_id)"
      >
        <template #trigger>
          <Button type="button" variant="outline" size="sm" class="h-8 gap-1.5">
            <Plus class="size-3.5" />
            {{ addLabel || (multi ? "Add" : "Pick") }}
          </Button>
        </template>
        <template #option="{ item: m }">
          <Avatar class="size-6">
            <AvatarImage v-if="m.avatar_url" :src="m.avatar_url" />
            <AvatarFallback class="text-xs" :seed="m.user_id">
              {{ m.first_name[0] }}{{ m.last_name[0] }}
            </AvatarFallback>
          </Avatar>
          {{ m.first_name }} {{ m.last_name }}
        </template>
      </SearchableSelect>

      <span
        v-if="selectedMembers.length === 0 && !canShowAdd"
        class="text-sm text-muted-foreground"
      >
        {{ emptyLabel || "None" }}
      </span>
    </div>
  </div>
</template>

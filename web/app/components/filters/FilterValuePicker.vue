<script setup lang="ts">
import type { FilterField, FilterOp, FilterValue, ProjectState, ProjectMember, ProjectLabel } from "~/types";
import type { ValueKind } from "./filterCatalog";
import {
  findOpDef,
  STATE_TYPE_OPTIONS,
  PRIORITY_OPTIONS,
  RELATIVE_DATE_OPTIONS,
} from "./filterCatalog";
import EntityMultiSelect from "~/components/shared/EntityMultiSelect.vue";

const props = defineProps<{
  field: FilterField;
  op: FilterOp;
  value: FilterValue | undefined;
  states: ProjectState[];
  labels: ProjectLabel[];
  members: ProjectMember[];
}>();

const emit = defineEmits<{
  "update:value": [value: FilterValue | undefined];
}>();

const valueKind = computed<ValueKind>(() => findOpDef(props.field, props.op)?.valueKind ?? "none");

// ----- helpers to coerce the polymorphic value into the right shape -----

const asStringArray = computed<string[]>(() => {
  return Array.isArray(props.value) ? (props.value as (string | number)[]).map(String) : [];
});

const asText = computed<string>(() => (typeof props.value === "string" ? props.value : ""));

const asNumber = computed<string>(() => (typeof props.value === "number" ? String(props.value) : ""));

const asDateValue = computed<string>(() => (typeof props.value === "string" ? props.value : ""));

const asRange = computed<{ from: string; to: string }>(() => {
  if (props.value && typeof props.value === "object" && !Array.isArray(props.value)) {
    return props.value as { from: string; to: string };
  }
  return { from: "", to: "" };
});

// ----- entity pickers -----

function updateStringArray(next: string[]) {
  emit("update:value", next);
}
function updateIntArray(next: string[]) {
  emit("update:value", next.map((s) => parseInt(s, 10)).filter((n) => !isNaN(n)));
}
</script>

<template>
  <div class="w-full">
    <!-- text -->
    <div v-if="valueKind === 'text'" class="p-2">
      <Input
        :model-value="asText"
        placeholder="Enter text"
        class="h-8 text-sm"
        @update:model-value="(v) => emit('update:value', String(v ?? ''))"
      />
    </div>

    <!-- number -->
    <div v-else-if="valueKind === 'number'" class="p-2">
      <Input
        :model-value="asNumber"
        type="number"
        placeholder="0"
        class="h-8 text-sm"
        @update:model-value="(v) => emit('update:value', v === '' || v === undefined ? undefined : Number(v))"
      />
    </div>

    <!-- state (uuid-array of project_states) -->
    <EntityMultiSelect
      v-else-if="valueKind === 'uuid-array' && field === 'state'"
      :items="states"
      :model-value="asStringArray"
      placeholder="Find state…"
      empty-message="No states"
      @update:model-value="updateStringArray"
    >
      <template #option="{ item }">
        <span
          class="size-2 rounded-full"
          :style="{ backgroundColor: (item as ProjectState).color || '#6B7280' }"
        />
        <span class="truncate">{{ (item as ProjectState).name }}</span>
      </template>
    </EntityMultiSelect>

    <!-- state_type (string-array enum) -->
    <EntityMultiSelect
      v-else-if="valueKind === 'string-array' && field === 'state_type'"
      :items="STATE_TYPE_OPTIONS"
      :model-value="asStringArray"
      item-key="id"
      placeholder="Find status…"
      @update:model-value="updateStringArray"
    >
      <template #option="{ item }">{{ (item as { label: string }).label }}</template>
    </EntityMultiSelect>

    <!-- priority (int-array) -->
    <EntityMultiSelect
      v-else-if="valueKind === 'int-array'"
      :items="PRIORITY_OPTIONS"
      :model-value="asStringArray"
      item-key="id"
      placeholder="Find priority…"
      @update:model-value="updateIntArray"
    >
      <template #option="{ item }">
        <span
          class="size-2 rounded-full"
          :style="{ backgroundColor: (item as { color: string }).color }"
        />
        <span>{{ (item as { label: string }).label }}</span>
      </template>
    </EntityMultiSelect>

    <!-- assignees / created_by (uuid-array of members) -->
    <EntityMultiSelect
      v-else-if="valueKind === 'uuid-array' && (field === 'assignees' || field === 'created_by')"
      :items="[{ user_id: '@me', first_name: 'Me', last_name: '', username: 'me', email: '' }, ...members]"
      :model-value="asStringArray"
      item-key="user_id"
      placeholder="Find member…"
      empty-message="No members"
      @update:model-value="updateStringArray"
    >
      <template #option="{ item }">
        <Avatar class="size-5">
          <AvatarFallback class="text-[10px]" :seed="String((item as ProjectMember).user_id)">
            {{ (item as ProjectMember).first_name?.[0] || '?' }}{{ (item as ProjectMember).last_name?.[0] || '' }}
          </AvatarFallback>
        </Avatar>
        <span class="truncate">
          {{ (item as ProjectMember).first_name }} {{ (item as ProjectMember).last_name }}
        </span>
      </template>
    </EntityMultiSelect>

    <!-- labels (uuid-array) -->
    <EntityMultiSelect
      v-else-if="valueKind === 'uuid-array' && field === 'labels'"
      :items="labels"
      :model-value="asStringArray"
      placeholder="Find label…"
      empty-message="No labels"
      @update:model-value="updateStringArray"
    >
      <template #option="{ item }">
        <span
          class="size-3 rounded"
          :style="{ backgroundColor: (item as ProjectLabel).color || '#3B82F6' }"
        />
        <span class="truncate">{{ (item as ProjectLabel).name }}</span>
      </template>
    </EntityMultiSelect>

    <!-- date (single keyword or ISO) -->
    <div v-else-if="valueKind === 'date'" class="space-y-1 p-2">
      <Input
        :model-value="asDateValue"
        type="text"
        placeholder="YYYY-MM-DD or today, this_week…"
        class="h-8 text-sm"
        @update:model-value="(v) => emit('update:value', String(v ?? ''))"
      />
      <div class="flex flex-wrap gap-1 pt-1">
        <button
          v-for="opt in RELATIVE_DATE_OPTIONS"
          :key="opt.id"
          type="button"
          class="rounded-full border px-2 py-0.5 text-xs text-muted-foreground transition hover:border-primary hover:text-primary"
          :class="asDateValue === opt.id ? 'border-primary bg-primary/10 text-primary' : ''"
          @click="emit('update:value', opt.id)"
        >
          {{ opt.label }}
        </button>
      </div>
    </div>

    <!-- date range -->
    <div v-else-if="valueKind === 'date-range'" class="space-y-2 p-2">
      <div class="space-y-1">
        <Label class="text-xs text-muted-foreground">From</Label>
        <Input
          :model-value="asRange.from"
          type="text"
          placeholder="YYYY-MM-DD or keyword"
          class="h-8 text-sm"
          @update:model-value="(v) => emit('update:value', { from: String(v ?? ''), to: asRange.to })"
        />
      </div>
      <div class="space-y-1">
        <Label class="text-xs text-muted-foreground">To</Label>
        <Input
          :model-value="asRange.to"
          type="text"
          placeholder="YYYY-MM-DD or keyword"
          class="h-8 text-sm"
          @update:model-value="(v) => emit('update:value', { from: asRange.from, to: String(v ?? '') })"
        />
      </div>
    </div>

    <!-- none -->
    <p v-else class="px-3 py-3 text-xs text-muted-foreground">
      No value needed.
    </p>
  </div>
</template>

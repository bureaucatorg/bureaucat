<script setup lang="ts">
import type { DateValue } from "reka-ui";
import { CalendarDate } from "@internationalized/date";
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

// ----- date picker helpers -----

/** Check if a string is a relative date keyword (not an ISO date). */
function isRelativeDate(v: string): boolean {
  return RELATIVE_DATE_OPTIONS.some((o) => o.id === v);
}

/** Parse an ISO date string (YYYY-MM-DD) into a CalendarDate, or return undefined. */
function parseCalendarDate(v: string): DateValue | undefined {
  if (!v || isRelativeDate(v)) return undefined;
  const m = v.match(/^(\d{4})-(\d{2})-(\d{2})$/);
  if (!m) return undefined;
  return new CalendarDate(parseInt(m[1]), parseInt(m[2]), parseInt(m[3]));
}

/** Format a CalendarDate to ISO date string. */
function calendarDateToString(d: DateValue): string {
  const y = String(d.year).padStart(4, "0");
  const m = String(d.month).padStart(2, "0");
  const day = String(d.day).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

const calendarValue = computed<DateValue | undefined>(() => parseCalendarDate(asDateValue.value));
const calendarFromValue = computed<DateValue | undefined>(() => parseCalendarDate(asRange.value.from));
const calendarToValue = computed<DateValue | undefined>(() => parseCalendarDate(asRange.value.to));

/** Which section is active: 'calendar' for picking a date, 'relative' for keywords. */
const dateMode = ref<"calendar" | "relative">(
  asDateValue.value && isRelativeDate(asDateValue.value) ? "relative" : "calendar"
);

// For date-range, track which field is picking: 'from' or 'to'
const rangeDateMode = ref<{ from: "calendar" | "relative"; to: "calendar" | "relative" }>({
  from: asRange.value.from && isRelativeDate(asRange.value.from) ? "relative" : "calendar",
  to: asRange.value.to && isRelativeDate(asRange.value.to) ? "relative" : "calendar",
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
    <div v-else-if="valueKind === 'date'" class="w-full">
      <!-- Tabs: Calendar / Relative -->
      <div class="flex border-b">
        <button
          type="button"
          class="flex-1 px-3 py-1.5 text-xs font-medium transition-colors"
          :class="dateMode === 'calendar' ? 'border-b-2 border-primary text-primary' : 'text-muted-foreground hover:text-foreground'"
          @click="dateMode = 'calendar'"
        >
          Pick date
        </button>
        <button
          type="button"
          class="flex-1 px-3 py-1.5 text-xs font-medium transition-colors"
          :class="dateMode === 'relative' ? 'border-b-2 border-primary text-primary' : 'text-muted-foreground hover:text-foreground'"
          @click="dateMode = 'relative'"
        >
          Relative
        </button>
      </div>

      <div v-if="dateMode === 'calendar'" class="flex justify-center p-1">
        <Calendar
          :model-value="calendarValue"
          layout="month-and-year"
          class="p-2"
          @update:model-value="(d: DateValue | undefined) => { if (d) emit('update:value', calendarDateToString(d)) }"
        />
      </div>

      <div v-else class="max-h-52 overflow-y-auto p-2">
        <div class="grid grid-cols-2 gap-1">
          <button
            v-for="opt in RELATIVE_DATE_OPTIONS"
            :key="opt.id"
            type="button"
            class="rounded-md px-2.5 py-1.5 text-left text-xs transition-colors hover:bg-accent hover:text-accent-foreground"
            :class="asDateValue === opt.id ? 'bg-primary/10 font-medium text-primary' : 'text-muted-foreground'"
            @click="emit('update:value', opt.id)"
          >
            {{ opt.label }}
          </button>
        </div>
      </div>

      <!-- Show current selection -->
      <div v-if="asDateValue" class="border-t px-3 py-1.5">
        <p class="text-xs text-muted-foreground">
          Selected: <span class="font-medium text-foreground">{{ asDateValue }}</span>
        </p>
      </div>
    </div>

    <!-- date range -->
    <div v-else-if="valueKind === 'date-range'" class="w-full">
      <!-- From date -->
      <div class="border-b">
        <div class="flex items-center justify-between bg-muted/30 px-3 py-1.5">
          <Label class="text-xs font-medium">From</Label>
          <div class="flex gap-1">
            <button
              type="button"
              class="rounded px-1.5 py-0.5 text-[10px] font-medium transition-colors"
              :class="rangeDateMode.from === 'calendar' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:text-foreground'"
              @click="rangeDateMode.from = 'calendar'"
            >
              Date
            </button>
            <button
              type="button"
              class="rounded px-1.5 py-0.5 text-[10px] font-medium transition-colors"
              :class="rangeDateMode.from === 'relative' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:text-foreground'"
              @click="rangeDateMode.from = 'relative'"
            >
              Relative
            </button>
          </div>
        </div>
        <div v-if="rangeDateMode.from === 'calendar'" class="flex justify-center p-1">
          <Calendar
            :model-value="calendarFromValue"
            layout="month-and-year"
            class="p-2"
            @update:model-value="(d: DateValue | undefined) => { if (d) emit('update:value', { from: calendarDateToString(d), to: asRange.to }) }"
          />
        </div>
        <div v-else class="max-h-36 overflow-y-auto p-2">
          <div class="grid grid-cols-2 gap-1">
            <button
              v-for="opt in RELATIVE_DATE_OPTIONS"
              :key="opt.id"
              type="button"
              class="rounded-md px-2 py-1 text-left text-xs transition-colors hover:bg-accent"
              :class="asRange.from === opt.id ? 'bg-primary/10 font-medium text-primary' : 'text-muted-foreground'"
              @click="emit('update:value', { from: opt.id, to: asRange.to })"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>
        <div v-if="asRange.from" class="border-t px-3 py-1">
          <p class="text-[11px] text-muted-foreground">From: <span class="font-medium text-foreground">{{ asRange.from }}</span></p>
        </div>
      </div>

      <!-- To date -->
      <div>
        <div class="flex items-center justify-between bg-muted/30 px-3 py-1.5">
          <Label class="text-xs font-medium">To</Label>
          <div class="flex gap-1">
            <button
              type="button"
              class="rounded px-1.5 py-0.5 text-[10px] font-medium transition-colors"
              :class="rangeDateMode.to === 'calendar' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:text-foreground'"
              @click="rangeDateMode.to = 'calendar'"
            >
              Date
            </button>
            <button
              type="button"
              class="rounded px-1.5 py-0.5 text-[10px] font-medium transition-colors"
              :class="rangeDateMode.to === 'relative' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:text-foreground'"
              @click="rangeDateMode.to = 'relative'"
            >
              Relative
            </button>
          </div>
        </div>
        <div v-if="rangeDateMode.to === 'calendar'" class="flex justify-center p-1">
          <Calendar
            :model-value="calendarToValue"
            layout="month-and-year"
            class="p-2"
            @update:model-value="(d: DateValue | undefined) => { if (d) emit('update:value', { from: asRange.from, to: calendarDateToString(d) }) }"
          />
        </div>
        <div v-else class="max-h-36 overflow-y-auto p-2">
          <div class="grid grid-cols-2 gap-1">
            <button
              v-for="opt in RELATIVE_DATE_OPTIONS"
              :key="opt.id"
              type="button"
              class="rounded-md px-2 py-1 text-left text-xs transition-colors hover:bg-accent"
              :class="asRange.to === opt.id ? 'bg-primary/10 font-medium text-primary' : 'text-muted-foreground'"
              @click="emit('update:value', { from: asRange.from, to: opt.id })"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>
        <div v-if="asRange.to" class="border-t px-3 py-1">
          <p class="text-[11px] text-muted-foreground">To: <span class="font-medium text-foreground">{{ asRange.to }}</span></p>
        </div>
      </div>
    </div>

    <!-- none -->
    <p v-else class="px-3 py-3 text-xs text-muted-foreground">
      No value needed.
    </p>
  </div>
</template>

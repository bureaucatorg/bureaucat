<script setup lang="ts">
import { Calendar as CalendarIcon, ChevronDown, X } from "lucide-vue-next";
import { CalendarDate, type DateValue } from "@internationalized/date";

const props = withDefaults(
  defineProps<{
    id?: string;
    placeholder?: string;
    defaultTime?: string;
    min?: string;
    max?: string;
    disabled?: boolean;
  }>(),
  { placeholder: "Set date", defaultTime: "09:00" }
);

const model = defineModel<string | undefined>();

const open = ref(false);
const dateDraft = ref<DateValue | undefined>();
const timeDraft = ref(props.defaultTime);

function isoToCalendarDate(iso: string | undefined): DateValue | undefined {
  if (!iso) return undefined;
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return undefined;
  return new CalendarDate(d.getFullYear(), d.getMonth() + 1, d.getDate());
}

function isoToTimeInput(iso: string | undefined): string {
  if (!iso) return props.defaultTime;
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return props.defaultTime;
  return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

const minValue = computed(() => isoToCalendarDate(props.min));
const maxValue = computed(() => isoToCalendarDate(props.max));

watch(open, (isOpen) => {
  if (!isOpen) return;
  dateDraft.value = isoToCalendarDate(model.value);
  timeDraft.value = isoToTimeInput(model.value);
});

function apply() {
  if (!dateDraft.value) return;
  const [hh, mm] = timeDraft.value.split(":").map((n) => parseInt(n, 10));
  const d = new Date(dateDraft.value.year, dateDraft.value.month - 1, dateDraft.value.day, hh || 0, mm || 0);
  model.value = d.toISOString();
  open.value = false;
}

function clear() {
  model.value = undefined;
  open.value = false;
}
</script>

<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <Button
        :id="id"
        variant="outline"
        class="w-full justify-start gap-2 px-3 font-normal"
        :class="!model && 'text-muted-foreground'"
        :disabled="disabled"
      >
        <CalendarIcon class="size-4 opacity-70" />
        <span class="flex-1 truncate text-left">{{ model ? formatDateTime(model) : placeholder }}</span>
        <ChevronDown class="size-4 opacity-50" />
      </Button>
    </PopoverTrigger>
    <PopoverContent class="w-auto p-0" align="start">
      <Calendar v-model="dateDraft" layout="month-and-year" :min-value="minValue" :max-value="maxValue" />
      <div class="flex items-center gap-2 border-t px-3 py-2">
        <CalendarIcon class="size-3.5 text-muted-foreground" />
        <Input v-model="timeDraft" type="time" class="h-8 flex-1 text-sm" />
        <Button size="sm" :disabled="!dateDraft" @click="apply">Set</Button>
      </div>
      <div v-if="model" class="border-t px-3 py-2">
        <Button variant="ghost" size="sm" class="w-full" @click="clear">
          <X class="mr-1.5 size-3.5" />
          Clear
        </Button>
      </div>
    </PopoverContent>
  </Popover>
</template>

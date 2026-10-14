<script setup lang="ts">
import { Calendar as CalendarIcon, ChevronDown } from "lucide-vue-next";
import type { DateValue } from "@internationalized/date";
import { STATS_PRESETS, activePresetDays, statsRangeLabel, statsToday } from "~/utils/stats";

const from = defineModel<DateValue>("from", { required: true });
const to = defineModel<DateValue>("to", { required: true });

const open = ref(false);
const maxDate = statsToday();
const activeDays = computed(() => activePresetDays(from.value, to.value));

function applyPreset(days: number) {
  to.value = maxDate.copy();
  from.value = maxDate.subtract({ days: days - 1 });
}
</script>

<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <Button variant="outline" size="sm" class="gap-2">
        <CalendarIcon class="size-4" />
        <span>{{ statsRangeLabel(from, to) }}</span>
        <ChevronDown class="size-3.5 opacity-60" />
      </Button>
    </PopoverTrigger>
    <PopoverContent align="end" class="w-auto p-0">
      <div class="flex flex-col sm:flex-row">
        <div class="flex flex-row flex-wrap gap-1 border-b p-2 sm:flex-col sm:flex-nowrap sm:border-b-0 sm:border-r">
          <Button
            v-for="p in STATS_PRESETS"
            :key="p.days"
            variant="ghost"
            size="sm"
            class="justify-start"
            :class="activeDays === p.days && 'bg-accent text-accent-foreground'"
            @click="applyPreset(p.days)"
          >
            {{ p.label }}
          </Button>
        </div>
        <div class="p-2">
          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <p class="mb-1 px-1 text-xs font-medium text-muted-foreground">From</p>
              <Calendar v-model="from" :max-value="to" layout="month-and-year" />
            </div>
            <div>
              <p class="mb-1 px-1 text-xs font-medium text-muted-foreground">To</p>
              <Calendar v-model="to" :min-value="from" :max-value="maxDate" layout="month-and-year" />
            </div>
          </div>
        </div>
      </div>
    </PopoverContent>
  </Popover>
</template>

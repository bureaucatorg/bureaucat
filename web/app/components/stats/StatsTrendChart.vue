<script setup lang="ts">
import { VisAxis, VisXYContainer, VisGroupedBar } from "@unovis/vue";
import {
  ChartContainer,
  ChartCrosshair,
  ChartTooltip,
  ChartTooltipContent,
  componentToString,
  type ChartConfig,
} from "~/components/ui/chart";
import { dayLabelFor, shortDay, yTick, type TrendSeries } from "~/utils/stats";

type Point = { day: string };

const props = defineProps<{
  title: string;
  series: TrendSeries[];
  data: Point[];
  emptyText?: string;
}>();

const config = computed<ChartConfig>(() =>
  Object.fromEntries(props.series.map((s) => [s.key, { label: s.label, color: s.color }])),
);
const value = (d: Point, key: string) => Number((d as Record<string, unknown>)[key] ?? 0);
const accessors = computed(() => props.series.map((s) => (d: Point) => value(d, s.key)));
const colors = computed(() => props.series.map((s) => s.color));
const isEmpty = computed(
  () => !!props.emptyText && !props.data.some((d) => props.series.some((s) => value(d, s.key) > 0)),
);
</script>

<template>
  <Card class="gap-3 py-4">
    <CardHeader class="flex flex-row items-center justify-between space-y-0 px-4 pb-1">
      <CardTitle class="text-base">{{ title }}</CardTitle>
      <div v-if="series.length > 1" class="flex items-center gap-3 text-xs text-muted-foreground">
        <span v-for="s in series" :key="s.key" class="flex items-center gap-1.5">
          <span class="size-2.5 rounded-full" :style="{ backgroundColor: s.color }" />
          {{ s.label }}
        </span>
      </div>
    </CardHeader>
    <CardContent class="px-4">
      <p v-if="isEmpty" class="py-12 text-center text-sm text-muted-foreground">{{ emptyText }}</p>
      <ChartContainer v-else :config="config" class="h-44 w-full">
        <VisXYContainer :data="data" :margin="{ top: 8, right: 8, bottom: 4, left: 4 }">
          <VisGroupedBar
            :x="(_d: Point, i: number) => i"
            :y="accessors"
            :color="colors"
            :rounded-corners="2"
            :bar-padding="0.15"
            :group-padding="series.length > 1 ? 0.15 : undefined"
          />
          <VisAxis
            type="x"
            :x="(_d: Point, i: number) => i"
            :tick-format="(i: number) => shortDay(data[i]?.day)"
            :num-ticks="5"
            :grid-line="false"
            :domain-line="false"
            :tick-line="false"
          />
          <VisAxis
            type="y"
            :tick-format="yTick"
            :num-ticks="4"
            :grid-line="false"
            :domain-line="false"
            :tick-line="false"
          />
          <ChartTooltip />
          <ChartCrosshair :template="componentToString(config, ChartTooltipContent, { labelFormatter: dayLabelFor(data) })" />
        </VisXYContainer>
      </ChartContainer>
    </CardContent>
  </Card>
</template>

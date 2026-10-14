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
import { categoryLabelFor, yTick, type StatsBar } from "~/utils/stats";

defineProps<{
  title: string;
  bars: StatsBar[];
  emptyText: string;
}>();

const config: ChartConfig = { count: { label: "Tasks" } };
</script>

<template>
  <Card class="gap-3 py-4">
    <CardHeader class="px-4 pb-1">
      <CardTitle class="text-base">{{ title }}</CardTitle>
    </CardHeader>
    <CardContent class="px-4">
      <ChartContainer v-if="bars.length" :config="config" class="h-48 w-full">
        <VisXYContainer :data="bars" :margin="{ top: 8, right: 8, bottom: 4, left: 4 }">
          <VisGroupedBar
            :x="(_d: StatsBar, i: number) => i"
            :y="(d: StatsBar) => d.count"
            :color="(d: StatsBar) => d.fill"
            :rounded-corners="6"
            :bar-padding="0.35"
          />
          <VisAxis
            type="x"
            :x="(_d: StatsBar, i: number) => i"
            :tick-format="(i: number) => bars[i]?.label ?? ''"
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
          <ChartCrosshair :template="componentToString(config, ChartTooltipContent, { labelFormatter: categoryLabelFor(bars) })" />
        </VisXYContainer>
      </ChartContainer>
      <p v-else class="py-12 text-center text-sm text-muted-foreground">{{ emptyText }}</p>
    </CardContent>
  </Card>
</template>

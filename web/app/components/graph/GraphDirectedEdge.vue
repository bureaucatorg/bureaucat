<script setup lang="ts">
import { BaseEdge, getStraightPath, type EdgeProps } from "@vue-flow/core";

const props = defineProps<EdgeProps<{ color: string; position?: number }>>();

// Node handles sit at node centres, so an end marker would hide under the
// target node; the arrow is drawn along the line (midpoint by default).
const path = computed(() =>
  getStraightPath({
    sourceX: props.sourceX,
    sourceY: props.sourceY,
    targetX: props.targetX,
    targetY: props.targetY,
  }),
);

const arrowTransform = computed(() => {
  const f = props.data?.position ?? 0.5;
  const mx = props.sourceX + (props.targetX - props.sourceX) * f;
  const my = props.sourceY + (props.targetY - props.sourceY) * f;
  const angle = (Math.atan2(props.targetY - props.sourceY, props.targetX - props.sourceX) * 180) / Math.PI;
  return `translate(${mx} ${my}) rotate(${angle})`;
});
</script>

<template>
  <BaseEdge :id="id" :path="path[0]" :style="style" :interaction-width="interactionWidth" />
  <path
    class="graph-edge-arrow"
    d="M -6 -5 L 6 0 L -6 5 Z"
    :transform="arrowTransform"
    :style="{ fill: data?.color, stroke: 'none' }"
  />
</template>

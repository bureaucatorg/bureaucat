<script setup lang="ts">
import { BaseEdge, EdgeLabelRenderer, type EdgeProps, type GraphNode } from "@vue-flow/core";
import { Lock } from "lucide-vue-next";

const props = defineProps<EdgeProps<{ restricted: boolean; selected: boolean }>>();

const CURVE = 28;

function center(n: GraphNode) {
  return {
    x: n.computedPosition.x + n.dimensions.width / 2,
    y: n.computedPosition.y + n.dimensions.height / 2,
  };
}

// Point where the ray from the node's centre towards (tx, ty) leaves its box.
function borderPoint(n: GraphNode, tx: number, ty: number) {
  const c = center(n);
  const dx = tx - c.x;
  const dy = ty - c.y;
  const hw = n.dimensions.width / 2 + 4;
  const hh = n.dimensions.height / 2 + 4;
  const scale = Math.min(dx ? hw / Math.abs(dx) : Infinity, dy ? hh / Math.abs(dy) : Infinity);
  return { x: c.x + dx * scale, y: c.y + dy * scale };
}

// Edges bow to their right, so A→B and B→A sit side by side instead of overlapping.
const geometry = computed(() => {
  const s = center(props.sourceNode);
  const t = center(props.targetNode);
  const len = Math.hypot(t.x - s.x, t.y - s.y) || 1;
  const ctrl = {
    x: (s.x + t.x) / 2 - ((t.y - s.y) / len) * CURVE * 2,
    y: (s.y + t.y) / 2 + ((t.x - s.x) / len) * CURVE * 2,
  };
  const p1 = borderPoint(props.sourceNode, ctrl.x, ctrl.y);
  const p2 = borderPoint(props.targetNode, ctrl.x, ctrl.y);
  const mid = { x: 0.25 * p1.x + 0.5 * ctrl.x + 0.25 * p2.x, y: 0.25 * p1.y + 0.5 * ctrl.y + 0.25 * p2.y };
  const angle = (Math.atan2(p2.y - ctrl.y, p2.x - ctrl.x) * 180) / Math.PI;
  return {
    path: `M ${p1.x} ${p1.y} Q ${ctrl.x} ${ctrl.y} ${p2.x} ${p2.y}`,
    arrow: `translate(${p2.x} ${p2.y}) rotate(${angle})`,
    mid,
  };
});

const color = computed(() => (props.data?.selected ? "var(--foreground)" : "var(--muted-foreground)"));
</script>

<template>
  <BaseEdge
    :id="id"
    :path="geometry.path"
    :interaction-width="16"
    :style="{ stroke: color, strokeWidth: data?.selected ? 2 : 1.25 }"
  />
  <path d="M -9 -5 L 0 0 L -9 5 Z" :transform="geometry.arrow" :style="{ fill: color, stroke: 'none' }" />
  <EdgeLabelRenderer v-if="data?.restricted">
    <div
      class="nodrag nopan pointer-events-none absolute flex size-5 items-center justify-center rounded-full border bg-card text-muted-foreground"
      :style="{ transform: `translate(-50%, -50%) translate(${geometry.mid.x}px, ${geometry.mid.y}px)` }"
    >
      <Lock class="size-3" />
    </div>
  </EdgeLabelRenderer>
</template>

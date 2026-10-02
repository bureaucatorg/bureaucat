<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { nodes, assigned, linked, tasks, users, STATES, PRIORITIES } from "../lib/hero-web-data";

const at = Object.fromEntries(nodes.map((n) => [n.id, n]));
const line = ([a, b]: [string, string]) => ({ x1: at[a]!.x, y1: at[a]!.y, x2: at[b]!.x, y2: at[b]!.y });
const taskWidth = (id: string) => id.length * 7.2 + 22;

const root = ref<HTMLDivElement | null>(null);
const svg = ref<SVGSVGElement | null>(null);
const hovered = ref<{ id: string; x: number; y: number } | null>(null);
let hoveredEl: SVGGElement | null = null;
let raf = 0;
let io: IntersectionObserver | null = null;

const CARD_W = 288;
const CARD_H = 230;

function place(el: SVGGElement) {
  if (!root.value || !hovered.value) return;
  const box = root.value.getBoundingClientRect();
  const r = el.getBoundingClientRect();
  hovered.value.x = r.left + r.width / 2 - box.left;
  hovered.value.y = r.top + r.height / 2 - box.top;
}

function enter(id: string, e: MouseEvent) {
  hoveredEl = e.currentTarget as SVGGElement;
  hovered.value = { id, x: 0, y: 0 };
  place(hoveredEl);
}

function leave() {
  hovered.value = null;
  hoveredEl = null;
}

const isHot = (a: string, b?: string) => !!hovered.value && (hovered.value.id === a || hovered.value.id === b);

const cardStyle = computed(() => {
  const h = hovered.value;
  const box = root.value?.getBoundingClientRect();
  if (!h || !box) return {};
  const left = h.x + 28 + CARD_W > box.width - 8 ? h.x - 28 - CARD_W : h.x + 28;
  const top = Math.min(Math.max(h.y - 40, 60), box.height - CARD_H - 8);
  return { left: `${left}px`, top: `${top}px`, width: `${CARD_W}px` };
});

const neighbours = (id: string, list: [string, string][]) =>
  list.filter(([a, b]) => a === id || b === id).map(([a, b]) => (a === id ? b : a));

const taskCard = computed(() => {
  const id = hovered.value?.id;
  const t = id ? tasks[id] : undefined;
  if (!id || !t) return null;
  return {
    id,
    ...t,
    state: STATES[t.state],
    priority: PRIORITIES[t.priority],
    assignees: neighbours(id, assigned).map((u) => ({ id: u, ...users[u]! })),
    linked: neighbours(id, linked),
  };
});

const userCard = computed(() => {
  const id = hovered.value?.id;
  const u = id ? users[id] : undefined;
  if (!id || !u) return null;
  const list = neighbours(id, assigned).map((t) => ({ id: t, ...tasks[t]!, stateInfo: STATES[tasks[t]!.state] }));
  return {
    id,
    ...u,
    tasks: list,
    awaiting: list.filter((t) => t.state === "approval").length,
    done: list.filter((t) => t.state === "done").length,
  };
});

// Each node drifts on its own slow orbit; edges are re-pinned to the moving endpoints every frame.
onMounted(() => {
  const el = svg.value;
  if (!el || window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
  const groups = el.querySelectorAll<SVGGElement>("[data-node]");
  const lines = el.querySelectorAll<SVGLineElement>("[data-edge]");
  const edges = [...assigned, ...linked];
  const index = Object.fromEntries(nodes.map((n, i) => [n.id, i]));
  const seeds = nodes.map((_, i) => ({ phase: i * 1.7, speed: 0.3 + (i % 5) * 0.06, amp: 8 + (i % 3) * 4 }));
  const pos = nodes.map((n) => ({ x: n.x, y: n.y }));
  let visible = true;

  const tick = (now: number) => {
    const t = now / 1000;
    nodes.forEach((n, i) => {
      const s = seeds[i]!;
      const p = pos[i]!;
      p.x = n.x + Math.sin(t * s.speed + s.phase) * s.amp;
      p.y = n.y + Math.cos(t * s.speed * 0.8 + s.phase * 1.3) * s.amp;
      groups[i]!.setAttribute("transform", `translate(${p.x.toFixed(1)} ${p.y.toFixed(1)})`);
    });
    edges.forEach(([a, b], k) => {
      const A = pos[index[a]!]!;
      const B = pos[index[b]!]!;
      const l = lines[k]!;
      l.setAttribute("x1", A.x.toFixed(1));
      l.setAttribute("y1", A.y.toFixed(1));
      l.setAttribute("x2", B.x.toFixed(1));
      l.setAttribute("y2", B.y.toFixed(1));
    });
    if (hoveredEl) place(hoveredEl);
    raf = visible ? requestAnimationFrame(tick) : 0;
  };

  io = new IntersectionObserver(([entry]) => {
    visible = !!entry?.isIntersecting;
    if (visible && !raf) raf = requestAnimationFrame(tick);
  });
  io.observe(el);
});

onBeforeUnmount(() => {
  cancelAnimationFrame(raf);
  io?.disconnect();
});
</script>

<template>
  <div ref="root" aria-hidden="true" class="pointer-events-none absolute inset-0 overflow-hidden">
    <svg
      ref="svg"
      class="hero-web absolute inset-0 h-full w-full"
      viewBox="0 0 1440 900"
      preserveAspectRatio="xMidYMid slice"
    >
      <line
        v-for="(e, i) in assigned"
        :key="`a${i}`"
        data-edge
        v-bind="line(e)"
        class="flow dim stroke-zinc-950 dark:stroke-zinc-400"
        :class="{ hot: isHot(e[0], e[1]) }"
        stroke-width="1.2"
        :style="{ animationDelay: `${-i * 0.37}s` }"
      />
      <line
        v-for="(e, i) in linked"
        :key="`l${i}`"
        data-edge
        v-bind="line(e)"
        class="flow flow-slow dim stroke-amber-500"
        :class="{ hot: isHot(e[0], e[1]) }"
        stroke-width="1.8"
      />
      <g
        v-for="n in nodes"
        :key="n.id"
        data-node
        class="node pointer-events-auto"
        :transform="`translate(${n.x} ${n.y})`"
        @mouseenter="enter(n.id, $event)"
        @mouseleave="leave"
      >
        <template v-if="n.user">
          <circle r="16" class="fill-paper" />
          <g class="dim" :class="{ hot: hovered?.id === n.id }">
            <circle r="16" class="fill-none stroke-zinc-950/50 dark:stroke-zinc-400/60" stroke-width="1.2" />
            <text y="3.5" text-anchor="middle" font-size="10" font-weight="700" class="fill-muted-foreground font-mono">{{ n.id }}</text>
          </g>
        </template>
        <template v-else>
          <rect :x="-taskWidth(n.id) / 2" y="-13" :width="taskWidth(n.id)" height="26" rx="7" class="fill-paper" />
          <g class="dim" :class="{ hot: hovered?.id === n.id }">
            <rect
              :x="-taskWidth(n.id) / 2"
              y="-13"
              :width="taskWidth(n.id)"
              height="26"
              rx="7"
              class="fill-none stroke-zinc-950/40 dark:stroke-zinc-400/50"
              stroke-width="1.2"
            />
            <text y="4" text-anchor="middle" font-size="11" class="fill-muted-foreground font-mono">{{ n.id }}</text>
          </g>
        </template>
      </g>
    </svg>

    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0 translate-y-1"
      leave-active-class="transition duration-100 ease-in"
      leave-to-class="opacity-0"
    >
      <div
        v-if="taskCard || userCard"
        class="absolute z-10 rounded-xl border bg-card p-4 text-left shadow-[0_20px_50px_-20px_rgb(0_0_0/0.35)]"
        :style="cardStyle"
      >
        <template v-if="taskCard">
          <div class="flex items-center justify-between font-mono text-[10px] tracking-[0.15em] text-muted-foreground uppercase">
            <span>{{ taskCard.id }}</span>
            <span>Due {{ taskCard.due }}</span>
          </div>
          <p class="mt-2 text-[15px] leading-snug font-semibold">{{ taskCard.title }}</p>
          <div class="mt-3 flex flex-wrap items-center gap-2">
            <span
              class="inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 font-mono text-[11px]"
              :style="{ borderColor: `${taskCard.state.color}66`, background: `${taskCard.state.color}1a` }"
            >
              <span class="size-1.5 rounded-full" :style="{ background: taskCard.state.color }" />
              {{ taskCard.state.name }}
            </span>
            <span class="inline-flex items-center gap-1.5 font-mono text-[11px]" :style="{ color: taskCard.priority.color }">
              <span class="flex items-end gap-[2px]">
                <span
                  v-for="b in 4"
                  :key="b"
                  class="w-[3px] rounded-[1px]"
                  :style="{ height: `${b * 3 + 1}px`, background: b <= taskCard.priority.bars ? 'currentColor' : 'var(--border)' }"
                />
              </span>
              {{ taskCard.priority.name }}
            </span>
          </div>
          <div class="mt-3 border-t border-dashed pt-3">
            <p class="font-mono text-[10px] tracking-[0.15em] text-muted-foreground uppercase">Assignees</p>
            <div class="mt-1.5 flex flex-col gap-1">
              <div v-for="a in taskCard.assignees" :key="a.id" class="flex items-center gap-2 text-xs">
                <span class="grid size-5 place-items-center rounded-full font-mono text-[8px] font-bold" :class="a.tone">{{ a.id }}</span>
                {{ a.name }}
              </div>
            </div>
          </div>
          <div v-if="taskCard.linked.length" class="mt-3 flex flex-wrap items-center gap-1.5">
            <span class="font-mono text-[10px] tracking-[0.15em] text-muted-foreground uppercase">Linked</span>
            <span
              v-for="l in taskCard.linked"
              :key="l"
              class="rounded-md border border-amber-500/40 bg-amber-500/10 px-1.5 py-0.5 font-mono text-[10px] text-amber-700 dark:text-amber-400"
            >{{ l }}</span>
          </div>
          <div class="mt-3 flex flex-wrap gap-1.5">
            <span v-for="l in taskCard.labels" :key="l" class="rounded-md bg-muted px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">#{{ l }}</span>
          </div>
        </template>

        <template v-else-if="userCard">
          <div class="flex items-center gap-3">
            <span class="grid size-9 place-items-center rounded-full font-mono text-xs font-bold" :class="userCard.tone">{{ userCard.id }}</span>
            <div>
              <p class="text-[15px] leading-tight font-semibold">{{ userCard.name }}</p>
              <p class="text-xs text-muted-foreground">{{ userCard.role }}</p>
            </div>
          </div>
          <div class="mt-3 grid grid-cols-3 gap-2 border-y border-dashed py-2.5 text-center">
            <div>
              <p class="font-mono text-base font-bold">{{ userCard.tasks.length }}</p>
              <p class="font-mono text-[9px] tracking-[0.12em] text-muted-foreground uppercase">Assigned</p>
            </div>
            <div>
              <p class="font-mono text-base font-bold text-amber-600 dark:text-amber-500">{{ userCard.awaiting }}</p>
              <p class="font-mono text-[9px] tracking-[0.12em] text-muted-foreground uppercase">Approval</p>
            </div>
            <div>
              <p class="font-mono text-base font-bold">{{ userCard.done }}</p>
              <p class="font-mono text-[9px] tracking-[0.12em] text-muted-foreground uppercase">Done</p>
            </div>
          </div>
          <ul class="mt-2.5 flex flex-col gap-1.5">
            <li v-for="t in userCard.tasks.slice(0, 5)" :key="t.id" class="flex items-center gap-2 text-xs">
              <span class="size-1.5 shrink-0 rounded-full" :style="{ background: t.stateInfo.color }" />
              <span class="shrink-0 font-mono text-[10px] text-muted-foreground">{{ t.id }}</span>
              <span class="truncate">{{ t.title }}</span>
            </li>
            <li v-if="userCard.tasks.length > 5" class="font-mono text-[10px] text-muted-foreground">
              +{{ userCard.tasks.length - 5 }} more
            </li>
          </ul>
        </template>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.hero-web {
  mask-image: radial-gradient(ellipse 45% 40% at 33% 52%, transparent 0%, black 85%);
}
.dim {
  opacity: 0.4;
  transition: opacity 0.2s ease;
}
.hot {
  opacity: 1;
}
.node {
  cursor: default;
}
.flow {
  stroke-dasharray: 4 6;
  animation: flow 1.4s linear infinite;
}
@keyframes flow {
  to { stroke-dashoffset: -20; }
}
.flow-slow {
  stroke-dasharray: 8 6;
  animation: flow-slow 2.6s linear infinite;
}
@keyframes flow-slow {
  to { stroke-dashoffset: -28; }
}
@media (prefers-reduced-motion: reduce) {
  .flow,
  .flow-slow {
    animation: none;
  }
}
</style>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from "vue";
import { CHAPTERS, CHAPTER_VH, mountLandingStage } from "../lib/landing-stage";

const props = defineProps<{ noBs: boolean }>();

const chapters = [
  {
    form: "INTAKE",
    title: "Meet the cat.",
    body: "Requests, sign-offs and to-dos pile up. BureauCat files every one of them as a task, so nothing stays loose on someone's desk.",
    blunt: "It's a task manager. With a cat.",
  },
  {
    form: "STRUCTURE",
    title: "Everything has a place.",
    body: "Workspaces hold projects, and projects hold tasks. Each task gets a numbered key like DEVOPS-780, not \"that thing from Tuesday\".",
    blunt: "Workspaces → projects → tasks.",
  },
  {
    form: "LIFECYCLE",
    title: "Watch it move.",
    body: "Tasks travel through states you define. The defaults include Approval Pending, so a sign-off is a column, not an email thread. Priority, assignees and followers ride along.",
    blunt: "States, priorities, assignees, followers.",
  },
  {
    form: "PLANNING",
    title: "Plan the work.",
    body: "Break a task into subtasks, move cards across the kanban board, and slot work into cycles and modules to see the bigger picture.",
    blunt: "Subtasks. Boards. Cycles. Modules.",
  },
  {
    form: "GRAPH",
    title: "See the web.",
    body: "The graph view maps who is on what and how tasks nest, across every project you are a member of.",
    blunt: "A graph of people and tasks.",
  },
  {
    form: "NOTICES",
    title: "Nobody misses a beat.",
    body: "Comments with @mentions and attachments. Updates within 15 minutes fold into one notification: in-app, as an opt-in email digest, or a Mattermost DM.",
    blunt: "Notifications, batched. In-app, email, Mattermost.",
  },
  {
    form: "RECORD",
    title: "On the record.",
    body: "Every change lands in an append-only audit log, each entry hash-chained to the one before. One click verifies nothing was rewritten.",
    blunt: "Audit log. Tamper-evident.",
  },
];

const section = ref<HTMLElement | null>(null);
const panel = ref<HTMLElement | null>(null);
const svg = ref<SVGSVGElement | null>(null);
const active = ref(0);
let reduced = false;
let stage: { destroy(): void } | null = null;

onMounted(() => {
  if (!section.value || !svg.value || !panel.value) return;
  reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  const panelEl = panel.value;
  stage = mountLandingStage({
    svg: svg.value,
    section: section.value,
    topOffset: () => parseFloat(getComputedStyle(panelEl).top) || 0,
    reducedMotion: reduced,
    onChapter: (i) => (active.value = i),
  });
});

onBeforeUnmount(() => stage?.destroy());

function goTo(i: number) {
  if (!section.value || !panel.value) return;
  const top = section.value.getBoundingClientRect().top + window.scrollY;
  const offset = parseFloat(getComputedStyle(panel.value).top) || 0;
  window.scrollTo({
    top: top - offset + (i + 0.6) * CHAPTER_VH * window.innerHeight,
    behavior: reduced ? "auto" : "smooth",
  });
}

const pad = (n: number) => String(n).padStart(2, "0");
</script>

<template>
  <section
    ref="section"
    aria-label="What BureauCat does"
    class="relative"
    :style="{ height: `calc(${CHAPTERS * CHAPTER_VH * 100}vh + 100vh - 3rem)` }"
  >
    <div
      v-for="(_, i) in chapters"
      :id="`chapter-${i + 1}`"
      :key="i"
      class="pointer-events-none absolute left-0 h-px w-px scroll-mt-12"
      :style="{ top: `${(i + 0.6) * CHAPTER_VH * 100}vh` }"
    />

    <div
      ref="panel"
      class="sticky top-12 flex h-[calc(100vh-3rem)] flex-col justify-center gap-2 overflow-hidden md:grid md:grid-cols-[minmax(0,4.5fr)_minmax(0,7.5fr)] md:items-center md:gap-8 mx-auto max-w-6xl px-4 md:px-6"
    >
      <!-- Copy -->
      <div class="order-2 flex flex-col gap-6 md:order-1">
        <div class="relative min-h-[9.5rem] md:min-h-[17rem]">
          <article
            v-for="(c, i) in chapters"
            :key="c.form"
            :aria-hidden="i !== active"
            class="absolute inset-x-0 top-0 transition-all duration-500 ease-out motion-reduce:transition-none"
            :class="i === active ? 'translate-y-0 opacity-100' : i < active ? '-translate-y-4 opacity-0' : 'translate-y-4 opacity-0'"
          >
            <p class="font-mono text-[11px] tracking-[0.2em] text-amber-700 uppercase dark:text-amber-400">
              § {{ pad(i + 1) }} · Form BC-{{ pad(i + 1) }} · {{ c.form }}
            </p>
            <h2 class="mt-3 text-3xl font-bold tracking-tight md:mt-4 md:text-5xl md:leading-[1.05]">
              {{ c.title }}
            </h2>
            <p class="mt-3 max-w-md text-[15px] leading-relaxed text-muted-foreground md:mt-5 md:text-lg">
              {{ props.noBs ? c.blunt : c.body }}
            </p>
          </article>
        </div>

        <nav aria-label="Chapters" class="flex items-center gap-1.5">
          <button
            v-for="(c, i) in chapters"
            :key="c.form"
            type="button"
            :aria-label="`Go to chapter ${i + 1}: ${c.title}`"
            :aria-current="i === active ? 'step' : undefined"
            class="group flex h-6 items-center"
            @click="goTo(i)"
          >
            <span
              class="block h-[3px] rounded-full transition-all duration-500 motion-reduce:transition-none"
              :class="i === active ? 'w-8 bg-amber-500' : i < active ? 'w-4 bg-foreground/40 group-hover:bg-foreground/60' : 'w-4 bg-foreground/15 group-hover:bg-foreground/30'"
            />
          </button>
          <span class="ml-3 font-mono text-[11px] tabular-nums text-muted-foreground">
            {{ pad(active + 1) }} / {{ pad(CHAPTERS) }}
          </span>
        </nav>
      </div>

      <!-- Stage -->
      <div class="relative order-1 -mx-2 aspect-[5/4] max-h-[55%] md:mx-0 md:order-2 md:aspect-auto md:h-full md:max-h-none">
        <svg
          ref="svg"
          role="img"
          :aria-label="`${chapters[active]?.title} ${chapters[active]?.body}`"
          class="font-sans absolute inset-0 h-full w-full select-none"
          preserveAspectRatio="xMidYMid meet"
        />
      </div>
    </div>
  </section>
</template>

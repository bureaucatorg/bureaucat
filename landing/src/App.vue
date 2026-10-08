<script setup lang="ts">
import { ref } from "vue";
import { ArrowRight, ArrowDown, Sparkles } from "lucide-vue-next";
import SiteHeader from "./components/SiteHeader.vue";
import CatLogo from "./components/CatLogo.vue";
import HeroWeb from "./components/HeroWeb.vue";
import LandingStory from "./components/LandingStory.vue";
import LandingDrawer from "./components/LandingDrawer.vue";
import LandingFaq from "./components/LandingFaq.vue";
import { appLink } from "./lib/config";

const noBS = ref(false);

function scrollToStory() {
  document.getElementById("chapter-1")?.scrollIntoView({ behavior: "smooth" });
}

const primaryBtn =
  "inline-flex h-10 items-center gap-2 rounded-md bg-amber-500 px-6 text-sm font-medium text-amber-950 transition-colors hover:bg-amber-400";
const outlineBtn =
  "inline-flex h-10 items-center rounded-md border border-border bg-background px-6 text-sm font-medium shadow-xs transition-colors hover:bg-muted";
</script>

<template>
  <div class="relative flex min-h-screen flex-col overflow-x-clip">
    <a
      href="#main-content"
      class="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-[60] focus:rounded-md focus:bg-background focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:shadow-md"
    >Skip to content</a>
    <div
      aria-hidden="true"
      class="pointer-events-none fixed inset-0 z-50 opacity-[0.025] dark:opacity-[0.04]"
      style="background-image: url(&quot;data:image/svg+xml,%3Csvg viewBox=%220 0 256 256%22 xmlns=%22http://www.w3.org/2000/svg%22%3E%3Cfilter id=%22n%22%3E%3CfeTurbulence type=%22fractalNoise%22 baseFrequency=%220.9%22 numOctaves=%224%22 stitchTiles=%22stitch%22/%3E%3C/filter%3E%3Crect width=%22100%25%22 height=%22100%25%22 filter=%22url(%23n)%22/%3E%3C/svg%3E&quot;)"
    />

    <SiteHeader />

    <main id="main-content" class="flex-1">
      <!-- Hero -->
      <section aria-labelledby="hero-title" class="relative">
        <div aria-hidden="true" class="ruled pointer-events-none absolute inset-0 -z-0" />
        <HeroWeb />
        <div
          class="pointer-events-none relative mx-auto grid min-h-[calc(100svh-3rem)] max-w-6xl items-center gap-14 px-4 pt-14 pb-28 lg:grid-cols-[minmax(0,1fr)_22rem] md:px-6 md:pt-10"
        >
          <div class="pointer-events-auto">
            <h1 id="hero-title" class="rise text-5xl leading-[1.02] font-bold tracking-[-0.035em] sm:text-6xl lg:text-[4.25rem]" style="--d: 120ms">
              <span class="mb-5 block font-mono text-[11px] font-medium tracking-[0.2em] text-amber-700 uppercase dark:text-amber-400">
                Self-hosted · Open-source task manager
              </span>
              <template v-if="!noBS">
                Bureaucracy<br />
                That Actually
                <span class="relative inline-block text-amber-600 dark:text-amber-500">
                  Moves
                  <svg class="absolute -bottom-2 left-0 w-full" viewBox="0 0 300 14" fill="none" aria-hidden="true">
                    <path class="underline-draw" d="M3 10C60 3 170 2 297 9" pathLength="1" stroke="currentColor" stroke-opacity="0.45" stroke-width="5" stroke-linecap="round" />
                  </svg>
                </span>
              </template>
              <template v-else>
                A No-Nonsense<br />
                <span class="relative inline-block text-amber-600 dark:text-amber-500">
                  Task Manager
                  <svg class="absolute -bottom-2 left-0 w-full" viewBox="0 0 300 14" fill="none" aria-hidden="true">
                    <path class="underline-draw" d="M3 10C60 3 170 2 297 9" pathLength="1" stroke="currentColor" stroke-opacity="0.45" stroke-width="5" stroke-linecap="round" />
                  </svg>
                </span>
              </template>
            </h1>

            <p class="rise mt-8 max-w-xl text-lg leading-relaxed text-muted-foreground md:text-xl" style="--d: 240ms">
              <template v-if="!noBS">
                Requests, sign-offs and to-dos, turned into tasks that move, with approval workflows, kanban boards and a tamper-evident audit log.
              </template>
              <template v-else>Create tasks. Assign people. Track progress. That's it.<br/>Open Source. Self-hosted.</template>
            </p>

            <div class="rise mt-9 flex flex-wrap items-center gap-3" style="--d: 360ms">
              <a :href="appLink('/signup')" :class="primaryBtn">Get started <ArrowRight class="size-4" /></a>
              <a :href="appLink('/signin')" :class="outlineBtn">Sign in</a>
              <button
                type="button"
                class="inline-flex h-10 items-center gap-2 rounded-md px-4 text-sm font-medium text-amber-700 transition-colors hover:bg-amber-500/10 hover:text-amber-800 dark:text-amber-400 dark:hover:text-amber-300"
                :aria-pressed="noBS"
                @click="noBS = !noBS"
              >
                <Sparkles class="size-4" />
                {{ noBS ? "Add the BS back" : "Cut the BS" }}
              </button>
            </div>
          </div>

          <!-- Dossier -->
          <div aria-hidden="true" class="dossier relative mx-auto w-full max-w-[22rem] md:max-w-sm">
            <div class="sheet-back absolute inset-0 rounded-2xl border bg-card shadow-sm" />
            <div class="sheet-mid absolute inset-0 rounded-2xl border bg-card shadow-sm" />
            <div class="peek absolute -top-11 right-10">
              <CatLogo :size="64" />
            </div>
            <div class="sheet-front relative rounded-2xl border bg-card p-6 shadow-[0_30px_60px_-30px_rgb(0_0_0/0.35)]">
              <div class="flex items-center justify-between font-mono text-[10px] tracking-[0.18em] text-muted-foreground uppercase">
                <span>Form BC-00</span><span>DEVOPS-780</span>
              </div>
              <p class="mt-4 text-xl leading-snug font-semibold tracking-tight">Host Qwen3.8:27B on the GPU Cluster</p>
              <div class="relative my-5 border-t border-dashed">
                <span class="absolute -top-2.5 -left-[34px] size-5 rounded-full border bg-paper" />
                <span class="absolute -top-2.5 -right-[34px] size-5 rounded-full border bg-paper" />
              </div>
              <dl class="grid grid-cols-[6.5rem_1fr] items-center gap-y-3.5 text-sm">
                <dt class="font-mono text-[10px] tracking-[0.18em] text-muted-foreground uppercase">State</dt>
                <dd class="relative h-6">
                  <span class="state-a absolute inset-y-0 left-0 inline-flex items-center gap-1.5 rounded-full border border-amber-500/40 bg-amber-500/10 px-2.5 font-mono text-xs">
                    <span class="size-2 rounded-full bg-[#F59E0B]" /> Approval Pending
                  </span>
                  <span class="state-b absolute inset-y-0 left-0 inline-flex items-center gap-1.5 rounded-full border border-emerald-500/40 bg-emerald-500/10 px-2.5 font-mono text-xs">
                    <span class="size-2 rounded-full bg-[#10B981]" /> In Progress
                  </span>
                </dd>
                <dt class="font-mono text-[10px] tracking-[0.18em] text-muted-foreground uppercase">Priority</dt>
                <dd class="inline-flex items-center gap-2 font-mono text-xs text-[#F97316]">
                  <span class="flex items-end gap-[2px]">
                    <span class="h-1 w-[3px] rounded-[1px] bg-current" /><span class="h-2 w-[3px] rounded-[1px] bg-current" /><span class="h-3 w-[3px] rounded-[1px] bg-current" /><span class="h-4 w-[3px] rounded-[1px] bg-muted-foreground/25" />
                  </span>
                  High
                </dd>
                <dt class="font-mono text-[10px] tracking-[0.18em] text-muted-foreground uppercase">Assignees</dt>
                <dd class="flex -space-x-1.5">
                  <span class="grid size-6 place-items-center rounded-full bg-sky-100 font-mono text-[9px] font-bold text-sky-800 ring-2 ring-card">AS</span>
                  <span class="grid size-6 place-items-center rounded-full bg-violet-100 font-mono text-[9px] font-bold text-violet-800 ring-2 ring-card">CC</span>
                </dd>
                <dt class="font-mono text-[10px] tracking-[0.18em] text-muted-foreground uppercase">Filed by</dt>
                <dd class="font-mono text-xs">you, just now</dd>
              </dl>
              <div class="stamp pointer-events-none absolute mix-blend-multiply dark:mix-blend-screen right-5 top-[11.5rem] rounded-md border-[3px] border-amber-600 px-3 py-1 font-mono text-lg font-bold tracking-[0.25em] text-amber-600">
                <span class="block rounded-sm border border-amber-600/70 px-2">APPROVED</span>
              </div>
            </div>
          </div>
        </div>

        <button
          type="button"
          class="rise absolute bottom-6 left-1/2 flex -translate-x-1/2 flex-col items-center gap-2 font-mono text-[10px] tracking-[0.25em] text-muted-foreground uppercase transition-colors hover:text-foreground"
          style="--d: 900ms"
          @click="scrollToStory"
        >
          Scroll to file
          <ArrowDown class="cue size-4" />
        </button>
      </section>

      <LandingStory :no-bs="noBS" />

      <LandingDrawer :no-bs="noBS" />

      <LandingFaq />

      <!-- Closing -->
      <section aria-labelledby="closing-title" class="relative overflow-hidden border-t border-border/60">
        <div aria-hidden="true" class="pointer-events-none absolute -top-40 left-1/2 h-80 w-[40rem] max-w-full -translate-x-1/2 rounded-full bg-amber-500/10 blur-3xl" />
        <div class="relative mx-auto max-w-3xl px-4 py-24 text-center md:py-32">
          <CatLogo :size="44" class="mx-auto" />
          <h2 id="closing-title" class="mt-6 text-3xl font-bold tracking-tight md:text-5xl">
            <template v-if="!noBS">Ready to make bureaucracy<br class="hidden sm:block" /> work for you?</template>
            <template v-else>Ready to get things done?</template>
          </h2>
          <p class="mt-5 text-lg text-muted-foreground">Sign up, create a project, file your first task.</p>
          <div class="mt-9 flex flex-wrap justify-center gap-3">
            <a :href="appLink('/signup')" :class="primaryBtn">Get started <ArrowRight class="size-4" /></a>
            <a :href="appLink('/signin')" :class="outlineBtn">Sign in</a>
          </div>
        </div>
      </section>
    </main>

    <footer class="border-t border-border/60">
      <div class="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 px-4 py-8 md:flex-row md:px-6">
        <div class="flex items-center gap-2">
          <CatLogo :size="20" />
          <span class="text-sm text-muted-foreground">BureauCat &copy; 2026</span>
        </div>
        <nav aria-label="Footer" class="flex flex-wrap justify-center gap-x-5 gap-y-2 text-sm text-muted-foreground">
          <a href="/about" class="transition-colors hover:text-foreground">About</a>
          <a href="/developers" class="transition-colors hover:text-foreground">Developers</a>
          <a href="/contact" class="transition-colors hover:text-foreground">Contact</a>
          <a href="/privacy" class="transition-colors hover:text-foreground">Privacy</a>
        </nav>
        <p class="text-center text-sm text-muted-foreground">
          Crafted by
          <a
            href="https://codingcoffee.dev"
            target="_blank"
            rel="noopener noreferrer"
            class="underline underline-offset-4 transition-colors hover:text-foreground"
          >Ameya Shenoy</a>
        </p>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.ruled {
  background-image: repeating-linear-gradient(to bottom, transparent 0 31px, var(--border) 31px 32px);
  mask-image: radial-gradient(ellipse 70% 60% at 50% 40%, black 20%, transparent 75%);
  opacity: 0.55;
}

.rise {
  animation: rise 0.8s cubic-bezier(0.2, 0.8, 0.2, 1) var(--d, 0ms) both;
}
@keyframes rise {
  from { opacity: 0; transform: translateY(18px); }
}

.underline-draw {
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  animation: draw 0.9s cubic-bezier(0.65, 0, 0.35, 1) 0.7s forwards;
}
@keyframes draw {
  to { stroke-dashoffset: 0; }
}

.dossier {
  animation: rise 1s cubic-bezier(0.2, 0.8, 0.2, 1) 0.3s both;
}
.sheet-back { transform: rotate(-7deg) translate(-14px, 10px); animation: drift-a 7s ease-in-out 2s infinite; }
.sheet-mid { transform: rotate(4deg) translate(12px, 6px); animation: drift-b 8s ease-in-out 2s infinite; }
.sheet-front { animation: thud 0.45s ease-out 1.55s both; }
@keyframes drift-a { 50% { transform: rotate(-5.5deg) translate(-10px, 4px); } }
@keyframes drift-b { 50% { transform: rotate(5.5deg) translate(16px, 10px); } }
@keyframes thud {
  0%, 100% { transform: none; }
  25% { transform: translateY(3px) rotate(-0.6deg); }
  60% { transform: translateY(-1px) rotate(0.3deg); }
}

.stamp {
  transform: rotate(-14deg);
  opacity: 0.9;
  animation: stamp 0.4s cubic-bezier(0.5, 0, 0.75, 0) 1.2s both;
}
@keyframes stamp {
  from { opacity: 0; transform: rotate(-20deg) scale(2.6); }
  to { opacity: 0.9; transform: rotate(-14deg) scale(1); }
}

.state-a { animation: fade-out 0.3s ease 1.6s both; }
.state-b { animation: fade-in 0.4s ease 1.75s both; }
@keyframes fade-out { to { opacity: 0; transform: translateY(-6px); } }
@keyframes fade-in { from { opacity: 0; transform: translateY(6px); } }

.peek {
  z-index: 0;
  animation: peek 0.7s cubic-bezier(0.34, 1.56, 0.64, 1) 2.1s both;
}
.sheet-front { z-index: 1; }
@keyframes peek {
  from { transform: translateY(60%) rotate(8deg); opacity: 0; }
}

.cue { animation: cue 1.8s ease-in-out infinite; }
@keyframes cue { 50% { transform: translateY(5px); } }

@media (prefers-reduced-motion: reduce) {
  .rise, .dossier, .sheet-back, .sheet-mid, .sheet-front, .stamp, .state-a, .state-b, .peek, .cue, .underline-draw {
    animation: none;
  }
  .underline-draw { stroke-dashoffset: 0; }
  .state-a { opacity: 0; }
}
</style>

<script setup lang="ts">
import { Play, Pause, RotateCcw, Timer, Settings2 } from "lucide-vue-next";

const CANONICAL = "https://bureaucat.org/pomodoro";
const OG_IMAGE = "https://bureaucat.org/pomodoro-og.svg";
const SEO_TITLE = "Pomodoro Timer — Focus, one tick at a time · Bureaucat";
const SEO_DESC =
  "A minimal, keyboard-driven Pomodoro timer by Bureaucat. Custom focus and break durations, beautiful watch-dial interface, desktop notifications, zero tracking. Free and ready in your browser.";
const SEO_KEYWORDS = [
  "pomodoro timer",
  "pomodoro technique",
  "focus timer",
  "productivity timer",
  "deep work timer",
  "tomato timer",
  "25 minute timer",
  "work timer",
  "study timer",
  "bureaucat",
  "minimal pomodoro",
  "browser pomodoro",
].join(", ");

useSeoMeta({
  title: SEO_TITLE,
  description: SEO_DESC,
  keywords: SEO_KEYWORDS,
  author: "Bureaucat",
  robots: "index, follow, max-image-preview:large, max-snippet:-1",
  themeColor: "#F59E0B",
  applicationName: "Bureaucat Pomodoro",

  ogTitle: "Pomodoro Timer — Bureaucat",
  ogDescription: SEO_DESC,
  ogType: "website",
  ogUrl: CANONICAL,
  ogSiteName: "Bureaucat",
  ogLocale: "en_US",
  ogImage: OG_IMAGE,
  ogImageUrl: OG_IMAGE,
  ogImageSecureUrl: OG_IMAGE,
  ogImageType: "image/svg+xml",
  ogImageWidth: 1200,
  ogImageHeight: 630,
  ogImageAlt: "A watch-style Pomodoro timer dial showing 20:00, with progress arc and session indicators — Bureaucat Pomodoro.",

  twitterCard: "summary_large_image",
  twitterTitle: "Pomodoro Timer — Bureaucat",
  twitterDescription: SEO_DESC,
  twitterImage: OG_IMAGE,
  twitterImageAlt: "Bureaucat Pomodoro — a minimal, watch-dial focus timer.",

  appleMobileWebAppCapable: "yes",
  appleMobileWebAppTitle: "Pomodoro",
  appleMobileWebAppStatusBarStyle: "black-translucent",
});

useHead({
  htmlAttrs: { lang: "en" },
  link: [
    { rel: "canonical", href: CANONICAL },
    { rel: "alternate", hreflang: "en", href: CANONICAL },
    { rel: "alternate", hreflang: "x-default", href: CANONICAL },
    { rel: "preload", as: "image", href: "/pomodoro-og.svg", type: "image/svg+xml" },
  ],
  meta: [
    { name: "format-detection", content: "telephone=no" },
    { name: "color-scheme", content: "light dark" },
    { name: "referrer", content: "strict-origin-when-cross-origin" },
    { name: "category", content: "Productivity" },
    { property: "og:image:alt", content: "Bureaucat Pomodoro — minimal focus timer" },
  ],
  script: [
    {
      type: "application/ld+json",
      innerHTML: JSON.stringify({
        "@context": "https://schema.org",
        "@graph": [
          {
            "@type": "WebApplication",
            "@id": `${CANONICAL}#webapp`,
            name: "Bureaucat Pomodoro",
            alternateName: "Pomodoro Timer",
            url: CANONICAL,
            description: SEO_DESC,
            image: OG_IMAGE,
            applicationCategory: "ProductivityApplication",
            applicationSubCategory: "Time Management",
            operatingSystem: "Web",
            browserRequirements: "Requires JavaScript. Modern browser recommended.",
            isAccessibleForFree: true,
            offers: {
              "@type": "Offer",
              price: "0",
              priceCurrency: "USD",
              availability: "https://schema.org/InStock",
            },
            featureList: [
              "Adjustable focus, short break, and long break durations",
              "Keyboard shortcuts (Space to toggle, R to reset)",
              "Desktop notifications on session end",
              "Session counter with 4-cycle grouping",
              "Watch-dial interface with amber progress arc",
              "Works offline after first load",
              "Zero tracking, zero account required",
            ],
            publisher: {
              "@type": "Organization",
              "@id": "https://bureaucat.org/#org",
              name: "Bureaucat",
              url: "https://bureaucat.org",
              logo: "https://bureaucat.org/logo.svg",
            },
            inLanguage: "en-US",
          },
          {
            "@type": "WebPage",
            "@id": CANONICAL,
            url: CANONICAL,
            name: SEO_TITLE,
            description: SEO_DESC,
            isPartOf: { "@id": "https://bureaucat.org/#website" },
            primaryImageOfPage: OG_IMAGE,
            breadcrumb: { "@id": `${CANONICAL}#breadcrumb` },
            inLanguage: "en-US",
          },
          {
            "@type": "BreadcrumbList",
            "@id": `${CANONICAL}#breadcrumb`,
            itemListElement: [
              {
                "@type": "ListItem",
                position: 1,
                name: "Home",
                item: "https://bureaucat.org/",
              },
              {
                "@type": "ListItem",
                position: 2,
                name: "Pomodoro",
                item: CANONICAL,
              },
            ],
          },
          {
            "@type": "HowTo",
            "@id": `${CANONICAL}#howto`,
            name: "How to use the Pomodoro technique with Bureaucat",
            description:
              "Use 20-minute focus intervals followed by short breaks to sustain deep work throughout the day.",
            totalTime: "PT2H",
            step: [
              {
                "@type": "HowToStep",
                position: 1,
                name: "Pick a task",
                text: "Choose a single task to focus on for the next session.",
              },
              {
                "@type": "HowToStep",
                position: 2,
                name: "Start Focus",
                text: "Press Space or Start. The timer counts down from 20 minutes by default.",
              },
              {
                "@type": "HowToStep",
                position: 3,
                name: "Work without interruption",
                text: "Stay on the chosen task. Note distractions to handle after the session.",
              },
              {
                "@type": "HowToStep",
                position: 4,
                name: "Short Break",
                text: "When the bell rings, switch to Short Break (5 min). Step away from the screen.",
              },
              {
                "@type": "HowToStep",
                position: 5,
                name: "Long Break every 4",
                text: "After four focus sessions, take a Long Break (15 min) to fully recover.",
              },
            ],
          },
          {
            "@type": "FAQPage",
            "@id": `${CANONICAL}#faq`,
            mainEntity: [
              {
                "@type": "Question",
                name: "What is the Pomodoro Technique?",
                acceptedAnswer: {
                  "@type": "Answer",
                  text: "A time-management method that breaks work into focused intervals (traditionally 25 minutes) separated by short breaks, helping sustain concentration across a day.",
                },
              },
              {
                "@type": "Question",
                name: "Can I customize the durations?",
                acceptedAnswer: {
                  "@type": "Answer",
                  text: "Yes. Click the settings icon next to the mode selector to set any duration between the allowed limits. Preferences are saved locally in your browser.",
                },
              },
              {
                "@type": "Question",
                name: "Does Bureaucat Pomodoro track or send my data anywhere?",
                acceptedAnswer: {
                  "@type": "Answer",
                  text: "No. The timer runs entirely in your browser. Nothing is recorded or transmitted.",
                },
              },
              {
                "@type": "Question",
                name: "Do I need to sign up?",
                acceptedAnswer: {
                  "@type": "Answer",
                  text: "No. /pomodoro is accessible without an account.",
                },
              },
            ],
          },
        ],
      }),
    },
  ],
});

type Mode = "focus" | "short" | "long";

const DEFAULTS: Record<Mode, number> = {
  focus: 20,
  short: 5,
  long: 15,
};

const LIMITS: Record<Mode, { min: number; max: number }> = {
  focus: { min: 1, max: 180 },
  short: { min: 1, max: 60 },
  long: { min: 1, max: 60 },
};

const STORAGE_KEY = "pomodoro.durations.v1";

const minutesByMode = reactive<Record<Mode, number>>({ ...DEFAULTS });

const labels: Record<Mode, string> = {
  focus: "Focus",
  short: "Short Break",
  long: "Long Break",
};

const mode = ref<Mode>("focus");
const remaining = ref(DEFAULTS.focus * 60);
const running = ref(false);
const completed = ref(0);
const pulse = ref(false);
const settingsOpen = ref(false);

let interval: ReturnType<typeof setInterval> | null = null;
let endAt = 0;

const total = computed(() => minutesByMode[mode.value] * 60);
const progress = computed(() => (total.value === 0 ? 0 : 1 - remaining.value / total.value));

const minutes = computed(() => Math.floor(remaining.value / 60).toString().padStart(2, "0"));
const seconds = computed(() => Math.floor(remaining.value % 60).toString().padStart(2, "0"));

function clampMinutes(m: Mode, v: number): number {
  const lim = LIMITS[m];
  if (!Number.isFinite(v)) return DEFAULTS[m];
  return Math.min(lim.max, Math.max(lim.min, Math.round(v)));
}

function loadDurations() {
  if (typeof localStorage === "undefined") return;
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return;
    const parsed = JSON.parse(raw) as Partial<Record<Mode, number>>;
    (Object.keys(DEFAULTS) as Mode[]).forEach((k) => {
      if (typeof parsed[k] === "number") {
        minutesByMode[k] = clampMinutes(k, parsed[k] as number);
      }
    });
  } catch {}
}

function saveDurations() {
  if (typeof localStorage === "undefined") return;
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(minutesByMode));
  } catch {}
}

function tick() {
  const ms = endAt - Date.now();
  if (ms <= 0) {
    remaining.value = 0;
    finish();
    return;
  }
  remaining.value = Math.ceil(ms / 1000);
}

function start() {
  if (running.value || remaining.value <= 0) return;
  running.value = true;
  endAt = Date.now() + remaining.value * 1000;
  interval = setInterval(tick, 250);
}

function pause() {
  if (!running.value) return;
  running.value = false;
  if (interval) clearInterval(interval);
  interval = null;
}

function toggle() {
  running.value ? pause() : start();
}

function reset() {
  pause();
  remaining.value = total.value;
}

function setMode(m: Mode) {
  if (m === mode.value) return;
  pause();
  mode.value = m;
  remaining.value = minutesByMode[m] * 60;
}

function updateDuration(m: Mode, value: number) {
  const clamped = clampMinutes(m, value);
  minutesByMode[m] = clamped;
  saveDurations();
  if (m === mode.value && !running.value) {
    remaining.value = clamped * 60;
  }
}

function restoreDefaults() {
  (Object.keys(DEFAULTS) as Mode[]).forEach((k) => (minutesByMode[k] = DEFAULTS[k]));
  saveDurations();
  if (!running.value) remaining.value = minutesByMode[mode.value] * 60;
}

function finish() {
  pause();
  pulse.value = true;
  setTimeout(() => (pulse.value = false), 2000);
  if (mode.value === "focus") {
    completed.value += 1;
  }
  if (typeof window !== "undefined" && "Notification" in window && Notification.permission === "granted") {
    new Notification(`${labels[mode.value]} complete`, {
      body: mode.value === "focus" ? "Nice. Take a break." : "Back to focus.",
      silent: true,
    });
  }
  try {
    const ctx = new (window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext)();
    const o = ctx.createOscillator();
    const g = ctx.createGain();
    o.connect(g);
    g.connect(ctx.destination);
    o.frequency.value = 880;
    g.gain.setValueAtTime(0, ctx.currentTime);
    g.gain.linearRampToValueAtTime(0.08, ctx.currentTime + 0.02);
    g.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + 0.8);
    o.start();
    o.stop(ctx.currentTime + 0.85);
  } catch {}
}

function onKey(e: KeyboardEvent) {
  const target = e.target as HTMLElement | null;
  const tag = target?.tagName;
  if (tag === "INPUT" || tag === "TEXTAREA" || tag === "BUTTON") return;
  if (target?.getAttribute?.("role") === "button") return;
  if (e.code === "Space") {
    e.preventDefault();
    toggle();
  }
  if (e.key === "r" || e.key === "R") {
    reset();
  }
}

onMounted(() => {
  loadDurations();
  remaining.value = minutesByMode[mode.value] * 60;
  window.addEventListener("keydown", onKey);
  if ("Notification" in window && Notification.permission === "default") {
    Notification.requestPermission().catch(() => {});
  }
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", onKey);
  if (interval) clearInterval(interval);
});

watch([minutes, seconds, mode], () => {
  if (typeof document !== "undefined") {
    document.title = `${minutes.value}:${seconds.value} · ${labels[mode.value]}`;
  }
});

// Positions for 8 secondary bezel ticks — 2 on each side, spaced evenly from the cardinal tick.
function bezelTickStyle(i: number): Record<string, string> {
  // i: 0..7 → (side, offset)
  const side = Math.floor(i / 2); // 0=top, 1=right, 2=bottom, 3=left
  const variant = i % 2; // 0=before, 1=after cardinal
  const pos = variant === 0 ? "30%" : "70%";
  if (side === 0) return { top: "12px", left: pos, width: "1px", height: "6px" };
  if (side === 1) return { right: "12px", top: pos, width: "6px", height: "1px" };
  if (side === 2) return { bottom: "12px", left: pos, width: "1px", height: "6px" };
  return { left: "12px", top: pos, width: "6px", height: "1px" };
}
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />
    <main id="main-content" class="flex-1">
      <div class="mx-auto flex max-w-3xl flex-col items-center px-6 py-12">
        <!-- Header -->
        <div class="mb-10 flex flex-col items-center gap-2 text-center">
          <div class="inline-flex items-center gap-2 text-xs uppercase tracking-[0.18em] text-muted-foreground">
            <Timer class="size-3.5" />
            Pomodoro
          </div>
          <h1 class="text-2xl font-semibold tracking-tight">
            {{ labels[mode] }}
          </h1>
        </div>

        <!-- Mode tabs + settings -->
        <div class="mb-12 flex items-center gap-2">
          <div class="inline-flex items-center rounded-full border border-border/60 bg-muted/40 p-1 backdrop-blur">
            <button
              v-for="m in (['focus', 'short', 'long'] as Mode[])"
              :key="m"
              type="button"
              class="relative rounded-full px-4 py-1.5 text-sm font-medium outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring"
              :class="mode === m
                ? 'bg-amber-500/15 text-amber-700 dark:text-amber-400'
                : 'text-muted-foreground hover:text-foreground'"
              @click="setMode(m)"
            >
              {{ labels[m] }}
            </button>
          </div>

          <Popover v-model:open="settingsOpen">
            <PopoverTrigger as-child>
              <button
                type="button"
                title="Customize durations"
                aria-label="Customize durations"
                class="flex size-9 items-center justify-center rounded-full border border-border/60 bg-muted/40 text-muted-foreground outline-none transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
              >
                <Settings2 class="size-4" />
              </button>
            </PopoverTrigger>
            <PopoverContent align="end" class="w-72">
              <div class="space-y-4">
                <div>
                  <div class="text-sm font-semibold">Durations</div>
                  <div class="text-xs text-muted-foreground">In minutes. Saved on this device.</div>
                </div>
                <div class="space-y-3">
                  <div
                    v-for="m in (['focus', 'short', 'long'] as Mode[])"
                    :key="m"
                    class="flex items-center justify-between gap-3"
                  >
                    <label :for="`dur-${m}`" class="text-sm">{{ labels[m] }}</label>
                    <Input
                      :id="`dur-${m}`"
                      type="number"
                      :min="LIMITS[m].min"
                      :max="LIMITS[m].max"
                      :model-value="minutesByMode[m]"
                      class="h-8 w-20 text-right tabular-nums"
                      @update:model-value="(v) => updateDuration(m, Number(v))"
                    />
                  </div>
                </div>
                <div class="flex items-center justify-between border-t border-border/60 pt-3">
                  <button
                    type="button"
                    class="text-xs text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
                    @click="restoreDefaults"
                  >
                    Restore defaults
                  </button>
                  <Button size="sm" variant="ghost" @click="settingsOpen = false">Done</Button>
                </div>
              </div>
            </PopoverContent>
          </Popover>
        </div>

        <!-- Watch assembly -->
        <div class="watch-assembly relative select-none" :class="pulse && 'pomodoro-pulse'">
          <!-- Top lugs -->
          <span class="lug lug-top-left" />
          <span class="lug lug-top-right" />

          <!-- Bottom lugs -->
          <span class="lug lug-bottom-left" />
          <span class="lug lug-bottom-right" />

          <!-- Left engraving plate -->
          <div class="side-plate side-plate-left">
            <span class="side-engrave">CAT · {{ minutesByMode[mode].toString().padStart(2, "0") }}M</span>
          </div>

          <!-- Right crown + pushers -->
          <div class="side-controls">
            <button
              type="button"
              class="pusher pusher-top"
              :aria-label="running ? 'Pause timer' : 'Start timer'"
              :title="running ? 'Pause' : 'Start'"
              @click="toggle"
            />
            <div
              class="crown"
              role="button"
              tabindex="0"
              aria-label="Open settings"
              title="Settings"
              @click="settingsOpen = true"
              @keydown.enter="settingsOpen = true"
              @keydown.space.prevent="settingsOpen = true"
            />
            <button
              type="button"
              class="pusher pusher-bottom"
              aria-label="Reset timer"
              title="Reset"
              @click="reset"
            />
          </div>

          <!-- Bezel -->
          <div class="watch-bezel relative flex size-[340px] items-center justify-center rounded-[2.5rem]">
            <!-- Corner screws -->
            <span class="screw screw-tl" />
            <span class="screw screw-tr" />
            <span class="screw screw-bl" />
            <span class="screw screw-br" />

            <!-- Bezel cardinal ticks -->
            <span class="watch-tick watch-tick-top" />
            <span class="watch-tick watch-tick-right" />
            <span class="watch-tick watch-tick-bottom" />
            <span class="watch-tick watch-tick-left" />

            <!-- Bezel minute ticks (secondary) -->
            <span
              v-for="i in 8"
              :key="`bt-${i}`"
              class="bezel-minute-tick"
              :style="bezelTickStyle(i - 1)"
            />

            <!-- Dial -->
            <div class="watch-dial relative flex size-[296px] items-center justify-center rounded-[2rem]">
              <!-- Progress outline along inner rounded rect -->
              <svg
                class="pointer-events-none absolute inset-0"
                width="296"
                height="296"
                viewBox="0 0 296 296"
                aria-hidden="true"
              >
                <rect
                  x="6"
                  y="6"
                  width="284"
                  height="284"
                  rx="28"
                  ry="28"
                  fill="none"
                  stroke="var(--w-rim-soft)"
                  stroke-width="2"
                />
                <rect
                  x="6"
                  y="6"
                  width="284"
                  height="284"
                  rx="28"
                  ry="28"
                  fill="none"
                  stroke="#F59E0B"
                  stroke-width="2.5"
                  stroke-linecap="round"
                  pathLength="1"
                  class="transition-[stroke-dashoffset] duration-500 ease-out"
                  stroke-dasharray="1 1"
                  :stroke-dashoffset="1 - progress"
                />
              </svg>

              <!-- Crosshair guides -->
              <div class="crosshair crosshair-v" />
              <div class="crosshair crosshair-h" />

              <!-- Brand -->
              <div
                class="absolute top-4 left-1/2 -translate-x-1/2 text-[0.6rem] font-semibold uppercase tracking-[0.32em]"
                style="color: var(--w-ink-strong)"
              >
                Bureaucat
              </div>
              <div
                class="absolute top-[34px] left-1/2 -translate-x-1/2 text-[0.5rem] uppercase tracking-[0.35em]"
                style="color: var(--w-ink-soft)"
              >
                Chrono · Pomodoro
              </div>

              <!-- Mode window (top-right sub window) -->
              <div class="mode-window">
                <div class="text-[0.45rem] uppercase tracking-[0.25em]" style="color: var(--w-ink-soft)">Mode</div>
                <div class="font-mono text-[0.8rem] font-semibold leading-none tracking-wider" style="color: var(--w-ink-strong)">
                  {{ mode === "focus" ? "FOC" : mode === "short" ? "BRK" : "LNG" }}
                </div>
              </div>

              <!-- Status badge (top-left sub window) -->
              <div class="status-window">
                <span
                  class="size-1.5 rounded-full"
                  :class="running ? 'bg-amber-500 animate-pulse' : remaining === total ? 'bg-muted-foreground/40' : 'bg-amber-500/60'"
                />
                <span class="font-mono text-[0.55rem] uppercase tracking-[0.2em]">
                  {{ running ? "RUN" : remaining === total ? "RDY" : "HLD" }}
                </span>
              </div>

              <!-- Time -->
              <div class="flex flex-col items-center">
                <div
                  class="flex items-baseline font-mono text-[5rem] font-extralight tabular-nums leading-none tracking-tight"
                  style="color: var(--w-ink-strong)"
                >
                  <span>{{ minutes }}</span>
                  <span class="mx-0.5" :class="running && 'animate-pulse'" style="color: var(--w-ink-soft)">:</span>
                  <span>{{ seconds }}</span>
                </div>
              </div>

              <!-- Subdial: sessions -->
              <div class="subdial">
                <div class="subdial-label">Sessions</div>
                <div class="subdial-pips">
                  <span
                    v-for="i in 4"
                    :key="`sp-${i}`"
                    class="subdial-pip"
                    :class="i <= (completed % 4 || (completed > 0 && completed % 4 === 0 ? 4 : 0))
                      ? 'subdial-pip-on'
                      : ''"
                  />
                </div>
                <div class="subdial-count tabular-nums">{{ completed }}/4</div>
              </div>

              <!-- Progress percent subdial -->
              <div class="subdial subdial-right">
                <div class="subdial-label">Elapsed</div>
                <div class="subdial-meter">
                  <div class="subdial-meter-fill" :style="{ width: `${Math.round(progress * 100)}%` }" />
                </div>
                <div class="subdial-count tabular-nums">{{ Math.round(progress * 100) }}%</div>
              </div>

              <!-- Bottom engraving -->
              <div
                class="absolute bottom-[18px] left-1/2 -translate-x-1/2 font-mono text-[0.55rem] uppercase tracking-[0.28em] tabular-nums"
                style="color: var(--w-ink-soft)"
              >
                {{ minutesByMode[mode].toString().padStart(2, "0") }}:00 · {{ labels[mode] }}
              </div>
            </div>
          </div>
        </div>

        <!-- Controls -->
        <div class="mt-10 flex items-center gap-3">
          <Button
            size="lg"
            class="h-11 min-w-[140px] rounded-full bg-amber-500 text-white shadow-sm hover:bg-amber-500/90 dark:bg-amber-500 dark:text-amber-950"
            @click="toggle"
          >
            <Play v-if="!running" class="mr-1.5 size-4" />
            <Pause v-else class="mr-1.5 size-4" />
            {{ running ? "Pause" : remaining === total ? "Start" : "Resume" }}
          </Button>
          <Button
            variant="ghost"
            size="lg"
            class="h-11 rounded-full text-muted-foreground hover:text-foreground"
            :disabled="remaining === total && !running"
            @click="reset"
          >
            <RotateCcw class="mr-1.5 size-4" />
            Reset
          </Button>
        </div>

        <!-- Session counter -->
        <div class="mt-12 flex flex-col items-center gap-3">
          <div class="flex items-center gap-1.5">
            <span
              v-for="i in 4"
              :key="i"
              class="size-1.5 rounded-full transition-all"
              :class="i <= (completed % 4 || (completed > 0 && completed % 4 === 0 ? 4 : 0))
                ? 'bg-amber-500 scale-110'
                : 'bg-border'"
            />
          </div>
          <div class="text-xs text-muted-foreground">
            <span class="tabular-nums">{{ completed }}</span>
            focus session<span v-if="completed !== 1">s</span> today
          </div>
        </div>

        <!-- Keyboard hints -->
        <div class="mt-16 flex items-center gap-6 text-[0.7rem] uppercase tracking-[0.18em] text-muted-foreground/70">
          <div class="flex items-center gap-1.5">
            <kbd class="rounded border border-border/60 bg-muted/40 px-1.5 py-0.5 font-mono text-[0.65rem] normal-case tracking-normal">Space</kbd>
            Toggle
          </div>
          <div class="flex items-center gap-1.5">
            <kbd class="rounded border border-border/60 bg-muted/40 px-1.5 py-0.5 font-mono text-[0.65rem] normal-case tracking-normal">R</kbd>
            Reset
          </div>
          <div class="flex items-center gap-1.5">
            <kbd class="rounded border border-border/60 bg-muted/40 px-1.5 py-0.5 font-mono text-[0.65rem] normal-case tracking-normal">Crown</kbd>
            Settings
          </div>
        </div>

        <!-- About / SEO content -->
        <section class="mt-24 w-full max-w-2xl border-t border-border/50 pt-12">
          <header class="mb-6 flex items-center gap-2 text-[0.65rem] uppercase tracking-[0.25em] text-muted-foreground">
            <span class="inline-block h-px w-4 bg-amber-500" />
            Notes on the technique
          </header>

          <h2 class="text-xl font-semibold tracking-tight">
            A minimal Pomodoro timer, built for focused work.
          </h2>

          <div class="mt-4 space-y-4 text-sm leading-relaxed text-muted-foreground">
            <p>
              The Pomodoro technique splits work into short, focused intervals separated by deliberate breaks.
              Bureaucat's timer keeps the method simple: a 20-minute focus session by default,
              a five-minute break to breathe, and a longer 15-minute break every four cycles.
              Durations are yours to change from the crown on the right.
            </p>
            <p>
              Everything runs in your browser. No account, no tracking, no data leaves this tab.
              Your preferred durations are stored locally; clear your site data to reset them.
            </p>
          </div>

          <h3 class="mt-10 text-sm font-semibold tracking-tight">How to use it</h3>
          <ol class="mt-3 space-y-2 text-sm text-muted-foreground list-decimal list-inside">
            <li>Pick a single task to focus on.</li>
            <li>Press <kbd class="rounded border border-border/60 bg-muted/40 px-1 font-mono text-xs">Space</kbd> or the Start button to begin.</li>
            <li>Work uninterrupted until the bell — note any distractions for later.</li>
            <li>When the focus session ends, switch to a short break.</li>
            <li>After four focus sessions, take a longer break to reset.</li>
          </ol>

          <h3 class="mt-10 text-sm font-semibold tracking-tight">FAQ</h3>
          <dl class="mt-3 space-y-4 text-sm">
            <div>
              <dt class="font-medium text-foreground">Can I adjust the durations?</dt>
              <dd class="mt-1 text-muted-foreground">
                Yes — open the crown or the settings icon to set any focus or break length within reasonable limits.
                Changes save automatically to your browser.
              </dd>
            </div>
            <div>
              <dt class="font-medium text-foreground">Does it track anything?</dt>
              <dd class="mt-1 text-muted-foreground">
                No. The timer runs locally. There is no analytics pixel, no server round-trip, no account requirement.
              </dd>
            </div>
            <div>
              <dt class="font-medium text-foreground">Will it notify me when a session ends?</dt>
              <dd class="mt-1 text-muted-foreground">
                Yes — if you grant notification permission, you'll get a desktop alert along with a soft chime.
              </dd>
            </div>
            <div>
              <dt class="font-medium text-foreground">Do I need to sign up?</dt>
              <dd class="mt-1 text-muted-foreground">
                No. <NuxtLink to="/pomodoro" class="text-amber-600 underline underline-offset-2 hover:text-amber-700 dark:text-amber-400">/pomodoro</NuxtLink> is free and works without an account.
              </dd>
            </div>
          </dl>
        </section>
      </div>
    </main>
  </div>
</template>

<style scoped>
/* ---------- Watch palette (scoped, theme-aware) ---------- */
.watch-assembly {
  /* Light: warm light-gray bezel, near-white dial, defined rim */
  --w-bezel-top: hsl(32 10% 91%);
  --w-bezel-bottom: hsl(32 10% 78%);
  --w-dial-top: hsl(32 6% 99%);
  --w-dial-bottom: hsl(32 10% 88%);
  --w-rim: hsl(32 12% 42%);
  --w-rim-soft: hsl(32 10% 68%);
  --w-ink-strong: hsl(32 12% 28%);
  --w-ink-soft: hsl(32 6% 52%);
  --w-window-bg: hsl(0 0% 100%);
  --w-crown-ridge: hsl(32 10% 38%);
  --w-shadow: 0 30px 40px rgba(40, 30, 20, 0.18);

  width: 340px;
  height: 340px;
  display: flex;
  align-items: center;
  justify-content: center;
  filter: drop-shadow(var(--w-shadow));
}
:global(.dark) .watch-assembly {
  --w-bezel-top: hsl(30 10% 15%);
  --w-bezel-bottom: hsl(30 8% 7%);
  --w-dial-top: hsl(30 8% 10%);
  --w-dial-bottom: hsl(0 0% 4%);
  --w-rim: hsl(30 10% 30%);
  --w-rim-soft: hsl(30 8% 18%);
  --w-ink-strong: hsl(30 14% 72%);
  --w-ink-soft: hsl(30 6% 45%);
  --w-window-bg: hsl(30 6% 4%);
  --w-crown-ridge: hsl(30 6% 38%);
  --w-shadow: 0 30px 50px rgba(0, 0, 0, 0.6);
}

/* ---------- Bezel ---------- */
.watch-bezel {
  background: linear-gradient(155deg, var(--w-bezel-top), var(--w-bezel-bottom));
  border: 1.5px solid var(--w-rim);
  box-shadow:
    inset 0 1px 0 hsl(0 0% 100% / 0.5),
    inset 0 -1px 0 hsl(0 0% 0% / 0.15),
    0 0 0 1px hsl(0 0% 0% / 0.04);
}
:global(.dark) .watch-bezel {
  box-shadow:
    inset 0 1px 0 hsl(0 0% 100% / 0.05),
    inset 0 -1px 0 hsl(0 0% 0% / 0.6),
    0 0 0 1px hsl(0 0% 0% / 0.4);
}

/* ---------- Dial ---------- */
.watch-dial {
  background: radial-gradient(circle at 30% 18%, var(--w-dial-top) 0%, var(--w-dial-bottom) 90%);
  box-shadow:
    inset 0 2px 5px hsl(0 0% 0% / 0.12),
    inset 0 -1px 2px hsl(0 0% 100% / 0.4),
    inset 0 0 0 1.5px var(--w-rim);
}
:global(.dark) .watch-dial {
  box-shadow:
    inset 0 2px 5px hsl(0 0% 0% / 0.7),
    inset 0 -1px 2px hsl(0 0% 100% / 0.03),
    inset 0 0 0 1.5px var(--w-rim);
}

/* ---------- Cardinal ticks ---------- */
.watch-tick {
  position: absolute;
  background: var(--w-ink-strong);
  border-radius: 9999px;
  pointer-events: none;
}
.watch-tick-top,
.watch-tick-bottom {
  width: 2px;
  height: 10px;
  left: 50%;
  transform: translateX(-50%);
}
.watch-tick-left,
.watch-tick-right {
  width: 10px;
  height: 2px;
  top: 50%;
  transform: translateY(-50%);
}
.watch-tick-top { top: 10px; }
.watch-tick-bottom { bottom: 10px; }
.watch-tick-left { left: 10px; }
.watch-tick-right { right: 10px; }

.bezel-minute-tick {
  position: absolute;
  background: var(--w-ink-soft);
  pointer-events: none;
  border-radius: 1px;
}

/* ---------- Corner screws ---------- */
.screw {
  position: absolute;
  width: 10px;
  height: 10px;
  border-radius: 9999px;
  background: radial-gradient(circle at 35% 35%, var(--w-bezel-top), var(--w-bezel-bottom));
  box-shadow:
    inset 0 1px 1px hsl(0 0% 0% / 0.25),
    inset 0 0 0 1px var(--w-rim);
  pointer-events: none;
}
.screw::after {
  content: "";
  position: absolute;
  inset: 2px;
  border-top: 1.2px solid var(--w-ink-strong);
  transform: rotate(40deg);
  transform-origin: center;
}
:global(.dark) .screw {
  background: radial-gradient(circle at 35% 35%, hsl(30 6% 14%), hsl(0 0% 4%));
  box-shadow:
    inset 0 1px 1px hsl(0 0% 0% / 0.8),
    inset 0 0 0 1px hsl(0 0% 0% / 0.6);
}
.screw-tl { top: 14px; left: 14px; }
.screw-tr { top: 14px; right: 14px; }
.screw-bl { bottom: 14px; left: 14px; }
.screw-br { bottom: 14px; right: 14px; }

/* ---------- Lugs ---------- */
.lug {
  position: absolute;
  width: 46px;
  height: 22px;
  border-radius: 8px;
  background: linear-gradient(180deg, var(--w-bezel-top), var(--w-bezel-bottom));
  border: 1px solid var(--w-rim);
  box-shadow:
    inset 0 1px 0 hsl(0 0% 100% / 0.3),
    0 4px 6px -4px rgba(0, 0, 0, 0.22);
  z-index: 0;
}
:global(.dark) .lug {
  box-shadow:
    inset 0 1px 0 hsl(0 0% 100% / 0.04),
    0 4px 8px -4px rgba(0, 0, 0, 0.7);
}
.lug::before {
  content: "";
  position: absolute;
  inset: 6px 10px;
  border-radius: 2px;
  background: var(--w-rim-soft);
}
.lug-top-left { top: -14px; left: 28px; }
.lug-top-right { top: -14px; right: 28px; }
.lug-bottom-left { bottom: -14px; left: 28px; }
.lug-bottom-right { bottom: -14px; right: 28px; }

/* ---------- Side engraving plate (left) ---------- */
.side-plate-left {
  position: absolute;
  left: -12px;
  top: 50%;
  transform: translateY(-50%);
  width: 14px;
  height: 96px;
  border-radius: 3px 0 0 3px;
  background: linear-gradient(90deg, var(--w-bezel-bottom), var(--w-bezel-top));
  border: 1px solid var(--w-rim);
  border-right: none;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: inset 0 1px 0 hsl(0 0% 100% / 0.25);
}
:global(.dark) .side-plate-left {
  box-shadow: inset 0 1px 0 hsl(0 0% 100% / 0.04);
}
.side-engrave {
  writing-mode: vertical-rl;
  transform: rotate(180deg);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 8px;
  letter-spacing: 0.28em;
  color: var(--w-ink-soft);
  text-transform: uppercase;
  white-space: nowrap;
}

/* ---------- Right-side crown & pushers ---------- */
.side-controls {
  position: absolute;
  right: -12px;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  flex-direction: column;
  gap: 14px;
  align-items: center;
  z-index: 0;
}
.pusher {
  width: 12px;
  height: 24px;
  border-radius: 3px;
  border: 1px solid var(--w-rim);
  border-left: none;
  background: linear-gradient(90deg, var(--w-bezel-bottom), var(--w-bezel-top) 80%);
  box-shadow:
    inset 0 1px 0 hsl(0 0% 100% / 0.3),
    0 2px 3px rgba(0, 0, 0, 0.18);
  cursor: pointer;
  transition: transform 0.12s ease, filter 0.12s ease;
  outline: none;
}
.pusher:hover { filter: brightness(1.05); }
.pusher:active { transform: translateX(-2px); filter: brightness(0.95); }
.pusher:focus-visible { box-shadow: 0 0 0 2px hsl(var(--ring)); }

.crown {
  width: 18px;
  height: 44px;
  border-radius: 4px;
  border: 1px solid var(--w-rim);
  border-left: none;
  background:
    repeating-linear-gradient(
      0deg,
      var(--w-crown-ridge) 0,
      var(--w-crown-ridge) 1.2px,
      var(--w-bezel-top) 1.2px,
      var(--w-bezel-top) 3.5px
    );
  box-shadow:
    inset 0 1px 0 hsl(0 0% 100% / 0.25),
    0 2px 4px rgba(0, 0, 0, 0.22);
  cursor: pointer;
  transition: transform 0.12s ease;
  outline: none;
}
.crown:hover { transform: translateX(-1px); }
.crown:active { transform: translateX(-3px); }
.crown:focus-visible { box-shadow: 0 0 0 2px hsl(var(--ring)); }

/* ---------- Dial crosshair ---------- */
.crosshair {
  position: absolute;
  background: var(--w-rim-soft);
  pointer-events: none;
  opacity: 0.5;
}
.crosshair-v {
  left: 50%;
  top: 24%;
  bottom: 24%;
  width: 1px;
  transform: translateX(-0.5px);
}
.crosshair-h {
  top: 50%;
  left: 24%;
  right: 24%;
  height: 1px;
  transform: translateY(-0.5px);
}

/* ---------- Mode window (top-right sub window) ---------- */
.mode-window {
  position: absolute;
  top: 60px;
  right: 40px;
  min-width: 46px;
  padding: 4px 8px;
  border-radius: 5px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  background: var(--w-window-bg);
  box-shadow:
    inset 0 1px 2px hsl(0 0% 0% / 0.18),
    inset 0 0 0 1px var(--w-rim);
}

/* ---------- Status window (top-left) ---------- */
.status-window {
  position: absolute;
  top: 60px;
  left: 40px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 7px;
  border-radius: 999px;
  background: var(--w-window-bg);
  color: var(--w-ink-strong);
  box-shadow:
    inset 0 1px 2px hsl(0 0% 0% / 0.15),
    inset 0 0 0 1px var(--w-rim);
}

/* ---------- Subdials ---------- */
.subdial {
  position: absolute;
  bottom: 48px;
  left: 34px;
  width: 72px;
  padding: 6px 8px;
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  background: var(--w-window-bg);
  box-shadow:
    inset 0 1px 2px hsl(0 0% 0% / 0.18),
    inset 0 0 0 1px var(--w-rim);
}
.subdial-right {
  left: auto;
  right: 34px;
}
.subdial-label {
  font-size: 7px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--w-ink-soft);
}
.subdial-pips {
  display: flex;
  gap: 4px;
}
.subdial-pip {
  width: 6px;
  height: 6px;
  border-radius: 9999px;
  background: var(--w-rim-soft);
  transition: background 0.2s;
}
.subdial-pip-on {
  background: rgb(245, 158, 11);
  box-shadow: 0 0 6px rgba(245, 158, 11, 0.45);
}
.subdial-count {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 9px;
  color: var(--w-ink-strong);
}
.subdial-meter {
  width: 100%;
  height: 4px;
  border-radius: 2px;
  background: var(--w-rim-soft);
  overflow: hidden;
}
.subdial-meter-fill {
  height: 100%;
  background: rgb(245, 158, 11);
  transition: width 0.5s ease-out;
}

/* ---------- Pulse animation on finish ---------- */
.pomodoro-pulse {
  animation: pomodoro-pulse 1.2s ease-out;
}
@keyframes pomodoro-pulse {
  0% { filter: drop-shadow(0 30px 40px rgba(0, 0, 0, 0.18)) drop-shadow(0 0 0 rgba(245, 158, 11, 0)); }
  40% { filter: drop-shadow(0 30px 40px rgba(0, 0, 0, 0.18)) drop-shadow(0 0 28px rgba(245, 158, 11, 0.5)); }
  100% { filter: drop-shadow(0 30px 40px rgba(0, 0, 0, 0.18)) drop-shadow(0 0 0 rgba(245, 158, 11, 0)); }
}
</style>

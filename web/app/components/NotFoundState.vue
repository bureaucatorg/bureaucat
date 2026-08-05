<script setup lang="ts">
import { FileQuestion } from "lucide-vue-next";

withDefaults(
  defineProps<{
    /** Big numeral in the docket header. */
    code?: string | number;
    /** Short headline, e.g. "Task not found". */
    title: string;
    /** Sentence explaining what happened. */
    message?: string;
    /** What was looked up — rendered on the REF. line of the form. */
    reference?: string;
    /** Words inside the rubber stamp. */
    stamp?: string;
  }>(),
  { code: 404, message: "", reference: "", stamp: "NOT ON FILE" }
);

const filedOn = new Date().toLocaleDateString("en-GB", {
  day: "2-digit",
  month: "short",
  year: "numeric",
});
</script>

<template>
  <div class="nf-paper relative flex min-h-[70vh] items-center justify-center overflow-hidden px-6 py-16">
    <div class="nf-card relative w-full max-w-xl">
      <!-- Docket -->
      <div class="relative rounded-lg border border-border bg-card/80 shadow-sm backdrop-blur-[1px]">
        <!-- Header strip -->
        <div class="flex items-center justify-between gap-3 border-b border-dashed border-border px-5 py-2.5">
          <span class="font-mono text-[0.7rem] uppercase tracking-[0.18em] text-muted-foreground">
            Records Office
          </span>
          <span class="font-mono text-[0.7rem] uppercase tracking-[0.18em] text-muted-foreground">
            Form {{ code }}
          </span>
        </div>

        <div class="relative px-5 pb-6 pt-5 sm:px-8">
          <!-- Rubber stamp -->
          <div class="nf-stamp pointer-events-none absolute right-3 top-2 select-none sm:right-6">
            <div
              class="rotate-[-11deg] rounded-md border-[3px] border-double border-amber-600/60 px-3 py-1.5 text-amber-700/70 dark:border-amber-500/50 dark:text-amber-500/70"
            >
              <p class="font-mono text-sm font-bold uppercase leading-none tracking-[0.14em]">
                {{ stamp }}
              </p>
              <p class="mt-1 font-mono text-[0.6rem] uppercase leading-none tracking-[0.2em] opacity-80">
                {{ filedOn }}
              </p>
            </div>
          </div>

          <!-- Numeral -->
          <div class="nf-fade flex items-end gap-3" style="--d: 60ms">
            <span class="font-mono text-[5.5rem] font-bold leading-[0.85] tracking-tighter text-foreground/90 sm:text-[7rem]">
              {{ code }}
            </span>
            <FileQuestion class="mb-3 size-6 shrink-0 text-amber-500" />
          </div>

          <h1 class="nf-fade mt-5 text-xl font-semibold tracking-tight sm:text-2xl" style="--d: 140ms">
            {{ title }}
          </h1>
          <p v-if="message" class="nf-fade mt-1.5 max-w-md text-sm text-muted-foreground" style="--d: 200ms">
            {{ message }}
          </p>

          <!-- Form lines -->
          <dl class="nf-fade mt-6 space-y-0 border-t border-dashed border-border pt-4 font-mono text-xs" style="--d: 260ms">
            <div v-if="reference" class="flex gap-3 py-1.5">
              <dt class="w-24 shrink-0 uppercase tracking-[0.14em] text-muted-foreground">Ref.</dt>
              <dd class="min-w-0 truncate text-foreground/80">{{ reference }}</dd>
            </div>
          </dl>

          <div class="nf-fade mt-6 flex flex-wrap items-center gap-2" style="--d: 320ms">
            <slot name="actions" />
          </div>
        </div>
      </div>

      <!-- Paper edge under the docket -->
      <div class="mx-3 h-2 rounded-b-lg border border-t-0 border-border/70 bg-muted/40" />
      <div class="mx-6 h-2 rounded-b-lg border border-t-0 border-border/50 bg-muted/25" />
    </div>
  </div>
</template>

<style scoped>
/* Faint ruled-paper grid that fades out towards the edges. */
.nf-paper::before {
  content: "";
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(to right, var(--border) 1px, transparent 1px),
    linear-gradient(to bottom, var(--border) 1px, transparent 1px);
  background-size: 28px 28px;
  mask-image: radial-gradient(ellipse 70% 60% at 50% 45%, black, transparent 100%);
  opacity: 0.7;
  pointer-events: none;
}

.nf-fade {
  animation: nf-rise 0.5s cubic-bezier(0.22, 1, 0.36, 1) backwards;
  animation-delay: var(--d, 0ms);
}

@keyframes nf-rise {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
}

/* The stamp lands after the form is on the desk. */
.nf-stamp {
  animation: nf-stamp 0.34s cubic-bezier(0.2, 1.4, 0.4, 1) 420ms backwards;
}

@keyframes nf-stamp {
  from {
    opacity: 0;
    transform: scale(1.7) rotate(6deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .nf-fade,
  .nf-stamp {
    animation: none;
  }
}
</style>

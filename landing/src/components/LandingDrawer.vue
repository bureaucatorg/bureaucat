<script setup lang="ts">
import { appLink } from "../lib/config";
import { KeyRound, Terminal, BookOpen, Filter, FileText, Search, Palette, Network, Timer, Users } from "lucide-vue-next";

defineProps<{ noBs: boolean }>();

const items = [
  { icon: KeyRound, title: "Single sign-on", body: "Google and Zitadel, alongside passwords." },
  { icon: Terminal, title: "Tokens & CLI", body: "Read-only or read-write API tokens, plus a CLI." },
  { icon: BookOpen, title: "API docs", body: "Every endpoint, documented.", to: appLink("/docs") },
  { icon: Filter, title: "Saved views", body: "Filter, group and sort. Keep it private or share it." },
  { icon: FileText, title: "Pages", body: "Rich-text docs that live next to the project." },
  { icon: Search, title: "Search & quick create", body: "Ctrl K to find anything. Shift C to file a task." },
  { icon: Palette, title: "Your branding", body: "Rename the app for your organisation." },
  { icon: Network, title: "Graph view", body: "See who is on what and how tasks connect." },
  { icon: Timer, title: "Pomodoro", body: "A focus timer. The forms won't fill themselves.", to: appLink("/pomodoro") },
  { icon: Users, title: "Admin tools", body: "Merge users, restore projects, instance stats." },
];
</script>

<template>
  <section aria-labelledby="drawer-title" class="mx-auto max-w-6xl px-4 py-24 md:px-6 md:py-32">
    <p class="font-mono text-[11px] tracking-[0.2em] text-amber-700 uppercase dark:text-amber-400">
      § 08 · Also in the drawer
    </p>
    <h2 id="drawer-title" class="mt-4 text-3xl font-bold tracking-tight md:text-5xl">
      {{ noBs ? "Also included." : "The rest of the paperwork." }}
    </h2>

    <ul class="mt-12 grid grid-cols-1 border-t border-l border-border/70 sm:grid-cols-2 lg:grid-cols-5">
      <li
        v-for="(item, i) in items"
        :key="item.title"
        class="group relative border-r border-b border-border/70 bg-card/40 p-6 transition-colors hover:bg-card"
      >
        <span aria-hidden="true" class="font-mono text-[10px] tracking-widest text-muted-foreground/70">
          {{ String(i + 1).padStart(2, "0") }}
        </span>
        <component
          :is="item.icon"
          class="mt-4 size-5 text-amber-600 transition-transform duration-300 group-hover:-translate-y-0.5 dark:text-amber-500"
        />
        <h3 class="mt-4 text-[15px] font-semibold">
          <a v-if="item.to" :href="item.to" class="after:absolute after:inset-0 hover:underline underline-offset-4">
            {{ item.title }}
          </a>
          <template v-else>{{ item.title }}</template>
        </h3>
        <p class="mt-1.5 text-sm leading-relaxed text-muted-foreground">{{ item.body }}</p>
      </li>
    </ul>
  </section>
</template>

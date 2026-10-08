<script setup lang="ts">
import { ChevronDown } from "lucide-vue-next";

const faqs = [
  {
    q: "What is BureauCat?",
    a: "BureauCat is an open-source, self-hosted task manager built around approval workflows. Requests and sign-offs become numbered tasks that move through states such as Approval Pending, and every change is recorded in a tamper-evident audit log.",
  },
  {
    q: "Is BureauCat free?",
    a: "Yes. BureauCat is free and open source under the GNU Affero General Public License v3.0 (AGPL-3.0). You host it yourself.",
  },
  {
    q: "How do I self-host BureauCat?",
    a: "Run the single bureaucat binary or the codingcoffee/bureaucat Docker image with a PostgreSQL database. Set DATABASE_URL and JWT_SECRET, then start it with bureaucat serve --migrate. The web app and API are served from one port.",
  },
  {
    q: "How do approvals work?",
    a: "Approval is a task state. The default workflow includes Approval Pending, so a request waiting for sign-off sits in its own column on the board, with assignees, followers and a full history, instead of in an email thread.",
  },
  {
    q: "What makes the audit log tamper-evident?",
    a: "The audit log is append-only and each entry is hash-chained (SHA-256) to the one before it. Rewriting any past entry breaks the chain, and one click verifies the whole log.",
  },
  {
    q: "Does BureauCat have an API?",
    a: "Yes. Every endpoint is documented with an OpenAPI spec and interactive docs. API tokens can be read-only or read-write, and the bureaucat binary doubles as a CLI.",
  },
];
</script>

<template>
  <section aria-labelledby="faq-title" class="mx-auto max-w-3xl px-4 pb-24 md:px-6 md:pb-32">
    <p class="font-mono text-[11px] tracking-[0.2em] text-amber-700 uppercase dark:text-amber-400">
      § 09 · Questions on file
    </p>
    <h2 id="faq-title" class="mt-4 text-3xl font-bold tracking-tight md:text-5xl">Frequently asked questions</h2>

    <div class="mt-10 border-t border-border/70">
      <details v-for="f in faqs" :key="f.q" class="group border-b border-border/70">
        <summary class="flex cursor-pointer list-none items-center justify-between gap-4 py-5 text-left font-semibold [&::-webkit-details-marker]:hidden">
          <h3 class="text-[15px] md:text-base">{{ f.q }}</h3>
          <ChevronDown class="size-4 shrink-0 text-muted-foreground transition-transform group-open:rotate-180" />
        </summary>
        <p class="pb-5 text-sm leading-relaxed text-muted-foreground md:text-[15px]">{{ f.a }}</p>
      </details>
    </div>
  </section>
</template>

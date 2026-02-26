<script setup lang="ts">
import { Clock, Shield, Eye, Zap, Sparkles } from "lucide-vue-next";

const { appName } = useSettings();

const quirkyDesc = "Bureaucracy That Actually Moves.";

useSeoMeta({
  title: () => appName.value,
  description: quirkyDesc,
  ogTitle: () => appName.value,
  ogDescription: quirkyDesc,
  ogImage: "/api/v1/og-image",
  ogType: "website",
  twitterCard: "summary_large_image",
  twitterTitle: () => appName.value,
  twitterDescription: quirkyDesc,
  twitterImage: "/api/v1/og-image",
});

const noBS = ref(false);

const features = computed(() =>
  noBS.value
    ? [
        {
          icon: Clock,
          stat: "60%",
          title: "Less waiting",
          description: "Tasks go to the right person. No email chains.",
        },
        {
          icon: Shield,
          stat: "100%",
          title: "Everything logged",
          description: "Full audit trail. Zero effort.",
        },
        {
          icon: Eye,
          stat: "0",
          title: "Nothing lost",
          description: "Every request tracked. Nothing slips.",
        },
      ]
    : [
        {
          icon: Clock,
          stat: "60%",
          title: "Less time chasing",
          description:
            "Automated routing gets requests to the right people instantly. No more email chains.",
        },
        {
          icon: Shield,
          stat: "100%",
          title: "Audit coverage",
          description:
            "Every action logged, every decision tracked. Complete compliance without overhead.",
        },
        {
          icon: Eye,
          stat: "0",
          title: "Requests lost",
          description:
            "Real-time visibility into every request. Nothing falls through the cracks.",
        },
      ],
);
</script>

<template>
  <div class="relative flex min-h-screen flex-col overflow-hidden">
    <!-- Grain overlay -->
    <div
      class="pointer-events-none fixed inset-0 z-50 opacity-[0.015] dark:opacity-[0.03]"
      style="
        background-image: url(&quot;data:image/svg+xml,%3Csvg viewBox=%220 0 256 256%22 xmlns=%22http://www.w3.org/2000/svg%22%3E%3Cfilter id=%22noise%22%3E%3CfeTurbulence type=%22fractalNoise%22 baseFrequency=%220.9%22 numOctaves=%224%22 stitchTiles=%22stitch%22/%3E%3C/filter%3E%3Crect width=%22100%25%22 height=%22100%25%22 filter=%22url(%23noise)%22/%3E%3C/svg%3E&quot;);
      "
    />

    <!-- Header -->
    <Navbar />

    <main class="flex-1">
      <!-- Hero Section -->
      <section class="relative">
        <!-- Background decoration -->
        <div class="absolute inset-0 -z-10 overflow-hidden">
          <div
            class="absolute -right-1/4 -top-1/4 h-[600px] w-[600px] rounded-full bg-amber-500/5 blur-3xl dark:bg-amber-500/10"
          />
          <div
            class="absolute -bottom-1/4 -left-1/4 h-[400px] w-[400px] rounded-full bg-amber-500/5 blur-3xl dark:bg-amber-500/10"
          />
          <!-- Grid pattern -->
          <div
            class="absolute inset-0 bg-[linear-gradient(to_right,transparent_0%,transparent_49%,var(--border)_49%,var(--border)_51%,transparent_51%,transparent_100%)] bg-[length:80px_80px] opacity-30"
          />
          <div
            class="absolute inset-0 bg-[linear-gradient(to_bottom,transparent_0%,transparent_49%,var(--border)_49%,var(--border)_51%,transparent_51%,transparent_100%)] bg-[length:80px_80px] opacity-30"
          />
        </div>

        <div class="mx-auto max-w-6xl px-6 pb-24 pt-20 md:pb-32 md:pt-28">
          <div class="flex flex-col items-start">
            <!-- Badge -->
            <div
              class="animate-fade-in mb-8 inline-flex items-center gap-2 rounded-full border border-amber-500/30 bg-amber-500/10 px-4 py-1.5 text-sm text-amber-700 dark:text-amber-400"
            >
              <Zap class="size-3.5" />
              <span v-if="!noBS">Workflow automation for modern teams</span>
              <span v-else>It's a task manager</span>
            </div>

            <!-- Main headline -->
            <h1
              class="animate-fade-in-up font-display text-5xl font-bold leading-[1.1] tracking-tight md:text-6xl lg:text-7xl"
              style="animation-delay: 100ms"
            >
              <template v-if="!noBS">
                Bureaucracy<br />
                <span class="relative">
                  That Actually
                  <span
                    class="relative z-10 text-amber-600 dark:text-amber-500"
                  >
                    Moves</span
                  >
                  <svg
                    class="absolute -bottom-2 left-0 w-full"
                    viewBox="0 0 300 12"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                      d="M2 10C50 2 150 2 298 10"
                      stroke="currentColor"
                      stroke-width="3"
                      stroke-linecap="round"
                      class="text-amber-500/40"
                    />
                  </svg>
                </span>
              </template>
              <template v-else>
                A No-Nonsense<br />
                <span class="relative">
                  <span class="relative z-10 text-amber-600 dark:text-amber-500"
                    >Task Manager</span
                  >
                  <svg
                    class="absolute -bottom-2 left-0 w-full"
                    viewBox="0 0 300 12"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                      d="M2 10C50 2 150 2 298 10"
                      stroke="currentColor"
                      stroke-width="3"
                      stroke-linecap="round"
                      class="text-amber-500/40"
                    />
                  </svg>
                </span>
              </template>
            </h1>

            <p
              class="animate-fade-in-up mt-8 max-w-xl text-lg leading-relaxed text-muted-foreground md:text-xl"
              style="animation-delay: 200ms"
            >
              <template v-if="!noBS">
                Route approvals intelligently. Track everything automatically.
                Ship faster without sacrificing compliance.
              </template>
              <template v-else>
                Create tasks. Assign people. Track progress. That's it.
              </template>
            </p>

            <!-- No BS button -->
            <Button
              variant="outline"
              size="sm"
              class="animate-fade-in-up mt-6 gap-2 border-amber-500/30 text-amber-700 hover:bg-amber-500/10 hover:text-amber-800 dark:text-amber-400 dark:hover:text-amber-300"
              style="animation-delay: 300ms"
              @click="noBS = !noBS"
            >
              <Sparkles class="size-3.5" />
              {{ noBS ? "Add the BS back" : "Cut the BS" }}
            </Button>
          </div>
        </div>
      </section>

      <!-- Stats Section -->
      <section id="features" class="border-y border-border/50 bg-muted/30">
        <div class="mx-auto max-w-6xl px-6">
          <div
            class="grid divide-y divide-border/50 md:grid-cols-3 md:divide-x md:divide-y-0"
          >
            <div
              v-for="(feature, index) in features"
              :key="feature.title"
              class="group relative px-6 py-12 transition-colors hover:bg-muted/50 md:px-10"
              :style="`animation-delay: ${index * 100}ms`"
            >
              <div class="mb-6 flex items-center justify-between">
                <div
                  class="flex size-12 items-center justify-center rounded-xl bg-amber-500/10 text-amber-600 transition-colors group-hover:bg-amber-500/20 dark:text-amber-500"
                >
                  <component :is="feature.icon" class="size-5" />
                </div>
                <span
                  class="font-mono text-4xl font-bold tracking-tighter text-foreground/10 transition-colors group-hover:text-amber-500/30"
                >
                  {{ feature.stat }}
                </span>
              </div>
              <h3 class="text-lg font-semibold">{{ feature.title }}</h3>
              <p class="mt-2 text-sm leading-relaxed text-muted-foreground">
                {{ feature.description }}
              </p>
            </div>
          </div>
        </div>
      </section>

      <!-- CTA Section -->
      <section
        class="relative overflow-hidden border-t border-border/50 py-24 md:py-32"
      >
        <!-- Background -->
        <div class="absolute inset-0 -z-10">
          <div
            class="absolute inset-0 bg-gradient-to-br from-amber-500/5 via-transparent to-amber-500/5"
          />
          <div
            class="absolute -right-40 -top-40 h-80 w-80 rounded-full bg-amber-500/10 blur-3xl"
          />
        </div>

        <div class="mx-auto max-w-6xl px-6">
          <div class="mx-auto max-w-2xl text-center">
            <h2
              class="font-display text-3xl font-bold tracking-tight md:text-4xl"
            >
              <template v-if="!noBS">
                Ready to make bureaucracy<br />work for you?
              </template>
              <template v-else> Ready to get things done? </template>
            </h2>
            <p class="mt-6 text-lg text-muted-foreground">
              <template v-if="!noBS">
                Join hundreds of teams that have transformed their approval
                workflows.
              </template>
              <template v-else>
                Sign up. Create a project. Start tracking.
              </template>
            </p>
          </div>
        </div>
      </section>
    </main>

    <!-- Footer -->
    <footer class="border-t border-border/50">
      <div class="mx-auto max-w-6xl px-6 py-8">
        <div
          class="flex flex-col items-center justify-between gap-4 md:flex-row"
        >
          <div class="flex items-center gap-2">
            <BurecatLogo :size="20" />
            <span class="text-sm text-muted-foreground">{{ appName }}</span>
          </div>
          <p class="text-center text-sm text-muted-foreground">
            Crafted by
            <a
              href="https://codingcoffee.dev"
              target="_blank"
              rel="noopener noreferrer"
              class="underline underline-offset-4 hover:text-foreground transition-colors"
              >codingcoffee</a
            >
          </p>
        </div>
      </div>
    </footer>
  </div>
</template>

<style>
@import url("https://fonts.googleapis.com/css2?family=DM+Sans:ital,opsz,wght@0,9..40,400;0,9..40,500;0,9..40,600;0,9..40,700&display=swap");

.font-display {
  font-family: "DM Sans", system-ui, sans-serif;
}

@keyframes fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes fade-in-up {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fade-in {
  animation: fade-in 0.6s ease-out forwards;
  opacity: 0;
}

.animate-fade-in-up {
  animation: fade-in-up 0.6s ease-out forwards;
  opacity: 0;
}
</style>

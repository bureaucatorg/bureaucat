<script setup lang="ts">
import {
  Clock,
  Shield,
  Eye,
  ArrowRight,
  Zap,
  Users,
  GitBranch,
  CheckCircle2,
  BarChart3,
  Lock,
  Workflow,
  Bell
} from "lucide-vue-next";

const features = [
  {
    icon: Clock,
    stat: "60%",
    title: "Less time chasing",
    description: "Automated routing gets requests to the right people instantly. No more email chains.",
  },
  {
    icon: Shield,
    stat: "100%",
    title: "Audit coverage",
    description: "Every action logged, every decision tracked. Complete compliance without overhead.",
  },
  {
    icon: Eye,
    stat: "0",
    title: "Requests lost",
    description: "Real-time visibility into every request. Nothing falls through the cracks.",
  },
];

const capabilities = [
  {
    icon: Workflow,
    title: "Visual Workflow Builder",
    description: "Design approval chains with drag-and-drop simplicity. No code required.",
  },
  {
    icon: GitBranch,
    title: "Conditional Routing",
    description: "Smart rules that route requests based on amount, department, or custom criteria.",
  },
  {
    icon: Users,
    title: "Delegation & Escalation",
    description: "Automatic escalation when approvers are away. Delegation with full audit trails.",
  },
  {
    icon: Bell,
    title: "Smart Notifications",
    description: "Right message, right channel, right time. Slack, email, or in-app.",
  },
  {
    icon: BarChart3,
    title: "Analytics Dashboard",
    description: "Bottleneck detection, approval times, and team performance at a glance.",
  },
  {
    icon: Lock,
    title: "Enterprise Security",
    description: "SSO, role-based access, and encryption at rest. SOC 2 compliant.",
  },
];

const steps = [
  { number: "01", title: "Define", description: "Map your approval process" },
  { number: "02", title: "Connect", description: "Integrate with existing tools" },
  { number: "03", title: "Launch", description: "Go live in minutes" },
];
</script>

<template>
  <div class="relative flex min-h-screen flex-col overflow-hidden">
    <!-- Grain overlay -->
    <div class="pointer-events-none fixed inset-0 z-50 opacity-[0.015] dark:opacity-[0.03]"
         style="background-image: url('data:image/svg+xml,%3Csvg viewBox=%220 0 256 256%22 xmlns=%22http://www.w3.org/2000/svg%22%3E%3Cfilter id=%22noise%22%3E%3CfeTurbulence type=%22fractalNoise%22 baseFrequency=%220.9%22 numOctaves=%224%22 stitchTiles=%22stitch%22/%3E%3C/filter%3E%3Crect width=%22100%25%22 height=%22100%25%22 filter=%22url(%23noise)%22/%3E%3C/svg%3E');" />

    <!-- Header -->
    <Navbar />

    <main class="flex-1">
      <!-- Hero Section -->
      <section class="relative">
        <!-- Background decoration -->
        <div class="absolute inset-0 -z-10 overflow-hidden">
          <div class="absolute -right-1/4 -top-1/4 h-[600px] w-[600px] rounded-full bg-amber-500/5 blur-3xl dark:bg-amber-500/10" />
          <div class="absolute -bottom-1/4 -left-1/4 h-[400px] w-[400px] rounded-full bg-amber-500/5 blur-3xl dark:bg-amber-500/10" />
          <!-- Grid pattern -->
          <div class="absolute inset-0 bg-[linear-gradient(to_right,transparent_0%,transparent_49%,var(--border)_49%,var(--border)_51%,transparent_51%,transparent_100%)] bg-[length:80px_80px] opacity-30" />
          <div class="absolute inset-0 bg-[linear-gradient(to_bottom,transparent_0%,transparent_49%,var(--border)_49%,var(--border)_51%,transparent_51%,transparent_100%)] bg-[length:80px_80px] opacity-30" />
        </div>

        <div class="mx-auto max-w-6xl px-6 pb-24 pt-20 md:pb-32 md:pt-28">
          <div class="flex flex-col items-start">
            <!-- Badge -->
            <div class="animate-fade-in mb-8 inline-flex items-center gap-2 rounded-full border border-amber-500/30 bg-amber-500/10 px-4 py-1.5 text-sm text-amber-700 dark:text-amber-400">
              <Zap class="size-3.5" />
              <span>Workflow automation for modern teams</span>
            </div>

            <!-- Main headline -->
            <h1 class="animate-fade-in-up font-display text-5xl font-bold leading-[1.1] tracking-tight md:text-6xl lg:text-7xl" style="animation-delay: 100ms">
              Bureaucracy<br />
              <span class="relative">
                That Actually
                <span class="relative z-10 text-amber-600 dark:text-amber-500"> Moves</span>
                <svg class="absolute -bottom-2 left-0 w-full" viewBox="0 0 300 12" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M2 10C50 2 150 2 298 10" stroke="currentColor" stroke-width="3" stroke-linecap="round" class="text-amber-500/40"/>
                </svg>
              </span>
            </h1>

            <p class="animate-fade-in-up mt-8 max-w-xl text-lg leading-relaxed text-muted-foreground md:text-xl" style="animation-delay: 200ms">
              Route approvals intelligently. Track everything automatically.
              Ship faster without sacrificing compliance.
            </p>

            <div class="animate-fade-in-up mt-10 flex flex-col gap-4 sm:flex-row" style="animation-delay: 300ms">
              <Button size="lg" class="group gap-2 bg-foreground px-6 text-background hover:bg-foreground/90">
                Start Free Trial
                <ArrowRight class="size-4 transition-transform group-hover:translate-x-1" />
              </Button>
              <Button variant="outline" size="lg" class="px-6">
                Watch Demo
              </Button>
            </div>

            <!-- Social proof -->
            <div class="animate-fade-in-up mt-16 flex items-center gap-6 border-t border-border/50 pt-8" style="animation-delay: 400ms">
              <div class="flex -space-x-2">
                <div v-for="i in 4" :key="i" class="flex size-10 items-center justify-center rounded-full border-2 border-background bg-muted text-xs font-medium">
                  {{ ['JD', 'AK', 'MR', 'SL'][i-1] }}
                </div>
              </div>
              <div class="text-sm">
                <span class="font-semibold">500+</span>
                <span class="text-muted-foreground"> teams ship faster with Bureaucat</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Stats Section -->
      <section id="features" class="border-y border-border/50 bg-muted/30">
        <div class="mx-auto max-w-6xl px-6">
          <div class="grid divide-y divide-border/50 md:grid-cols-3 md:divide-x md:divide-y-0">
            <div v-for="(feature, index) in features" :key="feature.title"
                 class="group relative px-6 py-12 transition-colors hover:bg-muted/50 md:px-10"
                 :style="`animation-delay: ${index * 100}ms`">
              <div class="mb-6 flex items-center justify-between">
                <div class="flex size-12 items-center justify-center rounded-xl bg-amber-500/10 text-amber-600 transition-colors group-hover:bg-amber-500/20 dark:text-amber-500">
                  <component :is="feature.icon" class="size-5" />
                </div>
                <span class="font-mono text-4xl font-bold tracking-tighter text-foreground/10 transition-colors group-hover:text-amber-500/30">
                  {{ feature.stat }}
                </span>
              </div>
              <h3 class="text-lg font-semibold">{{ feature.title }}</h3>
              <p class="mt-2 text-sm leading-relaxed text-muted-foreground">{{ feature.description }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- How it works -->
      <section id="how-it-works" class="relative py-24 md:py-32">
        <div class="mx-auto max-w-6xl px-6">
          <div class="mb-16 max-w-2xl">
            <span class="mb-4 inline-block font-mono text-sm uppercase tracking-widest text-amber-600 dark:text-amber-500">How it works</span>
            <h2 class="font-display text-3xl font-bold tracking-tight md:text-4xl">
              Three steps to workflow freedom
            </h2>
          </div>

          <div class="relative grid gap-8 md:grid-cols-3">
            <!-- Connecting line -->
            <div class="absolute left-0 right-0 top-12 hidden h-px bg-gradient-to-r from-transparent via-border to-transparent md:block" />

            <div v-for="(step, index) in steps" :key="step.number" class="relative">
              <div class="mb-6 flex items-center gap-4">
                <div class="relative flex size-12 items-center justify-center rounded-full border-2 border-amber-500/30 bg-background font-mono text-lg font-bold text-amber-600 dark:text-amber-500">
                  {{ step.number }}
                  <div class="absolute -inset-1 rounded-full bg-amber-500/10" />
                </div>
                <div v-if="index < steps.length - 1" class="h-px flex-1 bg-border md:hidden" />
              </div>
              <h3 class="text-xl font-semibold">{{ step.title }}</h3>
              <p class="mt-2 text-muted-foreground">{{ step.description }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- Capabilities Grid -->
      <section class="border-t border-border/50 bg-muted/20 py-24 md:py-32">
        <div class="mx-auto max-w-6xl px-6">
          <div class="mb-16 flex flex-col items-start justify-between gap-6 md:flex-row md:items-end">
            <div class="max-w-2xl">
              <span class="mb-4 inline-block font-mono text-sm uppercase tracking-widest text-amber-600 dark:text-amber-500">Capabilities</span>
              <h2 class="font-display text-3xl font-bold tracking-tight md:text-4xl">
                Everything you need to tame the chaos
              </h2>
            </div>
          </div>

          <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <Card v-for="capability in capabilities" :key="capability.title"
                  class="group border-border/50 bg-background/50 transition-all hover:border-amber-500/30 hover:shadow-lg hover:shadow-amber-500/5">
              <CardHeader class="pb-4">
                <div class="mb-4 flex size-11 items-center justify-center rounded-lg bg-muted transition-colors group-hover:bg-amber-500/10">
                  <component :is="capability.icon" class="size-5 text-muted-foreground transition-colors group-hover:text-amber-600 dark:group-hover:text-amber-500" />
                </div>
                <CardTitle class="text-base font-semibold">{{ capability.title }}</CardTitle>
              </CardHeader>
              <CardContent class="pt-0">
                <CardDescription class="text-sm leading-relaxed">{{ capability.description }}</CardDescription>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      <!-- CTA Section -->
      <section class="relative overflow-hidden border-t border-border/50 py-24 md:py-32">
        <!-- Background -->
        <div class="absolute inset-0 -z-10">
          <div class="absolute inset-0 bg-gradient-to-br from-amber-500/5 via-transparent to-amber-500/5" />
          <div class="absolute -right-40 -top-40 h-80 w-80 rounded-full bg-amber-500/10 blur-3xl" />
        </div>

        <div class="mx-auto max-w-6xl px-6">
          <div class="mx-auto max-w-2xl text-center">
            <h2 class="font-display text-3xl font-bold tracking-tight md:text-4xl">
              Ready to make bureaucracy<br />work for you?
            </h2>
            <p class="mt-6 text-lg text-muted-foreground">
              Join hundreds of teams that have transformed their approval workflows.
              Start free, no credit card required.
            </p>
            <div class="mt-10 flex flex-col justify-center gap-4 sm:flex-row">
              <Button size="lg" class="group gap-2 bg-foreground px-8 text-background hover:bg-foreground/90">
                Get Started Free
                <ArrowRight class="size-4 transition-transform group-hover:translate-x-1" />
              </Button>
              <Button variant="outline" size="lg" class="px-8">
                Talk to Sales
              </Button>
            </div>
            <p class="mt-6 flex items-center justify-center gap-2 text-sm text-muted-foreground">
              <CheckCircle2 class="size-4 text-green-500" />
              14-day free trial, cancel anytime
            </p>
          </div>
        </div>
      </section>
    </main>

    <!-- Footer -->
    <footer class="border-t border-border/50">
      <div class="mx-auto max-w-6xl px-6 py-8">
        <div class="flex flex-col items-center justify-between gap-4 md:flex-row">
          <div class="flex items-center gap-3">
            <div class="flex size-6 items-center justify-center rounded bg-foreground">
              <span class="font-mono text-xs font-bold text-background">B</span>
            </div>
            <span class="text-sm text-muted-foreground">Bureaucat</span>
          </div>
          <p class="text-center text-sm text-muted-foreground">
            Built with &lt;3
          </p>
        </div>
      </div>
    </footer>
  </div>
</template>

<style>
@import url('https://fonts.googleapis.com/css2?family=DM+Sans:ital,opsz,wght@0,9..40,400;0,9..40,500;0,9..40,600;0,9..40,700&display=swap');

.font-display {
  font-family: 'DM Sans', system-ui, sans-serif;
}

@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
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

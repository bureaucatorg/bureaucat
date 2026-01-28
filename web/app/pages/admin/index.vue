<script setup lang="ts">
import { Users, Key, Shield, ArrowRight } from "lucide-vue-next";

definePageMeta({
  middleware: ["admin"],
});

const adminModels = [
  {
    title: "Users",
    description: "Manage user accounts, create new users, and control access levels",
    icon: Users,
    href: "/admin/model/users",
    color: "text-blue-500",
    bgColor: "bg-blue-500/10",
  },
  {
    title: "Tokens",
    description: "Monitor active sessions, revoke tokens, and cleanup expired sessions",
    icon: Key,
    href: "/admin/model/tokens",
    color: "text-amber-500",
    bgColor: "bg-amber-500/10",
  },
];
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-12">
        <div class="mb-8">
          <div class="flex items-center gap-3 mb-2">
            <div class="flex size-10 items-center justify-center rounded-lg bg-foreground">
              <Shield class="size-5 text-background" />
            </div>
            <h1 class="text-3xl font-bold tracking-tight">Admin Dashboard</h1>
          </div>
          <p class="text-muted-foreground">
            Manage your application's data and settings
          </p>
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <NuxtLink
            v-for="model in adminModels"
            :key="model.title"
            :to="model.href"
            class="group"
          >
            <Card class="h-full transition-all hover:border-foreground/20 hover:shadow-lg">
              <CardHeader>
                <div class="flex items-center justify-between">
                  <div :class="[model.bgColor, 'flex size-12 items-center justify-center rounded-lg']">
                    <component :is="model.icon" :class="['size-6', model.color]" />
                  </div>
                  <ArrowRight class="size-5 text-muted-foreground transition-transform group-hover:translate-x-1" />
                </div>
                <CardTitle class="mt-4">{{ model.title }}</CardTitle>
                <CardDescription>{{ model.description }}</CardDescription>
              </CardHeader>
            </Card>
          </NuxtLink>
        </div>
      </div>
    </main>
  </div>
</template>

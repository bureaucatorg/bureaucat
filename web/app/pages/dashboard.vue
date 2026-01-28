<script setup lang="ts">
import { User, Mail, Calendar, Shield } from "lucide-vue-next";

definePageMeta({
  middleware: ["auth"],
});

const { user } = useAuth();

const formattedCreatedAt = computed(() => {
  if (!user.value?.created_at) return "";
  return new Date(user.value.created_at).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main class="flex-1">
      <div class="mx-auto max-w-4xl px-6 py-12">
        <!-- Welcome Section -->
        <div class="mb-8">
          <h1 class="text-3xl font-bold tracking-tight">
            Welcome back, {{ user?.first_name }}!
          </h1>
          <p class="mt-2 text-muted-foreground">
            Here's an overview of your account
          </p>
        </div>

        <!-- Profile Card -->
        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <User class="size-5" />
              Profile Information
            </CardTitle>
            <CardDescription>Your account details</CardDescription>
          </CardHeader>
          <CardContent class="space-y-6">
            <div class="flex items-center gap-4">
              <div class="flex size-16 items-center justify-center rounded-full bg-muted text-2xl font-semibold">
                {{ user?.first_name?.[0] }}{{ user?.last_name?.[0] }}
              </div>
              <div>
                <h3 class="text-lg font-semibold">
                  {{ user?.first_name }} {{ user?.last_name }}
                </h3>
                <p class="text-sm text-muted-foreground">@{{ user?.username }}</p>
              </div>
            </div>

            <div class="grid gap-4 sm:grid-cols-2">
              <div class="flex items-center gap-3 rounded-lg border p-4">
                <div class="flex size-10 items-center justify-center rounded-lg bg-muted">
                  <Mail class="size-5 text-muted-foreground" />
                </div>
                <div>
                  <p class="text-sm text-muted-foreground">Email</p>
                  <p class="font-medium">{{ user?.email }}</p>
                </div>
              </div>

              <div class="flex items-center gap-3 rounded-lg border p-4">
                <div class="flex size-10 items-center justify-center rounded-lg bg-muted">
                  <Shield class="size-5 text-muted-foreground" />
                </div>
                <div>
                  <p class="text-sm text-muted-foreground">Account Type</p>
                  <p class="font-medium capitalize">{{ user?.user_type }}</p>
                </div>
              </div>

              <div class="flex items-center gap-3 rounded-lg border p-4 sm:col-span-2">
                <div class="flex size-10 items-center justify-center rounded-lg bg-muted">
                  <Calendar class="size-5 text-muted-foreground" />
                </div>
                <div>
                  <p class="text-sm text-muted-foreground">Member Since</p>
                  <p class="font-medium">{{ formattedCreatedAt }}</p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  </div>
</template>

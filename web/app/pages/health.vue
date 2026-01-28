<script setup lang="ts">
import { CheckCircle, XCircle, RefreshCw } from "lucide-vue-next";

interface HealthResponse {
  all: boolean;
  db: boolean;
  api: boolean;
}

const health = ref<HealthResponse | null>(null);
const loading = ref(true);
const error = ref<string | null>(null);

async function fetchHealth() {
  loading.value = true;
  error.value = null;

  try {
    const response = await fetch("/api/v1/ht/");
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    health.value = await response.json();
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Failed to fetch health status";
    health.value = null;
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  fetchHealth();
});

const services = computed(() => [
  { name: "API", status: health.value?.api ?? false },
  { name: "Database", status: health.value?.db ?? false },
]);
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <header class="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div class="mx-auto flex h-16 max-w-5xl items-center justify-between px-4">
        <NuxtLink to="/" class="flex items-center gap-2">
          <span class="text-xl font-bold">Bureaucat</span>
        </NuxtLink>
        <ThemeToggle />
      </div>
    </header>

    <main class="flex-1">
      <div class="mx-auto max-w-2xl px-4 py-16">
        <div class="mb-8 flex items-center justify-between">
          <h1 class="text-3xl font-bold">System Health</h1>
          <Button variant="outline" size="sm" :disabled="loading" @click="fetchHealth">
            <RefreshCw :class="['mr-2 size-4', { 'animate-spin': loading }]" />
            Refresh
          </Button>
        </div>

        <div v-if="loading && !health" class="space-y-4">
          <Card>
            <CardHeader>
              <div class="flex items-center gap-3">
                <div class="size-6 animate-pulse rounded-full bg-muted" />
                <div class="h-6 w-32 animate-pulse rounded bg-muted" />
              </div>
            </CardHeader>
          </Card>
        </div>

        <div v-else-if="error" class="space-y-4">
          <Card class="border-destructive">
            <CardHeader>
              <div class="flex items-center gap-3">
                <XCircle class="size-6 text-destructive" />
                <CardTitle>Connection Error</CardTitle>
              </div>
              <CardDescription>{{ error }}</CardDescription>
            </CardHeader>
          </Card>
        </div>

        <div v-else class="space-y-4">
          <Card :class="health?.all ? 'border-green-500' : 'border-destructive'">
            <CardHeader>
              <div class="flex items-center gap-3">
                <CheckCircle v-if="health?.all" class="size-6 text-green-500" />
                <XCircle v-else class="size-6 text-destructive" />
                <CardTitle>{{ health?.all ? "All Systems Operational" : "System Issues Detected" }}</CardTitle>
              </div>
            </CardHeader>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Services</CardTitle>
            </CardHeader>
            <CardContent>
              <div class="space-y-3">
                <div v-for="service in services" :key="service.name" class="flex items-center justify-between">
                  <span class="font-medium">{{ service.name }}</span>
                  <Badge :variant="service.status ? 'default' : 'destructive'" :class="service.status ? 'bg-green-500 hover:bg-green-500' : ''">
                    {{ service.status ? "Healthy" : "Unhealthy" }}
                  </Badge>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </main>
  </div>
</template>

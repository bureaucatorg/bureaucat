<script setup lang="ts">
import type { NuxtError } from "#app";

const props = defineProps<{ error: NuxtError }>();

const isNotFound = computed(() => props.error?.statusCode === 404);

const title = computed(() =>
  isNotFound.value ? "This page isn't filed anywhere" : "Something went wrong"
);

const message = computed(() =>
  isNotFound.value
    ? "The clerk searched every drawer twice. There's no page at this address — it may have been moved, renamed, or never existed."
    : props.error?.message || "An unexpected error occurred while processing your request."
);

const route = useRoute();

useHead({ title: computed(() => (isNotFound.value ? "404 · Not found" : "Error")) });

// Leaving the error page needs clearError(), but the links must stay real
// anchors — so let the browser handle modifier/middle clicks itself.
function navigate(event: MouseEvent, to: string) {
  if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
    return;
  }
  event.preventDefault();
  clearError({ redirect: to });
}
</script>

<template>
  <div class="min-h-screen bg-background text-foreground">
    <header class="flex items-center gap-2 px-6 py-4">
      <NuxtLink to="/" class="flex items-center gap-2" @click="navigate($event, '/')">
        <img src="/logo.svg" alt="" class="size-7" >
        <span class="text-lg font-bold">Bureaucat</span>
      </NuxtLink>
    </header>

    <main id="main-content">
      <NotFoundState
        :code="error?.statusCode || 500"
        :title="title"
        :message="message"
        :reference="route.fullPath"
        :stamp="isNotFound ? 'NOT ON FILE' : 'REJECTED'"
      >
        <template #actions>
          <Button as-child>
            <NuxtLink to="/" @click="navigate($event, '/')">Back to home</NuxtLink>
          </Button>
          <Button variant="outline" as-child>
            <NuxtLink to="/projects" @click="navigate($event, '/projects')">
              Browse projects
            </NuxtLink>
          </Button>
        </template>
      </NotFoundState>
    </main>
  </div>
</template>

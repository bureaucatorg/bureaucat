<script setup lang="ts">
import { Loader2, Calendar, Mail, AtSign } from "lucide-vue-next";

definePageMeta({
  middleware: ["auth"],
});

const route = useRoute();
const userId = computed(() => route.params.id as string);

const { getAuthHeader } = useAuth();

const loading = ref(true);
const error = ref<string | null>(null);
const user = ref<{
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  user_type: string;
  created_at: string;
} | null>(null);

useHead({
  title: computed(() => {
    if (user.value) return `${user.value.first_name} ${user.value.last_name}`;
    return "Profile";
  }),
});

async function loadUser() {
  loading.value = true;
  error.value = null;

  try {
    const response = await fetch(`/api/v1/users/${userId.value}`, {
      headers: getAuthHeader(),
    });

    if (!response.ok) {
      error.value = "User not found";
      return;
    }

    user.value = await response.json();
  } catch {
    error.value = "Failed to load user";
  } finally {
    loading.value = false;
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

onMounted(() => {
  loadUser();
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex-1">
      <div class="mx-auto max-w-2xl px-6 py-8">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-20">
          <Loader2 class="size-8 animate-spin text-muted-foreground" />
        </div>

        <!-- Error -->
        <div
          v-else-if="error"
          class="flex flex-col items-center justify-center py-20"
        >
          <p class="text-lg text-destructive">{{ error }}</p>
          <Button class="mt-4" variant="outline" as-child>
            <NuxtLink to="/projects">Back to Projects</NuxtLink>
          </Button>
        </div>

        <!-- Profile -->
        <template v-else-if="user">
          <div class="flex items-start gap-6">
            <Avatar class="size-20">
              <AvatarFallback class="text-2xl">
                {{ user.first_name?.[0] }}{{ user.last_name?.[0] }}
              </AvatarFallback>
            </Avatar>

            <div class="flex-1">
              <h1 class="text-2xl font-bold">
                {{ user.first_name }} {{ user.last_name }}
              </h1>

              <div class="mt-3 space-y-2 text-sm text-muted-foreground">
                <div class="flex items-center gap-2">
                  <AtSign class="size-4" />
                  <span>{{ user.username }}</span>
                </div>
                <div class="flex items-center gap-2">
                  <Mail class="size-4" />
                  <span>{{ user.email }}</span>
                </div>
                <div class="flex items-center gap-2">
                  <Calendar class="size-4" />
                  <span>Member since {{ formatDate(user.created_at) }}</span>
                </div>
              </div>

              <Badge class="mt-3 capitalize" variant="secondary">
                {{ user.user_type }}
              </Badge>
            </div>
          </div>
        </template>
      </div>
    </main>
  </div>
</template>

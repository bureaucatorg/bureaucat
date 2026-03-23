<script setup lang="ts">
import { Loader2 } from "lucide-vue-next";

definePageMeta({
  middleware: ["guest"],
});

useSeoMeta({ title: "Authenticating..." });

const { refreshToken } = useAuth();
const route = useRoute();

const error = ref<string | null>(null);
const loading = ref(true);

onMounted(async () => {
  const status = route.query.status as string;
  const message = route.query.message as string;

  if (status === "error") {
    error.value = message || "SSO authentication failed";
    loading.value = false;
    return;
  }

  // After SSO callback, backend has set httpOnly cookies.
  // Call refreshToken() to exchange the refresh cookie for auth state.
  const success = await refreshToken();
  loading.value = false;

  if (success) {
    await navigateTo("/dashboard");
  } else {
    error.value = "Failed to complete authentication. Please try again.";
  }
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex flex-1 items-center justify-center px-4 py-12">
      <Card class="w-full max-w-md">
        <CardContent class="pt-6 text-center">
          <div v-if="loading" class="space-y-4 py-8">
            <Loader2 class="mx-auto size-8 animate-spin text-muted-foreground" />
            <p class="text-sm text-muted-foreground">Completing sign in...</p>
          </div>
          <div v-else-if="error" class="space-y-4 py-8">
            <p class="text-sm text-destructive">{{ error }}</p>
            <Button as-child variant="outline">
              <NuxtLink to="/signin">Back to Sign In</NuxtLink>
            </Button>
          </div>
        </CardContent>
      </Card>
    </main>
  </div>
</template>

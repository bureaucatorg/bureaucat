<script setup lang="ts">
import { Loader2 } from "lucide-vue-next";

definePageMeta({
  middleware: ["guest"],
});

const { refreshToken } = useAuth();
const route = useRoute();

const error = ref<{ code: number; title: string; message: string; reference: string; stamp: string; domain?: string } | null>(null);
const loading = ref(true);

useSeoMeta({ title: computed(() => (error.value ? "Sign in failed" : "Authenticating...")) });

onMounted(async () => {
  const status = route.query.status as string;

  if (status === "error") {
    const domain = (route.query.domain as string) || "";
    error.value =
      route.query.reason === "domain_not_allowed"
        ? {
            code: 403,
            title: "Your domain isn't on the list",
            message: `Accounts from ${domain ? `@${domain}` : "this domain"} aren't permitted here. Sign in with your organisation's account, or ask an admin to add your domain.`,
            reference: domain ? `@${domain}` : "",
            stamp: "NOT PERMITTED",
            domain,
          }
        : {
            code: 401,
            title: "Sign in didn't go through",
            message: (route.query.message as string) || "SSO authentication failed. Please try again.",
            reference: "",
            stamp: "REJECTED",
          };
    loading.value = false;
    return;
  }

  // After SSO callback, backend has set httpOnly cookies.
  // Call refreshToken() to exchange the refresh cookie for auth state.
  const success = await refreshToken();
  loading.value = false;

  if (success) {
    await navigateTo(safeRedirect(route.query.redirect));
  } else {
    error.value = {
      code: 401,
      title: "Sign in didn't go through",
      message: "Failed to complete authentication. Please try again.",
      reference: "",
      stamp: "REJECTED",
    };
  }
});
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex flex-1 flex-col">
      <NotFoundState
        v-if="error"
        :code="error.code"
        :title="error.title"
        :message="error.message"
        :reference="error.reference"
        :stamp="error.stamp"
      >
        <template v-if="error.domain" #message>
          Accounts from <strong class="font-semibold text-foreground">@{{ error.domain }}</strong> aren't permitted here.
          Sign in with your organisation's account, or ask an admin to add your domain.
        </template>
        <template #actions>
          <Button as-child>
            <NuxtLink to="/signin">Try another account</NuxtLink>
          </Button>
        </template>
      </NotFoundState>

      <div v-else class="flex flex-1 items-center justify-center px-4 py-12">
        <Card class="w-full max-w-md">
          <CardContent class="pt-6 text-center">
            <div class="space-y-4 py-8">
              <Loader2 class="mx-auto size-8 animate-spin text-muted-foreground" />
              <p class="text-sm text-muted-foreground">Completing sign in...</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  </div>
</template>

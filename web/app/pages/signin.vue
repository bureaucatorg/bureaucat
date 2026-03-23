<script setup lang="ts">
import { Eye, EyeOff, Loader2, Shield } from "lucide-vue-next";

definePageMeta({
  middleware: ["guest"],
});

useSeoMeta({ title: "Sign In" });

const { signin } = useAuth();
const { fetchSSOProviders, ssoProviders } = useSettings();
const route = useRoute();

const redirectTo = computed(() => {
  const r = route.query.redirect as string | undefined;
  // Only allow relative paths to prevent open redirect
  if (r && r.startsWith("/")) return r;
  return "/dashboard";
});

const identifier = ref("");
const password = ref("");
const showPassword = ref(false);
const loading = ref(false);
const error = ref<string | null>(null);

onMounted(async () => {
  await fetchSSOProviders();
});

const hasSSOProviders = computed(() => ssoProviders.value.google || ssoProviders.value.zitadel);

function startSSO(provider: string) {
  const params = new URLSearchParams();
  if (redirectTo.value !== "/dashboard") {
    params.set("redirect", redirectTo.value);
  }
  const qs = params.toString();
  window.location.href = `/api/v1/auth/sso/${provider}${qs ? `?${qs}` : ""}`;
}

async function handleSubmit() {
  error.value = null;

  if (!identifier.value || !password.value) {
    error.value = "Please fill in all fields";
    return;
  }

  loading.value = true;

  const result = await signin({
    identifier: identifier.value,
    password: password.value,
  });

  loading.value = false;

  if (result.success) {
    await navigateTo(redirectTo.value);
  } else {
    error.value = result.error || "Sign in failed";
  }
}
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main id="main-content" class="flex flex-1 items-center justify-center px-4 py-12">
      <Card class="w-full max-w-md">
        <CardHeader class="space-y-1 text-center">
          <CardTitle class="text-2xl font-bold">Welcome back</CardTitle>
          <CardDescription>Sign in to your account to continue</CardDescription>
        </CardHeader>
        <CardContent>
          <!-- SSO Buttons -->
          <div v-if="hasSSOProviders" class="space-y-3">
            <Button
              v-if="ssoProviders.google"
              variant="outline"
              class="w-full"
              @click="startSSO('google')"
            >
              <svg class="mr-2 size-4" viewBox="0 0 24 24">
                <path
                  d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"
                  fill="#4285F4"
                />
                <path
                  d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                  fill="#34A853"
                />
                <path
                  d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                  fill="#FBBC05"
                />
                <path
                  d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                  fill="#EA4335"
                />
              </svg>
              Sign in with Google
            </Button>
            <Button
              v-if="ssoProviders.zitadel"
              variant="outline"
              class="w-full"
              @click="startSSO('zitadel')"
            >
              <Shield class="mr-2 size-4" />
              Sign in with Zitadel
            </Button>

            <!-- Separator -->
            <div class="relative my-4">
              <div class="absolute inset-0 flex items-center">
                <Separator />
              </div>
              <div class="relative flex justify-center text-xs uppercase">
                <span class="bg-card px-2 text-muted-foreground">or continue with</span>
              </div>
            </div>
          </div>

          <form @submit.prevent="handleSubmit" class="space-y-4">
            <div v-if="error" role="alert" class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {{ error }}
            </div>

            <div class="space-y-2">
              <Label for="identifier">Email or Username</Label>
              <Input
                id="identifier"
                v-model="identifier"
                type="text"
                placeholder="you@example.com"
                required
                :disabled="loading"
              />
            </div>

            <div class="space-y-2">
              <Label for="password">Password</Label>
              <div class="relative">
                <Input
                  id="password"
                  v-model="password"
                  :type="showPassword ? 'text' : 'password'"
                  placeholder="Enter your password"
                  required
                  :disabled="loading"
                  class="pr-10"
                />
                <button
                  type="button"
                  :aria-label="showPassword ? 'Hide password' : 'Show password'"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 rounded-sm outline-none"
                  @click="showPassword = !showPassword"
                >
                  <Eye v-if="!showPassword" class="size-4" />
                  <EyeOff v-else class="size-4" />
                </button>
              </div>
            </div>

            <Button type="submit" class="w-full" :disabled="loading">
              <Loader2 v-if="loading" class="mr-2 size-4 animate-spin" />
              Sign In
            </Button>
          </form>
        </CardContent>
        <CardFooter class="flex justify-center">
          <p class="text-sm text-muted-foreground">
            Don't have an account?
            <NuxtLink to="/signup" class="text-foreground underline-offset-4 hover:underline">Sign up</NuxtLink>
          </p>
        </CardFooter>
      </Card>
    </main>
  </div>
</template>

<script setup lang="ts">
import { Eye, EyeOff, Loader2 } from "lucide-vue-next";

definePageMeta({
  middleware: ["guest"],
});

useSeoMeta({ title: "Sign In" });

const { signin } = useAuth();

const identifier = ref("");
const password = ref("");
const showPassword = ref(false);
const loading = ref(false);
const error = ref<string | null>(null);

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
    await navigateTo("/dashboard");
  } else {
    error.value = result.error || "Sign in failed";
  }
}
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main class="flex flex-1 items-center justify-center px-4 py-12">
      <Card class="w-full max-w-md">
        <CardHeader class="space-y-1 text-center">
          <CardTitle class="text-2xl font-bold">Welcome back</CardTitle>
          <CardDescription>Sign in to your account to continue</CardDescription>
        </CardHeader>
        <CardContent>
          <form @submit.prevent="handleSubmit" class="space-y-4">
            <div v-if="error" class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
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
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
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

<script setup lang="ts">
import { Eye, EyeOff, Loader2, Check, X } from "lucide-vue-next";
import { validatePassword, validateEmail } from "~/utils/validators";

definePageMeta({
  middleware: ["guest"],
});

useSeoMeta({ title: "Sign Up" });

const { signup } = useAuth();
const { signupSettings, fetchSignupSettings } = useSettings();
const runtimeConfig = useRuntimeConfig();

onMounted(() => {
  fetchSignupSettings();
});

const isProduction = computed(() => runtimeConfig.public.nodeEnv === "production");

const form = reactive({
  firstName: "",
  lastName: "",
  username: "",
  email: "",
  password: "",
  confirmPassword: "",
});

const showPassword = ref(false);
const showConfirmPassword = ref(false);
const loading = ref(false);
const error = ref<string | null>(null);

const passwordValidation = computed(() => validatePassword(form.password, isProduction.value));

const passwordRequirements = computed(() => [
  { label: "At least 8 characters", met: form.password.length >= 8 },
  { label: "One uppercase letter", met: /[A-Z]/.test(form.password) },
  { label: "One lowercase letter", met: /[a-z]/.test(form.password) },
  { label: "One number", met: /[0-9]/.test(form.password) },
  { label: "One special character", met: /[!@#$%^&*(),.?":{}|<>]/.test(form.password) },
]);

const formErrors = computed(() => {
  const errors: Record<string, string> = {};

  if (form.email && !validateEmail(form.email)) {
    errors.email = "Please enter a valid email address";
  }

  if (isProduction.value && form.password && !passwordValidation.value.isValid) {
    errors.password = "Password does not meet requirements";
  }

  if (form.confirmPassword && form.password !== form.confirmPassword) {
    errors.confirmPassword = "Passwords do not match";
  }

  return errors;
});

const isFormValid = computed(() => {
  return (
    form.firstName &&
    form.lastName &&
    form.username &&
    form.email &&
    form.password &&
    form.confirmPassword &&
    validateEmail(form.email) &&
    (isProduction.value ? passwordValidation.value.isValid : form.password.length > 0) &&
    form.password === form.confirmPassword
  );
});

async function handleSubmit() {
  error.value = null;

  if (!isFormValid.value) {
    error.value = "Please fix the errors above";
    return;
  }

  loading.value = true;

  const result = await signup({
    first_name: form.firstName,
    last_name: form.lastName,
    username: form.username,
    email: form.email,
    password: form.password,
  });

  loading.value = false;

  if (result.success) {
    await navigateTo("/dashboard");
  } else {
    error.value = result.error || "Sign up failed";
  }
}
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <Navbar />

    <main class="flex flex-1 items-center justify-center px-4 py-12">
      <Card class="w-full max-w-md">
        <CardHeader class="space-y-1 text-center">
          <CardTitle class="text-2xl font-bold">Create an account</CardTitle>
          <CardDescription v-if="signupSettings.enabled">Enter your details to get started</CardDescription>
          <CardDescription v-else>New account registration is currently disabled</CardDescription>
        </CardHeader>
        <CardContent v-if="!signupSettings.enabled">
          <div class="rounded-md bg-muted p-4 text-center text-sm text-muted-foreground">
            Signups are not available at this time. Please contact your administrator for access.
          </div>
        </CardContent>
        <CardContent v-else>
          <form @submit.prevent="handleSubmit" class="space-y-4">
            <div v-if="error" class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {{ error }}
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <Label for="firstName">First Name</Label>
                <Input
                  id="firstName"
                  v-model="form.firstName"
                  type="text"
                  placeholder="John"
                  required
                  :disabled="loading"
                />
              </div>
              <div class="space-y-2">
                <Label for="lastName">Last Name</Label>
                <Input
                  id="lastName"
                  v-model="form.lastName"
                  type="text"
                  placeholder="Doe"
                  required
                  :disabled="loading"
                />
              </div>
            </div>

            <div class="space-y-2">
              <Label for="username">Username</Label>
              <Input
                id="username"
                v-model="form.username"
                type="text"
                placeholder="johndoe"
                required
                :disabled="loading"
              />
            </div>

            <div class="space-y-2">
              <Label for="email">Email</Label>
              <Input
                id="email"
                v-model="form.email"
                type="email"
                placeholder="you@example.com"
                required
                :disabled="loading"
                :class="{ 'border-destructive': formErrors.email }"
              />
              <p v-if="formErrors.email" class="text-xs text-destructive">{{ formErrors.email }}</p>
            </div>

            <div class="space-y-2">
              <Label for="password">Password</Label>
              <div class="relative">
                <Input
                  id="password"
                  v-model="form.password"
                  :type="showPassword ? 'text' : 'password'"
                  placeholder="Create a password"
                  required
                  :disabled="loading"
                  class="pr-10"
                  :class="{ 'border-destructive': formErrors.password }"
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

              <!-- Password requirements (production only) -->
              <div v-if="isProduction && form.password" class="mt-2 space-y-1">
                <div
                  v-for="req in passwordRequirements"
                  :key="req.label"
                  class="flex items-center gap-2 text-xs"
                  :class="req.met ? 'text-green-600 dark:text-green-500' : 'text-muted-foreground'"
                >
                  <Check v-if="req.met" class="size-3" />
                  <X v-else class="size-3" />
                  <span>{{ req.label }}</span>
                </div>
              </div>
            </div>

            <div class="space-y-2">
              <Label for="confirmPassword">Confirm Password</Label>
              <div class="relative">
                <Input
                  id="confirmPassword"
                  v-model="form.confirmPassword"
                  :type="showConfirmPassword ? 'text' : 'password'"
                  placeholder="Confirm your password"
                  required
                  :disabled="loading"
                  class="pr-10"
                  :class="{ 'border-destructive': formErrors.confirmPassword }"
                />
                <button
                  type="button"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  @click="showConfirmPassword = !showConfirmPassword"
                >
                  <Eye v-if="!showConfirmPassword" class="size-4" />
                  <EyeOff v-else class="size-4" />
                </button>
              </div>
              <p v-if="formErrors.confirmPassword" class="text-xs text-destructive">{{ formErrors.confirmPassword }}</p>
            </div>

            <Button type="submit" class="w-full" :disabled="loading || !isFormValid">
              <Loader2 v-if="loading" class="mr-2 size-4 animate-spin" />
              Create Account
            </Button>
          </form>
        </CardContent>
        <CardFooter class="flex justify-center">
          <p class="text-sm text-muted-foreground">
            <template v-if="signupSettings.enabled">
              Already have an account?
              <NuxtLink to="/signin" class="text-foreground underline-offset-4 hover:underline">Sign in</NuxtLink>
            </template>
            <template v-else>
              <NuxtLink to="/signin" class="text-foreground underline-offset-4 hover:underline">Sign in</NuxtLink>
              with an existing account
            </template>
          </p>
        </CardFooter>
      </Card>
    </main>
  </div>
</template>

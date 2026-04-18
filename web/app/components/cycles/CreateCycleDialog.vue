<script setup lang="ts">
import { Loader2, Calendar as CalendarIcon } from "lucide-vue-next";
import { toast } from "vue-sonner";
import { CalendarDate, type DateValue } from "@internationalized/date";
import type { Cycle } from "~/types";

const props = defineProps<{
  projectKey: string;
  // If provided, dialog behaves as an edit form for this cycle.
  cycle?: Cycle | null;
}>();
const open = defineModel<boolean>("open", { default: false });
const emit = defineEmits<{ created: []; saved: [Cycle] }>();

const { createCycle, updateCycle } = useCycles();

const isEdit = computed(() => !!props.cycle);

const loading = ref(false);
const error = ref<string | null>(null);
const form = ref<{
  title: string;
  description: string;
  startDate: DateValue | undefined;
  endDate: DateValue | undefined;
}>({
  title: "",
  description: "",
  startDate: undefined,
  endDate: undefined,
});

const startOpen = ref(false);
const endOpen = ref(false);

function isoToCalendar(s?: string): DateValue | undefined {
  if (!s) return undefined;
  const [y, m, d] = s.split("-").map(Number);
  if (!y || !m || !d) return undefined;
  return new CalendarDate(y, m, d);
}

function hydrate() {
  error.value = null;
  if (props.cycle) {
    form.value = {
      title: props.cycle.title,
      description: props.cycle.description ?? "",
      startDate: isoToCalendar(props.cycle.start_date),
      endDate: isoToCalendar(props.cycle.end_date),
    };
  } else {
    form.value = {
      title: "",
      description: "",
      startDate: undefined,
      endDate: undefined,
    };
  }
}

watch(open, (isOpen) => {
  if (isOpen) hydrate();
});

function formatDateValue(d: DateValue | undefined): string {
  if (!d) return "Pick a date";
  return `${d.year}-${String(d.month).padStart(2, "0")}-${String(d.day).padStart(2, "0")}`;
}

function toIsoDate(d: DateValue): string {
  return `${d.year}-${String(d.month).padStart(2, "0")}-${String(d.day).padStart(2, "0")}`;
}

const canSubmit = computed(
  () => !!form.value.title.trim() && !!form.value.startDate && !!form.value.endDate
);

async function handleSubmit() {
  if (!form.value.startDate || !form.value.endDate) return;
  error.value = null;

  const start = form.value.startDate as CalendarDate;
  const end = form.value.endDate as CalendarDate;
  if (end.compare(start) < 0) {
    error.value = "End date must be on or after start date";
    return;
  }

  loading.value = true;
  const payload = {
    title: form.value.title.trim(),
    description: form.value.description.trim() || undefined,
    start_date: toIsoDate(start),
    end_date: toIsoDate(end),
  };
  const result = isEdit.value
    ? await updateCycle(props.projectKey, props.cycle!.id, payload)
    : await createCycle(props.projectKey, payload);
  loading.value = false;

  if (result.success) {
    toast.success(isEdit.value ? "Cycle updated" : "Cycle created");
    open.value = false;
    if (isEdit.value && result.data) emit("saved", result.data);
    else emit("created");
  } else {
    error.value = result.error || (isEdit.value ? "Failed to update cycle" : "Failed to create cycle");
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>{{ isEdit ? "Edit cycle" : "New Cycle" }}</DialogTitle>
        <DialogDescription>
          A cycle groups tasks into a time-boxed window, like a sprint.
        </DialogDescription>
      </DialogHeader>
      <form class="space-y-4" @submit.prevent="handleSubmit">
        <div
          v-if="error"
          class="rounded-md bg-destructive/10 p-3 text-sm text-destructive"
        >
          {{ error }}
        </div>

        <div class="space-y-2">
          <Label for="cycle_title">Title</Label>
          <Input
            id="cycle_title"
            v-model="form.title"
            placeholder="Q2 Sprint 3"
            required
            :disabled="loading"
          />
        </div>

        <div class="space-y-2">
          <Label for="cycle_desc">Description</Label>
          <Textarea
            id="cycle_desc"
            v-model="form.description"
            placeholder="What's this cycle about?"
            rows="3"
            :disabled="loading"
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div class="space-y-2">
            <Label>Start date</Label>
            <Popover v-model:open="startOpen">
              <PopoverTrigger as-child>
                <Button
                  type="button"
                  variant="outline"
                  class="w-full justify-start gap-2 font-normal"
                  :class="!form.startDate && 'text-muted-foreground'"
                  :disabled="loading"
                >
                  <CalendarIcon class="size-4" />
                  {{ formatDateValue(form.startDate) }}
                </Button>
              </PopoverTrigger>
              <PopoverContent class="w-auto p-0" align="start">
                <Calendar v-model="form.startDate" layout="month-and-year" />
              </PopoverContent>
            </Popover>
          </div>
          <div class="space-y-2">
            <Label>End date</Label>
            <Popover v-model:open="endOpen">
              <PopoverTrigger as-child>
                <Button
                  type="button"
                  variant="outline"
                  class="w-full justify-start gap-2 font-normal"
                  :class="!form.endDate && 'text-muted-foreground'"
                  :disabled="loading"
                >
                  <CalendarIcon class="size-4" />
                  {{ formatDateValue(form.endDate) }}
                </Button>
              </PopoverTrigger>
              <PopoverContent class="w-auto p-0" align="start">
                <Calendar
                  v-model="form.endDate"
                  layout="month-and-year"
                  :min-value="form.startDate"
                />
              </PopoverContent>
            </Popover>
          </div>
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            :disabled="loading"
            @click="open = false"
          >
            Cancel
          </Button>
          <Button type="submit" :disabled="loading || !canSubmit">
            <Loader2 v-if="loading" class="mr-2 size-4 animate-spin" />
            {{ isEdit ? "Save changes" : "Create Cycle" }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { Loader2 } from "lucide-vue-next";
import type {
  CreateTemplateRequest,
  EditableTemplateField,
  ProjectLabel,
  ProjectMember,
  TaskTemplate,
  TemplateField,
} from "~/types";
import { mdToHtml } from "~/utils/markdown";

const props = defineProps<{
  template?: TaskTemplate;
  labels: ProjectLabel[];
  members: ProjectMember[];
  loading: boolean;
  submitLabel: string;
}>();

const emit = defineEmits<{
  submit: [data: CreateTemplateRequest];
  cancel: [];
}>();

const NO_PRIORITY = "none";
const priorities = [
  { value: "0", label: "No priority" },
  { value: "1", label: "Low" },
  { value: "2", label: "Medium" },
  { value: "3", label: "High" },
  { value: "4", label: "Urgent" },
];

function toEditable(f: TemplateField, i: number): EditableTemplateField {
  return {
    uid: `${f.key}-${i}`,
    key: f.key,
    label: f.label,
    type: f.type,
    description: f.description ?? "",
    required: f.required,
    optionsText: (f.options ?? []).join("\n"),
    multiple: !!f.multiple,
    min: f.min != null ? String(f.min) : "",
    max: f.max != null ? String(f.max) : "",
  };
}

function slugify(label: string) {
  return (
    label
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "_")
      .replace(/^_+|_+$/g, "")
      .slice(0, 40) || "field"
  );
}

function toNumber(v: string | number) {
  const s = String(v).trim();
  return s === "" || !Number.isFinite(Number(s)) ? undefined : Number(s);
}

function fromEditable(list: EditableTemplateField[]): TemplateField[] {
  const used = new Set<string>();
  return list.map((f) => {
    const base = f.key || slugify(f.label);
    let key = base;
    for (let n = 2; used.has(key); n++) key = `${base}_${n}`;
    used.add(key);
    return {
      key,
      label: f.label.trim(),
      type: f.type,
      description: f.description.trim() || undefined,
      required: f.required,
      options:
        f.type === "select"
          ? f.optionsText.split("\n").map((o) => o.trim()).filter(Boolean)
          : undefined,
      multiple: f.type === "select" ? f.multiple : undefined,
      min: f.type === "number" ? toNumber(f.min) : undefined,
      max: f.type === "number" ? toNumber(f.max) : undefined,
    };
  });
}

const t = props.template;
const form = ref({
  name: t?.name ?? "",
  title: t?.title ?? "",
  // Legacy templates are stored as markdown; convert so the editor shows them formatted.
  description: mdToHtml(t?.description ?? ""),
  fields: (t?.fields ?? []).map(toEditable),
  priority: t?.default_priority != null ? String(t.default_priority) : NO_PRIORITY,
  labels: [...(t?.default_labels ?? [])],
  assignees: [...(t?.default_assignees ?? [])],
});

const fieldsInvalid = computed(() =>
  form.value.fields.some(
    (f) => !f.label.trim() || (f.type === "select" && !f.optionsText.trim())
  )
);

const selectedLabels = computed(() => props.labels.filter((l) => form.value.labels.includes(l.id)));
const availableLabels = computed(() => props.labels.filter((l) => !form.value.labels.includes(l.id)));
const selectedMembers = computed(() =>
  props.members.filter((m) => form.value.assignees.includes(m.user_id))
);
const availableMembers = computed(() =>
  props.members.filter((m) => !form.value.assignees.includes(m.user_id))
);

function handleSubmit() {
  if (!form.value.name.trim() || fieldsInvalid.value) return;
  const labelIds = new Set(props.labels.map((l) => l.id));
  const memberIds = new Set(props.members.map((m) => m.user_id));
  emit("submit", {
    name: form.value.name.trim(),
    title: form.value.title,
    description: form.value.description,
    fields: fromEditable(form.value.fields),
    default_priority: form.value.priority === NO_PRIORITY ? null : Number(form.value.priority),
    default_labels: form.value.labels.filter((id) => labelIds.has(id)),
    default_assignees: form.value.assignees.filter((id) => memberIds.has(id)),
  });
}
</script>

<template>
  <form class="space-y-5" @submit.prevent="handleSubmit">
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <div class="space-y-2">
        <Label>Name</Label>
        <Input v-model="form.name" placeholder="e.g. Provision EC2 instance" :disabled="loading" />
      </div>
      <div class="space-y-2">
        <Label>Default title</Label>
        <Input v-model="form.title" placeholder="Pre-filled task title" :disabled="loading" />
      </div>
    </div>

    <div class="space-y-2">
      <Label>Description</Label>
      <p class="text-xs text-muted-foreground">
        With fields, this is shown as guidance above them. Without fields, it pre-fills the task description.
      </p>
      <TiptapEditor v-model="form.description" :disabled="loading" compact />
    </div>

    <div class="space-y-2">
      <Label>Fields</Label>
      <p class="text-xs text-muted-foreground">
        People must fill required fields to create a task. Answers are added to the task description.
      </p>
      <TemplateFieldsEditor v-model="form.fields" :disabled="loading" />
    </div>

    <div class="space-y-3">
      <Label>Defaults</Label>
      <div class="grid grid-cols-1 items-start gap-4 sm:grid-cols-3">
        <div class="min-w-0 space-y-2">
          <p class="text-xs text-muted-foreground">Priority</p>
          <Select v-model="form.priority" :disabled="loading">
            <SelectTrigger class="w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="NO_PRIORITY">No default priority</SelectItem>
              <SelectItem v-for="p in priorities" :key="p.value" :value="p.value">
                {{ p.label }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div class="min-w-0 space-y-2">
          <p class="text-xs text-muted-foreground">Labels</p>
          <TokenSelect
            :selected="selectedLabels"
            :available="availableLabels"
            :get-key="(l) => l.id"
            :get-search-text="(l) => l.name"
            :chip-style="(l) => ({ backgroundColor: l.color + '20', color: l.color })"
            :chip-class="() => 'pl-2 pr-1 font-medium'"
            :disabled="loading"
            placeholder="Add labels..."
            empty-text="No labels found"
            @add="(l) => form.labels.push(l.id)"
            @remove="(l) => (form.labels = form.labels.filter((id) => id !== l.id))"
          >
            <template #chip="{ item: label }">
              <span class="truncate">{{ label.name }}</span>
            </template>
            <template #option="{ item: label }">
              <div class="size-3 shrink-0 rounded-full" :style="{ backgroundColor: label.color }" />
              {{ label.name }}
            </template>
          </TokenSelect>
        </div>
        <div class="min-w-0 space-y-2">
          <p class="text-xs text-muted-foreground">Assignees</p>
          <TokenSelect
            :selected="selectedMembers"
            :available="availableMembers"
            :get-key="(m) => m.user_id"
            :get-search-text="(m) => `${m.first_name} ${m.last_name} ${m.username}`"
            :disabled="loading"
            placeholder="Add assignees..."
            empty-text="No members found"
            @add="(m) => form.assignees.push(m.user_id)"
            @remove="(m) => (form.assignees = form.assignees.filter((id) => id !== m.user_id))"
          >
            <template #chip="{ item: member }">
              <Avatar class="size-5">
                <AvatarImage v-if="member.avatar_url" :src="member.avatar_url" />
                <AvatarFallback class="text-[10px]" :seed="member.user_id">
                  {{ member.first_name[0] }}{{ member.last_name[0] }}
                </AvatarFallback>
              </Avatar>
              <span class="truncate">{{ member.first_name }} {{ member.last_name }}</span>
            </template>
            <template #option="{ item: member }">
              <Avatar class="size-6">
                <AvatarImage v-if="member.avatar_url" :src="member.avatar_url" />
                <AvatarFallback class="text-xs" :seed="member.user_id">
                  {{ member.first_name[0] }}{{ member.last_name[0] }}
                </AvatarFallback>
              </Avatar>
              {{ member.first_name }} {{ member.last_name }}
            </template>
          </TokenSelect>
        </div>
      </div>
    </div>

    <div class="flex justify-end gap-2">
      <Button type="button" variant="outline" size="sm" :disabled="loading" @click="emit('cancel')">
        Cancel
      </Button>
      <Button type="submit" size="sm" :disabled="loading || !form.name.trim() || fieldsInvalid">
        <Loader2 v-if="loading" class="mr-1.5 size-4 animate-spin" />
        {{ submitLabel }}
      </Button>
    </div>
  </form>
</template>

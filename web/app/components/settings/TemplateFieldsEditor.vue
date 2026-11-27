<script setup lang="ts">
import { ArrowDown, ArrowUp, Plus, Trash2 } from "lucide-vue-next";
import type { EditableTemplateField, TemplateFieldType } from "~/types";

defineProps<{ disabled?: boolean }>();

const fields = defineModel<EditableTemplateField[]>({ required: true });

const fieldTypes: { value: TemplateFieldType; label: string }[] = [
  { value: "text", label: "Short text" },
  { value: "textarea", label: "Long text" },
  { value: "number", label: "Number" },
  { value: "select", label: "Select" },
];

let seq = 0;

function addField() {
  fields.value = [
    ...fields.value,
    {
      uid: `new-${Date.now()}-${seq++}`,
      key: "",
      label: "",
      type: "text",
      description: "",
      required: true,
      optionsText: "",
      multiple: false,
      min: "",
      max: "",
    },
  ];
}

function removeField(i: number) {
  fields.value = fields.value.filter((_, idx) => idx !== i);
}

function move(i: number, dir: -1 | 1) {
  const j = i + dir;
  if (j < 0 || j >= fields.value.length) return;
  const next = [...fields.value];
  [next[i], next[j]] = [next[j]!, next[i]!];
  fields.value = next;
}
</script>

<template>
  <div class="space-y-3">
    <div
      v-for="(f, i) in fields"
      :key="f.uid"
      class="space-y-3 rounded-md border p-3"
    >
      <div class="flex flex-col gap-2 sm:flex-row sm:items-center">
        <Input
          v-model="f.label"
          placeholder="Field label, e.g. RAM (GB)"
          class="h-8 flex-1"
          :disabled="disabled"
        />
        <div class="flex items-center gap-2">
          <Select v-model="f.type" :disabled="disabled">
            <SelectTrigger class="h-8 w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="t in fieldTypes" :key="t.value" :value="t.value">
                {{ t.label }}
              </SelectItem>
            </SelectContent>
          </Select>
          <label class="flex items-center gap-1.5 text-xs text-muted-foreground">
            <Switch v-model:checked="f.required" :disabled="disabled" />
            Required
          </label>
          <div class="ml-auto flex items-center sm:ml-0">
            <Button
              type="button"
              variant="ghost"
              size="icon"
              class="size-7"
              aria-label="Move up"
              :disabled="disabled || i === 0"
              @click="move(i, -1)"
            >
              <ArrowUp class="size-3.5" />
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              class="size-7"
              aria-label="Move down"
              :disabled="disabled || i === fields.length - 1"
              @click="move(i, 1)"
            >
              <ArrowDown class="size-3.5" />
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              class="size-7 text-destructive hover:text-destructive"
              aria-label="Remove field"
              :disabled="disabled"
              @click="removeField(i)"
            >
              <Trash2 class="size-3.5" />
            </Button>
          </div>
        </div>
      </div>

      <Input
        v-model="f.description"
        placeholder="Help text (optional)"
        class="h-8"
        :disabled="disabled"
      />

      <div v-if="f.type === 'select'" class="space-y-2">
        <Textarea
          v-model="f.optionsText"
          placeholder="One option per line"
          rows="3"
          :disabled="disabled"
        />
        <label class="flex items-center gap-2 text-xs text-muted-foreground">
          <Checkbox v-model="f.multiple" :disabled="disabled" />
          Allow multiple
        </label>
      </div>

      <div v-if="f.type === 'number'" class="grid grid-cols-2 gap-2">
        <Input
          v-model="f.min"
          type="number"
          placeholder="Min (optional)"
          class="h-8"
          :disabled="disabled"
        />
        <Input
          v-model="f.max"
          type="number"
          placeholder="Max (optional)"
          class="h-8"
          :disabled="disabled"
        />
      </div>
    </div>

    <Button type="button" variant="outline" size="sm" :disabled="disabled" @click="addField">
      <Plus class="mr-1.5 size-4" />
      Add field
    </Button>
  </div>
</template>

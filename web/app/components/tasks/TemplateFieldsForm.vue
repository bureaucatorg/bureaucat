<script setup lang="ts">
import type { TemplateField, TemplateFieldValue } from "~/types";
import { renderRichText } from "~/utils/markdown";

const props = defineProps<{
  fields: TemplateField[];
  errors: Record<string, string>;
  guidance?: string;
  disabled?: boolean;
}>();

const values = defineModel<Record<string, TemplateFieldValue>>({ required: true });

const guidanceHtml = computed(() => renderRichText(props.guidance));

function fieldId(f: TemplateField) {
  return `tf-${f.key}`;
}

function stringValue(key: string) {
  const v = values.value[key];
  return v === undefined ? "" : String(v);
}

function setValue(key: string, v: TemplateFieldValue) {
  values.value = { ...values.value, [key]: v };
}

function selectedOptions(key: string): string[] {
  const v = values.value[key];
  return Array.isArray(v) ? v : [];
}

function addOption(key: string, opt: string) {
  setValue(key, [...selectedOptions(key), opt]);
}

function removeOption(key: string, opt: string) {
  setValue(key, selectedOptions(key).filter((o) => o !== opt));
}

function availableOptions(f: TemplateField) {
  const picked = new Set(selectedOptions(f.key));
  return (f.options ?? []).filter((o) => !picked.has(o));
}
</script>

<template>
  <div class="space-y-4 rounded-lg border bg-muted/20 p-4">
    <div
      v-if="guidanceHtml"
      class="prose prose-sm max-w-none text-muted-foreground dark:prose-invert"
      v-html="guidanceHtml"
    />

    <div v-for="f in fields" :key="f.key" class="space-y-1.5">
      <Label :for="fieldId(f)">
        {{ f.label }}
        <span v-if="f.required" class="text-destructive" aria-hidden="true">*</span>
      </Label>
      <p v-if="f.description" class="text-xs text-muted-foreground">{{ f.description }}</p>

      <Input
        v-if="f.type === 'text'"
        :id="fieldId(f)"
        :model-value="stringValue(f.key)"
        :disabled="disabled"
        :aria-invalid="!!errors[f.key]"
        @update:model-value="(v) => setValue(f.key, String(v))"
      />

      <Textarea
        v-else-if="f.type === 'textarea'"
        :id="fieldId(f)"
        :model-value="stringValue(f.key)"
        :disabled="disabled"
        :aria-invalid="!!errors[f.key]"
        rows="3"
        @update:model-value="(v) => setValue(f.key, String(v))"
      />

      <Input
        v-else-if="f.type === 'number'"
        :id="fieldId(f)"
        type="number"
        inputmode="decimal"
        :min="f.min"
        :max="f.max"
        :model-value="stringValue(f.key)"
        :disabled="disabled"
        :aria-invalid="!!errors[f.key]"
        @update:model-value="(v) => setValue(f.key, String(v))"
      />

      <TokenSelect
        v-else-if="f.type === 'select' && f.multiple"
        :selected="selectedOptions(f.key)"
        :available="availableOptions(f)"
        :get-key="(o: string) => o"
        :get-search-text="(o: string) => o"
        :disabled="disabled"
        placeholder="Select options..."
        empty-text="No options left"
        @add="(o: string) => addOption(f.key, o)"
        @remove="(o: string) => removeOption(f.key, o)"
      >
        <template #chip="{ item }">
          <span class="truncate">{{ item }}</span>
        </template>
        <template #option="{ item }">
          {{ item }}
        </template>
      </TokenSelect>

      <Select
        v-else-if="f.type === 'select'"
        :model-value="stringValue(f.key) || undefined"
        :disabled="disabled"
        @update:model-value="(v) => setValue(f.key, String(v ?? ''))"
      >
        <SelectTrigger :id="fieldId(f)" class="w-full" :aria-invalid="!!errors[f.key]">
          <SelectValue placeholder="Select an option" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="o in f.options" :key="o" :value="o">{{ o }}</SelectItem>
        </SelectContent>
      </Select>

      <p v-if="errors[f.key]" class="text-xs text-destructive">{{ errors[f.key] }}</p>
    </div>
  </div>
</template>

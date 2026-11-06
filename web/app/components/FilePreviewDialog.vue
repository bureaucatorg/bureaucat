<script setup lang="ts">
import { Download, ExternalLink, Loader2 } from "lucide-vue-next";
import type { Attachment } from "~/composables/useAttachments";
import { previewKind } from "~/utils/filePreview";
import { parseCsv } from "~/utils/csv";
import { renderMarkdown } from "~/utils/markdown";

const MAX_TEXT_BYTES = 2 * 1024 * 1024;
const MAX_CSV_ROWS = 1000;

const props = defineProps<{ attachment: Attachment | null }>();
const emit = defineEmits<{ close: [] }>();

const kind = computed(() => (props.attachment ? previewKind(props.attachment) : null));
const loading = ref(false);
const error = ref("");
const text = ref("");

const csvRows = computed(() => (kind.value === "csv" ? parseCsv(text.value) : []));
const csvHeader = computed(() => csvRows.value[0] ?? []);
const csvBody = computed(() => csvRows.value.slice(1, MAX_CSV_ROWS + 1));
const csvTotal = computed(() => Math.max(csvRows.value.length - 1, 0));
const csvTruncated = computed(() => csvTotal.value > MAX_CSV_ROWS);
const csvColumns = computed(() =>
  Array.from({ length: Math.max(0, ...csvRows.value.map((r) => r.length)) }, (_, i) => i)
);
const csvNumeric = computed(() =>
  csvColumns.value.map((c) => {
    const values = csvBody.value.map((r) => r[c]?.trim()).filter(Boolean);
    return values.length > 0 && values.every((v) => /^-?[\d,]*\.?\d+%?$/.test(v!));
  })
);
const markdownHtml = computed(() => (kind.value === "markdown" ? renderMarkdown(text.value) : ""));

watch(
  () => props.attachment,
  async (a) => {
    text.value = "";
    error.value = "";
    if (!a || kind.value === "pdf") return;
    if (a.size_bytes > MAX_TEXT_BYTES) {
      error.value = "This file is too large to preview. Download it instead.";
      return;
    }
    loading.value = true;
    try {
      const res = await fetch(a.url);
      if (!res.ok) throw new Error();
      text.value = await res.text();
    } catch {
      error.value = "Could not load the file.";
    } finally {
      loading.value = false;
    }
  },
  { immediate: true }
);
</script>

<template>
  <Dialog :open="!!attachment" @update:open="(v) => !v && emit('close')">
    <DialogContent
      :show-close-button="false"
      class="flex max-h-[90vh] max-w-[calc(100%-1rem)] flex-col gap-0 overflow-hidden p-0 sm:max-w-5xl"
      :class="kind === 'pdf' ? 'h-[90vh]' : 'min-h-48'"
    >
      <div class="flex items-center gap-2 border-b px-4 py-2.5">
        <div class="min-w-0 flex-1">
          <DialogTitle class="truncate text-sm font-medium">
            {{ attachment?.filename }}
          </DialogTitle>
          <p v-if="kind === 'csv' && text" class="text-xs text-muted-foreground">
            {{ csvTotal.toLocaleString() }} rows · {{ csvColumns.length }} columns
          </p>
        </div>
        <DialogDescription class="sr-only">File preview</DialogDescription>
        <Button v-if="kind === 'pdf'" variant="ghost" size="icon" class="size-8" as-child>
          <a :href="attachment?.url" target="_blank" rel="noopener" aria-label="Open in new tab">
            <ExternalLink class="size-4" />
          </a>
        </Button>
        <Button variant="ghost" size="icon" class="size-8" as-child>
          <a :href="attachment?.url" :download="attachment?.filename" aria-label="Download">
            <Download class="size-4" />
          </a>
        </Button>
        <DialogClose as-child>
          <Button variant="ghost" size="sm" class="h-8">Close</Button>
        </DialogClose>
      </div>

      <div class="min-h-0 flex-1 overflow-auto">
        <iframe
          v-if="kind === 'pdf'"
          :src="attachment?.url"
          :title="attachment?.filename"
          class="size-full border-0"
        />
        <div v-else-if="loading" class="flex items-center justify-center py-16">
          <Loader2 class="size-5 animate-spin text-muted-foreground" />
        </div>
        <p v-else-if="error" class="p-6 text-center text-sm text-muted-foreground">{{ error }}</p>
        <div
          v-else-if="kind === 'markdown'"
          class="prose prose-sm max-w-none p-4 dark:prose-invert sm:p-6"
          v-html="markdownHtml"
        />
        <pre
          v-else-if="kind === 'text'"
          class="whitespace-pre-wrap break-words p-4 font-mono text-xs leading-relaxed sm:p-6"
        >{{ text }}</pre>
        <div v-else-if="kind === 'csv'" class="[&_[data-slot=table-container]]:overflow-visible">
          <Table class="text-xs">
            <TableHeader>
              <TableRow class="hover:bg-transparent">
                <TableHead class="sticky top-0 z-10 h-8 w-px bg-muted px-3 text-right font-normal text-muted-foreground shadow-[inset_0_-1px_0_var(--border)]">
                  #
                </TableHead>
                <TableHead
                  v-for="c in csvColumns"
                  :key="c"
                  class="sticky top-0 z-10 h-8 border-l bg-muted px-3 font-semibold shadow-[inset_0_-1px_0_var(--border)]"
                  :class="csvNumeric[c] && 'text-right'"
                >
                  {{ csvHeader[c] }}
                </TableHead>
                <TableHead class="sticky top-0 z-10 h-8 w-full border-l bg-muted shadow-[inset_0_-1px_0_var(--border)]" />
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="(row, r) in csvBody" :key="r" class="even:bg-muted/30">
                <TableCell class="w-px px-3 py-1.5 text-right tabular-nums text-muted-foreground">
                  {{ r + 1 }}
                </TableCell>
                <TableCell
                  v-for="c in csvColumns"
                  :key="c"
                  class="max-w-80 truncate border-l px-3 py-1.5"
                  :class="csvNumeric[c] && 'text-right tabular-nums'"
                  :title="row[c]"
                >
                  {{ row[c] }}
                </TableCell>
                <TableCell class="w-full border-l p-0" />
              </TableRow>
            </TableBody>
          </Table>
          <p v-if="csvTruncated" class="border-t p-3 text-center text-xs text-muted-foreground">
            Showing the first {{ MAX_CSV_ROWS.toLocaleString() }} rows. Download the file to see the rest.
          </p>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

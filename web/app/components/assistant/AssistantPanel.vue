<script setup lang="ts">
import { Bot, Check, ChevronsUpDown, ChevronRight, Loader2, SendHorizontal, MessageSquarePlus, MessagesSquare, Trash2, Pencil, X, CircleAlert, Paperclip, FileText, MapPin } from "lucide-vue-next";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "~/components/ui/collapsible";
import { toast } from "vue-sonner";
import type { AssistantAction, AssistantAttachment } from "~/composables/useAssistant";
import { pageContextFromRoute, describePageContext } from "~/composables/useAssistant";
import { renderAssistantMarkdown } from "~/utils/markdown";

const { user } = useAuth();
const {
  enabled,
  open,
  conversations,
  current,
  messages,
  actions,
  steps,
  busy,
  error,
  setOpen,
  listConversations,
  loadConversation,
  newConversation,
  deleteConversation,
  renameConversation,
  fetchToolOutput,
  send,
  resolve,
} = useAssistant();

const { uploadFile } = useUploads();
const route = useRoute();

const input = ref("");
const attachments = ref<AssistantAttachment[]>([]);
const uploading = ref(false);
const fileInput = ref<HTMLInputElement | null>(null);
const MAX_ATTACHMENTS = 10;
const pickerOpen = ref(false);
const confirmDelete = ref(false);
const editingTitle = ref(false);
const titleDraft = ref("");
const titleInput = ref<{ $el: HTMLInputElement } | null>(null);
const scroller = ref<HTMLElement | null>(null);

const actionsByCall = computed(() => new Map(actions.value.map((a) => [a.tool_call_id, a])));

const suggestions = [
  "What's assigned to me?",
  "Summarise overdue tasks in my projects",
  "What changed on my tasks this week?",
];

function toolLabel(name: string) {
  return name.replace(/_/g, " ");
}

const toolOutputs = ref<Record<string, { loading: boolean; text?: string; failed?: boolean }>>({});

async function loadToolOutput(callId: string, open: boolean) {
  const cached = toolOutputs.value[callId];
  if (!open || !current.value || (cached && !cached.failed)) return;
  toolOutputs.value[callId] = { loading: true };
  const text = await fetchToolOutput(current.value.id, callId);
  toolOutputs.value[callId] = text === null ? { loading: false, failed: true } : { loading: false, text };
}

watch(() => current.value?.id, () => (toolOutputs.value = {}));

function formatOutput(output: string) {
  try {
    return JSON.stringify(JSON.parse(output), null, 2);
  } catch {
    return output;
  }
}

function formatArg(value: unknown) {
  const text = typeof value === "string" ? value : JSON.stringify(value);
  return text.length > 160 ? `${text.slice(0, 160)}…` : text;
}

const canSend = computed(() => !busy.value && !uploading.value && (!!input.value.trim() || attachments.value.length > 0));

async function submit() {
  if (!canSend.value) return;
  const text = input.value.trim();
  const files = attachments.value;
  input.value = "";
  attachments.value = [];
  const page = pageContextFromRoute(route);
  await send(text, files, page ? { page } : {});
}

async function onFiles(e: Event) {
  const target = e.target as HTMLInputElement;
  const files = Array.from(target.files ?? []);
  target.value = "";
  if (attachments.value.length + files.length > MAX_ATTACHMENTS) {
    toast.error(`You can attach up to ${MAX_ATTACHMENTS} files per message`);
    return;
  }
  uploading.value = true;
  for (const file of files) {
    const result = await uploadFile(file);
    if (result.success && result.data) {
      attachments.value.push({
        upload_id: result.data.id,
        filename: result.data.filename,
        mime_type: result.data.mime_type,
        size_bytes: result.data.size_bytes,
      });
    } else {
      toast.error(`${file.name}: ${result.error || "upload failed"}`);
    }
  }
  uploading.value = false;
}

// Keep in-app links (e.g. task links) in the same tab; external links still open a new tab.
function onLinkClick(e: MouseEvent) {
  const link = (e.target as HTMLElement).closest("a");
  const href = link?.getAttribute("href");
  if (!href?.startsWith("/") || href.startsWith("//") || e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
  e.preventDefault();
  navigateTo(href);
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Enter" && !e.shiftKey && !e.isComposing) {
    e.preventDefault();
    submit();
  }
}

async function pick(id: string) {
  pickerOpen.value = false;
  confirmDelete.value = false;
  await loadConversation(id);
}

function startNew() {
  confirmDelete.value = false;
  newConversation();
}

function startRename() {
  if (!current.value) return;
  titleDraft.value = current.value.title;
  editingTitle.value = true;
  nextTick(() => titleInput.value?.$el.select());
}

async function saveTitle() {
  if (!editingTitle.value || !current.value) return;
  editingTitle.value = false;
  const title = titleDraft.value.trim();
  if (!title || title === current.value.title) return;
  if (!(await renameConversation(current.value.id, title))) toast.error("Could not rename chat");
}

async function removeCurrent() {
  if (!current.value) return;
  if (!confirmDelete.value) {
    confirmDelete.value = true;
    return;
  }
  confirmDelete.value = false;
  await deleteConversation(current.value.id);
}

function statusClass(a: AssistantAction) {
  return {
    pending: "border-amber-500/40 bg-amber-500/5",
    approved: "border-emerald-500/30",
    rejected: "opacity-60",
    failed: "border-destructive/40",
  }[a.status];
}

watch(
  () => [messages.value.length, actions.value.length, steps.value.length, busy.value],
  async () => {
    await nextTick();
    scroller.value?.scrollTo({ top: scroller.value.scrollHeight });
  }
);

watch(
  open,
  (isOpen) => {
    if (isOpen && enabled.value) listConversations();
  },
  { immediate: true }
);
</script>

<template>
  <aside
    v-if="open && !enabled && user?.user_type === 'admin'"
    class="fixed inset-y-0 left-12 z-40 flex w-[calc(100vw-3rem)] flex-col border-r bg-background md:w-[28rem] lg:w-[36rem]"
    aria-label="Assistant"
  >
    <header class="flex h-12 shrink-0 items-center justify-between border-b px-4">
      <p class="font-medium">Assistant</p>
      <Button variant="ghost" size="icon" title="Close" @click="setOpen(false)">
        <X class="size-4" />
      </Button>
    </header>
    <div class="flex flex-1 flex-col items-center justify-center gap-4 px-6 text-center">
      <Bot class="size-8 text-muted-foreground" />
      <p class="text-sm text-muted-foreground">
        The assistant isn't enabled yet. Choose a provider, model and API key in the admin settings.
      </p>
      <Button as-child size="sm">
        <NuxtLink to="/admin" @click="setOpen(false)">Open admin settings</NuxtLink>
      </Button>
    </div>
  </aside>
  <aside
    v-else-if="open && enabled"
    class="fixed inset-y-0 left-12 z-40 flex w-[calc(100vw-3rem)] flex-col border-r bg-background md:w-[28rem] lg:w-[36rem]"
    aria-label="Assistant"
  >
    <header class="flex h-12 shrink-0 items-center gap-1 border-b px-2">
      <Input
        v-if="editingTitle"
        ref="titleInput"
        v-model="titleDraft"
        maxlength="100"
        aria-label="Chat title"
        class="h-8 min-w-0 flex-1"
        @keydown.enter.prevent="saveTitle"
        @keydown.esc.prevent="editingTitle = false"
        @blur="saveTitle"
      />
      <Popover v-else v-model:open="pickerOpen">
        <PopoverTrigger as-child>
          <Button variant="ghost" role="combobox" class="min-w-0 flex-1 justify-between px-2 font-normal">
            <span class="truncate">{{ current?.title || "New chat" }}</span>
            <ChevronsUpDown class="size-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent align="start" class="w-80 p-0">
          <Command>
            <CommandInput placeholder="Search chats..." />
            <CommandList>
              <CommandEmpty>No chats yet.</CommandEmpty>
              <CommandGroup>
                <CommandItem
                  v-for="c in conversations"
                  :key="c.id"
                  :value="`${c.title} ${c.id}`"
                  @select="pick(c.id)"
                >
                  <Check :class="['size-4', current?.id === c.id ? 'opacity-100' : 'opacity-0']" />
                  <span class="truncate">{{ c.title || "Untitled" }}</span>
                </CommandItem>
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
      <Button
        v-if="current && !editingTitle"
        variant="ghost"
        size="icon"
        title="Rename chat"
        :disabled="busy"
        @click="startRename"
      >
        <Pencil class="size-4" />
      </Button>
      <Button
        v-if="current"
        variant="ghost"
        size="icon"
        :title="confirmDelete ? 'Click again to delete' : 'Delete chat'"
        :class="confirmDelete && 'text-destructive'"
        :disabled="busy"
        @click="removeCurrent"
      >
        <Trash2 class="size-4" />
      </Button>
      <Button variant="ghost" size="icon" title="New chat" :disabled="busy" @click="startNew">
        <MessageSquarePlus class="size-4" />
      </Button>
      <Button variant="ghost" size="icon" title="Close" @click="setOpen(false)">
        <X class="size-4" />
      </Button>
    </header>

    <div ref="scroller" class="flex-1 overflow-y-auto px-4 py-4" @click="onLinkClick">
      <div v-if="!messages.length" class="flex h-full flex-col items-center justify-center gap-4 text-center">
        <Bot class="size-8 text-muted-foreground" />
        <p class="text-sm text-muted-foreground">
          Ask about your tasks and projects. Any change needs your approval first.
        </p>
        <div class="flex flex-col gap-2">
          <Button
            v-for="s in suggestions"
            :key="s"
            variant="outline"
            size="sm"
            class="h-auto whitespace-normal py-1.5 font-normal"
            @click="input = s"
          >
            {{ s }}
          </Button>
        </div>
      </div>

      <div v-else class="flex flex-col gap-3">
        <template v-for="m in messages" :key="m.id">
          <div v-if="m.role === 'user'" class="ml-8 flex flex-col items-end gap-1 self-end">
            <Collapsible v-if="m.meta?.chat?.messages?.length" class="group w-full">
              <CollapsibleTrigger
                class="ml-auto flex items-center gap-1 text-[11px] text-muted-foreground hover:text-foreground"
                title="Recent Mattermost messages shared with the assistant"
              >
                <MessagesSquare class="size-3" />
                {{ m.meta.chat.messages.length }} {{ m.meta.chat.where }} message{{ m.meta.chat.messages.length === 1 ? "" : "s" }} as context
                <ChevronRight class="size-3 transition-transform group-data-[state=open]:rotate-90" />
              </CollapsibleTrigger>
              <CollapsibleContent>
                <div class="mt-1 mb-1 flex flex-col gap-2 rounded-md border px-2.5 py-2 text-xs">
                  <div v-for="(c, i) in m.meta.chat.messages" :key="i" class="min-w-0">
                    <div class="flex flex-wrap items-baseline gap-x-1">
                      <span class="font-medium">{{ c.author }}</span>
                      <span v-if="c.is_assistant" class="text-muted-foreground">(assistant)</span>
                      <span v-else-if="c.bureaucat_user" class="text-muted-foreground">({{ c.bureaucat_user }})</span>
                    </div>
                    <p class="whitespace-pre-wrap break-words text-muted-foreground">{{ c.text }}</p>
                  </div>
                </div>
              </CollapsibleContent>
            </Collapsible>
            <div v-if="m.content" class="whitespace-pre-wrap break-words rounded-lg bg-muted px-3 py-2 text-sm">
              {{ m.content }}
            </div>
            <a
              v-for="a in m.attachments ?? []"
              :key="a.upload_id"
              :href="`/api/v1/uploads/${a.upload_id}`"
              target="_blank"
              rel="noopener noreferrer"
              class="flex max-w-full items-center gap-1.5 rounded-md border px-2 py-1 text-xs text-muted-foreground hover:text-foreground"
            >
              <FileText class="size-3.5 shrink-0" />
              <span class="truncate">{{ a.filename }}</span>
            </a>
            <span
              v-if="m.meta?.page"
              class="flex items-center gap-1 text-[11px] text-muted-foreground"
              title="The page you were on, shared with the assistant"
            >
              <MapPin class="size-3" />
              on {{ describePageContext(m.meta.page) }}
            </span>
          </div>
          <template v-else>
            <div v-if="m.content" class="prose max-w-none break-words text-sm" v-html="renderAssistantMarkdown(m.content)" />
            <template v-for="tc in m.tool_calls" :key="tc.id">
              <div
                v-if="actionsByCall.get(tc.id)"
                class="rounded-md border p-3 text-sm"
                :class="statusClass(actionsByCall.get(tc.id)!)"
              >
                <div class="flex items-center justify-between gap-2">
                  <p class="font-medium">{{ actionsByCall.get(tc.id)!.title }}</p>
                  <Badge variant="outline" class="capitalize">{{ actionsByCall.get(tc.id)!.status }}</Badge>
                </div>
                <dl class="mt-2 space-y-1 text-xs">
                  <div v-for="(v, k) in actionsByCall.get(tc.id)!.args" :key="k" class="flex gap-2">
                    <dt class="shrink-0 text-muted-foreground">{{ toolLabel(String(k)) }}</dt>
                    <dd class="min-w-0 whitespace-pre-wrap break-words">{{ formatArg(v) }}</dd>
                  </div>
                </dl>
                <p v-if="actionsByCall.get(tc.id)!.result" class="mt-2 text-xs text-destructive">
                  {{ actionsByCall.get(tc.id)!.result }}
                </p>
                <div v-if="actionsByCall.get(tc.id)!.status === 'pending'" class="mt-3 flex justify-end gap-2">
                  <Button size="sm" variant="outline" :disabled="busy" @click="resolve(actionsByCall.get(tc.id)!, false)">
                    Reject
                  </Button>
                  <Button size="sm" :disabled="busy" @click="resolve(actionsByCall.get(tc.id)!, true)">
                    Approve
                  </Button>
                </div>
              </div>
              <Collapsible v-else class="group" @update:open="loadToolOutput(tc.id, $event)">
                <CollapsibleTrigger class="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground">
                  <CircleAlert v-if="tc.is_error" class="size-3 text-destructive" />
                  <Check v-else class="size-3" />
                  {{ toolLabel(tc.name) }}
                  <ChevronRight class="size-3 transition-transform group-data-[state=open]:rotate-90" />
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <div class="mt-1 space-y-2 rounded-md border px-2.5 py-2 text-xs">
                    <dl v-if="Object.keys(tc.args ?? {}).length" class="space-y-1">
                      <div v-for="(v, k) in tc.args" :key="k" class="flex gap-2">
                        <dt class="shrink-0 text-muted-foreground">{{ toolLabel(String(k)) }}</dt>
                        <dd class="min-w-0 whitespace-pre-wrap break-words">{{ formatArg(v) }}</dd>
                      </div>
                    </dl>
                    <p v-if="!tc.has_output" class="text-muted-foreground">No output recorded.</p>
                    <p v-else-if="toolOutputs[tc.id]?.loading" class="flex items-center gap-1.5 text-muted-foreground">
                      <Loader2 class="size-3 animate-spin" /> Loading output…
                    </p>
                    <p v-else-if="toolOutputs[tc.id]?.failed" class="text-destructive">Could not load output.</p>
                    <pre
                      v-else-if="toolOutputs[tc.id]?.text !== undefined"
                      class="max-h-64 overflow-auto whitespace-pre-wrap break-words rounded bg-muted p-2 font-mono text-[11px]"
                    >{{ formatOutput(toolOutputs[tc.id]!.text!) }}</pre>
                  </div>
                </CollapsibleContent>
              </Collapsible>
            </template>
          </template>
        </template>

        <p
          v-for="(s, i) in steps"
          :key="`step-${i}`"
          class="flex items-center gap-1.5 text-xs text-muted-foreground"
        >
          <Loader2 v-if="s.status === 'running'" class="size-3 animate-spin" />
          <CircleAlert v-else-if="s.status === 'error'" class="size-3 text-destructive" />
          <Check v-else class="size-3" />
          {{ toolLabel(s.tool) }}
        </p>
        <p v-if="busy && !steps.some((s) => s.status === 'running')" class="flex items-center gap-1.5 text-xs text-muted-foreground">
          <Loader2 class="size-3 animate-spin" />
          Thinking…
        </p>
      </div>
    </div>

    <form class="shrink-0 border-t p-3" @submit.prevent="submit">
      <p v-if="error" class="mb-2 text-xs text-destructive">{{ error }}</p>
      <div v-if="attachments.length || uploading" class="mb-2 flex flex-wrap gap-1.5">
        <span
          v-for="(a, i) in attachments"
          :key="a.upload_id"
          class="flex max-w-full items-center gap-1 rounded-md border px-2 py-1 text-xs"
        >
          <FileText class="size-3.5 shrink-0 text-muted-foreground" />
          <span class="truncate">{{ a.filename }}</span>
          <button
            type="button"
            class="rounded-sm text-muted-foreground hover:text-foreground"
            :title="`Remove ${a.filename}`"
            @click="attachments.splice(i, 1)"
          >
            <X class="size-3" />
          </button>
        </span>
        <span v-if="uploading" class="flex items-center gap-1 px-1 text-xs text-muted-foreground">
          <Loader2 class="size-3 animate-spin" />
          Uploading…
        </span>
      </div>
      <div class="relative">
        <Textarea
          v-model="input"
          rows="3"
          placeholder="Ask anything… (Shift+Enter for a new line)"
          class="resize-none pr-20"
          :disabled="busy"
          @keydown="onKeydown"
        />
        <input ref="fileInput" type="file" multiple class="hidden" @change="onFiles" />
        <Button
          type="button"
          variant="ghost"
          size="icon"
          class="absolute bottom-2 right-11 size-8"
          title="Attach files"
          :disabled="busy || uploading"
          @click="fileInput?.click()"
        >
          <Paperclip class="size-4" />
        </Button>
        <Button
          type="submit"
          size="icon"
          class="absolute bottom-2 right-2 size-8"
          title="Send"
          :disabled="!canSend"
        >
          <SendHorizontal class="size-4" />
        </Button>
      </div>
    </form>
  </aside>
</template>

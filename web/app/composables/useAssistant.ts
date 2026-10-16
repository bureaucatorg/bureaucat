export interface AssistantToolCall {
  id: string;
  name: string;
  args: Record<string, unknown>;
  has_output?: boolean;
  is_error?: boolean;
}

export interface AssistantAttachment {
  upload_id: string;
  filename: string;
  mime_type: string;
  size_bytes: number;
}

export interface AssistantPageContext {
  kind: string;
  project_key?: string;
  task_number?: number;
  page_number?: number;
  cycle_id?: string;
  module_id?: string;
  user_id?: string;
  tab?: string;
}

export interface AssistantChatMessage {
  author: string;
  bureaucat_user?: string;
  is_assistant?: boolean;
  text: string;
}

export interface AssistantChatContext {
  where: "channel" | "thread";
  messages: AssistantChatMessage[];
}

export interface AssistantMeta {
  page?: AssistantPageContext;
  chat?: AssistantChatContext;
}

export interface AssistantMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  tool_calls: AssistantToolCall[];
  attachments?: AssistantAttachment[];
  meta?: AssistantMeta;
  created_at: string;
}

export interface AssistantAction {
  id: string;
  tool_call_id: string;
  tool: string;
  title: string;
  args: Record<string, unknown>;
  status: "pending" | "approved" | "rejected" | "failed";
  result?: string;
}

export interface AssistantConversation {
  id: string;
  title: string;
  updated_at: string;
}

interface AssistantEvent {
  type: "step" | "message" | "pending" | "action" | "error" | "done";
  text?: string;
  tool?: string;
  status?: "running" | "done" | "error";
  action?: AssistantAction;
}

export interface AssistantStep {
  tool: string;
  status: "running" | "done" | "error";
}

// sessionStorage is per tab and survives a reload, but is not shared with newly opened tabs.
const OPEN_KEY = "bureaucat:assistant-open";

const STATIC_PAGES: Record<string, string> = {
  "/dashboard": "dashboard",
  "/projects": "projects",
  "/workspaces": "workspaces",
  "/views": "views",
  "/graph": "graph",
  "/cycles/active": "active_cycles",
  "/modules/active": "active_modules",
  "/settings": "settings",
};

// Describes the current route for the agent; the server re-validates it.
export function pageContextFromRoute(route: { path: string; params: Record<string, unknown>; query: Record<string, unknown> }): AssistantPageContext | undefined {
  const path = route.path.replace(/\/$/, "") || "/";
  if (STATIC_PAGES[path]) return { kind: STATIC_PAGES[path] };
  const key = typeof route.params.key === "string" ? route.params.key.toUpperCase() : undefined;
  let m: RegExpMatchArray | null;
  if ((m = path.match(/^\/projects\/[^/]+\/tasks\/(\d+)$/))) return { kind: "task", project_key: key, task_number: Number(m[1]) };
  if ((m = path.match(/^\/projects\/[^/]+\/pages\/(\d+)$/))) return { kind: "page", project_key: key, page_number: Number(m[1]) };
  if ((m = path.match(/^\/projects\/[^/]+\/cycles\/([^/]+)$/))) return { kind: "cycle", project_key: key, cycle_id: m[1] };
  if ((m = path.match(/^\/projects\/[^/]+\/modules\/([^/]+)$/))) return { kind: "module", project_key: key, module_id: m[1] };
  if (/^\/projects\/[^/]+$/.test(path)) {
    const tab = typeof route.query.tab === "string" ? route.query.tab : "tasks";
    return { kind: "project", project_key: key, tab };
  }
  if ((m = path.match(/^\/profile\/([^/]+)$/))) return { kind: "profile", user_id: m[1] };
  return undefined;
}

export function describePageContext(page?: AssistantPageContext): string {
  if (!page) return "";
  switch (page.kind) {
    case "task":
      return `${page.project_key}-${page.task_number}`;
    case "project":
      return page.tab && page.tab !== "tasks" ? `${page.project_key} ${page.tab}` : `${page.project_key}`;
    case "page":
      return `${page.project_key} page ${page.page_number}`;
    case "cycle":
      return `${page.project_key} cycle`;
    case "module":
      return `${page.project_key} module`;
    case "profile":
      return "a profile";
    default:
      return page.kind.replace(/_/g, " ");
  }
}
const CONVERSATION_KEY = "bureaucat:assistant-conversation";

// Module-level shared state so the sidebar button and the panel stay in sync.
const enabled = ref(false);
const statusLoaded = ref(false);
const open = ref(false);
const conversations = ref<AssistantConversation[]>([]);
const current = ref<AssistantConversation | null>(null);
const messages = ref<AssistantMessage[]>([]);
const actions = ref<AssistantAction[]>([]);
const steps = ref<AssistantStep[]>([]);
const busy = ref(false);
const error = ref("");

export function useAssistant() {
  const { getAuthHeader } = useAuth();

  function api(path: string, init: RequestInit = {}) {
    return fetch(`/api/v1/assistant${path}`, {
      ...init,
      headers: { "Content-Type": "application/json", ...getAuthHeader(), ...(init.headers || {}) },
      credentials: "include",
    });
  }

  async function fetchStatus() {
    try {
      const res = await api("/status");
      enabled.value = res.ok && (await res.json()).enabled === true;
    } catch {
      enabled.value = false;
    }
    statusLoaded.value = true;
    let savedId: string | null = null;
    try {
      open.value = sessionStorage.getItem(OPEN_KEY) === "1";
      savedId = sessionStorage.getItem(CONVERSATION_KEY);
    } catch {}
    if (enabled.value && savedId && !current.value) {
      await loadConversation(savedId);
      if (!current.value) error.value = "";
    }
  }

  function rememberConversation(id: string | null) {
    try {
      if (id) sessionStorage.setItem(CONVERSATION_KEY, id);
      else sessionStorage.removeItem(CONVERSATION_KEY);
    } catch {}
  }

  function setOpen(value: boolean) {
    open.value = value;
    try {
      sessionStorage.setItem(OPEN_KEY, value ? "1" : "0");
    } catch {}
  }

  async function listConversations() {
    try {
      const res = await api("/conversations");
      if (res.ok) conversations.value = await res.json();
    } catch {}
  }

  async function loadConversation(id: string) {
    error.value = "";
    steps.value = [];
    try {
      const res = await api(`/conversations/${id}`);
      if (!res.ok) {
        error.value = "Conversation not found";
        if (res.status === 404) rememberConversation(null);
        return;
      }
      const data = await res.json();
      current.value = { id: data.id, title: data.title, updated_at: data.updated_at };
      rememberConversation(data.id);
      messages.value = data.messages;
      actions.value = data.actions;
    } catch {
      error.value = "Network error";
    }
  }

  function newConversation() {
    current.value = null;
    rememberConversation(null);
    messages.value = [];
    actions.value = [];
    steps.value = [];
    error.value = "";
  }

  function reset() {
    newConversation();
    conversations.value = [];
    setOpen(false);
    enabled.value = false;
  }

  async function deleteConversation(id: string) {
    const res = await api(`/conversations/${id}`, { method: "DELETE" });
    if (!res.ok) return false;
    conversations.value = conversations.value.filter((c) => c.id !== id);
    if (current.value?.id === id) newConversation();
    return true;
  }

  async function fetchToolOutput(convId: string, callId: string): Promise<string | null> {
    const res = await api(`/conversations/${convId}/tool-calls/${encodeURIComponent(callId)}/output`);
    if (!res.ok) return null;
    return (await res.json()).output;
  }

  async function renameConversation(id: string, title: string) {
    const res = await api(`/conversations/${id}`, { method: "PATCH", body: JSON.stringify({ title }) });
    if (!res.ok) return false;
    const saved: string = (await res.json()).title;
    const c = conversations.value.find((c) => c.id === id);
    if (c) c.title = saved;
    if (current.value?.id === id) current.value.title = saved;
    return true;
  }

  async function send(text: string, attachments: AssistantAttachment[] = [], meta: AssistantMeta = {}) {
    if (busy.value) return;
    error.value = "";
    if (!current.value) {
      const res = await api("/conversations", { method: "POST" });
      if (!res.ok) {
        error.value = "Could not start a conversation";
        return;
      }
      current.value = await res.json();
      rememberConversation(current.value!.id);
    }
    for (const a of actions.value) {
      if (a.status === "pending") a.status = "rejected";
    }
    messages.value.push({
      id: `local-${Date.now()}`,
      role: "user",
      content: text,
      tool_calls: [],
      attachments,
      meta,
      created_at: new Date().toISOString(),
    });
    await stream(`/conversations/${current.value!.id}/messages`, {
      text,
      upload_ids: attachments.map((a) => a.upload_id),
      meta,
    });
  }

  async function resolve(action: AssistantAction, approve: boolean) {
    if (busy.value || !current.value) return;
    error.value = "";
    await stream(`/conversations/${current.value.id}/actions/${action.id}`, { approve });
  }

  async function stream(path: string, body: unknown) {
    busy.value = true;
    steps.value = [];
    const convId = current.value!.id;
    try {
      const res = await api(path, { method: "POST", body: JSON.stringify(body) });
      if (!res.ok || !res.body) {
        const data = await res.json().catch(() => ({}));
        error.value = data.message || "Assistant request failed";
        return;
      }
      const reader = res.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";
      for (;;) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        let idx;
        while ((idx = buffer.indexOf("\n\n")) >= 0) {
          const chunk = buffer.slice(0, idx);
          buffer = buffer.slice(idx + 2);
          const line = chunk.split("\n").find((l) => l.startsWith("data: "));
          if (line) handleEvent(JSON.parse(line.slice(6)));
        }
      }
    } catch {
      error.value = "Connection lost";
    } finally {
      busy.value = false;
      const keepError = error.value;
      if (current.value?.id === convId) await loadConversation(convId);
      await listConversations();
      error.value = keepError;
    }
  }

  function handleEvent(ev: AssistantEvent) {
    switch (ev.type) {
      case "step": {
        const existing = steps.value.findLast((s) => s.tool === ev.tool && s.status === "running");
        if (existing && ev.status !== "running") existing.status = ev.status!;
        else steps.value.push({ tool: ev.tool!, status: ev.status! });
        break;
      }
      case "message":
        messages.value.push({
          id: `local-${Date.now()}-${messages.value.length}`,
          role: "assistant",
          content: ev.text || "",
          tool_calls: [],
          created_at: new Date().toISOString(),
        });
        break;
      case "pending":
        if (ev.action) actions.value.push(ev.action);
        break;
      case "action": {
        const i = actions.value.findIndex((a) => a.id === ev.action?.id);
        if (i >= 0 && ev.action) actions.value[i] = ev.action;
        break;
      }
      case "error":
        error.value = ev.text || "Assistant error";
        break;
    }
  }

  return {
    enabled,
    statusLoaded,
    open,
    conversations,
    current,
    messages,
    actions,
    steps,
    busy,
    error,
    fetchStatus,
    setOpen,
    listConversations,
    loadConversation,
    newConversation,
    reset,
    deleteConversation,
    renameConversation,
    fetchToolOutput,
    send,
    resolve,
  };
}

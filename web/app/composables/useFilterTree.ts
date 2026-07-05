/**
 * useFilterTree owns the current filter document for a project view and
 * synchronises it with the URL.
 *
 * URL contract:
 *   ?f=<base64url(JSON(FilterTree))>   authoritative filter
 *   ?view=<slug>                        active saved view (informational)
 *   ?group_by=<key>                     only meaningful on ?tab=board
 *   ?sort_by=<key>&sort_dir=<asc|desc>
 *
 * Legacy URLs containing ?state_id=, ?priority=, ?assigned_to=, ?q=,
 * ?from_date=, ?to_date= are detected and migrated on first load.
 */

import type { LocationQueryRaw } from "vue-router";
import type {
  FilterTree,
  FilterNode,
  Predicate,
  SortKey,
  SortDir,
  ViewGroupBy,
} from "~/types";

const LEGACY_PARAMS = [
  "state_id",
  "state_type",
  "created_by",
  "assigned_to",
  "priority",
  "q",
  "from_date",
  "to_date",
] as const;

const DEFAULT_SORT_BY: SortKey = "created_at";
const DEFAULT_SORT_DIR: SortDir = "desc";
const DEFAULT_GROUP_BY: ViewGroupBy = "state";

// Query params that make up the "filter state" of a project view. These are
// persisted to localStorage per project so that navigating back to the project
// (e.g. via the breadcrumb, which carries no query) restores the last filters.
const PERSIST_KEYS = ["f", "q", "view", "sort_by", "sort_dir", "group_by"] as const;

function filterStorageKey(projectKey: string): string {
  return `bureaucat:filters:${projectKey}`;
}

function emptyTree(): FilterTree {
  return { children: [] };
}

// ---- base64url helpers (no padding) ----

function encodeBase64Url(json: string): string {
  // Use the UTF-8-safe trick to encode arbitrary JSON.
  const bytes = new TextEncoder().encode(json);
  let bin = "";
  for (const b of bytes) bin += String.fromCharCode(b);
  return btoa(bin).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

function decodeBase64Url(s: string): string {
  const padded = s.replace(/-/g, "+").replace(/_/g, "/");
  const bin = atob(padded);
  const bytes = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
  return new TextDecoder().decode(bytes);
}

function encodeTree(tree: FilterTree): string {
  return encodeBase64Url(JSON.stringify(tree));
}

function decodeTree(raw: string): FilterTree | null {
  try {
    const parsed = JSON.parse(decodeBase64Url(raw));
    if (!parsed || !Array.isArray(parsed.children)) return null;
    return parsed as FilterTree;
  } catch {
    return null;
  }
}

function hasAnyLegacyParam(query: Record<string, unknown>): boolean {
  return LEGACY_PARAMS.some((k) => query[k]);
}

/**
 * migrateLegacyQuery converts the old ?state_id=...&priority=... shape to an
 * equivalent FilterTree so that historic bookmarks continue to work.
 */
function migrateLegacyQuery(query: Record<string, unknown>): FilterTree {
  const children: FilterNode[] = [];
  const s = (k: string) => (typeof query[k] === "string" ? (query[k] as string) : "");

  if (s("state_id")) {
    children.push({ predicate: { field: "state", op: "in", value: [s("state_id")] } });
  }
  if (s("state_type")) {
    children.push({ predicate: { field: "state_type", op: "in", value: [s("state_type")] } });
  }
  if (s("created_by")) {
    children.push({ predicate: { field: "created_by", op: "in", value: [s("created_by")] } });
  }
  if (s("assigned_to")) {
    children.push({ predicate: { field: "assignees", op: "has_any", value: [s("assigned_to")] } });
  }
  if (s("priority")) {
    const n = parseInt(s("priority"), 10);
    if (!isNaN(n)) {
      children.push({ predicate: { field: "priority", op: "in", value: [n] } });
    }
  }
  // Note: legacy ?q= is preserved as a separate ?q= param, not as a chip.
  if (s("from_date")) {
    children.push({ predicate: { field: "created_at", op: "after", value: s("from_date") } });
  }
  if (s("to_date")) {
    children.push({ predicate: { field: "created_at", op: "before", value: s("to_date") } });
  }
  return { children };
}

/**
 * structurally compare two FilterTrees for equality. Used to flag whether the
 * live tree has drifted from the active saved view's tree.
 */
export function filterTreesEqual(a: FilterTree | null, b: FilterTree | null): boolean {
  if (a === b) return true;
  if (!a || !b) return false;
  return JSON.stringify(a) === JSON.stringify(b);
}

export function useFilterTree() {
  const route = useRoute();
  const router = useRouter();

  // The live tree — reactive; synced from the URL on mount and whenever ?f= changes.
  const tree = ref<FilterTree>(emptyTree());

  const projectKey = () =>
    typeof route.params.key === "string" ? route.params.key : "";

  // Distinguishes URL writes we initiate (filter mutations, sort/group/search,
  // view selection) from query changes we did NOT cause (breadcrumb to the bare
  // project URL, browser back/forward). Internal writes persist the new state;
  // external ones that land on the empty project URL restore saved filters
  // instead of being mistaken for a "clear".
  let internalWrite = false;
  function replaceQuery(query: LocationQueryRaw) {
    internalWrite = true;
    return router.replace({ query });
  }

  // Persist the filter-related query params for the current project. Called
  // whenever any of them change. Cleared filters store as {} so they don't get
  // re-applied on the next visit.
  function persistFilters() {
    if (!import.meta.client) return;
    const key = projectKey();
    if (!key) return;
    // Only persist filter changes made on the overview page itself. (Callers
    // gate this via the internalWrite flag, so this only runs for our own URL
    // writes, which always happen here — but the guard keeps it safe.)
    if (route.path !== `/projects/${key}`) return;
    const data: Record<string, string> = {};
    for (const k of PERSIST_KEYS) {
      const v = route.query[k];
      if (typeof v === "string" && v) data[k] = v;
    }
    try {
      localStorage.setItem(filterStorageKey(key), JSON.stringify(data));
    } catch {
      // ignore quota / unavailable storage
    }
  }

  function readPersistedFilters(): Record<string, string> | null {
    if (!import.meta.client) return null;
    const key = projectKey();
    if (!key) return null;
    try {
      const raw = localStorage.getItem(filterStorageKey(key));
      if (!raw) return null;
      const parsed = JSON.parse(raw);
      return parsed && typeof parsed === "object" ? parsed : null;
    } catch {
      return null;
    }
  }

  // Decode the tree sitting in the URL, if any.
  function readFromUrl(): FilterTree {
    const raw = route.query.f;
    if (typeof raw === "string" && raw) {
      return decodeTree(raw) ?? emptyTree();
    }
    return emptyTree();
  }

  // Push the current tree back into the URL under ?f= (omits when empty).
  function writeToUrl(next: FilterTree, opts: { resetPage?: boolean } = {}) {
    const q = { ...route.query };
    if (next.children.length === 0) {
      delete q.f;
    } else {
      q.f = encodeTree(next);
    }
    if (opts.resetPage) delete q.page;
    void replaceQuery(q);
  }

  function setTree(next: FilterTree, opts?: { resetPage?: boolean }) {
    tree.value = next;
    writeToUrl(next, { resetPage: true, ...opts });
  }

  function clearTree() {
    setTree(emptyTree());
  }

  /**
   * Clear everything that scopes the task list — filter tree, free-text search
   * and the active saved view — in a single router.replace so the writes can't
   * race and strand stale params in the URL.
   */
  function clearAll() {
    tree.value = emptyTree();
    const q = { ...route.query };
    delete q.f;
    delete q.q;
    delete q.view;
    delete q.page;
    void replaceQuery(q);
  }

  function addPredicate(p: Predicate) {
    setTree({ children: [...tree.value.children, { predicate: p }] });
  }

  function replacePredicateAt(index: number, p: Predicate) {
    const next = [...tree.value.children];
    next[index] = { predicate: p };
    setTree({ children: next });
  }

  function removeNodeAt(index: number) {
    const next = tree.value.children.filter((_, i) => i !== index);
    setTree({ children: next });
  }

  // Sort and group-by accessors backed by URL query.
  const sortBy = computed<SortKey>({
    get: () => ((route.query.sort_by as SortKey) ?? DEFAULT_SORT_BY),
    set: (v) => {
      const q = { ...route.query };
      if (!v || v === DEFAULT_SORT_BY) delete q.sort_by;
      else q.sort_by = v;
      delete q.page;
      void replaceQuery(q);
    },
  });
  const sortDir = computed<SortDir>({
    get: () => ((route.query.sort_dir as SortDir) ?? DEFAULT_SORT_DIR),
    set: (v) => {
      const q = { ...route.query };
      if (!v || v === DEFAULT_SORT_DIR) delete q.sort_dir;
      else q.sort_dir = v;
      delete q.page;
      void replaceQuery(q);
    },
  });
  const groupBy = computed<ViewGroupBy>({
    get: () => ((route.query.group_by as ViewGroupBy) ?? DEFAULT_GROUP_BY),
    set: (v) => {
      const q = { ...route.query };
      if (!v || v === DEFAULT_GROUP_BY) delete q.group_by;
      else q.group_by = v;
      void replaceQuery(q);
    },
  });

  const activeViewSlug = computed(() => {
    const v = route.query.view;
    return typeof v === "string" && v ? v : null;
  });

  /** Free-text search bound to ?q= — kept out of the chip tree by design. */
  const searchQuery = computed<string>({
    get: () => (typeof route.query.q === "string" ? (route.query.q as string) : ""),
    set: (v) => {
      const q = { ...route.query };
      if (v) q.q = v;
      else delete q.q;
      delete q.page;
      void replaceQuery(q);
    },
  });

  /**
   * effectiveTree is what the API actually receives. The free-text search
   * box is emitted as a single `search contains X` predicate — the server
   * expands that opcode to match both title and description internally.
   */
  const effectiveTree = computed<FilterTree>(() => {
    if (!searchQuery.value) return tree.value;
    const searchNode: FilterNode = {
      predicate: { field: "search", op: "contains", value: searchQuery.value },
    };
    return { children: [searchNode, ...tree.value.children] };
  });

  function setActiveView(slug: string | null) {
    const q = { ...route.query };
    if (slug) q.view = slug;
    else delete q.view;
    void replaceQuery(q);
  }

  /**
   * Run once on mount. Detects legacy URL params and rewrites to ?f=, then
   * hydrates the live tree from the URL. If the URL carries no filter params
   * at all (e.g. arriving via the breadcrumb), the last-used filters for this
   * project are restored from localStorage.
   */
  async function hydrateFromUrl() {
    if (hasAnyLegacyParam(route.query)) {
      const migrated = migrateLegacyQuery(route.query as Record<string, unknown>);
      const q = { ...route.query };
      for (const k of LEGACY_PARAMS) delete q[k];
      if (migrated.children.length > 0) {
        q.f = encodeTree(migrated);
      }
      await replaceQuery(q);
      tree.value = migrated;
      return;
    }

    // No filter params in the URL → try restoring the project's saved filters.
    const hasFilterParam = PERSIST_KEYS.some((k) => route.query[k]);
    if (!hasFilterParam) {
      const saved = readPersistedFilters();
      if (saved && Object.keys(saved).length > 0) {
        await replaceQuery({ ...route.query, ...saved });
        tree.value = saved.f ? decodeTree(saved.f) ?? emptyTree() : emptyTree();
        return;
      }
    }

    tree.value = readFromUrl();
  }

  // Keep the local tree in sync if the URL changes externally (back/forward).
  watch(
    () => route.query.f,
    () => {
      tree.value = readFromUrl();
    }
  );

  // React to filter-param changes. Our own writes persist the new state; a
  // change we didn't make that lands on the bare project URL (breadcrumb,
  // back/forward) restores the saved filters instead of losing them.
  watch(
    () => PERSIST_KEYS.map((k) => route.query[k]),
    () => {
      if (internalWrite) {
        // A change we made — persist it. An empty state is stored as {} so an
        // explicit clear is remembered and not re-applied on the next visit.
        internalWrite = false;
        persistFilters();
        return;
      }
      // A navigation we didn't initiate. If we've landed on the project overview
      // with no filter params, restore the last-used filters.
      const key = projectKey();
      if (!key || route.path !== `/projects/${key}`) return;
      if (PERSIST_KEYS.some((k) => route.query[k])) return;
      const saved = readPersistedFilters();
      if (saved && Object.keys(saved).length > 0) {
        void replaceQuery({ ...route.query, ...saved });
        tree.value = saved.f ? decodeTree(saved.f) ?? emptyTree() : emptyTree();
      }
    }
  );

  return {
    tree: computed(() => tree.value),
    setTree,
    clearTree,
    clearAll,
    addPredicate,
    replacePredicateAt,
    removeNodeAt,
    sortBy,
    sortDir,
    groupBy,
    activeViewSlug,
    setActiveView,
    searchQuery,
    effectiveTree,
    hydrateFromUrl,
    encodeTree,
    decodeTree,
  };
}

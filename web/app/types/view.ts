import type { FilterTree, ViewGroupBy, SortKey, SortDir } from "./task";

export type ViewVisibility = "private" | "shared";

/** A saved view returned by the API. */
export interface ProjectView {
  id: string;
  project_id: string;
  slug: string;
  name: string;
  description?: string;
  visibility: ViewVisibility;
  owner_id: string;
  filter_tree: FilterTree;
  group_by: ViewGroupBy;
  sort_by: SortKey;
  sort_dir: SortDir;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface CreateViewInput {
  name: string;
  description?: string;
  visibility?: ViewVisibility;
  filter_tree: FilterTree;
  group_by?: ViewGroupBy;
  sort_by?: SortKey;
  sort_dir?: SortDir;
}

export interface UpdateViewInput {
  name?: string;
  description?: string | null;
  visibility?: ViewVisibility;
  filter_tree?: FilterTree;
  group_by?: ViewGroupBy;
  sort_by?: SortKey;
  sort_dir?: SortDir;
  position?: number;
}

/**
 * Non-persisted presets rendered in the view strip before user-created views.
 * These are not stored in project_views; they are built in the frontend.
 */
export type PresetViewId = "all" | "mine" | "overdue";

export interface PresetView {
  id: PresetViewId;
  name: string;
  filter_tree: FilterTree;
}

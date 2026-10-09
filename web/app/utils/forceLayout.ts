export interface Point {
  x: number;
  y: number;
}

const GRID_COL_WIDTH = 200;
const GRID_ROW_HEIGHT = 70;
const GRID_GAP = 120;

/**
 * Lays out linked nodes with a force simulation, then packs nodes that have no
 * edges into a grid just below them (repulsion would otherwise push them far out).
 */
export function forceLayout(
  ids: string[],
  edges: [string, string][],
  iterations = 300,
  k = 110,
): Map<string, Point> {
  const known = new Set(ids);
  const linked = new Set<string>();
  for (const [a, b] of edges) {
    if (known.has(a) && known.has(b)) {
      linked.add(a);
      linked.add(b);
    }
  }
  const positions = simulate(ids.filter((id) => linked.has(id)), edges, iterations, k);

  const isolated = ids.filter((id) => !linked.has(id));
  if (!isolated.length) return positions;

  let minX = 0;
  let maxX = 0;
  let maxY = -GRID_GAP;
  if (positions.size) {
    const pts = [...positions.values()];
    minX = Math.min(...pts.map((p) => p.x));
    maxX = Math.max(...pts.map((p) => p.x));
    maxY = Math.max(...pts.map((p) => p.y));
  }
  const cols = Math.max(4, Math.round((maxX - minX) / GRID_COL_WIDTH) + 1);
  isolated.forEach((id, i) => {
    positions.set(id, {
      x: minX + (i % cols) * GRID_COL_WIDTH,
      y: maxY + GRID_GAP + Math.floor(i / cols) * GRID_ROW_HEIGHT,
    });
  });
  return positions;
}

/**
 * Fruchterman-Reingold layout. Nodes start on a circle in the given order,
 * so the result is deterministic for the same input.
 */
function simulate(
  ids: string[],
  edges: [string, string][],
  iterations: number,
  k: number,
): Map<string, Point> {
  const n = ids.length;
  const index = new Map(ids.map((id, i) => [id, i]));
  const x = new Float64Array(n);
  const y = new Float64Array(n);
  const radius = k * Math.sqrt(n);

  for (let i = 0; i < n; i++) {
    const angle = (2 * Math.PI * i) / n;
    x[i] = radius * Math.cos(angle);
    y[i] = radius * Math.sin(angle);
  }

  const links: [number, number][] = [];
  for (const [a, b] of edges) {
    const ia = index.get(a);
    const ib = index.get(b);
    if (ia !== undefined && ib !== undefined) links.push([ia, ib]);
  }

  const dx = new Float64Array(n);
  const dy = new Float64Array(n);
  let temp = radius / 4;
  const cooling = temp / iterations;

  for (let iter = 0; iter < iterations; iter++) {
    dx.fill(0);
    dy.fill(0);

    for (let i = 0; i < n; i++) {
      for (let j = i + 1; j < n; j++) {
        const ddx = x[i]! - x[j]!;
        const ddy = y[i]! - y[j]!;
        const dist = Math.hypot(ddx, ddy) || 0.01;
        const f = (k * k) / dist / dist;
        dx[i]! += ddx * f;
        dy[i]! += ddy * f;
        dx[j]! -= ddx * f;
        dy[j]! -= ddy * f;
      }
    }

    for (const [a, b] of links) {
      const ddx = x[a]! - x[b]!;
      const ddy = y[a]! - y[b]!;
      const dist = Math.hypot(ddx, ddy) || 0.01;
      const f = dist / k;
      dx[a]! -= ddx * f;
      dy[a]! -= ddy * f;
      dx[b]! += ddx * f;
      dy[b]! += ddy * f;
    }

    for (let i = 0; i < n; i++) {
      dx[i]! -= x[i]! * 0.05;
      dy[i]! -= y[i]! * 0.05;
      const len = Math.hypot(dx[i]!, dy[i]!);
      if (len > 0) {
        const step = Math.min(len, temp);
        x[i]! += (dx[i]! / len) * step;
        y[i]! += (dy[i]! / len) * step;
      }
    }

    temp = Math.max(temp - cooling, 1);
  }

  return new Map(ids.map((id, i) => [id, { x: x[i]!, y: y[i]! }]));
}

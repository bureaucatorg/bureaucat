const NS = "http://www.w3.org/2000/svg";
const W = 800;
const H = 640;
export const CHAPTERS = 7;
// Scroll distance per chapter, as a fraction of the viewport height.
export const CHAPTER_VH = 0.9;

type Attrs = Record<string, string | number>;
type Child = SVGElement | string;

function h<T extends SVGElement = SVGGElement>(tag: string, attrs: Attrs = {}, children: Child[] = []): T {
  const el = document.createElementNS(NS, tag) as T;
  for (const [k, v] of Object.entries(attrs)) el.setAttribute(k, String(v));
  for (const c of children) el.append(typeof c === "string" ? document.createTextNode(c) : c);
  return el;
}

const clamp = (v: number, a = 0, b = 1) => Math.min(b, Math.max(a, v));
const lerp = (a: number, b: number, t: number) => a + (b - a) * t;
const range = (p: number, a: number, b: number) => clamp((p - a) / (b - a));
const easeOut = (t: number) => 1 - Math.pow(1 - t, 3);
const easeIn = (t: number) => t * t * t;
const easeInOut = (t: number) => (t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2);
const easeBack = (t: number) => {
  const c1 = 1.70158;
  const c3 = c1 + 1;
  return 1 + c3 * Math.pow(t - 1, 3) + c1 * Math.pow(t - 1, 2);
};
const spring = (t: number) => (t <= 0 ? 0 : t >= 1 ? 1 : 1 - Math.exp(-5 * t) * Math.cos(9 * t));
const f2 = (n: number) => n.toFixed(2);

function tf(el: SVGElement, x: number, y: number, s = 1, r = 0) {
  el.setAttribute("transform", `translate(${f2(x)} ${f2(y)}) rotate(${f2(r)}) scale(${f2(Math.max(s, 0.0001))})`);
}
function op(el: SVGElement, o: number) {
  el.setAttribute("opacity", f2(clamp(o)));
}
function draw(el: SVGElement, t: number) {
  el.setAttribute("stroke-dashoffset", f2(1 - clamp(t)));
}
function setText(el: SVGElement, text: string) {
  if (el.textContent !== text) el.textContent = text;
}
// Scales `el` about (cx, cy) in its parent's coordinate space.
function about(el: SVGElement, cx: number, cy: number, sx: number, sy = sx, r = 0) {
  el.setAttribute(
    "transform",
    `translate(${f2(cx)} ${f2(cy)}) rotate(${f2(r)}) scale(${f2(Math.max(sx, 0.0001))} ${f2(Math.max(sy, 0.0001))}) translate(${f2(-cx)} ${f2(-cy)})`,
  );
}

const MONO = "font-mono";
const mono = (x: number, y: number, size: number, text: string, cls = "fill-muted-foreground", extra: Attrs = {}) =>
  h("text", { x, y, "font-size": size, class: `${MONO} ${cls}`, ...extra }, [text]);
const sans = (x: number, y: number, size: number, text: string, cls = "fill-foreground", extra: Attrs = {}) =>
  h("text", { x, y, "font-size": size, class: `font-sans ${cls}`, ...extra }, [text]);
const card = (x: number, y: number, w: number, hh: number, r = 10, cls = "fill-card stroke-border") =>
  h("rect", { x, y, width: w, height: hh, rx: r, class: cls, "stroke-width": 1.2 });
const lines = (x: number, y: number, widths: number[], gap = 8) =>
  h("g", {}, widths.map((w, i) => h("rect", { x, y: y + i * gap, width: w, height: 3.5, rx: 1.75, class: "fill-muted-foreground/25" })));

const STATES = [
  { name: "Backlog", color: "#6B7280", type: "backlog" },
  { name: "Todo", color: "#3B82F6", type: "unstarted" },
  { name: "Approval Pending", color: "#F59E0B", type: "unstarted" },
  { name: "In Progress", color: "#10B981", type: "started" },
  { name: "Blocked", color: "#EF4444", type: "started" },
  { name: "Testing", color: "#8B5CF6", type: "started" },
  { name: "Done", color: "#22C55E", type: "completed" },
];
const PRIORITIES = [
  { name: "No priority", color: "#6B7280" },
  { name: "Low", color: "#3B82F6" },
  { name: "Medium", color: "#EAB308" },
  { name: "High", color: "#F97316" },
  { name: "Urgent", color: "#EF4444" },
];

const TONES = {
  sky: ["fill-sky-100", "fill-sky-800"],
  violet: ["fill-violet-100", "fill-violet-800"],
  emerald: ["fill-emerald-100", "fill-emerald-800"],
  rose: ["fill-rose-100", "fill-rose-800"],
} as const;
type Tone = keyof typeof TONES;

function avatar(x: number, y: number, r: number, initials: string, tone: Tone) {
  const [bg, fg] = TONES[tone];
  return h("g", { transform: `translate(${x} ${y})` }, [
    h("circle", { r, class: `${bg} stroke-card`, "stroke-width": 2 }),
    mono(0, r * 0.36, r * 0.9, initials, fg, { "text-anchor": "middle", "font-weight": 700 }),
  ]);
}

interface Scene {
  root: SVGGElement;
  update(p: number, t: number): void;
}
interface Shared {
  nod: number;
}

// ─── § 01 · Meet the cat: loose paperwork orbits, then files itself ───
function sceneIntake(): Scene {
  const root = h("g");
  const n = 9;
  const tray = h("g", {}, [
    h("path", { d: "M330 548 h140 l-10 22 h-120 z", class: "fill-amber-500/15 stroke-amber-600/60", "stroke-width": 1.5 }),
  ]);
  const counter = mono(400, 596, 12, "IN-TRAY · 0 / 9 FILED", "fill-muted-foreground", { "text-anchor": "middle", "letter-spacing": 1.5 });
  root.append(tray, counter);
  const tones = STATES.map((s) => s.color);
  const papers = Array.from({ length: n }, (_, k) => {
    const g = h("g", {}, [
      card(-20, -26, 40, 52, 4),
      h("rect", { x: -13, y: -18, width: 12, height: 3.5, rx: 1.75, fill: tones[k % tones.length]! }),
      lines(-13, -9, [26, 22, 26, 16]),
    ]);
    root.append(g);
    return g;
  });
  let filed = -1;
  return {
    root,
    update(p, t) {
      let done = 0;
      papers.forEach((g, k) => {
        const a = (k / n) * Math.PI * 2 + t * 0.18 + p * 1.4;
        const depth = Math.sin(a);
        const ox = 400 + Math.cos(a) * 310;
        const oy = 300 + depth * 205;
        const os = 0.85 + depth * 0.25;
        const or = Math.sin(t * 0.6 + k * 1.7) * 28 + k * 37;
        const sx = 400 + (k % 2 ? 2 : -2);
        const sy = 540 - k * 3.4;
        const sr = ((k % 3) - 1) * 2.5;
        const f = easeInOut(range(p, 0.12 + k * 0.045, 0.45 + k * 0.045));
        if (f > 0.99) done++;
        tf(g, lerp(ox, sx, f), lerp(oy, sy, f), lerp(os, 1.15, f), lerp(or, sr, f));
      });
      op(tray, range(p, 0.05, 0.25));
      op(counter, range(p, 0.1, 0.3));
      if (done !== filed) {
        filed = done;
        setText(counter, `IN-TRAY · ${done} / ${n} FILED`);
      }
    },
  };
}

// ─── § 02 · Workspaces → projects → tasks ───
function sceneHierarchy(): Scene {
  const root = h("g");
  const back = h("g", {}, [
    h("rect", { x: 80, y: 90, width: 640, height: 440, rx: 18, class: "fill-none stroke-border", "stroke-dasharray": "6 6", "stroke-width": 1.2 }),
    mono(104, 80, 11, "WORKSPACE · OPS", "fill-muted-foreground/70", { "letter-spacing": 1.5 }),
  ]);
  const wsFill = h("rect", { x: 80, y: 90, width: 640, height: 440, rx: 18, class: "fill-card/70" });
  const wsLine = h("rect", { x: 80, y: 90, width: 640, height: 440, rx: 18, pathLength: 1, "stroke-dasharray": "1 1", class: "fill-none stroke-foreground/30", "stroke-width": 1.4 });
  const wsLabel = h("g", {}, [
    mono(104, 124, 12, "WORKSPACE", "fill-muted-foreground", { "letter-spacing": 1.5 }),
    mono(184, 124, 12, "ENG", "fill-amber-600 dark:fill-amber-500", { "font-weight": 700, "letter-spacing": 1.5 }),
    mono(696, 124, 11, "+2 projects", "fill-muted-foreground/70", { "text-anchor": "end" }),
  ]);
  const prLine = h("rect", { x: 112, y: 146, width: 576, height: 352, rx: 14, pathLength: 1, "stroke-dasharray": "1 1", class: "fill-background/60 stroke-foreground/25", "stroke-width": 1.2 });
  const prLabel = h("g", {}, [
    h("rect", { x: 136, y: 166, width: 18, height: 18, rx: 5, class: "fill-amber-500" }),
    mono(145, 179.5, 10, "D", "fill-amber-950", { "text-anchor": "middle", "font-weight": 700 }),
    mono(164, 180, 12, "PROJECT", "fill-muted-foreground", { "letter-spacing": 1.5 }),
    mono(228, 180, 12, "DEVOPS", "fill-foreground", { "font-weight": 700, "letter-spacing": 1.5 }),
  ]);
  root.append(back, wsFill, wsLine, wsLabel, prLine, prLabel);

  const tasks = [
    { id: "DEVOPS-778", title: "Q3 vendor audit", s: 6 },
    { id: "DEVOPS-779", title: "Laptop request", s: 3 },
    { id: "DEVOPS-780", title: "Renew contract", s: 2 },
  ];
  const cards = tasks.map((tk, k) => {
    const x = 136 + k * 180;
    const st = STATES[tk.s]!;
    const g = h("g", {}, [
      card(x, 206, 168, 108, 10),
      mono(x + 14, 230, 11, tk.id, "fill-muted-foreground"),
      sans(x + 14, 254, 15, tk.title, "fill-foreground", { "font-weight": 600 }),
      lines(x + 14, 268, [110, 80], 9),
      h("circle", { cx: x + 19, cy: 297, r: 4.5, fill: st.color }),
      mono(x + 30, 301, 10.5, st.name, "fill-muted-foreground"),
    ]);
    root.append(g);
    return g;
  });
  const ghost = h("g", {}, [
    h("rect", { x: 136, y: 334, width: 168, height: 108, rx: 10, class: "fill-none stroke-amber-500/70", "stroke-dasharray": "5 5", "stroke-width": 1.4 }),
    mono(150, 362, 11, "DEVOPS-781", "fill-amber-600 dark:fill-amber-500", { "font-weight": 700 }),
    sans(150, 386, 15, "New task", "fill-muted-foreground", { "font-weight": 600 }),
  ]);
  const caret = h("rect", { x: 216, y: 373, width: 1.6, height: 17, class: "fill-amber-500" });
  ghost.append(caret);
  const hint = mono(330, 394, 11, "← numbers itself, per project", "fill-muted-foreground/80");
  root.append(ghost, hint);

  return {
    root,
    update(p, t) {
      const b = easeOut(range(p, 0.08, 0.3));
      tf(back, 24 * b, -24 * b);
      op(back, b * 0.8);
      draw(wsLine, easeInOut(range(p, 0, 0.3)));
      op(wsFill, range(p, 0.15, 0.35));
      op(wsLabel, range(p, 0.18, 0.32));
      draw(prLine, easeInOut(range(p, 0.2, 0.45)));
      op(prLabel, range(p, 0.35, 0.48));
      cards.forEach((g, k) => {
        const f = range(p, 0.4 + k * 0.08, 0.6 + k * 0.08);
        tf(g, 0, (1 - easeBack(f)) * 36);
        op(g, f * 2);
      });
      const gf = range(p, 0.72, 0.84);
      tf(ghost, 0, (1 - easeOut(gf)) * 20);
      op(ghost, gf);
      op(caret, Math.sin(t * 6) > 0 ? 1 : 0);
      op(hint, range(p, 0.8, 0.9));
    },
  };
}

// ─── § 03 · Lifecycle: one task rides the default states ───
function sceneLifecycle(shared: Shared): Scene {
  const root = h("g");
  const X0 = 80;
  const DX = 106.67;
  const RY = 150;
  const sx = (i: number) => X0 + i * DX;

  const groups: [string, number, number][] = [
    ["backlog", 0, 0],
    ["unstarted", 1, 2],
    ["started", 3, 5],
    ["completed", 6, 6],
  ];
  for (const [name, a, b] of groups) {
    const x1 = sx(a) - 18;
    const x2 = sx(b) + 18;
    root.append(
      h("path", { d: `M${x1} 112 v-6 H${x2} v6`, class: "fill-none stroke-muted-foreground/40", "stroke-width": 1 }),
      mono((x1 + x2) / 2, 98, 10, name, "fill-muted-foreground/80", { "text-anchor": "middle", "letter-spacing": 1 }),
    );
  }
  root.append(h("line", { x1: X0, y1: RY, x2: sx(6), y2: RY, class: "stroke-border", "stroke-width": 3, "stroke-linecap": "round" }));
  const trail = h("line", { x1: X0, y1: RY, x2: X0, y2: RY, "stroke-width": 3, "stroke-linecap": "round" });
  root.append(trail);
  const nodes = STATES.map((s, i) => {
    const words = s.name.split(" ");
    const label = h("text", { x: sx(i), y: 182, "font-size": 10.5, class: `${MONO} fill-muted-foreground`, "text-anchor": "middle" },
      words.map((w, j) => h("tspan", { x: sx(i), dy: j ? 13 : 0 }, [w])));
    const dot = h("circle", { cx: sx(i), cy: RY, r: 7, class: "fill-background", stroke: s.color, "stroke-width": 2.5 });
    root.append(dot, label);
    return dot;
  });
  const pulse = h("circle", { cx: 0, cy: 0, r: 12, fill: "none", stroke: "#F59E0B", "stroke-width": 2 });
  const token = h("g", {}, [pulse, h("circle", { r: 11, class: "stroke-card", "stroke-width": 3 }), h("circle", { r: 3.5, class: "fill-card" })]);
  const tokenFill = token.children[1] as SVGElement;
  const burst = h("g", {}, Array.from({ length: 10 }, (_, k) => {
    const a = (k / 10) * Math.PI * 2;
    return h("line", { x1: Math.cos(a) * 18, y1: Math.sin(a) * 18, x2: Math.cos(a) * 30, y2: Math.sin(a) * 30, stroke: "#22C55E", "stroke-width": 2.5, "stroke-linecap": "round" });
  }));
  tf(burst, sx(6), RY);
  root.append(burst, token);

  const C = { x: 200, y: 250 };
  const tcard = h("g");
  const glow = h("rect", { x: C.x - 6, y: C.y - 6, width: 412, height: 222, rx: 18, fill: "none", stroke: "#F59E0B", "stroke-width": 2 });
  const pill = h("rect", { x: C.x + 108, y: C.y + 102, height: 24, rx: 12, "stroke-width": 1.2 });
  const pillDot = h("circle", { cx: C.x + 121, cy: C.y + 114, r: 4.5 });
  const pillText = mono(C.x + 132, C.y + 118, 11.5, "Backlog", "fill-foreground", { "font-weight": 500 });
  const prBars = [0, 1, 2, 3].map((i) =>
    h("rect", { x: C.x + 110 + i * 5, y: C.y + 152 - (i + 1) * 3.2, width: 3.2, height: (i + 1) * 3.2, rx: 0.8, class: "fill-muted-foreground/25" }));
  const prText = mono(C.x + 138, C.y + 152, 11.5, "No priority", "fill-foreground", { "font-weight": 500 });
  const av1 = avatar(C.x + 120, C.y + 178, 12, "AR", "sky");
  const av2 = avatar(C.x + 142, C.y + 178, 12, "KM", "violet");
  const follow = h("g", {}, [
    h("rect", { x: C.x + 270, y: C.y + 166, width: 112, height: 24, rx: 12, class: "fill-amber-500/10 stroke-amber-500/40", "stroke-width": 1 }),
    h("path", { d: `M${C.x + 281} ${C.y + 178} q8 -8 16 0 q-8 8 -16 0 z`, class: "fill-none stroke-amber-600 dark:stroke-amber-500", "stroke-width": 1.4 }),
    h("circle", { cx: C.x + 289, cy: C.y + 178, r: 2, class: "fill-amber-600 dark:fill-amber-500" }),
    mono(C.x + 303, C.y + 182, 11, "2 following", "fill-amber-700 dark:fill-amber-400"),
  ]);
  const stamp = h("g", {}, [
    h("rect", { x: -70, y: -24, width: 140, height: 48, rx: 6, fill: "none", class: "stroke-amber-600", "stroke-width": 3.5 }),
    h("rect", { x: -63, y: -17, width: 126, height: 34, rx: 3, fill: "none", class: "stroke-amber-600", "stroke-width": 1.2 }),
    mono(0, 7, 19, "APPROVED", "fill-amber-600", { "text-anchor": "middle", "font-weight": 700, "letter-spacing": 3 }),
  ]);
  tcard.append(
    glow,
    card(C.x, C.y, 400, 210, 14),
    h("circle", { cx: C.x, cy: C.y + 88, r: 7, class: "fill-background stroke-border" }),
    h("circle", { cx: C.x + 400, cy: C.y + 88, r: 7, class: "fill-background stroke-border" }),
    h("line", { x1: C.x + 14, y1: C.y + 88, x2: C.x + 386, y2: C.y + 88, class: "stroke-border", "stroke-dasharray": "4 5", "stroke-width": 1.2 }),
    mono(C.x + 24, C.y + 34, 11.5, "DEVOPS-780", "fill-muted-foreground", { "letter-spacing": 1 }),
    sans(C.x + 24, C.y + 64, 21, "Renew vendor contract", "fill-foreground", { "font-weight": 650, "letter-spacing": -0.4 }),
    mono(C.x + 24, C.y + 118, 10, "STATE", "fill-muted-foreground/80", { "letter-spacing": 1.5 }),
    mono(C.x + 24, C.y + 152, 10, "PRIORITY", "fill-muted-foreground/80", { "letter-spacing": 1.5 }),
    mono(C.x + 24, C.y + 182, 10, "ASSIGNEES", "fill-muted-foreground/80", { "letter-spacing": 1.5 }),
    pill, pillDot, pillText, ...prBars, prText, av2, av1, follow, stamp,
  );
  root.append(tcard);

  const keys: [number, number][] = [
    [0, 0], [0.08, 0], [0.16, 1], [0.24, 1], [0.32, 2], [0.55, 2],
    [0.63, 3], [0.7, 3], [0.8, 5], [0.86, 5], [0.95, 6], [1, 6],
  ];
  let lastState = -1;
  let lastPr = -1;
  return {
    root,
    update(p, t) {
      let pos = 0;
      let state = 0;
      for (let i = 0; i < keys.length - 1; i++) {
        const [pa, ia] = keys[i]!;
        const [pb, ib] = keys[i + 1]!;
        if (p >= pa && p <= pb) {
          const u = easeInOut(range(p, pa, pb));
          pos = lerp(ia, ib, u);
          state = u > 0.5 ? ib : ia;
          break;
        }
        if (p > pb) {
          pos = ib;
          state = ib;
        }
      }
      const s = STATES[state]!;
      const x = X0 + pos * DX;
      tf(token, x, RY);
      tokenFill.setAttribute("fill", s.color);
      trail.setAttribute("x2", f2(x));
      trail.setAttribute("stroke", s.color);
      nodes.forEach((d, i) => (d.style.fill = i !== 4 && sx(i) <= x + 1 ? STATES[i]!.color : ""));

      if (state !== lastState) {
        lastState = state;
        setText(pillText, s.name);
        pill.setAttribute("fill", `${s.color}1f`);
        pill.setAttribute("stroke", `${s.color}66`);
        pillDot.setAttribute("fill", s.color);
        pill.setAttribute("width", f2(36 + s.name.length * 7));
      }

      const waiting = range(p, 0.3, 0.33) * (1 - range(p, 0.55, 0.6));
      const ring = (t * 0.9) % 1;
      pulse.setAttribute("r", f2(12 + ring * 22));
      op(pulse, waiting * (1 - ring));
      op(glow, waiting * (0.35 + Math.sin(t * 4) * 0.25));

      const sp = range(p, 0.42, 0.47);
      const impact = range(p, 0.47, 0.53);
      const shake = Math.sin(impact * Math.PI * 5) * (1 - impact) * 4;
      shared.nod = Math.sin(impact * Math.PI) * 10;
      tf(stamp, C.x + 316, C.y + 128, lerp(2.2, 0.85, easeIn(sp)), -12);
      op(stamp, sp > 0 ? Math.min(1, sp * 3) * 0.92 : 0);
      tf(tcard, shake, impact > 0 && impact < 1 ? Math.abs(shake) * 0.5 : 0);

      const pr = Math.floor(range(p, 0.02, 0.28) * 3.99);
      if (pr !== lastPr) {
        lastPr = pr;
        const P = PRIORITIES[pr]!;
        setText(prText, P.name);
        prText.style.fill = pr ? P.color : "";
        prBars.forEach((b, i) => (b.style.fill = i < pr ? P.color : ""));
      }
      [av1, av2].forEach((a, i) => {
        const f = range(p, 0.08 + i * 0.05, 0.16 + i * 0.05);
        tf(a, C.x + 120 + i * 22, C.y + 178 - (1 - easeBack(f)) * 30, 1);
        op(a, f * 2);
      });
      op(follow, range(p, 0.2, 0.27));

      const bf = range(p, 0.95, 1);
      burst.setAttribute("transform", `translate(${sx(6)} ${RY}) scale(${f2(0.6 + bf * 0.8)})`);
      op(burst, bf > 0 && bf < 1 ? 1 - bf : 0);
    },
  };
}

// ─── § 04 · Subtasks, the board, cycles and modules ───
function scenePlanning(): Scene {
  const root = h("g");
  const parent = h("g", {}, [
    card(250, 40, 300, 56, 12),
    mono(268, 63, 11, "DEVOPS-780", "fill-muted-foreground"),
    sans(268, 83, 15, "Renew vendor contract", "fill-foreground", { "font-weight": 600 }),
    h("rect", { x: 462, y: 56, width: 74, height: 22, rx: 11, class: "fill-amber-500/12 stroke-amber-500/40", "stroke-width": 1 }),
    mono(499, 71, 10.5, "3 subtasks", "fill-amber-700 dark:fill-amber-400", { "text-anchor": "middle" }),
  ]);
  root.append(parent);

  const cols = [
    { name: "Todo", color: "#3B82F6", x: 70 },
    { name: "In Progress", color: "#10B981", x: 305 },
    { name: "Done", color: "#22C55E", x: 540 },
  ];
  const board = h("g");
  cols.forEach((c) => {
    board.append(
      h("rect", { x: c.x, y: 220, width: 190, height: 240, rx: 12, class: "fill-muted/50 stroke-border", "stroke-width": 1 }),
      h("circle", { cx: c.x + 18, cy: 241, r: 5, fill: c.color }),
      mono(c.x + 30, 245, 11.5, c.name, "fill-foreground", { "font-weight": 600 }),
    );
  });
  const fillers = [
    { col: 0, slot: 1, id: "DEVOPS-776", title: "Insurance renewal", label: "#0EA5E9" },
    { col: 2, slot: 1, id: "DEVOPS-771", title: "Access review", label: "#A855F7" },
    { col: 1, slot: 1, id: "DEVOPS-774", title: "Budget sign-off", label: "#F43F5E" },
  ].map((f) => {
    const x = cols[f.col]!.x + 8;
    const y = 262 + f.slot * 52;
    const g = h("g", {}, [
      card(x, y, 174, 44, 8),
      mono(x + 10, y + 17, 9.5, f.id, "fill-muted-foreground"),
      sans(x + 10, y + 34, 12.5, f.title, "fill-foreground", { "font-weight": 550 }),
      h("circle", { cx: x + 160, cy: y + 14, r: 4, fill: f.label }),
    ]);
    board.append(g);
    return g;
  });
  root.append(board);

  const subs = [
    { id: "DEVOPS-781", title: "Collect quotes", col: 2, label: "#10B981" },
    { id: "DEVOPS-782", title: "Legal review", col: 1, label: "#A855F7" },
    { id: "DEVOPS-783", title: "Sign & file", col: 0, label: "#F59E0B" },
  ];
  const tree = subs.map((_, k) => {
    const y = 132 + k * 46;
    const path = h("path", { d: `M270 96 V${y} H296`, pathLength: 1, "stroke-dasharray": "1 1", class: "fill-none stroke-amber-500/70", "stroke-width": 1.5 });
    root.append(path);
    return path;
  });
  const subEls = subs.map((s) => {
    const rect = card(0, 0, 220, 40, 8);
    const g = h("g", {}, [
      rect,
      mono(10, 16, 9.5, s.id, "fill-muted-foreground"),
      sans(10, 32, 12.5, s.title, "fill-foreground", { "font-weight": 550 }),
    ]);
    const dot = h("circle", { cy: 14, r: 4, fill: s.label });
    g.append(dot);
    root.append(g);
    return { g, rect, dot };
  });

  const cycle = h("g", {}, [
    mono(70, 500, 11, "CYCLE · Sprint 14", "fill-foreground", { "font-weight": 600, "letter-spacing": 1 }),
    mono(730, 500, 11, "Oct 06 → Oct 19", "fill-muted-foreground", { "text-anchor": "end" }),
    h("rect", { x: 70, y: 510, width: 660, height: 10, rx: 5, class: "fill-muted" }),
    ...Array.from({ length: 13 }, (_, i) =>
      h("line", { x1: 70 + ((i + 1) * 660) / 14, y1: 524, x2: 70 + ((i + 1) * 660) / 14, y2: 529, class: "stroke-muted-foreground/40", "stroke-width": 1 })),
  ]);
  const cycleFill = h("rect", { x: 70, y: 510, width: 0, height: 10, rx: 5, class: "fill-amber-500" });
  const cycleMark = h("path", { d: "M0 0 l-5 -8 h10 z", class: "fill-amber-600" });
  cycle.append(cycleFill, cycleMark);
  const mods = [
    { x: 70, name: "Vendor onboarding", status: "in_progress", c: "#10B981" },
    { x: 410, name: "Compliance", status: "planned", c: "#3B82F6" },
  ].map((m) => {
    const g = h("g", {}, [
      h("rect", { x: m.x, y: 548, width: 320, height: 34, rx: 8, class: "fill-card stroke-border", "stroke-width": 1 }),
      h("rect", { x: m.x + 10, y: 557, width: 16, height: 16, rx: 4, fill: "none", class: "stroke-foreground/60", "stroke-width": 1.4 }),
      h("rect", { x: m.x + 14, y: 561, width: 8, height: 8, rx: 1.5, class: "fill-foreground/60" }),
      mono(m.x + 36, 569, 10, "MODULE", "fill-muted-foreground", { "letter-spacing": 1.2 }),
      sans(m.x + 86, 570, 13, m.name, "fill-foreground", { "font-weight": 600 }),
      h("rect", { x: m.x + 222, y: 556, width: 88, height: 18, rx: 9, fill: `${m.c}1f`, stroke: `${m.c}66` }),
      mono(m.x + 266, 568.5, 10, m.status, "fill-foreground", { "text-anchor": "middle" }),
    ]);
    root.append(g);
    return g;
  });
  root.append(cycle);

  return {
    root,
    update(p) {
      const pf = range(p, 0, 0.1);
      tf(parent, 0, (1 - easeOut(pf)) * -20);
      op(parent, pf);
      const fly = easeInOut(range(p, 0.34, 0.56));
      subEls.forEach(({ g, rect, dot }, k) => {
        const appear = range(p, 0.1 + k * 0.06, 0.22 + k * 0.06);
        draw(tree[k]!, appear);
        op(tree[k]!, 1 - range(p, 0.34, 0.42));
        const fx = 296;
        const fy = 112 + k * 46;
        const c = cols[subs[k]!.col]!;
        const tx = c.x + 8;
        const ty = 262;
        const arc = Math.sin(fly * Math.PI) * -40;
        tf(g, lerp(fx, tx, fly) + (1 - easeOut(appear)) * 20, lerp(fy, ty, fly) + arc, 1, Math.sin(fly * Math.PI) * (k - 1) * 6);
        op(g, appear);
        const w = lerp(220, 174, fly);
        rect.setAttribute("width", f2(w));
        dot.setAttribute("cx", f2(w - 14));
      });
      const bf = range(p, 0.3, 0.42);
      op(board, bf);
      tf(board, 0, (1 - easeOut(bf)) * 16);
      fillers.forEach((g, i) => op(g, range(p, 0.4 + i * 0.04, 0.5 + i * 0.04)));

      const cf = range(p, 0.55, 0.62);
      op(cycle, cf);
      const fill = 660 * 0.64 * easeOut(range(p, 0.58, 0.82));
      cycleFill.setAttribute("width", f2(fill));
      tf(cycleMark, 70 + fill, 508);
      mods.forEach((m, i) => {
        const f = range(p, 0.66 + i * 0.06, 0.78 + i * 0.06);
        tf(m, 0, (1 - easeBack(f)) * 20);
        op(m, f);
      });
    },
  };
}

// ─── § 05 · Graph view: people and tasks, with subtask edges ───
function sceneGraph(): Scene {
  const root = h("g");
  const cam = h("g");
  root.append(cam);
  type N = { id: string; x: number; y: number; kind: "user" | "task"; tone?: Tone; el?: SVGElement };
  const nodes: N[] = [
    { id: "DEVOPS-780", x: 400, y: 300, kind: "task" },
    { id: "DEVOPS-781", x: 280, y: 210, kind: "task" },
    { id: "DEVOPS-782", x: 525, y: 205, kind: "task" },
    { id: "DEVOPS-783", x: 405, y: 430, kind: "task" },
    { id: "ACCT-91", x: 125, y: 315, kind: "task" },
    { id: "ACCT-92", x: 680, y: 330, kind: "task" },
    { id: "AR", x: 185, y: 130, kind: "user", tone: "sky" },
    { id: "KM", x: 640, y: 115, kind: "user", tone: "violet" },
    { id: "SN", x: 205, y: 470, kind: "user", tone: "emerald" },
    { id: "DP", x: 620, y: 485, kind: "user", tone: "rose" },
  ];
  const edges: [string, string, boolean][] = [
    ["DEVOPS-780", "DEVOPS-781", true],
    ["DEVOPS-780", "DEVOPS-782", true],
    ["DEVOPS-780", "DEVOPS-783", true],
    ["AR", "DEVOPS-781", false],
    ["AR", "ACCT-91", false],
    ["KM", "DEVOPS-782", false],
    ["KM", "ACCT-92", false],
    ["SN", "DEVOPS-783", false],
    ["SN", "ACCT-91", false],
    ["DP", "DEVOPS-780", false],
    ["DP", "ACCT-92", false],
  ];
  const edgeEls = edges.map(([, , sub]) => {
    const l = h("line", sub
      ? { class: "stroke-amber-500", "stroke-width": 2, "stroke-dasharray": "5 5", "stroke-linecap": "round" }
      : { class: "stroke-foreground/25", "stroke-width": 1.4 });
    cam.append(l);
    return l;
  });
  for (const n of nodes) {
    if (n.kind === "user") {
      n.el = h("g", {}, [
        h("circle", { r: 26, class: "fill-none stroke-amber-500/40", "stroke-width": 1.2, "stroke-dasharray": "2 4" }),
        avatar(0, 0, 20, n.id, n.tone!),
      ]);
    } else {
      const w = n.id.length * 7.4 + 24;
      const proj = n.id.split("-")[0];
      n.el = h("g", {}, [
        h("rect", { x: -w / 2, y: -16, width: w, height: 32, rx: 9, class: "fill-card stroke-border", "stroke-width": 1.2 }),
        h("rect", { x: -w / 2, y: -16, width: 4, height: 32, rx: 2, class: proj === "DEVOPS" ? "fill-amber-500" : "fill-sky-500" }),
        mono(2, 4.5, 12, n.id, "fill-foreground", { "text-anchor": "middle", "font-weight": 500 }),
      ]);
    }
    cam.append(n.el!);
  }
  const legend = h("g", {}, [
    h("circle", { cx: 238, cy: 598, r: 6, class: "fill-sky-100 stroke-sky-800/40" }),
    mono(250, 602, 11, "person", "fill-muted-foreground"),
    h("rect", { x: 318, y: 591, width: 18, height: 14, rx: 4, class: "fill-card stroke-border" }),
    mono(342, 602, 11, "task", "fill-muted-foreground"),
    h("line", { x1: 396, y1: 598, x2: 420, y2: 598, class: "stroke-foreground/40", "stroke-width": 1.4 }),
    mono(426, 602, 11, "assigned", "fill-muted-foreground"),
    h("line", { x1: 504, y1: 598, x2: 528, y2: 598, class: "stroke-amber-500", "stroke-width": 2, "stroke-dasharray": "5 4" }),
    mono(534, 602, 11, "subtask", "fill-muted-foreground"),
  ]);
  root.append(legend);
  const pos = new Map<string, { x: number; y: number }>();

  return {
    root,
    update(p, t) {
      about(cam, 400, 300, lerp(0.88, 1, easeOut(range(p, 0, 0.7))));
      nodes.forEach((n, k) => {
        const f = range(p, 0.04 + k * 0.035, 0.42 + k * 0.035);
        const s = spring(f);
        const a = k * 2.399;
        const sx = 400 + Math.cos(a) * 30;
        const sy = 300 + Math.sin(a) * 30;
        const x = lerp(sx, n.x, s) + Math.sin(t * 0.7 + k) * 3;
        const y = lerp(sy, n.y, s) + Math.cos(t * 0.6 + k * 1.3) * 3;
        pos.set(n.id, { x, y });
        tf(n.el!, x, y, lerp(0.3, 1, easeOut(f)));
        op(n.el!, f * 3);
      });
      edges.forEach(([a, b], i) => {
        const A = pos.get(a)!;
        const B = pos.get(b)!;
        const d = easeOut(range(p, 0.42 + i * 0.025, 0.6 + i * 0.025));
        const l = edgeEls[i]!;
        l.setAttribute("x1", f2(A.x));
        l.setAttribute("y1", f2(A.y));
        l.setAttribute("x2", f2(lerp(A.x, B.x, d)));
        l.setAttribute("y2", f2(lerp(A.y, B.y, d)));
        op(l, d > 0 ? 1 : 0);
      });
      op(legend, range(p, 0.7, 0.82));
    },
  };
}

// ─── § 06 · Comments, 15-minute batching, email digest, Mattermost ───
function sceneNotify(): Scene {
  const root = h("g");
  const bubble = h("g", {}, [
    card(60, 50, 410, 190, 16),
    avatar(96, 88, 18, "SN", "emerald"),
    sans(124, 85, 13.5, "S. Nair", "fill-foreground", { "font-weight": 600 }),
    mono(124, 102, 10.5, "commented on DEVOPS-780", "fill-muted-foreground"),
    sans(84, 138, 15, "Quotes are in, see attached.", "fill-foreground"),
    h("rect", { x: 82, y: 148, width: 62, height: 23, rx: 6, class: "fill-amber-500/15" }),
    sans(88, 165, 15, "@Kiran", "fill-amber-700 dark:fill-amber-400", { "font-weight": 600 }),
    sans(150, 165, 15, "can you sign off?", "fill-foreground"),
    h("rect", { x: 84, y: 182, width: 72, height: 44, rx: 6, class: "fill-sky-100 dark:fill-sky-950 stroke-border", "stroke-width": 1 }),
    h("circle", { cx: 140, cy: 194, r: 5, class: "fill-amber-400" }),
    h("path", { d: "M86 224 l20 -20 l14 12 l10 -8 l24 16 z", class: "fill-sky-300 dark:fill-sky-800" }),
    mono(170, 200, 11.5, "quote-v2.png", "fill-foreground"),
    mono(170, 217, 10, "248 KB · click to preview", "fill-muted-foreground"),
  ]);
  root.append(bubble);

  const bell = h("g", {}, [
    h("path", { d: "M-14 8 v-10 a14 14 0 0 1 28 0 v10 l4 5 h-36 z", class: "fill-card stroke-foreground", "stroke-width": 2, "stroke-linejoin": "round" }),
    h("path", { d: "M-5 16 a5 5 0 0 0 10 0", class: "fill-none stroke-foreground", "stroke-width": 2 }),
  ]);
  const badge = h("g", {}, [h("circle", { cx: 13, cy: -14, r: 8, class: "fill-amber-500" }), mono(13, -10.5, 10, "1", "fill-amber-950", { "text-anchor": "middle", "font-weight": 700 })]);
  bell.append(badge);
  root.append(bell);

  const note = h("g", {}, [
    card(500, 150, 240, 64, 12),
    h("circle", { cx: 518, cy: 170, r: 4, class: "fill-amber-500" }),
    mono(530, 174, 11, "DEVOPS-780", "fill-muted-foreground"),
    sans(518, 198, 13.5, "Renew vendor contract", "fill-foreground", { "font-weight": 600 }),
    h("rect", { x: 688, y: 160, width: 42, height: 22, rx: 11, class: "fill-amber-500" }),
  ]);
  const count = mono(709, 175.5, 11.5, "×1", "fill-amber-950", { "text-anchor": "middle", "font-weight": 700 });
  note.append(count);
  root.append(note);

  const WX0 = 100;
  const WX1 = 700;
  const win = h("rect", { x: WX0, y: 300, width: WX1 - WX0, height: 56, rx: 12, class: "fill-amber-500/5 stroke-amber-500/50", "stroke-dasharray": "5 5", "stroke-width": 1.4 });
  const winLabel = mono(WX0, 288, 11, "ONE 15-MINUTE WINDOW", "fill-muted-foreground", { "letter-spacing": 1.5 });
  const ticks = h("g", {}, ["0m", "5m", "10m", "15m"].map((l, i) =>
    mono(WX0 + (i * (WX1 - WX0)) / 3, 376, 10, l, "fill-muted-foreground/80", { "text-anchor": i === 0 ? "start" : i === 3 ? "end" : "middle" })));
  const head = h("line", { y1: 296, y2: 360, class: "stroke-amber-600", "stroke-width": 2 });
  const closed = mono(WX1, 288, 11, "CLOSED → SEND", "fill-amber-700 dark:fill-amber-400", { "text-anchor": "end", "letter-spacing": 1.5, "font-weight": 700 });
  root.append(win, winLabel, ticks, head, closed);

  const events = [
    { x: 175, label: "comment" },
    { x: 310, label: "state" },
    { x: 450, label: "assignee" },
    { x: 585, label: "label" },
  ].map((e) => {
    const w = e.label.length * 7 + 20;
    const chip = h("g", {}, [
      h("rect", { x: -w / 2, y: -12, width: w, height: 24, rx: 12, class: "fill-card stroke-border", "stroke-width": 1.2 }),
      mono(0, 4, 11, e.label, "fill-foreground", { "text-anchor": "middle" }),
    ]);
    const spark = h("circle", { r: 4.5, class: "fill-amber-500" });
    root.append(chip, spark);
    return { ...e, chip, spark };
  });

  const env = h("g", {}, [
    h("rect", { x: -48, y: -32, width: 96, height: 64, rx: 8, class: "fill-card stroke-foreground/70", "stroke-width": 1.8 }),
    h("path", { d: "M-48 -26 L0 8 L48 -26", class: "fill-none stroke-foreground/70", "stroke-width": 1.8, "stroke-linejoin": "round" }),
    mono(0, 56, 10.5, "EMAIL DIGEST · OPT-IN", "fill-muted-foreground", { "text-anchor": "middle", "letter-spacing": 1 }),
  ]);
  const dm = h("g", {}, [
    h("path", { d: "M-92 -30 h184 a10 10 0 0 1 10 10 v36 a10 10 0 0 1 -10 10 h-160 l-14 12 v-12 h-10 a10 10 0 0 1 -10 -10 v-36 a10 10 0 0 1 10 -10 z", class: "fill-card stroke-foreground/70", "stroke-width": 1.8 }),
    mono(-88, -10, 10.5, "bureaucat bot", "fill-muted-foreground"),
    sans(-88, 10, 13, "4 updates on DEVOPS-780", "fill-foreground", { "font-weight": 600 }),
    mono(0, 56, 10.5, "MATTERMOST DM", "fill-muted-foreground", { "text-anchor": "middle", "letter-spacing": 1 }),
  ]);
  root.append(env, dm);

  let lastN = -1;
  return {
    root,
    update(p, t) {
      const bf = range(p, 0, 0.1);
      tf(bubble, 0, (1 - easeOut(bf)) * 20);
      op(bubble, bf);
      const wf = range(p, 0.06, 0.14);
      op(win, wf);
      op(winLabel, wf);
      op(ticks, wf);
      const hx = lerp(WX0, WX1, range(p, 0.12, 0.62));
      head.setAttribute("x1", f2(hx));
      head.setAttribute("x2", f2(hx));
      op(head, wf * (1 - range(p, 0.62, 0.66)));
      const shut = range(p, 0.62, 0.68);
      op(closed, shut);
      win.setAttribute("stroke-dasharray", shut > 0.5 ? "0" : "5 5");

      let landed = 0;
      events.forEach((e) => {
        const appear = hx >= e.x ? 1 : 0;
        const ap = range(hx, e.x - 10, e.x + 30);
        tf(e.chip, e.x, 328 - (1 - easeBack(ap)) * 18);
        const tStart = 0.12 + ((e.x - WX0) / (WX1 - WX0)) * 0.5;
        const f = range(p, tStart + 0.02, tStart + 0.08);
        const merge = easeInOut(range(p, 0.62, 0.7));
        const sx = lerp(e.x, 709, f);
        const sy = lerp(316, 171, f) - Math.sin(f * Math.PI) * 60;
        tf(e.spark, sx, sy, 1 - f * 0.4);
        op(e.spark, f > 0 && f < 1 ? 1 : 0);
        if (f >= 1) landed++;
        op(e.chip, (appear ? ap : 0) * (1 - merge * 0.6));
      });
      const nf = range(p, 0.14, 0.2);
      tf(note, 0, (1 - easeOut(nf)) * -14);
      op(note, nf);
      const n = Math.max(1, landed);
      if (n !== lastN) {
        lastN = n;
        setText(count, `×${n}`);
      }

      const ring = range(p, 0.64, 0.8);
      const swing = ring > 0 && ring < 1 ? Math.sin(t * 22) * 14 * (1 - ring) : 0;
      tf(bell, 680, 92, 1.2, swing);
      op(bell, range(p, 0.02, 0.1));
      const bb = range(p, 0.2, 0.26);
      tf(badge, 0, 0, easeBack(bb));
      op(badge, bb);

      const out = range(p, 0.68, 0.88);
      [
        { el: env, tx: 200, ty: 520, delay: 0 },
        { el: dm, tx: 590, ty: 520, delay: 0.05 },
      ].forEach(({ el, tx, ty, delay }) => {
        const f = easeOut(range(out, delay, 0.75 + delay));
        const x = lerp(620, tx, f);
        const y = lerp(190, ty, f) - Math.sin(f * Math.PI) * 50;
        tf(el, x, y, lerp(0.3, 1, f), (1 - f) * (tx < 400 ? -20 : 20));
        op(el, f * 2);
      });
    },
  };
}

// ─── § 07 · Append-only, hash-chained activity log ───
function sceneAudit(): Scene {
  const root = h("g");
  const rows = [
    { type: "task_created", detail: "DEVOPS-780 · Renew vendor contract", hash: "3f9a…c21e", color: "#6B7280" },
    { type: "assignee_added", detail: "+ Kiran", hash: "b71d…04a9", color: "#3B82F6" },
    { type: "state_changed", detail: "Todo → Approval Pending", hash: "e02c…9f3b", color: "#F59E0B" },
    { type: "comment_created", detail: "“Signed off.”", hash: "5c48…ad17", color: "#8B5CF6" },
    { type: "state_changed", detail: "Approval Pending → In Progress", hash: "91fe…63d0", color: "#10B981" },
  ];
  const X = 110;
  const RW = 580;
  const y0 = 56;
  const gap = 88;
  const links: SVGGElement[] = [];
  const els = rows.map((r, k) => {
    const y = y0 + k * gap;
    const check = h("circle", { cx: X + RW - 26, cy: y + 32, r: 11, "stroke-width": 1.5, class: "fill-none stroke-border" });
    const tick = h("path", { d: `M${X + RW - 31} ${y + 32} l4 4 l7 -8`, fill: "none", stroke: "white", "stroke-width": 2, "stroke-linecap": "round", "stroke-linejoin": "round" });
    const g = h("g", {}, [
      card(X, y, RW, 64, 12),
      h("circle", { cx: X + 26, cy: y + 32, r: 7, fill: r.color }),
      mono(X + 46, y + 27, 12.5, r.type, "fill-foreground", { "font-weight": 600 }),
      sans(X + 46, y + 47, 12.5, r.detail, "fill-muted-foreground"),
      mono(X + RW - 50, y + 27, 11, `sha256 ${r.hash}`, "fill-foreground/80", { "text-anchor": "end" }),
      mono(X + RW - 50, y + 46, 10, k ? `prev ${rows[k - 1]!.hash}` : "prev ∅ genesis", "fill-muted-foreground/80", { "text-anchor": "end" }),
      check,
      tick,
    ]);
    root.append(g);
    if (k) {
      const ly = y - 24;
      const link = h("g", {}, [
        h("rect", { x: X + 21, y: ly - 6, width: 10, height: 18, rx: 5, fill: "none", "stroke-width": 2 }),
        h("rect", { x: X + 21, y: ly + 6, width: 10, height: 18, rx: 5, fill: "none", "stroke-width": 2 }),
      ]);
      root.append(link);
      links.push(link);
    }
    return { g, check, tick, y };
  });
  const scan = h("g", {}, [
    h("rect", { x: X - 20, y: -18, width: RW + 40, height: 36, class: "fill-emerald-500/10" }),
    h("line", { x1: X - 20, x2: X + RW + 20, y1: 0, y2: 0, stroke: "#22C55E", "stroke-width": 2 }),
  ]);
  const banner = h("g", {}, [
    h("rect", { x: 230, y: 510, width: 340, height: 40, rx: 20, fill: "#22C55E1f", stroke: "#22C55E80", "stroke-width": 1.2 }),
    mono(400, 535, 12.5, "✓ chain intact · 5 / 5 verified", "fill-foreground", { "text-anchor": "middle", "font-weight": 600, "letter-spacing": 0.5 }),
  ]);
  root.append(scan, banner);

  return {
    root,
    update(p) {
      const sy = lerp(y0 - 10, y0 + 4 * gap + 74, easeInOut(range(p, 0.55, 0.85)));
      els.forEach((e, k) => {
        const f = range(p, k * 0.09, k * 0.09 + 0.14);
        tf(e.g, (1 - easeOut(f)) * 50, 0);
        op(e.g, f);
        const ok = p > 0.55 && sy > e.y + 32;
        e.check.style.fill = ok ? "#22C55E" : "";
        e.check.style.stroke = ok ? "#22C55E" : "";
        op(e.tick, ok ? 1 : 0);
        if (k) {
          const link = links[k - 1]!;
          op(link, range(p, k * 0.09 + 0.08, k * 0.09 + 0.16));
          link.setAttribute("stroke", ok ? "#22C55E" : "#F59E0B");
        }
      });
      const sv = range(p, 0.55, 0.58) * (1 - range(p, 0.85, 0.88));
      tf(scan, 0, sy);
      op(scan, sv);
      const bf = range(p, 0.86, 0.92);
      tf(banner, 0, (1 - easeBack(bf)) * 16);
      op(banner, bf);
    },
  };
}

// ─── The cat: built from BurecatLogo geometry, lives across every chapter ───
interface Cat {
  root: SVGGElement;
  update(o: { x: number; y: number; s: number; r: number; build: number; sleep: number; look: { x: number; y: number } | null; t: number; still: boolean }): void;
}

function buildCat(): Cat {
  const root = h("g");
  const body = h("g");
  body.setAttribute("transform", "translate(-16 -17)");
  root.append(body);
  const earL = h("g", {}, [h("path", { d: "M5.5 14L2.5 3L13 10Z", class: "fill-amber-500" }), h("path", { d: "M7 12L4.5 5.5L11.5 9.5Z", class: "fill-amber-600" })]);
  const earR = h("g", {}, [h("path", { d: "M26.5 14L29.5 3L19 10Z", class: "fill-amber-500" }), h("path", { d: "M25 12L27.5 5.5L20.5 9.5Z", class: "fill-amber-600" })]);
  const head = h("circle", { cx: 16, cy: 18, r: 11, class: "fill-amber-500" });
  const glass = (cx: number) => h("circle", { cx, cy: 16.5, r: 3.8, pathLength: 1, "stroke-dasharray": "1 1", class: "stroke-zinc-950", "stroke-width": 1, fill: "none" });
  const gL = glass(11.5);
  const gR = glass(20.5);
  const bridge = h("path", { d: "M15.3 16C15.5 15.2 16.5 15.2 16.7 16", pathLength: 1, "stroke-dasharray": "1 1", class: "stroke-zinc-950", "stroke-width": 0.8, fill: "none" });
  const eye = (cx: number) => {
    const look = h("g", {}, [h("circle", { cx: cx + 0.3, cy: 16.2, r: 0.9, class: "fill-amber-500" }), h("circle", { cx: cx + 0.8, cy: 15.7, r: 0.45, class: "fill-zinc-50" })]);
    const g = h("g", {}, [h("ellipse", { cx, cy: 16.8, rx: 1.8, ry: 2, class: "fill-zinc-950" }), look]);
    return { g, look, cx };
  };
  const eL = eye(11.5);
  const eR = eye(20.5);
  const shut = h("g", {}, [
    h("path", { d: "M9.8 17.2 q1.7 -1.6 3.4 0", class: "stroke-zinc-950", "stroke-width": 0.8, fill: "none", "stroke-linecap": "round" }),
    h("path", { d: "M18.8 17.2 q1.7 -1.6 3.4 0", class: "stroke-zinc-950", "stroke-width": 0.8, fill: "none", "stroke-linecap": "round" }),
  ]);
  const face = h("g", {}, [
    h("path", { d: "M14.8 21L16 20L17.2 21Z", class: "fill-amber-700" }),
    h("path", { d: "M14 22.5C14.8 23.2 17.2 23.2 18 22.5", class: "stroke-amber-800", "stroke-width": 0.7, "stroke-linecap": "round", fill: "none" }),
  ]);
  const wh = (x1: number, y1: number, x2: number, y2: number) =>
    h("line", { x1, y1, x2, y2, pathLength: 1, "stroke-dasharray": "1 1", class: "stroke-amber-600", "stroke-width": 0.5, "stroke-linecap": "round" });
  const whL = h("g", {}, [wh(7.5, 19, 2, 17.5), wh(7.5, 20.5, 1.5, 20.5)]);
  const whR = h("g", {}, [wh(24.5, 19, 30, 17.5), wh(24.5, 20.5, 30.5, 20.5)]);
  const whiskers = [...whL.children, ...whR.children] as SVGElement[];
  const zz = [0, 1, 2].map(() => mono(0, 0, 5, "z", "fill-muted-foreground", { "font-weight": 700 }));
  body.append(earL, earR, head, whL, whR, eL.g, eR.g, shut, gL, gR, bridge, face, ...zz);

  const gaze = { x: 0, y: 0 };
  return {
    root,
    update({ x, y, s, r, build, sleep, look, t, still }) {
      tf(root, x, y, s, r);
      const hb = easeBack(range(build, 0, 0.4));
      about(head, 16, 18, hb);
      const twitch = still ? 0 : Math.max(0, Math.sin(t * 1.3) - 0.97) * 300;
      const eb = easeBack(range(build, 0.25, 0.6));
      about(earL, 8, 12, eb, eb, -(1 - eb) * 30);
      about(earR, 24, 12, eb, eb, (1 - eb) * 30 + twitch);
      const gb = easeInOut(range(build, 0.45, 0.8));
      draw(gL, gb);
      draw(gR, gb);
      draw(bridge, range(build, 0.7, 0.85));
      const phase = still ? 1 : (t % 3.7) / 0.16;
      const blink = phase < 1 ? 1 - Math.sin(phase * Math.PI) : 1;
      const open = range(build, 0.55, 0.75) * blink * (1 - sleep);
      about(eL.g, eL.cx, 16.8, 1, open);
      about(eR.g, eR.cx, 16.8, 1, open);
      op(shut, sleep);
      op(face, range(build, 0.7, 0.9));
      whiskers.forEach((w) => draw(w, range(build, 0.75, 1)));
      const wig = still ? 0 : Math.sin(t * 2.1) * 2;
      about(whL, 7.5, 20, 1, 1, wig);
      about(whR, 24.5, 20, 1, 1, -wig);

      let tx = still ? 0 : Math.sin(t * 0.5) * 0.4;
      let ty = still ? 0 : Math.cos(t * 0.37) * 0.25;
      if (look && !still) {
        const ex = x;
        const ey = y - 0.2 * s;
        const dx = look.x - ex;
        const dy = look.y - ey;
        const d = Math.hypot(dx, dy) || 1;
        const m = Math.min(d / 220, 1) * 0.65;
        tx = (dx / d) * m;
        ty = (dy / d) * m;
      }
      gaze.x = lerp(gaze.x, tx, 0.15);
      gaze.y = lerp(gaze.y, ty, 0.15);
      tf(eL.look, gaze.x, gaze.y);
      tf(eR.look, gaze.x, gaze.y);

      zz.forEach((z, i) => {
        const k = still ? 0.2 + i * 0.3 : (t * 0.35 + i / 3) % 1;
        tf(z, 27 + k * 6 + i * 1.5, 6 - k * 10, 0.7 + k * 0.6, -12);
        op(z, sleep * Math.sin(k * Math.PI));
      });
    },
  };
}

// [x, y, scale, rotation] for the cat in each chapter.
const CAT_POSES: [number, number, number, number][] = [
  [400, 290, 8.5, 0],
  [698, 552, 2.6, -8],
  [690, 520, 3, 0],
  [728, 76, 2.2, 6],
  [86, 562, 2.4, 0],
  [400, 588, 2.2, 0],
  [698, 566, 2.8, -4],
];

export interface StageOptions {
  svg: SVGSVGElement;
  section: HTMLElement;
  topOffset: () => number;
  reducedMotion: boolean;
  onChapter: (i: number) => void;
}

export function mountLandingStage(o: StageOptions) {
  const { svg, section } = o;
  svg.setAttribute("viewBox", `0 0 ${W} ${H}`);
  const shared: Shared = { nod: 0 };
  const scenes = [sceneIntake(), sceneHierarchy(), sceneLifecycle(shared), scenePlanning(), sceneGraph(), sceneNotify(), sceneAudit()];
  const layer = h("g");
  scenes.forEach((s) => layer.append(s.root));
  const cat = buildCat();
  svg.append(layer, cat.root);

  let target = -1;
  let current = -1;
  let chapter = -1;
  let raf = 0;
  let visible = false;
  let look: { x: number; y: number } | null = null;
  const t0 = performance.now();

  const measure = () => {
    const r = section.getBoundingClientRect();
    target = (o.topOffset() - r.top) / (CHAPTER_VH * window.innerHeight);
  };

  const render = () => {
    const t = o.reducedMotion ? 0 : (performance.now() - t0) / 1000;
    current = o.reducedMotion ? target : lerp(current, target, 0.12);
    if (Math.abs(current - target) < 0.0005) current = target;
    const P = current;
    const active = clamp(Math.floor(P + 0.08), 0, CHAPTERS - 1);
    if (active !== chapter) {
      chapter = active;
      o.onChapter(active);
    }

    shared.nod = 0;
    scenes.forEach((s, i) => {
      const local = P - i;
      let vis: number;
      let p: number;
      if (o.reducedMotion) {
        vis = i === active ? 1 : 0;
        p = 1;
      } else {
        const fin = i === 0 ? 1 : easeOut(range(local, -0.2, 0));
        const fout = i === CHAPTERS - 1 ? 1 : 1 - range(local, 0.85, 1.02);
        vis = fin * fout;
        p = clamp(local / 0.8);
      }
      s.root.style.display = vis <= 0.001 ? "none" : "";
      if (vis <= 0.001) return;
      const zoom = local < 0.5 ? lerp(0.96, 1, vis) : lerp(1.04, 1, vis);
      about(s.root, 400, 320, i === 0 && local < 0.5 ? 1 : zoom);
      op(s.root, vis);
      s.update(p, t);
    });

    let pose: [number, number, number, number];
    if (o.reducedMotion) {
      pose = CAT_POSES[active]!;
    } else {
      const seg = clamp(Math.floor(P - 0.82), 0, CHAPTERS - 2);
      const u = easeInOut(range(P, seg + 0.82, seg + 1.02));
      const a = CAT_POSES[P < 0.82 ? 0 : seg]!;
      const b = CAT_POSES[P < 0.82 ? 0 : seg + 1]!;
      pose = [lerp(a[0], b[0], u), lerp(a[1], b[1], u), lerp(a[2], b[2], u), lerp(a[3], b[3], u)];
    }
    const bob = o.reducedMotion ? 0 : Math.sin(t * 1.6) * 0.6 * pose[2] * 0.4;
    cat.update({
      x: pose[0],
      y: pose[1] + bob,
      s: pose[2],
      r: pose[3] + shared.nod,
      build: o.reducedMotion ? 1 : range(P, -0.95, -0.2),
      sleep: o.reducedMotion ? (active === CHAPTERS - 1 ? 1 : 0) : range(P, 6.62, 6.82),
      look,
      t,
      still: o.reducedMotion,
    });
  };

  const loop = () => {
    render();
    raf = visible ? requestAnimationFrame(loop) : 0;
  };
  const start = () => {
    if (!raf) raf = requestAnimationFrame(loop);
  };

  const onScroll = () => {
    measure();
    if (o.reducedMotion) render();
  };
  const onPointer = (e: PointerEvent) => {
    const m = svg.getScreenCTM();
    if (!m) return;
    const pt = new DOMPoint(e.clientX, e.clientY).matrixTransform(m.inverse());
    look = { x: pt.x, y: pt.y };
  };

  const io = new IntersectionObserver(([entry]) => {
    visible = !!entry?.isIntersecting;
    if (visible && !o.reducedMotion) start();
    if (visible && o.reducedMotion) render();
  });
  io.observe(section);
  window.addEventListener("scroll", onScroll, { passive: true });
  window.addEventListener("resize", onScroll, { passive: true });
  window.addEventListener("pointermove", onPointer, { passive: true });
  measure();
  current = target;
  render();

  return {
    destroy() {
      cancelAnimationFrame(raf);
      io.disconnect();
      window.removeEventListener("scroll", onScroll);
      window.removeEventListener("resize", onScroll);
      window.removeEventListener("pointermove", onPointer);
      svg.replaceChildren();
    },
  };
}

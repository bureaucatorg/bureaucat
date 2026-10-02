import { readFile, writeFile } from "node:fs/promises";
import { render } from "../dist-ssr/entry-server.js";

const file = new URL("../dist/index.html", import.meta.url);
const template = await readFile(file, "utf8");
const marker = '<div id="app"></div>';
if (!template.includes(marker)) throw new Error(`prerender: ${marker} not found in dist/index.html`);

await writeFile(file, template.replace(marker, `<div id="app">${await render()}</div>`));
console.log("prerender: dist/index.html");

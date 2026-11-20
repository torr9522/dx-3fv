import { readFile, writeFile } from "node:fs/promises";

const [input, output] = process.argv.slice(2);
if (!input || !output) {
  throw new Error("usage: sanitize-node-metadata.mjs INPUT OUTPUT");
}

const value = JSON.parse(await readFile(input, "utf8"));

function sanitize(item) {
  if (Array.isArray(item)) {
    return item.map(sanitize);
  }
  if (item && typeof item === "object") {
    return Object.fromEntries(
      Object.entries(item)
        .filter(([key]) => key !== "path" && key !== "paths")
        .map(([key, child]) => [key, sanitize(child)]),
    );
  }
  return item;
}

await writeFile(output, `${JSON.stringify(sanitize(value), null, 2)}\n`);

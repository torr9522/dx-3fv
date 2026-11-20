import { readFile, writeFile } from "node:fs/promises";

const [input, output] = process.argv.slice(2);
if (!input || !output) {
  throw new Error("usage: sanitize-go-module-metadata.mjs INPUT OUTPUT");
}

const source = await readFile(input, "utf8");

function parseJsonStream(text) {
  const values = [];
  let start = -1;
  let depth = 0;
  let inString = false;
  let escaped = false;

  for (let index = 0; index < text.length; index += 1) {
    const character = text[index];
    if (start < 0) {
      if (/\s/u.test(character)) {
        continue;
      }
      if (character !== "{") {
        throw new Error(`unexpected JSON stream character at offset ${index}`);
      }
      start = index;
    }
    if (inString) {
      if (escaped) {
        escaped = false;
      } else if (character === "\\") {
        escaped = true;
      } else if (character === '"') {
        inString = false;
      }
      continue;
    }
    if (character === '"') {
      inString = true;
    } else if (character === "{") {
      depth += 1;
    } else if (character === "}") {
      depth -= 1;
      if (depth === 0) {
        values.push(JSON.parse(text.slice(start, index + 1)));
        start = -1;
      }
    }
  }

  if (start >= 0 || depth !== 0 || inString) {
    throw new Error("incomplete JSON object stream");
  }
  return values;
}

const modules = parseJsonStream(source);

const pathFields = new Set(["Dir", "GoMod", "GoModSum"]);

function sanitize(value) {
  if (Array.isArray(value)) {
    return value.map(sanitize);
  }
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value)
        .filter(([key]) => !pathFields.has(key))
        .map(([key, child]) => [key, sanitize(child)]),
    );
  }
  return value;
}

await writeFile(output, `${JSON.stringify(sanitize(modules), null, 2)}\n`);

// One-off: syntax-check every inline <script> block in the given HTML files.
// Usage: node playground/check_inline_js.mjs file.html [more.html...]
import { readFileSync } from "node:fs";
import vm from "node:vm";

let failed = false;
for (const file of process.argv.slice(2)) {
  const html = readFileSync(file, "utf8");
  const re = /<script(?![^>]*\bsrc=)[^>]*>([\s\S]*?)<\/script>/g;
  let m;
  let i = 0;
  while ((m = re.exec(html)) !== null) {
    i++;
    const code = m[1];
    if (!code.trim()) continue;
    try {
      new vm.Script(code.replace(/\{\{[^}]*\}\}/g, "0"));
    } catch (err) {
      console.error(`${file} inline script #${i}: ${err.message}`);
      failed = true;
    }
  }
}
process.exit(failed ? 1 : 0);

import { spawn } from "node:child_process";
import fs from "node:fs/promises";
import path from "node:path";
import os from "node:os";
import process from "node:process";

import { chromium } from "playwright";

function repoRoot() {
  return process.cwd();
}

async function readFirstUrlLine(stream) {
  stream.setEncoding("utf8");
  let buffer = "";
  for await (const chunk of stream) {
    buffer += chunk;
    const lines = buffer.split(/\r?\n/);
    buffer = lines.pop() ?? "";
    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed.startsWith("URL=")) return trimmed.slice("URL=".length);
    }
  }
  throw new Error("demo server did not print URL=...");
}

async function waitForServer(url, { timeoutMs }) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    try {
      const res = await fetch(`${url}/prompts`, { redirect: "manual" });
      if (res.status === 200 || res.status === 303) return;
    } catch {
      // ignore
    }
    await new Promise((r) => setTimeout(r, 150));
  }
  throw new Error(`server not ready after ${timeoutMs}ms: ${url}`);
}

async function shutdownServer(url) {
  try {
    await fetch(`${url}/__shutdown`, { method: "POST" });
  } catch {
    // best effort
  }
}

async function ensureStableRendering(page) {
  await page.addStyleTag({
    content: `
      *, *::before, *::after { transition: none !important; animation: none !important; }
      #hidden-data { display: none !important; }
    `,
  });
}

async function capturePage(page, url, destPath, { waitForSelector, waitForFunction, beforeScreenshot } = {}) {
  await page.goto(url, { waitUntil: "domcontentloaded" });
  await ensureStableRendering(page);
  if (waitForSelector) await page.waitForSelector(waitForSelector, { timeout: 30_000 });
  if (waitForFunction) await page.waitForFunction(waitForFunction, { timeout: 30_000 });
  if (beforeScreenshot) await beforeScreenshot(page);
  await page.waitForTimeout(350); // allow charts/websocket status to paint
  await page.screenshot({ path: destPath, fullPage: false });
}

async function main() {
  const root = repoRoot();
  const assetsDir = path.join(root, "assets");

  const tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), "llm-tournament-shots-"));
  const dbPath = path.join(tmpDir, "demo.db");

  const env = {
    ...process.env,
    CGO_ENABLED: "1",
  };

  const server = spawn(
    "go",
    ["run", "./tools/screenshots/cmd/demo-server", "-db", dbPath, "-addr", "127.0.0.1:0", "-seed=true"],
    { cwd: root, env, stdio: ["ignore", "pipe", "inherit"] }
  );

  const url = await readFirstUrlLine(server.stdout);
  await waitForServer(url, { timeoutMs: 15_000 });

  const browser = await chromium.launch();
  const context = await browser.newContext({
    viewport: { width: 1920, height: 1080 },
    deviceScaleFactor: 1,
  });
  const page = await context.newPage();

  try {
    await capturePage(page, `${url}/results`, path.join(assetsDir, "ui-results.png"), {
      waitForSelector: ".table",
      waitForFunction: () => {
        const cells = document.querySelectorAll("td");
        for (const cell of cells) {
          const bg = getComputedStyle(cell).backgroundColor;
          if (bg && bg !== "rgba(0, 0, 0, 0)" && bg !== "transparent" && bg !== "rgba(255, 255, 255, 0)") {
            return true;
          }
        }
        return false;
      },
    });

    await capturePage(page, `${url}/prompts`, path.join(assetsDir, "ui-prompts.png"), {
      waitForSelector: ".card",
    });
    await capturePage(page, `${url}/profiles`, path.join(assetsDir, "ui-profiles.png"), {
      waitForSelector: ".card",
    });
    await capturePage(page, `${url}/stats`, path.join(assetsDir, "ui-stats.png"), {
      waitForSelector: "canvas",
    });
    await capturePage(page, `${url}/settings`, path.join(assetsDir, "ui-settings.png"), {
      waitForSelector: ".card",
    });

    // Pick a stable model + prompt for Evaluate screenshot.
    await capturePage(
      page,
      `${url}/evaluate?model=${encodeURIComponent("gpt-5.2")}&prompt=0`,
      path.join(assetsDir, "ui-evaluate.png"),
      {
        waitForSelector: ".card",
        beforeScreenshot: async (p) => {
          const buttons = p.locator("button[data-score]").all();
          if (buttons.length > 0) {
            await buttons[3].click({ timeout: 30_000 });
            await p.waitForTimeout(150);
          }
        },
      }
    );
  } finally {
    await context.close();
    await browser.close();
    await shutdownServer(url);
  }

  const exitCode = await new Promise((resolve) => server.on("close", resolve));
  if (exitCode !== 0) {
    throw new Error(`demo server exited with code ${exitCode}`);
  }
}

await main();

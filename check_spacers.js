const { chromium } = require("playwright");
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  await page.goto("http://localhost:8080/results", {
    waitUntil: "networkidle",
  });
  await page.waitForTimeout(100); // Very short wait, before JS runs

  const info = await page.evaluate(() => {
    const profileRow = document.querySelector(".profile-row");
    return {
      profileRowExists: !!profileRow,
      theadRowCount: document.querySelectorAll("thead tr").length,
    };
  });

  console.log("Before JS runs:", JSON.stringify(info, null, 2));

  await browser.close();
})();

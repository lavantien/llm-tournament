const { chromium } = require("playwright");
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  await page.goto("http://localhost:8080/results", {
    waitUntil: "networkidle",
  });
  await page.waitForTimeout(3000);

  const info = await page.evaluate(() => {
    const tbody = document.querySelector("tbody");
    const rows = tbody ? Array.from(tbody.querySelectorAll("tr")) : [];
    const scoreCell = document.querySelector("td.score-cell");
    const cellRect = scoreCell ? scoreCell.getBoundingClientRect() : null;
    const dividerCell = document.querySelector("td.profile-group-divider");
    const dividerComputed = dividerCell
      ? window.getComputedStyle(dividerCell)
      : null;

    return {
      rowCount: rows.length,
      scoreCellCount: document.querySelectorAll("td.score-cell").length,
      scoreCellWidth: Math.round(cellRect?.width || 0),
      scoreCellHeight: Math.round(cellRect?.height || 0),
      dividerCount: document.querySelectorAll("td.profile-group-divider")
        .length,
      dividerBorderLeft: dividerComputed?.borderLeftWidth,
      dividerBorderLeftColor: dividerComputed?.borderLeftColor,
      tableWidth: Math.round(
        document.querySelector(".results-table").getBoundingClientRect()
          .width || 0,
      ),
    };
  });

  console.log("Info:", JSON.stringify(info, null, 2));

  await browser.close();
})();

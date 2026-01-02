/**
 * Utility functions for the results page
 */

/**
 * Safely parses JSON with fallback to default value
 * @param {string} json - JSON string to parse
 * @param {*} defaultValue - Default value if parsing fails
 * @returns {*} Parsed JSON or default value
 */
function safeJsonParse(json, defaultValue = {}) {
  if (!json || typeof json !== "string") return defaultValue;
  try {
    return JSON.parse(json);
  } catch (error) {
    console.error("Error parsing JSON:", error);
    return defaultValue;
  }
}

/**
 * Gets the CSS class for a score
 * @param {number} score - Score value
 * @returns {string} CSS class for the score
 */
function getScoreClass(score) {
  return SCORE_VALUES.includes(score) ? `score-${score}` : "score-0";
}

/**
 * Gets the label for a score
 * @param {number} score - Score value
 * @returns {string} Label for the score
 */
function getScoreLabel(score) {
  return SCORE_LABELS[score] || "N/A";
}

/**
 * Gets the CSS variable name for a score color
 * @param {number} score - Score value
 * @returns {string} CSS variable name
 */
function getScoreColorVar(score) {
  return `--score-color-${score}`;
}

/**
 * Logs score colors for debugging (requires score-utils.js to be loaded first)
 */
function logScoreColors() {
  if (typeof getScoreColor !== 'function') {
    console.error("getScoreColor is not available. Make sure score-utils.js is loaded.");
    return;
  }
  console.log("Score colors used in chart:");
  SCORE_VALUES.forEach((score) => {
    const color = getScoreColor(score);
    console.log(`${score}: ${color}`);
  });
}

/**
 * Scrolls to the top of the page smoothly
 */
function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

/**
 * Scrolls to the bottom of the page smoothly
 */
function scrollToBottom() {
  window.scrollTo({
    top: document.body.scrollHeight,
    behavior: "smooth",
  });
}

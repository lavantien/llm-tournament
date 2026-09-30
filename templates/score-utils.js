/**
 * Score helpers. Requires constants.js (SCORE_COLORS, SCORE_LABELS) to be
 * loaded first.
 */

/**
 * Gets the color for a score
 * @param {number} score - The score value
 * @returns {string} The color as a hex code
 */
function getScoreColor(score) {
  return SCORE_COLORS[score] || SCORE_COLORS[0];
}

/**
 * Gets the label for a score value
 * @param {number} score - The score value
 * @returns {string} Human-readable label for the score
 */
function getScoreLabel(score) {
  return SCORE_LABELS[score] || "N/A";
}

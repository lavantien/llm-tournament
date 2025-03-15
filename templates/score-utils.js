/**
 * Utility functions for working with evaluation scores
 */

// Score value constants
const SCORE_VALUES = [0, 20, 40, 60, 80, 100];

// Map score values to their descriptions
const SCORE_LABELS = {
    0: "N/A",
    20: "1/5",
    40: "2/5",
    60: "3/5",
    80: "4/5",
    100: "5/5"
};

/**
 * Get the appropriate CSS class for a score value
 * @param {number} score - The score value (0, 20, 40, 60, 80, 100)
 * @returns {string} CSS class name for the score
 */
function getScoreClass(score) {
    return SCORE_VALUES.includes(score) ? `score-${score}` : 'score-0';
}

/**
 * Get the label for a score value
 * @param {number} score - The score value
 * @returns {string} Human-readable label for the score
 */
function getScoreLabel(score) {
    return SCORE_LABELS[score] || "N/A";
}

/**
 * Get CSS variable name for a score
 * @param {number} score - The score value
 * @returns {string} CSS variable name
 */
function getScoreColorVar(score) {
    return `--score-color-${score}`;
}

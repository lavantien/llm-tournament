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
    if (!json || typeof json !== 'string') return defaultValue;
    try {
        return JSON.parse(json);
    } catch (error) {
        console.error('Error parsing JSON:', error);
        return defaultValue;
    }
}

/**
 * Gets the appropriate score color from CSS variables
 * @param {number} score - Score value
 * @returns {string} Color for the score
 */
function getScoreColor(score) {
    // First try to get from CSS variables (they're in body)
    const bodyStyles = getComputedStyle(document.body);
    const varName = `--score-color-${score}`;
    const cssColor = bodyStyles.getPropertyValue(varName).trim();
    
    // Return CSS variable if available
    return cssColor || '#808080';
}

/**
 * Gets the CSS class for a score
 * @param {number} score - Score value
 * @returns {string} CSS class for the score
 */
function getScoreClass(score) {
    return SCORE_VALUES.includes(score) ? `score-${score}` : 'score-0';
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
 * Initializes score color CSS variables
 */
function initScoreColorVariables() {
    SCORE_VALUES.forEach(score => {
        const defaultColor = document.body.style.getPropertyValue(`--score-color-${score}`);
        if (!defaultColor) {
            // Set fallback colors if not defined in CSS
            const fallbackColors = {
                0: '#808080',
                20: '#ffa500',
                40: '#ffcc00',
                60: '#ffff00',
                80: '#ccff00',
                100: '#00ff00'
            };
            document.body.style.setProperty(`--score-color-${score}`, fallbackColors[score]);
        }
    });
    console.log("Score color CSS variables initialized");
}

/**
 * Logs score colors for debugging
 */
function logScoreColors() {
    const rootStyles = getComputedStyle(document.documentElement);
    console.log("Score colors used in chart:");
    SCORE_VALUES.forEach(score => {
        const varName = `--score-color-${score}`;
        console.log(`${score}: ${rootStyles.getPropertyValue(varName).trim() || 'Not set'}`);
    });
}

/**
 * Scrolls to the top of the page smoothly
 */
function scrollToTop() {
    window.scrollTo({top: 0, behavior: "smooth"});
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

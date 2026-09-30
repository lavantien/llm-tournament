/**
 * Shared constants for every page. Load this file before score-utils.js and
 * page scripts; it is the single source for score values, labels, and colors.
 */

// Score value constants
const SCORE_VALUES = [0, 20, 40, 60, 80, 100];

// Labels mirror the Go ScoreOptions map in templates/shared.go.
const SCORE_LABELS = {
  0: "0/5",
  20: "1/5",
  40: "2/5",
  60: "3/5",
  80: "4/5",
  100: "5/5",
};

// Score colors mirror the scoreColor funcmap in templates/shared.go.
const SCORE_COLORS = {
  0: "#808080",
  20: "#ffa500",
  40: "#ffd700",
  60: "#00bfff",
  80: "#a77bff",
  100: "#7cff6b",
};

// Profile group borders in the results grid
const PROFILE_BORDER_WIDTH = "5px";

// WebSocket configuration
const MAX_WS_RETRIES = 3;
const WS_RETRY_DELAY_MS = 1000;

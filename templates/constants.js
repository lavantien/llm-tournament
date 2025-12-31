/**
 * Global constants for the results page
 */

// Score-related constants
const SCORE_VALUES = [0, 20, 40, 60, 80, 100];
const SCORE_LABELS = {
  0: "Incorrect",
  20: "Minimal",
  40: "Partial",
  60: "Mostly Correct",
  80: "Correct",
  100: "Perfect",
};

// Default styles
const DEFAULT_BORDER_COLOR = "#333";
const DEFAULT_BORDER_WIDTH = "1px";
const DEFAULT_BORDER_STYLE = "solid";

// Profile styling constants
const PROFILE_BORDER_WIDTH = "5px";
const PROFILE_HEADER_BG_OPACITY = 0.2;

// Table cell dimensions
const SCORE_CELL_SIZE = "50px";

// WebSocket configuration
const MAX_WS_RETRIES = 3;
const WS_RETRY_DELAY_MS = 1000;

// Table row animation constants
const ROW_HOVER_OPACITY = 0.8;
const ROW_HOVER_SCALE = 1.1;

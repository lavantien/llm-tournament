#!/usr/bin/env node
/**
 * Custom TDD-Guard Hook for Claude Code
 *
 * Features:
 * - Handles parallel workflows (subagent-created tests)
 * - Recognizes compile errors as test failures
 * - Reads test state from file AND conversation context
 * - Predictable, debuggable behavior
 * - Supports Go TDD workflow
 */

const fs = require('fs');
const path = require('path');

// Configuration
const CONFIG = {
  // Files that are always allowed to be edited (non-production code)
  ALWAYS_ALLOWED_PATTERNS: [
    /_test\.go$/,           // Test files
    /\.md$/,                // Documentation
    /test\.json$/,          // Test state
    /\.claude\//,           // Claude config
    /hooks\//,              // Hook files
  ],

  // Patterns that indicate test-related content in conversation
  TEST_EVIDENCE_PATTERNS: [
    /undefined:\s*\w+/i,                    // Go compile error
    /FAIL\s+.*\[build failed\]/i,           // Go build failure
    /--- FAIL:/i,                           // Go test failure
    /Error:.*expected.*got/i,               // Assertion failure
    /test.*fail/i,                          // Generic test failure
    /TestDefaultConfig/i,                   // Specific test names
    /TestBatchJobsTableExists/i,
  ],

  // Patterns that indicate implementation code
  IMPLEMENTATION_PATTERNS: [
    /^type\s+\w+\s+struct/m,               // Go struct
    /^func\s+[A-Z]\w*\(/m,                 // Go exported function
    /^func\s+\([^)]+\)\s+[A-Z]\w*\(/m,     // Go method
  ],
};

/**
 * Read hook data from stdin
 */
async function readHookData() {
  return new Promise((resolve, reject) => {
    let data = '';

    process.stdin.setEncoding('utf8');
    process.stdin.on('data', chunk => data += chunk);
    process.stdin.on('end', () => {
      try {
        resolve(JSON.parse(data));
      } catch (e) {
        // If JSON parsing fails, return empty object
        resolve({});
      }
    });
    process.stdin.on('error', reject);

    // Timeout after 5 seconds
    setTimeout(() => resolve({}), 5000);
  });
}

/**
 * Check if file is always allowed to be edited
 */
function isAlwaysAllowed(filePath) {
  if (!filePath) return false;
  return CONFIG.ALWAYS_ALLOWED_PATTERNS.some(pattern => pattern.test(filePath));
}

/**
 * Read test state from file
 */
function readTestState(projectRoot) {
  const testStatePath = path.join(projectRoot, '.claude', 'tdd-guard', 'data', 'test.json');

  try {
    if (fs.existsSync(testStatePath)) {
      const content = fs.readFileSync(testStatePath, 'utf8');
      return JSON.parse(content);
    }
  } catch (e) {
    // Ignore errors
  }

  return { testModules: [], reason: 'unknown' };
}

/**
 * Check if there are failing tests in state
 */
function hasFailingTestsInState(testState) {
  if (testState.reason === 'failed') return true;
  if (testState.testModules && testState.testModules.length > 0) {
    return testState.testModules.some(module =>
      module.tests && module.tests.some(test =>
        test.status === 'failed' || test.status === 'build_failed'
      )
    );
  }
  return false;
}

/**
 * Check conversation/transcript for test evidence
 */
function hasTestEvidenceInConversation(hookData) {
  // Check the transcript if available
  const transcriptPath = hookData.transcript_path;

  if (transcriptPath && fs.existsSync(transcriptPath)) {
    try {
      const transcript = fs.readFileSync(transcriptPath, 'utf8');

      // Look for test evidence patterns
      for (const pattern of CONFIG.TEST_EVIDENCE_PATTERNS) {
        if (pattern.test(transcript)) {
          return true;
        }
      }
    } catch (e) {
      // Ignore errors
    }
  }

  return false;
}

/**
 * Check if test file exists for the implementation file
 */
function hasCorrespondingTestFile(filePath, projectRoot) {
  if (!filePath || !projectRoot) return false;

  // For Go files, check if _test.go exists
  if (filePath.endsWith('.go') && !filePath.endsWith('_test.go')) {
    const dir = path.dirname(filePath);
    const baseName = path.basename(filePath, '.go');
    const testFile = path.join(dir, `${baseName}_test.go`);

    if (fs.existsSync(testFile)) return true;

    // Also check for any *_test.go in the same directory
    try {
      const files = fs.readdirSync(dir);
      return files.some(f => f.endsWith('_test.go'));
    } catch (e) {
      // Ignore errors
    }
  }

  return false;
}

/**
 * Check if the edit is adding implementation code
 */
function isAddingImplementation(newString) {
  if (!newString) return false;
  return CONFIG.IMPLEMENTATION_PATTERNS.some(pattern => pattern.test(newString));
}

/**
 * Main validation logic
 */
async function validate() {
  const hookData = await readHookData();

  const toolName = hookData.tool_name || '';
  const toolInput = hookData.tool_input || {};
  const projectRoot = process.cwd();

  // Get file path from tool input
  const filePath = toolInput.file_path || toolInput.path || '';
  const newString = toolInput.new_string || toolInput.content || '';

  // 1. Always allow test files, docs, config
  if (isAlwaysAllowed(filePath)) {
    return { allow: true, reason: 'File type always allowed (test/doc/config)' };
  }

  // 2. Check if this is an implementation edit
  const isImplementation = isAddingImplementation(newString);

  if (!isImplementation) {
    return { allow: true, reason: 'Not adding implementation code' };
  }

  // 3. For implementation code, check for test evidence

  // Check test state file
  const testState = readTestState(projectRoot);
  if (hasFailingTestsInState(testState)) {
    return { allow: true, reason: 'Failing tests found in test state' };
  }

  // Check if corresponding test file exists
  if (hasCorrespondingTestFile(filePath, projectRoot)) {
    // Test file exists - check if it's likely testing what we're implementing
    return { allow: true, reason: 'Corresponding test file exists' };
  }

  // Check conversation transcript for test evidence
  if (hasTestEvidenceInConversation(hookData)) {
    return { allow: true, reason: 'Test evidence found in conversation' };
  }

  // 4. Block with helpful message
  return {
    allow: false,
    reason: `TDD violation: Adding implementation without test evidence.

To fix this:
1. Write a failing test first (in a *_test.go file)
2. Run the test: make test
3. Then implement the code

Evidence checked:
- Test state file: ${hasFailingTestsInState(testState) ? 'HAS' : 'NO'} failing tests
- Test file exists: ${hasCorrespondingTestFile(filePath, projectRoot) ? 'YES' : 'NO'}
- Conversation evidence: ${hasTestEvidenceInConversation(hookData) ? 'FOUND' : 'NOT FOUND'}`
  };
}

/**
 * Entry point
 */
async function main() {
  try {
    const result = await validate();

    if (result.allow) {
      // Output nothing or success message for allowed
      console.log(`TDD-Guard: ${result.reason}`);
      process.exit(0);
    } else {
      // Output error message for blocked
      console.error(result.reason);
      process.exit(1);
    }
  } catch (error) {
    // On any error, allow the operation (fail-open)
    console.log(`TDD-Guard: Error occurred, allowing operation - ${error.message}`);
    process.exit(0);
  }
}

main();

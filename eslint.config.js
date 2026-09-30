// Minimal ESLint flat config. This repo has no ESLint toolchain of its own;
// the config exists so environments that auto-run eslint (pre-commit hooks
// using npx) parse cleanly and skip vendored, generated, and binary files.
export default [
  {
    ignores: [
      "node_modules/",
      "templates/vendor/",
      "templates/output.css",
      "templates/output.css.map",
      "coverage*",
      "release/",
      "playground/",
    ],
  },
  {
    files: ["templates/**/*.js"],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: "script",
      globals: {
        window: "readonly",
        document: "readonly",
        console: "readonly",
        navigator: "readonly",
        WebSocket: "readonly",
        fetch: "readonly",
        XMLHttpRequest: "readonly",
        URL: "readonly",
        FormData: "readonly",
        Chart: "readonly",
        marked: "readonly",
        confirm: "readonly",
        alert: "readonly",
        setTimeout: "readonly",
        location: "readonly",
        require: "readonly",
        module: "readonly",
      },
    },
  },
  {
    files: ["tools/**/*.mjs"],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: "module",
    },
  },
];

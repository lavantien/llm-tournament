/** @type {import('tailwindcss').Config} */
export default {
  content: ["./templates/**/*.html", "./tools/screenshots/**/*.html"],
  plugins: [require("@tailwindcss/forms"), require("daisyui")],
  daisyui: {
    themes: ["coffee"],
    darkTheme: "coffee",
  },
};

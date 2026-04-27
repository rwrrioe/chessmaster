import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}", "./lib/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: "#050505",
        bone: "#FDFBF7",
      },
      fontFamily: {
        display: ["'Clash Display'", "'Geist'", "ui-sans-serif", "system-ui"],
        sans: ["'Geist'", "ui-sans-serif", "system-ui"],
        mono: ["'Geist Mono'", "ui-monospace", "monospace"],
        editorial: ["'PP Editorial New'", "ui-serif", "Georgia"],
      },
      borderRadius: {
        "2rem": "2rem",
        "1.5rem": "1.5rem",
      },
      transitionTimingFunction: {
        fluid: "cubic-bezier(0.32,0.72,0,1)",
      },
    },
  },
  plugins: [],
};
export default config;

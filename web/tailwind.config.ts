import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: "#050505",
        bone: "#FDFBF7",
      },
      fontFamily: {
        display: ["'Clash Display'", "ui-sans-serif", "system-ui"],
        sans: ["'Geist'", "ui-sans-serif", "system-ui"],
        editorial: ["'PP Editorial New'", "ui-serif", "Georgia"],
      },
    },
  },
  plugins: [],
};
export default config;

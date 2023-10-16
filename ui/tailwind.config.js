/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    fontFamily: {
      sans: ["Outfit", "sans-serif"],
    },
    colors: {
      inherit: "inherit",
      current: "currentColor",
      transparent: "transparent",
      primary: {
        100: "#FBFDFF",
        300: "#F1F8FF",
        400: "#E9F4FF",
        500: "#C2D1DF",
        600: "#798C8C",
      },
      pink: {
        100: "#FFE5F1",
        400: "#FFA8D0",
        500: "#F463A0",
      },
      blue: {
        100: "#B7DAFF",
        400: "#82BFFF",
        500: "#3C9BFF",
      },
      nuances: {
        white: "#FFFFFF",
        black: "#000000",
        50: "#F2F2F2",
        100: "#E5E5E5",
        200: "#CCCCCC",
        300: "#737373",
        400: "#252525",
      },
      status: {
        error: {
          default: "#E14029",
          light: "#F9D9D4",
        },
        success: {
          default: "#6BEF70",
          light: "#E1FCE2",
        },
        warning: {
          default: "#FFEA51",
          light: "#FFFBDC",
        },
      },
      overlay: "#00000040",
    },
    boxShadow: {
      light: "0px 0px 32px 0px rgba(121, 140, 140, 0.16)",
      dark: "0px 4px 4px 0px rgba(0, 0, 0, 0.25)",
    },
    backgroundImage: {
      "background-light":
        "url('@/assets/backgrounds/background-light.png'), linear-gradient(to right, #E9F4FF, #E9F4FF)",
      "background-dark":
        "url('@/assets/backgrounds/background-dark.png'), linear-gradient(to right, #252525, #252525)",
    },
  },
  plugins: [],
};

import {
  defineConfig,
  presetAttributify,
  presetIcons,
  transformerDirectives,
  transformerVariantGroup,
  transformerAttributifyJsx,
} from "unocss";
import presetWind3 from "@unocss/preset-wind3";

export default defineConfig({
  shortcuts: [
    ["flex-center", "flex items-center justify-center"],
    ["hstack", "flex items-center"],
    ["vstack", "hstack flex-col"],
    ["no-outline", "outline-none focus:outline-none"],
    ["shadow-menu", "user-shadow-menu"],
    ["macos-window-btn", "size-3 text-black rounded-full flex-center no-outline"],
    ["border-menu", "user-color-border"],
    [
      "menu-box",
      "fixed top-8.5 user-color-ftext user-color-fbg border user-color-border user-rounded-lg shadow-menu",
    ],
    [
      "safari-btn",
      "h-6 outline-none focus:outline-none user-rounded-1 flex-center border user-color-border",
    ],
    ["cc-btn", "rounded-full p-2 user-color-bg"],
    [
      "cc-btn-active",
      "rounded-full p-2 user-color-bg",
    ],
    ["cc-text", "text-xs user-color-ftext opacity-70"],
    ["cc-grid", "user-color-fbg user-rounded-xl cc-grid-shadow backdrop-blur-2xl"],
    ["battery-level", "absolute rounded-[1px] h-2 top-1/2 -mt-1 ml-0.5 left-0"],
    ["flex-center-v", "flex items-center"],
    ["cc-mode", "p-2 rounded-full user-color-bg"],
  ],
  rules: [
    [
      "cc-grid-shadow",
      {
          "box-shadow":
          "0px 1px 5px 0px color-mix(in srgb, var(--user-text-1-color) 30%, transparent)",
      },
    ],
    [
      "user-shadow-menu",
      {
          "box-shadow":
          "0 4px 12px color-mix(in srgb, var(--user-text-1-color) 25%, transparent)",
      },
    ],
    [
      /^user-rounded-(.+)$/,
      ([, match]) => {
        const propertyMap = {
          t: ["border-top-left-radius", "border-top-right-radius"],
          r: ["border-top-right-radius", "border-bottom-right-radius"],
          b: ["border-bottom-left-radius", "border-bottom-right-radius"],
          l: ["border-top-left-radius", "border-bottom-left-radius"],
          tl: ["border-top-left-radius"],
          tr: ["border-top-right-radius"],
          bl: ["border-bottom-left-radius"],
          br: ["border-bottom-right-radius"],
        };
        const sizeMap = {
          xs: "0.125rem",
          sm: "0.25rem",
          md: "0.375rem",
          lg: "0.5rem",
          xl: "0.75rem",
          "2xl": "1rem",
          "3xl": "1.5rem",
          full: "9999px",
        };
        const declarations = {};
        const parts = match.split("-");
        if (parts.length === 0) {
          return declarations;
        }

        const sizeKey = parts.pop();
        let val = sizeMap[sizeKey];

        if (!val) {
          const numVal = parseFloat(sizeKey);
          if (!isNaN(numVal)) {
            val = `${numVal * 0.25}rem`;
          } else {
            return declarations;
          }
        }
        const radiusValue = `calc(var(--user-round-enabled) * ${val})`;
        if (parts.length > 0) {
          for (let i = 0; i < parts.length; i++) {
            const position = parts[i];
            if (propertyMap[position]) {
              propertyMap[position].forEach((cssProp) => {
                declarations[cssProp] = radiusValue;
              });
            } else {
            }
          }
        } else {
          declarations["border-radius"] = radiusValue;
        }
        return declarations;
      },
    ],
    [
      /^user-color-(.+)$/,
      ([, match]) => {
        const declarations = {};
        const propertyKey = match.toLowerCase();
        if (propertyKey === "ftext" || propertyKey === "text-1") {
          declarations["color"] = "var(--user-text-1-color)";
        } else if (propertyKey === "text-2") {
          declarations["color"] = "var(--user-text-2-color)";
        } else if (propertyKey === "text-3" || propertyKey === "muted") {
          declarations["color"] = "var(--user-text-3-color)";
        } else if (propertyKey === "fbg" || propertyKey === "bg-1") {
          declarations["background-color"] = "var(--user-bg-1-color)";
        } else if (propertyKey === "bg-2") {
          declarations["background-color"] = "var(--user-bg-2-color)";
        } else if (propertyKey === "bg-3") {
          declarations["background-color"] = "var(--user-bg-3-color)";
        } else if (propertyKey === "surface") {
          declarations["background-color"] = "var(--user-surface-color)";
          declarations["color"] = "var(--user-text-1-color)";
        } else if (propertyKey === "surface-muted") {
          declarations["background-color"] = "var(--user-surface-muted-color)";
          declarations["color"] = "var(--user-text-1-color)";
        } else if (propertyKey === "readable") {
          declarations["background-color"] = "var(--user-readable-surface-color)";
          declarations["color"] = "var(--user-text-1-color)";
        } else if (propertyKey === "control") {
          declarations["background-color"] = "var(--user-control-color)";
          declarations["color"] = "var(--user-text-1-color)";
        } else if (propertyKey === "control-hover" || propertyKey === "hover") {
          declarations["background-color"] = "var(--user-control-hover-color)";
          declarations["color"] = "var(--user-text-1-color)";
        } else if (propertyKey === "line" || propertyKey === "divider") {
          declarations["border-color"] = "var(--user-divider-color)";
        } else if (propertyKey === "text") {
          declarations["color"] = "var(--user-primary-color)";
        } else if (propertyKey === "bg" || propertyKey === "primary-bg") {
          declarations["background-color"] = "var(--user-primary-color)";
          declarations["color"] = "var(--user-primary-text-color)";
        } else if (propertyKey === "border" || propertyKey === "primary-border") {
          declarations["border-color"] = "var(--user-primary-color)";
        } else if (propertyKey === "outline") {
          declarations["outline-color"] = "var(--user-primary-color)";
        } else if (propertyKey === "fill") {
          declarations["fill"] = "var(--user-primary-color)";
        } else if (propertyKey === "stroke") {
          declarations["stroke"] = "var(--user-primary-color)";
        }
        return declarations;
      },
    ],
  ],
  presets: [
    presetWind3(),
    presetAttributify(),
    presetIcons({
      warn: true,
      extraProperties: {
        display: "inline-block",
      },
    }),
  ],
  transformers: [
    transformerDirectives(),
    transformerVariantGroup(),
    transformerAttributifyJsx(),
  ],
});

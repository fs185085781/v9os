const DEFAULT_PRIMARY = "green";

export const colorMap = {
  green: {
    primaryColor: "#18a058",
    primaryColorHover: "#36ad6a",
    primaryColorPressed: "#0c7a43",
    primaryColorSuppl: "#36ad6a",
  },
  blue: {
    primaryColor: "#2080f0",
    primaryColorHover: "#4098fc",
    primaryColorPressed: "#1060c9",
    primaryColorSuppl: "#4098fc",
  },
  orange: {
    primaryColor: "#ff9900",
    primaryColorHover: "#ffad33",
    primaryColorPressed: "#f29100",
    primaryColorSuppl: "#ffad33",
  },
  purple: {
    primaryColor: "#722ed1",
    primaryColorHover: "#9254de",
    primaryColorPressed: "#ab7ae0",
    primaryColorSuppl: "#9254de",
  },
  red: {
    primaryColor: "#d03050",
    primaryColorHover: "#f56c6c",
    primaryColorPressed: "#c45656",
    primaryColorSuppl: "#f56c6c",
  },
  cyan: {
    primaryColor: "#0fb9b1",
    primaryColorHover: "#2de0c9",
    primaryColorPressed: "#0ea5a5",
    primaryColorSuppl: "#2de0c9",
  },
  pink: {
    primaryColor: "#f759ab",
    primaryColorHover: "#ff85c0",
    primaryColorPressed: "#f5317f",
    primaryColorSuppl: "#ff85c0",
  },
  yellow: {
    primaryColor: "#fadb14",
    primaryColorHover: "#ffec3d",
    primaryColorPressed: "#d4b106",
    primaryColorSuppl: "#ffec3d",
  },
  gray: {
    primaryColor: "#8c8c8c",
    primaryColorHover: "#a6a6a6",
    primaryColorPressed: "#737373",
    primaryColorSuppl: "#a6a6a6",
  },
  deepBlue: {
    primaryColor: "#1d39c4",
    primaryColorHover: "#2540e9",
    primaryColorPressed: "#10239c",
    primaryColorSuppl: "#2540e9",
  },
  deepPurple: {
    primaryColor: "#531dab",
    primaryColorHover: "#693ac9",
    primaryColorPressed: "#3c1380",
    primaryColorSuppl: "#693ac9",
  },
  brown: {
    primaryColor: "#ad4e00",
    primaryColorHover: "#d46b08",
    primaryColorPressed: "#873b00",
    primaryColorSuppl: "#d46b08",
  },
};

const roundMap = {
  true: {},
  false: {
    borderRadius: "0px",
    borderRadiusSmall: "0px",
    borderRadiusMedium: "0px",
    borderRadiusLarge: "0px",
  },
};

const clamp = (value, min, max) => Math.max(min, Math.min(max, value));

const alphaHex = (alpha) =>
  Math.round(clamp(alpha, 0, 1) * 255)
    .toString(16)
    .padStart(2, "0");

const normalizeHex = (color) => {
  const value = String(color || "").trim();
  if (/^#[0-9a-f]{3}$/i.test(value)) {
    return `#${value
      .slice(1)
      .split("")
      .map((item) => item + item)
      .join("")}`;
  }
  if (/^#[0-9a-f]{6}$/i.test(value)) {
    return value;
  }
  return "";
};

const parseColor = (color) => {
  const hex = normalizeHex(color);
  if (hex) {
    const num = parseInt(hex.slice(1), 16);
    return {
      r: (num >> 16) & 255,
      g: (num >> 8) & 255,
      b: num & 255,
    };
  }
  const match = String(color || "").match(
    /rgba?\(\s*([\d.]+)\s*,\s*([\d.]+)\s*,\s*([\d.]+)/i,
  );
  if (match) {
    return {
      r: clamp(Number(match[1]), 0, 255),
      g: clamp(Number(match[2]), 0, 255),
      b: clamp(Number(match[3]), 0, 255),
    };
  }
  return { r: 32, g: 128, b: 240 };
};

const rgba = (color, alpha) => {
  const { r, g, b } = parseColor(color);
  return `rgba(${Math.round(r)}, ${Math.round(g)}, ${Math.round(b)}, ${clamp(
    alpha,
    0,
    1,
  ).toFixed(2)})`;
};

const luminance = (color) => {
  const { r, g, b } = parseColor(color);
  const [sr, sg, sb] = [r, g, b].map((channel) => {
    const value = channel / 255;
    return value <= 0.03928
      ? value / 12.92
      : Math.pow((value + 0.055) / 1.055, 2.4);
  });
  return 0.2126 * sr + 0.7152 * sg + 0.0722 * sb;
};

export const contrastRatio = (a, b) => {
  const l1 = luminance(a);
  const l2 = luminance(b);
  const lighter = Math.max(l1, l2);
  const darker = Math.min(l1, l2);
  return (lighter + 0.05) / (darker + 0.05);
};

export const readableTextColor = (backgroundColor) => {
  const black = "#000000";
  const white = "#ffffff";
  return contrastRatio(backgroundColor, black) >=
    contrastRatio(backgroundColor, white)
    ? black
    : white;
};

export const accentTextColor = (backgroundColor) => {
  const white = "#ffffff";
  const black = "#000000";
  if (contrastRatio(backgroundColor, white) >= 3) {
    return white;
  }
  return contrastRatio(backgroundColor, black) >= 3 ? black : readableTextColor(backgroundColor);
};

export const resolveColorTheme = (color, colorDesc) => {
  let colorTheme = colorMap[color] || colorMap[DEFAULT_PRIMARY];
  if (color === "diy" && colorDesc) {
    const colors = colorDesc
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean);
    if (colors.length === 1) {
      colors.push(colors[0], colors[0]);
    } else if (colors.length === 2) {
      colors.push(colors[0]);
    }
    colorTheme = {
      primaryColorHover: colors[0],
      primaryColor: colors[1],
      primaryColorPressed: colors[2],
      primaryColorSuppl: colors[0],
    };
  }
  return { ...colorTheme };
};

const createRole = (name, color, options = {}) => {
  const bgAlpha = options.bgAlpha ?? 0.14;
  return {
    [`--user-${name}-color`]: color,
    [`--user-${name}-bg-color`]: rgba(color, bgAlpha),
    [`--user-${name}-border-color`]: rgba(color, 0.34),
    [`--user-${name}-text-color`]: accentTextColor(color),
  };
};

export const buildPersonalTheme = (settings = {}) => {
  const isDark = settings.Theme === "dark";
  const transparent = clamp(Number(settings.Transparent) || 0, 0, 100);
  const surfaceAlpha = (100 - transparent) / 100;
  const readableAlpha = clamp(surfaceAlpha, isDark ? 0.72 : 0.78, 1);
  const baseBg = isDark ? "#000000" : "#ffffff";
  const text1 = isDark ? "rgba(255, 255, 255, 0.88)" : "rgba(0, 0, 0, 0.88)";
  const text2 = isDark ? "rgba(255, 255, 255, 0.72)" : "rgba(0, 0, 0, 0.68)";
  const text3 = isDark ? "rgba(255, 255, 255, 0.48)" : "rgba(0, 0, 0, 0.46)";
  const colorTheme = resolveColorTheme(settings.Color, settings.ColorDesc);
  const primaryText = accentTextColor(colorTheme.primaryColor);
  const infoText = accentTextColor("#2080f0");
  const successText = accentTextColor("#18a058");
  const warningText = accentTextColor("#f0a020");
  const errorText = accentTextColor("#d03050");
  const roundTheme = roundMap[String(settings.Round)] || roundMap.true;

  const vars = {
    "--user-primary-color": colorTheme.primaryColor,
    "--user-primary-color-hover": colorTheme.primaryColorHover,
    "--user-primary-color-pressed": colorTheme.primaryColorPressed,
    "--user-primary-color-suppl": colorTheme.primaryColorSuppl,
    "--user-primary-text-color": primaryText,
    "--user-primary-tcolor": primaryText,

    "--user-bg-color": `${baseBg}${alphaHex(surfaceAlpha)}`,
    "--user-bg-1-color": `${baseBg}${alphaHex(surfaceAlpha)}`,
    "--user-bg-2-color": rgba(baseBg, clamp(surfaceAlpha * 0.82, 0, 1)),
    "--user-bg-3-color": rgba(baseBg, clamp(surfaceAlpha * 0.64, 0, 1)),
    "--user-bg-filter-color": rgba(baseBg, isDark ? 0.05 : 0.08),
    "--user-glass-tint-color": isDark
      ? "rgba(0, 0, 0, 0.18)"
      : "rgba(255, 255, 255, 0.3)",
    "--user-glass-blur": `${transparent > 0 ? 3 : 0}px`,
    "--user-surface-color": rgba(baseBg, surfaceAlpha),
    "--user-surface-muted-color": rgba(baseBg, clamp(surfaceAlpha * 0.72, 0, 1)),
    "--user-readable-surface-color": rgba(baseBg, readableAlpha),
    "--user-control-color": isDark
      ? "rgba(255, 255, 255, 0.08)"
      : "rgba(0, 0, 0, 0.04)",
    "--user-control-hover-color": isDark
      ? "rgba(255, 255, 255, 0.13)"
      : "rgba(0, 0, 0, 0.07)",
    "--user-border-color": isDark
      ? "rgba(255, 255, 255, 0.18)"
      : "rgba(0, 0, 0, 0.13)",
    "--user-divider-color": isDark
      ? "rgba(255, 255, 255, 0.1)"
      : "rgba(0, 0, 0, 0.08)",
    "--user-hover-color": isDark
      ? "rgba(255, 255, 255, 0.1)"
      : "rgba(0, 0, 0, 0.06)",
    "--user-active-color": rgba(colorTheme.primaryColor, isDark ? 0.26 : 0.16),

    "--user-text-color": text1,
    "--user-text-1-color": text1,
    "--user-text-2-color": text2,
    "--user-text-3-color": text3,
    "--user-text-muted-color": text3,
    "--user-round-enabled": settings.Round === "false" ? "0" : "1",

    ...createRole("success", "#18a058"),
    ...createRole("warning", "#f0a020"),
    ...createRole("error", "#d03050"),
    ...createRole("info", "#2080f0"),
  };

  const common = {
    ...roundTheme,
    ...colorTheme,
    baseColor: rgba(baseBg, surfaceAlpha),
    bodyColor: rgba(baseBg, surfaceAlpha),
    cardColor: rgba(baseBg, readableAlpha),
    modalColor: rgba(baseBg, readableAlpha),
    popoverColor: rgba(baseBg, readableAlpha),
    tableColor: rgba(baseBg, readableAlpha),
    inputColor: isDark ? "rgba(255, 255, 255, 0.08)" : "rgba(0, 0, 0, 0.04)",
    inputColorDisabled: isDark
      ? "rgba(255, 255, 255, 0.05)"
      : "rgba(0, 0, 0, 0.03)",
    actionColor: vars["--user-control-color"],
    tabColor: vars["--user-control-color"],
    tableHeaderColor: vars["--user-control-color"],
    codeColor: isDark ? "rgba(255, 255, 255, 0.12)" : "rgba(0, 0, 0, 0.05)",
    tagColor: vars["--user-control-color"],
    hoverColor: vars["--user-hover-color"],
    pressedColor: isDark ? "rgba(255, 255, 255, 0.07)" : "rgba(0, 0, 0, 0.08)",
    tableColorHover: vars["--user-hover-color"],
    tableColorStriped: isDark
      ? "rgba(255, 255, 255, 0.05)"
      : "rgba(0, 0, 0, 0.025)",
    borderColor: vars["--user-border-color"],
    dividerColor: vars["--user-divider-color"],
    textColorBase: text1,
    textColor1: text1,
    textColor2: text2,
    textColor3: text3,
    placeholderColor: text3,
    textColorDisabled: text3,
    iconColor: text3,
  };

  const headerColor = isDark
    ? "rgba(255, 255, 255, 0.08)"
    : "rgba(0, 0, 0, 0.04)";
  const headerHoverColor = isDark
    ? "rgba(255, 255, 255, 0.12)"
    : "rgba(0, 0, 0, 0.07)";

  return {
    vars,
    colorTheme,
    roundTheme,
    naiveThemeOverrides: {
      common,
      DataTable: {
        thColor: headerColor,
        thColorHover: headerHoverColor,
        thColorSorting: headerHoverColor,
        thColorModal: headerColor,
        thColorHoverModal: headerHoverColor,
        thColorSortingModal: headerHoverColor,
        thColorPopover: headerColor,
        thColorHoverPopover: headerHoverColor,
        thColorSortingPopover: headerHoverColor,
      },
      Button: {
        textColorPrimary: primaryText,
        textColorHoverPrimary: primaryText,
        textColorPressedPrimary: primaryText,
        textColorFocusPrimary: primaryText,
        textColorDisabledPrimary: primaryText,
        textColorInfo: infoText,
        textColorHoverInfo: infoText,
        textColorPressedInfo: infoText,
        textColorFocusInfo: infoText,
        textColorDisabledInfo: infoText,
        textColorSuccess: successText,
        textColorHoverSuccess: successText,
        textColorPressedSuccess: successText,
        textColorFocusSuccess: successText,
        textColorDisabledSuccess: successText,
        textColorWarning: warningText,
        textColorHoverWarning: warningText,
        textColorPressedWarning: warningText,
        textColorFocusWarning: warningText,
        textColorDisabledWarning: warningText,
        textColorError: errorText,
        textColorHoverError: errorText,
        textColorPressedError: errorText,
        textColorFocusError: errorText,
        textColorDisabledError: errorText,
      },
    },
  };
};

export const applyThemeVars = (target, vars) => {
  const style = target?.style || target;
  if (!style || !vars) return;
  Object.entries(vars).forEach(([key, value]) => {
    if (value != null) {
      style.setProperty(key, String(value));
    }
  });
};

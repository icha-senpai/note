import {colorRGB, contrastColor, normalizeColor} from "./color";

export type InterfaceColorRole = "accent" | "background" | "surface" | "text";
export type InterfaceColors = Partial<Record<"0" | "1", Partial<Record<InterfaceColorRole, string>>>>;

const variables = ["--b3-theme-primary", "--b3-theme-primary-light", "--b3-theme-primary-lighter", "--b3-theme-primary-lightest", "--b3-theme-on-primary",
    "--b3-theme-background", "--b3-body-background", "--b3-theme-background-light", "--b3-theme-on-background", "--b3-theme-surface", "--b3-theme-surface-light", "--b3-theme-surface-lighter",
    "--b3-theme-on-surface", "--b3-theme-on-surface-light", "--b3-theme-on-surface-lighter"];

export const applyInterfaceColors = (colors: InterfaceColors, mode: number) => {
    const style = document.documentElement.style;
    variables.forEach(variable => style.removeProperty(variable));
    const values = colors?.[String(mode) as "0" | "1"] || {};
    const accent = normalizeColor(values.accent);
    const background = normalizeColor(values.background);
    const surface = normalizeColor(values.surface);
    const text = normalizeColor(values.text);
    if (accent) {
        style.setProperty("--b3-theme-primary", accent);
        style.setProperty("--b3-theme-on-primary", contrastColor(accent));
        ["light", "lighter", "lightest"].forEach((suffix, index) => style.setProperty(`--b3-theme-primary-${suffix}`, `rgba(${colorRGB(accent).join(",")},${[0.54, 0.38, 0.12][index]})`));
    }
    if (background) {
        style.setProperty("--b3-theme-background", background);
        style.setProperty("--b3-body-background", background);
        style.setProperty("--b3-theme-background-light", `color-mix(in srgb, ${background} 92%, ${contrastColor(background)})`);
        style.setProperty("--b3-theme-on-background", contrastColor(background));
    }
    if (surface) {
        style.setProperty("--b3-theme-surface", surface);
        style.setProperty("--b3-theme-surface-light", `rgba(${colorRGB(surface).join(",")},0.86)`);
        style.setProperty("--b3-theme-surface-lighter", `color-mix(in srgb, ${surface} 85%, ${contrastColor(surface)})`);
    }
    if (text || surface) {
        const foreground = text || contrastColor(surface);
        style.setProperty("--b3-theme-on-surface", foreground);
        style.setProperty("--b3-theme-on-surface-light", `rgba(${colorRGB(foreground).join(",")},0.68)`);
        style.setProperty("--b3-theme-on-surface-lighter", `rgba(${colorRGB(foreground).join(",")},0.38)`);
    }
    if (text) { style.setProperty("--b3-theme-on-background", text); }
};

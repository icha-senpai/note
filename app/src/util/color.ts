export const normalizeColor = (value: string): string | undefined => {
    const text = typeof value === "string" ? value.trim() : "";
    if (/^#[\da-f]{6}$/i.test(text)) { return text.toLowerCase(); }
    if (/^#[\da-f]{3}$/i.test(text)) { return "#" + text.slice(1).split("").map(part => part + part).join("").toLowerCase(); }
    const rgb = /^rgb\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)$/i.exec(text);
    if (rgb && rgb.slice(1).every(part => Number(part) <= 255)) {
        return "#" + rgb.slice(1).map(part => Number(part).toString(16).padStart(2, "0")).join("");
    }
    return undefined;
};

export const colorRGB = (color: string): number[] => {
    const hex = normalizeColor(color) || "#000000";
    return [1, 3, 5].map(index => parseInt(hex.slice(index, index + 2), 16));
};

export const contrastColor = (color: string): string => {
    const channels = colorRGB(color).map(channel => {
        const value = channel / 255;
        return value <= 0.04045 ? value / 12.92 : Math.pow((value + 0.055) / 1.055, 2.4);
    });
    const luminance = channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722;
    return (luminance + 0.05) / 0.05 >= 1.05 / (luminance + 0.05) ? "#000000" : "#ffffff";
};

// Stored palette IDs stay compatible with older databases; custom colors are opaque HEX.
export const selectColorStyle = (value: string): string => {
    const hex = normalizeColor(value);
    if (hex) { return `background-color:${hex};color:${contrastColor(hex)}`; }
    const index = /^(?:[1-9]|1[0-4])$/.test(value) ? value : "1";
    return `background-color:var(--b3-font-background${index});color:var(--b3-font-color${index})`;
};

export const selectColorBackground = (value: string): string => {
    const hex = normalizeColor(value);
    if (hex) { return `color-mix(in srgb, ${hex} 16%, var(--b3-theme-background))`; }
    return `var(--b3-font-background${/^(?:[1-9]|1[0-4])$/.test(value) ? value : "1"})`;
};

export const selectColorValue = (background: string, foreground: string): string => normalizeColor(background) ||
    /^var\(--b3-font-background([1-9]|1[0-4])\)$/.exec(background)?.[1] ||
    /^var\(--b3-font-color([1-9]|1[0-4])\)$/.exec(foreground)?.[1] || "1";

// JSON Canvas reserves 1-6 for red, orange, yellow, green, cyan, and purple.
export const canvasColor = (value: string): string | undefined => normalizeColor(value) ||
    (/^[1-6]$/.test(value) ? ["#e05252", "#e88d39", "#e6bc3f", "#54a76a", "#42a6b5", "#a06ac4"][Number(value) - 1] : undefined);

export interface ColorPalette { id: string; name: string; colors: string[]; }
export interface ColorLibrary { recent: string[]; palettes: ColorPalette[]; }

export const sanitizeColorLibrary = (data: unknown): ColorLibrary => {
    const colors = (items: unknown) => Array.isArray(items) ? Array.from(new Set(items.map(normalizeColor).filter(Boolean))).slice(0, 16) : [];
    const source = data && typeof data === "object" ? data as Partial<ColorLibrary> : {};
    const palettes: ColorPalette[] = [];
    if (Array.isArray(source.palettes)) {
        source.palettes.slice(0, 8).forEach(palette => {
            if (!palette || typeof palette.id !== "string" || typeof palette.name !== "string" || palettes.some(item => item.id === palette.id)) { return; }
            palettes.push({id: palette.id.slice(0, 80), name: palette.name.trim().slice(0, 80), colors: colors(palette.colors)});
        });
    }
    return {recent: colors(source.recent), palettes};
};

import {Dialog} from "../dialog";
import {escapeAttr, escapeHtml} from "./escape";
import {colorRGB, ColorLibrary, normalizeColor, sanitizeColorLibrary} from "./color";
import {fetchSyncPost} from "./fetch";
import {Constants} from "../constants";

export const resolveColor = (value: string, fallback = "#3575f0"): string => {
    const direct = normalizeColor(value);
    if (direct) { return direct; }
    const variable = /^var\((--b3-[a-z\d-]+)\)$/.exec(/^(?:[1-9]|1[0-4])$/.test(value) ? `var(--b3-font-color${value})` : value || "");
    return (variable && normalizeColor(getComputedStyle(document.documentElement).getPropertyValue(variable[1]))) || fallback;
};

// Kernel storage belongs to the workspace and survives Electron's changing ports.
const storageKey = "local-scribli-colors";
const readLibrary = (): ColorLibrary => sanitizeColorLibrary(window.scribli.storage?.[storageKey]);

export const openColorPicker = (options: {
    title?: string;
    color?: string;
    onApply: (color: string | undefined) => void | Promise<void>;
    allowReset?: boolean;
}) => {
    const languages = window.scribli.languages;
    let current = resolveColor(options.color || "");
    let busy = false;
    const dialog = new Dialog({title: options.title || languages.color, width: "min(95vw, 440px)", content: `<div class="b3-dialog__content scribli-color-picker">
<div class="scribli-color-picker__inputs"><input type="color" data-color-picker="native" aria-label="${languages.color}">
<label>HEX<input class="b3-text-field" data-color-picker="hex" aria-label="HEX" spellcheck="false" maxlength="32"></label>
${["R", "G", "B"].map(channel => `<label>${channel}<input class="b3-text-field" type="number" min="0" max="255" step="1" data-color-picker="rgb" aria-label="${channel}"></label>`).join("")}</div>
<div class="scribli-color-picker__error ft__error" role="status"></div>
<div class="scribli-color-picker__library"></div>
<div class="fn__hr"></div><div class="fn__flex"><input class="b3-text-field fn__flex-1" data-color-picker="name" maxlength="80" aria-label="${languages.colorPaletteName}" placeholder="${languages.colorPaletteName}"><span class="fn__space"></span><button type="button" class="b3-button b3-button--outline" data-color-picker="create">${languages.colorSavePalette}</button></div>
<div class="fn__hr"></div><div class="fn__flex">${options.allowReset ? `<button type="button" class="b3-button b3-button--outline" data-color-picker="reset">${languages.default}</button>` : ""}<span class="fn__flex-1"></span><button type="button" class="b3-button b3-button--outline" data-color-picker="cancel">${languages.cancel}</button><span class="fn__space"></span><button type="button" class="b3-button" data-color-picker="apply">${languages.apply}</button></div></div>`});
    const root = dialog.element.querySelector<HTMLElement>(".scribli-color-picker");
    const hex = root.querySelector<HTMLInputElement>('[data-color-picker="hex"]');
    const native = root.querySelector<HTMLInputElement>('[data-color-picker="native"]');
    const rgb = Array.from(root.querySelectorAll<HTMLInputElement>('[data-color-picker="rgb"]'));
    const apply = root.querySelector<HTMLButtonElement>('[data-color-picker="apply"]');
    const error = root.querySelector<HTMLElement>('[role="status"]');
    const swatch = (color: string) => `<button type="button" class="scribli-color-picker__swatch" data-color-swatch="${color}" aria-label="${color}" title="${color}" style="background-color:${color}"></button>`;
    const renderLibrary = () => {
        const library = readLibrary();
        const presets = Array.from({length: 14}, (_, index) => resolveColor(`var(--b3-font-color${index + 1})`));
        root.querySelector(".scribli-color-picker__library").innerHTML = `<div class="scribli-color-picker__section">${languages.colorPresets}</div><div class="scribli-color-picker__swatches">${presets.map(swatch).join("")}</div>
<div class="scribli-color-picker__section">${languages.colorRecent}</div><div class="scribli-color-picker__swatches">${library.recent.map(swatch).join("") || `<span class="ft__on-surface">${languages.colorNoRecent}</span>`}</div>
${library.palettes.map(palette => `<div class="scribli-color-picker__section fn__flex"><span class="fn__flex-1">${escapeHtml(palette.name)}</span><button type="button" class="block__icon block__icon--show" data-color-add="${escapeAttr(palette.id)}" aria-label="${languages.colorAddToPalette}"${palette.colors.length >= 16 ? " disabled" : ""}><svg><use xlink:href="#iconAdd"></use></svg></button><button type="button" class="block__icon block__icon--show" data-color-remove="${escapeAttr(palette.id)}" aria-label="${languages.colorDeletePalette}"><svg><use xlink:href="#iconTrashcan"></use></svg></button></div><div class="scribli-color-picker__swatches">${palette.colors.map(swatch).join("")}</div>`).join("")}`;
        root.querySelector<HTMLButtonElement>('[data-color-picker="create"]').disabled = library.palettes.length >= 8;
    };
    const sync = (color: string, updateHex = true) => {
        current = color;
        if (updateHex) { hex.value = color; }
        native.value = color;
        colorRGB(color).forEach((channel, index) => { rgb[index].value = String(channel); });
        error.textContent = "";
        apply.disabled = false;
        root.querySelectorAll<HTMLElement>("[data-color-swatch]").forEach(button => button.classList.toggle("scribli-color-picker__swatch--current", button.dataset.colorSwatch === color));
    };
    const invalid = () => { error.textContent = languages.colorInvalid; apply.disabled = true; };
    const write = async (library: ColorLibrary) => {
        busy = true;
        try {
            const value = sanitizeColorLibrary(library);
            const response = await fetchSyncPost("/api/storage/setLocalStorageVal", {app: Constants.SCRIBLI_APPID, key: storageKey, val: value});
            if (response.code !== 0) { throw new Error(languages.colorSaveFailed); }
            window.scribli.storage[storageKey] = value;
            renderLibrary(); sync(current); return true;
        } catch { error.textContent = languages.colorSaveFailed; return false; }
        finally { busy = false; }
    };
    hex.addEventListener("input", () => { const color = normalizeColor(hex.value); if (color) { sync(color, false); } else { invalid(); } });
    hex.addEventListener("blur", () => { const color = normalizeColor(hex.value); if (color) { sync(color); } });
    rgb.forEach(input => input.addEventListener("input", () => {
        if (rgb.some(channel => channel.value === "" || !/^\d{1,3}$/.test(channel.value) || Number(channel.value) > 255)) { invalid(); return; }
        sync(normalizeColor(`rgb(${rgb.map(channel => channel.value).join(",")})`));
    }));
    native.addEventListener("input", () => sync(native.value));
    root.addEventListener("click", async event => {
        const target = (event.target as HTMLElement).closest<HTMLElement>("button");
        if (!target || busy) { return; }
        if (target.dataset.colorSwatch) { sync(target.dataset.colorSwatch); return; }
        const action = target.dataset.colorPicker;
        if (action === "cancel") { dialog.destroy(); return; }
        const library = readLibrary();
        if (action === "create") {
            if (apply.disabled || library.palettes.length >= 8) { return; }
            const nameInput = root.querySelector<HTMLInputElement>('[data-color-picker="name"]');
            if (!nameInput.value.trim()) { nameInput.focus(); return; }
            library.palettes.push({id: Lute.NewNodeID(), name: nameInput.value.trim(), colors: [current]});
            if (await write(library)) { nameInput.value = ""; }
            return;
        }
        if (target.dataset.colorAdd && !apply.disabled) {
            library.palettes.find(palette => palette.id === target.dataset.colorAdd)?.colors.push(current);
            await write(library); return;
        }
        if (target.dataset.colorRemove) {
            library.palettes = library.palettes.filter(palette => palette.id !== target.dataset.colorRemove);
            await write(library); return;
        }
        if (action === "apply" || action === "reset") {
            if (action === "apply" && apply.disabled) { return; }
            target.setAttribute("disabled", "");
            busy = true;
            try {
                await options.onApply(action === "reset" ? undefined : current);
                if (action === "apply") { library.recent = [current, ...library.recent.filter(color => color !== current)]; if (!await write(library)) { target.removeAttribute("disabled"); return; } }
                dialog.destroy();
            } catch (reason) {
                error.textContent = reason instanceof Error ? reason.message : languages.failed;
                target.removeAttribute("disabled");
            }
            finally { busy = false; }
        }
    });
    renderLibrary(); sync(current);
    return dialog;
};

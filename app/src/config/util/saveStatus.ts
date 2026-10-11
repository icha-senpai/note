import type {SettingSaveState} from "./saveQueue";

const states = new Map<string, {state: SettingSaveState, retry: () => void}>();

export const setSettingSaveState = (namespace: string, state: SettingSaveState, retry: () => void) => {
    states.set(namespace, {state, retry});
    window.dispatchEvent(new Event("scribli-settings-save"));
};

export const bindSettingSaveStatus = (element: HTMLElement) => {
    const update = () => {
        const entries = Array.from(states.values());
        const failed = entries.some(item => item.state === "failed");
        const saving = entries.some(item => item.state === "saving");
        const text = element.querySelector("[data-type=save-status]");
        text.textContent = failed ? window.scribli.languages.settingsSaveFailed :
            saving ? window.scribli.languages.settingsSaving : entries.length ? window.scribli.languages.settingsSaved : "";
        element.querySelector("button").classList.toggle("fn__none", !failed);
    };
    element.querySelector("button").addEventListener("click", () => {
        states.forEach(item => { if (item.state === "failed") { item.retry(); } });
    });
    window.addEventListener("scribli-settings-save", update);
    update();
    return () => window.removeEventListener("scribli-settings-save", update);
};

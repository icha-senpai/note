import {frequentSlashItems, normalizeSlashUsage} from "./frequentSlash";

const key = () => `scribliFrequentSlash:${window.scribli.config.system.workspaceDir}`;
const enabled = () => window.scribli.config.editor.frequentSlash !== false;
const read = () => {
    try { return normalizeSlashUsage(JSON.parse(localStorage.getItem(key()) || "{}")); }
    catch { return {}; }
};

export const withFrequentSlashItems = (items: IHintData[]): IHintData[] => {
    if (!enabled()) { return items; }
    const frequent = frequentSlashItems(items, read());
    return frequent.length ? [...frequent, {html: "separator", value: ""}, ...items] : items;
};

export const recordSlashUsage = (id?: string) => {
    if (!id || !/^[a-zA-Z0-9_-]{1,80}$/.test(id) || !enabled() || window.scribli.config.readonly || window.scribli.config.editor.readOnly || window.scribli.isPublish) { return; }
    const usage = read();
    usage[id] = Math.min((Object.prototype.hasOwnProperty.call(usage, id) ? usage[id] : 0) + 1, Number.MAX_SAFE_INTEGER);
    try { localStorage.setItem(key(), JSON.stringify(usage)); }
    catch { /* Command insertion remains available when local storage is restricted. */ }
};

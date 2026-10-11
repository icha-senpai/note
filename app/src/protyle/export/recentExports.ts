import {Dialog} from "../../dialog";
import {escapeAttr, escapeHtml} from "../../util/escape";
import {useShell} from "../../util/pathName";
import {writeText} from "../util/compatibility";

interface RecentExport {path: string; date: number}
const storageKey = () => `scribliRecentExports:${window.scribli.config.system.workspaceDir}`;

const readExports = (): RecentExport[] => {
    try {
        const items = JSON.parse(localStorage.getItem(storageKey()) || "[]");
        return Array.isArray(items) ? items.filter(item => typeof item.path === "string" && typeof item.date === "number").slice(0, 20) : [];
    } catch { return []; }
};

export const recordExport = (path: string) => {
    const items = [{path, date: Date.now()}, ...readExports().filter(item => item.path !== path)].slice(0, 20);
    try { localStorage.setItem(storageKey(), JSON.stringify(items)); } catch { /* Storage can be unavailable in restricted browser sessions. */ }
};

export const openRecentExports = () => {
    const items = readExports();
    const dialog = new Dialog({
        title: window.scribli.languages.recentExports,
        width: "min(90vw, 760px)",
        content: `<div class="b3-dialog__content b3-list b3-list--background">${items.length ? items.map((item, index) => `<div class="b3-label fn__flex fn__flex-wrap" data-index="${index}">
<div class="fn__flex-1" style="min-width:180px;overflow-wrap:anywhere"><div>${escapeHtml(item.path)}</div><div class="ft__smaller ft__on-surface">${escapeHtml(new Date(item.date).toLocaleString())}</div></div>
<button type="button" class="b3-button b3-button--outline" data-type="folder">${window.scribli.languages.showInFolder}</button><span class="fn__space"></span><button type="button" class="b3-button b3-button--text" data-type="copy">${window.scribli.languages.copyPath}</button></div>`).join("") : `<div class="b3-list--empty">${window.scribli.languages.noRecentExports}</div>`}</div>
<div class="b3-dialog__action"><button type="button" class="b3-button b3-button--cancel" data-type="clear" ${items.length ? "" : "disabled"}>${escapeAttr(window.scribli.languages.clearList)}</button></div>`,
    });
    dialog.element.addEventListener("click", event => {
        const button = (event.target as HTMLElement).closest<HTMLButtonElement>("button[data-type]");
        if (!button) { return; }
        if (button.dataset.type === "clear") {
            localStorage.removeItem(storageKey());
            dialog.destroy();
            openRecentExports();
            return;
        }
        const item = items[Number(button.closest<HTMLElement>("[data-index]")?.dataset.index)];
        if (!item) { return; }
        if (button.dataset.type === "folder") { useShell("showItemInFolder", item.path); }
        if (button.dataset.type === "copy") { void writeText(item.path); }
    });
};

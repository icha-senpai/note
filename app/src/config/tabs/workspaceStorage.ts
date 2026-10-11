import {fetchSyncPost} from "../../util/fetch";
import {escapeHtml} from "../../util/escape";

const formatSize = (bytes: number) => {
    const units = ["B", "KB", "MB", "GB", "TB"];
    let index = 0;
    while (bytes >= 1024 && index < units.length - 1) { bytes /= 1024; index++; }
    return `${bytes.toLocaleString(undefined, {maximumFractionDigits: index ? 2 : 0})} ${units[index]}`;
};

export const mountWorkspaceStorage = (root: HTMLElement) => {
    const panel = root.querySelector<HTMLElement>("[data-type=workspace-storage]");
    const button = panel.querySelector<HTMLButtonElement>("button");
    const output = panel.querySelector<HTMLElement>("[role=status]");
    const load = async () => {
        if (button.disabled) { return; }
        button.disabled = true;
        output.textContent = window.scribli.languages.loading;
        try {
            const response = await fetchSyncPost("/api/system/getWorkspaceStorage", {});
            if (response.code !== 0) { throw new Error(response.msg); }
            if (!panel.isConnected) { return; }
            const stats = response.data as {directories: Record<string, number>; assetsSize: number; totalSize: number; calculatedAt: number};
            const names: Record<string, string> = {data: window.scribli.languages.storageData, repo: window.scribli.languages.dataSnapshot,
                history: window.scribli.languages.history, temp: window.scribli.languages.storageTemporary, conf: window.scribli.languages.config, other: window.scribli.languages.storageOther};
            output.innerHTML = `<div class="fn__flex"><strong class="fn__flex-1">${window.scribli.languages.total}</strong><strong>${formatSize(stats.totalSize)}</strong></div>` +
                ["data", "repo", "history", "temp", "conf", "other"].map(key => `<div class="fn__flex"><span class="fn__flex-1">${escapeHtml(names[key])}</span><span>${formatSize(stats.directories[key] || 0)}</span></div>`).join("") +
                `<div class="fn__hr"></div><div class="ft__smaller">${window.scribli.languages.storageAssetsIncluded}: ${formatSize(stats.assetsSize)}</div><div class="ft__smaller ft__on-surface">${escapeHtml(new Date(stats.calculatedAt).toLocaleString())}</div>`;
        } catch (error) {
            if (panel.isConnected) { output.textContent = error instanceof Error ? error.message : window.scribli.languages.failed; }
        } finally { button.disabled = false; }
    };
    button.addEventListener("click", () => { void load(); });
};

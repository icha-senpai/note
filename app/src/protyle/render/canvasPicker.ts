import {Dialog} from "../../dialog";
import {fetchSyncPost} from "../../util/fetch";
import {Constants} from "../../constants";
import {upDownHint} from "../../util/upDownHint";

interface CanvasReference {value: string; label: string; path?: string}

const plainText = (html: string) => {
    const template = document.createElement("template");
    template.innerHTML = html || "";
    return template.content.textContent.trim();
};

export const pickCanvasReference = (kind: string, blockID?: string): Promise<CanvasReference | null> => new Promise(resolve => {
    let completed = false;
    let requestID = 0;
    let timer: number;
    let items: CanvasReference[] = [];
    const finish = (item: CanvasReference | null) => {
        if (completed) { return; }
        completed = true;
        clearTimeout(timer);
        resolve(item);
        dialog.destroy();
    };
    const title = {document: "canvasDocumentPicker", block: "canvasBlockPicker", database: "database", asset: "assets"}[kind];
    const dialog = new Dialog({
        title: window.scribli.languages[title],
        width: "min(90vw, 720px)",
        content: `<div class="fn__flex-column b3-dialog__content">
<input class="b3-text-field fn__block" data-type="reference-search" aria-label="${window.scribli.languages.search}" placeholder="${window.scribli.languages.canvasSearchOrLink}">
${kind === "asset" ? `<div class="fn__hr"></div><label class="b3-button b3-button--outline" style="position:relative;align-self:flex-start">${window.scribli.languages.browseLocalFile}<input class="b3-form__upload" type="file" data-type="asset-file" aria-label="${window.scribli.languages.browseLocalFile}"></label>` : ""}
<div class="fn__hr"></div><div class="ft__smaller ft__error fn__none" data-type="picker-error" role="alert"></div>
<ul class="b3-list b3-list--background" data-type="reference-results" style="max-height:45vh;overflow:auto"></ul>
</div><div class="b3-dialog__action"><button type="button" class="b3-button b3-button--cancel" data-type="cancel">${window.scribli.languages.cancel}</button></div>`,
        destroyCallback() { if (!completed) { completed = true; clearTimeout(timer); resolve(null); } },
    });
    const input = dialog.element.querySelector<HTMLInputElement>('[data-type="reference-search"]');
    const list = dialog.element.querySelector<HTMLUListElement>('[data-type="reference-results"]');
    const error = dialog.element.querySelector<HTMLElement>('[data-type="picker-error"]');
    dialog.element.dataset.key = "canvas-reference-picker";
    const search = async () => {
        const token = ++requestID;
        let matches: CanvasReference[] = [];
        error.classList.add("fn__none");
        const query = input.value.trim();
        list.textContent = window.scribli.languages.searching;
        try {
            let response: IWebSocketData;
            const blockLink = query.match(/^(?:scribli:\/\/blocks\/)?(\d{14}-[a-z0-9]{7})(?:[?#].*)?$/);
            if (blockLink && ["document", "block"].includes(kind)) {
                response = await fetchSyncPost("/api/block/getBlockInfo", {id: blockLink[1]});
                if (response.code !== 0) { throw new Error(response.msg); }
                const id = kind === "document" ? response.data.rootID : blockLink[1];
                const info = await fetchSyncPost("/api/block/getBlockDOM", {id});
                if (info.code !== 0) { throw new Error(info.msg); }
                matches = [{value: id, label: (kind === "document" || id === response.data.rootID ? response.data.rootTitle : "") || plainText(info.data.dom).slice(0, 160) || id, path: response.data.rootTitle}];
            } else if (kind === "database") {
                response = await fetchSyncPost("/api/av/searchAttributeView", {keyword: query});
                if (response.code !== 0) { throw new Error(response.msg); }
                matches = (response.data.results || []).map((item: {avID: string, avName: string, hPath: string}) => ({value: item.avID, label: item.avName || window.scribli.languages.database, path: item.hPath}));
            } else if (kind === "asset") {
                response = await fetchSyncPost("/api/search/searchAsset", {k: query});
                if (response.code !== 0) { throw new Error(response.msg); }
                matches = (response.data || []).map((item: {path: string, hName: string}) => ({value: item.path, label: item.hName || item.path.split("/").pop(), path: item.path}));
            } else {
                response = await fetchSyncPost("/api/search/fullTextSearchBlock", {
                    query, page: 1, pageSize: 30, method: 0, groupBy: 0,
                    ...(kind === "document" ? {types: {document: true}} : {}),
                });
                if (response.code !== 0) { throw new Error(response.msg); }
                matches = (response.data.blocks || []).map((item: {id: string, content: string, hPath: string}) => ({value: item.id, label: plainText(item.content) || item.hPath?.split("/").pop() || item.id, path: item.hPath}));
            }
            if (token !== requestID || completed) { return; }
            items = matches;
            list.replaceChildren();
            items.forEach((item, index) => {
                const row = document.createElement("li");
                row.className = "b3-list-item" + (index === 0 ? " b3-list-item--focus" : "");
                row.dataset.index = String(index);
                row.tabIndex = 0;
                row.setAttribute("role", "button");
                const content = document.createElement("div");
                content.className = "b3-list-item--two fn__flex-1";
                const label = document.createElement("div");
                label.className = "b3-list-item__text";
                label.textContent = item.label;
                const path = document.createElement("div");
                path.className = "b3-list-item__meta";
                path.textContent = item.path || "";
                content.append(label, path);
                row.append(content);
                row.addEventListener("click", () => finish(item));
                row.addEventListener("keydown", event => {
                    if (event.key === "Enter" || event.key === " ") { event.preventDefault(); event.stopPropagation(); finish(item); }
                });
                list.append(row);
            });
            if (items.length === 0) {
                const empty = document.createElement("li");
                empty.className = "b3-list--empty";
                empty.textContent = window.scribli.languages.canvasNoMatches;
                list.append(empty);
            }
        } catch (failure) {
            if (token !== requestID || completed) { return; }
            list.replaceChildren();
            error.textContent = failure.message || window.scribli.languages.canvasSearchFailed;
            error.classList.remove("fn__none");
        }
    };
    input.addEventListener("input", event => {
        if ((event as InputEvent).isComposing) { return; }
        ++requestID;
        list.replaceChildren();
        items = [];
        clearTimeout(timer);
        timer = window.setTimeout(() => { void search(); }, Constants.TIMEOUT_INPUT);
    });
    input.addEventListener("compositionend", () => { clearTimeout(timer); void search(); });
    input.addEventListener("keydown", event => {
        event.stopPropagation();
        if (event.isComposing) { return; }
        upDownHint(list, event);
        if (event.key === "Enter") {
            event.preventDefault();
            const row = list.querySelector<HTMLElement>(".b3-list-item--focus");
            if (row) { finish(items[Number(row.dataset.index)]); }
        }
        if (event.key === "Escape") { finish(null); }
    });
    dialog.element.querySelector('[data-type="cancel"]').addEventListener("click", () => finish(null));
    dialog.element.querySelector<HTMLInputElement>('[data-type="asset-file"]')?.addEventListener("change", async event => {
        const file = (event.target as HTMLInputElement).files?.[0];
        if (!file) { return; }
        error.classList.add("fn__none");
        input.disabled = true;
        try {
            const data = new FormData();
            data.append("file[]", file);
            if (blockID) { data.append("id", blockID); }
            const response = await fetchSyncPost(Constants.UPLOAD_ADDRESS, data);
            const assetPath = response.data?.succMap?.[file.name] as string;
            if (response.code !== 0 || !assetPath) { throw new Error(response.msg || window.scribli.languages.canvasAssetImportFailed); }
            finish({value: assetPath, label: file.name});
        } catch (failure) {
            error.textContent = failure.message;
            error.classList.remove("fn__none");
        } finally { input.disabled = false; }
    });
    input.focus();
    void search();
});

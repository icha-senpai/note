import {Constants} from "../../constants";
import {Dialog} from "../../dialog";
import {confirmDialog} from "../../dialog/confirmDialog";
import {showMessage} from "../../dialog/message";
import {fetchPost, fetchSyncPost} from "../../util/fetch";
import {isMobile} from "../../util/functions";
import {isEncryptedBox} from "../../util/pathName";
import {saveExportFile} from "../util/compatibility";


type BoolKey = "addTitle" | "inlineMemo" | "includeSubDocs" | "includeRelatedDocs" | "markdownYFM" | "removeAssetsID";
type IntKey = "blockRefMode" | "blockEmbedMode" | "fileAnnotationRefMode";
type StringKey = "blockRefTextLeft" | "blockRefTextRight" | "tagOpenMarker" | "tagCloseMarker";

interface IExportMdOptions {
    id?: string;
    ids?: string[];
    notebook?: string;
}

export const openExportOptionsDialog = (onConfirm: (options: IExportMdOptionsPayload) => void, showSubDocs = true, showRelatedDocs = true) => {
    const conf = window.scribli.config.export;
    const bool = (id: BoolKey) => `<input id="${id}" class="b3-switch fn__flex-center" type="checkbox" ${conf[id] ? "checked" : ""}>`;
    const select = (id: IntKey, options: {value: number; label: string}[]) => {
        const opts = options.map(o =>
            `<option value="${o.value}" ${conf[id] === o.value ? "selected" : ""}>${o.label}</option>`).join("");
        return `<select id="${id}" class="b3-select fn__flex-center fn__size200">${opts}</select>`;
    };
    const row = (title: string, desc: string, control: string) =>
        `<label class="fn__flex b3-label config-item config-wrap">
            <div class="fn__flex-1">
                <div class="config-name">${title}</div>
                <div class="b3-label__text">${desc}</div>
            </div>
            <span class="fn__space"></span>
            ${control}
        </label>`;
    const textPair = (leftId: StringKey, rightId: StringKey) =>
        `<input id="${leftId}" class="b3-text-field fn__flex-center fn__size96" value="${conf[leftId] ?? ""}">
        <span class="fn__space"></span>
        <input id="${rightId}" class="b3-text-field fn__flex-center fn__size96" value="${conf[rightId] ?? ""}">`;

    const dialog = new Dialog({
        title: window.scribli.languages.export + " Markdown",
        content: `<div class="b3-dialog__content export-md__content">
    ${row(window.scribli.languages.export17, window.scribli.languages.export18, bool("addTitle"))}
    ${showSubDocs ? row(window.scribli.languages.includeSubDocs, window.scribli.languages.includeSubDocsTip, bool("includeSubDocs")) : ""}
    ${showRelatedDocs ? row(window.scribli.languages.includeRelatedDocs, window.scribli.languages.includeRelatedDocsTip, bool("includeRelatedDocs")) : ""}
    ${row(window.scribli.languages.export23, window.scribli.languages.export24, bool("markdownYFM"))}
    ${row(window.scribli.languages.removeAssetsID, window.scribli.languages.removeAssetsIDTip, bool("removeAssetsID"))}
    ${row(window.scribli.languages.export31, window.scribli.languages.export32, bool("inlineMemo"))}
    ${row(window.scribli.languages.ref, window.scribli.languages.export11,
        select("blockRefMode", [
            {value: 2, label: window.scribli.languages.export2},
            {value: 3, label: window.scribli.languages.export3},
            {value: 4, label: window.scribli.languages.export4},
        ]))}
    ${row(window.scribli.languages.blockEmbed, window.scribli.languages.export12,
        select("blockEmbedMode", [
            {value: 0, label: window.scribli.languages.export0},
            {value: 1, label: window.scribli.languages.export1},
        ]))}
    ${row(window.scribli.languages.export5, window.scribli.languages.export6,
        select("fileAnnotationRefMode", [
            {value: 0, label: window.scribli.languages.export7},
            {value: 1, label: window.scribli.languages.export8},
        ]))}
    ${row(window.scribli.languages.export13, window.scribli.languages.export14,
        textPair("blockRefTextLeft", "blockRefTextRight"))}
    ${row(window.scribli.languages.export15, window.scribli.languages.export16,
        textPair("tagOpenMarker", "tagCloseMarker"))}
</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages.confirm}</button>
</div>`,
        width: "520px",
        height: isMobile() ? "70vh" : "60vh",
    });
    dialog.element.setAttribute("data-key", Constants.DIALOG_EXPORTMARKDOWN);

    const el = dialog.element;
    const boolVal = (id: BoolKey) => {
        const input = el.querySelector("#" + id) as HTMLInputElement;
        return input ? input.checked : conf[id];
    };
    const collect = (): IExportMdOptionsPayload => ({
        addTitle: (el.querySelector("#addTitle") as HTMLInputElement).checked,
        inlineMemo: (el.querySelector("#inlineMemo") as HTMLInputElement).checked,
        blockRefMode: parseInt((el.querySelector("#blockRefMode") as HTMLSelectElement).value, 10),
        blockEmbedMode: parseInt((el.querySelector("#blockEmbedMode") as HTMLSelectElement).value, 10),
        fileAnnotationRefMode: parseInt((el.querySelector("#fileAnnotationRefMode") as HTMLSelectElement).value, 10),
        blockRefTextLeft: (el.querySelector("#blockRefTextLeft") as HTMLInputElement).value,
        blockRefTextRight: (el.querySelector("#blockRefTextRight") as HTMLInputElement).value,
        tagOpenMarker: (el.querySelector("#tagOpenMarker") as HTMLInputElement).value,
        tagCloseMarker: (el.querySelector("#tagCloseMarker") as HTMLInputElement).value,
        includeSubDocs: boolVal("includeSubDocs"),
        includeRelatedDocs: boolVal("includeRelatedDocs"),
        markdownYFM: (el.querySelector("#markdownYFM") as HTMLInputElement).checked,
        removeAssetsID: (el.querySelector("#removeAssetsID") as HTMLInputElement).checked,
    });

    const btnsElement = el.querySelectorAll(".b3-button");
    btnsElement[0].addEventListener("click", () => {
        dialog.destroy();
    });
    btnsElement[1].addEventListener("click", () => {
        const payload = collect();
        dialog.destroy();
        onConfirm(payload);
    });
};

interface IExportMdOptionsPayload {
    addTitle: boolean;
    inlineMemo: boolean;
    blockRefMode: number;
    blockEmbedMode: number;
    fileAnnotationRefMode: number;
    blockRefTextLeft: string;
    blockRefTextRight: string;
    tagOpenMarker: string;
    tagCloseMarker: string;
    includeSubDocs: boolean;
    includeRelatedDocs: boolean;
    markdownYFM: boolean;
    removeAssetsID: boolean;
}

export const exportMarkdownZip = async(options: IExportMdOptions) => {
    let showSubDocs = true;
    let showRelatedDocs = true;
    let encrypted = false;
    if (options.id) {
        const docInfo = await fetchSyncPost("/api/block/getDocInfo", {id: options.id});
        const data = docInfo.data;
        showSubDocs = 0 < data.subFileCount;
        showRelatedDocs = 0 < (data.refCount || 0) || 0 < (data.attrViews?.length || 0);
        const blockInfo = await fetchSyncPost("/api/block/getBlockInfo", {id: options.id});
        encrypted = blockInfo.code === 0 && isEncryptedBox(blockInfo.data.box);
    }
    openExportOptionsDialog(params => {
        const exportMarkdown = () => {
        const msgId = showMessage(window.scribli.languages.exporting, -1);
        const cb = (response: IWebSocketData) => saveExportFile(response.data.zip, msgId);
        if (options.id) {
            fetchPost("/api/export/exportMd", {id: options.id, ...params}, cb);
        } else if (options.ids) {
            fetchPost("/api/export/exportMds", {ids: options.ids, ...params}, cb);
        } else {
            fetchPost("/api/export/exportNotebookMd", {notebook: options.notebook, ...params}, cb);
        }
        };
        if (encrypted) {
            confirmDialog("⚠️ " + window.scribli.languages.export, window.scribli.languages.encryptedExportRiskTip, exportMarkdown);
            return;
        }
        exportMarkdown();
    }, showSubDocs, showRelatedDocs);
};

import {fetchPost} from "../util/fetch";
import {getDisplayName, getNotebookName, isEncryptedBox} from "../util/pathName";
import {confirmDialog} from "../dialog/confirmDialog";
import {hasTopClosestByTag} from "../protyle/util/hasClosest";
import {showMessage} from "../dialog/message";
import {escapeHtml} from "../util/escape";

export const deleteFile = (notebookId: string, pathString: string) => {
    if (window.scribli.config.fileTree.removeDocWithoutConfirm) {
        fetchPost("/api/filetree/removeDoc", {
            notebook: notebookId,
            path: pathString
        });
        return;
    }
    const docInfoParam: IObject = {
        id: getDisplayName(pathString, true, true)
    };
    if (isEncryptedBox(notebookId)) {
        docInfoParam.notebook = notebookId;
    }
    fetchPost("/api/block/getDocInfo", docInfoParam, (response) => {
        const fileName = escapeHtml(response.data.name);
        let tip = `${window.scribli.languages.confirmDeleteTip.replace("${x}", fileName)}
<div class="fn__hr"></div>
<div class="ft__smaller ft__on-surface">${window.scribli.languages.rollbackTip.replace("${x}", window.scribli.config.editor.historyRetentionDays)}</div>`;
        if (response.data.subFileCount > 0) {
            tip = `${window.scribli.languages.andSubFile.replace("${x}", fileName).replace("${y}", response.data.subFileCount)}
<div class="fn__hr"></div>
<div class="ft__smaller ft__on-surface">${window.scribli.languages.rollbackTip.replace("${x}", window.scribli.config.editor.historyRetentionDays)}</div>`;
        }
        confirmDialog(window.scribli.languages.deleteOpConfirm, tip, () => {
            fetchPost("/api/filetree/removeDoc", {
                notebook: notebookId,
                path: pathString
            });
        }, undefined, true);
    });
};

export const deleteFiles = (liElements: Element[]) => {
    if (liElements.length === 1) {
        const itemTopULElement = hasTopClosestByTag(liElements[0], "UL");
        if (itemTopULElement) {
            const itemNotebookId = itemTopULElement.getAttribute("data-url");
            if (liElements[0].getAttribute("data-type") === "navigation-file") {
                deleteFile(itemNotebookId, liElements[0].getAttribute("data-path"));
            } else {
                confirmDialog(window.scribli.languages.deleteOpConfirm,
                    `${window.scribli.languages.confirmDeleteTip.replace("${x}", Lute.EscapeHTMLStr(getNotebookName(itemNotebookId)))}
<div class="fn__hr"></div>
<div class="ft__smaller ft__on-surface">${window.scribli.languages.rollbackTip.replace("${x}", window.scribli.config.editor.historyRetentionDays)}</div>`, () => {
                        fetchPost("/api/notebook/removeNotebook", {
                            notebook: itemNotebookId,
                        });
                    }, undefined, true);
            }
        }
    } else {
        const paths: string[] = [];
        liElements.forEach(item => {
            const dataPath = item.getAttribute("data-path");
            if (dataPath !== "/") {
                paths.push(item.getAttribute("data-path"));
            }
        });
        if (paths.length === 0) {
            showMessage(window.scribli.languages.notBatchRemove);
            return;
        }
        confirmDialog(window.scribli.languages.deleteOpConfirm,
            `${window.scribli.languages.confirmRemoveAll.replace("${count}", paths.length)}
<div class="fn__hr"></div>
<div class="ft__smaller ft__on-surface">${window.scribli.languages.rollbackTip.replace("${x}", window.scribli.config.editor.historyRetentionDays)}</div>`, () => {
                fetchPost("/api/filetree/removeDocs", {
                    paths
                });
            }, undefined, true);
    }
};

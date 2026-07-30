import {showMessage} from "../dialog/message";
import {hasTopClosestByTag} from "../protyle/util/hasClosest";
import {Files} from "../layout/dock/Files";
import {Editor} from "../editor";
import {openFileById} from "../editor/util";
import {getActiveTab, getDockByType} from "../layout/tabUtil";
import {fetchPost} from "./fetch";
import {getDisplayName, getOpenNotebookCount, pathPosix} from "./pathName";
import {Constants} from "../constants";
import {replaceFileName, validateName} from "../editor/rename";
import {hideElements} from "../protyle/ui/hideElements";
import {App} from "../index";
import {NewDocTargetByHPath, NewDocTargetSubDoc, getNewDocTargetFromSavePath, getNewDocTargetFromTree} from "./parseNewDocTarget";

export const getBlockRefAnchorText = (title: string) => {
    const trimmed = (title || "").trim();
    if (!trimmed) {
        return window.scribli.languages._kernel[16];
    }
    return trimmed.substring(0, window.scribli.config.editor.blockRefDynamicAnchorTextMaxLen);
};

type NewDocRequest = {
    app: App;
    notebookId: string;
    currentPath: string;
    hasFocusTarget: boolean;
    name?: string;
    paths?: string[];
    listDocTree?: boolean;
    onCreated?: (id: string, title: string) => void;
};

export const newFile = (app: App, name?: string) => {
    if (getOpenNotebookCount() === 0) {
        showMessage(window.scribli.languages.newFileTip);
        return;
    }
    const {notebookId, currentPath, hasFocusTarget} = getNewFilePath();
    if (name === undefined) {
        runNewDoc({
            app,
            notebookId,
            currentPath,
            hasFocusTarget,
        });
    } else {
        runNewDoc({
            app,
            notebookId,
            currentPath,
            hasFocusTarget,
            name: replaceFileName(name.trim()),
            onCreated: () => hideElements(["dialog"]),
        });
    }
};

export const newFileInProtyle = (protyle: IProtyle, onCreated: (id: string, title: string) => void) => {
    runNewDoc({
        app: protyle.app,
        notebookId: protyle.notebookId,
        currentPath: protyle.path,
        hasFocusTarget: true,
        onCreated,
    });
};

export const newFileInTree = (app: App, notebookId: string, currentPath: string, paths?: string[]) => {
    runNewDocInTree({
        app,
        notebookId,
        currentPath,
        hasFocusTarget: true,
        paths,
        listDocTree: true,
    });
};

export const newFileBySelect = (protyle: IProtyle, selectText: string, nodeElement: HTMLElement, pathDir: string, targetNotebookId: string) => {
    const newFileName = replaceFileName(selectText.trim() ? selectText.trim() : protyle.lute.BlockDOM2Content(nodeElement.outerHTML).replace(/\n/g, "").trim());
    const hPath = pathPosix().join(pathDir, newFileName || window.scribli.languages._kernel[16]);
    fetchPost("/api/filetree/getIDsByHPath", {
        path: hPath,
        notebook: targetNotebookId
    }, (idResponse) => {
        const refText = getBlockRefAnchorText(newFileName);
        if (idResponse.data && idResponse.data.length > 0) {
            const refElement = protyle.toolbar.setInlineMark(protyle, "block-ref", "range", {
                type: "id",
                color: `${idResponse.data[0]}${Constants.ZWSP}d${Constants.ZWSP}${refText}`
            });
            if (refElement[0]) {
                protyle.toolbar.range.selectNodeContents(refElement[0]);
            }
        } else {
            fetchPost("/api/filetree/createDocWithMd", {
                notebook: targetNotebookId,
                path: hPath,
                parentID: protyle.notebookId === targetNotebookId ? protyle.block.rootID : "",
                markdown: "",
                titleEmpty: newFileName === "",
            }, response => {
                const refElement = protyle.toolbar.setInlineMark(protyle, "block-ref", "range", {
                    type: "id",
                    color: `${response.data}${Constants.ZWSP}d${Constants.ZWSP}${refText}`
                });
                if (refElement[0]) {
                    protyle.toolbar.range.selectNodeContents(refElement[0]);
                }
            });
        }
        hideElements(["toolbar"], protyle);
    });
};

export const getRefCreateSavePath = (notebookId: string, currentPath: string, cb: (targetNotebookId: string, hPath: string) => void) => {
    fetchPost("/api/filetree/getRefCreateSavePath", {
        notebook: notebookId
    }, (data) => {
        let targetPath = currentPath;
        if (notebookId !== data.data.box) {
            targetPath = data.data.path || "/";
        }
        if (data.data.path) {
            if (data.data.path.startsWith("/")) {
                cb(data.data.box, getDisplayName(data.data.path, false, true));
            } else {
                fetchPost("/api/filetree/getHPathByPath", {
                    notebook: data.data.box,
                    path: targetPath
                }, (response) => {
                    cb(data.data.box, getDisplayName(pathPosix().join(response.data, data.data.path), false, true));
                });
            }
        } else {
            fetchPost("/api/filetree/getHPathByPath", {
                notebook: data.data.box,
                path: targetPath
            }, (response) => {
                cb(data.data.box, getDisplayName(response.data, false, true));
            });
        }
    });
};

function getNewFilePath(): Pick<NewDocRequest, "notebookId" | "currentPath" | "hasFocusTarget"> {
    let notebookId = "";
    let currentPath = "";
    let hasFocusTarget = false;
    const tab = getActiveTab(false);
    if (tab?.model instanceof Editor) {
        notebookId = tab.model.editor.protyle.notebookId;
        currentPath = tab.model.editor.protyle.path;
        hasFocusTarget = true;
    }
    if (!notebookId) {
        const fileModel = getDockByType("file").data.file;
        if (fileModel instanceof Files) {
            const currentElement = fileModel.element.querySelector(".b3-list-item--focus");
            if (currentElement) {
                const topElement = hasTopClosestByTag(currentElement, "UL");
                if (topElement) {
                    notebookId = topElement.getAttribute("data-url");
                }
                currentPath = currentElement.getAttribute("data-path");
                hasFocusTarget = true;
            }
        }
    }
    if (!notebookId) {
        const openNotebook = window.scribli.notebooks.find(item => !item.closed);
        if (openNotebook) {
            notebookId = openNotebook.id;
            currentPath = "/";
        }
    }
    return {notebookId, currentPath, hasFocusTarget};
}

function runNewDoc(request: NewDocRequest) {
    fetchPost("/api/filetree/getDocCreateSavePath", {notebook: request.notebookId}, (savePathResponse) => {
        const templatePath = savePathResponse.data.path as string;
        const targetNotebookId = savePathResponse.data.box as string;
        getNewDocHPath(targetNotebookId, request.notebookId, request.currentPath, (hPath) => {
            createNewDoc(request, templatePath, targetNotebookId, hPath);
        });
    });
}

function getNewDocHPath(targetNotebookId: string, currentNotebookId: string, currentPath: string, callback: (hPath: string) => void) {
    if (targetNotebookId !== currentNotebookId) {
        callback("/");
        return;
    }
    fetchPost("/api/filetree/getHPathByPath", {
        notebook: targetNotebookId,
        path: currentPath,
    }, (hPathResponse) => {
        callback(hPathResponse.data);
    });
}

function createNewDoc(request: NewDocRequest, templatePath: string, targetNotebookId: string, hPath: string) {
    const target = getNewDocTargetFromSavePath({
        templatePath,
        hPath: hPath || "/",
        targetNotebookId,
        currentNotebookId: request.notebookId,
        name: request.name,
        hasFocusTarget: request.hasFocusTarget,
        currentPath: request.currentPath,
    });
    if (target.kind === "hPath") {
        createNewDocByHPath(request, target);
    } else if (target.kind === "subDoc") {
        createNewDocAsSubDoc(request, target);
    }
}

function runNewDocInTree(request: NewDocRequest) {
    fetchPost("/api/filetree/getDocCreateSavePath", {notebook: request.notebookId}, (savePathResponse) => {
        const target = getNewDocTargetFromTree({
            templatePath: savePathResponse.data.path as string,
            currentNotebookId: request.notebookId,
            currentPath: request.currentPath,
            name: request.name,
        });
        createNewDocAsSubDoc(request, target);
    });
}

function getCreateDocParentID(hasFocusTarget: boolean, notebookId: string, currentPath: string, targetNotebookId: string): string | undefined {
    return hasFocusTarget && notebookId === targetNotebookId && currentPath !== "/"
        ? getDisplayName(currentPath, true, true)
        : undefined;
}

function createNewDocByHPath(request: NewDocRequest, target: NewDocTargetByHPath) {
    if (target.title && !validateName(target.title)) {
        return;
    }
    const parentID = getCreateDocParentID(request.hasFocusTarget, request.notebookId, request.currentPath, target.targetNotebookId);
    fetchPost("/api/filetree/createDocWithMd", {
        notebook: target.targetNotebookId,
        path: target.hPath,
        parentID,
        markdown: "",
        titleEmpty: !target.title,
    }, (response) => {
        openCreatedDoc(request.app, response.data, request.onCreated, target.title);
    });
}

function createNewDocAsSubDoc(request: NewDocRequest, target: NewDocTargetSubDoc) {
    const id = Lute.NewNodeID();
    const newPath = pathPosix().join(getDisplayName(target.parentPath, false, true), id + ".sy");
    if (request.paths) {
        request.paths[request.paths.indexOf(undefined)] = newPath;
    }
    fetchPost("/api/filetree/createDoc", {
        notebook: target.targetNotebookId,
        path: newPath,
        title: target.title,
        md: "",
        sorts: request.paths,
        listDocTree: request.listDocTree,
    }, () => {
        openCreatedDoc(request.app, id, request.onCreated, target.title);
    });
}

function openCreatedDoc(app: App, id: string, onCreated?: (id: string, title: string) => void, title?: string) {
    if (onCreated) {
        onCreated(id, title || "");
    }
    openFileById({
        app,
        id,
        action: [Constants.CB_GET_CONTEXT, Constants.CB_GET_OPENNEW]
    });
}

/**
 * 
 */
export const newFileByRefHint = (
    protyle: IProtyle,
    name: string,
    onCreated?: (id: string, title: string) => void,
    presetId?: string,
) => {
    const requestName = replaceFileName(name.trim());
    fetchPost("/api/filetree/getRefCreateSavePath", {notebook: protyle.notebookId}, (savePathResponse) => {
        const templatePath = savePathResponse.data.path as string;
        const targetNotebookId = savePathResponse.data.box as string;
        getNewDocHPath(targetNotebookId, protyle.notebookId, protyle.path, (hPath) => {
            const target = getNewDocTargetFromSavePath({
                templatePath,
                hPath: hPath || "/",
                targetNotebookId,
                currentNotebookId: protyle.notebookId,
                name: requestName,
                hasFocusTarget: true,
                currentPath: protyle.path,
            });
            if (target.kind === "hPath") {
                createRefDocByHPath(protyle, target, onCreated, presetId);
            } else {
                createRefDocAsSubDoc(target, onCreated, presetId);
            }
        });
    });
};

function createRefDocByHPath(
    protyle: IProtyle,
    target: NewDocTargetByHPath,
    onCreated?: (id: string, title: string) => void,
    presetId?: string,
) {
    if (target.title && !validateName(target.title)) {
        return;
    }
    const parentID = getCreateDocParentID(true, protyle.notebookId, protyle.path, target.targetNotebookId);
    fetchPost("/api/filetree/createDocWithMd", {
        notebook: target.targetNotebookId,
        path: target.hPath,
        parentID,
        markdown: "",
        titleEmpty: !target.title,
        id: presetId,
    }, (response) => {
        onCreated?.(response.data, target.title || "");
    });
}

function createRefDocAsSubDoc(
    target: NewDocTargetSubDoc,
    onCreated?: (id: string, title: string) => void,
    presetId?: string,
) {
    const id = presetId || Lute.NewNodeID();
    const newPath = pathPosix().join(getDisplayName(target.parentPath, false, true), id + ".sy");
    fetchPost("/api/filetree/createDoc", {
        notebook: target.targetNotebookId,
        path: newPath,
        title: target.title,
        md: "",
    }, () => {
        onCreated?.(id, target.title || "");
    });
}

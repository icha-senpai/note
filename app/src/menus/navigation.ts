import {copySubMenu, exportMd, movePathToMenu, openFileAttr, renameMenu,} from "./commonMenuItem";
/// #if !BROWSER
import {FileFilter, ipcRenderer} from "electron";
import * as path from "path";
/// #endif
import {MenuItem} from "./Menu";
import {getDisplayName, getNotebookName, getTopPaths, isEncryptedBox, pathPosix, useShell} from "../util/pathName";
import {showMessage} from "../dialog/message";
import {confirmDialog} from "../dialog/confirmDialog";
import {fetchPost, fetchSyncPost} from "../util/fetch";
import {onGetnotebookconf} from "./onGetnotebookconf";
import {openSearch} from "../search/spread";
import {Constants} from "../constants";
import {newFileInTree} from "../util/newFile";
import {hasClosestByTag, hasTopClosestByTag} from "../protyle/util/hasClosest";
import {deleteFiles} from "../editor/deleteFile";
import {openFileById} from "../editor/util";
import {getDockByType} from "../layout/tabUtil";
import {Files} from "../layout/dock/Files";
import {openCardByData} from "../card/openCard";
import {viewCards} from "../card/viewCards";
import {App} from "../index";
import {openDocHistory} from "../history/doc";
import {openEditorTab} from "./util";
import {makeCard} from "../card/makeCard";
import {transaction} from "../protyle/wysiwyg/transaction";
import {emitOpenMenu} from "../plugin/EventBus";
import {saveExportFile} from "../protyle/util/compatibility";
import {exportMarkdownZip} from "../protyle/export/exportMd";
import {addFilesToDatabase} from "../protyle/render/av/addToDatabase";
import {openEmojiPanel} from "../emoji";

const confirmEncryptedExport = (notebookId: string, callback: () => void) => {
    if (!isEncryptedBox(notebookId)) {
        callback();
        return;
    }
    confirmDialog(window.scribli.languages.export, window.scribli.languages.encryptedExportRiskTip, callback);
};

const initMultiMenu = (selectItemElements: NodeListOf<Element>, app: App) => {
    window.scribli.menus.menu.element.setAttribute("data-from", Constants.MENU_FROM_DOC_TREE_MORE_ITEMS);
    const fileItemElement = Array.from(selectItemElements).find(item => {
        if (item.getAttribute("data-type") === "navigation-file") {
            return true;
        }
    });
    if (!fileItemElement) {
        return window.scribli.menus.menu;
    }
    const blockIDs: string[] = [];
    const notebookId = fileItemElement.parentElement?.getAttribute("data-url") || "";
    selectItemElements.forEach(item => {
        const id = item.getAttribute("data-node-id");
        if (id) {
            blockIDs.push(id);
        }
    });

    if (blockIDs.length > 0) {
        window.scribli.menus.menu.append(new MenuItem({
            id: "copy",
            label: window.scribli.languages.copy,
            type: "submenu",
            icon: "iconCopy",
            submenu: copySubMenu(blockIDs).concat([{
                id: "duplicate",
                iconHTML: "",
                label: window.scribli.languages.duplicate,
                accelerator: window.scribli.config.keymap.editor.general.duplicate.custom,
                click() {
                    blockIDs.forEach((id) => {
                        fetchPost("/api/filetree/duplicateDoc", {
                            id
                        });
                    });
                }
            }])
        }).element);
    }

    window.scribli.menus.menu.append(movePathToMenu(getTopPaths(
        Array.from(selectItemElements)
    )));

    if (blockIDs.length > 0) {
        window.scribli.menus.menu.append(new MenuItem({
            id: "addToDatabase",
            label: window.scribli.languages.addToDatabase,
            accelerator: window.scribli.config.keymap.general.addToDatabase.custom,
            icon: "iconDatabase",
            click: () => {
                addFilesToDatabase(Array.from(selectItemElements));
            }
        }).element);
    }
    window.scribli.menus.menu.append(new MenuItem({
        id: "delete",
        icon: "iconTrashcan",
        label: window.scribli.languages.delete,
        accelerator: "⌦",
        click: () => {
            deleteFiles(Array.from(selectItemElements));
        }
    }).element);

    if (blockIDs.length === 0) {
        return window.scribli.menus.menu;
    }
    window.scribli.menus.menu.append(new MenuItem({id: "separator_1", type: "separator"}).element);
    if (!window.scribli.config.readonly) {
        const riffCardMenu = [{
            id: "quickMakeCard",
            iconHTML: "",
            accelerator: window.scribli.config.keymap.editor.general.quickMakeCard.custom,
            label: window.scribli.languages.quickMakeCard,
            click: () => {
                transaction(undefined, [{
                    action: "addFlashcards",
                    deckID: Constants.QUICK_DECK_ID,
                    blockIDs,
                }]);
            }
        }, {
            id: "removeCard",
            iconHTML: "",
            label: window.scribli.languages.removeCard,
            click: () => {
                transaction(undefined, [{
                    action: "removeFlashcards",
                    deckID: Constants.QUICK_DECK_ID,
                    blockIDs,
                }]);
            }
        }];
        if (window.scribli.config.flashcard.deck) {
            riffCardMenu.push({
                id: "addToDeck",
                iconHTML: "",
                label: window.scribli.languages.addToDeck,
                click: () => {
                    makeCard(app, blockIDs);
                }
            });
        }
        window.scribli.menus.menu.append(new MenuItem({
            id: "riffCard",
            label: window.scribli.languages.riffCard,
            icon: "iconRiffCard",
            submenu: riffCardMenu,
        }).element);
        window.scribli.menus.menu.append(new MenuItem({id: "separator_2", type: "separator"}).element);
    }
    openEditorTab(app, blockIDs);
    window.scribli.menus.menu.append(new MenuItem({
        id: "export",
        label: window.scribli.languages.export,
        type: "submenu",
        icon: "iconUpload",
        submenu: [{
            id: "exportScribliZip",
            label: "Scribli .sy.zip",
            icon: "iconScribli",
            click: () => {
                confirmEncryptedExport(notebookId, () => {
                    const msgId = showMessage(window.scribli.languages.exporting, -1);
                    fetchPost("/api/export/exportSYs", {
                        ids: blockIDs,
                    }, response => {
                        saveExportFile(response.data.zip, msgId);
                    });
                });
            }
        }, {
            id: "exportMarkdown",
            label: "Markdown .zip",
            icon: "iconMarkdown",
            click: () => {
                confirmEncryptedExport(notebookId, () => exportMarkdownZip({ids: blockIDs}));
            }
        }]
    }).element);
    if (app.plugins) {
        emitOpenMenu({
            plugins: app.plugins,
            type: "open-menu-doctree",
            detail: {
                elements: selectItemElements,
                type: "docs"
            },
            separatorPosition: "top",
        });
    }
    return window.scribli.menus.menu;
};

export const initNavigationMenu = (app: App, liElement: HTMLElement) => {
    window.scribli.menus.menu.remove();
    window.scribli.menus.menu.element.setAttribute("data-name", Constants.MENU_DOC_TREE_MORE);
    const fileElement = hasClosestByTag(liElement, "DIV");
    if (!fileElement) {
        return window.scribli.menus.menu;
    }
    if (!liElement.classList.contains("b3-list-item--focus")) {
        fileElement.querySelectorAll(".b3-list-item--focus").forEach(item => {
            item.classList.remove("b3-list-item--focus");
            item.removeAttribute("select-end");
            item.removeAttribute("select-start");
        });
        liElement.classList.add("b3-list-item--focus");
    }
    const selectItemElements = fileElement.querySelectorAll(".b3-list-item--focus");
    if (selectItemElements.length > 1) {
        return initMultiMenu(selectItemElements, app);
    }
    window.scribli.menus.menu.element.setAttribute("data-from", Constants.MENU_FROM_DOC_TREE_MORE_NOTEBOOK);
    const notebookId = liElement.parentElement.getAttribute("data-url");
    const name = getNotebookName(notebookId);
    const boxDocID = liElement.getAttribute("data-node-id");
    if (boxDocID && window.scribli.config.fileTree.parentDocClickExpand &&
        Number(liElement.getAttribute("data-count")) > 0) {
        window.scribli.menus.menu.append(new MenuItem({
            id: "openDocument",
            label: window.scribli.languages.openDocument,
            icon: "iconOpen",
            click: () => {
                openFileById({
                    app,
                    id: boxDocID,
                    action: [Constants.CB_GET_FOCUS, Constants.CB_GET_SCROLL],
                });
            }
        }).element);
    }
    if (!window.scribli.config.readonly) {
        window.scribli.menus.menu.append(new MenuItem({
            id: "changeIcon",
            label: window.scribli.languages.changeIcon,
            icon: "iconEmoji",
            click: () => {
                const iconElement = liElement.querySelector<HTMLElement>(".b3-list-item__icon");
                if (!iconElement) {
                    return;
                }
                const rect = iconElement.getBoundingClientRect();
                openEmojiPanel(notebookId, "notebook", {
                    x: rect.left,
                    y: rect.bottom,
                    h: rect.height,
                    w: rect.width,
                }, undefined, iconElement.querySelector<HTMLElement>("img"));
            }
        }).element);
        window.scribli.menus.menu.append(renameMenu({
            path: "/",
            notebookId,
            name,
            type: "notebook"
        }));
        window.scribli.menus.menu.append(new MenuItem({
            id: "config",
            label: window.scribli.languages.config,
            icon: "iconSettings",
            click: () => {
                fetchPost("/api/notebook/getNotebookConf", {
                    notebook: notebookId
                }, (data) => {
                    onGetnotebookconf(data.data);
                });
            }
        }).element);
        const subMenu = sortMenu("notebook", parseInt(liElement.parentElement.getAttribute("data-sortmode")), (sort) => {
            fetchPost("/api/notebook/setNotebookConf", {
                notebook: notebookId,
                conf: {
                    sortMode: sort
                }
            }, () => {
                liElement.parentElement.setAttribute("data-sortmode", sort.toString());
                const files = (getDockByType("file").data["file"] as Files);
                const toggleElement = liElement.querySelector(".b3-list-item__arrow--open");
                if (toggleElement) {
                    toggleElement.classList.remove("b3-list-item__arrow--open");
                    liElement.nextElementSibling?.remove();
                    files.getLeaf(liElement, notebookId);
                }
            });
            return true;
        });
        window.scribli.menus.menu.append(new MenuItem({
            id: "sort",
            icon: "iconSort",
            label: window.scribli.languages.sort,
            type: "submenu",
            submenu: subMenu,
        }).element);
    }
    if (!window.scribli.config.readonly && !isEncryptedBox(notebookId)) {
        window.scribli.menus.menu.append(new MenuItem({
            id: "riffCard",
            label: window.scribli.languages.riffCard,
            type: "submenu",
            icon: "iconRiffCard",
            submenu: [{
                id: "spaceRepetition",
                iconHTML: "",
                label: window.scribli.languages.spaceRepetition,
                accelerator: window.scribli.config.keymap.editor.general.spaceRepetition.custom,
                click: () => {
                    fetchPost("/api/riff/getNotebookRiffDueCards", {notebook: notebookId}, (response) => {
                        openCardByData(app, response.data, "notebook", notebookId, name);
                    });
                }
            }, {
                id: "manage",
                iconHTML: "",
                label: window.scribli.languages.manage,
                click: () => {
                    viewCards(app, notebookId, name, "Notebook");
                }
            }],
        }).element);
    }
    window.scribli.menus.menu.append(new MenuItem({
        id: "search",
        label: window.scribli.languages.search,
        accelerator: window.scribli.config.keymap.general.search.custom,
        icon: "iconSearch",
        click() {
            openSearch({
                app,
                hotkey: Constants.DIALOG_SEARCH,
                notebookId,
            });
        }
    }).element);
    if (!window.scribli.config.readonly) {
        window.scribli.menus.menu.append(new MenuItem({
            id: "replace",
            label: window.scribli.languages.replace,
            accelerator: window.scribli.config.keymap.general.replace.custom,
            icon: "iconReplace",
            click() {
                openSearch({
                    app,
                    hotkey: Constants.DIALOG_REPLACE,
                    notebookId,
                });
            }
        }).element);
    }
    if (!window.scribli.config.readonly) {
        window.scribli.menus.menu.append(new MenuItem({id: "separator_1", type: "separator"}).element);
        window.scribli.menus.menu.append(new MenuItem({
            id: "close",
            label: window.scribli.languages.close,
            icon: "iconClose",
            click: () => {
                fetchPost("/api/notebook/closeNotebook", {
                    notebook: notebookId
                });
            }
        }).element);
        window.scribli.menus.menu.append(new MenuItem({
            id: "delete",
            icon: "iconTrashcan",
            label: window.scribli.languages.delete,
            accelerator: "⌦",
            click: () => {
                deleteFiles(Array.from(fileElement.querySelectorAll(".b3-list-item--focus")));
            }
        }).element);
    }
    window.scribli.menus.menu.append(new MenuItem({id: "separator_2", type: "separator"}).element);
    /// #if !BROWSER
    window.scribli.menus.menu.append(new MenuItem({
        id: "showInFolder",
        icon: "iconFolder",
        label: window.scribli.languages.showInFolder,
        click: () => {
            useShell("openPath", path.join(window.scribli.config.system.dataDir, notebookId));
        }
    }).element);
    /// #endif
    genImportMenu(notebookId, "/");

    window.scribli.menus.menu.append(new MenuItem({
        id: "export",
        label: window.scribli.languages.export,
        type: "submenu",
        icon: "iconUpload",
        submenu: [{
            id: "exportScribliZip",
            label: "Scribli .sy.zip",
            icon: "iconScribli",
            click: () => {
                confirmEncryptedExport(notebookId, () => {
                    const msgId = showMessage(window.scribli.languages.exporting, -1);
                    fetchPost("/api/export/exportNotebookSY", {
                        id: notebookId,
                    }, response => {
                        saveExportFile(response.data.zip, msgId);
                    });
                });
            }
        }, {
            id: "exportMarkdown",
            label: "Markdown .zip",
            icon: "iconMarkdown",
            click: () => {
                confirmEncryptedExport(notebookId, () => exportMarkdownZip({notebook: notebookId}));
            }
        }]
    }).element);
    if (app.plugins) {
        emitOpenMenu({
            plugins: app.plugins,
            type: "open-menu-doctree",
            detail: {
                elements: selectItemElements,
                type: "notebook"
            },
            separatorPosition: "top",
        });
    }
    return window.scribli.menus.menu;
};

export const initFileMenu = (app: App, notebookId: string, pathString: string, liElement: Element) => {
    window.scribli.menus.menu.remove();
    window.scribli.menus.menu.element.setAttribute("data-name", Constants.MENU_DOC_TREE_MORE);
    const fileElement = hasClosestByTag(liElement, "DIV");
    if (!fileElement) {
        return window.scribli.menus.menu;
    }
    if (!liElement.classList.contains("b3-list-item--focus")) {
        fileElement.querySelectorAll(".b3-list-item--focus").forEach(item => {
            item.classList.remove("b3-list-item--focus");
            item.removeAttribute("select-end");
            item.removeAttribute("select-start");
        });
        liElement.classList.add("b3-list-item--focus");
    }
    const selectItemElements = fileElement.querySelectorAll(".b3-list-item--focus");
    if (selectItemElements.length > 1) {
        return initMultiMenu(selectItemElements, app);
    }
    const id = liElement.getAttribute("data-node-id");
    let name = liElement.getAttribute("data-name");
    name = getDisplayName(name, false, true);
    if (window.scribli.config.fileTree.parentDocClickExpand && Number(liElement.getAttribute("data-count")) > 0) {
        window.scribli.menus.menu.append(new MenuItem({
            id: "openDocument",
            label: window.scribli.languages.openDocument,
            icon: "iconOpen",
            click: () => {
                openFileById({
                    app,
                    id,
                    action: [Constants.CB_GET_FOCUS, Constants.CB_GET_SCROLL],
                });
            }
        }).element);
    }
    if (!window.scribli.config.readonly) {
        const topElement = hasTopClosestByTag(liElement, "UL");
        if (window.scribli.config.fileTree.sort === 6 || (topElement && topElement.dataset.sortmode === "6")) {
            window.scribli.menus.menu.append(new MenuItem({
                id: "newDocAbove",
                icon: "iconBefore",
                label: window.scribli.languages.newDocAbove,
                click: () => {
                    const paths: string[] = [];
                    Array.from(liElement.parentElement.children).forEach((item) => {
                        if (item.tagName === "LI") {
                            if (item === liElement) {
                                paths.push(undefined);
                            }
                            paths.push(item.getAttribute("data-path"));
                        }
                    });
                    newFileInTree(app, notebookId, pathPosix().dirname(pathString), paths);
                }
            }).element);
            window.scribli.menus.menu.append(new MenuItem({
                id: "newDocBelow",
                icon: "iconAfter",
                label: window.scribli.languages.newDocBelow,
                click: () => {
                    const paths: string[] = [];
                    Array.from(liElement.parentElement.children).forEach((item) => {
                        if (item.tagName === "LI") {
                            paths.push(item.getAttribute("data-path"));
                            if (item === liElement) {
                                paths.push(undefined);
                            }
                        }
                    });
                    newFileInTree(app, notebookId, pathPosix().dirname(pathString), paths);
                }
            }).element);
            window.scribli.menus.menu.append(new MenuItem({id: "separator_1", type: "separator"}).element);
        }
        window.scribli.menus.menu.append(new MenuItem({
            id: "copy",
            label: window.scribli.languages.copy,
            type: "submenu",
            icon: "iconCopy",
            submenu: (copySubMenu([id]) as IMenu[]).concat([{
                id: "duplicate",
                iconHTML: "",
                label: window.scribli.languages.duplicate,
                accelerator: window.scribli.config.keymap.editor.general.duplicate.custom,
                click() {
                    fetchPost("/api/filetree/duplicateDoc", {
                        id
                    });
                }
            }])
        }).element);
        window.scribli.menus.menu.append(movePathToMenu(getTopPaths(
            Array.from(fileElement.querySelectorAll(".b3-list-item--focus"))
        )));
        window.scribli.menus.menu.append(new MenuItem({
            id: "addToDatabase",
            label: window.scribli.languages.addToDatabase,
            accelerator: window.scribli.config.keymap.general.addToDatabase.custom,
            icon: "iconDatabase",
            click: () => {
                addFilesToDatabase([liElement]);
            }
        }).element);
        window.scribli.menus.menu.append(new MenuItem({
            id: "delete",
            icon: "iconTrashcan",
            label: window.scribli.languages.delete,
            accelerator: "⌦",
            click: () => {
                deleteFiles(Array.from(fileElement.querySelectorAll(".b3-list-item--focus")));
            }
        }).element);
        window.scribli.menus.menu.append(new MenuItem({id: "separator_2", type: "separator"}).element);
        window.scribli.menus.menu.append(renameMenu({
            path: pathString,
            notebookId,
            name,
            type: "file",
            docId: id,
        }));
        window.scribli.menus.menu.append(new MenuItem({
            id: "attr",
            label: window.scribli.languages.attr,
            icon: "iconAttr",
            click() {
                const docInfoParam: IObject = {
                    id
                };
                if (isEncryptedBox(notebookId)) {
                    docInfoParam.notebook = notebookId;
                }
                fetchPost("/api/block/getDocInfo", docInfoParam, (response) => {
                    openFileAttr(response.data.ial);
                });
            }
        }).element);
        if (!window.scribli.config.readonly && !isEncryptedBox(notebookId)) {
            const riffCardMenu = [{
                id: "spaceRepetition",
                iconHTML: "",
                label: window.scribli.languages.spaceRepetition,
                accelerator: window.scribli.config.keymap.editor.general.spaceRepetition.custom,
                click: () => {
                    fetchPost("/api/riff/getTreeRiffDueCards", {rootID: id}, (response) => {
                        openCardByData(app, response.data, "doc", id, name);
                    });
                }
            }, {
                id: "manage",
                iconHTML: "",
                label: window.scribli.languages.manage,
                click: () => {
                    fetchPost("/api/filetree/getHPathByID", {
                        id
                    }, (response) => {
                        viewCards(app, id, pathPosix().join(getNotebookName(notebookId), response.data), "Tree");
                    });
                }
            }, {
                id: "quickMakeCard",
                iconHTML: "",
                accelerator: window.scribli.config.keymap.editor.general.quickMakeCard.custom,
                label: window.scribli.languages.quickMakeCard,
                click: () => {
                    transaction(undefined, [{
                        action: "addFlashcards",
                        deckID: Constants.QUICK_DECK_ID,
                        blockIDs: [id]
                    }]);
                }
            }, {
                id: "removeCard",
                iconHTML: "",
                label: window.scribli.languages.removeCard,
                click: () => {
                    transaction(undefined, [{
                        action: "removeFlashcards",
                        deckID: Constants.QUICK_DECK_ID,
                        blockIDs: [id]
                    }]);
                }
            }];
            if (window.scribli.config.flashcard.deck) {
                riffCardMenu.push({
                    id: "addToDeck",
                    iconHTML: "",
                    label: window.scribli.languages.addToDeck,
                    click: () => {
                        makeCard(app, [id]);
                    }
                });
            }
            window.scribli.menus.menu.append(new MenuItem({
                id: "riffCard",
                label: window.scribli.languages.riffCard,
                type: "submenu",
                icon: "iconRiffCard",
                submenu: riffCardMenu,
            }).element);
        }
        window.scribli.menus.menu.append(new MenuItem({
            id: "search",
            label: window.scribli.languages.search,
            icon: "iconSearch",
            accelerator: window.scribli.config.keymap.general.search.custom,
            async click() {
                const searchPath = getDisplayName(pathString, false, true);
                openSearch({
                    app,
                    hotkey: Constants.DIALOG_SEARCH,
                    notebookId,
                    searchPath
                });
            }
        }).element);
        window.scribli.menus.menu.append(new MenuItem({
            id: "replace",
            label: window.scribli.languages.replace,
            accelerator: window.scribli.config.keymap.general.replace.custom,
            icon: "iconReplace",
            async click() {
                const searchPath = getDisplayName(pathString, false, true);
                openSearch({
                    app,
                    hotkey: Constants.DIALOG_REPLACE,
                    notebookId,
                    searchPath
                });
            }
        }).element);
        window.scribli.menus.menu.append(new MenuItem({id: "separator_3", type: "separator"}).element);
    }
    openEditorTab(app, [id], notebookId, pathString);
    if (!window.scribli.config.readonly) {
        window.scribli.menus.menu.append(new MenuItem({
            id: "fileHistory",
            label: window.scribli.languages.fileHistory,
            icon: "iconHistory",
            click() {
                openDocHistory({app, id, notebookId, pathString: name});
            }
        }).element);
    }
    genImportMenu(notebookId, pathString);
    window.scribli.menus.menu.append(exportMd(id));
    if (app.plugins) {
        emitOpenMenu({
            plugins: app.plugins,
            type: "open-menu-doctree",
            detail: {
                elements: selectItemElements,
                type: "doc"
            },
            separatorPosition: "top",
        });
    }
    window.scribli.menus.menu.element.setAttribute("data-from", Constants.MENU_FROM_DOC_TREE_MORE_DOC);
    return window.scribli.menus.menu;
};

export const genImportMenu = (notebookId: string, pathString: string) => {
    if (window.scribli.config.readonly) {
        return;
    }
    const reloadDocTree = () => {
        const files = (getDockByType("file").data["file"] as Files);
        const liElement = files.element.querySelector(`[data-path="${pathString}"]`);
        liElement.querySelector(".b3-list-item__toggle").classList.remove("fn__hidden");
        files.getLeaf(liElement, notebookId, true);
        window.scribli.menus.menu.remove();
    };
    /// #if !BROWSER
    const importstdmd = (label: string, isDoc?: boolean) => {
        return {
            id: isDoc ? "importMarkdownDoc" : "importMarkdownFolder",
            icon: isDoc ? "iconMarkdown" : "iconFolder",
            label,
            click: async () => {
                let filters: FileFilter[] = [];
                if (isDoc) {
                    filters = [{name: "Markdown", extensions: ["md", "markdown"]}];
                }
                const localPath = await ipcRenderer.invoke(Constants.SCRIBLI_GET, {
                    cmd: "showOpenDialog",
                    defaultPath: window.scribli.config.system.homeDir,
                    filters,
                    properties: [isDoc ? "openFile" : "openDirectory"],
                });
                if (localPath.filePaths.length === 0) {
                    return;
                }
                fetchPost("/api/import/importStdMd", {
                    notebook: notebookId,
                    localPath: localPath.filePaths[0],
                    toPath: pathString,
                }, () => {
                    reloadDocTree();
                });
            }
        };
    };
    const importEbook = () => {
        return {
            id: "importEbook",
            icon: "iconFile",
            label: "Ebook (.epub, .mobi, .azw, .azw3)",
            click: async () => {
                const localPath = await ipcRenderer.invoke(Constants.SCRIBLI_GET, {
                    cmd: "showOpenDialog",
                    defaultPath: window.scribli.config.system.homeDir,
                    filters: [{name: "Ebook", extensions: ["epub", "mobi", "azw", "azw3"]}],
                    properties: ["openFile"],
                });
                if (localPath.filePaths.length === 0) {
                    return;
                }
                fetchPost("/api/import/importEbook", {
                    notebook: notebookId,
                    localPath: localPath.filePaths[0],
                    toPath: pathString,
                }, () => {
                    reloadDocTree();
                });
            }
        };
    };
    const importDocument = () => {
        return {
            id: "importDocument",
            icon: "iconDocx",
            label: "Document (.docx, .html, .rtf, .odt, .txt)",
            click: async () => {
                const localPath = await ipcRenderer.invoke(Constants.SCRIBLI_GET, {
                    cmd: "showOpenDialog",
                    defaultPath: window.scribli.config.system.homeDir,
                    filters: [{name: "Document", extensions: ["docx", "html", "htm", "rtf", "odt", "txt"]}],
                    properties: ["openFile"],
                });
                if (localPath.filePaths.length === 0) {
                    return;
                }
                fetchPost("/api/import/importDocument", {
                    notebook: notebookId,
                    localPath: localPath.filePaths[0],
                    toPath: pathString,
                }, () => {
                    reloadDocTree();
                });
            }
        };
    };
    /// #endif
    window.scribli.menus.menu.append(new MenuItem({
        id: "import",
        icon: "iconDownload",
        label: window.scribli.languages.import,
        submenu: [
            {
                id: "importScribliZip",
                icon: "iconScribli",
                label: 'Scribli .sy.zip<input class="b3-form__upload" type="file" accept="application/zip">',
                bind: (element) => {
                    element.querySelector(".b3-form__upload").addEventListener("change", (event: InputEvent & {
                        target: HTMLInputElement
                    }) => {
                        const formData = new FormData();
                        formData.append("file", event.target.files[0]);
                        formData.append("notebook", notebookId);
                        formData.append("toPath", pathString);
                        fetchPost("/api/import/importSY", formData, () => {
                            reloadDocTree();
                        });
                    });
                }
            },
            {
                id: "importMarkdownZip",
                icon: "iconMarkdown",
                label: 'Markdown .zip<input class="b3-form__upload" type="file" accept="application/zip">',
                bind: (element) => {
                    element.querySelector(".b3-form__upload").addEventListener("change", (event: InputEvent & {
                        target: HTMLInputElement
                    }) => {
                        const formData = new FormData();
                        formData.append("file", event.target.files[0]);
                        formData.append("notebook", notebookId);
                        formData.append("toPath", pathString);
                        fetchPost("/api/import/importZipMd", formData, () => {
                            reloadDocTree();
                        });
                    });
                }
            },
            /// #if !BROWSER
            importstdmd("Markdown " + window.scribli.languages.doc, true),
            importstdmd("Markdown " + window.scribli.languages.folder),
            importDocument(),
            importEbook()
            /// #endif
        ],
    }).element);
};

export const sortMenu = (type: "notebooks" | "notebook", sortMode: number, clickEvent: (sort: number) => void) => {
    const sortMenu: IMenu[] = [{
        id: "fileNameASC",
        checked: sortMode === 0,
        iconHTML: "",
        label: window.scribli.languages.fileNameASC,
        click: () => {
            clickEvent(0);
        }
    }, {
        id: "fileNameDESC",
        checked: sortMode === 1,
        iconHTML: "",
        label: window.scribli.languages.fileNameDESC,
        click: () => {
            clickEvent(1);
        }
    }, {
        id: "fileNameNatASC",
        checked: sortMode === 4,
        iconHTML: "",
        label: window.scribli.languages.fileNameNatASC,
        click: () => {
            clickEvent(4);
        }
    }, {
        id: "fileNameNatDESC",
        checked: sortMode === 5,
        iconHTML: "",
        label: window.scribli.languages.fileNameNatDESC,
        click: () => {
            clickEvent(5);
        }
    }, {id: "separator_1", type: "separator"}, {
        id: "createdASC",
        checked: sortMode === 9,
        iconHTML: "",
        label: window.scribli.languages.createdASC,
        click: () => {
            clickEvent(9);
        }
    }, {
        id: "createdDESC",
        checked: sortMode === 10,
        iconHTML: "",
        label: window.scribli.languages.createdDESC,
        click: () => {
            clickEvent(10);
        }
    }, {
        id: "modifiedASC",
        checked: sortMode === 2,
        iconHTML: "",
        label: window.scribli.languages.modifiedASC,
        click: () => {
            clickEvent(2);
        }
    }, {
        id: "modifiedDESC",
        checked: sortMode === 3,
        iconHTML: "",
        label: window.scribli.languages.modifiedDESC,
        click: () => {
            clickEvent(3);
        }
    }, {id: "separator_2", type: "separator"}, {
        id: "refCountASC",
        checked: sortMode === 7,
        iconHTML: "",
        label: window.scribli.languages.refCountASC,
        click: () => {
            clickEvent(7);
        }
    }, {
        id: "refCountDESC",
        checked: sortMode === 8,
        iconHTML: "",
        label: window.scribli.languages.refCountDESC,
        click: () => {
            clickEvent(8);
        }
    }, {id: "separator_3", type: "separator"}, {
        id: "docSizeASC",
        checked: sortMode === 11,
        iconHTML: "",
        label: window.scribli.languages.docSizeASC,
        click: () => {
            clickEvent(11);
        }
    }, {
        id: "docSizeDESC",
        checked: sortMode === 12,
        iconHTML: "",
        label: window.scribli.languages.docSizeDESC,
        click: () => {
            clickEvent(12);
        }
    }, {id: "separator_4", type: "separator"}, {
        id: "subDocCountASC",
        checked: sortMode === 13,
        iconHTML: "",
        label: window.scribli.languages.subDocCountASC,
        click: () => {
            clickEvent(13);
        }
    }, {
        id: "subDocCountDESC",
        checked: sortMode === 14,
        iconHTML: "",
        label: window.scribli.languages.subDocCountDESC,
        click: () => {
            clickEvent(14);
        }
    }, {id: "separator_5", type: "separator"}, {
        id: "customSort",
        checked: sortMode === 6,
        iconHTML: "",
        label: window.scribli.languages.customSort,
        click: () => {
            clickEvent(6);
        }
    }];
    if (type === "notebook") {
        sortMenu.push({
            id: "sortByFiletree",
            checked: sortMode === 15,
            iconHTML: "",
            label: window.scribli.languages.sortByFiletree,
            click: () => {
                clickEvent(15);
            }
        });
    }
    return sortMenu;
};

import {fetchPost, fetchSyncPost} from "../../util/fetch";
import {MenuItem} from "../../menus/Menu";
import {copySubMenu, exportMd, movePathToMenu, openFileAttr, openFileWechatNotify,} from "../../menus/commonMenuItem";
import {deleteFile} from "../../editor/deleteFile";
import {encodeBase64, updateHotkeyTip} from "../util/compatibility";
/// #if !MOBILE
import {openBacklink, openGraph, openOutline} from "../../layout/dock/util";
import * as path from "path";
/// #else
import {openMobileFileById} from "../../mobile/editor";
/// #endif
import {Constants} from "../../constants";
import {openCardByData} from "../../card/openCard";
import {viewCards} from "../../card/viewCards";
import {getDisplayName, getNotebookName, isEncryptedBox, pathPosix, useShell} from "../../util/pathName";
import {makeCard, quickMakeCard} from "../../card/makeCard";
import {emitOpenMenu} from "../../plugin/EventBus";
import * as dayjs from "dayjs";
import {hideTooltip} from "../../dialog/tooltip";
import {popSearch} from "../../mobile/menu/search";
import {openSearch} from "../../search/spread";
import {openDocHistory} from "../../history/doc";
import {openNewWindowById} from "../../window/openNewWindow";
import {transferBlockRef} from "../../menus/block";
import {addEditorToDatabase} from "../render/av/addToDatabase";
import {openFileById} from "../../editor/util";
import {hasTopClosestByClassName} from "../util/hasClosest";
import {showMessage} from "../../dialog/message";
import {removeZWJ} from "../util/normalizeText";

export const openTitleMenu = (protyle: IProtyle, position: IPosition, from: string) => {
    hideTooltip();
    if (!window.scribli.menus.menu.element.classList.contains("fn__none") &&
        window.scribli.menus.menu.element.getAttribute("data-name") === Constants.MENU_TITLE) {
        window.scribli.menus.menu.remove();
        return;
    }
    const docInfoParam: IObject = {
        id: protyle.block.rootID
    };
    if (isEncryptedBox(protyle.notebookId)) {
        docInfoParam.notebook = protyle.notebookId;
    }
    fetchPost("/api/block/getDocInfo", docInfoParam, (response) => {
        window.scribli.menus.menu.remove();
        window.scribli.menus.menu.element.setAttribute("data-name", Constants.MENU_TITLE);
        const isBoxDoc = protyle.notebookId === protyle.block.rootID;
        const popoverElement = hasTopClosestByClassName(protyle.element, "block__popover", true);
        window.scribli.menus.menu.element.setAttribute("data-from", popoverElement ? popoverElement.dataset.level + "popover-" + from : "app-" + from);
        const submenu = copySubMenu([protyle.block.rootID], true, undefined, protyle.block.showAll ? protyle.block.id : protyle.block.rootID);
        submenu.push({
            iconHTML: "",
            label: window.scribli.languages.copyDoc,
            accelerator: undefined,
            click: async () => {
                const [responseHTML, responseText] = await Promise.all([
                    fetchSyncPost("/api/block/getBlockDOM", {
                        id: protyle.block.rootID,
                        notebook: protyle.notebookId,
                    }),
                    fetchSyncPost("/api/export/exportMdContent", {
                        id: protyle.block.rootID,
                        refMode: 3,
                        embedMode: 1,
                        yfm: false,
                        fillCSSVar: false,
                        adjustHeadingLevel: false
                    })
                ]);

                const textHTML = `<!--data-scribli='${encodeBase64(responseHTML.data.dom)}'-->${removeZWJ(responseHTML.data.dom)}`;
                await navigator.clipboard.write([
                    new ClipboardItem({
                        "text/plain": new Blob([responseText.data.content], {type: "text/plain"}),
                        "text/html": new Blob([textHTML], {type: "text/html"}),
                    })
                ]);

                showMessage(window.scribli.languages.copied);
            }
        });
        window.scribli.menus.menu.append(new MenuItem({
            id: "copy",
            label: window.scribli.languages.copy,
            icon: "iconCopy",
            type: "submenu",
            submenu,
        }).element);
        if (!protyle.disabled) {
            if (!isBoxDoc) {
                window.scribli.menus.menu.append(movePathToMenu([protyle.path]));
            }
            const range = getSelection().rangeCount > 0 ? getSelection().getRangeAt(0) : undefined;
            window.scribli.menus.menu.append(new MenuItem({
                id: "addToDatabase",
                label: window.scribli.languages.addToDatabase,
                accelerator: window.scribli.config.keymap.general.addToDatabase.custom,
                icon: "iconDatabase",
                click: () => {
                    addEditorToDatabase(protyle, range, "title");
                }
            }).element);
            if (!isBoxDoc) {
                window.scribli.menus.menu.append(new MenuItem({
                    id: "delete",
                    icon: "iconTrashcan",
                    label: window.scribli.languages.delete,
                    click: () => {
                        deleteFile(protyle.notebookId, protyle.path);
                    }
                }).element);
            }
        }
        /// #if !MOBILE
        window.scribli.menus.menu.append(new MenuItem({id: "separator_1", type: "separator"}).element);
        window.scribli.menus.menu.append(new MenuItem({
            id: "outline",
            icon: "iconOutline",
            label: window.scribli.languages.outline,
            accelerator: window.scribli.config.keymap.editor.general.outline.custom,
            click: () => {
                openOutline({
                    app: protyle.app,
                    rootId: protyle.block.rootID,
                    title: protyle.options.render.title ? (protyle.title.editElement.textContent || window.scribli.languages.untitled) : "",
                    isPreview: !protyle.preview.element.classList.contains("fn__none")
                });
            }
        }).element);
        window.scribli.menus.menu.append(new MenuItem({
            id: "backlinks",
            icon: "iconLink",
            label: window.scribli.languages.backlinks,
            accelerator: window.scribli.config.keymap.editor.general.backlinks.custom,
            click: () => {
                openBacklink({
                    app: protyle.app,
                    blockId: protyle.block.id,
                    rootId: protyle.block.rootID,
                    useBlockId: protyle.block.showAll,
                    title: protyle.title ? (protyle.title.editElement.textContent || window.scribli.languages.untitled) : null
                });
            }
        }).element);
        window.scribli.menus.menu.append(new MenuItem({
            id: "graphView",
            icon: "iconGraph",
            label: window.scribli.languages.graphView,
            accelerator: window.scribli.config.keymap.editor.general.graphView.custom,
            click: () => {
                openGraph({
                    app: protyle.app,
                    blockId: protyle.block.id,
                    rootId: protyle.block.rootID,
                    useBlockId: protyle.block.showAll,
                    title: protyle.title ? (protyle.title.editElement.textContent || window.scribli.languages.untitled) : null
                });
            }
        }).element);
        /// #endif
        window.scribli.menus.menu.append(new MenuItem({id: "separator_2", type: "separator"}).element);
        window.scribli.menus.menu.append(new MenuItem({
            id: "attr",
            label: window.scribli.languages.attr,
            icon: "iconAttr",
            accelerator: window.scribli.config.keymap.editor.general.attr.custom + "/" + updateHotkeyTip("⇧" + window.scribli.languages.click),
            click() {
                openFileAttr(response.data.ial, "bookmark", protyle);
            }
        }).element);
        if (!window.scribli.config.readonly) {
            if (window.scribli.config.cloudRegion === 0) {
                window.scribli.menus.menu.append(new MenuItem({
                    id: "wechatReminder",
                    label: window.scribli.languages.wechatReminder,
                    icon: "iconMp",
                    click() {
                        openFileWechatNotify(protyle);
                    }
                }).element);
            }
            const isCardMade = !!response.data.ial[Constants.CUSTOM_RIFF_DECKS];
            if (!isEncryptedBox(protyle.notebookId)) {
            const riffCardMenu: IMenu[] = [{
                id: "spaceRepetition",
                iconHTML: "",
                label: window.scribli.languages.spaceRepetition,
                accelerator: window.scribli.config.keymap.editor.general.spaceRepetition.custom,
                click: () => {
                    fetchPost("/api/riff/getTreeRiffDueCards", {rootID: protyle.block.rootID}, (response) => {
                        openCardByData(protyle.app, response.data, "doc", protyle.block.rootID, response.data.name);
                    });
                }
            }, {
                id: "manage",
                iconHTML: "",
                label: window.scribli.languages.manage,
                click: () => {
                    fetchPost("/api/filetree/getHPathByID", {
                        id: protyle.block.rootID
                    }, (response) => {
                        viewCards(protyle.app, protyle.block.rootID, pathPosix().join(getNotebookName(protyle.notebookId), (response.data)), "Tree");
                    });
                }
            }, {
                id: isCardMade ? "removeCard" : "quickMakeCard",
                iconHTML: "",
                label: isCardMade ? window.scribli.languages.removeCard : window.scribli.languages.quickMakeCard,
                accelerator: window.scribli.config.keymap.editor.general.quickMakeCard.custom,
                click: () => {
                    let titleElement = protyle.title?.element;
                    if (!titleElement) {
                        titleElement = document.createElement("div");
                        titleElement.setAttribute("data-node-id", protyle.block.rootID);
                        titleElement.setAttribute(Constants.CUSTOM_RIFF_DECKS, response.data.ial[Constants.CUSTOM_RIFF_DECKS]);
                    }
                    quickMakeCard(protyle, [titleElement]);
                }
            }];
            if (window.scribli.config.flashcard.deck) {
                riffCardMenu.push({
                    id: "addToDeck",
                    iconHTML: "",
                    label: window.scribli.languages.addToDeck,
                    click: () => {
                        makeCard(protyle.app, [protyle.block.rootID]);
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
        }
        window.scribli.menus.menu.append(new MenuItem({
            id: "search",
            label: window.scribli.languages.search,
            icon: "iconSearch",
            accelerator: window.scribli.config.keymap.general.search.custom,
            async click() {
                const searchPath = isBoxDoc ? "" : getDisplayName(protyle.path, false, true);
                /// #if MOBILE
                const pathResponse = isBoxDoc ? undefined : await fetchSyncPost("/api/filetree/getHPathByPath", {
                        notebook: protyle.notebookId,
                        path: searchPath + ".sy"
                    });
                popSearch(protyle.app, {
                    hasReplace: false,
                    hPath: isBoxDoc ? getNotebookName(protyle.notebookId) : pathPosix().join(getNotebookName(protyle.notebookId), pathResponse.data),
                    idPath: [isBoxDoc ? protyle.notebookId : pathPosix().join(protyle.notebookId, searchPath)],
                    page: 1,
                });
                /// #else
                openSearch({
                    app: protyle.app,
                    hotkey: Constants.DIALOG_SEARCH,
                    notebookId: protyle.notebookId,
                    searchPath
                });
                /// #endif
            }
        }).element);
        if (!protyle.disabled) {
            transferBlockRef(protyle.block.rootID);
        }
        window.scribli.menus.menu.append(new MenuItem({id: "separator_3", type: "separator"}).element);
        if (!protyle.model) {
            window.scribli.menus.menu.append(new MenuItem({
                id: "openBy",
                label: window.scribli.languages.openBy,
                icon: "iconOpen",
                click() {
                    /// #if !MOBILE
                    openFileById({
                        app: protyle.app,
                        id: protyle.block.id,
                        action: protyle.block.rootID !== protyle.block.id ? [Constants.CB_GET_ALL, Constants.CB_GET_FOCUS] : [Constants.CB_GET_CONTEXT],
                    });
                    /// #else
                    openMobileFileById(protyle.app, protyle.block.id, protyle.block.rootID !== protyle.block.id ? [Constants.CB_GET_ALL] : [Constants.CB_GET_CONTEXT]);
                    /// #endif
                }
            }).element);
        }
        /// #if !BROWSER
        window.scribli.menus.menu.append(new MenuItem({
            id: "openByNewWindow",
            label: window.scribli.languages.openByNewWindow,
            icon: "iconOpenWindow",
            click() {
                openNewWindowById(protyle.block.rootID);
            }
        }).element);
        window.scribli.menus.menu.append(new MenuItem({
            id: "showInFolder",
            icon: "iconFolder",
            label: window.scribli.languages.showInFolder,
            click: () => {
                useShell("showItemInFolder", path.join(window.scribli.config.system.dataDir, protyle.notebookId, protyle.path));
            }
        }).element);
        /// #endif
        if (!protyle.disabled) {
            window.scribli.menus.menu.append(new MenuItem({
                id: "fileHistory",
                label: window.scribli.languages.fileHistory,
                icon: "iconHistory",
                click() {
                    openDocHistory({
                        app: protyle.app,
                        id: protyle.block.rootID,
                        notebookId: protyle.notebookId,
                        pathString: response.data.name
                    });
                }
            }).element);
        }
        window.scribli.menus.menu.append(exportMd(protyle.block.showAll ? protyle.block.id : protyle.block.rootID));

        window.scribli.menus.menu.append(new MenuItem({id: "separator_4", type: "separator"}).element);
        if (protyle?.app?.plugins) {
            emitOpenMenu({
                plugins: protyle.app.plugins,
                type: "click-editortitleicon",
                detail: {
                    protyle,
                    data: response.data,
                },
                separatorPosition: "bottom",
            });
        }
        window.scribli.menus.menu.append(new MenuItem({
            id: "updateAndCreatedAt",
            iconHTML: "",
            type: "readonly",
            // 不能换行，否则移动端间距过大
            label: `${window.scribli.languages.modifiedAt} ${dayjs(response.data.ial.updated).format("YYYY-MM-DD HH:mm:ss")}<br>${window.scribli.languages.createdAt} ${dayjs(response.data.ial.id.substr(0, 14)).format("YYYY-MM-DD HH:mm:ss")}`
        }).element);
        /// #if MOBILE
        window.scribli.menus.menu.fullscreen();
        /// #else
        window.scribli.menus.menu.popup(position);
        /// #endif
    });
};

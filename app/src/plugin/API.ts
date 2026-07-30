import {confirmDialog} from "../dialog/confirmDialog";
import {Plugin} from "./index";
import {hideMessage, showMessage} from "../dialog/message";
import {Dialog} from "../dialog";
import {fetchGet, fetchPost, fetchSyncPost} from "../util/fetch";
import {getBackend, getFrontend} from "../util/functions";
import {openFile, openFileById} from "../editor/util";
import {openNewWindow, openNewWindowById} from "../window/openNewWindow";
import {Tab} from "../layout/Tab";
import {saveExportFile, updateHotkeyTip} from "../protyle/util/compatibility";
import * as platformUtils from "./platformUtils";
import {App} from "../index";
import {Constants} from "../constants";
import {Setting} from "./Setting";
import {Menu} from "./Menu";
import {Protyle} from "../protyle";
import {exitScribli, lockScreen} from "../dialog/processSystem";
import {Model} from "../layout/Model";
import {getActiveTab, getDockByType} from "../layout/tabUtil";
import {getAllModels, getAllTabs} from "../layout/getAll";
import {exportLayout} from "../layout/util";
import {getAllEditor} from "../layout/getAll";
import {openSetting} from "../config";
import {openAttr, openFileAttr} from "../menus/commonMenuItem";
import {globalCommand} from "../boot/globalEvent/command/global";
import {hasClosestByClassName} from "../protyle/util/hasClosest";
import type {Files} from "../layout/dock/Files";
import {ProtyleMethod} from "./ProtyleMethod";
import {openEmojiPanel} from "../emoji";

const openWindow = (options: {
    position?: IPosition,
    height?: number,
    width?: number,
    tab?: Tab,
    alwaysOnTop?: boolean,
    doc?: {
        id: string,
    },
}) => {
    if (options.doc && options.doc.id) {
        openNewWindowById(options.doc.id, {
            alwaysOnTop: options.alwaysOnTop,
            position: options.position,
            width: options.width,
            height: options.height
        });
        return;
    }
    if (options.tab) {
        openNewWindow(options.tab, {
            alwaysOnTop: options.alwaysOnTop,
            position: options.position,
            width: options.width,
            height: options.height
        });
        return;
    }
};

const openTab = (options: {
    app: App,
    doc?: {
        id: string,
        action?: TProtyleAction []
        zoomIn?: boolean
        mode?: TEditorMode
    },
    pdf?: {
        path: string,
        page?: number,
        id?: string,    // File Annotation id
    },
    asset?: {
        path: string,
    },
    search?: Config.IUILayoutTabSearchConfig
    card?: {
        type: TCardType,
        id?: string,
        title?: string
    },
    custom?: {
        title: string,
        icon: string,
        data?: any
        id: string
    }
    position?: "right" | "bottom",
    keepCursor?: boolean
    removeCurrentTab?: boolean
    afterOpen?: (model?: Model) => void
}) => {
    if (options.doc) {
        if (options.doc.zoomIn) {
            if (options.doc.action && !options.doc.action.includes(Constants.CB_GET_ALL)) {
                options.doc.action.push(Constants.CB_GET_ALL);
            } else {
                options.doc.action = [Constants.CB_GET_ALL];
            }
        }
        if (!options.doc.action) {
            options.doc.action = [];
        }
        return openFileById({
            app: options.app,
            keepCursor: options.keepCursor,
            removeCurrentTab: options.removeCurrentTab,
            position: options.position,
            afterOpen: options.afterOpen,
            id: options.doc.id,
            action: options.doc.action,
            zoomIn: options.doc.zoomIn,
            scrollPosition: "start",
            mode: options.doc.mode,
        });
    }
    if (options.asset) {
        return openFile({
            app: options.app,
            keepCursor: options.keepCursor,
            removeCurrentTab: options.removeCurrentTab,
            position: options.position,
            afterOpen: options.afterOpen,
            assetPath: options.asset.path,
        });
    }
    if (options.pdf) {
        return openFile({
            app: options.app,
            keepCursor: options.keepCursor,
            removeCurrentTab: options.removeCurrentTab,
            position: options.position,
            afterOpen: options.afterOpen,
            assetPath: options.pdf.path,
            page: options.pdf.id || options.pdf.page,
        });
    }
    if (options.search) {
        if (!options.search.idPath) {
            options.search.idPath = [];
        }
        if (!options.search.hPath) {
            options.search.hPath = "";
        }
        return openFile({
            app: options.app,
            keepCursor: options.keepCursor,
            removeCurrentTab: options.removeCurrentTab,
            position: options.position,
            afterOpen: options.afterOpen,
            searchData: options.search,
        });
    }
    if (options.card) {
        return openFile({
            app: options.app,
            keepCursor: options.keepCursor,
            removeCurrentTab: options.removeCurrentTab,
            position: options.position,
            afterOpen: options.afterOpen,
            custom: {
                icon: "iconRiffCard",
                title: window.scribli.languages.spaceRepetition,
                data: {
                    cardType: options.card.type,
                    id: options.card.id || "",
                    title: options.card.title,
                },
                id: "scribli-card"
            },
        });
    }
    if (options.custom) {
        return openFile(options);
    }

};

const getModelByDockType = (type: TDock | string) => {
    return getDockByType(type).data[type];
};

const openAttributePanel = (options: {
    data?: Record<string, string>
    nodeElement?: HTMLElement,
    focusName: "bookmark" | "name" | "alias" | "memo" | "av" | "custom",
    protyle?: IProtyle,
}) => {
    if (options.data) {
        openFileAttr(options.data, options.focusName, options.protyle);
    } else {
        openAttr(options.nodeElement, options.focusName, options.protyle);
    }
};

const saveLayout = (cb: () => void) => {
    exportLayout({cb, errorExit: false});
};

const getActiveEditor = (wndActive = true) => {
    let editor;
    const range = getSelection().rangeCount > 0 ? getSelection().getRangeAt(0) : null;
    const allEditor = getAllEditor();
    if (range) {
        editor = allEditor.find(item => {
            if (item.protyle.element.contains(range.startContainer)) {
                return true;
            }
        });
    }
    if (!editor) {
        editor = allEditor.find(item => {
            if (!item.protyle.element.classList.contains("fn__none") &&
                hasClosestByClassName(item.protyle.element, "layout__wnd--active", true)) {
                return true;
            }
        });
    }
    if (!editor && !wndActive) {
        let activeTime = 0;
        allEditor.forEach(item => {
            let headerElement = item.protyle.model?.parent.headElement;
            if (!headerElement && item.protyle.element.getBoundingClientRect().height > 0) {
                const tabBodyElement = item.protyle.element.closest(".fn__flex-1[data-id]");
                if (tabBodyElement) {
                    headerElement = document.querySelector(`.layout-tab-bar .item[data-id="${tabBodyElement.getAttribute("data-id")}"]`);
                }
            }
            if (headerElement) {
                if (headerElement.classList.contains("item--focus") && parseInt(headerElement.dataset.activetime) > activeTime) {
                    activeTime = parseInt(headerElement.dataset.activetime);
                    editor = item;
                }
            } else if (item.protyle.element.getBoundingClientRect().height > 0) {
                editor = item;
            }
        });
    }
    return editor;
};

export const expandDocTree = async (options: {
    id: string,
    isSetCurrent?: boolean
}) => {
    let isNotebook = false;
    window.scribli.notebooks.find(item => {
        if (options.id === item.id) {
            isNotebook = true;
            return true;
        }
    });
    let liElement: HTMLElement;
    let notebookId = options.id;
    const file = getModelByDockType("file") as Files;
    if (typeof options.isSetCurrent === "undefined") {
        options.isSetCurrent = true;
    }
    if (isNotebook) {
        liElement = file.element.querySelector(`.b3-list[data-url="${options.id}"]`)?.firstElementChild as HTMLElement;
    } else {
        const response = await fetchSyncPost("/api/block/getBlockInfo", {id: options.id});
        if (response.code === -1) {
            return;
        }
        notebookId = response.data.box;
        liElement = await file.selectItem(response.data.box, response.data.path, undefined, undefined, options.isSetCurrent);
    }
    if (!liElement) {
        return;
    }
    if (options.isSetCurrent || typeof options.isSetCurrent === "undefined") {
        file.setCurrent(liElement);
    }
    const toggleElement = liElement.querySelector(".b3-list-item__arrow");
    if (toggleElement.classList.contains("b3-list-item__arrow--open")) {
        return;
    }
    file.getLeaf(liElement, notebookId);
};

const openEmoji = (options: {
    position: IPosition,
    selectedCB?: (emoji: string) => void,
    dynamicIconURL?: string
    hideDynamicIcon?: boolean
    hideCustomIcon?: boolean
}) => {
    let dynamicImgElement: HTMLImageElement;
    if (options.dynamicIconURL) {
        dynamicImgElement = document.createElement("img");
        dynamicImgElement.src = options.dynamicIconURL;
    }
    openEmojiPanel("", "av", options.position, options.selectedCB, dynamicImgElement, {
        dynamic: options.hideDynamicIcon,
        custom: options.hideCustomIcon
    });
};

export const API = {
    adaptHotkey: updateHotkeyTip,
    confirm: confirmDialog,
    Constants,
    showMessage,
    hideMessage,
    fetchPost,
    fetchSyncPost,
    fetchGet,
    getFrontend,
    getBackend,
    getModelByDockType,
    openTab,
    openWindow,
    lockScreen,
    exitScribli,
    Protyle,
    ProtyleMethod,
    Plugin,
    Dialog,
    Menu,
    Setting,
    getAllEditor,
    saveExportFile,
    getActiveTab,
    getAllModels,
    getAllTabs,
    getActiveEditor,
    platformUtils,
    openSetting,
    openAttributePanel,
    saveLayout,
    globalCommand,
    expandDocTree,
    openEmoji
};

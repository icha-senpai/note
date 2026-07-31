import {focusByRange} from "./selection";
import {fetchPost, fetchSyncPost} from "../../util/fetch";
import {Constants} from "../../constants";
/// #if !BROWSER
import {ipcRenderer} from "electron";
/// #endif
import {getDefaultSubType, getDefaultType} from "../../search/getDefault";
import {hideMessage, showMessage} from "../../dialog/message";
import {isScribliUriProtocol} from "../../util/pathName";
import {isBrowser} from "../../util/functions";
import type {App} from "../../index";

export const isPhablet = () => {
    return /Android|webOS|iPod|BlackBerry|IEMobile|Opera Mini|Mobile|Tablet/i.test(navigator.userAgent) || isIPhone() || isIPad();
};

export const encodeBase64 = (text: string): string => {
    if (typeof Buffer !== "undefined") {
        return Buffer.from(text, "utf8").toString("base64");
    } else {
        const encoder = new TextEncoder();
        const bytes = encoder.encode(text);
        let binary = "";
        const chunkSize = 0x8000;

        for (let i = 0; i < bytes.length; i += chunkSize) {
            const chunk = bytes.subarray(i, Math.min(i + chunkSize, bytes.length));
            binary += String.fromCharCode(...chunk);
        }

        return btoa(binary);
    }
};

export const getTextScribliFromTextHTML = (html: string) => {
    const internalDataReg = /<!--data-scribli='[^']+'-->/g;
    const legacyInternalDataReg = /<!--data-Scribli='[^']+'-->/g;
    if (html.trimStart().startsWith("<html") &&
        html.substring(0, html.indexOf(">")).includes('xmlns:x="urn:schemas-microsoft-com:office:excel"')) {
        return {
            textScribli: "",
            textHtml: html.replace(internalDataReg, "").replace(legacyInternalDataReg, "")
        };
    }
    const scribliMatch = html.match(/<!--data-scribli='([^']+)'-->/) || html.match(/<!--data-Scribli='([^']+)'-->/);
    let textScribli = "";
    let textHtml = html;
    if (scribliMatch) {
        try {
            if (typeof Buffer !== "undefined") {
                const decodedBytes = Buffer.from(scribliMatch[1], "base64");
                textScribli = decodedBytes.toString("utf8");
            } else {
                const decoder = new TextDecoder();
                const bytes = Uint8Array.from(atob(scribliMatch[1]), char => char.charCodeAt(0));
                textScribli = decoder.decode(bytes);
            }
            textHtml = html.replace(internalDataReg, "").replace(legacyInternalDataReg, "");
        } catch (e) {
            console.log("Failed to decode Scribli data from HTML comment:", e);
        }
    }
    return {
        textScribli,
        textHtml
    };
};

export const saveExportFile = async (uri: string, msgId?: string) => {
    if (!uri) {
        return;
    }
    /// #if !BROWSER
    try {
        const resolved = new URL(uri, `${location.origin}/`);
        const pathSeg = resolved.pathname.substring(resolved.pathname.lastIndexOf("/") + 1);
        let fileName: string;
        try {
            fileName = decodeURIComponent(pathSeg);
        } catch {
            fileName = pathSeg;
        }
        if (!fileName) {
            fileName = "download";
        }
        const result = await ipcRenderer.invoke(Constants.SCRIBLI_GET, {
            cmd: "showSaveDialog",
            defaultPath: fileName,
            properties: ["showOverwriteConfirmation"],
        });
        if (result.canceled || !result.filePath) {
            if (msgId) {
                hideMessage(msgId);
            }
            return;
        }
        const copyResponse = await (await fetch("/api/export/copyExportFile", {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify({
                srcPath: resolved.pathname,
                dest: result.filePath,
            }),
        })).json();
        if (copyResponse.code !== 0) {
            throw new Error(copyResponse.msg);
        }
        if (msgId) {
            hideMessage(msgId);
        }
        showMessage(window.scribli.languages.exported);
        return;
    } catch (e) {
        if (msgId) {
            hideMessage(msgId);
        }
        showMessage("saveExportFile failed: " + e);
    }
    /// #else
    try {
        const openUrl = new URL(uri, `${location.origin}/`);
        openUrl.searchParams.set("download", "true");
        window.open(openUrl.href);
        if (msgId) {
            hideMessage(msgId);
        }
    } catch (e) {
        if (msgId) {
            hideMessage(msgId);
        }
        showMessage("saveExportFile failed: " + e);
    }
    /// #endif
};

export const readText = () => {
    if (typeof navigator.clipboard === "undefined") {
        alert(window.scribli.languages.clipboardPermissionDenied);
        return "";
    }
    return navigator.clipboard.readText().catch(() => {
        alert(window.scribli.languages.clipboardPermissionDenied);
    }) || "";
};

/// #if !BROWSER
export const getLocalFiles = async () => {
    let localFiles: ILocalFiles[] = [];
    if ("darwin" === window.scribli.config.system.os) {
        const xmlString = await ipcRenderer.invoke(Constants.SCRIBLI_GET, {
            cmd: "clipboardRead",
            format: "NSFilenamesPboardType",
        });
        if (xmlString) {
            const domParser = new DOMParser();
            const xmlDom = domParser.parseFromString(xmlString, "application/xml");
            Array.from(xmlDom.getElementsByTagName("string")).forEach(item => {
                localFiles.push({path: item.childNodes[0].nodeValue, size: null});
            });
        }
    } else {
        const xmlString = await fetchSyncPost("/api/clipboard/readFilePaths", {});
        if (xmlString.data.length > 0) {
            localFiles = xmlString.data;
        }
    }
    return localFiles;
};
/// #endif

export const readClipboard = async () => {
    const text: IClipboardData = {textPlain: "", textHTML: "", scribliHTML: ""};
    if (typeof navigator.clipboard === "undefined") {
        alert(window.scribli.languages.clipboardPermissionDenied);
        return text;
    }
    try {
        const clipboardContents = await navigator.clipboard.read().catch(() => {
            alert(window.scribli.languages.clipboardPermissionDenied);
        });
        if (!clipboardContents) {
            return text;
        }
        for (const item of clipboardContents) {
            if (item.types.includes("text/html")) {
                const blob = await item.getType("text/html");
                text.textHTML = await blob.text();
                const textObj = getTextScribliFromTextHTML(text.textHTML);
                text.textHTML = textObj.textHtml;
                text.scribliHTML = textObj.textScribli;
            }
            if (item.types.includes("text/plain")) {
                const blob = await item.getType("text/plain");
                text.textPlain = await blob.text();
            }
            if (item.types.includes("image/png")) {
                const blob = await item.getType("image/png");
                text.files = [new File([blob], "image.png", {type: "image/png", lastModified: Date.now()})];
            }
        }
        /// #if !BROWSER
        if (!text.textHTML && !text.files) {
            text.localFiles = await getLocalFiles();
        }
        /// #endif
        return text;
    } catch (e) {
        return text;
    }
};

export const writeText = (text: string) => {
    let range: Range;
    if (getSelection().rangeCount > 0) {
        range = getSelection().getRangeAt(0).cloneRange();
    }
    try {
        navigator.clipboard.writeText(text);
    } catch (e) {
        const textElement = document.createElement("textarea");
        textElement.value = text;
        textElement.style.position = "fixed";  //avoid scrolling to bottom
        document.body.appendChild(textElement);
        textElement.focus();
        textElement.select();
        document.execCommand("copy");
        document.body.removeChild(textElement);
        if (range) {
            focusByRange(range);
        }
    }
};

export const copyPlainText = (text: string) => {
    text = text.replace(new RegExp(Constants.ZWSP, "g"), "");
    writeText(text);
};

export const getEventName = () => {
    if (isIPhone()) {
        return "touchstart";
    } else {
        return "click";
    }
};

export const isOnlyMeta = (event: KeyboardEvent | MouseEvent) => {
    if (isMac()) {
        // mac
        if (event.metaKey && !event.ctrlKey) {
            return true;
        }
        return false;
    } else {
        if (!event.metaKey && event.ctrlKey) {
            return true;
        }
        return false;
    }
};

export const isNotCtrl = (event: KeyboardEvent | MouseEvent) => {
    if (!event.metaKey && !event.ctrlKey) {
        return true;
    }
    return false;
};

export const isDisabledFeature = (feature: string): boolean => {
    return window.scribli.config.system.disabledFeatures?.indexOf(feature) > -1;
};

export const isIPhone = () => {
    return navigator.userAgent.indexOf("iPhone") > -1;
};

export const isSafari = () => {
    const userAgent = navigator.userAgent;
    return userAgent.includes("Safari") && !userAgent.includes("Chrome") && !userAgent.includes("Chromium");
};

export const isIPad = () => {
    return navigator.userAgent.indexOf("iPad") > -1;
};

export const isMac = () => {
    return navigator.platform.toUpperCase().indexOf("MAC") > -1;
};

export const isWin11 = async () => {
    if (!(navigator as any).userAgentData || !(navigator as any).userAgentData.getHighEntropyValues) {
        return false;
    }
    const ua = await (navigator as any).userAgentData.getHighEntropyValues(["platformVersion"]);
    if ((navigator as any).userAgentData.platform === "Windows") {
        if (parseInt(ua.platformVersion.split(".")[0]) >= 13) {
            return true;
        }
    }
    return false;
};

export const isWindows = () => {
    return navigator.platform.toUpperCase().indexOf("WIN") > -1;
};

export const isInEdge = () => {
    const ua = navigator.userAgent;
    return ua.indexOf("EdgA/") > -1 || ua.indexOf("Edge/") > -1;
};

export function isChromeBrowser(): boolean {
    const nav = window.navigator as Navigator & {
        userAgentData: {
            brands: {
                brand: string;
                version: string;
            }[]
        }
    };
    if (nav.userAgentData && Array.isArray(nav.userAgentData.brands)) {
        const brands = nav.userAgentData.brands.map((b) => b.brand);
        if (brands.some((brand) => /Edge|Opera|OPR/i.test(brand))) {
            return false;
        }
        return brands.some((brand) => /Chrome|Chromium/i.test(brand));
    }
    const ua = nav.userAgent || "";
    const isChromium = /\bChrome\/\d+/i.test(ua) || /\bChromium\/\d+/i.test(ua);
    const isEdge = /\bEdg(e|A|iOS)?\/\d+/i.test(ua); // Edge Chromium
    const isOpera = /\b(OPR|Opera)\/\d+/i.test(ua);

    return isChromium && !isEdge && !isOpera;
}

export const updateHotkeyAfterTip = (hotkey: string, split = " ") => {
    if (hotkey) {
        return split + updateHotkeyTip(hotkey);
    }
    return "";
};

export const updateHotkeyTip = (hotkey: string) => {
    if (!hotkey || isMac()) {
        return hotkey;
    }
    const keys = [];
    if ((hotkey.indexOf("⌘") > -1 || hotkey.indexOf("⌃") > -1)) keys.push("Ctrl");
    if (hotkey.indexOf("⇧") > -1) keys.push("Shift");
    if (hotkey.indexOf("⌥") > -1) keys.push("Alt");

    const lastKey = hotkey.replace(/[⌘⇧⌥⌃]/g, "");
    if (lastKey) {
        keys.push({
            "⇥": "Tab",
            "⌫": "Backspace",
            "⌦": "Delete",
            "↩": "Enter"
        }[lastKey] || lastKey);
    }
    return keys.join("+");
};

export const getLocalStorage = (cb: () => void) => {
    fetchPost("/api/storage/getLocalStorage", undefined, (response) => {
        window.scribli.storage = response.data;
        const defaultStorage: any = {};
        defaultStorage[Constants.LOCAL_SEARCHASSET] = {
            keys: [],
            col: "",
            row: "",
            layout: 0,
            method: 0,
            types: {},
            sort: 0,
            k: "",
        };
        defaultStorage[Constants.LOCAL_SEARCHUNREF] = {
            col: "",
            row: "",
            layout: 0,
        };
        Constants.SCRIBLI_ASSETS_SEARCH.forEach(type => {
            defaultStorage[Constants.LOCAL_SEARCHASSET].types[type] = true;
        });
        defaultStorage[Constants.LOCAL_SEARCHKEYS] = {
            keys: [],
            replaceKeys: [],
            col: "",
            row: "",
            layout: 0,
            colTab: "",
            rowTab: "",
            layoutTab: 0
        };
        defaultStorage[Constants.LOCAL_PDFTHEME] = {
            light: "light",
            dark: "dark",
            annoColor: "var(--b3-pdf-background1)"
        };
        defaultStorage[Constants.LOCAL_LAYOUTS] = [];   // {name: "", layout:{}, time: number, filespaths: IFilesPath[]}
        defaultStorage[Constants.LOCAL_AI] = [];   // {name: "", memo: ""}
        defaultStorage[Constants.LOCAL_PLUGIN_DOCKS] = {};  // { pluginName: {dockId: IPluginDockTab}}
        defaultStorage[Constants.LOCAL_PLUGINTOPUNPIN] = [];
        defaultStorage[Constants.LOCAL_OUTLINE] = {keepCurrentExpand: false};
        defaultStorage[Constants.LOCAL_FILEPOSITION] = {}; // {id: IScrollAttr}
        defaultStorage[Constants.LOCAL_DIALOGPOSITION] = {}; // {id: IPosition}
        defaultStorage[Constants.LOCAL_HISTORY] = {
            notebookId: "%",
            type: 0,
            operation: "all",
            sideWidth: "256px",
            sideDocWidth: "256px",
            sideDiffWidth: "256px",
        };
        defaultStorage[Constants.LOCAL_FLASHCARD] = {
            fullscreen: false
        };
        defaultStorage[Constants.LOCAL_EXPORTWORD] = {removeAssets: false, mergeSubdocs: false};
        defaultStorage[Constants.LOCAL_EXPORTPDF] = {
            landscape: false,
            marginType: "0",
            scale: 1,
            pageSize: "A4",
            removeAssets: true,
            keepFold: false,
            mergeSubdocs: false,
            watermark: false,
            paged: true
        };
        defaultStorage[Constants.LOCAL_EXPORTIMG] = {
            keepFold: false,
            watermark: false
        };
        defaultStorage[Constants.LOCAL_DOCINFO] = {
            id: "",
        };
        defaultStorage[Constants.LOCAL_IMAGES] = {
            file: "1f4c4",
            note: "1f5c3",
            folder: "1f4d1"
        };
        defaultStorage[Constants.LOCAL_EMOJIS] = {
            currentTab: "emoji"
        };
        defaultStorage[Constants.LOCAL_FONTSTYLES] = [];
        defaultStorage[Constants.LOCAL_CLOSED_TABS] = [];
        defaultStorage[Constants.LOCAL_FILESPATHS] = [];    // IFilesPath[]
        defaultStorage[Constants.LOCAL_SEARCHDATA] = {
            removed: true,
            page: 1,
            sort: 0,
            group: 0,
            hasReplace: false,
            method: 0,
            hPath: "",
            idPath: [],
            k: "",
            r: "",
            types: getDefaultType(),
            subTypes: getDefaultSubType(),
            replaceTypes: Object.assign({}, Constants.SCRIBLI_DEFAULT_REPLACETYPES),
        };
        defaultStorage[Constants.LOCAL_ZOOM] = 1;
        defaultStorage[Constants.LOCAL_MOVE_PATH] = {keys: [], k: ""};
        defaultStorage[Constants.LOCAL_RECENT_DOCS] = {type: "viewedAt"};   // TRecentDocsSort

        [Constants.LOCAL_EXPORTIMG, Constants.LOCAL_SEARCHKEYS, Constants.LOCAL_PDFTHEME,
            Constants.LOCAL_EXPORTWORD, Constants.LOCAL_EXPORTPDF, Constants.LOCAL_DOCINFO, Constants.LOCAL_FONTSTYLES,
            Constants.LOCAL_SEARCHDATA, Constants.LOCAL_ZOOM, Constants.LOCAL_LAYOUTS, Constants.LOCAL_AI,
            Constants.LOCAL_PLUGINTOPUNPIN, Constants.LOCAL_SEARCHASSET, Constants.LOCAL_FLASHCARD,
            Constants.LOCAL_DIALOGPOSITION, Constants.LOCAL_SEARCHUNREF, Constants.LOCAL_HISTORY,
            Constants.LOCAL_OUTLINE, Constants.LOCAL_FILEPOSITION, Constants.LOCAL_FILESPATHS, Constants.LOCAL_IMAGES,
            Constants.LOCAL_PLUGIN_DOCKS, Constants.LOCAL_EMOJIS, Constants.LOCAL_MOVE_PATH, Constants.LOCAL_RECENT_DOCS,
            Constants.LOCAL_CLOSED_TABS].forEach((key) => {
            if (typeof response.data[key] === "string") {
                try {
                    const parseData = JSON.parse(response.data[key]);
                    if (typeof parseData === "number") {
                        window.scribli.storage[key] = parseData;
                    } else {
                        window.scribli.storage[key] = Object.assign(defaultStorage[key], parseData);
                    }
                } catch (e) {
                    window.scribli.storage[key] = defaultStorage[key];
                }
            } else if (typeof response.data[key] === "undefined") {
                window.scribli.storage[key] = defaultStorage[key];
            }
        });
        if (!window.scribli.storage[Constants.LOCAL_SEARCHDATA].replaceTypes ||
            Object.keys(window.scribli.storage[Constants.LOCAL_SEARCHDATA].replaceTypes).length === 0) {
            window.scribli.storage[Constants.LOCAL_SEARCHDATA].replaceTypes = Object.assign({}, Constants.SCRIBLI_DEFAULT_REPLACETYPES);
        }
        // Migrate stored search data to include subTypes when absent
        if (!window.scribli.storage[Constants.LOCAL_SEARCHDATA].subTypes ||
            Object.keys(window.scribli.storage[Constants.LOCAL_SEARCHDATA].subTypes).length === 0) {
            window.scribli.storage[Constants.LOCAL_SEARCHDATA].subTypes = getDefaultSubType();
        }
        cb();
    });
};

export const setStorageVal = (key: string, val: any, cb?: () => void) => {
    if (window.scribli.config.readonly || window.scribli.isPublish) {
        return;
    }
    fetchPost("/api/storage/setLocalStorageVal", {
        app: Constants.SCRIBLI_APPID,
        key,
        val,
    }, () => {
        if (cb) {
            cb();
        }
    });
};

export const initWindowOpenOverride = (app: App, openExternal?: (url: string) => void) => {
    const originalOpen = window.open;
    window.open = function (url?: string | URL, target?: string, features?: string): WindowProxy | null {
        const urlStr = typeof url === "string" ? url : (url ? String(url) : "");
        if (isScribliUriProtocol(urlStr) && (!isBrowser() || target !== "_blank")) {
            void import("../../util/uri").then(({processScribliUri}) => processScribliUri(app, urlStr));
            return null;
        }
        return originalOpen.call(window, url, target, features);
    };
};

/// #if !BROWSER
export const initNativeDialogOverride = () => {
    const originalAlert = window.alert;
    const originalConfirm = window.confirm;

    window.alert = function (message: string) {
        try {
            ipcRenderer.sendSync(Constants.SCRIBLI_ALERT_DIALOG, {
                title: window.scribli.languages.scribliNote,
                message,
                buttons: [window.scribli.languages.confirm],
                noLink: true,
            });
            return undefined;
        } catch (error) {
            return originalAlert.call(this, message);
        }
    };

    window.confirm = function (message: string): boolean {
        try {
            const buttonIndex = ipcRenderer.sendSync(Constants.SCRIBLI_CONFIRM_DIALOG, {
                title: window.scribli?.languages?.scribliNote || "Scribli",
                message,
                buttons: [window.scribli?.languages?.cancel || "Cancel", window.scribli?.languages?.confirm || "OK"],
                cancelId: 0,
                defaultId: 1,
                noLink: true,
            });
            return buttonIndex === 1;
        } catch (error) {
            return originalConfirm.call(this, message);
        }
    };
};
/// #endif

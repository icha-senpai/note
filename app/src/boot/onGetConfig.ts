import {adjustLayout, exportLayout, JSONToLayout, resetLayout, resizeTopBar} from "../layout/util";
import {resizeTabs, setTabPosition} from "../layout/tabUtil";
import {initWindowOpenOverride, isWindows, setStorageVal} from "../protyle/util/compatibility";
/// #if !BROWSER
import {initNativeDialogOverride} from "../protyle/util/compatibility";
/// #endif
/// #if !BROWSER
import {ipcRenderer, webFrame} from "electron";
import * as fs from "fs";
import * as path from "path";
import {afterExport} from "../protyle/export/util";
import {onWindowsMsg} from "../window/onWindowsMsg";
/// #endif
import {Constants} from "../constants";
import {appearanceConfigApi} from "../config/tabs/appearanceRuntime";
import {fetchPost, fetchSyncPost} from "../util/fetch";
import {initAssets, setInlineStyle} from "../util/assets";
import {renderSnippet} from "../config/util/snippets";
import {openFile} from "../editor/util";
import {exitScribli} from "../dialog/processSystem";
import {isWindow, setToolbarLeftMac} from "../util/functions";
import {initStatus} from "../layout/status";
import {showMessage} from "../dialog/message";
import {replaceLocalPath} from "../editor/rename";
import {initBar} from "../layout/topBar";
import {openChangelog} from "./openChangelog";
import {App} from "../index";
import {initWindowEvent} from "./globalEvent/event";
import {sendGlobalShortcut} from "./globalEvent/keydown";
import {closeWindow} from "../window/closeWin";
import {correctHotkey} from "./globalEvent/commonHotkey";
import {recordBeforeResizeTop} from "../protyle/util/resize";
import {processScribliUri} from "../util/uri";
import {getAllEditor} from "../layout/getAll";
import {openDesktopOnboarding} from "../onboarding";

export const onGetConfig = (isStart: boolean, app: App) => {
    correctHotkey(app);
    document.body.classList.toggle("body--windows", isWindows());
    /// #if !BROWSER
    ipcRenderer.invoke(Constants.SCRIBLI_INIT, {
        languages: window.scribli.languages["_trayMenu"],
        workspaceDir: window.scribli.config.system.workspaceDir,
        port: location.port
    });
    webFrame.setZoomFactor(window.scribli.storage[Constants.LOCAL_ZOOM]);
    const position = Constants.SIZE_ZOOM.find((item) => item.zoom === window.scribli.storage[Constants.LOCAL_ZOOM]).position;
    ipcRenderer.send(Constants.SCRIBLI_CMD, {
        cmd: "setTrafficLightPosition",
        zoom: window.scribli.storage[Constants.LOCAL_ZOOM],
        position: {
            x: position.x,
            y: (window.scribli.config.appearance.hideToolbar ? 5 * window.scribli.storage[Constants.LOCAL_ZOOM] : 0) + position.y
        },
    });
    /// #endif
    if (!window.scribli.config.uiLayout || (window.scribli.config.uiLayout && !window.scribli.config.uiLayout.left)) {
        window.scribli.config.uiLayout = Constants.SCRIBLI_EMPTY_LAYOUT;
    }
    initWindowEvent(app);
    fetchPost("/api/system/getEmojiConf", {}, response => {
        window.scribli.emojis = response.data as IEmoji[];
        try {
            JSONToLayout(app, isStart);
            setTimeout(() => {
                adjustLayout();
            });
            /// #if !BROWSER
            sendGlobalShortcut(app);
            /// #endif
            openChangelog();
        } catch (e) {
            resetLayout();
        }
        openDesktopOnboarding(app);
    });
    initBar(app);
    initStatus();
    initWindow(app);
    initWindowOpenOverride(app);
    /// #if !BROWSER
    initNativeDialogOverride();
    /// #endif
    appearanceConfigApi.apply(window.scribli.config.appearance);
    initAssets();
    setInlineStyle();
    renderSnippet();
    if (window.scribli.config.system.safeMode) {
        showMessage(window.scribli.languages.safeModeTip);
    }
    let resizeTimeout = 0;
    let firstResize = true;
    window.addEventListener("resize", () => {
        if (firstResize) {
            recordBeforeResizeTop();
            firstResize = false;
        }
        window.clearTimeout(resizeTimeout);
        resizeTimeout = window.setTimeout(() => {
            adjustLayout();
            resizeTabs();
            resizeTopBar();
            setTabPosition(true);
            window.scribli.menus.menu.resetPosition();
            firstResize = true;
            if (getSelection().rangeCount > 0) {
                const range = getSelection().getRangeAt(0);
                getAllEditor().forEach(item => {
                    if (item.protyle.wysiwyg.element.contains(range.startContainer)) {
                        item.protyle.toolbar.render(item.protyle, range);
                    }
                });
            }
            window.scribli.dialogs.forEach(item => {
                item.resize();
            });
        }, Constants.TIMEOUT_RESIZE);
    });
};

export const initWindow = async (app: App) => {
    /// #if !BROWSER
    ipcRenderer.send(Constants.SCRIBLI_CMD, {
        cmd: "setSpellCheckerLanguages",
        languages: window.scribli.config.editor.spellcheckLanguages
    });
    const winOnClose = (close = false) => {
        exportLayout({
            cb() {
                if (window.scribli.config.appearance.closeButtonBehavior === 1 && !close) {
                    if ("windows" === window.scribli.config.system.os) {
                        ipcRenderer.send(Constants.SCRIBLI_CONFIG_TRAY, {
                            languages: window.scribli.languages["_trayMenu"],
                        });
                    } else {
                        ipcRenderer.send(Constants.SCRIBLI_CMD, "closeButtonBehavior");
                    }
                } else {
                    exitScribli();
                }
            },
            errorExit: true
        });
    };

    ipcRenderer.send(Constants.SCRIBLI_EVENT);
    ipcRenderer.on(Constants.SCRIBLI_EVENT, (event, cmd) => {
        if (cmd === "focus") {
            window.scribli.altIsPressed = false;
            window.scribli.ctrlIsPressed = false;
            window.scribli.shiftIsPressed = false;
            document.body.classList.remove("body--blur");
        } else if (cmd === "blur") {
            document.body.classList.add("body--blur");
        } else if (cmd === "enter-full-screen") {
            document.body.classList.add("body--fullscreen");
            setToolbarLeftMac(window.scribli.storage[Constants.LOCAL_ZOOM]);
            setTabPosition();
        } else if (cmd === "leave-full-screen") {
            document.body.classList.remove("body--fullscreen");
            setToolbarLeftMac(window.scribli.storage[Constants.LOCAL_ZOOM]);
            setTabPosition();
        } else if (cmd === "maximize") {
            document.body.classList.add("body--maximize");
        } else if (cmd === "unmaximize") {
            document.body.classList.remove("body--maximize");
        }
    });
    if (!isWindow()) {
        ipcRenderer.on(Constants.SCRIBLI_OPEN_URL, (event, url) => {
            processScribliUri(app, url);
        });
    }
    ipcRenderer.on(Constants.SCRIBLI_OPEN_FILE, (event, data) => {
        if (!data.app) {
            data.app = app;
        }
        openFile(data);
    });
    ipcRenderer.on(Constants.SCRIBLI_SAVE_CLOSE, (event, close) => {
        if (isWindow()) {
            closeWindow(app);
        } else {
            winOnClose(close);
        }
    });
    ipcRenderer.on(Constants.SCRIBLI_SEND_WINDOWS, (e, ipcData: IWebSocketData) => {
        onWindowsMsg(ipcData, app);
    });
    ipcRenderer.on(Constants.SCRIBLI_HOTKEY, (e, data) => {
        let matchCommand = false;
        app.plugins.find(item => {
            item.commands.find(command => {
                if (command.globalCallback && data.hotkey === command.customHotkey) {
                    matchCommand = true;
                    command.globalCallback();
                    return true;
                }
            });
            if (matchCommand) {
                return true;
            }
        });
    });
    ipcRenderer.on(Constants.SCRIBLI_EXPORT_PDF, async (e, ipcData) => {
        const msgId = showMessage(window.scribli.languages.exporting, -1);
        window.scribli.storage[Constants.LOCAL_EXPORTPDF] = {
            removeAssets: ipcData.removeAssets,
            keepFold: ipcData.keepFold,
            mergeSubdocs: ipcData.mergeSubdocs,
            watermark: ipcData.watermark,
            landscape: ipcData.pdfOptions.landscape,
            marginType: ipcData.pdfOptions.marginType,
            pageSize: ipcData.pageSize,
            scale: ipcData.pdfOptions.scale,
            marginTop: ipcData.pdfOptions.margins.top,
            marginRight: ipcData.pdfOptions.margins.right,
            marginBottom: ipcData.pdfOptions.margins.bottom,
            marginLeft: ipcData.pdfOptions.margins.left,
            paged: ipcData.paged,
        };
        setStorageVal(Constants.LOCAL_EXPORTPDF, window.scribli.storage[Constants.LOCAL_EXPORTPDF]);
        try {
            if (window.scribli.config.export.pdfFooter.trim()) {
                const response = await fetchSyncPost("/api/template/renderSprig", {template: window.scribli.config.export.pdfFooter});
                ipcData.pdfOptions.displayHeaderFooter = true;
                ipcData.pdfOptions.headerTemplate = "<span></span>";
                ipcData.pdfOptions.footerTemplate = `<div style="text-align:center;width:100%;font-size:10px;line-height:12px;">
${response.data.replace("%pages", "<span class=totalPages></span>").replace("%page", "<span class=pageNumber></span>")}
</div>`;
            }
            const pdfData = await ipcRenderer.invoke(Constants.SCRIBLI_GET, {
                cmd: "printToPDF",
                pdfOptions: ipcData.pdfOptions,
                webContentsId: ipcData.webContentsId
            });
            const savePath = ipcData.filePaths[0];
            let pdfFilePath = path.join(savePath, replaceLocalPath(ipcData.rootTitle) + ".pdf");
            const responseUnique = await fetchSyncPost("/api/file/getUniqueFilename", {path: pdfFilePath});
            pdfFilePath = responseUnique.data.path;
            fetchPost("/api/export/exportHTML", {
                id: ipcData.rootId,
                pdf: true,
                removeAssets: ipcData.removeAssets,
                merge: ipcData.mergeSubdocs,
                savePath,
            }, () => {
                fs.writeFileSync(pdfFilePath, pdfData);
                ipcRenderer.send(Constants.SCRIBLI_CMD, {cmd: "destroy", webContentsId: ipcData.webContentsId});
                fetchPost("/api/export/processPDF", {
                    id: ipcData.rootId,
                    merge: ipcData.mergeSubdocs,
                    path: pdfFilePath,
                    removeAssets: ipcData.removeAssets,
                    watermark: ipcData.watermark
                }, async () => {
                    afterExport(pdfFilePath, msgId);
                    if (ipcData.removeAssets) {
                        const removePromise = (dir: string) => {
                            return new Promise(function (resolve) {
                                fs.stat(dir, function (err, stat) {
                                    if (!stat) {
                                        return;
                                    }

                                    if (stat.isDirectory()) {
                                        fs.readdir(dir, function (err, files) {
                                            files = files.map(file => path.join(dir, file)); // a/b  a/m
                                            Promise.all(files.map(file => removePromise(file))).then(function () {
                                                fs.rm(dir, resolve);
                                            });
                                        });
                                    } else {
                                        fs.unlink(dir, resolve);
                                    }
                                });
                            });
                        };

                        const assetsDir = path.join(savePath, "assets");
                        await removePromise(assetsDir);
                        if (1 > fs.readdirSync(assetsDir).length) {
                            fs.rmdirSync(assetsDir);
                        }
                    }
                });
            });
        } catch (e) {
            console.error(e);
            showMessage(window.scribli.languages.exportPDFLowMemory, 0, "error", msgId);
            ipcRenderer.send(Constants.SCRIBLI_CMD, {cmd: "destroy", webContentsId: ipcData.webContentsId});
        }
        ipcRenderer.send(Constants.SCRIBLI_CMD, {cmd: "hide", webContentsId: ipcData.webContentsId});
    });

    if (isWindow()) {
        const isAlwaysOnTop = await ipcRenderer.invoke(Constants.SCRIBLI_GET, {
            cmd: "isAlwaysOnTop",
        });
        document.body.insertAdjacentHTML("beforeend", `<div class="toolbar__window">
<div class="toolbar__window-drag"></div>
<div class="toolbar__item ariaLabel" aria-label="${window.scribli.languages[isAlwaysOnTop ? "unpin" : "pin"]}" id="pinWindow">
    <svg>
        <use xlink:href="#icon${isAlwaysOnTop ? "Unpin" : "Pin"}"></use>
    </svg>
</div></div>`);
        const pinElement = document.getElementById("pinWindow");
        pinElement.addEventListener("click", () => {
            if (pinElement.getAttribute("aria-label") === window.scribli.languages.pin) {
                pinElement.querySelector("use").setAttribute("xlink:href", "#iconUnpin");
                pinElement.setAttribute("aria-label", window.scribli.languages.unpin);
                ipcRenderer.send(Constants.SCRIBLI_CMD, "setAlwaysOnTopTrue");
            } else {
                pinElement.querySelector("use").setAttribute("xlink:href", "#iconPin");
                pinElement.setAttribute("aria-label", window.scribli.languages.pin);
                ipcRenderer.send(Constants.SCRIBLI_CMD, "setAlwaysOnTopFalse");
            }
        });
    }

    const isFullScreen = await ipcRenderer.invoke(Constants.SCRIBLI_GET, {
        cmd: "isFullScreen",
    });
    if (isFullScreen) {
        document.body.classList.add("body--fullscreen");
    }
    setToolbarLeftMac(window.scribli.storage[Constants.LOCAL_ZOOM]);
    const isMaximized = await ipcRenderer.invoke(Constants.SCRIBLI_GET, {
        cmd: "isMaximized",
    });
    if (isMaximized) {
        document.body.classList.add("body--maximize");
    }

    if ("darwin" !== window.scribli.config.system.os) {
        document.body.classList.add("body--win32");

        const controlsHTML = `<div class="toolbar__item ariaLabel toolbar__item--win" aria-label="${window.scribli.languages.min}" id="minWindow">
    <svg>
        <use xlink:href="#iconMin"></use>
    </svg>
</div>
<div aria-label="${window.scribli.languages.max}" class="ariaLabel toolbar__item toolbar__item--win" id="maxWindow">
    <svg>
        <use xlink:href="#iconMax"></use>
    </svg>
</div>
<div aria-label="${window.scribli.languages.restore}" class="ariaLabel toolbar__item toolbar__item--win" id="restoreWindow">
    <svg>
        <use xlink:href="#iconRestore"></use>
    </svg>
</div>
<div aria-label="${window.scribli.languages.close}" class="ariaLabel toolbar__item toolbar__item--close" id="closeWindow">
    <svg>
        <use xlink:href="#iconClose"></use>
    </svg>
</div>`;
        if (isWindow()) {
            document.querySelector(".toolbar__window").insertAdjacentHTML("beforeend", controlsHTML);
        } else {
            document.getElementById("windowControls").innerHTML = controlsHTML;
        }
        const maxBtnElement = document.getElementById("maxWindow");
        const restoreBtnElement = document.getElementById("restoreWindow");

        restoreBtnElement.addEventListener("click", () => {
            ipcRenderer.send(Constants.SCRIBLI_CMD, "restore");
        });
        maxBtnElement.addEventListener("click", () => {
            ipcRenderer.send(Constants.SCRIBLI_CMD, "maximize");
        });

        const minBtnElement = document.getElementById("minWindow");
        const closeBtnElement = document.getElementById("closeWindow");
        minBtnElement.addEventListener("click", () => {
            if (minBtnElement.classList.contains("window-controls__item--disabled")) {
                return;
            }
            ipcRenderer.send(Constants.SCRIBLI_CMD, "minimize");
        });
        closeBtnElement.addEventListener("click", () => {
            if (isWindow()) {
                closeWindow(app);
            } else {
                winOnClose();
            }
        });
    }
    /// #else
    if (!isWindow()) {
        document.querySelector(".toolbar").classList.add("toolbar--browser");
    }
    if (isWindows()) {
        document.body.classList.add("body--win32-browser");
    }
    /// #endif
};

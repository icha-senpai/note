import {
    isInMobileApp,
    setStorageVal,
    updateHotkeyTip
} from "../protyle/util/compatibility";
import {exitScribli, processSync} from "../dialog/processSystem";
import {goBack, goForward} from "../util/backForward";
import {syncGuide} from "../sync/syncGuide";
import {workspaceMenu} from "../menus/workspace";
import {MenuItem} from "../menus/Menu";
import {setMode} from "../util/assets";
import {openSetting} from "../config";
import {openSearch} from "../search/spread";
import {App} from "../index";
/// #if !BROWSER
import {ipcRenderer, webFrame} from "electron";
/// #endif
import {Constants} from "../constants";
import {isBrowser, isWindow, setToolbarLeftMac} from "../util/functions";
import {fetchPost} from "../util/fetch";
import {hasFeatureAccess} from "../util/featureAccess";
import * as dayjs from "dayjs";
import {exportLayout, resizeTopBar} from "./util";
import {setTabPosition} from "./tabUtil";
import {commandPanel} from "../boot/globalEvent/command/panel";
import {openTopBarMenu} from "../plugin/openTopBarMenu";
import {getWorkspaceName, setTitle} from "../util/processTitle";

const sendTrafficLightPosition = (zoom: number) => {
    /// #if !BROWSER
    const position = Constants.SIZE_ZOOM.find((item) => item.zoom === zoom).position;
    ipcRenderer.send(Constants.SCRIBLI_CMD, {
        cmd: "setTrafficLightPosition",
        zoom,
        position: {
            x: position.x,
            y: ((window.scribli.config.appearance.hideToolbar && !isWindow()) ? 5 * zoom : 0) + position.y,
        },
    });
    /// #endif
};

/** 同步顶栏隐藏后的布局（运行时切换 hideToolbar 时调用） */
export const syncHideToolbarLayout = () => {
    document.body.classList.toggle("body--toolbar-hide", window.scribli.config.appearance.hideToolbar);
    resizeTopBar();
    /// #if !BROWSER
    if (!isWindow()) {
        sendTrafficLightPosition(window.scribli.storage[Constants.LOCAL_ZOOM]);
        if (!window.scribli.config.appearance.hideToolbar) {
            const title = document.querySelector('.layout-tab-bar .item--focus[data-type="tab-header"] .item__text')?.textContent || "";
            setTitle(title, title ? false : true);
        }
    } else {
        return;
    }
    /// #endif
    setTabPosition(false, true);
};

export const updateBarModeIcon = () => {
    document.querySelector("#barMode use")?.setAttribute(
        "xlink:href",
        `#icon${window.scribli.config.appearance.modeOS ? "Mode" : (window.scribli.config.appearance.mode === 0 ? "Light" : "Dark")}`
    );
};

export const initBar = (app: App) => {
    const toolbarElement = document.getElementById("toolbar");
    toolbarElement.innerHTML = `
<div id="barWorkspace" class="ariaLabel toolbar__item" aria-label="${window.scribli.languages.mainMenu} ${updateHotkeyTip(window.scribli.config.keymap.general.mainMenu.custom)}">
    <span class="toolbar__text">${getWorkspaceName()}</span>
    <svg class="toolbar__svg"><use xlink:href="#iconDown"></use></svg>
</div>
<div id="barSync" class="ariaLabel toolbar__item${window.scribli.config.readonly ? " fn__none" : ""}">
    <svg><use xlink:href="#iconCloudSucc"></use></svg>
</div>
<button id="barBack" class="ariaLabel toolbar__item toolbar__item--disabled" aria-label="${window.scribli.languages.goBack} ${updateHotkeyTip(window.scribli.config.keymap.general.goBack.custom)}">
    <svg><use xlink:href="#iconBack"></use></svg>
</button>
<button id="barForward" class="ariaLabel toolbar__item toolbar__item--disabled" aria-label="${window.scribli.languages.goForward} ${updateHotkeyTip(window.scribli.config.keymap.general.goForward.custom)}">
    <svg><use xlink:href="#iconForward"></use></svg>
</button>
<div class="fn__flex-1 fn__ellipsis" id="drag"><span class="fn__none">开发版，使用前请进行备份 Development version, please backup before use</span></div>
<div id="toolbarAccount" class="fn__flex${window.scribli.config.readonly ? " fn__none" : ""}"></div>
<div id="barPlugins" class="toolbar__item ariaLabel" aria-label="${window.scribli.languages.plugin}">
    <svg><use xlink:href="#iconPlugin"></use></svg>
</div>
<div id="barCommand" class="toolbar__item ariaLabel" aria-label="${window.scribli.languages.commandPanel} ${updateHotkeyTip(window.scribli.config.keymap.general.commandPanel.custom)}">
    <svg><use xlink:href="#iconTerminal"></use></svg>
</div>
<div id="barSearch" class="toolbar__item ariaLabel" aria-label="${window.scribli.languages.globalSearch} ${updateHotkeyTip(window.scribli.config.keymap.general.globalSearch.custom)}">
    <svg><use xlink:href="#iconSearch"></use></svg>
</div>
<div id="barZoom" class="toolbar__item ariaLabel${(window.scribli.storage[Constants.LOCAL_ZOOM] === 1 || isBrowser()) ? " fn__none" : ""}" aria-label="${window.scribli.languages.zoom}">
    <svg><use xlink:href="#iconZoom${window.scribli.storage[Constants.LOCAL_ZOOM] > 1 ? "In" : "Out"}"></use></svg>
</div>
<div id="barMode" class="toolbar__item ariaLabel${window.scribli.config.readonly ? " fn__none" : ""}" aria-label="${window.scribli.languages.appearanceMode}">
    <svg><use xlink:href="#icon${window.scribli.config.appearance.modeOS ? "Mode" : (window.scribli.config.appearance.mode === 0 ? "Light" : "Dark")}"></use></svg>
</div>
<div id="barExit" class="ft__error toolbar__item ariaLabel${isInMobileApp() ? "" : " fn__none"}" aria-label="${window.scribli.languages.safeQuit}">
    <svg><use xlink:href="#iconQuit"></use></svg>
</div>
<div id="barMore" class="toolbar__item ariaLabel" aria-label="${window.scribli.languages.more}">
    <svg><use xlink:href="#iconMore"></use></svg>
</div>
<div class="fn__flex" id="windowControls"></div>`;
    processSync();
    toolbarElement.addEventListener("click", (event: MouseEvent) => {
        let target = event.target as HTMLElement;
        if (typeof event.detail === "string") {
            target = toolbarElement.querySelector("#" + event.detail);
        }
        while (!target.classList.contains("toolbar")) {
            const targetId = typeof event.detail === "string" ? event.detail : target.id;
            if (targetId === "barBack") {
                goBack(app);
                event.stopPropagation();
                break;
            } else if (targetId === "barMore") {
                if (!window.scribli.menus.menu.element.classList.contains("fn__none") &&
                    window.scribli.menus.menu.element.getAttribute("data-name") === Constants.MENU_BAR_MORE) {
                    window.scribli.menus.menu.remove();
                    return;
                }
                window.scribli.menus.menu.remove();
                window.scribli.menus.menu.element.setAttribute("data-name", Constants.MENU_BAR_MORE);
                (target.getAttribute("data-hideids") || "").split(",").forEach((itemId) => {
                    // data-hideids 可能为空字符串，split(",") 会得到 [""]，导致 querySelector("#") 抛出无效选择器异常
                    if (!itemId) {
                        return;
                    }
                    const hideElement = toolbarElement.querySelector("#" + itemId);
                    const useElement = hideElement.querySelector("use");
                    const menuOptions: IMenu = {
                        label: itemId === "toolbarAccount" ? window.scribli.languages.account : hideElement.getAttribute("aria-label"),
                        icon: itemId === "toolbarAccount" ? "iconAccount" : (useElement ? useElement.getAttribute("xlink:href").substring(1) : undefined),
                        click: () => {
                            if (itemId.startsWith("plugin")) {
                                hideElement.dispatchEvent(new CustomEvent("click"));
                            } else {
                                toolbarElement.dispatchEvent(new CustomEvent("click", {detail: itemId}));
                            }
                        }
                    };
                    if (!useElement && hideElement.querySelector("svg")) {
                        const svgElement = hideElement.querySelector("svg").cloneNode(true) as HTMLElement;
                        svgElement.classList.add("b3-menu__icon");
                        menuOptions.iconHTML = svgElement.outerHTML;
                    }
                    window.scribli.menus.menu.append(new MenuItem(menuOptions).element);
                });
                const rect = target.getBoundingClientRect();
                window.scribli.menus.menu.popup({x: rect.right, y: rect.bottom, isLeft: true});
                event.stopPropagation();
                break;
            } else if (targetId === "barForward") {
                goForward(app);
                event.stopPropagation();
                break;
            } else if (targetId === "barSync") {
                syncGuide(app);
                event.stopPropagation();
                break;
            } else if (targetId === "barWorkspace") {
                workspaceMenu(app, target.getBoundingClientRect());
                event.stopPropagation();
                break;
            } else if (targetId === "barExit") {
                event.stopPropagation();
                exportLayout({
                    errorExit: true,
                    cb: exitScribli,
                });
                break;
            } else if (targetId === "barMode") {
                if (!window.scribli.menus.menu.element.classList.contains("fn__none") &&
                    window.scribli.menus.menu.element.getAttribute("data-name") === Constants.MENU_BAR_MODE) {
                    window.scribli.menus.menu.remove();
                    return;
                }
                window.scribli.menus.menu.remove();
                window.scribli.menus.menu.element.setAttribute("data-name", Constants.MENU_BAR_MODE);
                window.scribli.menus.menu.append(new MenuItem({
                    id: "themeLight",
                    label: window.scribli.languages.themeLight,
                    icon: "iconLight",
                    current: window.scribli.config.appearance.mode === 0 && !window.scribli.config.appearance.modeOS,
                    click: () => {
                        setMode(0);
                    }
                }).element);
                window.scribli.menus.menu.append(new MenuItem({
                    id: "themeDark",
                    label: window.scribli.languages.themeDark,
                    current: window.scribli.config.appearance.mode === 1 && !window.scribli.config.appearance.modeOS,
                    icon: "iconDark",
                    click: () => {
                        setMode(1);
                    }
                }).element);
                window.scribli.menus.menu.append(new MenuItem({
                    id: "themeOS",
                    label: window.scribli.languages.themeOS,
                    current: window.scribli.config.appearance.modeOS,
                    icon: "iconMode",
                    click: () => {
                        setMode(2);
                    }
                }).element);
                let rect = target.getBoundingClientRect();
                if (rect.width === 0) {
                    rect = toolbarElement.querySelector("#barMore").getBoundingClientRect();
                }
                window.scribli.menus.menu.popup({x: rect.right, y: rect.bottom, isLeft: true});
                event.stopPropagation();
                break;
            } else if (targetId === "toolbarAccount") {
                if (!window.scribli.config.readonly) {
                    openSetting(app, "sync");
                }
                event.stopPropagation();
                break;
            } else if (targetId === "barSearch") {
                openSearch({
                    app,
                    hotkey: Constants.DIALOG_GLOBALSEARCH
                });
                event.stopPropagation();
                break;
            } else if (targetId === "barPlugins") {
                openTopBarMenu(app, target);
                event.stopPropagation();
                break;
            } else if (targetId === "barCommand") {
                commandPanel(app);
                event.stopPropagation();
                break;
            } else if (targetId === "barZoom") {
                if (!window.scribli.menus.menu.element.classList.contains("fn__none") &&
                    window.scribli.menus.menu.element.getAttribute("data-name") === Constants.MENU_BAR_ZOOM) {
                    window.scribli.menus.menu.remove();
                    return;
                }
                window.scribli.menus.menu.remove();
                window.scribli.menus.menu.element.setAttribute("data-name", Constants.MENU_BAR_ZOOM);
                window.scribli.menus.menu.append(new MenuItem({
                    label: window.scribli.languages.zoomIn,
                    icon: "iconZoomIn",
                    accelerator: "⌘=",
                    click: () => {
                        setZoom("zoomIn");
                    }
                }).element);
                window.scribli.menus.menu.append(new MenuItem({
                    label: window.scribli.languages.zoomOut,
                    accelerator: "⌘-",
                    icon: "iconZoomOut",
                    click: () => {
                        setZoom("zoomOut");
                    }
                }).element);
                window.scribli.menus.menu.append(new MenuItem({
                    label: window.scribli.languages.reset,
                    accelerator: "⌘0",
                    click: () => {
                        setZoom("restore");
                    }
                }).element);
                let rect = target.getBoundingClientRect();
                if (rect.width === 0) {
                    rect = toolbarElement.querySelector("#barMore").getBoundingClientRect();
                }
                window.scribli.menus.menu.popup({x: rect.right, y: rect.bottom, isLeft: true});
                event.stopPropagation();
                break;
            }
            target = target.parentElement;
        }
    });
    const barSyncElement = toolbarElement.querySelector("#barSync");
    barSyncElement.addEventListener("mouseenter", (event) => {
        event.stopPropagation();
        event.preventDefault();
        fetchPost("/api/sync/getSyncInfo", {}, (response) => {
            let html = "";
            if (!window.scribli.config.sync.enabled || (0 === window.scribli.config.sync.provider && !hasFeatureAccess())) {
                html = response.data.stat;
            } else {
                html = window.scribli.languages._kernel[82].replace("%s", dayjs(response.data.synced).format("YYYY-MM-DD HH:mm")) + "<br>";
                html += "&emsp;" + response.data.stat;
                if (response.data.kernels.length > 0) {
                    html += "<br>";
                    html += window.scribli.languages.currentKernel + "<br>";
                    html += "&emsp;" + response.data.kernel + "/" + window.scribli.config.system.kernelVersion + " (" + window.scribli.config.system.os + "/" + window.scribli.config.system.name + ")<br>";
                    html += window.scribli.languages.otherOnlineKernels + "<br>";
                    response.data.kernels.forEach((item: {
                        os: string;
                        ver: string;
                        hostname: string;
                        id: string;
                    }) => {
                        html += `&emsp;${item.id}/${item.ver} (${item.os}/${item.hostname}) <br>`;
                    });
                }
            }
            barSyncElement.setAttribute("aria-label", html);
        });
    });
    barSyncElement.setAttribute("aria-label", window.scribli.config.sync.stat || (window.scribli.languages.syncNow + " " + updateHotkeyTip(window.scribli.config.keymap.general.syncNow.custom)));
    if (window.scribli.config.appearance.hideToolbar) {
        document.body.classList.add("body--toolbar-hide");
    }
};

export const setZoom = (type: "zoomIn" | "zoomOut" | "restore") => {
    /// #if !BROWSER
    let zoom = 1;
    if (type === "zoomIn") {
        Constants.SIZE_ZOOM.find((item, index) => {
            if (item.zoom === window.scribli.storage[Constants.LOCAL_ZOOM]) {
                zoom = Constants.SIZE_ZOOM[index + 1]?.zoom || 3;
                return true;
            }
        });
    } else if (type === "zoomOut") {
        Constants.SIZE_ZOOM.find((item, index) => {
            if (item.zoom === window.scribli.storage[Constants.LOCAL_ZOOM]) {
                zoom = Constants.SIZE_ZOOM[index - 1]?.zoom || 0.67;
                return true;
            }
        });
    }

    webFrame.setZoomFactor(zoom);
    setToolbarLeftMac(zoom);
    sendTrafficLightPosition(zoom);
    window.scribli.storage[Constants.LOCAL_ZOOM] = zoom;
    setStorageVal(Constants.LOCAL_ZOOM, zoom);
    if (!isWindow()) {
        const barZoomElement = document.getElementById("barZoom");
        if (zoom === 1) {
            barZoomElement.classList.add("fn__none");
        } else {
            if (zoom > 1) {
                barZoomElement.querySelector("use").setAttribute("xlink:href", "#iconZoomIn");
            } else {
                barZoomElement.querySelector("use").setAttribute("xlink:href", "#iconZoomOut");
            }
            barZoomElement.classList.remove("fn__none");
        }
    }
    /// #endif
};

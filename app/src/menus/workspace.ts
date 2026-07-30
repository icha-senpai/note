import {MenuItem} from "./Menu";
/// #if !BROWSER
import {ipcRenderer} from "electron";
/// #endif
import {openHistory} from "../history/history";
import {getOpenNotebookCount, originalPath, pathPosix, useShell} from "../util/pathName";
import {fetchNewDailyNote, mountHelp, newDailyNote} from "../util/mount";
import {fetchPost} from "../util/fetch";
import {Constants} from "../constants";
import {isIPad, setStorageVal, writeText} from "../protyle/util/compatibility";
import {openCard} from "../card/openCard";
import {openSetting} from "../config";
import {getAllDocks} from "../layout/getAll";
import {exportLayout, getAllLayout} from "../layout/util";
import {getDockByType} from "../layout/tabUtil";
import {exitScribli, lockScreen} from "../dialog/processSystem";
import {showMessage} from "../dialog/message";
import {unicode2Emoji} from "../emoji";
import {Dock} from "../layout/dock";
import {escapeAttr, escapeHtml} from "../util/escape";
import {viewCards} from "../card/viewCards";
import {Dialog} from "../dialog";
import {hasClosestByClassName} from "../protyle/util/hasClosest";
import {confirmDialog} from "../dialog/confirmDialog";
import {App} from "../index";
import {isBrowser} from "../util/functions";
import {openRecentDocs} from "../business/openRecentDocs";
import * as dayjs from "dayjs";
import {upDownHint} from "../util/upDownHint";
import {openDataMigration} from "./dataMigration";

const editLayout = (layoutName?: string) => {
    const dialog = new Dialog({
        positionId: Constants.DIALOG_SAVEWORKSPACE,
        title: layoutName ? window.scribli.languages.edit : window.scribli.languages.save,
        content: `<div class="b3-dialog__content">
        <input class="b3-text-field fn__block" value="${layoutName || ""}" placeholder="${window.scribli.languages.memo}">
</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--remove${layoutName ? "" : " fn__none"}">${window.scribli.languages.delete}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text${layoutName ? "" : " fn__none"}">${window.scribli.languages.rename}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages[layoutName ? "updateLayout" : "confirm"]}</button>
</div>`,
        width: "520px",
    });
    dialog.element.setAttribute("data-key", Constants.DIALOG_SAVEWORKSPACE);
    const btnsElement = dialog.element.querySelectorAll(".b3-button");
    const inputElement = dialog.element.querySelector("input");
    inputElement.select();
    inputElement.focus();
    dialog.bindInput(inputElement, () => {
        btnsElement[3].dispatchEvent(new CustomEvent("click"));
    });
    btnsElement[0].addEventListener("click", () => {
        window.scribli.storage[Constants.LOCAL_LAYOUTS].find((layoutItem: ISaveLayout, index: number) => {
            if (layoutItem.name === layoutName) {
                window.scribli.storage[Constants.LOCAL_LAYOUTS].splice(index, 1);
                setStorageVal(Constants.LOCAL_LAYOUTS, window.scribli.storage[Constants.LOCAL_LAYOUTS]);
                return true;
            }
        });
        dialog.destroy();
    });
    btnsElement[1].addEventListener("click", () => {
        dialog.destroy();
    });
    btnsElement[2].addEventListener("click", () => {
        const value = inputElement.value;
        if (!value) {
            showMessage(window.scribli.languages["_kernel"]["142"]);
            return;
        }
        dialog.destroy();
        window.scribli.storage[Constants.LOCAL_LAYOUTS].find((layoutItem: ISaveLayout) => {
            if (layoutItem.name === layoutName) {
                layoutItem.name = value;
                layoutItem.time = Date.now();
                setStorageVal(Constants.LOCAL_LAYOUTS, window.scribli.storage[Constants.LOCAL_LAYOUTS]);
                return true;
            }
        });
    });
    btnsElement[3].addEventListener("click", () => {
        const value = inputElement.value;
        if (!value) {
            showMessage(window.scribli.languages["_kernel"]["142"]);
            return;
        }
        dialog.destroy();
        if (layoutName) {
            window.scribli.storage[Constants.LOCAL_LAYOUTS].find((layoutItem: ISaveLayout) => {
                if (layoutItem.name === layoutName) {
                    layoutItem.name = value;
                    layoutItem.time = Date.now();
                    layoutItem.layout = getAllLayout();
                    layoutItem.filesPaths = window.scribli.storage[Constants.LOCAL_FILESPATHS];
                    setStorageVal(Constants.LOCAL_LAYOUTS, window.scribli.storage[Constants.LOCAL_LAYOUTS]);
                    return true;
                }
            });
            return;
        }
        const hadName = window.scribli.storage[Constants.LOCAL_LAYOUTS].find((item: ISaveLayout) => {
            if (item.name === value) {
                confirmDialog(window.scribli.languages.save, window.scribli.languages.exportTplTip, () => {
                    item.layout = getAllLayout();
                    item.time = Date.now();
                    item.filesPaths = window.scribli.storage[Constants.LOCAL_FILESPATHS];
                    setStorageVal(Constants.LOCAL_LAYOUTS, window.scribli.storage[Constants.LOCAL_LAYOUTS]);
                });
                return true;
            }
        });
        if (hadName) {
            return;
        }
        window.scribli.storage[Constants.LOCAL_LAYOUTS].push({
            name: value,
            time: Date.now(),
            layout: getAllLayout(),
            filesPaths: window.scribli.storage[Constants.LOCAL_FILESPATHS]
        });
        setStorageVal(Constants.LOCAL_LAYOUTS, window.scribli.storage[Constants.LOCAL_LAYOUTS]);
    });
};

const togglePinDock = (id: "switchLeftDock" | "switchRightDock" | "switchBottomDock", dock: Dock, pinIcon: string, unpinIcon: string) => {
    return {
        id,
        label: `${dock.pin ? window.scribli.languages.switchToFloatingLayout : window.scribli.languages.switchToFixedLayout}`,
        icon: `${dock.pin ? unpinIcon : pinIcon}`,
        accelerator: window.scribli.config.keymap.general[id].custom,
        current: !dock.pin,
        click() {
            dock.togglePin();
        }
    };
};

export const workspaceMenu = (app: App, rect: DOMRect) => {
    if (!window.scribli.menus.menu.element.classList.contains("fn__none") &&
        window.scribli.menus.menu.element.getAttribute("data-name") === Constants.MENU_BAR_WORKSPACE) {
        window.scribli.menus.menu.remove();
        return;
    }
    fetchPost("/api/system/getWorkspaces", {}, (response) => {
        window.scribli.menus.menu.remove();
        window.scribli.menus.menu.element.setAttribute("data-name", Constants.MENU_BAR_WORKSPACE);
        if (!window.scribli.config.readonly) {
            window.scribli.menus.menu.append(new MenuItem({
                id: "config",
                label: window.scribli.languages.config,
                icon: "iconSettings",
                accelerator: window.scribli.config.keymap.general.config.custom,
                click: () => {
                    openSetting(app);
                }
            }).element);
        }
        const dockMenu: IMenu[] = [];
        getAllDocks().forEach(item => {
            dockMenu.push({
                id: item.type,
                icon: item.icon,
                accelerator: item.hotkey,
                label: item.title,
                click() {
                    getDockByType(item.type).toggleModel(item.type);
                }
            });
        });
        if (!window.scribli.config.readonly) {
            dockMenu.push({id: "separator_1", type: "separator"});
            dockMenu.push(togglePinDock("switchLeftDock", window.scribli.layout.leftDock, "iconPanelLeft", "iconPanelLeftDashed"));
            dockMenu.push(togglePinDock("switchRightDock", window.scribli.layout.rightDock, "iconPanelRight", "iconPanelRightDashed"));
            dockMenu.push(togglePinDock("switchBottomDock", window.scribli.layout.bottomDock, "iconPanelBottom", "iconPanelBottomDashed"));
        }
        window.scribli.menus.menu.append(new MenuItem({
            id: "panels",
            label: window.scribli.languages.panels,
            icon: "iconDock",
            type: "submenu",
            submenu: dockMenu
        }).element);
        if (!window.scribli.config.readonly) {
            let workspaceSubMenu: IMenu[];
            /// #if !BROWSER
            workspaceSubMenu = [{
                id: "newOrOpenBy",
                label: `${window.scribli.languages.new} / ${window.scribli.languages.openBy}`,
                iconHTML: "",
                click: async () => {
                    const localPath = await ipcRenderer.invoke(Constants.SCRIBLI_GET, {
                        cmd: "showOpenDialog",
                        defaultPath: window.scribli.config.system.homeDir,
                        properties: ["openDirectory", "createDirectory"],
                    });
                    if (localPath.filePaths.length === 0) {
                        return;
                    }
                    fetchPost("/api/system/checkWorkspaceDir", {path: localPath.filePaths[0]}, (response) => {
                        if (response.data.isWorkspace) {
                            openWorkspace(localPath.filePaths[0]);
                        } else {
                            confirmDialog("🏗️ " + window.scribli.languages.createWorkspace, window.scribli.languages.createWorkspaceTip + `<br><br><code class="fn__code">${localPath.filePaths[0]}</code>`, () => {
                                openWorkspace(localPath.filePaths[0]);
                            });
                        }
                    });
                }
            }];
            workspaceSubMenu.push({id: "separator_1", type: "separator"});
            response.data.forEach((item: IWorkspace) => {
                workspaceSubMenu.push(workspaceItem(item) as IMenu);
            });
            /// #else
            workspaceSubMenu = [{
                id: "new",
                label: window.scribli.languages.new,
                iconHTML: "",
                click() {
                    const createWorkspaceDialog = new Dialog({
                        title: window.scribli.languages.new,
                        content: `<div class="b3-dialog__content">
    <input class="b3-text-field fn__block">
</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages.confirm}</button>
</div>`,
                        width: "520px",
                    });
                    createWorkspaceDialog.element.setAttribute("data-key", Constants.DIALOG_CREATEWORKSPACE);
                    const inputElement = createWorkspaceDialog.element.querySelector("input");
                    inputElement.focus();
                    const btnsElement = createWorkspaceDialog.element.querySelectorAll(".b3-button");
                    btnsElement[0].addEventListener("click", () => {
                        createWorkspaceDialog.destroy();
                    });
                    btnsElement[1].addEventListener("click", () => {
                        fetchPost("/api/system/createWorkspaceDir", {
                            path: pathPosix().join(pathPosix().dirname(window.scribli.config.system.workspaceDir), inputElement.value)
                        }, () => {
                            createWorkspaceDialog.destroy();
                        });
                    });
                }
            }, {
                id: "openBy",
                label: `${window.scribli.languages.openBy}...`,
                iconHTML: "",
                click() {
                    fetchPost("/api/system/getWorkspaces", {}, (response) => {
                        let selectHTML = "";
                        response.data.forEach((item: IWorkspace, index: number) => {
                            selectHTML += `<option value="${item.path}"${index === 0 ? ' selected="selected"' : ""}>${pathPosix().basename(item.path)}</option>`;
                        });
                        const openWorkspaceDialog = new Dialog({
                            title: window.scribli.languages.openBy,
                            content: `<div class="b3-dialog__content">
    <select class="b3-text-field fn__block">${selectHTML}</select>
</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages.confirm}</button>
</div>`,
                            width: "520px",
                        });
                        openWorkspaceDialog.element.setAttribute("data-key", Constants.DIALOG_OPENWORKSPACE);
                        const btnsElement = openWorkspaceDialog.element.querySelectorAll(".b3-button");
                        btnsElement[0].addEventListener("click", () => {
                            openWorkspaceDialog.destroy();
                        });
                        btnsElement[1].addEventListener("click", () => {
                            const openPath = openWorkspaceDialog.element.querySelector("select").value;
                            if (openPath === window.scribli.config.system.workspaceDir) {
                                openWorkspaceDialog.destroy();
                                return;
                            }
                            confirmDialog(window.scribli.languages.confirm, `${pathPosix().basename(window.scribli.config.system.workspaceDir)} -> ${pathPosix().basename(openPath)}?`, () => {
                                fetchPost("/api/system/setWorkspaceDir", {
                                    path: openPath
                                }, () => {
                                    exitScribli(false);
                                });
                            });
                        });
                    });
                }
            }];
            workspaceSubMenu.push({id: "separator_1", type: "separator"});
            response.data.forEach((item: IWorkspace) => {
                workspaceSubMenu.push({
                    iconHTML: "",
                    action: "iconCloseRound",
                    current: window.scribli.config.system.workspaceDir === item.path,
                    label: pathPosix().basename(item.path),
                    bind(menuElement) {
                        menuElement.addEventListener("click", (event) => {
                            if (hasClosestByClassName(event.target as Element, "b3-menu__action")) {
                                event.preventDefault();
                                event.stopPropagation();
                                fetchPost("/api/system/removeWorkspaceDir", {path: item.path}, () => {
                                    confirmDialog(window.scribli.languages.deleteOpConfirm, window.scribli.languages.removeWorkspacePhysically.replace("${x}", item.path), () => {
                                        fetchPost("/api/system/removeWorkspaceDirPhysically", {path: item.path});
                                    }, undefined, true);
                                });
                                return;
                            }
                            confirmDialog(window.scribli.languages.confirm, `${pathPosix().basename(window.scribli.config.system.workspaceDir)} -> ${pathPosix().basename(item.path)}?`, () => {
                                fetchPost("/api/system/setWorkspaceDir", {
                                    path: item.path
                                }, () => {
                                    exitScribli(false);
                                });
                            });
                        });
                    }
                });
            });
            /// #endif
            if (!isBrowser()) {
                window.scribli.menus.menu.append(new MenuItem({
                    id: "workspaceList",
                    label: window.scribli.languages.workspaceList,
                    icon: "iconWorkspace",
                    type: "submenu",
                    submenu: workspaceSubMenu,
                }).element);
            }
        }
        const layoutSubMenu: IMenu[] = [{
            id: "save",
            iconHTML: "",
            label: window.scribli.languages.save,
            click() {
                editLayout();
            }
        }];
        if (window.scribli.storage[Constants.LOCAL_LAYOUTS].length > 0) {
            layoutSubMenu.push({id: "separator_1", type: "separator"});
            layoutSubMenu.push({
                iconHTML: "",
                type: "empty",
                label: `<input class="b3-text-field fn__block" style="margin: 4px 0" placeholder="${window.scribli.languages.search}">
<div class="b3-list b3-list--background" style="width: 220px"></div>`,
                bind(menuElement) {
                    const genListHTML = (isInit = false) => {
                        let html = "";
                        window.scribli.storage[Constants.LOCAL_LAYOUTS].sort((a: ISaveLayout, b: ISaveLayout) => {
                            return a.name.localeCompare(b.name, undefined, {numeric: true});
                        }).forEach((item: ISaveLayout) => {
                            if (inputElement.value === "" || item.name.toLowerCase().indexOf(inputElement.value.toLowerCase()) > -1) {
                                html += `<div data-name="${item.name}" class="b3-list-item b3-list-item--narrow b3-list-item--hide-action${!isInit && !html ? " b3-list-item--focus" : ""} ariaLabel" data-position="8east" aria-label="${escapeAttr(item.name)}" >
    <div class="b3-list-item__text">${item.name}</div>
    <span class="b3-list-item__meta">${item.time ? dayjs(item.time).format("YYYY-MM-DD HH:mm") : ""}</span>
    <span class="b3-list-item__action">
        <svg><use xlink:href="#iconEdit"></use></svg>
    </span>
</div>`;
                            }
                        });
                        return html;
                    };
                    const inputElement = menuElement.querySelector(".b3-text-field") as HTMLInputElement;
                    const listElement = menuElement.querySelector(".b3-list");
                    inputElement.addEventListener("focus", () => {
                        if (!menuElement.querySelector(".b3-list-item--focus")) {
                            menuElement.querySelector(".b3-list-item")?.classList.add("b3-list-item--focus");
                        }
                    });
                    inputElement.addEventListener("blur", () => {
                        menuElement.querySelector(".b3-list-item--focus")?.classList.remove("b3-list-item--focus");
                    });
                    inputElement.addEventListener("keydown", (event) => {
                        event.stopPropagation();
                        if (event.isComposing) {
                            return;
                        }
                        upDownHint(listElement, event);
                        if (event.key === "Escape" || (event.key === "ArrowLeft" && inputElement.value === "")) {
                            window.scribli.menus.menu.remove(true);
                        } else if (event.key === "Enter") {
                            const currentElement = listElement.querySelector(".b3-list-item--focus");
                            if (currentElement) {
                                listElement.dispatchEvent(new CustomEvent("click", {detail: currentElement.getAttribute("data-name")}));
                            }
                        }
                    });
                    inputElement.addEventListener("compositionend", () => {
                        listElement.innerHTML = genListHTML();
                    });
                    inputElement.addEventListener("input", (event: InputEvent) => {
                        if (event.isComposing) {
                            return;
                        }
                        event.stopPropagation();
                        listElement.innerHTML = genListHTML();
                    });
                    listElement.addEventListener("click", (event: MouseEvent) => {
                        if (window.scribli.config.readonly) {
                            return;
                        }
                        const actionElement = hasClosestByClassName(event.target as Element, "b3-list-item__action");
                        if (actionElement) {
                            event.preventDefault();
                            event.stopPropagation();
                            editLayout(actionElement.parentElement.dataset.name);
                            window.scribli.menus.menu.remove();
                            return;
                        }
                        const liElement = hasClosestByClassName(event.target as Element, "b3-list-item");
                        if (liElement || event.detail) {
                            const itemData: ISaveLayout = window.scribli.storage[Constants.LOCAL_LAYOUTS].find((item: ISaveLayout) => {
                                if (typeof event.detail === "string") {
                                    return item.name === event.detail;
                                } else if (liElement) {
                                    return item.name === liElement.dataset.name;
                                }
                            });
                            if (itemData) {
                                fetchPost("/api/system/setUILayout", {layout: itemData.layout}, () => {
                                    if (itemData.filesPaths) {
                                        window.scribli.storage[Constants.LOCAL_FILESPATHS] = itemData.filesPaths;
                                        setStorageVal(Constants.LOCAL_FILESPATHS, itemData.filesPaths, () => {
                                            window.location.reload();
                                        });
                                    } else {
                                        window.location.reload();
                                    }
                                });
                            }
                            event.preventDefault();
                            event.stopPropagation();
                        }
                    });
                    listElement.innerHTML = genListHTML(true);
                }
            });
        }
        if (!window.scribli.config.readonly) {
            window.scribli.menus.menu.append(new MenuItem({
                id: "layout",
                label: window.scribli.languages.layout,
                icon: "iconLayout",
                type: "submenu",
                submenu: layoutSubMenu
            }).element);
        }
        window.scribli.menus.menu.append(new MenuItem({id: "separator_1", type: "separator"}).element);
        if (!window.scribli.config.readonly) {
            if (getOpenNotebookCount() < 2) {
                window.scribli.menus.menu.append(new MenuItem({
                    id: "dailyNote",
                    label: window.scribli.languages.dailyNote,
                    icon: "iconCalendar",
                    accelerator: window.scribli.config.keymap.general.dailyNote.custom,
                    click: () => {
                        newDailyNote(app);
                    }
                }).element);
            } else {
                const submenu: IMenu[] = [];
                window.scribli.notebooks.forEach(item => {
                    if (!item.closed) {
                        submenu.push({
                            label: escapeHtml(item.name),
                            iconHTML: unicode2Emoji(item.icon || window.scribli.storage[Constants.LOCAL_IMAGES].note, "b3-menu__icon", true),
                            accelerator: window.scribli.storage[Constants.LOCAL_DAILYNOTEID] === item.id ? window.scribli.config.keymap.general.dailyNote.custom : "",
                            click: () => {
                                fetchNewDailyNote(app, item.id);
                                window.scribli.storage[Constants.LOCAL_DAILYNOTEID] = item.id;
                                setStorageVal(Constants.LOCAL_DAILYNOTEID, window.scribli.storage[Constants.LOCAL_DAILYNOTEID]);
                            }
                        });
                    }
                });
                window.scribli.menus.menu.append(new MenuItem({
                    id: "dailyNote",
                    label: window.scribli.languages.dailyNote,
                    icon: "iconCalendar",
                    type: "submenu",
                    submenu
                }).element);
            }
            if (!window.scribli.config.readonly) {
                window.scribli.menus.menu.append(new MenuItem({
                    id: "riffCard",
                    label: window.scribli.languages.riffCard,
                    type: "submenu",
                    icon: "iconRiffCard",
                    submenu: [{
                        id: "spaceRepetition",
                        iconHTML: "",
                        label: window.scribli.languages.spaceRepetition,
                        accelerator: window.scribli.config.keymap.general.riffCard.custom,
                        click: () => {
                            openCard(app);
                        }
                    }, {
                        id: "manage",
                        iconHTML: "",
                        label: window.scribli.languages.manage,
                        click: () => {
                            viewCards(app, "", window.scribli.languages.all, "");
                        }
                    }],
                }).element);
            }
            window.scribli.menus.menu.append(new MenuItem({
                id: "recentDocs",
                label: window.scribli.languages.recentDocs,
                icon: "iconRecentDocs",
                accelerator: window.scribli.config.keymap.general.recentDocs.custom,
                click: () => {
                    openRecentDocs();
                }
            }).element);
            window.scribli.menus.menu.append(new MenuItem({
                id: "lockScreen",
                label: window.scribli.languages.lockScreen,
                icon: "iconLock",
                accelerator: window.scribli.config.keymap.general.lockScreen.custom,
                click: () => {
                    lockScreen(app);
                }
            }).element);
            window.scribli.menus.menu.append(new MenuItem({
                id: "dataHistory",
                label: window.scribli.languages.dataHistory,
                icon: "iconHistory",
                accelerator: window.scribli.config.keymap.general.dataHistory.custom,
                click: () => {
                    openHistory(app);
                }
            }).element);
            if (!window.scribli.config.readonly) {
                window.scribli.menus.menu.append(new MenuItem({
                    id: "dataMigration",
                    label: window.scribli.languages.dataMigration,
                    icon: "iconDatabaseBackup",
                    click: () => {
                        openDataMigration();
                    }
                }).element);
            }
            window.scribli.menus.menu.append(new MenuItem({id: "separator_2", type: "separator"}).element);
        }
        window.scribli.menus.menu.append(new MenuItem({
            id: "userGuide",
            label: window.scribli.languages.userGuide,
            icon: "iconHelp",
            ignore: window.scribli.config.readonly,
            click: () => {
                mountHelp();
            }
        }).element);
        /// #if !BROWSER
        window.scribli.menus.menu.append(new MenuItem({
            id: "debug",
            label: window.scribli.languages.debug,
            icon: "iconBug",
            click: () => {
                ipcRenderer.send(Constants.SCRIBLI_CMD, "openDevTools");
            }
        }).element);
        /// #endif
        if (isIPad() || !isBrowser()) {
            window.scribli.menus.menu.append(new MenuItem({id: "separator_3", type: "separator"}).element);
            window.scribli.menus.menu.append(new MenuItem({
                id: "safeQuit",
                label: window.scribli.languages.safeQuit,
                icon: "iconQuit",
                warning: true,
                click: () => {
                    exportLayout({
                        errorExit: true,
                        cb: exitScribli,
                    });
                }
            }).element);
        }
        window.scribli.menus.menu.popup({x: rect.left, y: rect.bottom});
    });
};

const openWorkspace = (workspace: string) => {
    /// #if !BROWSER
    if (workspace === window.scribli.config.system.workspaceDir) {
        return;
    }
    fetchPost("/api/system/setWorkspaceDir", {
        path: workspace
    }, () => {
        ipcRenderer.send(Constants.SCRIBLI_OPEN_WORKSPACE, {
            workspace,
            lang: window.scribli.config.appearance.lang
        });
    });
    /// #endif
};

const workspaceItem = (item: IWorkspace) => {
    /// #if !BROWSER
    const submenu = [{
        id: "showInFolder",
        icon: "iconFolder",
        label: window.scribli.languages.showInFolder,
        click() {
            useShell("showItemInFolder", item.path);
        }
    }, {
        id: "copyPath",
        icon: "iconCopy",
        label: window.scribli.languages.copyPath,
        click() {
            writeText(item.path);
            showMessage(window.scribli.languages.copied);
        }
    }];
    if (item.path !== window.scribli.config.system.workspaceDir) {
        submenu.splice(0, 0, {
            id: "openBy",
            icon: "iconOpenWindow",
            label: window.scribli.languages.openBy,
            click() {
                openWorkspace(item.path);
            }
        });
        if (item.closed) {
            submenu.push({
                id: "removeWorkspaceTip",
                icon: "iconTrashcan",
                label: window.scribli.languages.removeWorkspaceTip,
                click() {
                    fetchPost("/api/system/removeWorkspaceDir", {path: item.path});
                }
            });
        }
    }
    return {
        label: `<div aria-label="${item.path}" class="fn__ellipsis ariaLabel" style="max-width: 256px">
    ${originalPath().basename(item.path)}
</div>`,
        current: !item.closed,
        iconHTML: "",
        type: "submenu",
        submenu,
        click() {
            openWorkspace(item.path);
        },
    };
    /// #endif
};

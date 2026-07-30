import {Constants} from "../constants";
import {fetchPost} from "../util/fetch";
import {exportLayout} from "../layout/util";
import {getDockByType} from "../layout/tabUtil";
import {Files} from "../layout/dock/Files";
import {getAllEditor} from "../layout/getAll";
/// #if !BROWSER
import {ipcRenderer} from "electron";
/// #endif
import {hideMessage, showMessage} from "./message";
import {Dialog} from "./index";
import {isMobile} from "../util/functions";
import {confirmDialog} from "./confirmDialog";
import {escapeHtml} from "../util/escape";
import {hideAllElements} from "../protyle/ui/hideElements";
import {App} from "../index";
import {saveScroll} from "../protyle/scroll/saveScroll";
import {setStorageVal} from "../protyle/util/compatibility";
import {Plugin} from "../plugin";

export const setRefDynamicText = (data: {
    "blockID": string,
    "defBlockID": string,
    "refText": string,
    "rootID": string
}) => {
    getAllEditor().forEach(editor => {
        editor.protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${data.blockID}"] span[data-type~="block-ref"][data-subtype="d"][data-id="${data.defBlockID}"]`).forEach(item => {
            item.innerHTML = data.refText;
        });
    });
};

export const setDefRefCount = (data: {
    "blockID": string,
    "refCount": number,
    "rootRefCount": number,
    "rootID": string
}) => {
    getAllEditor().forEach(editor => {
        if (editor.protyle.block.rootID === data.rootID && editor.protyle.title) {
            const attrElement = editor.protyle.title.element.querySelector(".protyle-attr");
            const countElement = attrElement.querySelector(".protyle-attr--refcount");
            if (countElement) {
                if (data.rootRefCount === 0) {
                    countElement.remove();
                } else {
                    countElement.textContent = data.rootRefCount.toString();
                }
            } else if (data.rootRefCount > 0) {
                attrElement.insertAdjacentHTML("beforeend", `<div class="protyle-attr--refcount popover__block">${data.rootRefCount}</div>`);
            }
        }
        if (data.rootID === data.blockID) {
            return;
        }
        editor.protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${data.blockID}"]`).forEach(item => {
            const countElement = item.lastElementChild?.querySelector(".protyle-attr--refcount");
            if (countElement) {
                if (data.refCount === 0) {
                    countElement.remove();
                } else {
                    countElement.textContent = data.refCount.toString();
                }
            } else if (data.refCount > 0) {
                const attrElement = item.lastElementChild;
                if (attrElement.childElementCount > 0) {
                    attrElement.lastElementChild.insertAdjacentHTML("afterend", `<div class="protyle-attr--refcount popover__block">${data.refCount}</div>`);
                } else {
                    attrElement.innerHTML = `<div class="protyle-attr--refcount popover__block">${data.refCount}</div>${Constants.ZWSP}`;
                }
            }
            if (data.refCount === 0) {
                item.removeAttribute("refcount");
            } else {
                item.setAttribute("refcount", data.refCount.toString());
            }
        });
    });

    const liElement = (getDockByType("file")?.data["file"] as Files)?.element.querySelector(`li[data-node-id="${data.rootID}"]`);
    if (liElement) {
        const counterElement = liElement.querySelector(".counter");
        if (counterElement) {
            if (data.rootRefCount === 0) {
                counterElement.remove();
            } else {
                counterElement.textContent = data.rootRefCount.toString();
            }
        } else if (data.rootRefCount > 0) {
            liElement.insertAdjacentHTML("beforeend", `<span class="popover__block counter b3-tooltips b3-tooltips__nw" aria-label="${window.scribli.languages.ref}">${data.rootRefCount}</span>`);
        }
    }
};

export const lockScreen = async (app: App) => {
    if (window.scribli.config.readonly || window.scribli.isPublish) {
        return;
    }
    app.plugins.forEach(item => {
        item.eventBus.emit("lock-screen");
    });
    exportLayout({
        errorExit: false,
        cb() {
            fetchPost("/api/system/logoutAuth");
        }
    });

};

// forceQuit is used when the kernel is disconnected and /api/system/exit cannot run.
export const forceQuit = () => {
    /// #if !BROWSER
    ipcRenderer.send(Constants.SCRIBLI_QUIT, location.port);
    /// #else
    window.close();
    /// #endif
};

const installNewVersion = (installPkgPath: string, setCurrentWorkspace: boolean) => {
    if (!installPkgPath) {
        showMessage(window.scribli.languages._kernel[104], 7000, "error");
        return;
    }
    /// #if !BROWSER
    ipcRenderer.invoke(Constants.SCRIBLI_INSTALL_UPDATE, {
        port: location.port,
        setCurrentWorkspace,
    }).then((accepted: boolean) => {
        if (!accepted) {
            showMessage(window.scribli.languages._kernel[104], 7000, "error");
        }
    }).catch(() => {
        showMessage(window.scribli.languages._kernel[104], 7000, "error");
    });
    /// #else
    fetchPost("/api/system/exit", {
        force: true,
        setCurrentWorkspace,
        execInstallPkg: 1,
    });
    /// #endif
};

export const exitScribli = async (setCurrentWorkspace = true) => {
    hideAllElements(["util"]);
    fetchPost("/api/system/exit", {force: false, setCurrentWorkspace}, (response) => {
        if (response.code === 1) {
            const msgId = showMessage(response.msg, response.data.closeTimeout, "error");
            const buttonElement = document.querySelector(`#message [data-id="${msgId}"] button`);
            if (buttonElement) {
                buttonElement.addEventListener("click", () => {
                    if (response.data.installPkgPath) {
                        installNewVersion(response.data.installPkgPath, setCurrentWorkspace);
                        return;
                    }
                    fetchPost("/api/system/exit", {force: true, setCurrentWorkspace}, () => {
                        /// #if !BROWSER
                        ipcRenderer.send(Constants.SCRIBLI_QUIT, location.port);
                        /// #endif
                    });
                });
            }
        } else if (response.code === 2) {
            hideMessage();

            /// #if !BROWSER
            if ("std" === window.scribli.config.system.container) {
                ipcRenderer.send(Constants.SCRIBLI_SHOW_WINDOW);
            }
            /// #endif

            confirmDialog(window.scribli.languages.updateVersion, response.msg, () => {
                installNewVersion(response.data.installPkgPath, setCurrentWorkspace);
            }, () => {
                fetchPost("/api/system/exit", {
                    force: true,
                    setCurrentWorkspace,
                    execInstallPkg: 1
                }, () => {
                    /// #if !BROWSER
                    ipcRenderer.send(Constants.SCRIBLI_QUIT, location.port);
                    /// #endif
                });
            });
        } else {
            /// #if !BROWSER
            ipcRenderer.send(Constants.SCRIBLI_QUIT, location.port);
            /// #endif
        }
    });
};

export const transactionError = (msg?: string) => {
    if (document.getElementById("transactionError")) {
        return;
    }
    const dialog = new Dialog({
        disableClose: true,
        title: `${window.scribli.languages.stateExcepted} v${Constants.SCRIBLI_VERSION}`,
        content: `<div class="b3-dialog__content" style="max-height: calc(100vh - 182px)" id="transactionError">
    ${window.scribli.languages.rebuildIndexTip}
    ${msg ? `<div class="fn__hr"></div>${escapeHtml(msg.trim())}` : ""}
</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--text">${window.scribli.languages._kernel[97]}</button>
    <div class="fn__space"></div>
    <button class="b3-button">${window.scribli.languages.rebuildDataIndex}</button>
</div>`,
        width: isMobile() ? "92vw" : "520px",
    });
    dialog.element.setAttribute("data-key", Constants.DIALOG_STATEEXCEPTED);
    const btnsElement = dialog.element.querySelectorAll(".b3-button");
    btnsElement[0].addEventListener("click", () => {
        exportLayout({
            errorExit: true,
            cb: exitScribli
        });
    });
    btnsElement[1].addEventListener("click", () => {
        refreshFileTree();
        dialog.destroy();
    });
};

export const refreshFileTree = (cb?: () => void) => {
    window.scribli.storage[Constants.LOCAL_FILEPOSITION] = {};
    setStorageVal(Constants.LOCAL_FILEPOSITION, window.scribli.storage[Constants.LOCAL_FILEPOSITION]);
    fetchPost("/api/system/rebuildDataIndex", {}, () => {
        if (cb) {
            cb();
        }
    });
};

let statusTimeout: number;
export const progressStatus = (data: IWebSocketData) => {
    const msgElement = document.querySelector("#status .status__msg");
    if (msgElement) {
        clearTimeout(statusTimeout);
        msgElement.innerHTML = data.msg;
        statusTimeout = window.setTimeout(() => {
            msgElement.innerHTML = "";
        }, 12000);
    }
};

export const progressLoading = (data: IWebSocketData) => {
    let progressElement = document.getElementById("progress");
    if (!progressElement) {
        document.body.insertAdjacentHTML("beforeend", `<div id="progress" style="z-index: ${++window.scribli.zIndex}"></div>`);
        progressElement = document.getElementById("progress");
    }
    if (data.code === 2) {
        progressElement.remove();
        return;
    }
    if (data.code === 0) {
        progressElement.innerHTML = `<div class="b3-dialog__scrim" style="opacity: 1"></div>
<div class="b3-dialog__loading">
    <div style="text-align: right">${data.data.current}/${data.data.total}</div>
    <div style="margin: 8px 0;height: 8px;border-radius: var(--b3-border-radius);overflow: hidden;background-color:#fff;"><div style="width: ${data.data.current / data.data.total * 100}%;transition: var(--b3-transition);background-color: var(--b3-theme-primary);height: 8px;"></div></div>
    <div class="ft__breakword">${escapeHtml(data.msg)}</div>
</div>`;
    } else if (data.code === 1) {
        if (progressElement.lastElementChild) {
            progressElement.lastElementChild.lastElementChild.innerHTML = escapeHtml(data.msg);
        } else {
            progressElement.innerHTML = `<div class="b3-dialog__scrim" style="opacity: 1"></div>
<div class="b3-dialog__loading">
    <div style="margin: 8px 0;height: 8px;border-radius: var(--b3-border-radius);overflow: hidden;background-color:#fff;"><div style="background-color: var(--b3-theme-primary);height: 8px;background-image: linear-gradient(-45deg, rgba(255, 255, 255, 0.2) 25%, transparent 25%, transparent 50%, rgba(255, 255, 255, 0.2) 50%, rgba(255, 255, 255, 0.2) 75%, transparent 75%, transparent);animation: stripMove 450ms linear infinite;background-size: 50px 50px;"></div></div>
    <div class="ft__breakword">${escapeHtml(data.msg)}</div>
</div>`;
        }
    }
};

export const progressBackgroundTask = (tasks: { action: string }[]) => {
    const backgroundTaskElement = document.querySelector(".status__backgroundtask");
    if (!backgroundTaskElement) {
        return;
    }
    if (tasks.length === 0) {
        backgroundTaskElement.classList.add("fn__none");
        if (!window.scribli.menus.menu.element.classList.contains("fn__none") &&
            window.scribli.menus.menu.element.getAttribute("data-name") === Constants.MENU_STATUS_BACKGROUND_TASK) {
            window.scribli.menus.menu.remove();
        }
    } else {
        backgroundTaskElement.classList.remove("fn__none");
        backgroundTaskElement.setAttribute("data-tasks", JSON.stringify(tasks));
        backgroundTaskElement.innerHTML = tasks[0].action + '<div class="fn__progress"><div></div></div>';
    }
};

export const bootSync = () => {
    fetchPost("/api/sync/getBootSync", {}, response => {
        if (response.code === 1) {
            const dialog = new Dialog({
                width: isMobile() ? "92vw" : "50vw",
                title: "🌩️ " + window.scribli.languages.bootSyncFailed,
                content: `<div class="b3-dialog__content">${response.msg}</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages.syncNow}</button>
</div>`
            });
            dialog.element.setAttribute("data-key", Constants.DIALOG_BOOTSYNCFAILED);
            const btnsElement = dialog.element.querySelectorAll(".b3-button");
            btnsElement[0].addEventListener("click", () => {
                dialog.destroy();
            });
            btnsElement[1].addEventListener("click", () => {
                if (btnsElement[1].getAttribute("disabled")) {
                    return;
                }
                btnsElement[1].setAttribute("disabled", "disabled");
                fetchPost("/api/sync/performBootSync", {}, (syncResponse) => {
                    if (syncResponse.code === 0) {
                        dialog.destroy();
                    }
                    btnsElement[1].removeAttribute("disabled");
                });
            });
        }
    });
};

export const processSync = (data?: IWebSocketData, plugins?: Plugin[]) => {
    if (data?.code === 1) {
        window.dispatchEvent(new CustomEvent("scribli-sync-success"));
    }
    const iconElement = document.querySelector("#barSync");
    if (!iconElement) {
        return;
    }
    const useElement = iconElement.querySelector("use");
    if (!data) {
        iconElement.classList.remove("toolbar__item--active");
        if (!window.scribli.config.sync.enabled || window.scribli.config.sync.provider === 0) {
            useElement.setAttribute("xlink:href", "#iconCloudOff");
        } else {
            useElement.setAttribute("xlink:href", "#iconCloudSucc");
        }
        return;
    }
    iconElement.firstElementChild.classList.remove("fn__rotate");
    if (data.code === 0) {  // syncing
        iconElement.classList.add("toolbar__item--active");
        iconElement.firstElementChild.classList.add("fn__rotate");
        useElement.setAttribute("xlink:href", "#iconRefresh");
    } else if (data.code === 2) {    // error
        iconElement.classList.remove("toolbar__item--active");
        useElement.setAttribute("xlink:href", "#iconCloudError");
    } else if (data.code === 1) {   // success
        iconElement.classList.remove("toolbar__item--active");
        useElement.setAttribute("xlink:href", "#iconCloudSucc");
    }
    plugins.forEach((item) => {
        if (data.code === 0) {
            item.eventBus.emit("sync-start", data);
        } else if (data.code === 1) {
            item.eventBus.emit("sync-end", data);
        } else if (data.code === 2) {
            item.eventBus.emit("sync-fail", data);
        }
    });
};

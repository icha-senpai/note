import * as compatibility from "../protyle/util/compatibility";
/// #if !BROWSER
import {ipcRenderer} from "electron";
import {Constants} from "../constants";
/// #endif
export const readText = compatibility.readText;
export const writeText = compatibility.writeText;
export const copyPlainText = compatibility.copyPlainText;
export const getEventName = compatibility.getEventName;
export const isOnlyMeta = compatibility.isOnlyMeta;
export const isNotCtrl = compatibility.isNotCtrl;
export const isHuawei = compatibility.isHuawei;
export const isIPhone = compatibility.isIPhone;
export const isIPad = compatibility.isIPad;
export const isMac = compatibility.isMac;
export const updateHotkeyTip = compatibility.updateHotkeyTip;
export const getLocalStorage = compatibility.getLocalStorage;
export const setStorageVal = compatibility.setStorageVal;

export const getStorageVal = (key: string): any => {
    return window.scribli.storage?.[key] ?? null;
};

/**
 */
export const sendNotification = (options: {
    channel?: string,
    title?: string,
    body?: string,
    delayInSeconds?: number,
    timeoutType?: "default" | "never"
}): Promise<number> => {
    return new Promise((resolve) => {
        const title = options.title || "";
        const body = options.body || "";
        const delayInSeconds = options.delayInSeconds || 0;
        if (!title.trim() && !body.trim()) {
            resolve(-1);
            return;
        }

        /// #if BROWSER
        resolve(-1);
        /// #else
        const timeoutId = window.setTimeout(() => {
            ipcRenderer.send(Constants.SCRIBLI_CMD, {
                cmd: "notification",
                title,
                body,
                timeoutType: options.timeoutType || "default"
            });
        }, delayInSeconds * 1000);
        resolve(timeoutId);
        /// #endif
    });
};

export const cancelNotification = (id: number) => {
    if (id < 0) {
        return;
    }
    /// #if BROWSER
    return;
    /// else
    clearTimeout(id);
    /// #endif
};

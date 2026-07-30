/// #if !BROWSER
import {ipcRenderer} from "electron";
/// #endif
import {fetchPost} from "../../util/fetch";
import {Constants} from "../../constants";
import {exportLayout} from "../../layout/util";
import {exitScribli} from "../../dialog/processSystem";

export const sendAppSetting = (controlId: string, value: unknown) => {
    switch (controlId) {
        case "system.autoLaunch2": {
            const autoLaunchMode = value as Config.ISystem["autoLaunch2"];
            fetchPost("/api/system/setAutoLaunch", {autoLaunch: autoLaunchMode}, () => {
                window.scribli.config.system.autoLaunch2 = autoLaunchMode;
                /// #if !BROWSER
                ipcRenderer.send(Constants.SCRIBLI_AUTO_LAUNCH, {
                    openAtLogin: 0 !== autoLaunchMode,
                    openAsHidden: 2 === autoLaunchMode,
                });
                /// #endif
            });
            break;
        }
        case "system.lockScreenMode": {
            const lockScreenMode = (value ? 1 : 0) as Config.ISystem["lockScreenMode"];
            fetchPost("/api/system/setFollowSystemLockScreen", {lockScreenMode}, () => {
                window.scribli.config.system.lockScreenMode = lockScreenMode;
            });
            break;
        }
        case "system.networkServe": {
            const networkServe = Boolean(value) as Config.ISystem["networkServe"];
            fetchPost("/api/system/setNetworkServe", {networkServe}, () => {
                void exportLayout({
                    errorExit: true,
                    cb: exitScribli,
                });
            });
            break;
        }
        case "system.networkServeTLS": {
            const networkServeTLS = Boolean(value) as Config.ISystem["networkServeTLS"];
            fetchPost("/api/system/setNetworkServeTLS", {networkServeTLS}, () => {
                void exportLayout({
                    errorExit: true,
                    cb: exitScribli,
                });
            });
            break;
        }
        default:
            console.warn(`[config] sendAppSetting: unhandled controlId "${controlId}"`);
            break;
    }
};

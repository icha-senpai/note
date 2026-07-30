/// #if !BROWSER
import {ipcRenderer} from "electron";
import * as path from "path";
/// #endif
import type {SettingTabBuilder} from "../setting/builder";
import {Constants} from "../../constants";
import {fetchPost} from "../../util/fetch";
import {exportLayout} from "../../layout/util";
import {exitScribli} from "../../dialog/processSystem";
import {showMessage} from "../../dialog/message";
import {isMac, saveExportFile} from "../../protyle/util/compatibility";
/// #if !BROWSER
import {afterExport} from "../../protyle/export/util";
/// #endif
import {genConfigItemMainHtml, genConfigItemName} from "../render/fragments";
import {sendAppSetting} from "./appRuntime";

const genImportUploadButtonHtml = (inputId: string, label: string): string =>
    `<button class="b3-button b3-button--outline fn__flex-center fn__size200" style="position: relative">
    <input id="${inputId}" class="b3-form__upload" type="file">
    <svg><use xlink:href="#iconDownload"></use></svg>
    ${label}
</button>`;

const registerAppGeneralGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("general", window.scribli.languages.configGroupGeneral);

    /// #if !BROWSER
    if (!window.scribli.config.system.isMicrosoftStore && window.scribli.config.system.container === "std" && window.scribli.config.system.os !== "linux") {
        group.select("system.autoLaunch2", {
            title: window.scribli.languages.autoLaunch,
            desc: window.scribli.languages.autoLaunchTip,
            options: [
                {value: 0, label: window.scribli.languages.autoLaunchMode0},
                {value: 1, label: window.scribli.languages.autoLaunchMode1},
                ...(!isMac() ? [{value: 2, label: window.scribli.languages.autoLaunchMode2}] : []),
            ],
            save: (value) => sendAppSetting("system.autoLaunch2", value),
        });
    }
    /// #endif
    group.slot({
        key: "networkProxy",
        keywords: [
            window.scribli.languages.networkProxy,
            window.scribli.languages.about17,
            window.scribli.languages.directConnection,
            "SOCKS5",
            "HTTPS",
            "HTTP",
            "user:pass@IP",
            "Port",
            window.scribli.languages.confirm,
        ],
        html: genNetworkProxyHtml,
        afterMount: mountNetworkProxy,
    });
};

const genNetworkProxyHtml = (): string => {
    const proxy = window.scribli.config.system.networkProxy;
    return `<div class="b3-label config-item">
    ${genConfigItemName(window.scribli.languages.networkProxy)}
    <div class="b3-label__text">
        ${window.scribli.languages.about17}
    </div>
    <div class="b3-label__text fn__flex config-wrap" style="overflow: visible !important;">
        <select id="networkProxyScheme" class="b3-select">
            <option value="" ${proxy.scheme === "" ? "selected" : ""}>${window.scribli.languages.directConnection}</option>
            <option value="socks5" ${proxy.scheme === "socks5" ? "selected" : ""}>SOCKS5</option>
            <option value="https" ${proxy.scheme === "https" ? "selected" : ""}>HTTPS</option>
            <option value="http" ${proxy.scheme === "http" ? "selected" : ""}>HTTP</option>
        </select>
        <span class="fn__space"></span>
        <input id="networkProxyHost" placeholder="user:pass@IP" class="b3-text-field fn__block" value="${Lute.EscapeHTMLStr(proxy.host)}"/>
        <span class="fn__space"></span>
        <input id="networkProxyPort" placeholder="Port" class="b3-text-field fn__block" value="${Lute.EscapeHTMLStr(proxy.port)}" type="number"/>
        <span class="fn__space"></span>
        <button id="networkProxyConfirm" class="b3-button fn__size200 b3-button--outline">${window.scribli.languages.confirm}</button>
    </div>
</div>`;
};

const mountNetworkProxy = (root: HTMLElement) => {
    root.querySelector("#networkProxyConfirm")?.addEventListener("click", () => {
        const scheme = (root.querySelector("#networkProxyScheme") as HTMLSelectElement)?.value as Config.TSystemNetworkProxyScheme;
        const host = (root.querySelector("#networkProxyHost") as HTMLInputElement)?.value;
        const port = (root.querySelector("#networkProxyPort") as HTMLInputElement)?.value;
        fetchPost("/api/system/setNetworkProxy", {scheme, host, port}, async () => {
            Object.assign(window.scribli.config.system.networkProxy, {scheme, host, port});
            /// #if !BROWSER
            ipcRenderer.invoke(Constants.SCRIBLI_GET, {
                cmd: "setProxy",
                proxyURL: `${window.scribli.config.system.networkProxy.scheme}://${window.scribli.config.system.networkProxy.host}:${window.scribli.config.system.networkProxy.port}`,
            }).then(() => {
                exportLayout({
                    errorExit: false,
                    cb() {
                        window.location.reload();
                    },
                });
            });
            /// #endif
        });
    });
};

const registerAppDataGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("data", window.scribli.languages.configGroupData);

    group.button({
        id: "exportData",
        title: `${window.scribli.languages.export} Data`,
        desc: window.scribli.languages.exportDataTip,
        label: window.scribli.languages.export,
        icon: "iconUpload",
        afterMount: mountExportData,
    });
    group.slot({
        key: "importData",
        keywords: [window.scribli.languages.import, window.scribli.languages.importDataTip],
        html: () => `<div class="fn__flex b3-label config-item config-wrap">
    ${genConfigItemMainHtml(`${window.scribli.languages.import} Data`, window.scribli.languages.importDataTip)}
    <span class="fn__space"></span>
    ${genImportUploadButtonHtml("importData", window.scribli.languages.import)}
</div>`,
        afterMount: (root) => {
            root.querySelector("#importData")?.addEventListener("change", (event: Event) => {
                const target = event.target as HTMLInputElement;
                const formData = new FormData();
                formData.append("file", target.files[0]);
                fetchPost("/api/import/importData", formData);
            });
        },
    });
    group.button({
        id: "exportConf",
        title: window.scribli.languages.exportConf,
        desc: window.scribli.languages.exportConfTip,
        label: window.scribli.languages.export,
        icon: "iconUpload",
        afterMount: (root) => {
            root.querySelector("#exportConf")?.addEventListener("click", () => {
                fetchPost("/api/system/exportConf", {}, (response) => {
                    void saveExportFile(response.data.zip);
                });
            });
        },
    });
    group.slot({
        key: "importConf",
        keywords: [window.scribli.languages.importConf, window.scribli.languages.importConfTip],
        html: () => `<div class="fn__flex b3-label config-item config-wrap">
    ${genConfigItemMainHtml(window.scribli.languages.importConf, window.scribli.languages.importConfTip)}
    <span class="fn__space"></span>
    ${genImportUploadButtonHtml("importConf", window.scribli.languages.import)}
</div>`,
        afterMount: (root) => {
            root.querySelector("#importConf")?.addEventListener("change", (event: Event) => {
                const target = event.target as HTMLInputElement;
                const formData = new FormData();
                formData.append("file", target.files[0]);
                fetchPost("/api/system/importConf", formData, (response) => {
                    if (response.code !== 0) {
                        showMessage(response.msg);
                        return;
                    }
                    showMessage(window.scribli.languages.imported);
                    void exportLayout({
                        errorExit: true,
                        cb: exitScribli,
                    });
                });
            });
        },
    });
};

const mountExportData = (root: HTMLElement) => {
    root.querySelector("#exportData")?.addEventListener("click", async () => {
        /// #if BROWSER
        fetchPost("/api/export/exportData", {}, (response) => {
            saveExportFile(response.data.zip);
        });
        /// #else
        const result = await ipcRenderer.invoke(Constants.SCRIBLI_GET, {
            cmd: "showOpenDialog",
            title: window.scribli.languages.export + " " + "Data",
            properties: ["createDirectory", "openDirectory"],
        });
        if (result.canceled || result.filePaths.length === 0) {
            return;
        }
        const msgId = showMessage(window.scribli.languages.exporting, -1);
        fetchPost("/api/export/exportDataInFolder", {
            folder: result.filePaths[0],
        }, (response) => {
            afterExport(path.join(result.filePaths[0], response.data.name), msgId);
        });
        /// #endif
    });
};

const registerAppMaintenanceGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("maintenance", window.scribli.languages.configGroupMaintenance);

    group.button({
        id: "reloadUI",
        title: window.scribli.languages.reloadUI,
        desc: window.scribli.languages.reloadUITip,
        label: window.scribli.languages.reloadUI,
        icon: "iconRefresh",
        afterMount: (root) => {
            root.querySelector("#reloadUI")?.addEventListener("click", () => {
                fetchPost("/api/ui/reloadUI", {});
            });
        },
    });
    group.button({
        id: "vacuumDataIndex",
        title: window.scribli.languages.vacuumDataIndex,
        desc: window.scribli.languages.vacuumDataIndexTip,
        label: window.scribli.languages.vacuumDataIndex,
        icon: "iconRefresh",
        afterMount: (root) => {
            root.querySelector("#vacuumDataIndex")?.addEventListener("click", () => {
                fetchPost("/api/system/vacuumDataIndex", {});
            });
        },
    });
    group.button({
        id: "rebuildDataIndex",
        title: window.scribli.languages.rebuildDataIndex,
        desc: window.scribli.languages.rebuildDataIndexTip,
        label: window.scribli.languages.rebuildDataIndex,
        icon: "iconRefresh",
        afterMount: (root) => {
            root.querySelector("#rebuildDataIndex")?.addEventListener("click", () => {
                fetchPost("/api/system/rebuildDataIndex", {});
            });
        },
    });
    group.button({
        id: "clearTempFiles",
        title: window.scribli.languages.clearTempFiles,
        desc: window.scribli.languages.clearTempFilesTip,
        label: window.scribli.languages.purge,
        icon: "iconTrashcan",
        afterMount: (root) => {
            root.querySelector("#clearTempFiles")?.addEventListener("click", () => {
                fetchPost("/api/system/clearTempFiles", {});
            });
        },
    });
    group.button({
        id: "exportLog",
        title: window.scribli.languages.systemLog,
        desc: window.scribli.languages.systemLogTip,
        label: window.scribli.languages.export,
        icon: "iconUpload",
        afterMount: (root) => {
            root.querySelector("#exportLog")?.addEventListener("click", () => {
                fetchPost("/api/system/exportLog", {}, (response) => {
                    void saveExportFile(response.data.zip);
                });
            });
        },
    });
};

export const registerAppTab = (tab: SettingTabBuilder) => {
    registerAppGeneralGroup(tab);
    registerAppDataGroup(tab);
    registerAppMaintenanceGroup(tab);
};

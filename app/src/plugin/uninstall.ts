import {App} from "../index";
import {Plugin} from "./index";
import {getAllModels} from "../layout/getAll";
import {resizeTopBar} from "../layout/util";
import {setTabPosition} from "../layout/tabUtil";
/// #if !BROWSER
import {ipcRenderer} from "electron";
/// #endif
import {Constants} from "../constants";
import {setStorageVal} from "../protyle/util/compatibility";
import {getAllEditor} from "../layout/getAll";
import {unregisterAction} from "../layout/dock/agent/frontendActions";

export const uninstall = (app: App, name: string, isReload: boolean) => {
    app.plugins.find((plugin: Plugin, index) => {
        if (plugin.name === name) {
            try {
                plugin.onunload();
            } catch (e) {
                console.error(`plugin ${plugin.name} onunload error:`, e);
            }
            try {
                plugin.kernel.destroy();
            } catch (e) {
                console.error(`plugin ${plugin.name} kernel destroy error:`, e);
            }
            if (!isReload) {
                try {
                    plugin.uninstall();
                } catch (e) {
                    console.error(`plugin ${plugin.name} uninstall error:`, e);
                }
                window.scribli.storage[Constants.LOCAL_PLUGIN_DOCKS][plugin.name] = {};
                setStorageVal(Constants.LOCAL_PLUGIN_DOCKS, window.scribli.storage[Constants.LOCAL_PLUGIN_DOCKS]);
            }
            // rm tab
            const modelsKeys = Object.keys(plugin.models);
            getAllModels().custom.forEach(custom => {
                if (modelsKeys.includes(custom.type)) {
                    if (isReload) {
                        if (custom.update) {
                            custom.update();
                        }
                    } else {
                        custom.parent.parent.removeTab(custom.parent.id);
                    }
                }
            });
            // rm topBar
            for (let i = 0; i < plugin.topBarIcons.length; i++) {
                const item = plugin.topBarIcons[i];
                item.remove();
                plugin.topBarIcons.splice(i, 1);
                i--;
            }
            // rm agent actions
            plugin.agentActions.forEach(name => unregisterAction(name));
            // rm statusBar
            plugin.statusBarIcons.forEach(item => {
                item.remove();
            });
            // rm dock
            const docksKeys = Object.keys(plugin.docks);
            docksKeys.forEach(key => {
                if (window.scribli.layout.leftDock && Object.keys(window.scribli.layout.leftDock.data).includes(key)) {
                    window.scribli.layout.leftDock.remove(key);
                } else if (window.scribli.layout.rightDock && Object.keys(window.scribli.layout.rightDock.data).includes(key)) {
                    window.scribli.layout.rightDock.remove(key);
                } else if (window.scribli.layout.bottomDock && Object.keys(window.scribli.layout.bottomDock.data).includes(key)) {
                    window.scribli.layout.bottomDock.remove(key);
                }
            });
            resizeTopBar();
            setTabPosition(true);
            // rm listen
            Array.from(document.childNodes).find(item => {
                if (item.nodeType === 8 && item.textContent === name) {
                    item.remove();
                    return true;
                }
            });
            // rm plugin
            app.plugins.splice(index, 1);
            // rm icons
            document.querySelector(`svg[data-name="${plugin.name}"]`)?.remove();
            // rm protyle toolbar
            getAllEditor().forEach(editor => {
                editor.protyle.toolbar.update(editor.protyle);
            });
            // rm style
            document.getElementById("pluginsStyle" + name)?.remove();
            /// #if !BROWSER
            plugin.commands.forEach(command => {
                if (command.globalCallback && command.customHotkey) {
                    ipcRenderer.send(Constants.SCRIBLI_CMD, {
                        cmd: "unregisterGlobalShortcut",
                        accelerator: command.customHotkey
                    });
                }
            });
            /// #endif
            return true;
        }
    });
};

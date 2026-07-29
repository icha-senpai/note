import {openMobileFileById} from "../editor";
import {
    processSync,
    progressLoading,
    setDefRefCount,
    setRefDynamicText,
    transactionError
} from "../../dialog/processSystem";
import {App} from "../../index";
import {reloadPlugin} from "../../plugin/loader";
import {reloadEmoji} from "../../emoji";
import {renderSnippet} from "../../config/util/snippets";
import {redirectToCheckAuth} from "../../util/pathName";
import {reloadSync} from "../../util/reloadSync";
import {setEmpty} from "./setEmpty";
import {activateOnboarding} from "../../onboarding";
import {clearMobileBackForward} from "./MobileBackFoward";

let statusTimeout: number;
const statusElement = document.querySelector("#status") as HTMLElement;

export const onMessage = (app: App, data: IWebSocketData) => {
    if (data) {
        switch (data.cmd) {
            case "logoutAuth":
                redirectToCheckAuth();
                break;
            case "backgroundtask":
                if (!document.querySelector("#keyboardToolbar").classList.contains("fn__none") ||
                    window.scribli.config.appearance.hideStatusBar) {
                    return;
                }
                if (data.data.tasks.length === 0) {
                    statusElement.style.bottom = "";
                } else {
                    clearTimeout(statusTimeout);
                    statusElement.innerHTML = `<div class="fn__flex">${data.data.tasks[0].action}<div class="fn__progress"><div></div></div>`;
                    statusElement.style.bottom = "0";
                }
                break;
            case "setAppearance":
                window.location.reload();
                break;
            case "setSnippet":
                window.scribli.config.snippet = data.data;
                renderSnippet();
                break;
            case "setDefRefCount":
                setDefRefCount(data.data);
                break;
            case "reloadTag":
                window.scribli.mobile.docks.tag?.update();
                break;
            case "setRefDynamicText":
                setRefDynamicText(data.data);
                break;
            case "reloadPlugin":
                reloadPlugin(app, data.data);
                break;
            case "reloadEmojiConf":
                reloadEmoji();
                break;
            case "syncMergeResult":
                reloadSync(app, data.data);
                break;
            case "setConf":
                window.scribli.config = data.data;
                break;
            case "setPublish":
                window.scribli.config.publish = data.data;
                setPublish();
                break;
            case "reloaddoc":
                reloadSync(this, {upsertRootIDs: [data.data], removeRootIDs: []}, false, false, true);
                break;
            case "readonly":
                window.scribli.config.editor.readOnly = data.data;
                break;
            case "closeBox":
            case "removeBox": {
                const closesCurrentEditor = window.scribli.mobile.editor?.protyle.notebookId === data.data.box;
                clearMobileBackForward(closesCurrentEditor ? undefined : data.data.box);
                if (closesCurrentEditor) {
                    window.scribli.mobile.editor.destroy();
                    window.scribli.mobile.editor.protyle.element.innerHTML = "";
                    window.scribli.mobile.editor = undefined;
                    setEmpty(app);
                }
                break;
            }
            case "onboarding":
                void activateOnboarding(app, data.data);
                break;
            case "removeDoc":
                if (window.scribli.config.onboarding?.newUser && !window.scribli.config.onboarding.dismissed &&
                    data.data.ids.includes(window.scribli.config.onboarding.documentID)) {
                    void activateOnboarding(app, window.scribli.config.onboarding);
                }
                break;
            case "setLocalStorageVal":
                window.scribli.storage[data.data.key] = data.data.val;
                break;
            case "setLocalStorageVals":
                Object.keys(data.data.keyVals).forEach((k) => {
                    window.scribli.storage[k] = data.data.keyVals[k];
                });
                break;
            case "removeLocalStorageVal":
                delete window.scribli.storage[data.data.key];
                break;
            case "removeLocalStorageVals":
                data.data.keys.forEach((k: string) => {
                    delete window.scribli.storage[k];
                });
                break;
            case"progress":
                progressLoading(data);
                break;
            case"syncing":
                processSync(data, app.plugins);
                if (data.code === 1) {
                    document.getElementById("toolbarSync").classList.add("fn__none");
                }
                break;
            case "openFileById":
                openMobileFileById(app, data.data.id);
                break;
            case"txerr":
                transactionError(data.msg);
                break;
            case"statusbar":
                if (!document.querySelector("#keyboardToolbar").classList.contains("fn__none") ||
                    window.scribli.config.appearance.hideStatusBar) {
                    return;
                }
                clearTimeout(statusTimeout);
                statusElement.innerHTML = data.msg;
                statusElement.style.bottom = "0";
                statusTimeout = window.setTimeout(() => {
                    statusElement.style.bottom = "";
                }, 12000);
                break;
        }
    }
};

const setPublish = () => {
    const accessElement = window.scribli.mobile.docks.file.element.previousElementSibling.querySelector('[data-type="publish-access"]');
    if (!window.scribli.config.publish.enable) {
        accessElement.classList.remove("block__icon--active");
        accessElement.classList.add("fn__none");
        window.scribli.mobile.docks.file.element.querySelectorAll(".b3-list-item__icon").forEach(item => {
            item.classList.remove("fn__none");
            item.nextElementSibling.classList.add("fn__none");
        });
    } else {
        accessElement.classList.remove("fn__none");
    }

};

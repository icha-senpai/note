import {Constants} from "../constants";
import {fetchPost, fetchSyncPost} from "../util/fetch";
import {openDataMigration} from "../menus/dataMigration";
import {syncGuide} from "../sync/syncGuide";
import {setNoteBook} from "../util/pathName";
import type {App} from "../index";
import {openFileById} from "../editor/util";

export const ensureOnboarding = async () => {
    const onboarding = window.scribli.config.onboarding;
    if (!onboarding?.newUser || onboarding.dismissed || window.scribli.config.readonly || window.scribli.isPublish) {
        return;
    }
    try {
        const response = await fetchSyncPost("/api/system/ensureOnboarding", {});
        if (response.code === 0) {
            window.scribli.config.onboarding = response.data;
        }
    } catch (error) {
        console.warn("ensure onboarding failed", error);
    }
};

const shouldShowOnboarding = () => {
    return window.scribli.config.onboarding?.newUser &&
        window.scribli.config.onboarding.state === "completed" &&
        window.scribli.config.onboarding.documentID &&
        !window.scribli.config.onboarding.dismissed;
};

let pendingSyncHandler: (() => void) | undefined;

const dismissOnboarding = () => {
    if (pendingSyncHandler) {
        window.removeEventListener("scribli-sync-success", pendingSyncHandler);
        pendingSyncHandler = undefined;
    }
    const onboardingElement = document.querySelector(".onboarding");
    onboardingElement?.parentElement?.classList.remove("onboarding-container");
    onboardingElement?.remove();
    window.scribli.config.onboarding.dismissed = true;
    fetchPost("/api/system/dismissOnboarding", {});
};

const syncAndDismissOnSuccess = (app: App) => {
    if (pendingSyncHandler) {
        window.removeEventListener("scribli-sync-success", pendingSyncHandler);
    }
    pendingSyncHandler = () => {
        pendingSyncHandler = undefined;
        dismissOnboarding();
    };
    window.addEventListener("scribli-sync-success", pendingSyncHandler, {once: true});
    syncGuide(app);
};

const renderOnboarding = (app: App) => {
    if (!shouldShowOnboarding() || document.querySelector(".onboarding")) {
        return;
    }
    const element = document.createElement("section");
    element.className = "onboarding";
    element.innerHTML = `<button class="onboarding__close" data-type="close" aria-label="${window.scribli.languages.close}">
    <svg><use xlink:href="#iconCloseRound"></use></svg>
</button>
<div class="onboarding__title">&#x1F389; ${window.scribli.languages.onboardingWelcome}</div>
<div class="onboarding__desc">${window.scribli.languages.onboardingDescription}</div>
<button class="b3-button b3-button--outline fn__block" data-type="import">
    <svg><use xlink:href="#iconDownload"></use></svg>${window.scribli.languages.importExistingData}
</button>
<button class="b3-button b3-button--outline fn__block" data-type="sync">
    <svg><use xlink:href="#iconCloud"></use></svg>${window.scribli.languages.settingsAndSync}
</button>`;
    element.addEventListener("click", (event) => {
        const target = (event.target as HTMLElement).closest("[data-type]") as HTMLElement;
        if (!target) {
            return;
        }
        switch (target.dataset.type) {
            case "close":
                dismissOnboarding();
                break;
            case "import":
                openDataMigration({
                    mode: "onboarding",
                    notebookID: window.scribli.config.onboarding.notebookID,
                    onContentImportComplete: dismissOnboarding,
                });
                break;
            case "sync":
                syncAndDismissOnSuccess(app);
                break;
        }
    });
    let containerElement = document.body;
    const editorContainerElement = document.querySelector(".layout__center") as HTMLElement;
    if (editorContainerElement) {
        containerElement = editorContainerElement;
        containerElement.classList.add("onboarding-container");
        element.classList.add("onboarding--editor");
    }
    containerElement.append(element);
};

export const openDesktopOnboarding = (app: App) => {
    if (!shouldShowOnboarding()) {
        return;
    }
    window.setTimeout(() => {
        openFileById({
            app,
            id: window.scribli.config.onboarding.documentID,
            action: [Constants.CB_GET_FOCUSFIRST],
        });
        renderOnboarding(app);
    });
};

export const activateOnboarding = async (app: App, onboarding: Config.IConf["onboarding"]) => {
    window.scribli.config.onboarding = onboarding;
    await ensureOnboarding();
    setNoteBook(() => {
        openDesktopOnboarding(app);
    });
};

import type {SettingTabBuilder} from "../setting/builder";
import {Constants} from "../../constants";
import {fetchPost} from "../../util/fetch";
import {confirmDialog} from "../../dialog/confirmDialog";
import {showMessage} from "../../dialog/message";
import {processSync} from "../../dialog/processSystem";
import {writeText} from "../../protyle/util/compatibility";
import {bindSyncCloudListEvent, renderSyncCloudList, setKey} from "../../sync/syncGuide";
import {Dialog} from "../../dialog";
import {genConfigItemMainHtml, genConfigItemName} from "../render/fragments";
import {getSyncProviderConfigKeywords} from "./syncUi";
import {patchSyncConfig} from "./syncRuntime";

const registerSyncGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("sync", window.scribli.languages.configGroupSync);

    group.select("sync.provider", {
        title: window.scribli.languages.syncProvider,
        desc: window.scribli.languages.syncProviderTip,
        options: [
            {value: 2, label: "S3"},
            {value: 3, label: "WebDAV"},
            ...(["std", "docker"].includes(window.scribli.config.system.container) ? [{value: 4, label: window.scribli.languages.localFileSystem}] : []),
        ],
        save: (value) => patchSyncConfig("sync.provider", value),
    });
    group.slot({
        key: "syncProviderConfig",
        keywords: getSyncProviderConfigKeywords(),
        html: () => '<div id="syncProviderConfig" class="b3-label config-item"></div>',
    });
    group.switch("sync.enabled", {
        title: window.scribli.languages.openSyncTip1,
        desc: window.scribli.languages.openSyncTip2,
        save: (value) => patchSyncConfig("sync.enabled", value),
    });
    group.switch("sync.generateConflictDoc", {
        title: window.scribli.languages.generateConflictDoc,
        desc: window.scribli.languages.generateConflictDocTip,
        save: (value) => patchSyncConfig("sync.generateConflictDoc", value),
    });
    group.select("sync.mode", {
        title: window.scribli.languages.syncMode,
        desc: window.scribli.languages.syncModeTip,
        options: [
            {value: 1, label: window.scribli.languages.syncMode1},
            {value: 2, label: window.scribli.languages.syncMode2},
            {value: 3, label: window.scribli.languages.syncMode3},
        ],
        save: (value) => patchSyncConfig("sync.mode", value),
    });
    group.number("sync.interval", {
        title: window.scribli.languages.syncInterval,
        desc: window.scribli.languages.syncIntervalTip,
        min: 30,
        max: 43200,
        unit: window.scribli.languages.second,
        save: (value) => patchSyncConfig("sync.interval", value),
    });
    group.switch("sync.perception", {
        title: window.scribli.languages.syncPerception,
        desc: window.scribli.languages.syncPerceptionTip,
        save: (value) => patchSyncConfig("sync.perception", value),
    });
    group.slot({
        key: "syncCloudDir",
        keywords: [window.scribli.languages.cloudSyncDir, window.scribli.languages.cloudSyncDirTip, window.scribli.languages.config],
        html: () => `<div class="b3-label config-item" id="syncCloudDirBlock">
    <div class="fn__flex config-wrap">
        ${genConfigItemMainHtml(window.scribli.languages.cloudSyncDir, window.scribli.languages.cloudSyncDirTip)}
        <div class="fn__space"></div>
        <button class="b3-button b3-button--outline fn__flex-center fn__size200" data-action="config">
            <svg><use xlink:href="#iconSettings"></use></svg>${window.scribli.languages.config}
        </button>
    </div>
    <div id="syncCloudList" class="fn__none"></div>
</div>`,
        afterMount: mountSyncCloudDir,
    });
};

const mountSyncCloudDir = (root: HTMLElement) => {
    const cloudListElement = root.querySelector("#syncCloudList");
    if (cloudListElement) {
        bindSyncCloudListEvent(cloudListElement);
        root.querySelector('#syncCloudDirBlock [data-action="config"]')?.addEventListener("click", () => {
            const hidden = cloudListElement.classList.toggle("fn__none");
            if (!hidden) {
                renderSyncCloudList(cloudListElement, true);
            }
        });
    }
};

const registerRepoGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("repo", window.scribli.languages.configGroupLocalDataRepo);

    group.slot({
        key: "repoKey",
        keywords: [
            window.scribli.languages.dataRepoKey,
            window.scribli.languages.dataRepoKeyTip1,
            window.scribli.languages.dataRepoKeyTip2,
            window.scribli.languages.importKey,
            window.scribli.languages.genKey,
            window.scribli.languages.genKeyByPW,
            window.scribli.languages.copyKey,
            window.scribli.languages.resetRepo,
        ],
        html: () => `<div class="fn__flex b3-label config-item config-wrap">
    <div class="fn__flex-1 fn__flex-center">
        ${genConfigItemName(window.scribli.languages.dataRepoKey)}
        <div class="fn__hr--small"></div>
        <div class="b3-label__text">
            ${window.scribli.languages.dataRepoKeyTip1}
            <div class="fn__hr--small"></div>
            <span class="ft__error">${window.scribli.languages.dataRepoKeyTip2}</span>
        </div>
    </div>
    <div class="fn__space"></div>
    <div class="fn__size200 fn__flex-center fn__none" id="repoKeyActionsEmpty">
        <button class="b3-button b3-button--outline fn__block" id="importKey"><svg><use xlink:href="#iconDownload"></use></svg>${window.scribli.languages.importKey}</button>
        <div class="fn__hr"></div>
        <button class="b3-button b3-button--outline fn__block" id="initKey"><svg><use xlink:href="#iconLock"></use></svg>${window.scribli.languages.genKey}</button>
        <div class="fn__hr"></div>
        <button class="b3-button b3-button--outline fn__block" id="initKeyByPW"><svg><use xlink:href="#iconHand"></use></svg>${window.scribli.languages.genKeyByPW}</button>
    </div>
    <div class="fn__size200 fn__flex-center fn__none" id="repoKeyActionsSet">
        <button class="b3-button b3-button--outline fn__block" id="copyKey"><svg><use xlink:href="#iconCopy"></use></svg>${window.scribli.languages.copyKey}</button>
        <div class="fn__hr"></div>
        <button class="b3-button b3-button--outline fn__block" id="resetRepo"><svg><use xlink:href="#iconUndo"></use></svg>${window.scribli.languages.resetRepo}</button>
    </div>
</div>`,
        afterMount: mountRepoKey,
    });
    group.stack({
        key: "repoPurge",
        keywords: [
            window.scribli.languages.dataRepoPurge,
            window.scribli.languages.dataRepoPurgeTip,
            window.scribli.languages.dataRepoAutoPurgeIndexRetentionDays,
            window.scribli.languages.dataRepoAutoPurgeRetentionIndexesDaily,
        ],
        afterMount: (root) => {
            root.querySelector("#purgeRepo")?.addEventListener("click", () => {
                confirmDialog("♻️ " + window.scribli.languages.dataRepoPurge, window.scribli.languages.dataRepoPurgeConfirm, () => {
                    fetchPost("/api/repo/purgeRepo");
                });
            });
        },
    }, (stack) => {
        stack.title(window.scribli.languages.dataRepoPurge);
        stack.desc(window.scribli.languages.dataRepoPurgeTip);
        stack.button({
            id: "purgeRepo",
            label: window.scribli.languages.purge,
            icon: "iconTrashcan",
        });
        stack.number("repo.indexRetentionDays", {
            desc: window.scribli.languages.dataRepoAutoPurgeIndexRetentionDays,
            min: 1,
        });
        stack.number("repo.retentionIndexesDaily", {
            desc: window.scribli.languages.dataRepoAutoPurgeRetentionIndexesDaily,
            min: 1,
        });
    });
};

const mountRepoKey = (root: HTMLElement) => {
    const emptyElement = root.querySelector("#repoKeyActionsEmpty");
    const setElement = root.querySelector("#repoKeyActionsSet");
    const toggleRepoKeyActions = () => {
        const hasKey = Boolean(window.scribli.config.repo.key);
        emptyElement?.classList.toggle("fn__none", hasKey);
        setElement?.classList.toggle("fn__none", !hasKey);
    };
    toggleRepoKeyActions();
    root.querySelector("#importKey")?.addEventListener("click", () => {
        const passwordDialog = new Dialog({
            title: "🔑 " + window.scribli.languages.key,
            content: `<div class="b3-dialog__content" style="display:flex">
    <textarea spellcheck="false" style="resize: none;flex:1" class="b3-text-field fn__block" placeholder="${window.scribli.languages.keyPlaceholder}"></textarea>
</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages.confirm}</button>
</div>`,
            width: "520px",
            height: "260px",
        });
        passwordDialog.element.setAttribute("data-key", Constants.DIALOG_PASSWORD);
        const textAreaElement = passwordDialog.element.querySelector("textarea");
        textAreaElement.focus();
        const btnsElement = passwordDialog.element.querySelectorAll(".b3-button");
        btnsElement[0].addEventListener("click", () => {
            passwordDialog.destroy();
        });
        btnsElement[1].addEventListener("click", () => {
            fetchPost("/api/repo/importRepoKey", {key: textAreaElement.value}, (response) => {
                window.scribli.config.repo.key = response.data.key;
                toggleRepoKeyActions();
                passwordDialog.destroy();
            });
        });
    });
    root.querySelector("#initKey")?.addEventListener("click", () => {
        confirmDialog("🔑 " + window.scribli.languages.genKey, window.scribli.languages.initRepoKeyTip, () => {
            fetchPost("/api/repo/initRepoKey", {}, (response) => {
                window.scribli.config.repo.key = response.data.key;
                toggleRepoKeyActions();
            });
        });
    });
    root.querySelector("#initKeyByPW")?.addEventListener("click", () => {
        setKey(false, () => {
            toggleRepoKeyActions();
        });
    });
    root.querySelector("#copyKey")?.addEventListener("click", () => {
        writeText(window.scribli.config.repo.key);
        showMessage(window.scribli.languages.copied);
    });
    root.querySelector("#resetRepo")?.addEventListener("click", () => {
        confirmDialog("⚠️ " + window.scribli.languages.resetRepo, window.scribli.languages.resetRepoTip, () => {
            fetchPost("/api/repo/resetRepo", {}, () => {
                window.scribli.config.repo.key = "";
                window.scribli.config.sync.enabled = false;
                processSync();
                toggleRepoKeyActions();
            });
        });
    });
};

export const registerSyncTab = (tab: SettingTabBuilder) => {
    registerSyncGroup(tab);
    registerRepoGroup(tab);
};

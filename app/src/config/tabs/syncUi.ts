import {showMessage} from "../../dialog/message";
import {fetchPost, fetchSyncPost} from "../../util/fetch";
import {saveExportFile} from "../../protyle/util/compatibility";

/** 按当前配置刷新同步 Tab 可见性与动态面板（供 syncRuntime 调用） */
export const refreshSyncTabPanels = (root: Element) => {
    setSyncConfigItemVisible(root);
    setSyncModeRelatedConfigItemVisible(root);
    renderProviderConfig(root);
};

/** 仅刷新与同步模式相关的配置项可见性（供 syncRuntime 调用） */
export const refreshSyncModeRelatedItems = (root: Element) => {
    setSyncModeRelatedConfigItemVisible(root);
};

const setSyncConfigItemVisible = (root: Element) => {
    const visible = window.scribli.config.sync.provider !== 0;
    [
        "sync.enabled",
        "sync.generateConflictDoc",
        "sync.mode",
        "sync.interval",
        "sync.perception",
        "syncCloudDirBlock",
    ]
    .forEach((id) => {
        root.querySelector(`#${CSS.escape(id)}`)?.closest(".config-item")?.classList.toggle("fn__none", !visible);
    });
};

const setSyncModeRelatedConfigItemVisible = (root: Element) => {
    const syncModeElement = root.querySelector(`#${CSS.escape("sync.mode")}`) as HTMLSelectElement | null;
    if (!syncModeElement) {
        return;
    }
    const syncMode: Config.ISync["mode"] = Number(syncModeElement.value);
    const isAutoSync = syncMode === 1 && window.scribli.config.sync.provider !== 0;
    root.querySelector(`#${CSS.escape("sync.interval")}`)?.closest(".config-item")?.classList.toggle("fn__none", !isAutoSync);
    root.querySelector(`#${CSS.escape("sync.perception")}`)?.closest(".config-item")?.classList.add("fn__none");
};

/** 同步提供商配置区检索关键词（供 syncTab 注册 slot） */
export const getSyncProviderConfigKeywords = (): string[] => buildProviderConfigKeywords();

type SyncProviderConfigKey = Extract<keyof Config.ISync, "s3" | "webdav" | "local">;

type SyncProviderFieldDef =
    | {type: "input"; label: string; id: string; attrs?: string}
    | {type: "password"; label: string; id: string}
    | {type: "select"; label: string; id: string; options: {value: string; label: string}[]};

type SyncProviderIntroDef = {
    genIntro: () => string;
    genUnavailableIntro: () => string;
    isProviderConfigAllowed: () => boolean;
};

type SyncThirdPartyProviderDef = SyncProviderIntroDef & {
    configKey: SyncProviderConfigKey;
    api: string;
    getConfig: () => Config.ISync[SyncProviderConfigKey];
    fields: SyncProviderFieldDef[];
};

type SyncProviderDef = SyncProviderIntroDef | SyncThirdPartyProviderDef;

const isThirdPartySyncProviderDef = (def: SyncProviderDef): def is SyncThirdPartyProviderDef => {
    return "configKey" in def;
};

const genThirdPartyUnavailableIntro = (): string => {
    return `<div>
    ${window.scribli.languages.syncProviderConfigTip}
</div>`;
};

const isProviderConfigAllowed = () => true;

const SYNC_PROVIDER_DEFS: Record<Config.ISync["provider"], SyncProviderDef> = {
    0: {
        isProviderConfigAllowed: () => false,
        genIntro: () => `<div class="b3-label b3-label--inner">${window.scribli.languages.syncProviderConfigTip}</div>`,
        genUnavailableIntro: () => `<div class="b3-label b3-label--inner">${window.scribli.languages.syncProviderConfigTip}</div>`,
    },
    2: {
        isProviderConfigAllowed,
        configKey: "s3",
        api: "/api/sync/setSyncProviderS3",
        getConfig: () => window.scribli.config.sync.s3,
        genIntro: () => `<div class="b3-label b3-label--inner">
    ${window.scribli.languages.syncThirdPartyProviderS3Intro}
    <div class="fn__hr"></div>
    ${window.scribli.languages.syncThirdPartyProviderTip}
</div>`,
        genUnavailableIntro: genThirdPartyUnavailableIntro,
        fields: [
            {type: "input", label: "Endpoint", id: "endpoint"},
            {type: "input", label: "Access Key", id: "accessKey"},
            {type: "password", label: "Secret Key", id: "secretKey"},
            {type: "input", label: "Bucket", id: "bucket"},
            {type: "input", label: "Region ID", id: "region"},
            {type: "input", label: "Timeout (s)", id: "timeout", attrs: 'type="number" min="7" max="300"'},
            {type: "select", label: "Addressing", id: "pathStyle", options: [
                {value: "true", label: "Path-style"},
                {value: "false", label: "Virtual-hosted-style"},
            ]},
            {type: "select", label: "TLS Verify", id: "skipTlsVerify", options: [
                {value: "false", label: "Verify"},
                {value: "true", label: "Skip"},
            ]},
            {type: "input", label: "Concurrent Reqs", id: "concurrentReqs", attrs: 'type="number" min="1" max="16"'},
        ],
    },
    3: {
        isProviderConfigAllowed,
        configKey: "webdav",
        api: "/api/sync/setSyncProviderWebDAV",
        getConfig: () => window.scribli.config.sync.webdav,
        genIntro: () => `<div class="b3-label b3-label--inner">
    ${window.scribli.languages.syncThirdPartyProviderWebDAVIntro}
    <div class="fn__hr"></div>
    ${window.scribli.languages.syncThirdPartyProviderTip}
</div>`,
        genUnavailableIntro: genThirdPartyUnavailableIntro,
        fields: [
            {type: "input", label: "Endpoint", id: "endpoint"},
            {type: "input", label: "Username", id: "username"},
            {type: "password", label: "Password", id: "password"},
            {type: "input", label: "Timeout (s)", id: "timeout", attrs: 'type="number" min="7" max="300"'},
            {type: "select", label: "TLS Verify", id: "skipTlsVerify", options: [
                {value: "false", label: "Verify"},
                {value: "true", label: "Skip"},
            ]},
            {type: "input", label: "Concurrent Reqs", id: "concurrentReqs", attrs: 'type="number" min="1" max="16"'},
        ],
    },
    4: {
        isProviderConfigAllowed,
        configKey: "local",
        api: "/api/sync/setSyncProviderLocal",
        getConfig: () => window.scribli.config.sync.local,
        genIntro: () => `<div class="b3-label b3-label--inner">
    <div class="ft__error">
        ${window.scribli.languages.mobileNotSupport}
    </div>
    <div class="fn__hr"></div>
    ${window.scribli.languages.syncThirdPartyProviderLocalIntro}
</div>`,
        genUnavailableIntro: () => `${genThirdPartyUnavailableIntro()}<div class="ft__error">
    <div class="fn__hr--b"></div>
    ${window.scribli.languages.mobileNotSupport}
</div>`,
        fields: [
            {type: "input", label: "Endpoint", id: "endpoint"},
            {type: "input", label: "Timeout (s)", id: "timeout", attrs: 'type="number" min="7" max="300"'},
            {type: "input", label: "Concurrent Reqs", id: "concurrentReqs", attrs: 'type="number" min="1" max="1024"'},
        ],
    },
};

const buildProviderConfigKeywords = (): string[] => {
    return [
        // 服务不可用 / 本地等提示
        window.scribli.languages.syncProviderConfigTip,
        window.scribli.languages.mobileNotSupport,
        // S3 / WebDAV / 本地第三方
        window.scribli.languages.syncThirdPartyProviderS3Intro,
        window.scribli.languages.syncThirdPartyProviderWebDAVIntro,
        window.scribli.languages.syncThirdPartyProviderLocalIntro,
        window.scribli.languages.syncThirdPartyProviderTip,
        // 操作按钮
        window.scribli.languages.import,
        window.scribli.languages.export,
        // 表单标签与选项（硬编码英文）
        "Endpoint",
        "Access Key",
        "Secret Key",
        "Bucket",
        "Region ID",
        "Timeout (s)",
        "Addressing",
        "Path-style",
        "Virtual-hosted-style",
        "TLS Verify",
        "Verify",
        "Skip",
        "Concurrent Reqs",
        "Username",
        "Password",
    ];
};

const renderProviderConfig = (root: Element) => {
    const providerConfigElement = root.querySelector("#syncProviderConfig");
    if (!providerConfigElement) {
        return;
    }

    const def = SYNC_PROVIDER_DEFS[window.scribli.config.sync.provider];
    let html = "";
    if (def) {
        if (!def.isProviderConfigAllowed()) {
            html = def.genUnavailableIntro();
        } else if (isThirdPartySyncProviderDef(def)) {
            html = `${def.genIntro()}${def.fields.map(genProviderField).join("")}${genProviderActionButtons(def.configKey)}`;
        } else {
            html = def.genIntro();
        }
    }

    providerConfigElement.innerHTML = html;
    bindProviderConfigEvent(providerConfigElement, root);
};

const genProviderField = (field: SyncProviderFieldDef): string => {
    switch (field.type) {
        case "input":
            return genProviderFlexInput(field.label, field.id, field.attrs);
        case "password":
            return genProviderFlexPassword(field.label, field.id);
        case "select":
            return genProviderFlexSelect(field.label, field.id, field.options.map((option) => `
    <option value="${option.value}">${option.label}</option>`).join(""));
    }
};

const genProviderFlexInput = (label: string, id: string, attrs = "") => `<div class="b3-label b3-label--inner fn__flex">
    <div class="fn__flex-center fn__size200">${label}</div>
    <div class="fn__space"></div>
    <input id="${id}" class="b3-text-field fn__block"${attrs ? ` ${attrs}` : ""}>
</div>`;

const genProviderFlexPassword = (label: string, id: string) => `<div class="b3-label b3-label--inner fn__flex">
    <div class="fn__flex-center fn__size200">${label}</div>
    <div class="fn__space"></div>
    <div class="b3-form__icona fn__block">
        <input id="${id}" type="password" class="b3-text-field b3-form__icona-input">
        <svg class="b3-form__icona-icon" data-action="togglePassword"><use xlink:href="#iconEye"></use></svg>
    </div>
</div>`;

const genProviderFlexSelect = (label: string, id: string, optionsHtml: string) => `<div class="b3-label b3-label--inner fn__flex">
    <div class="fn__flex-center fn__size200">${label}</div>
    <div class="fn__space"></div>
    <select class="b3-select fn__block" id="${id}">
        ${optionsHtml}
    </select>
</div>`;

const genProviderActionButtons = (dataType: SyncProviderConfigKey) => {
    if (dataType !== "s3" && dataType !== "webdav") {
        return "";
    }
    return `<div class="b3-label b3-label--inner fn__flex fn__flex-wrap">
    <div class="fn__flex-1"></div>
    <button class="b3-button b3-button--outline fn__size200" style="position: relative">
        <input id="importSyncConfig" class="b3-form__upload" type="file" data-type="${dataType}">
        <svg><use xlink:href="#iconDownload"></use></svg>${window.scribli.languages.import}
    </button>
    <div class="fn__space"></div>
    <button class="b3-button b3-button--outline fn__size200" id="exportSyncConfig" data-type="${dataType}">
        <svg><use xlink:href="#iconUpload"></use></svg>${window.scribli.languages.export}
    </button>
</div>`;
};

const syncProviderConfigBoundElements = new WeakSet<Element>();

const bindProviderConfigEvent = (configElement: Element, root: Element) => {
    const togglePasswordIcon = configElement.querySelector('[data-action="togglePassword"]');
    togglePasswordIcon?.addEventListener("click", () => {
        const useElement = togglePasswordIcon.firstElementChild;
        const isEye = useElement.getAttribute("xlink:href") === "#iconEye";
        useElement.setAttribute("xlink:href", isEye ? "#iconEyeoff" : "#iconEye");
        (togglePasswordIcon.previousElementSibling as HTMLInputElement).setAttribute("type", isEye ? "text" : "password");
    });

    const importElement = configElement.querySelector("#importSyncConfig") as HTMLInputElement;
    importElement?.addEventListener("change", () => {
        const formData = new FormData();
        formData.append("file", importElement.files[0]);
        const isS3 = importElement.getAttribute("data-type") === "s3";
        fetchPost(isS3 ? "/api/sync/importSyncProviderS3" : "/api/sync/importSyncProviderWebDAV", formData, (response) => {
            if (isS3) {
                window.scribli.config.sync.s3 = response.data.s3;
            } else {
                window.scribli.config.sync.webdav = response.data.webdav;
            }
            renderProviderConfig(root);
            showMessage(window.scribli.languages.imported);
        });
    });

    const exportButton = configElement.querySelector("#exportSyncConfig");
    exportButton?.addEventListener("click", () => {
        fetchPost(exportButton.getAttribute("data-type") === "s3" ? "/api/sync/exportSyncProviderS3" : "/api/sync/exportSyncProviderWebDAV", {}, (response) => {
            void saveExportFile(response.data.zip);
        });
    });

    const provider = window.scribli.config.sync.provider;
    const def = SYNC_PROVIDER_DEFS[provider];
    if (!isThirdPartySyncProviderDef(def) || !def.isProviderConfigAllowed()) {
        return;
    }
    fillSyncProviderConfigValues(configElement);
    if (syncProviderConfigBoundElements.has(configElement)) {
        return;
    }
    syncProviderConfigBoundElements.add(configElement);
    configElement.addEventListener("change", (event: Event) => {
        const target = event.target as HTMLElement;
        if (!target.matches(".b3-text-field, .b3-select")) {
            return;
        }
        saveSyncProviderConfigValues(configElement);
    });
};

const saveSyncProviderConfigValues = (configElement: Element) => {
    const provider = window.scribli.config.sync.provider;
    const def = SYNC_PROVIDER_DEFS[provider];
    if (!isThirdPartySyncProviderDef(def)) {
        return;
    }
    const data = readProviderConfigFields(configElement, def.getConfig());
    const configKey = def.configKey;
    // 使用 fetchSyncPost：内核返回 code < 0 时 fetchPost 不会调用回调，此处需始终回写界面与已保存配置一致
    fetchSyncPost(def.api, {[configKey]: data})
        .then((response) => {
            if (response.code === 0 && response.data?.[configKey]) {
                window.scribli.config.sync[configKey] = response.data[configKey];
            }
        })
        .finally(() => {
            fillSyncProviderConfigValues(configElement);
        })
        .catch(() => {});
};

const fillSyncProviderConfigValues = (configElement: Element) => {
    const provider = window.scribli.config.sync.provider;
    const def = SYNC_PROVIDER_DEFS[provider];
    if (!isThirdPartySyncProviderDef(def)) {
        return;
    }
    const data = def.getConfig();
    (Object.keys(data) as (keyof typeof data & string)[]).forEach((key) => {
        const el = configElement.querySelector(`#${key}`) as HTMLInputElement | HTMLSelectElement | null;
        if (el) {
            el.value = String(data[key]);
        }
    });
};

const readProviderConfigFields = <T extends object>(configElement: Element, template: T): T => {
    const result = {} as Record<string, unknown>;
    (Object.keys(template) as (keyof T & string)[]).forEach((key) => {
        const el = configElement.querySelector(`#${key}`) as HTMLInputElement | HTMLSelectElement | null;
        if (!el) {
            return;
        }
        const sample = template[key];
        if (typeof sample === "boolean") {
            result[key] = el.value === "true";
        } else if (typeof sample === "number") {
            result[key] = parseInt(el.value, 10);
        } else {
            result[key] = el.value;
        }
    });
    return result as T;
};

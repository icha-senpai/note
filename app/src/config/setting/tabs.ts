import {editorConfigApi} from "../tabs/editorRuntime";
import {fileConfigApi} from "../tabs/fileRuntime";
import {flashcardConfigApi} from "../tabs/flashcardRuntime";
import {aiConfigApi} from "../tabs/aiRuntime";
import {secretsConfigApi} from "../tabs/secretsVariablesRuntime";
import {exportConfigApi} from "../tabs/exportRuntime";
import {searchConfigApi} from "../tabs/searchRuntime";
import {appearanceConfigApi} from "../tabs/appearanceRuntime";
import {mountSyncTabExtras, patchSyncConfig} from "../tabs/syncRuntime";
import {mountAccessTab} from "../tabs/accessRuntime";
import {collectAssetsTabSearchStrings, mountAssetsTab} from "../assets";
import {collectKeymapTabSearchStrings, mountKeymapTab} from "../tabs/keymapUi";
import {SettingBuilder, type SettingTab} from "./builder";
import {registerEditorTab} from "../tabs/editorTab";
import {registerFileTab} from "../tabs/fileTab";
import {registerFlashcardTab} from "../tabs/flashcardTab";
import {registerAiTab} from "../tabs/aiTab";
import {registerSecretsVariablesTab} from "../tabs/secretsVariablesTab";
import {registerExportTab} from "../tabs/exportTab";
import {registerSearchTab} from "../tabs/searchTab";
import {registerAppearanceTab} from "../tabs/appearanceTab";
import {registerSyncTab} from "../tabs/syncTab";
import {registerAccessTab} from "../tabs/accessTab";
import {registerAppTab} from "../tabs/appTab";
import {registerAboutTab} from "../tabs/aboutTab";

const setting = new SettingBuilder();
const settingTabs = {
    editor: setting.tab({
        id: "editor",
        icon: "iconEdit",
        title: () => window.scribli.languages.editor,
        defaultSave: editorConfigApi.patch,
    }, registerEditorTab),
    file: setting.tab({
        id: "file",
        icon: "iconFiles",
        title: () => window.scribli.languages.fileTree,
        defaultSave: fileConfigApi.patch,
    }, registerFileTab),
    appearance: setting.tab({
        id: "appearance",
        icon: "iconTheme",
        title: () => window.scribli.languages.appearance,
        defaultSave: appearanceConfigApi.patch,
    }, registerAppearanceTab),
    flashcard: setting.tab({
        id: "flashcard",
        icon: "iconRiffCard",
        title: () => window.scribli.languages.riffCard,
        defaultSave: flashcardConfigApi.patch,
    }, registerFlashcardTab),
    ai: setting.tab({
        id: "ai",
        icon: "iconSparkles",
        title: () => window.scribli.languages.ai,
        defaultSave: aiConfigApi.patch,
    }, registerAiTab),
    secretsVariables: setting.tab({
        id: "secretsVariables",
        icon: "iconSquareAsterisk",
        title: () => window.scribli.languages.secretsVariables,
        defaultSave: secretsConfigApi.patch,
    }, registerSecretsVariablesTab),
    assets: setting.panel({
        id: "assets",
        icon: "iconImage",
        title: () => window.scribli.languages.assets,
        searchStrings: collectAssetsTabSearchStrings,
        mount: mountAssetsTab,
    }),
    export: setting.tab({
        id: "export",
        icon: "iconUpload",
        title: () => window.scribli.languages.export,
        defaultSave: exportConfigApi.patch,
    }, registerExportTab),
    search: setting.tab({
        id: "search",
        icon: "iconSearch",
        title: () => window.scribli.languages.search,
        defaultSave: searchConfigApi.patch,
    }, registerSearchTab),
    keymap: setting.panel({
        id: "keymap",
        icon: "iconKeymap",
        title: () => window.scribli.languages.keymap,
        searchStrings: collectKeymapTabSearchStrings,
        mount: mountKeymapTab,
    }),
    sync: setting.tab({
        id: "sync",
        icon: "iconCloud",
        title: () => window.scribli.languages.settingsAndSync,
        defaultSave: patchSyncConfig,
        afterMount: mountSyncTabExtras,
    }, registerSyncTab),
    access: setting.tab({
        id: "access",
        icon: "iconLock",
        title: () => window.scribli.languages.authentication,
        afterMount: mountAccessTab,
    }, registerAccessTab),
    app: setting.tab({
        id: "app",
        icon: "iconLayoutGrid",
        title: () => window.scribli.languages.application,
    }, registerAppTab),
    about: setting.tab({
        id: "about",
        icon: "iconInfo",
        title: () => window.scribli.languages.about,
    }, registerAboutTab),
};

export type TSettingTab = keyof typeof settingTabs;

export const getSettingTab = (id: TSettingTab): SettingTab => settingTabs[id];

export interface ISettingTabShell<TId extends string = string> {
    id: TId;
    icon: string;
    title: string;
    hidden?: boolean;
}

let settingTabShellCache: ISettingTabShell<TSettingTab>[] | undefined;

export const getSettingTabDefs = (): ISettingTabShell<TSettingTab>[] => {
    if (settingTabShellCache) {
        return settingTabShellCache;
    }
    settingTabShellCache = (Object.entries(settingTabs) as [TSettingTab, SettingTab][]).map(([id, tab]) => ({
        id,
        icon: tab.icon,
        title: tab.title(),
        hidden: tab.hidden?.(),
    }));
    return settingTabShellCache;
};

export const settingTabToMenuId = (tabId: string): string =>
    "menuConfig" + tabId[0].toUpperCase() + tabId.slice(1);

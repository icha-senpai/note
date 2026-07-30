import {adjustDockPadding} from "../../layout/dock/util";
import {exportLayout} from "../../layout/util";
import {syncHideToolbarLayout, updateBarModeIcon} from "../../layout/topBar";
import {fetchPost} from "../../util/fetch";
import {loadAssets} from "../../util/assets";
import {remountOpenSettingTab} from "../setting/mount";
import {createConfigNamespaceApi} from "../util/namespaceApi";

/** 主题模式下拉框初值：合并 mode / modeOS */
export const appearanceThemeModeValue = (): number =>
    window.scribli.config.appearance.modeOS ? 2 : window.scribli.config.appearance.mode;

/** 主题模式选择：合并 mode / modeOS 后提交 */
export const saveThemeMode = (value: number) => {
    const OSThemeMode = window.matchMedia("(prefers-color-scheme: dark)").matches ? 1 : 0;
    fetchPost("/api/setting/setAppearance", {
        ...window.scribli.config.appearance,
        mode: (value === 2 ? OSThemeMode : value) as Config.IAppearance["mode"],
        modeOS: value === 2,
    });
};

const applyAppearanceConfig = async (data: Config.IAppearance) => {
    if (data.lang !== window.scribli.config.appearance.lang) {
        void exportLayout({
            cb() {
                window.location.reload();
            },
            errorExit: false,
        });
        return;
    }

    if (window.scribli.config.appearance.themeJS) {
        if (data.mode !== window.scribli.config.appearance.mode ||
            (data.mode === window.scribli.config.appearance.mode && (
                (data.mode === 0 && window.scribli.config.appearance.themeLight !== data.themeLight) ||
                (data.mode === 1 && window.scribli.config.appearance.themeDark !== data.themeDark))
            )
        ) {
            if (window.destroyTheme) {
                try {
                    await window.destroyTheme();
                    window.destroyTheme = undefined;
                    document.getElementById("themeScript").remove();
                } catch (e) {
                    console.error("destroyTheme error: " + e);
                }
            } else {
                void exportLayout({
                    errorExit: false,
                    cb() {
                        window.location.reload();
                    },
                });
                return;
            }
        }
    }

    const prevAppearance = window.scribli.config.appearance;
    window.scribli.config.appearance = data;

    document.getElementById("status")?.classList.toggle("fn__none", data.hideStatusBar);
    if (data.hideStatusBar !== prevAppearance.hideStatusBar) {
        adjustDockPadding();
    }
    if (data.hideToolbar !== prevAppearance.hideToolbar) {
        syncHideToolbarLayout();
    }
    updateBarModeIcon();

    loadAssets(data);
    void remountOpenSettingTab("appearance");
};

/** 外观 Tab 命名空间：设置面板注册项 save */
export const appearanceConfigApi = createConfigNamespaceApi<Config.IAppearance>({
    namespace: "appearance",
    getConfig: () => window.scribli.config.appearance,
    setConfig: (data) => {
        void applyAppearanceConfig(data);
    },
    apiPath: "/api/setting/setAppearance",
    applyFromResponse: false,
});

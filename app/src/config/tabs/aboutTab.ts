import type {SettingTabBuilder} from "../setting/builder";
import {Constants} from "../../constants";

const registerAboutVersionGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("version", "");

    group.slot({
        key: "version",
        keywords: [
            window.scribli.languages.currentVer,
            window.scribli.languages.isMsStoreVerTip,
        ],
        html: genAboutVersionHtml,
        afterMount: mountAboutVersionSlot,
    });
};

const genAboutVersionHtml = (): string => {
    if (window.scribli.config.system.isMicrosoftStore) {
        return `<div class="fn__flex b3-label config-item config-wrap">
    <div class="fn__flex-1">
        <div class="config-name">${window.scribli.languages.currentVer} v${Constants.SCRIBLI_VERSION}<span id="isInsider"></span></div>
        <div class="b3-label__text">${window.scribli.languages.isMsStoreVerTip}</div>
    </div>
</div>`;
    }
    return `<div class="fn__flex b3-label config-item config-wrap">
    <div class="fn__flex-1">
        <div class="config-name">${window.scribli.languages.currentVer} v${Constants.SCRIBLI_VERSION}<span id="isInsider"></span></div>
    </div>
</div>`;
};

const mountAboutVersionSlot = (root: HTMLElement) => {
    const isInsiderElement = root.querySelector("#isInsider");
    if (window.scribli.config.system.isInsider && isInsiderElement) {
        isInsiderElement.innerHTML = " <span class='ft__secondary'>Insider Preview</span>";
    }
};

const registerAboutInfoGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("info", "");
    group.slot({
        key: "aboutLogo",
        keywords: [
            window.scribli.languages.siyuanNote,
            window.scribli.languages.slogan,
            window.scribli.languages.about1,
            window.scribli.languages.feedback,
        ],
        html: () => `<div class="fn__flex b3-label config-item config-wrap">
    <div class="fn__flex-1">
        <div class="config-about__logo">
            <img src="/stage/icon.png">
            <span class="fn__space"></span>
            <span>${window.scribli.languages.siyuanNote}</span>
            <span class="fn__space"></span>
            <span class="ft__on-surface">${window.scribli.languages.slogan}</span>
        </div>
        <div class='fn__hr'></div>
        ${window.scribli.languages.about1}${window.scribli.config.system.container === "harmony" ? ` • ${window.scribli.languages.feedback} 845765@qq.com` : ""}
    </div>
</div>`,
    });
};

export const registerAboutTab = (tab: SettingTabBuilder) => {
    registerAboutVersionGroup(tab);
    registerAboutInfoGroup(tab);
};

import {openColorPicker, resolveColor} from "../../util/colorPicker";
import {InterfaceColorRole, InterfaceColors} from "../../util/interfaceColors";
import {appearanceConfigApi} from "./appearanceRuntime";
import {SettingTabBuilder} from "../setting/builder";

export const registerAppearanceColors = (tab: SettingTabBuilder) => {
    const group = tab.group("colors", window.scribli.languages.appearanceColors);
    const roles: Array<[InterfaceColorRole, string, string]> = [["accent", "appearanceAccent", "--b3-theme-primary"], ["background", "appearanceBackground", "--b3-theme-background"],
        ["surface", "appearanceSurface", "--b3-theme-surface"], ["text", "appearanceText", "--b3-theme-on-background"]];
    roles.forEach(([role, label, variable]) => group.slot({key: role, keywords: [window.scribli.languages.color], html: () =>
        `<div class="fn__flex b3-label config-item"><span class="fn__flex-1">${window.scribli.languages[label]}</span><button type="button" class="scribli-color-picker__swatch" data-interface-color="${role}" aria-label="${window.scribli.languages[label]}" style="background-color:${resolveColor(`var(${variable})`)}"></button></div>`,
    afterMount(root) {
        root.querySelector(`[data-interface-color="${role}"]`).addEventListener("click", () => {
            const mode = String(window.scribli.config.appearance.mode) as "0" | "1";
            openColorPicker({title: window.scribli.languages[label], color: resolveColor(`var(${variable})`), allowReset: true, onApply(color) {
                const colors: InterfaceColors = JSON.parse(JSON.stringify(window.scribli.config.appearance.interfaceColors || {}));
                colors[mode] = {...colors[mode]};
                if (color) { colors[mode][role] = color; } else { delete colors[mode][role]; }
                appearanceConfigApi.patch("appearance.interfaceColors", colors);
            }});
        });
    }}));
    group.slot({key: "reset", keywords: [window.scribli.languages.reset], html: () => `<div class="fn__flex b3-label config-item"><span class="fn__flex-1">${window.scribli.languages.appearanceColorReset}</span><button type="button" class="b3-button b3-button--outline" data-interface-reset>${window.scribli.languages.reset}</button></div>`, afterMount(root) {
        root.querySelector("[data-interface-reset]").addEventListener("click", () => {
            const colors: InterfaceColors = JSON.parse(JSON.stringify(window.scribli.config.appearance.interfaceColors || {}));
            delete colors[String(window.scribli.config.appearance.mode) as "0" | "1"];
            appearanceConfigApi.patch("appearance.interfaceColors", colors);
        });
    }});
};

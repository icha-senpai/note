import {MenuItem} from "./Menu";
import {copySubMenu} from "./commonMenuItem";

export const initSearchMenu = (id: string) => {
    window.scribli.menus.menu.remove();
    window.scribli.menus.menu.append(new MenuItem({
        id: "copy",
        icon: "iconCopy",
        label: window.scribli.languages.copy,
        type: "submenu",
        submenu: copySubMenu([id])
    }).element);
    return window.scribli.menus.menu;
};

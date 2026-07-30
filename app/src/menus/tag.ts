import {MenuItem} from "./Menu";
import {fetchPost} from "../util/fetch";
import {confirmDialog} from "../dialog/confirmDialog";
import {escapeHtml} from "../util/escape";
import {renameTag} from "../util/noRelyPCFunction";
import {getDockByType} from "../layout/tabUtil";
import {Tag} from "../layout/dock/Tag";
import {Constants} from "../constants";

export const openTagMenu = (element: HTMLElement, event: MouseEvent, labelName: string) => {
    if (!window.scribli.menus.menu.element.classList.contains("fn__none") &&
        window.scribli.menus.menu.element.getAttribute("data-name") === Constants.MENU_TAG) {
        window.scribli.menus.menu.remove();
        return;
    }
    window.scribli.menus.menu.remove();
    window.scribli.menus.menu.append(new MenuItem({
        icon: "iconEdit",
        label: window.scribli.languages.rename,
        click: () => {
            renameTag(labelName);
        }
    }).element);
    window.scribli.menus.menu.append(new MenuItem({
        icon: "iconTrashcan",
        label: window.scribli.languages.remove,
        click: () => {
            confirmDialog(window.scribli.languages.deleteOpConfirm, `${window.scribli.languages.confirmDelete} <b>${escapeHtml(labelName)}</b>?`, () => {
                fetchPost("/api/tag/removeTag", {label: labelName}, () => {
                    const dockTag = getDockByType("tag");
                    (dockTag.data.tag as Tag).update();
                });
            }, undefined, true);
        }
    }).element);
    window.scribli.menus.menu.element.setAttribute("data-name", Constants.MENU_TAG);
    window.scribli.menus.menu.popup({x: event.clientX - 11, y: event.clientY + 11, h: 22, w: 12});
};

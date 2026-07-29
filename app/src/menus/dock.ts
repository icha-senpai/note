import {MenuItem} from "./Menu";
import {Constants} from "../constants";

const moveMenuItem = (label: string, target: Element) => {
    return new MenuItem({
        id: label,
        label: window.scribli.languages[label],
        icon: label.replace("moveTo", "icon"),
        click: () => {
            if (label.indexOf("moveToLeft") > -1) {
                window.scribli.layout.leftDock.add(label.endsWith("Top") ? 0 : 1, target);
            } else if (label.indexOf("moveToRight") > -1) {
                window.scribli.layout.rightDock.add(label.endsWith("Top") ? 0 : 1, target);
            } else if (label.indexOf("moveToBottom") > -1) {
                window.scribli.layout.bottomDock.add(label.endsWith("Left") ? 0 : 1, target);
            }
        }
    });
};

export const initDockMenu = (target: Element) => {
    window.scribli.menus.menu.remove();
    window.scribli.menus.menu.element.setAttribute("data-name", Constants.MENU_DOCK);
    window.scribli.menus.menu.append(moveMenuItem("moveToLeftTop", target).element);
    window.scribli.menus.menu.append(moveMenuItem("moveToLeftBottom", target).element);
    window.scribli.menus.menu.append(moveMenuItem("moveToRightTop", target).element);
    window.scribli.menus.menu.append(moveMenuItem("moveToRightBottom", target).element);
    window.scribli.menus.menu.append(moveMenuItem("moveToBottomLeft", target).element);
    window.scribli.menus.menu.append(moveMenuItem("moveToBottomRight", target).element);
    return window.scribli.menus.menu;
};

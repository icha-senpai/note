/// #if !MOBILE
import {getAllModels} from "../../layout/getAll";
/// #endif
import {hasClosestByAttribute, hasClosestByClassName, hasTopClosestByClassName} from "../../protyle/util/hasClosest";
import {hideAllElements} from "../../protyle/ui/hideElements";
import {isWindow} from "../../util/functions";
import {writeText} from "../../protyle/util/compatibility";
import {showMessage} from "../../dialog/message";
import {cancelDrag} from "./dragover";
import {nbsp2space, removeZWJ} from "../../protyle/util/normalizeText";
import {getDockByType} from "../../layout/tabUtil";

export const globalClickHideMenu = (element: HTMLElement) => {
    if (!window.scribli.menus.menu.element.contains(element) && !hasClosestByAttribute(element, "data-menu", "true")) {
        if (getSelection().rangeCount > 0 && window.scribli.menus.menu.element.contains(getSelection().getRangeAt(0).startContainer) &&
            window.scribli.menus.menu.element.contains(document.activeElement)) {
            // https://ld246.com/article/1654567749834/comment/1654589171218#comments
        } else {
            window.scribli.menus.menu.remove();
        }
    }
};

export const globalClick = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    cancelDrag();

    globalClickHideMenu(target);

    const protyleElement = hasClosestByClassName(target, "protyle", true);
    if (protyleElement) {
        const wysiwygElement = protyleElement.querySelector(".protyle-wysiwyg");
        if (wysiwygElement.getAttribute("data-readonly") === "true" || !wysiwygElement.contains(target)) {
            wysiwygElement.dispatchEvent(new Event("focusin"));
        }
    }

    if (!hasTopClosestByClassName(target, "protyle-util") &&
        !hasTopClosestByClassName(target, "protyle-toolbar")) {
        document.querySelectorAll(".protyle-font").forEach((item: HTMLElement) => {
            item.parentElement.classList.add("fn__none");
        });
    }

    const copyElement = hasTopClosestByClassName(target, "protyle-action__copy");
    if (copyElement) {
        writeText(removeZWJ(nbsp2space(copyElement.parentElement.nextElementSibling.textContent.replace(/\n$/, ""))));
        showMessage(window.scribli.languages.copied, 2000);
        event.preventDefault();
        return;
    }

    /// #if !MOBILE
    // dock float 时，点击空白处，隐藏 dock。场景：文档树上重命名后
    if (!isWindow() && window.scribli.layout.leftDock &&
        !hasClosestByClassName(target, "b3-dialog--open", true) &&
        !hasClosestByClassName(target, "b3-menu") &&
        !hasClosestByClassName(target, "block__popover") &&
        !hasClosestByClassName(target, "dock") &&
        !hasClosestByClassName(target, "layout--float", true)
    ) {
        window.scribli.layout.bottomDock.hideDock();
        window.scribli.layout.leftDock.hideDock();
        window.scribli.layout.rightDock.hideDock();
    }
    // Dock item click
    const dockItemElement = hasClosestByClassName(target, "dock__item");
    if (dockItemElement) {
        const type = dockItemElement.getAttribute("data-type") as TDock;
        if (type) {
            getDockByType(type).toggleModel(type, false, true);
        }
    }

    if (!hasClosestByClassName(target, "pdf__outer")) {
        hideAllElements(["pdfutil"]);
    }

    // 点击空白，pdf 搜索、更多消失
    if (hasClosestByAttribute(target, "id", "secondaryToolbarToggleButton") ||
        hasClosestByAttribute(target, "id", "viewFindButton") ||
        hasClosestByAttribute(target, "id", "findbar")) {
        return;
    }
    let currentPDFViewerObject: any;
    getAllModels().asset.find(item => {
        if (item.pdfObject &&
            !item.pdfObject.appConfig.appContainer.classList.contains("fn__none")) {
            currentPDFViewerObject = item.pdfObject;
            return true;
        }
    });
    if (!currentPDFViewerObject) {
        return;
    }
    if (currentPDFViewerObject.secondaryToolbar.isOpen) {
        currentPDFViewerObject.secondaryToolbar.close();
    }
    if (
        !currentPDFViewerObject.supportsIntegratedFind &&
        currentPDFViewerObject.findBar.opened
    ) {
        currentPDFViewerObject.findBar.close();
    }
    /// #endif
};

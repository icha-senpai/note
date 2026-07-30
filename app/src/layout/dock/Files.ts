import {escapeAriaLabel, escapeHtml, escapeLessThans} from "../../util/escape";
import {Tab} from "../Tab";
import {Model} from "../Model";
import {setPanelFocus} from "../util";
import {getDockByType} from "../tabUtil";
import {Constants} from "../../constants";
import {getDocDisplayName, pathPosix, setNoteBook} from "../../util/pathName";
import {newFileInTree} from "../../util/newFile";
import {initFileMenu, initNavigationMenu, sortMenu} from "../../menus/navigation";
import {MenuItem} from "../../menus/Menu";
import {showMessage} from "../../dialog/message";
import {
    getPublishAccessLevel,
    getPublishAccessOptionByLevel,
    openPublishAccessDialog
} from "../../protyle/util/publishAccess";
import {fetchPost, fetchSyncPost} from "../../util/fetch";
import {openEmojiPanel, unicode2Emoji} from "../../emoji";
import {newEncryptedNotebook, newNotebook, openEncryptedNotebook} from "../../util/mount";
import {isNotCtrl, isOnlyMeta, setStorageVal, updateHotkeyAfterTip} from "../../protyle/util/compatibility";
import {openFileById} from "../../editor/util";
import {
    hasClosestByAttribute,
    hasClosestByClassName,
    hasClosestByTag,
    hasTopClosestByTag
} from "../../protyle/util/hasClosest";
import {App} from "../../index";
import {refreshFileTree} from "../../dialog/processSystem";
/// #if !BROWSER
import {ipcRenderer} from "electron";
/// #endif
import {hideTooltip, showTooltip} from "../../dialog/tooltip";
import {selectOpenTab} from "./util";
import {hideDragTip, showDragTip, transparentImgSrc} from "../../protyle/util/dragTip";
import {
    cancelFileTreeCollapse,
    collapseFileTree,
    expandFileTree,
    isFileTreeCollapsing
} from "./fileTreeAnimation";

export class Files extends Model {
    public element: HTMLElement;
    public parent: Tab;
    public closeElement: HTMLElement;
    public lastSelectedElement: Element = null;
    private actionsElement: HTMLElement;
    private reloadNotebookInfoTimeout: number;

    constructor(options: { tab: Tab, app: App }) {
        super({app: options.app});
        this.connect({
            type: "filetree",
            id: options.tab.id,
            msgCallback: this.handleMsgCallback.bind(this)
        });
        options.tab.panelElement.classList.add("fn__flex-column", "file-tree", "sy__file", "dockPanel");
        options.tab.panelElement.innerHTML = `<div class="block__icons">
    <div class="block__logo fn__flex-1">${window.scribli.languages.fileTree}</div>
    <span data-type="focus" class="block__icon ariaLabel" data-position="north" aria-label="${window.scribli.languages.selectOpen1}${updateHotkeyAfterTip(window.scribli.config.keymap.general.selectOpen1.custom)}"><svg><use xlink:href='#iconFocus'></use></svg></span>
    <span class="fn__space"></span>
    <span data-type="collapse" class="block__icon ariaLabel" data-position="north" aria-label="${window.scribli.languages.collapse}${updateHotkeyAfterTip(window.scribli.config.keymap.editor.general.collapse.custom)}">
        <svg><use xlink:href="#iconContract"></use></svg>
    </span>
    <div class="fn__space${window.scribli.config.readonly ? " fn__none" : ""}"></div>
    <div data-type="more" class="ariaLabel block__icon${window.scribli.config.readonly ? " fn__none" : ""}" data-position="north" aria-label="${window.scribli.languages.more}">
        <svg><use xlink:href="#iconMore"></use></svg>
    </div> 
    <span class="fn__space"></span>
    <span data-type="min" class="block__icon ariaLabel" data-position="north" aria-label="${window.scribli.languages.min}${updateHotkeyAfterTip(window.scribli.config.keymap.general.closeTab.custom)}"><svg><use xlink:href='#iconMin'></use></svg></span>
</div>
<div class="fn__flex-1" style="padding-top: 2px;"></div>
<ul class="b3-list fn__flex-column" style="min-height: auto;height:30px;transition: height  .2s cubic-bezier(0, 0, .2, 1) 0ms">
    <li class="b3-list-item" data-type="toggle">
        <span class="b3-list-item__toggle">
            <svg class="b3-list-item__arrow"><use xlink:href="#iconRight"></use></svg>
        </span>
        <span class="b3-list-item__text">${window.scribli.languages.closeNotebook}</span>
        <span class="counter" style="cursor: auto"></span>
    </li>
    <ul class="fn__none fn__flex-1"></ul>
</ul>`;
        this.actionsElement = options.tab.panelElement.firstElementChild as HTMLElement;
        this.element = this.actionsElement.nextElementSibling as HTMLElement;
        this.closeElement = options.tab.panelElement.lastElementChild as HTMLElement;
        this.closeElement.addEventListener("click", (event) => {
            setPanelFocus(this.element.parentElement);
            let target = event.target as HTMLElement;
            while (target && !target.isEqualNode(this.closeElement)) {
                const type = target.getAttribute("data-type");
                if (target.classList.contains("b3-list-item__icon")) {
                    event.preventDefault();
                    event.stopPropagation();
                    const rect = target.getBoundingClientRect();
                    openEmojiPanel(target.parentElement.getAttribute("data-url"), "notebook", {
                        x: rect.left,
                        y: rect.bottom,
                        h: rect.height,
                        w: rect.width,
                    }, undefined, target.querySelector("img"));
                    break;
                } else if (type === "toggle") {
                    const svgElement = target.querySelector("svg");
                    if (svgElement.classList.contains("b3-list-item__arrow--open")) {
                        this.closeElement.style.height = "30px";
                        svgElement.classList.remove("b3-list-item__arrow--open");
                        this.closeElement.lastElementChild.classList.add("fn__none");
                    } else {
                        this.closeElement.style.height = "40%";
                        svgElement.classList.add("b3-list-item__arrow--open");
                        this.closeElement.lastElementChild.classList.remove("fn__none");
                    }
                    window.scribli.menus.menu.remove();
                    event.stopPropagation();
                    event.preventDefault();
                    break;
                } else if (type === "open") {
                    const notebookId = target.getAttribute("data-url");
                    const liElement = target.closest("li");
                    if (liElement && liElement.getAttribute("data-encrypted") === "true") {
                        openEncryptedNotebook(this.app, notebookId, liElement.querySelector(".b3-list-item__text").textContent);
                    } else {
                        fetchPost("/api/notebook/openNotebook", {
                            notebook: notebookId
                        });
                    }
                    window.scribli.menus.menu.remove();
                    event.stopPropagation();
                    event.preventDefault();
                    break;
                }
                target = target.parentElement;
            }
        });
        this.actionsElement.querySelector('[data-type="collapse"]').addEventListener("click", () => {
            Array.from(this.element.children).forEach(item => {
                const liElement = item.firstElementChild;
                const toggleElement = liElement.querySelector(".b3-list-item__arrow");
                if (toggleElement.classList.contains("b3-list-item__arrow--open")) {
                    toggleElement.classList.remove("b3-list-item__arrow--open");
                    liElement.nextElementSibling.remove();
                }
            });
            window.scribli.storage[Constants.LOCAL_FILESPATHS] = [];
            setStorageVal(Constants.LOCAL_FILESPATHS, []);
        });
        this.actionsElement.addEventListener("click", (event: MouseEvent & { target: HTMLElement }) => {
            let target = event.target as HTMLElement;
            let isFocus = true;
            while (target && !target.isEqualNode(this.actionsElement)) {
                const type = target.getAttribute("data-type");
                if (type === "min") {
                    getDockByType("file").toggleModel("file", false, true);
                    event.preventDefault();
                    event.stopPropagation();
                    window.scribli.menus.menu.remove();
                    isFocus = false;
                    break;
                } else if (type === "focus") {
                    selectOpenTab();
                    event.preventDefault();
                    break;
                } else if (type === "more") {
                    this.initMoreMenu().popup({x: event.clientX, y: event.clientY});
                    event.preventDefault();
                    event.stopPropagation();
                    break;
                }
                target = target.parentElement;
            }
            if (isFocus) {
                setPanelFocus(this.element.parentElement);
            }
        });
        this.element.addEventListener("mousedown", (event) => {
            if (event.button !== 1 || !window.scribli.config.fileTree.openFilesUseCurrentTab) {
                return;
            }
            let target = event.target as HTMLElement;
            while (target && !target.isEqualNode(this.element)) {
                if (target.tagName === "LI" && target.getAttribute("data-node-id") && !target.getAttribute("data-opening")) {
                    target.setAttribute("data-opening", "true");
                    openFileById({
                        app: options.app,
                        removeCurrentTab: false,
                        id: target.getAttribute("data-node-id"),
                        action: [Constants.CB_GET_FOCUS, Constants.CB_GET_SCROLL],
                        afterOpen() {
                            target.removeAttribute("data-opening");
                        }
                    });
                    event.stopPropagation();
                    event.preventDefault();
                    break;
                }
                target = target.parentElement;
            }
        });
        this.element.addEventListener("click", (event) => {
            let target = event.target as HTMLElement;
            const ulElement = hasTopClosestByTag(target, "UL");
            let needFocus = true;
            if (ulElement) {
                const notebookId = ulElement.getAttribute("data-url");
                while (target && !target.isEqualNode(this.element)) {
                    if (isNotCtrl(event) && target.classList.contains("b3-list-item__icon") && window.scribli.config.system.container !== "ios") {
                        event.preventDefault();
                        event.stopPropagation();
                        const liElement = target.parentElement;
                        const isFile = liElement.getAttribute("data-type") === "navigation-file";
                        const isBoxDoc = liElement.getAttribute("data-type") === "navigation-root" && liElement.getAttribute("data-node-id");
                        if ((isFile || isBoxDoc) && window.scribli.config.fileTree.docIconClickExpand) {
                            if (Number(liElement.getAttribute("data-count")) > 0) {
                                this.getLeaf(liElement, notebookId);
                            } else {
                                needFocus = false;
                                if (!liElement.getAttribute("data-opening")) {
                                    this.lastSelectedElement = liElement;
                                    this.setCurrent(liElement, false);
                                    liElement.setAttribute("data-opening", "true");
                                    openFileById({
                                        app: options.app,
                                        id: liElement.getAttribute("data-node-id"),
                                        action: [Constants.CB_GET_FOCUS, Constants.CB_GET_SCROLL],
                                        afterOpen() {
                                            liElement.removeAttribute("data-opening");
                                        }
                                    });
                                }
                            }
                            break;
                        }
                        const rect = target.getBoundingClientRect();
                        if (isFile) {
                            openEmojiPanel(liElement.getAttribute("data-node-id"), "doc", {
                                x: rect.left,
                                y: rect.bottom,
                                h: rect.height,
                                w: rect.width,
                            }, undefined, target.querySelector("img"));
                        } else {
                            openEmojiPanel(target.parentElement.parentElement.getAttribute("data-url"), "notebook", {
                                x: rect.left,
                                y: rect.bottom,
                                h: rect.height,
                                w: rect.width,
                            }, undefined, target.querySelector("img"));
                        }
                        break;
                    } else if (isNotCtrl(event) && target.classList.contains("b3-list-item__toggle")) {
                        const liElement = target.parentElement;
                        if (liElement.querySelector(".b3-list-item__arrow--open")) {
                            collapseFileTree(liElement, () => this.getOpenPaths());
                        } else if (!isFileTreeCollapsing(liElement)) {
                            this.getLeaf(liElement, notebookId);
                        }
                        event.preventDefault();
                        event.stopPropagation();
                        window.scribli.menus.menu.remove();
                        break;
                    } else if (target.classList.contains("b3-list-item__switch")) {
                        event.preventDefault();
                        event.stopPropagation();
                        const rect = target.getBoundingClientRect();
                        openPublishAccessDialog(target.parentElement.getAttribute("data-type") === "navigation-root" ?
                            notebookId : target.parentElement.getAttribute("data-node-id"), {
                            x: rect.left,
                            y: rect.bottom,
                            h: rect.height,
                            w: rect.width,
                        }, (access) => {
                            target.innerHTML = access.iconHTML;
                            fetchPost("/api/filetree/setPublishAccess", {
                                id: access.id,
                                visible: access.visible,
                                password: access.password,
                                disable: access.disable,
                            });
                        });
                        break;
                    } else if (isNotCtrl(event) && target.classList.contains("b3-list-item__action")) {
                        const type = target.getAttribute("data-type");
                        const pathString = target.parentElement.getAttribute("data-path");
                        if (!window.scribli.config.readonly) {
                            if (type === "new") {
                                newFileInTree(options.app, notebookId, pathString);
                            } else if (type === "more-root") {
                                initNavigationMenu(options.app, target.parentElement).popup({
                                    x: event.clientX,
                                    y: event.clientY
                                });
                            }
                        }
                        if (type === "more-file") {
                            initFileMenu(options.app, notebookId, pathString, target.parentElement).popup({
                                x: event.clientX,
                                y: event.clientY
                            });
                        }
                        event.preventDefault();
                        event.stopPropagation();
                        break;
                    } else if (event.button === 0 && isNotCtrl(event) && !event.altKey && !event.shiftKey &&
                        target.classList.contains("b3-list-item__text") &&
                        (target.parentElement.getAttribute("data-type") === "navigation-file" ||
                            (target.parentElement.getAttribute("data-type") === "navigation-root" && target.parentElement.getAttribute("data-node-id"))) &&
                        window.scribli.config.fileTree.parentDocClickExpand &&
                        Number(target.parentElement.getAttribute("data-count")) > 0) {
                        this.getLeaf(target.parentElement, notebookId);
                        event.preventDefault();
                        event.stopPropagation();
                        window.scribli.menus.menu.remove();
                        break;
                    } else if (target.tagName === "LI") {
                        if (isOnlyMeta(event) && !event.altKey && !event.shiftKey) {
                            target.classList.toggle("b3-list-item--focus");
                            this.lastSelectedElement = target;
                        } else if (event.shiftKey && !event.altKey && isNotCtrl(event)) {
                            if (!document.contains(this.lastSelectedElement)) {
                                this.lastSelectedElement = null;
                            }
                            if (!this.lastSelectedElement) {
                                this.lastSelectedElement = this.element.querySelector(".b3-list-item--focus");
                            }
                            if (!this.lastSelectedElement) {
                                this.lastSelectedElement = target.parentElement.firstElementChild;
                            }
                            this.element.querySelectorAll(".b3-list-item--focus").forEach(item => {
                                item.classList.remove("b3-list-item--focus");
                            });

                            const allFiles = Array.from(this.element.querySelectorAll("li.b3-list-item"));

                            const startIndex = allFiles.indexOf(this.lastSelectedElement);
                            const endIndex = allFiles.indexOf(target);

                            const start = Math.min(startIndex, endIndex);
                            const end = Math.max(startIndex, endIndex);

                            for (let i = start; i <= end; i++) {
                                (allFiles[i] as HTMLElement).classList.add("b3-list-item--focus");
                            }
                        } else {
                            this.lastSelectedElement = target;
                            this.setCurrent(target, false);
                            if (target.getAttribute("data-type") === "navigation-file" ||
                                (target.getAttribute("data-type") === "navigation-root" && target.getAttribute("data-node-id"))) {
                                needFocus = false;
                                if (target.getAttribute("data-opening")) {
                                    return;
                                }
                                target.setAttribute("data-opening", "true");
                                if (event.altKey && isNotCtrl(event) && !event.shiftKey) {
                                    openFileById({
                                        app: options.app,
                                        id: target.getAttribute("data-node-id"),
                                        position: "right",
                                        action: [Constants.CB_GET_FOCUS, Constants.CB_GET_SCROLL],
                                        afterOpen() {
                                            target.removeAttribute("data-opening");
                                        }
                                    });
                                } else if (!event.altKey && isOnlyMeta(event) && event.shiftKey) {
                                    openFileById({
                                        app: options.app,
                                        id: target.getAttribute("data-node-id"),
                                        position: "bottom",
                                        action: [Constants.CB_GET_FOCUS, Constants.CB_GET_SCROLL],
                                        afterOpen() {
                                            target.removeAttribute("data-opening");
                                        }
                                    });
                                } else if (window.scribli.config.fileTree.openFilesUseCurrentTab &&
                                    event.altKey && isOnlyMeta(event) && !event.shiftKey) {
                                    openFileById({
                                        app: options.app,
                                        removeCurrentTab: false,
                                        id: target.getAttribute("data-node-id"),
                                        action: [Constants.CB_GET_FOCUS, Constants.CB_GET_SCROLL],
                                        afterOpen() {
                                            target.removeAttribute("data-opening");
                                        }
                                    });
                                } else {
                                    openFileById({
                                        app: options.app,
                                        id: target.getAttribute("data-node-id"),
                                        action: [Constants.CB_GET_FOCUS, Constants.CB_GET_SCROLL],
                                        afterOpen() {
                                            target.removeAttribute("data-opening");
                                        }
                                    });
                                }
                            } else if (target.getAttribute("data-type") === "navigation-root") {
                                this.getLeaf(target, notebookId);
                            }
                        }
                        this.element.querySelector('[select-end="true"]')?.removeAttribute("select-end");
                        this.element.querySelector('[select-start="true"]')?.removeAttribute("select-start");
                        window.scribli.menus.menu.remove();
                        event.stopPropagation();
                        event.preventDefault();
                        break;
                    }
                    target = target.parentElement;
                }
            }
            if (needFocus) {
                setPanelFocus(this.element.parentElement);
            }
        });
        this.element.addEventListener("dragstart", (event: DragEvent & { target: HTMLElement }) => {
            if (window.scribli.config.readonly) return;
            window.getSelection().removeAllRanges();
            hideTooltip();
            const liElement = hasClosestByTag(event.target, "LI");
            if (liElement) {
                this.parent.panelElement.classList.add("sy__file--disablehover");
                let selectElements: Element[] = Array.from(this.element.querySelectorAll(".b3-list-item--focus"));
                if (!liElement.classList.contains("b3-list-item--focus")) {
                    selectElements.forEach((item) => {
                        item.classList.remove("b3-list-item--focus");
                    });
                    liElement.classList.add("b3-list-item--focus");
                    selectElements = [liElement];
                }
                let ids = "";
                const ghostElement = document.createElement("ul");
                selectElements.forEach((item: HTMLElement, index) => {
                    ghostElement.append(item.cloneNode(true));
                    item.style.opacity = "0.38";
                    const itemNodeId = item.dataset.nodeId ||
                        item.dataset.path;
                    if (itemNodeId) {
                        ids += itemNodeId;
                        if (index < selectElements.length - 1) {
                            ids += ",";
                        }
                    }
                });
                ghostElement.setAttribute("style", `width: 219px;position: fixed;top:-${selectElements.length * 30}px`);
                ghostElement.setAttribute("class", "b3-list b3-list--background");
                document.body.append(ghostElement);
                if (window.scribli.touchDragActive) {
                    event.dataTransfer.setDragImage(ghostElement, 16, 16);
                    window.scribli.touchDragGhost = ghostElement;
                } else {
                    const transparentImg = new Image();
                    transparentImg.src = transparentImgSrc;
                    event.dataTransfer.setDragImage(transparentImg, 0, 0);
                    setTimeout(() => {
                        ghostElement.remove();
                    });
                }
                event.dataTransfer.setData(Constants.SCRIBLI_DROP_FILE, ids);
                event.dataTransfer.dropEffect = "move";
                window.scribli.dragTitle = (selectElements[0] as HTMLElement)?.querySelector(".b3-list-item__text")?.textContent?.trim() || "";
                window.scribli.dragElement = document.createElement("div");
                window.scribli.dragElement.innerText = ids;
            }
        });
        const dragOverLastObj: {
            element: HTMLElement,
            positionY: number,
            rafId: number,
            sourceOnlyRoot: boolean,
        } = {
            element: null,
            positionY: null,
            rafId: null,
            sourceOnlyRoot: null
        };
        this.element.addEventListener("dragend", (event) => {
            if (dragOverLastObj.rafId) {
                cancelAnimationFrame(dragOverLastObj.rafId);
                dragOverLastObj.rafId = null;
            }
            dragOverLastObj.element = null;
            dragOverLastObj.positionY = null;
            dragOverLastObj.sourceOnlyRoot = null;
            this.element.querySelectorAll(".dragover, .dragover__bottom, .dragover__top").forEach((item: HTMLElement) => {
                item.classList.remove("dragover", "dragover__bottom", "dragover__top");
            });
            this.parent.panelElement.classList.remove("sy__file--disablehover");
            this.element.querySelectorAll('.b3-list-item[style*="opacity: 0.38;"]').forEach((item: HTMLElement, index) => {
                item.style.opacity = "";
                // 
                if (index === 0 && hasClosestByClassName(document.elementFromPoint(event.clientX, event.clientY), "sy__file")) {
                    const ariaLabelElement = item.querySelector(".ariaLabel");
                    if (ariaLabelElement) {
                        showTooltip(ariaLabelElement.getAttribute("aria-label"), ariaLabelElement);
                    }
                }
            });
            window.scribli.dragElement = undefined;
            hideDragTip();
            window.scribli.dragTitle = "";
            /// #if !BROWSER
            ipcRenderer.send(Constants.SCRIBLI_SEND_WINDOWS, {cmd: "resetTabsStyle", data: "rmDragStyle"});
            /// #else
            document.querySelectorAll(".layout-tab-bars--drag").forEach(item => {
                item.classList.remove("layout-tab-bars--drag");
            });
            /// #endif
        });
        this.element.addEventListener("dragover", (event: DragEvent & { target: HTMLElement }) => {
            if (window.scribli.config.readonly || !window.scribli.dragElement || event.dataTransfer.types.includes(Constants.SCRIBLI_DROP_TAB)) {
                event.preventDefault();
                return;
            }
            if (dragOverLastObj.rafId) {
                event.preventDefault();
                return;
            }
            let gutterType = "";
            for (const item of event.dataTransfer.items) {
                if (item.type.startsWith(Constants.SCRIBLI_DROP_GUTTER)) {
                    gutterType = item.type;
                }
            }
            if (gutterType) {
                const gutterTypes = gutterType.replace(Constants.SCRIBLI_DROP_GUTTER, "").split(Constants.ZWSP);
                if (!["nodelistitem", "nodeheading"].includes(gutterTypes[0])) {
                    hideDragTip();
                }
            }
            dragOverLastObj.rafId = requestAnimationFrame(() => {
                dragOverLastObj.rafId = null;
                let liElement = event.target.closest("li");
                if (!liElement) {
                    liElement = document.elementFromPoint(event.clientX, event.clientY - 1).closest("li");
                }
                if (!liElement) {
                    dragOverLastObj.element = null;
                    hideDragTip();
                    event.preventDefault();
                    return;
                }
                const targetType = liElement.getAttribute("data-type");
                if (dragOverLastObj.element !== liElement) {
                    dragOverLastObj.element?.classList.remove("dragover", "dragover__bottom", "dragover__top");
                    if (gutterType) {
                        const gutterTypes = gutterType.replace(Constants.SCRIBLI_DROP_GUTTER, "").split(Constants.ZWSP);
                        if (!["nodelistitem", "nodeheading"].includes(gutterTypes[0])) {
                            event.preventDefault();
                            return;
                        }
                    } else if (liElement.classList.contains("b3-list-item--focus")) {
                        hideDragTip();
                        event.preventDefault();
                        return;
                    }

                    dragOverLastObj.sourceOnlyRoot = gutterType ? false : true;
                    if (dragOverLastObj.sourceOnlyRoot) {
                        const focusItems = this.element.querySelectorAll(".b3-list-item--focus");
                        for (let i = 0; i < focusItems.length; i++) {
                            if (focusItems[i].getAttribute("data-type") === "navigation-file") {
                                dragOverLastObj.sourceOnlyRoot = false;
                                break;
                            }
                        }
                    }
                    if (dragOverLastObj.sourceOnlyRoot && targetType !== "navigation-root") {
                        hideDragTip();
                        event.preventDefault();
                        return;
                    }
                }
                if (dragOverLastObj.element && dragOverLastObj.element === liElement && dragOverLastObj.positionY !== event.clientY) {
                    const notebookElement = hasClosestByAttribute(liElement, "data-sortmode", null);
                    if (!notebookElement) {
                        hideDragTip();
                        event.preventDefault();
                        return;
                    }
                    const notebookSort = notebookElement.getAttribute("data-sortmode");
                    if ((dragOverLastObj.sourceOnlyRoot && targetType === "navigation-root" && window.scribli.config.fileTree.sort === 6) ||
                        (!dragOverLastObj.sourceOnlyRoot && targetType !== "navigation-root" &&
                            (notebookSort === "6" || (window.scribli.config.fileTree.sort === 6 && notebookSort === "15")))
                    ) {
                        const nodeRect = liElement.getBoundingClientRect();
                        const dragHeight = nodeRect.height * .2;
                        liElement.classList.remove("dragover__top", "dragover__bottom", "dragover");
                        if (targetType === "navigation-root" && dragOverLastObj.sourceOnlyRoot) {
                            if (event.clientY > nodeRect.top + nodeRect.height / 2) {
                                liElement.classList.add("dragover__bottom");
                            } else {
                                liElement.classList.add("dragover__top");
                            }
                        } else if (event.clientY > nodeRect.bottom - dragHeight) {
                            liElement.classList.add("dragover__bottom");
                        } else if (event.clientY < nodeRect.top + dragHeight) {
                            liElement.classList.add("dragover__top");
                        }
                    }
                    if (liElement.classList.contains("dragover__top") || liElement.classList.contains("dragover__bottom") ||
                        (targetType === "navigation-root" && dragOverLastObj.sourceOnlyRoot)) {
                        // do nothing
                    } else {
                        liElement.classList.add("dragover");
                    }
                }
                if (dragOverLastObj.element !== liElement) {
                    dragOverLastObj.element = liElement;
                }
                dragOverLastObj.positionY = event.clientY;
                if (!gutterType) {
                    const name = liElement.querySelector(".b3-list-item__text")?.textContent || "";
                    const title = window.scribli.dragTitle || "";
                    if (liElement.classList.contains("dragover__top")) {
                        showDragTip(title, window.scribli.languages.dragTipMoveBefore.replace("${x}", name), event.clientX, event.clientY);
                    } else if (liElement.classList.contains("dragover__bottom")) {
                        showDragTip(title, window.scribli.languages.dragTipMoveAfter.replace("${x}", name), event.clientX, event.clientY);
                    } else if (liElement.classList.contains("dragover")) {
                        showDragTip(title, window.scribli.languages.dragTipMoveChild.replace("${x}", name), event.clientX, event.clientY);
                    } else {
                        hideDragTip();
                    }
                } else {
                    const gutterTypes = gutterType.replace(Constants.SCRIBLI_DROP_GUTTER, "").split(Constants.ZWSP);
                    if (["nodelistitem", "nodeheading"].includes(gutterTypes[0])) {
                        const name = liElement.querySelector(".b3-list-item__text")?.textContent || "";
                        const title = window.scribli.dragTitle || "";
                        let action: string;
                        if (liElement.classList.contains("dragover__top")) {
                            action = window.scribli.languages.dragTip2DocBefore.replace("${x}", name);
                        } else if (liElement.classList.contains("dragover__bottom")) {
                            action = window.scribli.languages.dragTip2DocAfter.replace("${x}", name);
                        } else {
                            action = window.scribli.languages.dragTip2DocChild.replace("${x}", name);
                        }
                        showDragTip(title, action, event.clientX, event.clientY);
                    }
                }
                event.preventDefault();
            });
            event.preventDefault();
        });
        let counter = 0;
        this.element.addEventListener("dragleave", () => {
            counter--;
            if (counter === 0) {
                this.element.querySelectorAll(".dragover, .dragover__bottom, .dragover__top").forEach((item: HTMLElement) => {
                    item.classList.remove("dragover", "dragover__bottom", "dragover__top");
                });
                hideDragTip();
            }
        });
        this.element.addEventListener("dragenter", (event) => {
            event.preventDefault();
            counter++;
        });
        this.element.addEventListener("drop", async (event: DragEvent & { target: HTMLElement }) => {
            counter = 0;
            hideDragTip();
            window.scribli.dragTitle = "";
            const newElement = this.element.querySelector(".dragover, .dragover__bottom, .dragover__top");
            if (!newElement) {
                return;
            }
            const newUlElement = hasTopClosestByTag(newElement, "UL");
            if (!newUlElement) {
                return;
            }
            const oldScrollTop = this.element.scrollTop;
            const toURL = newUlElement.getAttribute("data-url");
            const toPath = newElement.getAttribute("data-path");
            let gutterType = "";
            for (const item of event.dataTransfer.items) {
                if (item.type.startsWith(Constants.SCRIBLI_DROP_GUTTER)) {
                    gutterType = item.type;
                }
            }
            if (gutterType) {
                const gutterTypes = gutterType.replace(Constants.SCRIBLI_DROP_GUTTER, "").split(Constants.ZWSP);
                if (["nodelistitem", "nodeheading"].includes(gutterTypes[0])) {
                    const toDocOptions: {
                        targetNoteBook: string;
                        pushMode: number;
                        toTop?: boolean;
                        srcHeadingID?: string;
                        srcListItemID?: string;
                        targetPath?: string;
                        previousPath?: string;
                    } = {
                        targetNoteBook: toURL,
                        pushMode: 0,
                    };
                    if (newElement.classList.contains("dragover")) {
                        toDocOptions.targetPath = toPath;
                    } else if (newElement.classList.contains("dragover__bottom")) {
                        toDocOptions.previousPath = toPath;
                    } else if (newElement.classList.contains("dragover__top")) {
                        if (newElement.previousElementSibling) {
                            toDocOptions.previousPath = newElement.previousElementSibling.getAttribute("data-path");
                        } else {
                            const parentLi = newElement.parentElement.previousElementSibling as HTMLElement;
                            toDocOptions.targetPath = parentLi.getAttribute("data-path");
                            toDocOptions.toTop = true;
                        }
                    }
                    if (gutterTypes[0] === "nodeheading") {
                        toDocOptions.srcHeadingID = gutterTypes[2].split(",")[0];
                        fetchPost("/api/filetree/heading2Doc", toDocOptions);
                    } else {
                        toDocOptions.srcListItemID = gutterTypes[2].split(",")[0];
                        fetchPost("/api/filetree/li2Doc", toDocOptions);
                    }
                }
                newElement.classList.remove("dragover", "dragover__bottom", "dragover__top");
                window.scribli.dragElement = undefined;
                return;
            }
            window.scribli.dragElement = undefined;
            if (!event.dataTransfer.getData(Constants.SCRIBLI_DROP_FILE)) {
                newElement.classList.remove("dragover", "dragover__bottom", "dragover__top");
                return;
            }
            const selectRootElements: HTMLElement[] = [];
            const selectFileElements: HTMLElement[] = [];
            const fromPaths: string[] = [];
            this.element.querySelectorAll(".b3-list-item--focus").forEach((item: HTMLElement) => {
                if (item.getAttribute("data-type") === "navigation-root") {
                    selectRootElements.push(item);
                } else {
                    const dataPath = item.getAttribute("data-path");
                    const isChild = fromPaths.find(itemPath => {
                        if (dataPath.startsWith(itemPath.replace(".sy", ""))) {
                            return true;
                        }
                    });
                    if (!isChild) {
                        if (newElement.getAttribute("data-path").startsWith(item.dataset.path.replace(".sy", ""))) {
                            return;
                        }
                        selectFileElements.push(item);
                        fromPaths.push(dataPath);
                    }
                }
            });
            if (newElement.classList.contains("dragover")) {
                fetchPost("/api/filetree/moveDocs", {
                    toNotebook: toURL,
                    fromPaths,
                    toPath,
                });
                newElement.classList.remove("dragover", "dragover__bottom", "dragover__top");
                return;
            }
            if (newElement.classList.contains("dragover__bottom") || newElement.classList.contains("dragover__top")) {
                const ulSort = newUlElement.getAttribute("data-sortmode");
                if (window.scribli.config.fileTree.sort === 6 && selectRootElements.length > 0 &&
                    newElement.getAttribute("data-path") === "/") {
                    if (newElement.classList.contains("dragover__top")) {
                        selectRootElements.forEach(item => {
                            newElement.parentElement.before(item.parentElement);
                        });
                    } else {
                        selectRootElements.reverse().forEach(item => {
                            newElement.parentElement.after(item.parentElement);
                        });
                    }
                    const notebooks: string[] = [];
                    Array.from(this.element.children).forEach(item => {
                        notebooks.push(item.getAttribute("data-url"));
                    });
                    fetchPost("/api/notebook/changeSortNotebook", {
                        notebooks,
                    });
                } else if ((ulSort === "6" || (window.scribli.config.fileTree.sort === 6 && ulSort === "15")) && selectFileElements.length > 0) {
                    let hasMove = false;
                    const toDir = pathPosix().dirname(toPath);
                    const newElementClassList = newElement.getAttribute("class");
                    if (fromPaths.length > 0) {
                        await fetchSyncPost("/api/filetree/moveDocs", {
                            toNotebook: toURL,
                            fromPaths,
                            toPath: toDir === "/" ? "/" : toDir + ".sy",
                            callback: Constants.CB_MOVE_NOLIST,
                        });
                        selectFileElements.forEach(item => {
                            item.setAttribute("data-path", pathPosix().join(toDir, item.getAttribute("data-node-id") + ".sy"));
                        });
                        hasMove = true;
                    }
                    if (newElementClassList.includes("dragover__top")) {
                        selectFileElements.forEach(item => {
                            let nextULElement;
                            if (item.nextElementSibling && item.nextElementSibling.tagName === "UL") {
                                nextULElement = item.nextElementSibling;
                            }
                            newElement.before(item);
                            if (nextULElement) {
                                item.after(nextULElement);
                            }
                        });
                    } else if (newElementClassList.includes("dragover__bottom")) {
                        selectFileElements.reverse().forEach(item => {
                            let nextULElement;
                            if (item.nextElementSibling && item.nextElementSibling.tagName === "UL") {
                                nextULElement = item.nextElementSibling;
                            }
                            if (newElement.nextElementSibling && newElement.nextElementSibling.tagName === "UL") {
                                newElement.nextElementSibling.after(item);
                            } else {
                                newElement.after(item);
                            }
                            if (nextULElement) {
                                item.after(nextULElement);
                            }
                        });
                    }
                    const paths: string[] = [];
                    Array.from(newElement.parentElement.children).forEach(item => {
                        if (item.tagName === "LI") {
                            paths.push(item.getAttribute("data-path"));
                        }
                    });
                    fetchPost("/api/filetree/changeSort", {
                        paths,
                        notebook: toURL
                    }, () => {
                        if (hasMove) {
                            fetchPost("/api/filetree/listDocsByPath", {
                                notebook: toURL,
                                path: toDir === "/" ? "/" : toDir + ".sy",
                                app: Constants.SCRIBLI_APPID,
                            }, response => {
                                if (response.data.path === "/" && response.data.files.length === 0) {
                                    showMessage(window.scribli.languages.emptyContent);
                                    return;
                                }
                                this.onLsHTML(response.data, oldScrollTop);
                            });
                        }
                    });
                }
            }
            newElement.classList.remove("dragover", "dragover__bottom", "dragover__top");
        });
        this.init();
    }

    private handleMsgCallback(data: IWebSocketData) {
        if (data) {
            switch (data.cmd) {
                case "reloadDocInfo":
                    this.updateDocInfo(data);
                    break;
                case "moveDoc":
                    this.onMove(data);
                    break;
                case "reloadFiletree":
                    setNoteBook(() => {
                        this.init(false);
                    });
                    break;
                case "reloadNotebookInfo":
                    window.clearTimeout(this.reloadNotebookInfoTimeout);
                    this.reloadNotebookInfoTimeout = window.setTimeout(() => {
                        setNoteBook((notebooks) => {
                            notebooks.forEach((notebook) => {
                                const liElement = this.element.querySelector<HTMLElement>(
                                    `ul[data-url="${notebook.id}"] > li[data-type="navigation-root"]`
                                );
                                if (liElement) {
                                    this.updateSubFileCount(liElement, notebook.subFileCount);
                                }
                            });
                        });
                    }, 128);
                    break;
                case "mount":
                    this.onMount(data);
                    this.app.plugins.forEach((item) => {
                        item.eventBus.emit("opened-notebook", data);
                    });
                    break;
                case "createnotebook":
                    setNoteBook((notebooks) => {
                        let previousId: string;
                        notebooks.find(item => {
                            if (!item.closed) {
                                if (item.id === data.data.box.id) {
                                    if (previousId) {
                                        this.element.querySelector(`.b3-list[data-url="${previousId}"]`).insertAdjacentHTML("afterend", this.genNotebook(data.data.box));
                                    } else {
                                        this.element.insertAdjacentHTML("afterbegin", this.genNotebook(data.data.box));
                                    }
                                    return true;
                                }
                                previousId = item.id;
                            }
                        });
                    });
                    break;
                case "closeBox":
                case "removeBox":
                    this.onRemove(data);
                    this.app.plugins.forEach((item) => {
                        item.eventBus.emit("closed-notebook", data);
                    });
                    break;
                case "removeDoc":
                    this.onRemove(data);
                    break;
                case "create":
                    if (data.data.listDocTree) {
                        this.selectItem(data.data.box.id, data.data.path);
                    } else {
                        this.updateItemArrow(data.data.box.id, data.data.path);
                    }
                    break;
                case "createdailynote":
                case "heading2doc":
                case "li2doc":
                    this.selectItem(data.data.box.id, data.data.path);
                    break;
                case "renamenotebook": {
                    const notebook = window.scribli.notebooks.find((item) => item.id === data.data.box);
                    if (notebook) {
                        notebook.name = data.data.name;
                    }
                    this.element.querySelector(`[data-url="${data.data.box}"] .b3-list-item__text`).innerHTML = escapeHtml(data.data.name);
                    break;
                }
                case "rename":
                    this.onRename(data.data);
                    break;
            }
        }
    }

    private updateDocInfo(data: IWebSocketData) {
        const notebook = window.scribli.notebooks.find((item) => item.id === data.data.rootID);
        const subFileCount = notebook && window.scribli.isPublish ? notebook.subFileCount : data.data.subFileCount;
        if (notebook) {
            notebook.subFileCount = subFileCount;
        }
        const liElement = this.element.querySelector(
            `li[data-node-id="${data.data.rootID}"][data-type="navigation-file"], ` +
            `li[data-node-id="${data.data.rootID}"][data-type="navigation-root"]`
        );
        if (liElement) {
            if (liElement.getAttribute("data-type") === "navigation-file") {
                liElement.querySelector(".b3-list-item__text.ariaLabel")?.setAttribute("aria-label", this.genDocAriaLabel(data.data, escapeLessThans));
            }
            this.updateSubFileCount(liElement as HTMLElement, subFileCount);
        }
    }

    private updateSubFileCount(liElement: HTMLElement, subFileCount: number) {
        liElement.setAttribute("data-count", subFileCount.toString());
        if (subFileCount === 0) {
            liElement.querySelector(".b3-list-item__toggle")?.classList.add("fn__hidden");
            liElement.querySelector(".b3-list-item__arrow")?.classList.remove("b3-list-item__arrow--open");
            if (liElement.nextElementSibling?.tagName === "UL") {
                liElement.nextElementSibling.remove();
            }
        } else {
            liElement.querySelector(".b3-list-item__toggle")?.classList.remove("fn__hidden");
        }
        this.updateDocActionElement(liElement);
    }

    private updateDocActionElement(liElement: HTMLElement) {
        const iconElement = liElement.querySelector<HTMLElement>(".b3-list-item__icon");
        if (!iconElement) {
            return;
        }
        const isFile = liElement.getAttribute("data-type") === "navigation-file";
        const isBoxDoc = liElement.getAttribute("data-type") === "navigation-root" &&
            Boolean(liElement.getAttribute("data-node-id"));
        const hasChildren = (isFile || isBoxDoc) && Number(liElement.getAttribute("data-count")) > 0;
        const iconUsesDocAction = window.scribli.config.fileTree.docIconClickExpand && (isFile || isBoxDoc);
        const editingPublishAccess = this.element.classList.contains("file-tree__publish-access--active");
        iconElement.setAttribute("aria-label", iconUsesDocAction ?
            (hasChildren ? window.scribli.languages.docIconClickExpand : window.scribli.languages.openDocument) :
            window.scribli.languages.changeIcon);
        liElement.classList.toggle("file-tree__item--icon-expand", hasChildren && iconUsesDocAction &&
            !editingPublishAccess);
        liElement.classList.toggle("file-tree__item--icon-open", (isFile || isBoxDoc) && !hasChildren && iconUsesDocAction &&
            !editingPublishAccess);
        liElement.classList.toggle("file-tree__item--title-expand", hasChildren &&
            window.scribli.config.fileTree.parentDocClickExpand);
    }

    public updateDocActions() {
        this.element.querySelectorAll<HTMLElement>(
            'li[data-type="navigation-file"], li[data-type="navigation-root"]'
        ).forEach((item) => {
            this.updateDocActionElement(item);
        });
    }

    private updateItemArrow(notebookId: string, filePath: string) {
        const treeElement = this.element.querySelector(`[data-url="${notebookId}"]`);
        if (!treeElement) {
            return;
        }
        let currentPath = filePath;
        let liElement;
        while (!liElement) {
            liElement = treeElement.querySelector(`[data-path="${currentPath}"]`);
            if (!liElement) {
                const dirname = pathPosix().dirname(currentPath);
                if (dirname === "/") {
                    const rootElement = treeElement.firstElementChild as HTMLElement;
                    if (rootElement.querySelector(".b3-list-item__arrow--open")) {
                        this.getLeaf(rootElement, notebookId, true);
                    }
                    break;
                } else {
                    currentPath = dirname + ".sy";
                }
            } else {
                const hiddenElement = liElement.querySelector(".fn__hidden");
                if (hiddenElement) {
                    hiddenElement.classList.remove("fn__hidden");
                } else if (liElement.querySelector(".b3-list-item__arrow--open")) {
                    this.getLeaf(liElement, notebookId, true);
                }
                break;
            }
        }
    }

    private genNotebook(item: INotebook) {
        const editingPublishAccess = this.element.classList.contains("file-tree__publish-access--active");
        const iconContent = (item.encrypted && item.closed)
            ? "🔒️"
            : unicode2Emoji(item.icon || window.scribli.storage[Constants.LOCAL_IMAGES].note);
        const isBoxDoc = !item.closed && window.scribli.config.fileTree.boxDocEnabled;
        const hasChildren = isBoxDoc && item.subFileCount > 0;
        const iconUsesDocAction = isBoxDoc && window.scribli.config.fileTree.docIconClickExpand;
        const iconAriaLabel = iconUsesDocAction ?
            (hasChildren ? window.scribli.languages.docIconClickExpand : window.scribli.languages.openDocument) :
            window.scribli.languages.changeIcon;
        const actionClasses = `${iconUsesDocAction && hasChildren && !editingPublishAccess ? " file-tree__item--icon-expand" : ""}${
            iconUsesDocAction && !hasChildren && !editingPublishAccess ? " file-tree__item--icon-open" : ""}${
            hasChildren && window.scribli.config.fileTree.parentDocClickExpand ? " file-tree__item--title-expand" : ""}`;
        const emojiHTML = `<span class="b3-list-item__icon ariaLabel${isBoxDoc ? " popover__block" : ""}${editingPublishAccess ? " fn__none" : ""}" data-position="8east"${isBoxDoc ? ` data-id="${item.id}"` : ""} aria-label="${iconAriaLabel}">${iconContent}</span>`;
        const switchHTML = `<span class="b3-list-item__switch b3-tooltips b3-tooltips__e${editingPublishAccess ? "" : " fn__none"}" aria-label="${window.scribli.languages.publishAccess}">${getPublishAccessOptionByLevel("public").iconHTML}</span>`;
        if (item.closed) {
            return `<li data-url="${item.id}" class="b3-list-item b3-list-item--hide-action"${item.encrypted ? ' data-encrypted="true"' : ""}>
    <span class="b3-list-item__toggle fn__hidden">
        <svg class="b3-list-item__arrow"><use xlink:href="#iconRight"></use></svg>
    </span>
    ${emojiHTML}
    ${switchHTML}
    <span class="b3-list-item__text" style="cursor: default;">${escapeHtml(item.name)}</span>
    <span data-type="open" data-url="${item.id}" class="b3-list-item__action b3-tooltips b3-tooltips__w${(window.scribli.config.readonly) ? " fn__none" : ""}" aria-label="${window.scribli.languages.openBy}">
        <svg><use xlink:href="#iconOpen"></use></svg>
    </span>
</li>`;
        } else {
            return `<ul class="b3-list b3-list--background" data-url="${item.id}" data-sort="${item.sort}" data-sortmode="${item.sortMode}">
<li class="b3-list-item b3-list-item--hide-action${actionClasses}" ${window.scribli.config.fileTree.sort === 6 ? 'draggable="true"' : ""}
style="--file-toggle-width:22px;--file-action-offset:22px"
data-type="navigation-root" data-path="/" data-count="${item.subFileCount || 0}" data-node-id="${window.scribli.config.fileTree.boxDocEnabled ? item.id : ""}">
    <span class="b3-list-item__toggle b3-list-item__toggle--hl${isBoxDoc && !hasChildren ? " fn__hidden" : ""}">
        <svg class="b3-list-item__arrow"><use xlink:href="#iconRight"></use></svg>
    </span>
    ${emojiHTML}
    ${switchHTML}
    <span class="b3-list-item__text ariaLabel" data-position="parentE">${escapeHtml(item.name)}</span>
    <span data-type="more-root" class="b3-list-item__action b3-tooltips b3-tooltips__w${(window.scribli.config.readonly) ? " fn__none" : ""}" aria-label="${window.scribli.languages.more}">
        <svg><use xlink:href="#iconMore"></use></svg>
    </span>
    <span data-type="new" class="b3-list-item__action b3-tooltips b3-tooltips__w${(window.scribli.config.readonly) ? " fn__none" : ""}" aria-label="${window.scribli.languages.newSubDoc}">
        <svg><use xlink:href="#iconAdd"></use></svg>
    </span>
</li></ul>`;
        }
    }

    public init(init = true) {
        let html = "";
        let closeHtml = "";
        let closeCounter = 0;
        const scrollTop = this.element.scrollTop;
        window.scribli.notebooks.forEach((item) => {
            if (item.closed) {
                closeCounter++;
                closeHtml += this.genNotebook(item);
            } else {
                html += this.genNotebook(item);
            }
        });
        this.element.innerHTML = html;
        this.closeElement.lastElementChild.innerHTML = closeHtml;
        const counterElement = this.closeElement.querySelector(".counter");
        counterElement.textContent = closeCounter.toString();
        if (closeCounter) {
            this.closeElement.classList.remove("fn__none");
        } else {
            this.closeElement.classList.add("fn__none");
        }
        window.scribli.storage[Constants.LOCAL_FILESPATHS].forEach(async (item: IFilesPath) => {
            for (const openPath of item.openPaths) {
                await this.selectItem(item.notebookId, openPath, undefined, false, false);
            }
            this.element.scrollTop = scrollTop;
        });
        this.refreshPublishAccessSwitch();
        if (!init) {
            return;
        }
        const svgElement = this.closeElement.querySelector("svg");
        if (html !== "") {
            this.closeElement.style.height = "30px";
            svgElement.classList.remove("b3-list-item__arrow--open");
            this.closeElement.lastElementChild.classList.add("fn__none");
        } else {
            this.closeElement.style.height = "40%";
            svgElement.classList.add("b3-list-item__arrow--open");
            this.closeElement.lastElementChild.classList.remove("fn__none");
        }
    }

    private onRemove(data: IWebSocketData) {
        if (data.cmd === "closeBox" || data.cmd === "removeBox") {
            setNoteBook((notebooks) => {
                const targetElement = this.element.querySelector(`ul[data-url="${data.data.box}"] li[data-path="${"/"}"]`);
                if (targetElement) {
                    targetElement.parentElement.remove();
                    if (data.cmd === "closeBox") {
                        let closeHTML = "";
                        notebooks.find(item => {
                            if (item.closed) {
                                closeHTML += this.genNotebook(item);
                            }
                        });
                        this.closeElement.lastElementChild.innerHTML = closeHTML;
                        const counterElement = this.closeElement.querySelector(".counter");
                        counterElement.textContent = (parseInt(counterElement.textContent) + 1).toString();
                        this.closeElement.classList.remove("fn__none");
                    }
                }
            });
            if (data.cmd === "removeBox") {
                const removeElement = this.closeElement.querySelector(`li[data-url="${data.data.box}"]`);
                if (removeElement) {
                    removeElement.remove();
                    const counterElement = this.closeElement.querySelector(".counter");
                    counterElement.textContent = (parseInt(counterElement.textContent) - 1).toString();
                    if (counterElement.textContent === "0") {
                        this.closeElement.classList.add("fn__none");
                    }
                }
            }
            return;
        }
        data.data.ids.forEach((item: string) => {
            const targetElement = this.element.querySelector(`li.b3-list-item[data-node-id="${item}"]`);
            if (targetElement) {
                if (targetElement.nextElementSibling?.tagName === "UL") {
                    targetElement.nextElementSibling.remove();
                }
                const parentElement = targetElement.parentElement.previousElementSibling as HTMLElement;
                if (targetElement.parentElement.childElementCount === 1) {
                    if (parentElement) {
                        const iconElement = parentElement.querySelector("svg");
                        iconElement.classList.remove("b3-list-item__arrow--open");
                        if (parentElement.dataset.type !== "navigation-root" || parentElement.dataset.nodeId) {
                            iconElement.parentElement.classList.add("fn__hidden");
                        }
                        parentElement.setAttribute("data-count", "0");
                        this.updateDocActionElement(parentElement);
                        const emojiElement = iconElement.parentElement.nextElementSibling;
                        if (emojiElement.innerHTML === unicode2Emoji(window.scribli.storage[Constants.LOCAL_IMAGES].folder)) {
                            emojiElement.innerHTML = unicode2Emoji(window.scribli.storage[Constants.LOCAL_IMAGES].file);
                        }
                    }
                    targetElement.parentElement.remove();
                } else {
                    targetElement.remove();
                }
            }
        });
    }

    private onMount(data: IWebSocketData) {
        if (data.data.existed) {
            return;
        }
        const liElement = this.closeElement.querySelector(`li[data-url="${data.data.box.id}"]`) as HTMLElement;
        if (liElement) {
            const counterElement = this.closeElement.querySelector(".counter");
            counterElement.textContent = (parseInt(counterElement.textContent) - 1).toString();
            if (counterElement.textContent === "0") {
                this.closeElement.classList.add("fn__none");
            }
            liElement.remove();
        }
        setNoteBook((notebooks: INotebook[]) => {
            const notebook = notebooks.find((item) => item.id === data.data.box.id) || data.data.box;
            const html = this.genNotebook(notebook);
            if (this.element.childElementCount === 0) {
                this.element.innerHTML = html;
            } else {
                let previousId;
                notebooks.find((item, index) => {
                    if (item.id === data.data.box.id) {
                        while (index > 0) {
                            if (!notebooks[index - 1].closed) {
                                previousId = notebooks[index - 1].id;
                                break;
                            } else {
                                index--;
                            }
                        }
                        return true;
                    }
                });
                if (previousId) {
                    this.element.querySelector(`[data-url="${previousId}"]`).insertAdjacentHTML("afterend", html);
                } else {
                    this.element.insertAdjacentHTML("afterbegin", html);
                }
            }
        });
    }

    public onRename(data: { path: string, title: string, box: string }) {
        const fileItemElement = this.element.querySelector(`ul[data-url="${data.box}"] li[data-path="${data.path}"]`);
        if (!fileItemElement) {
            return;
        }
        fileItemElement.setAttribute("data-name", data.title);
        fileItemElement.querySelector(".b3-list-item__text").innerHTML = escapeHtml(data.title);
    }

    private onMove(response: IWebSocketData) {
        const sourceElement = this.element.querySelector(`ul[data-url="${response.data.fromNotebook}"] li[data-path="${response.data.fromPath}"]`) as HTMLElement;
        if (sourceElement) {
            if (sourceElement.nextElementSibling && sourceElement.nextElementSibling.tagName === "UL") {
                sourceElement.nextElementSibling.remove();
            }
            if (sourceElement.parentElement.childElementCount === 1) {
                if (sourceElement.parentElement.previousElementSibling) {
                    const parentLiElement = sourceElement.parentElement.previousElementSibling as HTMLElement;
                    if (parentLiElement.getAttribute("data-type") !== "navigation-root" || parentLiElement.dataset.nodeId) {
                        parentLiElement.querySelector(".b3-list-item__toggle").classList.add("fn__hidden");
                    }
                    parentLiElement.querySelector(".b3-list-item__arrow").classList.remove("b3-list-item__arrow--open");
                    parentLiElement.setAttribute("data-count", "0");
                    this.updateDocActionElement(parentLiElement);
                    const emojiElement = parentLiElement.querySelector(".b3-list-item__icon");
                    if (emojiElement.innerHTML === unicode2Emoji(window.scribli.storage[Constants.LOCAL_IMAGES].folder)) {
                        emojiElement.innerHTML = unicode2Emoji(window.scribli.storage[Constants.LOCAL_IMAGES].file);
                    }
                }
                sourceElement.parentElement.remove();
            } else {
                sourceElement.remove();
            }
        } else {
            const parentElement = this.element.querySelector(`ul[data-url="${response.data.fromNotebook}"] li[data-path="${pathPosix().dirname(response.data.fromPath)}.sy"]`) as HTMLElement;
            if (parentElement && parentElement.getAttribute("data-count") === "1") {
                parentElement.querySelector(".b3-list-item__toggle").classList.add("fn__hidden");
                parentElement.querySelector(".b3-list-item__arrow").classList.remove("b3-list-item__arrow--open");
            }
        }
        const newElement = this.element.querySelector(`[data-url="${response.data.toNotebook}"] li[data-path="${response.data.toPath}"]`) as HTMLElement;
        if (newElement) {
            newElement.querySelector(".b3-list-item__toggle").classList.remove("fn__hidden");
            if (newElement.getAttribute("data-type") === "navigation-root") {
                newElement.setAttribute("data-count", Math.max(1, Number(newElement.getAttribute("data-count"))).toString());
                this.updateDocActionElement(newElement);
            }
            const emojiElement = newElement.querySelector(".b3-list-item__icon");
            if (emojiElement.innerHTML === unicode2Emoji(window.scribli.storage[Constants.LOCAL_IMAGES].file)) {
                emojiElement.innerHTML = unicode2Emoji(window.scribli.storage[Constants.LOCAL_IMAGES].folder);
            }
            const arrowElement = newElement.querySelector(".b3-list-item__arrow");
            if (arrowElement.classList.contains("b3-list-item__arrow--open") && response.callback !== Constants.CB_MOVE_NOLIST) {
                this.getLeaf(newElement, response.data.toNotebook, true);
            }
        }
    }

    private onLsHTML(data: { files: IFile[], box: string, path: string }, scrollTop?: number) {
        if (data.files.length === 0) {
            return;
        }
        const liElement = this.element.querySelector(`ul[data-url="${data.box}"] li[data-path="${data.path}"]`);
        if (!liElement) {
            return;
        }
        let fileHTML = "";
        data.files.forEach((item: IFile) => {
            fileHTML += this.genFileHTML(item);
        });
        let nextElement = liElement.nextElementSibling;
        if (nextElement && nextElement.tagName === "UL") {
            const tempElement = document.createElement("template");
            tempElement.innerHTML = fileHTML;
            nextElement.querySelectorAll(":scope > .b3-list-item > .b3-list-item__toggle> .b3-list-item__arrow--open").forEach(item => {
                const openLiElement = hasClosestByClassName(item, "b3-list-item");
                if (openLiElement) {
                    const tempOpenLiElement = tempElement.content.querySelector(`.b3-list-item[data-node-id="${openLiElement.getAttribute("data-node-id")}"]`);
                    tempOpenLiElement.after(openLiElement.nextElementSibling);
                    tempOpenLiElement.querySelector(".b3-list-item__arrow").classList.add("b3-list-item__arrow--open");
                }
            });
            nextElement.innerHTML = tempElement.innerHTML;
            if (typeof scrollTop === "number") {
                this.element.scroll({top: scrollTop, behavior: "smooth"});
            }
            this.refreshPublishAccessSwitch();
            return;
        }
        liElement.querySelector(".b3-list-item__arrow").classList.add("b3-list-item__arrow--open");
        liElement.insertAdjacentHTML("afterend", `<ul>${fileHTML}</ul>`);
        nextElement = liElement.nextElementSibling;
        nextElement.setAttribute("style", "top: -1px;position: relative;");
        expandFileTree(nextElement as HTMLElement, () => {
            nextElement.removeAttribute("style");
            if (typeof scrollTop === "number") {
                this.element.scroll({top: scrollTop, behavior: "smooth"});
            }
        });
        this.refreshPublishAccessSwitch();
    }

    private async onLsSelect(data: {
        files: IFile[],
        box: string,
        path: string
    }, filePath: string, setStorage: boolean, isSetCurrent: boolean) {
        let fileHTML = "";
        data.files.forEach((item: IFile) => {
            fileHTML += this.genFileHTML(item);
        });
        if (fileHTML === "") {
            return;
        }
        const liElement = this.element.querySelector(`ul[data-url="${data.box}"] li[data-path="${data.path}"]`);
        if (!liElement) {
            return;
        }
        if (liElement.nextElementSibling && liElement.nextElementSibling.tagName === "UL") {
            liElement.nextElementSibling.remove();
        }
        const arrowElement = liElement.querySelector(".b3-list-item__arrow");
        arrowElement.classList.add("b3-list-item__arrow--open");
        arrowElement.parentElement.classList.remove("fn__hidden");
        const emojiElement = liElement.querySelector(".b3-list-item__icon");
        if (emojiElement.textContent === unicode2Emoji(window.scribli.storage[Constants.LOCAL_IMAGES].file)) {
            emojiElement.textContent = unicode2Emoji(window.scribli.storage[Constants.LOCAL_IMAGES].folder);
        }
        liElement.insertAdjacentHTML("afterend", `<ul>${fileHTML}</ul>`);
        let newLiElement;
        for (let i = 0; i < data.files.length; i++) {
            const item = data.files[i];
            if (filePath === item.path) {
                newLiElement = await this.selectItem(data.box, filePath, undefined, setStorage, isSetCurrent);
            } else if (filePath.startsWith(item.path.replace(".sy", ""))) {
                const response = await fetchSyncPost("/api/filetree/listDocsByPath", {
                    notebook: data.box,
                    path: item.path,
                    app: Constants.SCRIBLI_APPID,
                });
                newLiElement = await this.selectItem(response.data.box, filePath, response.data, setStorage, isSetCurrent);
            }
        }
        if (isSetCurrent) {
            this.setCurrent(newLiElement);
        }
        return newLiElement;
    }

    public setCurrent(target: HTMLElement, isScroll = true) {
        if (!target) {
            return;
        }
        this.element.querySelectorAll("li.b3-list-item--focus").forEach((liItem) => {
            liItem.classList.remove("b3-list-item--focus");
        });
        target.classList.add("b3-list-item--focus");

        if (isScroll) {
            const elementRect = this.element.getBoundingClientRect();
            this.element.scrollTop = this.element.scrollTop + (target.getBoundingClientRect().top - (elementRect.top + elementRect.height / 2));
        }
    }

    public getLeaf(liElement: Element, notebookId: string, focusUpdate = false) {
        const toggleElement = liElement.querySelector(".b3-list-item__arrow");
        if (cancelFileTreeCollapse(liElement)) {
            this.getOpenPaths();
            if (!focusUpdate) {
                return;
            }
        }
        const leafElement = liElement.nextElementSibling as HTMLElement;
        if (toggleElement.classList.contains("b3-list-item__arrow--open") && !focusUpdate) {
            toggleElement.classList.remove("b3-list-item__arrow--open");
            if (leafElement?.tagName === "UL") {
                leafElement.remove();
                this.getOpenPaths();
            } else {
                this.getOpenPaths();
            }
            return;
        }
        fetchPost("/api/filetree/listDocsByPath", {
            notebook: notebookId,
            path: liElement.getAttribute("data-path"),
            app: Constants.SCRIBLI_APPID,
        }, response => {
            if (response.data.path === "/" && response.data.files.length === 0) {
                newFileInTree(this.app, notebookId, "/");
                return;
            }
            this.onLsHTML(response.data);
            this.getOpenPaths();
        });
    }

    public async selectItem(notebookId: string, filePath: string, data?: {
        files: IFile[],
        box: string,
        path: string
    }, setStorage = true, isSetCurrent = true) {
        filePath = filePath.replace(/\/\/+/g, "/");
        const treeElement = this.element.querySelector(`[data-url="${notebookId}"]`);
        if (!treeElement) {
            return;
        }
        const boxDocID = window.scribli.config.fileTree.boxDocEnabled ? notebookId : "";
        if (boxDocID && filePath === `/${boxDocID}.sy`) {
            const boxDocElement = treeElement.querySelector("[data-type=\"navigation-root\"]") as HTMLElement;
            if (isSetCurrent) {
                this.setCurrent(boxDocElement);
            }
            return boxDocElement;
        }
        let currentPath = filePath;
        let liElement: HTMLElement;
        const visitedPaths = new Set<string>();
        while (!liElement) {
            if (visitedPaths.has(currentPath)) {
                return;
            }
            visitedPaths.add(currentPath);
            liElement = treeElement.querySelector(`[data-path="${currentPath}"]`);
            if (!liElement) {
                const dirname = pathPosix().dirname(currentPath);
                if (dirname === "/") {
                    currentPath = dirname;
                } else {
                    currentPath = dirname + ".sy";
                }
            }
        }

        if (liElement.getAttribute("data-path") === filePath) {
            if (setStorage) {
                this.getOpenPaths();
            }
            if (isSetCurrent) {
                this.setCurrent(liElement);
            }
            return liElement;
        }

        if (data && data.path === currentPath) {
            liElement = await this.onLsSelect(data, filePath, setStorage, isSetCurrent);
        } else {
            const response = await fetchSyncPost("/api/filetree/listDocsByPath", {
                notebook: notebookId,
                path: currentPath,
                app: Constants.SCRIBLI_APPID,
            });
            liElement = await this.onLsSelect(response.data, filePath, setStorage, isSetCurrent);
        }
        this.refreshPublishAccessSwitch();
        return liElement;
    }

    private getOpenPaths() {
        const filesPaths: IFilesPath[] = [];
        this.element.querySelectorAll(".b3-list[data-url]").forEach((item: HTMLElement) => {
            const notebookPaths: IFilesPath = {
                notebookId: item.getAttribute("data-url"),
                openPaths: []
            };
            item.querySelectorAll(".b3-list-item__arrow--open").forEach((openItem) => {
                const liElement = hasClosestByTag(openItem, "LI");
                if (liElement) {
                    notebookPaths.openPaths.push(liElement.getAttribute("data-path"));
                }
            });
            if (notebookPaths.openPaths.length > 0) {
                for (let i = 0; i < notebookPaths.openPaths.length; i++) {
                    for (let j = i + 1; j < notebookPaths.openPaths.length; j++) {
                        if (notebookPaths.openPaths[j].startsWith(notebookPaths.openPaths[i].replace(".sy", ""))) {
                            notebookPaths.openPaths.splice(i, 1);
                            j--;
                        }
                    }
                }
                notebookPaths.openPaths.forEach((openPath, index) => {
                    const nextPath = this.element.querySelector(`[data-url="${notebookPaths.notebookId}"] li[data-path="${openPath}"]`)?.nextElementSibling?.firstElementChild?.getAttribute("data-path");
                    if (nextPath) {
                        notebookPaths.openPaths[index] = nextPath;
                    }
                });
                filesPaths.push(notebookPaths);
            }
        });
        window.scribli.storage[Constants.LOCAL_FILESPATHS] = filesPaths;
        setStorageVal(Constants.LOCAL_FILESPATHS, filesPaths);
    }

    private genDocAriaLabel(item: IFile, escapeMethod: (text: string) => string) {
        return `${escapeMethod(getDocDisplayName(item.name, item.titleEmpty))} <small class='ft__on-surface'>${item.hSize}</small>${item.bookmark ? "<br>" + window.scribli.languages.bookmark + " " + escapeMethod(item.bookmark) : ""}${item.name1 ? "<br>" + window.scribli.languages.name + " " + escapeMethod(item.name1) : ""}${item.alias ? "<br>" + window.scribli.languages.alias + " " + escapeMethod(item.alias) : ""}${item.memo ? "<br>" + window.scribli.languages.memo + " " + escapeMethod(item.memo) : ""}${item.subFileCount !== 0 ? window.scribli.languages.includeSubFile.replace("x", item.subFileCount) : ""}<br>${window.scribli.languages.modifiedAt} ${item.hMtime}<br>${window.scribli.languages.createdAt} ${item.hCtime}`;
    }

    private genFileHTML(item: IFile) {
        let countHTML = "";
        if (item.count && item.count > 0) {
            countHTML = `<span class="popover__block counter b3-tooltips b3-tooltips__nw" aria-label="${window.scribli.languages.ref}">${item.count}</span>`;
        }
        const ariaLabel = this.genDocAriaLabel(item, escapeAriaLabel);
        const paddingLeft = (item.path.split("/").length - 1) * 18;
        const editingPublishAccess = this.element.classList.contains("file-tree__publish-access--active");
        const iconExpands = window.scribli.config.fileTree.docIconClickExpand;
        const iconAriaLabel = iconExpands ?
            (item.subFileCount > 0 ? window.scribli.languages.docIconClickExpand : window.scribli.languages.openDocument) :
            window.scribli.languages.changeIcon;
        const actionClasses = `${iconExpands && item.subFileCount > 0 && !editingPublishAccess ? " file-tree__item--icon-expand" : ""}${
            iconExpands && item.subFileCount === 0 && !editingPublishAccess ? " file-tree__item--icon-open" : ""}${
            window.scribli.config.fileTree.parentDocClickExpand && item.subFileCount > 0 ? " file-tree__item--title-expand" : ""}`;
        return `<li data-node-id="${item.id}" data-name="${Lute.EscapeHTMLStr(item.name)}" draggable="true" data-count="${item.subFileCount}" 
data-type="navigation-file" 
style="--file-toggle-width:${paddingLeft + 18}px;--file-action-offset:${paddingLeft + 20}px"
class="b3-list-item b3-list-item--hide-action${actionClasses}" data-path="${item.path}">
    <span style="padding-left: ${paddingLeft}px" class="b3-list-item__toggle b3-list-item__toggle--hl${item.subFileCount === 0 ? " fn__hidden" : ""}">
        <svg class="b3-list-item__arrow"><use xlink:href="#iconRight"></use></svg>
    </span>
    <span class="b3-list-item__icon ariaLabel popover__block${editingPublishAccess ? " fn__none" : ""}" data-position="8east" data-id="${item.id}" aria-label="${iconAriaLabel}">${unicode2Emoji(item.icon || (item.subFileCount === 0 ? window.scribli.storage[Constants.LOCAL_IMAGES].file : window.scribli.storage[Constants.LOCAL_IMAGES].folder))}</span>
    <span class="b3-list-item__switch b3-tooltips b3-tooltips__n${editingPublishAccess ? "" : " fn__none"}" aria-label="${window.scribli.languages.publishAccess}">${getPublishAccessOptionByLevel("public").iconHTML}</span>
    <span class="b3-list-item__text ariaLabel" data-delay="200" data-position="parentE"
aria-label="${ariaLabel}">${getDocDisplayName(item.name, item.titleEmpty, true)}</span>
    <span data-type="more-file" class="b3-list-item__action b3-tooltips b3-tooltips__nw" aria-label="${window.scribli.languages.more}">
        <svg><use xlink:href="#iconMore"></use></svg>
    </span>
    <span data-type="new" class="b3-list-item__action b3-tooltips b3-tooltips__nw${window.scribli.config.readonly ? " fn__none" : ""}" aria-label="${window.scribli.languages.newSubDoc}">
        <svg><use xlink:href="#iconAdd"></use></svg>
    </span>
    ${countHTML}
</li>`;
    }

    private initMoreMenu() {
        window.scribli.menus.menu.remove();
        if (!window.scribli.config.readonly) {
            window.scribli.menus.menu.append(new MenuItem({
                icon: "iconNewNoteBook",
                label: window.scribli.languages.newNotebook,
                click: () => {
                    newNotebook();
                }
            }).element);
            if (window.scribli.config.notebookCrypto?.enabled) {
                window.scribli.menus.menu.append(new MenuItem({
                    icon: "iconLock",
                    label: window.scribli.languages.newEncryptedNotebook,
                    click: () => {
                        newEncryptedNotebook();
                    }
                }).element);
            }
        }
        window.scribli.menus.menu.append(new MenuItem({
            icon: "iconRefresh",
            label: window.scribli.languages.rebuildDataIndex,
            click: () => {
                if (!this.element.getAttribute("disabled")) {
                    this.element.setAttribute("disabled", "disabled");
                    refreshFileTree(() => {
                        this.element.removeAttribute("disabled");
                        this.init(false);
                    });
                }
            }
        }).element);
        if (!window.scribli.config.readonly) {
            const subMenu = sortMenu("notebooks", window.scribli.config.fileTree.sort, (sort: number) => {
                fetchPost("/api/setting/setFiletree", {
                    ...window.scribli.config.fileTree,
                    sort,
                }, (response) => {
                    window.scribli.config.fileTree = response.data;
                    setNoteBook(() => {
                        this.init(false);
                    });
                });
            });
            window.scribli.menus.menu.append(new MenuItem({
                icon: "iconSort",
                label: window.scribli.languages.sort,
                type: "submenu",
                submenu: subMenu,
            }).element);
        }
        if (!window.scribli.config.readonly && window.scribli.config.publish.enable) {
            window.scribli.menus.menu.append(new MenuItem({
                icon: "iconEye",
                label: window.scribli.languages.publishAccess,
                checked: this.element.classList.contains("file-tree__publish-access--active"),
                click: () => {
                    this.element.classList.toggle("file-tree__publish-access--active");
                    const editingPublishAccess = this.element.classList.contains("file-tree__publish-access--active");
                    this.element.querySelectorAll(".b3-list-item__icon").forEach(item => {
                        item.classList.toggle("fn__none", editingPublishAccess);
                        item.nextElementSibling.classList.toggle("fn__none", !editingPublishAccess);
                    });
                    this.updateDocActions();
                    this.refreshPublishAccessSwitch();
                }
            }).element);
        }
        return window.scribli.menus.menu;
    }

    private refreshPublishAccessSwitch() {
        if (window.scribli.config.readonly || window.scribli.isPublish ||
            !this.element.classList.contains("file-tree__publish-access--active")) {
            return;
        }
        const ids: string[] = [];
        this.element.querySelectorAll("[data-url]").forEach((element: HTMLElement) => ids.push(element.getAttribute("data-url")));
        this.element.querySelectorAll("[data-type=\"navigation-file\"][data-node-id]").forEach((element: HTMLElement) => ids.push(element.getAttribute("data-node-id")));
        fetchPost("/api/filetree/getPublishAccess", {
            ids
        }, response => {
            response.data.publishAccess.forEach((item: IPublishAccessItem) => {
                const element = this.element.querySelector(`[data-url="${item.id}"] .b3-list-item__switch`) || this.element.querySelector(`[data-node-id="${item.id}"] .b3-list-item__switch`);
                if (element) {
                    element.innerHTML = getPublishAccessOptionByLevel(getPublishAccessLevel(item.visible, item.password, item.disable)).iconHTML;
                }
            });
        });
    }
}

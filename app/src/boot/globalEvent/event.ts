import {App} from "../../index";
import {windowMouseMove} from "./mousemove";
import {windowKeyUp} from "./keyup";
import {windowKeyDown} from "./keydown";
import {globalClick} from "./click";
import {goBack, goForward} from "../../util/backForward";
import {Constants} from "../../constants";
import {hasClosestByClassName, isInEmbedBlock} from "../../protyle/util/hasClosest";
import {hideTooltip} from "../../dialog/tooltip";
import {hideAllElements} from "../../protyle/ui/hideElements";
import {dragOverScroll, stopScrollAnimation} from "./dragover";
import {cancelManualTouch, initTouchDragBridge, isLastPointerMouse} from "../../util/touchDragBridge";
import {isWindow} from "../../util/functions";
import {getDockByType} from "../../layout/tabUtil";
import {fetchPost} from "../../util/fetch";

export const initWindowEvent = (app: App) => {
    let lastEncryptedNotebookTouch = 0;
    const touchEncryptedNotebooks = () => {
        if (window.scribli.isPublish) {
            return;
        }
        const now = Date.now();
        if (now - lastEncryptedNotebookTouch < 30000) {
            return;
        }
        lastEncryptedNotebookTouch = now;
        fetchPost("/api/notebook/touchEncryptedNotebooks", {});
    };
    window.addEventListener("pointerdown", touchEncryptedNotebooks, {passive: true});
    window.addEventListener("keydown", touchEncryptedNotebooks);
    document.addEventListener("touchstart", touchEncryptedNotebooks, {passive: true});

    document.body.addEventListener("mouseleave", () => {
        if (window.scribli.layout.leftDock) {
            window.scribli.layout.leftDock.hideDock();
            window.scribli.layout.rightDock.hideDock();
            window.scribli.layout.bottomDock.hideDock();
        }
        document.querySelectorAll(".protyle-gutters").forEach(item => {
            item.classList.add("fn__none");
            item.innerHTML = "";
        });
        hideTooltip();
    });
    let mouseIsEnter = false;
    document.body.addEventListener("mouseenter", () => {
        if (window.scribli.layout.leftDock) {
            mouseIsEnter = true;
            setTimeout(() => {
                mouseIsEnter = false;
            }, Constants.TIMEOUT_TRANSITION);
        }
    });

    window.addEventListener("mousemove", (event: MouseEvent) => {
        windowMouseMove(event, mouseIsEnter);
    });

    window.addEventListener("scroll", (event: Event) => {
        const scrollElement = event.target as HTMLElement;
        if (!scrollElement.parentElement || !scrollElement.parentElement.classList.contains("table")) {
            return;
        }
        const resizeElement = scrollElement.parentElement.querySelector(".table__resize") as HTMLElement;
        if (!resizeElement) {
            return;
        }
        const baseLeft = resizeElement.getAttribute("data-left");
        const style = resizeElement.getAttribute("style");
        if (baseLeft === null || !style || style.indexOf("display:block") === -1) {
            return;
        }
        const left = parseInt(baseLeft) - scrollElement.scrollLeft;
        resizeElement.setAttribute("style", style.replace(/left: ?-?\d+px;/, `left: ${Math.round(left)}px;`));
    }, true);

    let scrollTarget: HTMLElement | false;
    window.addEventListener("dragover", (event: DragEvent) => {
        const target = event.target as HTMLElement;
        if (event.dataTransfer.types.includes(Constants.SCRIBLI_DROP_TAB)) {
            if (!hasClosestByClassName(target, "layout-tab-bar")) {
                stopScrollAnimation();
            }
            return;
        }
        if (event.dataTransfer.types.includes("text/plain")) {
            return;
        }
        if (!isWindow() &&
            (!window.scribli.layout.leftDock.pin || !window.scribli.layout.rightDock.pin || !window.scribli.layout.bottomDock.pin)) {
            const fileDock = getDockByType("file");
            if (fileDock && !fileDock.pin &&
                document.querySelector('.dock__items > .dock__item--active[data-type="file"]')) {
                let gutterBlockType = "";
                for (const itemType of event.dataTransfer.types) {
                    if (itemType.startsWith(Constants.SCRIBLI_DROP_GUTTER)) {
                        gutterBlockType = itemType.replace(Constants.SCRIBLI_DROP_GUTTER, "").split(Constants.ZWSP)[0];
                        break;
                    }
                }
                if (["nodeheading", "nodelistitem"].includes(gutterBlockType)) {
                    const statusHeight = document.getElementById("status")?.clientHeight || 0;
                    const toolbarHeight = document.getElementById("toolbar")?.clientHeight || 0;
                    const inYRange = event.clientY > toolbarHeight && event.clientY < window.innerHeight - statusHeight;
                    const dockElement = fileDock.layout.element;
                    let onEdge = false;
                    if (dockElement.classList.contains("layout__dockl")) {
                        onEdge = inYRange &&
                            (fileDock.elements[0].clientWidth > 0 ? event.clientX < Math.max((document.getElementById("dockLeft")?.clientWidth || 0) + 1, 16) : event.clientX < 8);
                    } else if (dockElement.classList.contains("layout__dockr")) {
                        onEdge = inYRange &&
                            (fileDock.elements[0].clientWidth > 0 ? event.clientX > window.innerWidth - Math.max((document.getElementById("dockRight")?.clientWidth || 0) - 2, 16) : event.clientX > window.innerWidth - 8);
                    } else if (dockElement.classList.contains("layout__dockb")) {
                        onEdge = event.clientY > Math.min(window.innerHeight - 10, window.innerHeight - statusHeight);
                    }
                    const rect = dockElement.getBoundingClientRect();
                    if (onEdge ||
                        (event.clientX >= rect.left && event.clientX <= rect.right && event.clientY >= rect.top && event.clientY <= rect.bottom)) {
                        fileDock.showDock();
                    } else {
                        fileDock.hideDock();
                    }
                }
            }
        }
        const fileElement = hasClosestByClassName(target, "sy__file");
        const protyleElement = hasClosestByClassName(target, "protyle", true);
        if (!fileElement && !protyleElement) {
            document.querySelector(".drag-tip")?.remove();
            stopScrollAnimation();
            return;
        }
        if (!scrollTarget) {
            scrollTarget = fileElement || protyleElement;
        }
        if (scrollTarget && protyleElement && (
            scrollTarget.classList.contains("sy__file") || protyleElement !== scrollTarget
        )) {
            scrollTarget = protyleElement;
        } else if (scrollTarget && scrollTarget.classList.contains("protyle") && fileElement) {
            scrollTarget = fileElement;
        }
        if (hasClosestByClassName(target, "layout-tab-container__drag")) {
            stopScrollAnimation();
            return;
        }
        let scrollElement;
        if (scrollTarget && scrollTarget.classList.contains("sy__file")) {
            scrollElement = scrollTarget.firstElementChild.nextElementSibling;
        } else if (scrollTarget && scrollTarget.classList.contains("protyle")) {
            scrollElement = scrollTarget.querySelector(".protyle-content");
        }
        if (scrollTarget && scrollElement) {
            if ((event.dataTransfer.types.includes(Constants.SCRIBLI_DROP_FILE) && hasClosestByClassName(target, "layout-tab-bar")) ||
                (event.dataTransfer.types.includes("Files") && scrollTarget.classList.contains("sy__file")) ||
                (scrollTarget.classList.contains("protyle") && hasClosestByClassName(target, "dockPanel"))) {
                stopScrollAnimation();
            } else {
                dragOverScroll(event, scrollElement.getBoundingClientRect(), scrollElement);
            }
        } else {
            stopScrollAnimation();
        }
    });
    window.addEventListener("dragend", () => {
        stopScrollAnimation();
        document.querySelector(".drag-tip")?.remove();
        window.scribli.dragTitle = "";
    });
    window.addEventListener("dragleave", () => {
        stopScrollAnimation();
    });

    window.addEventListener("mouseup", (event) => {
        if (event.button === 3) {
            event.preventDefault();
            goBack(app);
        } else if (event.button === 4) {
            event.preventDefault();
            goForward(app);
        }
    });

    window.addEventListener("mousedown", (event) => {
        if (!hasClosestByClassName(event.target as Element, "protyle-toolbar")) {
            hideAllElements(["toolbar"]);
        }
    });

    window.addEventListener("keyup", (event) => {
        windowKeyUp(app, event);
    });

    window.addEventListener("keydown", (event) => {
        windowKeyDown(app, event);
    });

    window.addEventListener("blur", () => {
        window.scribli.ctrlIsPressed = false;
        window.scribli.shiftIsPressed = false;
        window.scribli.altIsPressed = false;
        document.body.classList.remove("body--shift-pressed");
        /// #if BROWSER
        /// #endif
    });

    window.addEventListener("click", (event) => {
        globalClick(event);
    });

    let time = 0;
    let startX = 0;
    let startY = 0;
    document.addEventListener("touchstart", (event) => {
        time = Date.now();
        startX = event.touches[0].clientX;
        startY = event.touches[0].clientY;
        // 
        const target = event.target as HTMLElement;
        if (hasClosestByClassName(target, "protyle-icons") ||
            hasClosestByClassName(target, "item") ||
            target.classList.contains("protyle-background__icon")) {
            return;
        }
        const embedBlockElement = isInEmbedBlock(target);
        if (embedBlockElement) {
            embedBlockElement.firstElementChild.classList.toggle("protyle-icons--show");
            return;
        }
    }, false);

    document.addEventListener("touchend", (event) => {
        cancelManualTouch();
        if (window.scribli.touchDragActive) {
            return;
        }
        if (Math.abs(startX - event.changedTouches[0].clientX) < Constants.SIZE_DRAG_THRESHOLD &&
            Math.abs(startY - event.changedTouches[0].clientY) < Constants.SIZE_DRAG_THRESHOLD &&
            Date.now() - time > Constants.TIMEOUT_LONGPRESS &&
            !isLastPointerMouse()) {
            event.target.dispatchEvent(new MouseEvent("contextmenu", {
                bubbles: true,
                cancelable: true,
                clientX: event.changedTouches[0].clientX,
                clientY: event.changedTouches[0].clientY,
            }));
            event.stopImmediatePropagation();
            event.preventDefault();
        }
    });
    initTouchDragBridge();
};

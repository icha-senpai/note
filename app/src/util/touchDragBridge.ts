import {stopScrollAnimation} from "../boot/globalEvent/dragover";
import {Constants} from "../constants";

interface LongPressGate {
    startX: number;
    startY: number;
    touchStartTime: number;
    requireLongPress: boolean;
    longPressCancelled: boolean;
    isMouse: boolean;
}

const shouldYieldToScroll = (gate: LongPressGate, clientX: number, clientY: number): boolean => {
    const dx = clientX - gate.startX;
    const dy = clientY - gate.startY;
    if (Math.abs(dx) < Constants.SIZE_DRAG_THRESHOLD && Math.abs(dy) < Constants.SIZE_DRAG_THRESHOLD) {
        return true;
    }
    if (gate.isMouse) {
        if (gate.requireLongPress) {
            return Date.now() - gate.touchStartTime < Constants.TIMEOUT_MOUSE_DRAG_DELAY;
        }
        return false;
    }
    if (!gate.requireLongPress) {
        return false;
    }
    if (gate.longPressCancelled) {
        return true;
    }
    if (Date.now() - gate.touchStartTime < Constants.TIMEOUT_LONGPRESS) {
        gate.longPressCancelled = true;
        return true;
    }
    return false;
};

interface TouchDragState {
    dataTransfer: DataTransfer | null;
    ghostElement: HTMLElement | null;
    isDragging: boolean;
    draggableElement: HTMLElement;
    editorElement: HTMLElement | null;
}

let dragState: (TouchDragState & LongPressGate) | null = null;
let lastDragOverElement: Element | null = null;

let manualState: (LongPressGate) | null = null;

let lastPointerType: string = "";

const isMouseInput = (touch: Touch): boolean => {
    const hasContactArea = (touch.radiusX ?? 0) > 0 || (touch.radiusY ?? 0) > 0;
    return !hasContactArea && lastPointerType === "mouse";
};

export const isLastPointerMouse = (): boolean => {
    return lastPointerType === "mouse";
};

const handleTouchStart = (e: TouchEvent) => {
    if (dragState || manualState) return;
    if (e.touches.length !== 1) return;

    const target = e.target as HTMLElement;

    if (!target.classList.contains("av__widthdrag")) {
        const draggable = getDraggableAncestor(target);
        if (draggable) {
            const touch = e.touches[0];
            dragState = {
                dataTransfer: null,
                ghostElement: null,
                isDragging: false,
                draggableElement: draggable,
                editorElement: null,
                startX: touch.clientX,
                startY: touch.clientY,
                touchStartTime: Date.now(),
                requireLongPress: draggable.closest(".sy__file") !== null ||
                    draggable.closest(".sy__outline") !== null ||
                    draggable.closest(".av__gallery-item") !== null ||
                    draggable.closest(".layout-tab-bar") !== null ||
                    draggable.closest(".protyle-action") !== null,
                longPressCancelled: false,
                isMouse: isMouseInput(touch),
            };
            return;
        }
    }

    // 
    if (target.tagName === "SELECT" || target.tagName === "OPTION" || target.closest("select")) {
        return;
    }
    if (!target.closest(".dock") &&
        !(target.closest(".b3-dialog") &&  ["resize__move", "resize__rd", "resize__r", "resize__rt",
            "resize__d", "resize__l", "resize__ld", "resize__lt", "resize__t"].some(cls => target.closest("." + cls))) &&
        !target.closest(".sy__outline") &&
        !target.closest(".layout__resize") &&
        !target.closest(".layout__resize--lr") &&
        !target.closest(".layout__dockresize") &&
        !target.closest(".layout__dockresize--lr") &&
        !target.closest(".search__drag") &&
        // Editor-internal resize handles (not native Drag API)
        !target.closest(".av__widthdrag") &&
        !target.closest(".av__drag-fill") &&
        !target.closest(".protyle-action__drag") &&
        !target.closest(".table__resize") &&
        !target.closest(".sb__resize") &&
        !target.closest(".protyle-background__img") &&
        !target.closest(".b3-chip")) return;

    const touch = e.touches[0];
    const mouseEvent = new MouseEvent("mousedown", {
        bubbles: true,
        cancelable: true,
        clientX: touch.clientX,
        clientY: touch.clientY,
        button: 0,
        view: window,
    });
    target.dispatchEvent(mouseEvent);
    manualState = {
        startX: touch.clientX,
        startY: touch.clientY,
        touchStartTime: Date.now(),
        requireLongPress: target.closest(".sy__outline") !== null,
        longPressCancelled: false,
        isMouse: isMouseInput(touch),
    };
};

const handleTouchMove = (e: TouchEvent) => {
    if (dragState) {
        const touch = e.touches[0];
        if (!dragState.isDragging) {
            if (shouldYieldToScroll(dragState, touch.clientX, touch.clientY)) {
                return;
            }
            e.preventDefault();
            startTouchDrag(touch);
            return;
        }
        e.preventDefault();
        continueTouchDrag(touch);
        return;
    }

    if (!manualState) return;
    const touch = e.touches[0];
    if (!document.onmousemove || typeof document.onmousemove !== "function") return;

    if (shouldYieldToScroll(manualState, touch.clientX, touch.clientY)) {
        return;
    }

    e.preventDefault();
    window.scribli.touchDragActive = true;
    const elementUnderFinger = document.elementFromPoint(touch.clientX, touch.clientY);
    if (elementUnderFinger) {
        elementUnderFinger.dispatchEvent(new MouseEvent("mousemove", {
            clientX: touch.clientX,
            clientY: touch.clientY,
            cancelable: true,
            bubbles: true,
        }));
    }
};

const handleTouchEnd = (e: TouchEvent) => {
    if (dragState) {
        if (dragState.isDragging) {
            e.preventDefault();
            endTouchDrag(e.changedTouches[0]);
        }
        cleanupDrag();
        return;
    }
    if (!manualState) return;
    cancelManualTouch();
};

const getDraggableAncestor = (el: HTMLElement): HTMLElement | null => {
    let current: HTMLElement | null = el;
    while (current) {
        if (current.getAttribute?.("draggable") === "true") {
            return current;
        }
        if (current === document.body) break;
        current = current.parentElement;
    }
    return null;
};

const getElementUnderTouch = (clientX: number, clientY: number): Element | null => {
    if (dragState?.ghostElement) {
        dragState.ghostElement.style.display = "none";
    }
    const el = document.elementFromPoint(clientX, clientY);
    if (dragState?.ghostElement) {
        dragState.ghostElement.style.display = "";
    }
    return el;
};

const positionGhost = (clientX: number, clientY: number) => {
    if (dragState?.ghostElement) {
        // Offset ghost so it's visible beside the finger, not hidden under it
        dragState.ghostElement.style.left = `${clientX + 12}px`;
        dragState.ghostElement.style.top = `${clientY + 12}px`;
    }
};

const clearDragoverClasses = () => {
    document.querySelectorAll(".dragover__top, .dragover__bottom, .dragover__left, .dragover__right, .dragover").forEach((item) => {
        item.classList.remove("dragover__top", "dragover__bottom", "dragover__left", "dragover__right", "dragover");
    });
};

const startTouchDrag = (touch: Touch) => {
    const dt = new DataTransfer();
    dragState.dataTransfer = dt;
    dragState.isDragging = true;

    dragState.editorElement = dragState.draggableElement.closest(".protyle-wysiwyg") as HTMLElement;

    window.scribli.touchDragActive = true;
    window.scribli.touchDragGhost = null;

    const dragStartEvent = new DragEvent("dragstart", {
        bubbles: true,
        cancelable: true,
        clientX: touch.clientX,
        clientY: touch.clientY,
        dataTransfer: dt,
        view: window,
    });
    dragState.draggableElement.dispatchEvent(dragStartEvent);

    dragState.ghostElement = window.scribli.touchDragGhost || null;
    if (dragState.ghostElement) {
        dragState.ghostElement.style.pointerEvents = "none";
        dragState.ghostElement.style.zIndex = (++window.scribli.zIndex).toString();
        // Position first, then show — avoids flash at wrong position
        positionGhost(touch.clientX, touch.clientY);
        dragState.ghostElement.style.opacity = "0.6";
    }

    if (dragState.editorElement) {
        const dragEnterEvent = new DragEvent("dragenter", {
            bubbles: false,
            cancelable: true,
            clientX: touch.clientX,
            clientY: touch.clientY,
            dataTransfer: dt,
            view: window,
        });
        dragState.editorElement.dispatchEvent(dragEnterEvent);
    }
};

const continueTouchDrag = (touch: Touch) => {
    if (!dragState.isDragging) return;

    const elementUnderTouch = getElementUnderTouch(touch.clientX, touch.clientY);

    // Track dragenter / dragleave across container-level elements.
    // Only dispatch when element's parent changes, to avoid flickering
    // when moving between siblings of the same parent.
    if (elementUnderTouch !== lastDragOverElement) {
        const prevContainer = lastDragOverElement?.parentElement;
        const currContainer = elementUnderTouch?.parentElement;
        if (prevContainer !== currContainer || (!prevContainer && currContainer) || (prevContainer && !currContainer)) {
            if (lastDragOverElement) {
                const dragLeaveEvent = new DragEvent("dragleave", {
                    bubbles: true,
                    cancelable: true,
                    clientX: touch.clientX,
                    clientY: touch.clientY,
                    dataTransfer: dragState.dataTransfer,
                    view: window,
                });
                lastDragOverElement.dispatchEvent(dragLeaveEvent);
            }
            if (elementUnderTouch) {
                const dragEnterEvent = new DragEvent("dragenter", {
                    bubbles: true,
                    cancelable: true,
                    clientX: touch.clientX,
                    clientY: touch.clientY,
                    dataTransfer: dragState.dataTransfer,
                    view: window,
                });
                elementUnderTouch.dispatchEvent(dragEnterEvent);
            }
        }
        lastDragOverElement = elementUnderTouch;
    }

    if (elementUnderTouch) {
        const dragOverEvent = new DragEvent("dragover", {
            bubbles: true,
            cancelable: true,
            clientX: touch.clientX,
            clientY: touch.clientY,
            dataTransfer: dragState.dataTransfer,
            view: window,
        });
        elementUnderTouch.dispatchEvent(dragOverEvent);
    }

    positionGhost(touch.clientX, touch.clientY);
};

const endTouchDrag = (touch: Touch) => {
    if (!dragState.isDragging) return;

    const elementUnderTouch = getElementUnderTouch(touch.clientX, touch.clientY);
    if (elementUnderTouch) {
        const dropEvent = new DragEvent("drop", {
            bubbles: true,
            cancelable: true,
            clientX: touch.clientX,
            clientY: touch.clientY,
            dataTransfer: dragState.dataTransfer,
            view: window,
        });
        elementUnderTouch.dispatchEvent(dropEvent);
    }

    const dragEndEvent = new DragEvent("dragend", {
        bubbles: true,
        cancelable: true,
        clientX: touch.clientX,
        clientY: touch.clientY,
        dataTransfer: dragState.dataTransfer,
        view: window,
    });
    dragState.draggableElement.dispatchEvent(dragEndEvent);

    clearDragoverClasses();
};

const cleanupDrag = () => {
    stopScrollAnimation();
    clearDragoverClasses();

    if (dragState?.ghostElement) {
        dragState.ghostElement.remove();
    }

    window.scribli.touchDragActive = false;
    window.scribli.touchDragGhost = null;
    dragState = null;
    lastDragOverElement = null;
};

const handleCancel = () => {
    cleanupDrag();
    cancelManualTouch();
};

export const cancelManualTouch = () => {
    if (manualState && document.onmouseup && typeof document.onmouseup === "function") {
        document.onmouseup(new MouseEvent("mouseup", {bubbles: true}));
    }
    manualState = null;
    window.scribli.touchDragActive = false;
};

export const initTouchDragBridge = () => {
    document.addEventListener("pointerdown", (e: PointerEvent) => {
        lastPointerType = e.pointerType;
    }, {passive: true});

    document.addEventListener("touchstart", handleTouchStart, {passive: false});
    document.addEventListener("touchmove", handleTouchMove, {passive: false});
    document.addEventListener("touchend", handleTouchEnd);
    document.addEventListener("touchcancel", handleCancel);
};

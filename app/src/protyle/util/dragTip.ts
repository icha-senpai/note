
export const transparentImgSrc = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=";

const dragTipState = {
    rafId: 0, title: "", action: "", x: 0, y: 0,
    element: null as HTMLElement, titleElement: null as HTMLElement, actionElement: null as HTMLElement,
    lastTitle: "", lastAction: ""
};

const renderDragTip = () => {
    dragTipState.rafId = 0;
    if (!dragTipState.element || !dragTipState.element.isConnected) {
        dragTipState.element = (document.querySelector(".drag-tip") as HTMLElement) || null;
        if (!dragTipState.element) {
            dragTipState.element = document.createElement("div");
            dragTipState.element.className = "tooltip drag-tip";
            dragTipState.element.style.animation = "none";
            dragTipState.element.style.pointerEvents = "none";
            dragTipState.element.style.zIndex = "1000000";
            dragTipState.element.style.fontSize = "14px";
            dragTipState.element.style.lineHeight = "20px";
            dragTipState.element.style.top = "0";
            dragTipState.element.style.left = "0";
            dragTipState.titleElement = document.createElement("div");
            dragTipState.titleElement.className = "drag-tip__title";
            dragTipState.titleElement.style.cssText = "max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--b3-tooltips-color);";
            dragTipState.actionElement = document.createElement("div");
            dragTipState.actionElement.className = "drag-tip__action";
            dragTipState.actionElement.style.cssText = "color:var(--b3-tooltips-second-color);font-size:12px;";
            dragTipState.element.append(dragTipState.titleElement, dragTipState.actionElement);
            document.body.append(dragTipState.element);
        } else {
            dragTipState.titleElement = dragTipState.element.querySelector(".drag-tip__title");
            dragTipState.actionElement = dragTipState.element.querySelector(".drag-tip__action");
        }
        dragTipState.lastTitle = "";
        dragTipState.lastAction = "";
    }
    if (dragTipState.lastTitle !== dragTipState.title) {
        dragTipState.titleElement.textContent = dragTipState.title;
        dragTipState.lastTitle = dragTipState.title;
        dragTipState.titleElement.style.display = dragTipState.title ? "" : "none";
    }
    if (dragTipState.lastAction !== dragTipState.action) {
        dragTipState.actionElement.textContent = dragTipState.action;
        dragTipState.lastAction = dragTipState.action;
    }
    dragTipState.element.style.transform = `translate(${dragTipState.x + 16}px, ${dragTipState.y + 16}px)`;
};

export const showDragTip = (title: string, action: string, x: number, y: number) => {
    dragTipState.title = title;
    dragTipState.action = action;
    dragTipState.x = x;
    dragTipState.y = y;
    if (!dragTipState.rafId) {
        dragTipState.rafId = requestAnimationFrame(renderDragTip);
    }
};

let caretLineElement: HTMLElement | null = null;

export const showCaretLine = (left: number, top: number, height: number) => {
    if (!caretLineElement) {
        caretLineElement = document.createElement("div");
        caretLineElement.style.cssText = "position:fixed;width:2px;background-color:var(--b3-theme-primary-light);z-index:1000000;pointer-events:none;border-radius:var(--b3-border-radius);";
        document.body.append(caretLineElement);
    }
    caretLineElement.style.left = left + "px";
    caretLineElement.style.top = top + "px";
    caretLineElement.style.height = height + "px";
    caretLineElement.style.display = "";
};

export const hideCaretLine = () => {
    caretLineElement?.remove();
    caretLineElement = null;
};

export const hideDragTip = () => {
    if (dragTipState.rafId) {
        cancelAnimationFrame(dragTipState.rafId);
        dragTipState.rafId = 0;
    }
    dragTipState.element?.remove();
    dragTipState.element = null;
    dragTipState.titleElement = null;
    dragTipState.actionElement = null;
    dragTipState.lastTitle = "";
    dragTipState.lastAction = "";
    hideCaretLine();
};

import {getTopBarHeight} from "../layout/getTopBarHeight";

export const setPosition = (element: HTMLElement, left: number, top: number, targetHeight = 0, targetLeft = 0, sticky = false) => {
    element.style.top = top + "px";
    element.style.left = left + "px";
    const rect = element.getBoundingClientRect();
    const topBarHeight = getTopBarHeight();
    if (rect.top < topBarHeight) {
        element.style.top = topBarHeight + "px";
    } else if (rect.bottom > window.innerHeight) {
        const y = top - rect.height - targetHeight;
        if (y > topBarHeight && (y + rect.height) < window.innerHeight) {
            element.style.top = y + "px";
        } else {
            element.style.top = Math.max(topBarHeight, window.innerHeight - rect.height) + "px";
        }
    }

    if (sticky) {
        const lockedBottom = element.dataset.positionBottom;
        const lockedX = element.dataset.positionX;
        const sameAnchor = element.dataset.positionTop === String(top);
        if (sameAnchor && lockedBottom !== undefined) {
            if (top + rect.height <= window.innerHeight) {
                element.style.top = top + "px";
            } else {
                const newTop = parseFloat(lockedBottom) - rect.height;
                element.style.top = (newTop >= topBarHeight ? newTop : topBarHeight) + "px";
            }
        }
        if (sameAnchor && lockedX !== undefined) {
            element.style.left = lockedX + "px";
        }

        if (!(sameAnchor && lockedX !== undefined)) {
            if (rect.right > window.innerWidth) {
                element.style.left = window.innerWidth - rect.width - targetLeft + "px";
            } else if (rect.left < 0) {
                element.style.left = "0";
            }
        }

        element.dataset.positionTop = String(top);
        const actualRect = element.getBoundingClientRect();
        element.dataset.positionBottom = String(actualRect.bottom);
        element.dataset.positionX = String(parseFloat(element.style.left));
    } else {
        if (rect.right > window.innerWidth) {
            element.style.left = window.innerWidth - rect.width - targetLeft + "px";
        } else if (rect.left < 0) {
            element.style.left = "0";
        }
    }
};

import {isMobile} from "../util/functions";

export const showTooltip = (
    message: string,
    target: Element,
    tooltipClass?: string,
    event?: MouseEvent,
    space: number = 0.5,
) => {
    if (isMobile() || !message) {
        return;
    }
    let targetRect = target.getBoundingClientRect();
    const clientRects = Array.from(target.getClientRects());
    if (clientRects.length > 1) {
        if (event) {
            clientRects.forEach(item => {
                if (event.clientY >= item.top - 3 && event.clientY <= item.bottom) {
                    targetRect = item;
                }
            });
        } else {
            let lastWidth = 0;
            clientRects.forEach(item => {
                if (item.width > lastWidth) {
                    targetRect = item;
                }
                lastWidth = item.width;
            });
        }
    }
    if (targetRect.height === 0) {
        hideTooltip();
        return;
    }
    const messageElement = document.getElementById("tooltip");
    messageElement.className = tooltipClass ? `tooltip tooltip--${tooltipClass}` : "tooltip";
    messageElement.innerHTML = window.DOMPurify.sanitize(message);
    messageElement.removeAttribute("style");
    const position = target.getAttribute("data-position");
    const parentRect = target.parentElement.getBoundingClientRect();

    let left;
    let top;
    if (position === "parentE") {
        // parentE: file tree and outline, backlink & viewcard
        top = Math.max(0, parentRect.top - (messageElement.clientHeight - parentRect.height) / 2);
        if (top > window.innerHeight - messageElement.clientHeight) {
            top = window.innerHeight - messageElement.clientHeight;
        }
        left = parentRect.right + 8;
        if (left + messageElement.clientWidth > window.innerWidth) {
            left = parentRect.left - messageElement.clientWidth - 8;
        }
    } else if (position === "parentW") {
        top = Math.max(0, parentRect.top - (messageElement.clientHeight - parentRect.height) / 2);
        if (top > window.innerHeight - messageElement.clientHeight) {
            top = window.innerHeight - messageElement.clientHeight;
        }
        left = parentRect.left - messageElement.clientWidth;
        if (left < 0) {
            left = parentRect.right;
        }
    } else if (position?.endsWith("west")) {
        const positionDiff = parseInt(position) || space;
        top = Math.max(0, targetRect.top - (messageElement.clientHeight - targetRect.height) / 2);
        if (top > window.innerHeight - messageElement.clientHeight) {
            top = window.innerHeight - messageElement.clientHeight;
        }
        left = targetRect.left - messageElement.clientWidth - positionDiff;
        if (left < 0) {
            left = targetRect.right;
        }
    } else if (position?.endsWith("east")) {
        const positionDiff = parseInt(position) || space;
        top = Math.max(0, targetRect.top - (messageElement.clientHeight - targetRect.height) / 2);
        if (top > window.innerHeight - messageElement.clientHeight) {
            top = window.innerHeight - messageElement.clientHeight;
        }
        left = targetRect.right + positionDiff;
        if (left + messageElement.clientWidth > window.innerWidth) {
            left = targetRect.left - messageElement.clientWidth - positionDiff;
        }
    } else if (position?.endsWith("north")) {
        const positionDiff = parseInt(position) || space;
        left = Math.max(0, targetRect.left - (messageElement.clientWidth - targetRect.width) / 2);
        top = targetRect.top - messageElement.clientHeight - positionDiff;
        if (top < 0) {
            if (targetRect.top < window.innerHeight - targetRect.bottom) {
                top = targetRect.bottom + positionDiff;
                messageElement.style.maxHeight = (window.innerHeight - top) + "px";
            } else {
                top = 0;
                messageElement.style.maxHeight = (targetRect.top - positionDiff) + "px";
            }
        }
        if (left + messageElement.clientWidth > window.innerWidth) {
            left = window.innerWidth - messageElement.clientWidth;
        }
    } else {
        const positionDiff = parseInt(position) || space;
        left = Math.max(0, targetRect.left - (messageElement.clientWidth - targetRect.width) / 2);
        top = targetRect.bottom + positionDiff;

        if (top + messageElement.clientHeight > window.innerHeight) {
            if (targetRect.top - positionDiff > window.innerHeight - top) {
                top = Math.max(0, targetRect.top - positionDiff - messageElement.clientHeight);
                messageElement.style.maxHeight = (targetRect.top - positionDiff) + "px";
            } else {
                messageElement.style.maxHeight = (window.innerHeight - top) + "px";
            }
        }
        if (left + messageElement.clientWidth > window.innerWidth) {
            left = window.innerWidth - messageElement.clientWidth;
        }
    }
    messageElement.style.top = top + "px";
    messageElement.style.left = Math.max(0, left) + "px";
    const tooltipDelay = target.getAttribute("data-delay");
    if (tooltipDelay) {
        messageElement.style.animationDelay = tooltipDelay + "ms";
    }
};

export const hideTooltip = () => {
    document.getElementById("tooltip").classList.add("fn__none");
};

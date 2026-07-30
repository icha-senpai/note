import {Constants} from "../../../constants";
import {getRowHTML} from "./row";

const BUFFER_RATIO = 1;

interface IBodyState {
    renderedStart: number;
    renderedEnd: number;
    dataOffset: number;
    view: IAVView;
    topSpacerHeight: number;
    pinIndex?: number;
    rowHeight?: number;
    selectedRowIds?: Set<string>;
}

const dataStore = new Map<string, {
    protyle: IProtyle;
    data: IAV;
}>();
const bodyStates = new WeakMap<HTMLElement, IBodyState>();
const trimPending = new WeakSet<HTMLElement>();
let lastScrollTop: number;

const measureHeightDiff = (el: HTMLElement, mutate: () => void): number => {
    const before = el?.scrollHeight || 0;
    mutate();
    return Math.abs((el?.scrollHeight || 0) - before);
};

const doTrim = (blockElement: HTMLElement, elementRect: DOMRect): void => {
    const viewportHeight = elementRect.bottom - elementRect.top;
    const buffer = viewportHeight * BUFFER_RATIO;
    const topLimit = elementRect.top - buffer;
    const bottomLimit = elementRect.bottom + buffer;
    const blockRect = blockElement.getBoundingClientRect();

    const stored = dataStore.get(blockElement.getAttribute("data-av-id") + blockElement.getAttribute(Constants.CUSTOM_SY_AV_VIEW));
    if (!stored) {
        return;
    }
    const protyle = stored.protyle;
    const isScrollingUp = lastScrollTop && lastScrollTop > protyle.contentElement.scrollTop;
    lastScrollTop = protyle.contentElement.scrollTop;

    if ((blockRect.bottom < elementRect.top && !isScrollingUp) || (blockRect.top > elementRect.bottom && isScrollingUp)) {
        return;
    }

    const type = blockElement.getAttribute("data-av-type") as TAVView;
    const bodies = blockElement.querySelectorAll(".av__body:not(.fn__none)") as NodeListOf<HTMLElement>;
    bodies.forEach((bodyEl: HTMLElement) => {
        const state = bodyStates.get(bodyEl);
        if (!state) {
            return;
        }
        const dataRows = type === "table" ? (state.view as IAVTable).rows : (state.view as IAVKanban).cards;
        const dataStart = state.dataOffset;
        const dataEnd = dataStart + dataRows.length - 1;
        let currentRows;
        let bottomElement;
        if (type === "table") {
            currentRows = bodyEl.querySelectorAll(".av__row:not(.av__row--header):not(.av__row--footer):not(.av__row--util)") as NodeListOf<HTMLElement>;
            bottomElement = bodyEl.querySelector(".av__row--util");
        } else {
            currentRows = bodyEl.querySelectorAll(".av__gallery-item") as NodeListOf<HTMLElement>;
            bottomElement = bodyEl.querySelector(".av__gallery-add");
        }
        if (currentRows.length === 0) {
            return;
        }
        const trimRange = viewportHeight + buffer * 2;
        if (bodyEl.dataset.avLocateWindow !== "true" &&
            dataRows.length <= Math.ceil(trimRange / Math.max(state.rowHeight || currentRows[0].offsetHeight, 1))) {
            const spacerEl = bodyEl.querySelector(".av__spacer") as HTMLElement;
            if (spacerEl) {
                spacerEl.remove();
                state.topSpacerHeight = 0;
                state.renderedStart = dataStart;
                state.renderedEnd = dataEnd;
                bodyStates.set(bodyEl, state);
            }
            return;
        }
        let topElement = currentRows[0];
        if (!topElement.isConnected) {
            return;
        }
        try {
            const spacerElement = bodyEl.querySelector(".av__spacer") as HTMLElement;
        if (!state.selectedRowIds) {
            state.selectedRowIds = new Set();
        }
        const restoreSelect = () => {
            if (state.selectedRowIds.size === 0) {
                return;
            }
            bodyEl.querySelectorAll(type === "table" ? ".av__row[data-id]" : ".av__gallery-item[data-id]").forEach((row: HTMLElement) => {
                if (state.selectedRowIds.has(row.getAttribute("data-id"))) {
                    row.classList.add(type === "table" ? "av__row--select" : "av__gallery-item--select");
                    const use = row.querySelector(".av__firstcol use") as SVGUseElement;
                    if (use) {
                        use.setAttribute("xlink:href", "#iconCheck");
                    }
                }
            });
        };
        let firstVisibleIndex: number;
        let lastVisibleIndex: number;
        const toRemoveAbove: HTMLElement[] = [];
        const toRemoveBelow: HTMLElement[] = [];
        let galleryColumn = type === "table" ? 1 : 0;
        const rowHeight = state.rowHeight || currentRows[0].offsetHeight;
        state.rowHeight = rowHeight;
        const firstTop = currentRows[0].getBoundingClientRect().top;
        if (spacerElement && state.renderedStart > 0) {
            const viewportStartTop = Math.max(elementRect.top, blockRect.top);
            const renderedStartTop = spacerElement.getBoundingClientRect().bottom;
            const rowsPerViewport = Math.ceil(viewportHeight / Math.max(rowHeight, 1));
            if (renderedStartTop - viewportStartTop > rowHeight * rowsPerViewport) {
                currentRows.forEach(row => row.remove());
                spacerElement.remove();
                const newEnd = Math.min(dataStart + rowsPerViewport - 1, dataEnd);
                let rowsHTML = "";
                const viewType = blockElement.getAttribute("data-av-type") as TAVView;
                for (let i = dataStart; i <= newEnd; i++) {
                    rowsHTML += getRowHTML({
                        data: state.view,
                        row: dataRows[i - dataStart],
                        rowIndex: i,
                        pinIndex: state.pinIndex,
                        type: viewType
                    });
                }
                if (bottomElement && bottomElement.isConnected) {
                    bottomElement.insertAdjacentHTML("beforebegin", rowsHTML);
                }
                restoreSelect();
                state.renderedStart = dataStart;
                state.renderedEnd = newEnd;
                state.topSpacerHeight = 0;
                return;
            }
        }
        let foundFirstVisible = false;
        for (let i = 0; i < currentRows.length; i++) {
            const rect = currentRows[i].getBoundingClientRect();
            if (rect.top === firstTop) {
                galleryColumn++;
            }
            if (rect.top > topLimit) {
                if (!foundFirstVisible) {
                    foundFirstVisible = true;
                    firstVisibleIndex = parseInt(currentRows[i].getAttribute("data-index"));
                }
            } else {
                if (!isScrollingUp && toRemoveAbove.length + 10 < currentRows.length) {
                    toRemoveAbove.push(currentRows[i]);
                }
            }
            if (rect.bottom < bottomLimit) {
                lastVisibleIndex = parseInt(currentRows[i].getAttribute("data-index"));
            } else {
                if (isScrollingUp && toRemoveBelow.length + 10 < currentRows.length) {
                    toRemoveBelow.push(currentRows[i]);
                }
                if (type === "table" && !isScrollingUp) {
                    break;
                }
            }
            if (i === currentRows.length - 1 && !isScrollingUp && rect.bottom < bottomLimit) {
                lastVisibleIndex = Math.min(state.renderedEnd + Math.ceil((bottomLimit - rect.bottom) / rowHeight) * galleryColumn, dataEnd);
            }
        }
        if (type === "gallery" && toRemoveAbove.length > 0 && !isScrollingUp) {
            const lastRemoved = toRemoveAbove[toRemoveAbove.length - 1];
            const firstKept = lastRemoved.nextElementSibling as HTMLElement;
            if (firstKept && firstKept.offsetTop === lastRemoved.offsetTop) {
                const incompleteTop = lastRemoved.offsetTop;
                while (toRemoveAbove.length > 0 &&
                    (toRemoveAbove[toRemoveAbove.length - 1] as HTMLElement).offsetTop === incompleteTop) {
                    toRemoveAbove.pop();
                }
            }
        }
        if (isScrollingUp && firstTop > topLimit) {
            firstVisibleIndex = Math.max(dataStart, state.renderedStart - Math.ceil((firstTop - topLimit) / rowHeight) * galleryColumn);
        }
        if (!isScrollingUp) {
            if (toRemoveAbove.length > 0) {
                topElement = toRemoveAbove[toRemoveAbove.length - 1].nextElementSibling as HTMLElement;
                let removeHeight = 0;
                if (type === "gallery") {
                    const galleryEl = bodyEl.querySelector(".av__gallery") as HTMLElement;
                    removeHeight = measureHeightDiff(galleryEl, () => {
                        toRemoveAbove.forEach((row) => {
                            row.remove();
                        });
                    });
                } else if (type === "table" && topElement) {
                    const removeStartTop = toRemoveAbove[0].getBoundingClientRect().top;
                    const removeEndTop = topElement.getBoundingClientRect().top;
                    removeHeight = Math.round(removeEndTop - removeStartTop);
                    toRemoveAbove.forEach((row) => {
                        row.remove();
                    });
                } else { // kanban
                    removeHeight = toRemoveAbove.reduce((sum, row) => sum + row.offsetHeight + 16, 0);
                    toRemoveAbove.forEach((row) => {
                        row.remove();
                    });
                }
                state.topSpacerHeight += removeHeight;
                state.renderedStart = state.renderedStart + toRemoveAbove.length;

                if (spacerElement) {
                    spacerElement.style.height = state.topSpacerHeight + "px";
                } else if (state.topSpacerHeight > 0 && topElement.isConnected) {
                    topElement.insertAdjacentHTML("beforebegin", `<div class="av__spacer" style="height:${state.topSpacerHeight}px"></div>`);
                }
            }

            if (lastVisibleIndex > state.renderedEnd) {
                const rowsPerViewport = Math.ceil(viewportHeight / Math.max(rowHeight, 1));
                const maxRowsPerFrame = rowsPerViewport * (galleryColumn || 1);
                if (lastVisibleIndex > state.renderedEnd + maxRowsPerFrame) {
                    lastVisibleIndex = state.renderedEnd + maxRowsPerFrame;
                }
                let rowsHTML = "";
                const viewType = blockElement.getAttribute("data-av-type") as TAVView;
                for (let i = state.renderedEnd + 1; i <= lastVisibleIndex; i++) {
                    rowsHTML += getRowHTML({
                        data: state.view,
                        row: dataRows[i - dataStart],
                        rowIndex: i,
                        pinIndex: state.pinIndex,
                        type: viewType
                    });
                }
                if (bottomElement && bottomElement.isConnected) {
                    bottomElement.insertAdjacentHTML("beforebegin", rowsHTML);
                }
                restoreSelect();
                state.renderedEnd = lastVisibleIndex;
            }
        } else {
            if (toRemoveBelow.length > 0) {
                toRemoveBelow.forEach(row => row.remove());
                state.renderedEnd = state.renderedEnd - toRemoveBelow.length;
            }

            if (typeof firstVisibleIndex === "number" && firstVisibleIndex < state.renderedStart) {
                let rowsHTML = "";
                const viewType = blockElement.getAttribute("data-av-type") as TAVView;
                for (let i = firstVisibleIndex; i < state.renderedStart; i++) {
                    rowsHTML += getRowHTML({
                        data: state.view,
                        row: dataRows[i - dataStart],
                        rowIndex: i,
                        pinIndex: state.pinIndex,
                        type: viewType
                    });
                }
                if (!topElement.isConnected) {
                    return;
                }
                let renderedHeight = 0;
                if (type === "gallery") {
                    const galleryEl = bodyEl.querySelector(".av__gallery") as HTMLElement;
                    renderedHeight = measureHeightDiff(galleryEl, () => {
                        topElement.insertAdjacentHTML("beforebegin", rowsHTML);
                    });
                } else {
                    topElement.insertAdjacentHTML("beforebegin", rowsHTML);
                    let newRowElement = topElement.previousElementSibling as HTMLElement;
                    while (newRowElement) {
                        if (type === "table") {
                            renderedHeight += newRowElement.offsetHeight;
                        } else { // kanban
                            renderedHeight += newRowElement.offsetHeight + 16;
                        }
                        newRowElement = newRowElement.previousElementSibling as HTMLElement;
                        if (!newRowElement || newRowElement.classList.contains("av__spacer") ||
                            newRowElement.classList.contains("av__row--header")) {
                            break;
                        }
                    }
                }
                state.topSpacerHeight = Math.max(0, state.topSpacerHeight - renderedHeight);
                if (state.topSpacerHeight === 0) {
                    spacerElement?.remove();
                } else if (spacerElement) {
                    spacerElement.style.height = state.topSpacerHeight + "px";
                }
                state.renderedStart = firstVisibleIndex;
                restoreSelect();
            }
        }
        } finally {
            bodyStates.set(bodyEl, state);
        }
    });
};

export const getBodyVirtualData = (bodyEl: HTMLElement, endSelector: string, firstRowIndex: number): IAVVirtualData => {
    const endMarker = bodyEl.querySelector(endSelector);
    let lastRow = endMarker ? endMarker.previousElementSibling as HTMLElement : null;
    while (lastRow && !lastRow.getAttribute("data-index")) {
        lastRow = lastRow.previousElementSibling as HTMLElement;
    }
    let renderedStart = firstRowIndex;
    let renderedEnd = parseInt(lastRow?.getAttribute("data-index") || "");
    const ghostElements = bodyEl.querySelectorAll('[data-type="ghost"]');
    if (ghostElements.length > 0) {
        let prev = (ghostElements[0] as HTMLElement).previousElementSibling as HTMLElement;
        while (prev && prev.getAttribute("data-type") === "ghost") {
            prev = prev.previousElementSibling as HTMLElement;
        }
        const prevIndex = prev?.getAttribute("data-index");
        if (prevIndex) {
            renderedEnd = Math.max(renderedEnd, parseInt(prevIndex) + ghostElements.length);
        } else {
            renderedStart = 0;
            renderedEnd = Math.max(renderedEnd, ghostElements.length - 1);
        }
    }
    return {
        renderedStart,
        renderedEnd,
        topSpacerHeight: bodyEl.querySelector(".av__spacer")?.clientHeight || 0,
    };
};

const getBodyData = (bodyEl: HTMLElement) => {
    const avEl = bodyEl.closest(".av") as HTMLElement;
    if (!avEl) return null;
    const stored = dataStore.get(avEl.getAttribute("data-av-id") + avEl.getAttribute(Constants.CUSTOM_SY_AV_VIEW));
    if (!stored) return null;

    const groupId = bodyEl.dataset.groupId;
    return groupId ? stored.data.view.groups.find((g: IAVView) => g.id === groupId) : stored.data.view;
};

export const getAvBodyData = (bodyEl: HTMLElement): IAVView | null => {
    return getBodyData(bodyEl);
};

export const updateAVRowSelect = (bodyEl: HTMLElement, rowId: string, selected: boolean): void => {
    const state = bodyStates.get(bodyEl);
    if (!state) {
        return;
    }
    if (!state.selectedRowIds) {
        state.selectedRowIds = new Set();
    }
    if (selected) {
        state.selectedRowIds.add(rowId);
    } else {
        state.selectedRowIds.delete(rowId);
    }
};

export const resetAVRowSelect = (bodyEl: HTMLElement, rowIds: string[]): void => {
    const state = bodyStates.get(bodyEl);
    if (!state) {
        return;
    }
    state.selectedRowIds = new Set(rowIds);
};

export const getAVSelectStat = (bodyEl: HTMLElement): { selectCount: number, loadedCount: number } | null => {
    const state = bodyStates.get(bodyEl);
    if (!state || !state.selectedRowIds) {
        return null;
    }
    const dataRows = state.view ? ((state.view as IAVTable).rows || (state.view as IAVKanban).cards || []) : [];
    return {
        selectCount: state.selectedRowIds.size,
        loadedCount: dataRows.length,
    };
};

export const trimAVRows = (blockElement: HTMLElement, elementRect: DOMRect): void => {
    if (blockElement.getAttribute(Constants.ATTRIBUTE_V_SCROLL) !== "true" || trimPending.has(blockElement)) {
        return;
    }
    trimPending.add(blockElement);
    requestAnimationFrame(() => {
        trimPending.delete(blockElement);
        doTrim(blockElement, elementRect);
    });
};

export const trimAVRowsSync = (blockElement: HTMLElement, elementRect: DOMRect): void => {
    if (blockElement.getAttribute(Constants.ATTRIBUTE_V_SCROLL) !== "true") {
        return;
    }
    doTrim(blockElement, elementRect);
};

export const initVirtualScroll = (options: {
    protyle: IProtyle,
    blockElement: HTMLElement,
    data: IAV,
}): void => {
    if (options.blockElement.getAttribute(Constants.ATTRIBUTE_V_SCROLL) !== "true") {
        return;
    }
    dataStore.set(options.blockElement.getAttribute("data-av-id") +
        options.blockElement.getAttribute(Constants.CUSTOM_SY_AV_VIEW), {
        protyle: options.protyle,
        data: options.data,
    });

    options.blockElement.querySelectorAll(".av__body").forEach((item: HTMLElement) => {
        const dataOffset = item.dataset.avLocateWindow === "true" ? options.data.target?.offset || 0 : 0;
        const view = getBodyData(item);
        if (!view) {
            return;
        }
        const selectedRowIds = new Set<string>();
        item.querySelectorAll(options.data.viewType === "table" ? ".av__row--select" : ".av__gallery-item--select").forEach((row: HTMLElement) => {
            const id = row.getAttribute("data-id");
            if (id) {
                selectedRowIds.add(id);
            }
        });
        if (options.data.viewType === "table") {
            const firstRow = item.querySelector(".av__row[data-id]") as HTMLElement;
            let lastRow = item.querySelector(".av__row--util")?.previousElementSibling as HTMLElement;
            while (lastRow && !lastRow.dataset.index) {
                lastRow = lastRow.previousElementSibling as HTMLElement;
            }
            if (!firstRow || !lastRow) {
                return;
            }
            bodyStates.set(item, {
                renderedStart: parseInt(firstRow.dataset.index),
                pinIndex: parseInt(item.querySelector(".av__row--header > .block__icons")?.getAttribute("data-pinindex")),
                renderedEnd: parseInt(lastRow.dataset.index),
                dataOffset,
                view,
                topSpacerHeight: item.querySelector(".av__spacer")?.clientHeight || 0,
                selectedRowIds,
            });
        } else {
            const firstItem = item.querySelector(".av__gallery-item") as HTMLElement;
            let lastItem = item.querySelector(".av__gallery-add")?.previousElementSibling as HTMLElement;
            while (lastItem && !lastItem.dataset.index) {
                lastItem = lastItem.previousElementSibling as HTMLElement;
            }
            if (!firstItem || !lastItem) {
                return;
            }
            bodyStates.set(item, {
                renderedStart: parseInt(firstItem.dataset.index),
                renderedEnd: parseInt(lastItem.dataset.index),
                dataOffset,
                view,
                topSpacerHeight: item.querySelector(".av__spacer")?.clientHeight || 0,
                selectedRowIds,
            });
        }
    });
};

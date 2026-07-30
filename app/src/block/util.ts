import {focusByWbr, getEditorRange, getUndoFocusContext} from "../protyle/util/selection";
import {hasClosestBlock, hasClosestByClassName} from "../protyle/util/hasClosest";
import {
    getContenteditableElement,
    getEmbedChildOperationParentID,
    getParentBlock,
    getPreviousBlockSibling,
    getTopAloneElement
} from "../protyle/wysiwyg/getBlock";
import {genListItemElement, updateListOrder} from "../protyle/wysiwyg/list";
import {transaction, turnsIntoOneTransaction, updateTransaction} from "../protyle/wysiwyg/transaction";
import {scrollCenter} from "../util/highlightById";
import {Constants} from "../constants";
import {hideElements} from "../protyle/ui/hideElements";
import {blockRender} from "../protyle/render/blockRender";
import {fetchPost, fetchSyncPost} from "../util/fetch";
import {openFileById} from "../editor/util";
import {mathRender} from "../protyle/render/mathRender";

export const cancelSB = async (protyle: IProtyle, nodeElement: Element, range?: Range) => {
    const doOperations: IOperation[] = [];
    const undoOperations: IOperation[] = [];
    let previousId = getPreviousBlockSibling(nodeElement)?.getAttribute("data-node-id");
    nodeElement.classList.remove("protyle-wysiwyg--select");
    nodeElement.removeAttribute("select-start");
    nodeElement.removeAttribute("select-end");
    const id = nodeElement.getAttribute("data-node-id");
    nodeElement.querySelectorAll(".sb__resize").forEach(handle => handle.remove());
    const sbElement = nodeElement.cloneNode() as HTMLElement;
    sbElement.innerHTML = nodeElement.lastElementChild.outerHTML;
    let parentID = getEmbedChildOperationParentID(nodeElement) || getParentBlock(nodeElement)?.getAttribute("data-node-id");
    if (!previousId && !parentID) {
        if (protyle.block.showAll || protyle.options.backlinkData) {
            const idData = await fetchSyncPost("/api/block/getBlockSiblingID", {
                id,
                notebook: protyle.notebookId,
            });
            previousId = idData.data.previous;
            parentID = idData.data.parent;
        } else {
            parentID = protyle.block.rootID;
        }
    }
    undoOperations.push({
        action: "insert",
        id,
        data: sbElement.outerHTML,
        previousID: previousId,
        parentID,
    });
    Array.from(nodeElement.children).forEach((item, index) => {
        if (index === nodeElement.childElementCount - 1) {
            doOperations.push({
                action: "delete",
                id,
            });
            if (range) {
                getContenteditableElement(nodeElement)?.insertAdjacentHTML("afterbegin", "<wbr>");
            }
            nodeElement.lastElementChild.remove();
            nodeElement.replaceWith(...nodeElement.children);
            if (range) {
                focusByWbr(protyle.wysiwyg.element, range);
            }
            return;
        }
        doOperations.push({
            action: "move",
            id: item.getAttribute("data-node-id"),
            previousID: previousId,
            parentID,
        });
        undoOperations.push({
            action: "move",
            id: item.getAttribute("data-node-id"),
            previousID: getPreviousBlockSibling(item)?.getAttribute("data-node-id"),
            parentID: id
        });
        previousId = item.getAttribute("data-node-id");
    });
    mathRender(protyle.wysiwyg.element);
    doOperations.forEach(item => {
        const element = protyle.wysiwyg.element.querySelector(`[data-node-id="${item.id}"]`);
        if (element && element.getAttribute("data-type") === "NodeBlockQueryEmbed") {
            element.removeAttribute("data-render");
            blockRender(protyle, element);
        }
    });
    return {
        doOperations, undoOperations, previousId
    };
};

export const genSBElement = (layout: string, id?: string, attrHTML?: string) => {
    const sbElement = document.createElement("div");
    sbElement.setAttribute("data-node-id", id || Lute.NewNodeID());
    sbElement.setAttribute("data-type", "NodeSuperBlock");
    sbElement.setAttribute("class", "sb");
    sbElement.setAttribute("data-sb-layout", layout);
    sbElement.innerHTML = attrHTML || `<div class="protyle-attr" contenteditable="false">${Constants.ZWSP}</div>`;
    return sbElement;
};

export const refreshSbResize = (sbElement: Element) => {
    if (!sbElement || !sbElement.classList.contains("sb")) {
        return;
    }
    sbElement.querySelectorAll(":scope > .sb__resize").forEach(item => item.remove());
    if (sbElement.getAttribute("data-sb-layout") !== "col") {
        return;
    }
    const children = Array.from(sbElement.querySelectorAll(":scope > [data-node-id]"));
    for (let i = 0; i < children.length - 1; i++) {
        const handle = document.createElement("span");
        handle.setAttribute("class", "sb__resize");
        handle.setAttribute("contenteditable", "false");
        children[i].after(handle);
    }
};

export const rebalanceSbWidth = (sbElement: Element): Array<{id: string, oldHTML: string}> => {
    if (!sbElement || sbElement.getAttribute("data-sb-layout") !== "col") {
        return [];
    }
    const children = Array.from(sbElement.querySelectorAll(":scope > [data-node-id]")) as HTMLElement[];
    if (children.length < 2) {
        return [];
    }
    if (!children.some(c => c.style.width)) {
        return [];
    }
    const handle = sbElement.querySelector(":scope > .sb__resize") as HTMLElement;
    let gapPx = 20;
    if (handle) {
        const hs = getComputedStyle(handle);
        gapPx = handle.offsetWidth + parseFloat(hs.marginLeft) + parseFloat(hs.marginRight);
    }
    const childCount = children.length;
    const gapShare = ((childCount - 1) * gapPx) / childCount + 0.5;
    const avgRatio = 1 / childCount;
    const ratios: number[] = children.map(c => {
        const match = c.style.width.match(/calc\(([\d.]+)%/);
        return match ? parseFloat(match[1]) / 100 : avgRatio;
    });
    const totalRatio = ratios.reduce((s, r) => s + r, 0) || 1;
    const changes: Array<{id: string, oldHTML: string}> = [];
    children.forEach((child, i) => {
        const oldHTML = child.outerHTML;
        const pct = Math.round((ratios[i] / totalRatio) * 100 * 10) / 10;
        child.style.width = `calc(${pct}% - ${gapShare}px)`;
        child.style.flex = "none";
        changes.push({id: child.getAttribute("data-node-id"), oldHTML});
    });
    return changes;
};

export const refreshSbAndPersistWidth = (sbElement: Element,
                                          doOperations: IOperation[], undoOperations: IOperation[]) => {
    if (!sbElement || !sbElement.parentElement) {
        return;
    }
    refreshSbResize(sbElement);
    const widthChanges = rebalanceSbWidth(sbElement);
    widthChanges.forEach(change => {
        const targetEl = sbElement.querySelector(`[data-node-id="${change.id}"]`);
        if (targetEl) {
            doOperations.push({action: "update", id: change.id, data: targetEl.outerHTML});
            undoOperations.splice(0, 0, {action: "update", id: change.id, data: change.oldHTML});
        }
    });
};

export const jumpToParent = (protyle: IProtyle, nodeElement: Element, type: "parent" | "next" | "previous") => {
    fetchPost("/api/block/getBlockSiblingID", {
        id: nodeElement.getAttribute("data-node-id"),
        notebook: protyle.notebookId,
    }, (response) => {
        const targetId = response.data[type];
        if (!targetId) {
            return;
        }
        openFileById({
            app: protyle.app,
            id: targetId,
            action: targetId !== protyle.block.rootID && protyle.block.showAll ? [Constants.CB_GET_ALL, Constants.CB_GET_FOCUS] : [Constants.CB_GET_FOCUS]
        });
    });
};

export const insertEmptyBlock = async (protyle: IProtyle, position: InsertPosition, target?: string | Element) => {
    const range = getEditorRange(protyle.wysiwyg.element);
    let blockElement: Element;
    if (typeof target === "string") {
        blockElement = protyle.wysiwyg.element.querySelector(`[data-node-id="${target}"]`);
    } else if (target) {
        blockElement = target;
    } else {
        const selectElements = protyle.wysiwyg.element.querySelectorAll(".protyle-wysiwyg--select");
        if (selectElements.length > 0) {
            if (position === "beforebegin") {
                blockElement = selectElements[0];
            } else {
                blockElement = selectElements[selectElements.length - 1];
            }
            hideElements(["select"], protyle);
        } else {
            blockElement = hasClosestBlock(range.startContainer) as HTMLElement;
            blockElement = getTopAloneElement(blockElement);
            // 
            if (blockElement.classList.contains("list")) {
                blockElement = hasClosestByClassName(range.startContainer, "li") as HTMLElement;
            } else if (blockElement.classList.contains("bq") || blockElement.classList.contains("callout")) {
                blockElement = hasClosestBlock(range.startContainer) as HTMLElement;
            }
        }
    }
    if (!blockElement) {
        return;
    }
    const undoFocusContext = getUndoFocusContext(protyle.wysiwyg.element, range);
    protyle.observerLoad?.disconnect();
    let newElement = genEmptyElement(false, true);
    let orderIndex = 1;
    const previousBlockElement = getPreviousBlockSibling(blockElement);
    if (blockElement.getAttribute("data-type") === "NodeListItem") {
        newElement = genListItemElement(blockElement, 0, true) as HTMLDivElement;
        orderIndex = parseInt(blockElement.parentElement.firstElementChild.getAttribute("data-marker"));
    } else if (position === "beforebegin" &&
        previousBlockElement?.getAttribute("data-type") === "NodeHeading" &&
        previousBlockElement.getAttribute("fold") === "1") {
        newElement = genHeadingElement(previousBlockElement, false, true) as HTMLDivElement;
    } else if (position === "afterend" && blockElement &&
        blockElement.getAttribute("data-type") === "NodeHeading" &&
        blockElement.getAttribute("fold") === "1") {
        newElement = genHeadingElement(blockElement, false, true) as HTMLDivElement;
    }

    const parentOldHTML = blockElement.parentElement.outerHTML;
    const newId = newElement.getAttribute("data-node-id");
    blockElement.insertAdjacentElement(position, newElement);
    if (blockElement.getAttribute("data-type") === "NodeListItem" && blockElement.getAttribute("data-subtype") === "o" &&
        !newElement.parentElement.classList.contains("protyle-wysiwyg")) {
        updateListOrder(newElement.parentElement, orderIndex);
        updateTransaction(protyle, newElement.parentElement, parentOldHTML, undoFocusContext);
    } else {
        let doOperations: IOperation[];
        if (position === "beforebegin") {
            doOperations = [{
                action: "insert",
                data: newElement.outerHTML,
                id: newId,
                nextID: blockElement.getAttribute("data-node-id"),
            }];
        } else {
            doOperations = [{
                action: "insert",
                data: newElement.outerHTML,
                id: newId,
                previousID: blockElement.getAttribute("data-node-id"),
            }];
        }
        const undoOperations: IOperation[] = [{
            action: "delete",
            id: newId,
            context: undoFocusContext,
        }];
        if (blockElement.parentElement.classList.contains("sb") &&
            blockElement.parentElement.getAttribute("data-sb-layout") === "col") {
            const mergeOperations = await turnsIntoOneTransaction({
                protyle,
                selectsElement: position === "afterend" ? [blockElement, blockElement.nextElementSibling] : [blockElement.previousElementSibling, blockElement],
                type: "BlocksMergeSuperBlock",
                level: "row",
                unfocus: true,
                getOperations: true,
            });
            doOperations.push(...mergeOperations.doOperations);
            undoOperations.splice(0, 0, ...mergeOperations.undoOperations);
        }
        transaction(protyle, doOperations, undoOperations);
    }
    focusByWbr(protyle.wysiwyg.element, range);
    scrollCenter(protyle);
};

export const genEmptyBlock = (zwsp = true, wbr = true, string?: string) => {
    let html = "";
    if (zwsp) {
        html = Constants.ZWSP;
    }
    if (wbr) {
        html += "<wbr>";
    }
    if (string) {
        html += string;
    }
    return `<div data-node-id="${Lute.NewNodeID()}" data-type="NodeParagraph" class="p"><div contenteditable="true" spellcheck="${window.scribli.config.editor.spellcheck}">${html}</div><div contenteditable="false" class="protyle-attr">${Constants.ZWSP}</div></div>`;
};

export const genEmptyElement = (zwsp = true, wbr = true, id?: string) => {
    const element = document.createElement("div");
    element.setAttribute("data-node-id", id || Lute.NewNodeID());
    element.setAttribute("data-type", "NodeParagraph");
    element.classList.add("p");
    element.innerHTML = `<div contenteditable="true" spellcheck="${window.scribli.config.editor.spellcheck}">${zwsp ? Constants.ZWSP : ""}${wbr ? "<wbr>" : ""}</div><div class="protyle-attr" contenteditable="false">${Constants.ZWSP}</div>`;
    return element;
};

export const genHeadingElement = (headElement: Element, getHTML = false, addWbr = false) => {
    const html = `<div data-subtype="${headElement.getAttribute("data-subtype")}" data-node-id="${Lute.NewNodeID()}" data-type="NodeHeading" class="${headElement.className}"><div contenteditable="true" spellcheck="false">${addWbr ? "<wbr>" : ""}</div><div class="protyle-attr" contenteditable="false">${Constants.ZWSP}</div></div>`;
    if (getHTML) {
        return html;
    } else {
        const tempElement = document.createElement("template");
        tempElement.innerHTML = html;
        return tempElement.content.firstElementChild;
    }
};

export const getLangByType = (type: string) => {
    let lang = type;
    switch (type) {
        case "NodeIFrame":
            lang = "IFrame";
            break;
        case "NodeAttributeView":
            lang = window.scribli.languages.database;
            break;
        case "NodeThematicBreak":
            lang = window.scribli.languages.line;
            break;
        case "NodeWidget":
            lang = window.scribli.languages.widget;
            break;
        case "NodeVideo":
            lang = window.scribli.languages.video;
            break;
        case "NodeAudio":
            lang = window.scribli.languages.audio;
            break;
        case "NodeBlockQueryEmbed":
            lang = window.scribli.languages.blockEmbed;
            break;
    }
    return lang;
};

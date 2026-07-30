import {fetchPost, fetchSyncPost} from "../../util/fetch";
import {focusBlock, focusByWbr, focusSideBlock, getEditorRange} from "../util/selection";
import {
    getContenteditableElement,
    getEmbedChildOperationContext,
    getEmbedChildOperationParentID,
    getFirstBlock,
    getNextBlockSibling,
    getParentBlock,
    getPreviousBlockSibling,
    getSbChildBlockCount,
    getTopAloneElement
} from "./getBlock";
import {Constants} from "../../constants";
import {blockRender} from "../render/blockRender";
import {processRender} from "../util/processCode";
import {highlightRender} from "../render/highlightRender";
import {hasClosestBlock, hasClosestByAttribute, hasTopClosestByAttribute, isInEmbedBlock} from "../util/hasClosest";
import {zoomOut} from "../../menus/protyle";
import {disabledProtyle, enableProtyle, onGet} from "../util/onGet";
import {getAllModels} from "../../layout/getAll";
import {avRender, refreshAV} from "../render/av/render";
import {removeFoldHeading} from "../util/heading";
import {cancelSB, genEmptyElement, genSBElement, refreshSbResize} from "../../block/util";
import {hideElements} from "../ui/hideElements";
import {reloadProtyle} from "../util/reload";
import {countBlockWord} from "../../layout/status";
import {resize} from "../util/resize";
import {processClonePHElement} from "../render/util";
import {scrollCenter} from "../../util/highlightById";
import {setFold} from "../util/blockFold";
import {isEncryptedBox} from "../../util/pathName";
import {queueTransaction} from "../util/transactionQueue";

const removeTopElement = (updateElement: Element, protyle: IProtyle) => {
    const topAloneElement = getTopAloneElement(updateElement);
    const doOperations: IOperation[] = [];
    if (topAloneElement !== updateElement) {
        updateElement.remove();
        doOperations.push({
            action: "delete",
            id: topAloneElement.getAttribute("data-node-id")
        });
    }
    topAloneElement.remove();
    if (protyle.wysiwyg.element.childElementCount === 0) {
        if (protyle.block.rootID === protyle.block.id) {
            const newId = Lute.NewNodeID();
            const newElement = genEmptyElement(false, false, newId);
            doOperations.push({
                action: "insert",
                data: newElement.outerHTML,
                id: newId,
                parentID: protyle.block.parentID
            });
            protyle.wysiwyg.element.innerHTML = newElement.outerHTML;
        } else {
            zoomOut({
                protyle,
                id: protyle.block.rootID,
                isPushBack: false,
                focusId: protyle.block.id,
            });
        }
    }
    if (doOperations.length > 0) {
        transaction(protyle, doOperations, []);
    }
};

const syncFoldAttr = (element: Element, operation: IOperation) => {
    const attrs = JSON.parse(operation.data);
    if (!Object.prototype.hasOwnProperty.call(attrs, "fold")) {
        return;
    }
    element.querySelectorAll(`[data-node-id="${operation.id}"]`).forEach(item => {
        if (attrs.fold === "1") {
            item.setAttribute("fold", "1");
        } else {
            item.removeAttribute("fold");
        }
    });
};

const promiseTransaction = (options: {
    protyle: IProtyle,
    doOperations: IOperation[],
    undoOperations: IOperation[],
    skipSync: boolean,
    callback?: () => void,
}) => {
    const protyle = options.protyle;
    const pendingEmbedElements = new Set<Element>();
    let range: Range;
    if (getSelection().rangeCount > 0) {
        range = getSelection().getRangeAt(0);
    }
    const isEmbedChildOperation = !!(range && getEmbedChildOperationContext(range.startContainer));
    if (!options.skipSync) {
        options.doOperations.forEach((operation: IOperation) => {
            if (operation.action === "update") {
                let updatedEmbed = false;

                const updateElements = Array.from(
                    protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`)
                );
                const currentUpdateElement = updateElements.find(item =>
                    item.getAttribute(Constants.ATTRIBUTE_EDITING) === "true" && getEmbedChildOperationContext(item));
                const currentEmbedContext = currentUpdateElement && getEmbedChildOperationContext(currentUpdateElement);
                const currentEmbedElement = currentEmbedContext && isInEmbedBlock(currentUpdateElement, false);

                const updateHTML = (item: Element, html: string, force = false) => {
                    if (!force && item.getAttribute(Constants.ATTRIBUTE_EDITING) === "true") {
                        item.removeAttribute(Constants.ATTRIBUTE_EDITING);
                        return;
                    }
                    const tempElement = document.createElement("template");
                    tempElement.innerHTML = html;
                    tempElement.content.querySelectorAll(".protyle-wysiwyg--select").forEach(selectItem => {
                        selectItem.classList.remove("protyle-wysiwyg--select");
                    });
                    const wbrElement = tempElement.content.querySelector("wbr");
                    if (wbrElement) {
                        wbrElement.remove();
                    }
                    item.outerHTML = tempElement.innerHTML;
                    updatedEmbed = true;
                };

                const allTempElement = document.createElement("template");
                allTempElement.innerHTML = operation.data;
                updateElements.forEach((item) => {
                    if ((currentEmbedElement && isInEmbedBlock(item, false) === currentEmbedElement) ||
                        (range && (item === range.startContainer || item.contains(range.startContainer)))) {
                        item.removeAttribute(Constants.ATTRIBUTE_EDITING);
                    } else {
                        updateHTML(item, operation.data, !!currentEmbedElement && !isInEmbedBlock(item));
                    }
                });
                protyle.wysiwyg.element.querySelectorAll(".protyle-wysiwyg__embed").forEach(item => {
                    if (item === currentEmbedContext?.resultElement ||
                        (range && (item === range.startContainer || item.contains(range.startContainer)))) {
                        item.removeAttribute(Constants.ATTRIBUTE_EDITING);
                    } else {
                        // 
                        const newTempElement = allTempElement.content.querySelector(`[data-node-id="${item.getAttribute("data-id")}"]`);
                        if (newTempElement && !isInEmbedBlock(newTempElement)) {
                            updateHTML(item.querySelector("[data-node-id]"), newTempElement.outerHTML);
                        } else {
                            item.removeAttribute(Constants.ATTRIBUTE_EDITING);
                        }
                    }
                });
                if (updatedEmbed) {
                    processRender(protyle.wysiwyg.element);
                    highlightRender(protyle.wysiwyg.element);
                    avRender(protyle.wysiwyg.element, protyle);
                }
                return;
            }
            if (operation.action === "delete" || operation.action === "append") {
                if ((operation.action === "delete" && isEmbedChildOperation) || protyle.options.backlinkData) {
                    Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`)).forEach(item => {
                        if (!isInEmbedBlock(item) && (!range || !item.contains(range.startContainer))) {
                            item.remove();
                        }
                    });
                }
                protyle.wysiwyg.element.querySelectorAll('[data-type="NodeBlockQueryEmbed"]').forEach((item) => {
                    if (item.querySelector(`[data-node-id="${operation.id}"]`)) {
                        pendingEmbedElements.add(item);
                    }
                });
                hideElements(["gutter"], protyle);
                return;
            }
            if (operation.action === "move") {
                if (protyle.options.backlinkData) {
                    const updateElements: Element[] = [];
                    Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`)).forEach(item => {
                        if (!isInEmbedBlock(item)) {
                            const topElement = hasTopClosestByAttribute(item, "data-node-id", null);
                            if (topElement && !topElement.contains(range.startContainer)) {
                                updateElements.push(item);
                            }
                        }
                    });
                    const originSbs: Element[] = [];
                    updateElements.forEach(item => {
                        const sb = item.closest('[data-type="NodeSuperBlock"]');
                        if (sb && !originSbs.includes(sb)) {
                            originSbs.push(sb);
                        }
                    });
                    let hasFind = false;
                    if (operation.previousID && updateElements.length > 0) {
                        Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.previousID}"]`)).forEach(item => {
                            if (!isInEmbedBlock(item) && !getNextBlockSibling(item)?.contains(range.startContainer)) {
                                item.after(processClonePHElement(updateElements[0].cloneNode(true) as Element));
                                hasFind = true;
                            }
                        });
                    } else if (updateElements.length > 0) {
                        Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.parentID}"]`)).forEach(item => {
                            if (!isInEmbedBlock(item) && !getFirstBlock(item).contains(range.startContainer)) {
                                const cloneElement = processClonePHElement(updateElements[0].cloneNode(true) as Element);
                                if (item.firstElementChild?.classList.contains("protyle-action")) {
                                    item.firstElementChild.after(cloneElement);
                                } else if (item.classList.contains("callout")) {
                                    item.querySelector(".callout-content").prepend(cloneElement);
                                } else {
                                    item.prepend(cloneElement);
                                }
                                hasFind = true;
                            }
                        });
                    }
                    updateElements.forEach(item => {
                        if (hasFind) {
                            item.remove();
                        } else if (!hasFind && item.parentElement) {
                            removeTopElement(item, protyle);
                        }
                    });
                    refreshSbs(...originSbs);
                }
                protyle.wysiwyg.element.querySelectorAll('[data-type="NodeBlockQueryEmbed"]').forEach((item) => {
                    if (item.querySelector(`[data-node-id="${operation.id}"],[data-node-id="${operation.parentID}"],[data-node-id="${operation.previousID}"]`)) {
                        pendingEmbedElements.add(item);
                    }
                });
                const moveEls = [operation.id, operation.parentID, operation.previousID]
                    .map(id => id ? protyle.wysiwyg.element.querySelector(`[data-node-id="${id}"]`) : null)
                    .filter(Boolean) as Element[];
                refreshSbs(...moveEls);
                return;
            }
            if (operation.action === "insert") {
                // 
                const insertedElement = protyle.wysiwyg.element.querySelector(`[data-node-id="${operation.id}"]`);
                const currentEmbedElement = insertedElement && isInEmbedBlock(insertedElement, false);
                if (insertedElement) {
                    protyle.wysiwyg.element.querySelectorAll("[data-type=\"NodeBlockQueryEmbed\"]").forEach(item => {
                        if (item !== currentEmbedElement && containsOperationAnchor(item, operation)) {
                            pendingEmbedElements.add(item);
                        }
                    });
                    getDocumentEmbedResults(protyle.wysiwyg.element, operation.parentID).forEach(item => {
                        const embedElement = isInEmbedBlock(item, false);
                        if (embedElement && embedElement !== currentEmbedElement) {
                            pendingEmbedElements.add(embedElement);
                        }
                    });
                }
                const cursorElements: Element[] = [];
                if (operation.previousID) {
                    Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.previousID}"]`)).forEach(item => {
                        const embedElement = isInEmbedBlock(item, false);
                        if (embedElement) {
                            if (embedElement !== currentEmbedElement) {
                                pendingEmbedElements.add(embedElement);
                            }
                            return;
                        }
                        if (getNextBlockSibling(item)?.getAttribute("data-node-id") !== operation.id &&
                            (!range || !item.contains(range.startContainer)) &&
                            !hasClosestByAttribute(item, "data-node-id", operation.id) &&
                            !item.parentElement.classList.contains("protyle-wysiwyg__embed")) {
                            item.insertAdjacentHTML("afterend", operation.data);
                            cursorElements.push(item.nextElementSibling);
                        }
                    });
                } else if (operation.nextID) {
                    Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.nextID}"]`)).forEach(item => {
                        const embedElement = isInEmbedBlock(item, false);
                        if (embedElement) {
                            if (embedElement !== currentEmbedElement) {
                                pendingEmbedElements.add(embedElement);
                            }
                            return;
                        }
                        if (getPreviousBlockSibling(item)?.getAttribute("data-node-id") !== operation.id &&
                            (!range || !item.contains(range.startContainer)) &&
                            !hasClosestByAttribute(item, "data-node-id", operation.id) &&
                            !item.parentElement.classList.contains("protyle-wysiwyg__embed")) {
                            item.insertAdjacentHTML("beforebegin", operation.data);
                            cursorElements.push(item.previousElementSibling);
                        }
                    });
                } else {
                    Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.parentID}"]`)).forEach(item => {
                        const embedElement = isInEmbedBlock(item, false);
                        if (embedElement) {
                            if (embedElement !== currentEmbedElement) {
                                pendingEmbedElements.add(embedElement);
                            }
                            return;
                        }
                        if (!range || !item.contains(range.startContainer)) {
                            if (item.firstElementChild && item.firstElementChild.classList.contains("protyle-action") &&
                                item.firstElementChild.nextElementSibling?.getAttribute("data-node-id") !== operation.id) {
                                item.firstElementChild.insertAdjacentHTML("afterend", operation.data);
                                cursorElements.push(item.firstElementChild.nextElementSibling);
                            } else if (item.classList.contains("callout") &&
                                item.querySelector("[data-node-id]")?.getAttribute("data-node-id") !== operation.id) {
                                item.querySelector(".callout-content").insertAdjacentHTML("afterbegin", operation.data);
                                cursorElements.push(item.querySelector("[data-node-id]"));
                            } else if (item.firstElementChild.getAttribute("data-node-id") !== operation.id) {
                                item.insertAdjacentHTML("afterbegin", operation.data);
                                cursorElements.push(item.firstElementChild);
                            }
                        }
                    });
                    getDocumentEmbedResults(protyle.wysiwyg.element, operation.parentID).forEach(item => {
                        const embedElement = isInEmbedBlock(item, false);
                        if (embedElement && embedElement !== currentEmbedElement) {
                            pendingEmbedElements.add(embedElement);
                        }
                    });
                }
                // 
                protyle.wysiwyg.element.querySelectorAll('[data-type="NodeHeading"]').forEach(item => {
                    if (item.lastElementChild.getAttribute("spin") === "1") {
                        item.lastElementChild.remove();
                    }
                });
                cursorElements.forEach(item => {
                    processRender(item);
                    highlightRender(item);
                    avRender(item, protyle);
                    blockRender(protyle, item);
                    item.querySelectorAll("wbr").forEach(wbrItem => {
                        wbrItem.remove();
                    });
                });
                protyle.wysiwyg.element.querySelectorAll("[parent-heading]").forEach(item => {
                    item.remove();
                });
                const insertedEl = protyle.wysiwyg.element.querySelector(`[data-node-id="${operation.id}"]`);
                refreshSbs(insertedEl);
                return;
            }
            if (operation.action === "setAttrs") {
                syncFoldAttr(protyle.wysiwyg.element, operation);
                const gutterFoldElement = protyle.gutter.element.querySelector('[data-type="fold"]');
                if (gutterFoldElement) {
                    gutterFoldElement.removeAttribute("disabled");
                }
                protyle.wysiwyg.element.querySelectorAll('[data-type="NodeBlockQueryEmbed"]').forEach((item) => {
                    if (item.querySelector(`[data-node-id="${operation.id}"]`)) {
                        pendingEmbedElements.add(item);
                    }
                });
            }
        });
        if (protyle.wysiwyg.element.childElementCount === 0 &&
            !protyle.block.showAll) {
            const newID = Lute.NewNodeID();
            const emptyElement = genEmptyElement(false, true, newID);
            protyle.wysiwyg.element.insertAdjacentElement("afterbegin", emptyElement);
            transaction(protyle, [{
                action: "insert",
                data: emptyElement.outerHTML,
                id: newID,
                parentID: protyle.block.parentID
            }]);
            focusByWbr(emptyElement, range);
        }
    }
    queueTransaction(protyle, () => fetchPost("/api/transactions", {
        session: protyle.id,
        app: Constants.SCRIBLI_APPID,
        transactions: [{
            doOperations: options.doOperations,
            undoOperations: options.undoOperations,
        }]
    }, (response) => {
        const ids: string[] = [];
        protyle.wysiwyg.element.querySelectorAll(".protyle-wysiwyg--select").forEach(item => {
            ids.push(item.getAttribute("data-node-id"));
        });
        countBlockWord(ids, protyle.block.rootID, true);
        if (!options.skipSync) {
            response.data[0].doOperations.forEach((operation: IOperation) => {
                if (operation.action === "unfoldHeading" || operation.action === "foldHeading") {
                    processFold(operation, protyle);
                    return;
                }
            });
        }
        pendingEmbedElements.forEach(item => {
            if (item.isConnected) {
                item.removeAttribute("data-render");
                blockRender(protyle, item);
            }
        });
        options.callback?.();
    }));
};

const containsOperationAnchor = (element: Element, operation: IOperation) => {
    const ids = new Set<string>();
    [operation.previousID, operation.nextID, operation.parentID].forEach(id => {
        if (id) {
            ids.add(id);
        }
    });
    return Array.from(element.querySelectorAll("[data-node-id]")).some(item => {
        const id = item.getAttribute("data-node-id");
        return isInEmbedBlock(item, false) === element && !!id && ids.has(id);
    });
};

const getDocumentEmbedResults = (element: Element, targetID?: string) => {
    if (!targetID) {
        return [];
    }
    return Array.from(element.querySelectorAll<HTMLElement>(
        ".protyle-wysiwyg__embed[data-allow-child-operation=\"true\"]"
    )).filter(item => item.getAttribute("data-id") === targetID && !getEmbedChildOperationContext(item)?.targetElement);
};

const refreshSbs = (...elements: (Element | undefined | null)[]) => {
    const sbs = new Set<Element>();
    elements.forEach(el => {
        if (el) {
            const sb = el.closest('[data-type="NodeSuperBlock"]');
            if (sb && sb.parentElement) {
                sbs.add(sb);
            }
        }
    });
    sbs.forEach(sb => refreshSbResize(sb));
};

const deleteBlock = (updateElements: Element[], id: string, protyle: IProtyle, isUndo: boolean) => {
    if (isUndo && updateElements[0]) {
        focusSideBlock(updateElements[0]);
    }
    const sbParents: Element[] = [];
    updateElements.forEach(item => {
        const sbAncestor = item.closest('[data-type="NodeSuperBlock"]');
        if (sbAncestor && !sbParents.includes(sbAncestor)) {
            sbParents.push(sbAncestor);
        }
        if (isUndo) {
            // 
            item.remove();
        } else {
            const topElement = getTopAloneElement(item);
            if (topElement) {
                topElement.remove();
            }
        }
    });
    protyle.wysiwyg.element.querySelectorAll('[data-type="NodeBlockQueryEmbed"]').forEach((item) => {
        if (item.querySelector(`[data-node-id="${id}"]`)) {
            item.removeAttribute("data-render");
            blockRender(protyle, item);
        }
    });
    refreshSbs(...sbParents);
};

const updateBlock = (updateElements: Element[], protyle: IProtyle, operation: IOperation, isUndo: boolean) => {
    const range = getSelection().rangeCount > 0 ? getSelection().getRangeAt(0) : null;
    updateElements.forEach(item => {
        let isRangeBlock = false;
        if (range && item.contains(range.startContainer)) {
            isRangeBlock = true;
        }
        //  
        // 
        let tableScrollLeft: number;
        let tableScrollTop: number;
        let contentScrollTop: number;
        if (item.classList.contains("table")) {
            tableScrollLeft = (item.firstElementChild as HTMLElement).scrollLeft;
            tableScrollTop = (item.firstElementChild as HTMLElement).scrollTop;
            if (isRangeBlock) {
                contentScrollTop = protyle.contentElement.scrollTop;
            }
        }
        item.insertAdjacentHTML("afterend",
            item.getAttribute("data-subtype") === "echarts" ? protyle.lute.SpinBlockDOM(operation.data) : operation.data);
        item = item.nextElementSibling;
        item.previousElementSibling.remove();

        const wbrElement = item.querySelector("wbr");
        if (isRangeBlock && isUndo) {
            if (wbrElement) {
                focusByWbr(item, range);
            } else {
                focusBlock(item);
            }
        }
        wbrElement?.remove();
        if (tableScrollLeft > 0) {
            (item.firstElementChild as HTMLElement).scrollLeft = tableScrollLeft;
        }
        if (tableScrollTop > 0) {
            (item.firstElementChild as HTMLElement).scrollTop = tableScrollTop;
        }
        if (contentScrollTop > 0) {
            protyle.contentElement.scrollTop = contentScrollTop;
            protyle.scroll.lastScrollTop = contentScrollTop - 1;
        }

        processRender(item);
        highlightRender(item);
        avRender(item, protyle);
        blockRender(protyle, item);
    });
};

export const onTransaction = (protyle: IProtyle, operations: IOperation[], isUndo: boolean) => {
    if (protyle.wysiwyg.element.firstElementChild?.classList.contains("protyle-password")) {
        return;
    }
    operations.forEach(operation => {
        const updateElements: Element[] = [];
        Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`)).forEach(item => {
            updateElements.push(item);
        });
        if (operation.action === "setAttrs") {
            syncFoldAttr(protyle.wysiwyg.element, operation);
            return;
        }
        if (operation.action === "unfoldHeading") {
            protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`).forEach(item => {
                item.removeAttribute("fold");
                if (isUndo) {
                    if (operation.retData) {
                        removeUnfoldRepeatBlock(operation.retData, protyle);
                        item.insertAdjacentHTML("afterend", operation.retData);
                    }
                    return;
                }
                const embedElement = isInEmbedBlock(item);
                if (embedElement) {
                    embedElement.removeAttribute("data-render");
                    blockRender(protyle, embedElement);
                    return;
                }
                if (operation.retData) {
                    removeUnfoldRepeatBlock(operation.retData, protyle);
                    item.insertAdjacentHTML("afterend", operation.retData);
                }
                if (operation.data === "remove") {
                    item.remove();
                }
            });
            if (operation.retData) {
                if (protyle.disabled) {
                    disabledProtyle(protyle);
                }
                processRender(protyle.wysiwyg.element);
                highlightRender(protyle.wysiwyg.element);
                avRender(protyle.wysiwyg.element, protyle);
                blockRender(protyle, protyle.wysiwyg.element);
                refreshSbs(...Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`)));
            }
            return;
        }
        if (operation.action === "foldHeading") {
            protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`).forEach(item => {
                item.setAttribute("fold", "1");
                if (!operation.retData) {
                    removeFoldHeading(item);
                }
            });
            if (operation.retData) {
                operation.retData.forEach((item: string) => {
                    let embedElement: HTMLElement | false;
                    Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${item}"]`)).find(itemElement => {
                        embedElement = isInEmbedBlock(itemElement);
                        if (embedElement) {
                            return true;
                        }
                        itemElement.remove();
                    });
                    if (embedElement) {
                        embedElement.removeAttribute("data-render");
                        blockRender(protyle, embedElement);
                    }
                });
                refreshSbs(...Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`)));
                if (protyle.wysiwyg.element.childElementCount === 0) {
                    zoomOut({
                        protyle,
                        id: protyle.block.rootID,
                        isPushBack: false,
                        focusId: operation.id,
                    });
                }
            }
            return;
        }
        if (operation.action === "delete") {
            if (updateElements.length > 0 || !isUndo) {
                deleteBlock(updateElements, operation.id, protyle, isUndo);
            } else if (isUndo) {
                zoomOut({
                    protyle,
                    id: protyle.block.rootID,
                    isPushBack: false,
                    focusId: operation.id,
                    callback() {
                        Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`)).forEach(item => {
                            if (!isInEmbedBlock(item)) {
                                updateElements.push(item);
                            }
                        });
                        deleteBlock(updateElements, operation.id, protyle, isUndo);
                    }
                });
            }
            return;
        }
        if (operation.action === "update") {
            if (updateElements.length === 0) {
                const newUpdateElement = protyle.wysiwyg.element.querySelector("[data-node-id]");
                if (newUpdateElement) {
                    const newUpdateId = newUpdateElement.getAttribute("data-node-id");
                    const tempElement = document.createElement("template");
                    tempElement.innerHTML = operation.data;
                    const newTempElement = tempElement.content.querySelector(`[data-node-id="${newUpdateId}"]`);
                    if (newTempElement) {
                        updateElements.push(newUpdateElement);
                        operation.data = newTempElement.outerHTML;
                        operation.id = newUpdateId;
                        // 
                        for (let i = 1; i < protyle.wysiwyg.element.childElementCount; i++) {
                            protyle.wysiwyg.element.childNodes[i].remove();
                            i--;
                        }
                    }
                }
            }
            if (updateElements.length > 0) {
                updateBlock(updateElements, protyle, operation, isUndo);
            } else if (isUndo) {
                zoomOut({
                    protyle,
                    id: protyle.block.rootID,
                    isPushBack: false,
                    focusId: operation.id,
                    callback() {
                        Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`)).forEach(item => {
                            if (!isInEmbedBlock(item)) {
                                updateElements.push(item);
                            }
                        });
                        updateBlock(updateElements, protyle, operation, isUndo);
                    }
                });
            }
            return;
        }
        if (operation.action === "updateAttrs") {
            const data = operation.data as any;
            const attrsResult: Record<string, string> = {};
            let bookmarkHTML = "";
            let nameHTML = "";
            let aliasHTML = "";
            let memoHTML = "";
            let avHTML = "";
            Object.keys(data.new).forEach(key => {
                attrsResult[key] = data.new[key];
                const escapeHTML = Lute.EscapeHTMLStr(data.new[key]);
                if (key === "bookmark") {
                    bookmarkHTML = `<div class="protyle-attr--bookmark">${escapeHTML}</div>`;
                } else if (key === "name") {
                    nameHTML = `<div class="protyle-attr--name"><svg><use xlink:href="#iconN"></use></svg>${escapeHTML}</div>`;
                } else if (key === "alias") {
                    aliasHTML = `<div class="protyle-attr--alias"><svg><use xlink:href="#iconA"></use></svg>${escapeHTML}</div>`;
                } else if (key === "memo") {
                    memoHTML = `<div class="protyle-attr--memo ariaLabel" aria-label="${escapeHTML}" data-position="north"><svg><use xlink:href="#iconM"></use></svg></div>`;
                } else if (key === "custom-avs" && data.new["av-names"]) {
                    avHTML = `<div class="protyle-attr--av"><svg><use xlink:href="#iconDatabase"></use></svg>${(data.new["av-names"])}</div>`;
                }
            });
            let nodeAttrHTML = bookmarkHTML + nameHTML + aliasHTML + memoHTML + avHTML;
            if (protyle.block.rootID === operation.id) {
                if (protyle.title) {
                    if (data.new["custom-avs"] && !data.new["av-names"]) {
                        nodeAttrHTML += protyle.title.element.querySelector(".protyle-attr--av")?.outerHTML || "";
                    }
                    const refElement = protyle.title.element.querySelector(".protyle-attr--refcount");
                    if (refElement) {
                        nodeAttrHTML += refElement.outerHTML;
                    }
                    if (data.new[Constants.CUSTOM_RIFF_DECKS] && data.new[Constants.CUSTOM_RIFF_DECKS] !== data.old[Constants.CUSTOM_RIFF_DECKS]) {
                        protyle.title.element.style.animation = "addCard 450ms linear";
                        protyle.title.element.setAttribute(Constants.CUSTOM_RIFF_DECKS, data.new[Constants.CUSTOM_RIFF_DECKS]);
                        setTimeout(() => {
                            protyle.title.element.style.animation = "";
                        }, 450);
                    } else if (!data.new[Constants.CUSTOM_RIFF_DECKS]) {
                        protyle.title.element.removeAttribute(Constants.CUSTOM_RIFF_DECKS);
                    }
                    protyle.title.element.querySelector(".protyle-attr").innerHTML = nodeAttrHTML;
                }
                protyle.wysiwyg.renderCustom(attrsResult);
                if (data.new[Constants.CUSTOM_SY_FULLWIDTH] !== data.old[Constants.CUSTOM_SY_FULLWIDTH]) {
                    resize(protyle);
                }
                if (data.new[Constants.CUSTOM_SY_READONLY] !== data.old[Constants.CUSTOM_SY_READONLY]) {
                    let customReadOnly = data.new[Constants.CUSTOM_SY_READONLY];
                    if (!customReadOnly) {
                        customReadOnly = window.scribli.config.editor.readOnly ? "true" : "false";
                    }
                    if (customReadOnly === "true") {
                        disabledProtyle(protyle);
                    } else {
                        enableProtyle(protyle);
                    }
                }
                if (data.new.icon !== data.old.icon ||
                    data.new["title-img"] !== data.old["title-img"] ||
                    data.new.tags !== data.old.tags && protyle.background) {
                    protyle.background.ial.icon = data.new.icon;
                    protyle.background.ial.tags = data.new.tags;
                    protyle.background.ial["title-img"] = data.new["title-img"];
                    protyle.background.render(protyle.background.ial, protyle.block.rootID);
                    protyle.model?.parent.setDocIcon(data.new.icon);
                }
                return;
            }
            protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`).forEach((item: HTMLElement) => {
                if (item.getAttribute("data-type") === "NodeThematicBreak") {
                    return;
                }
                Object.keys(data.old).forEach(key => {
                    item.removeAttribute(key);
                    if (key === "custom-avs") {
                        item.removeAttribute("av-names");
                    }
                });
                if (data.new.style && data.new[Constants.CUSTOM_RIFF_DECKS] && data.new[Constants.CUSTOM_RIFF_DECKS] !== data.old[Constants.CUSTOM_RIFF_DECKS]) {
                    data.new.style += ";animation:addCard 450ms linear";
                }
                Object.keys(data.new).forEach(key => {
                    if ("id" === key) {
                        return;
                    }

                    item.setAttribute(key, data.new[key]);
                    if (key === Constants.CUSTOM_RIFF_DECKS &&
                        data.new[Constants.CUSTOM_RIFF_DECKS] !== data.old[Constants.CUSTOM_RIFF_DECKS]) {
                        item.style.animation = "addCard 450ms linear";
                        setTimeout(() => {
                            if (item.parentElement) {
                                item.style.animation = "";
                            } else {
                                protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`).forEach((realItem: HTMLElement) => {
                                    realItem.style.animation = "";
                                });
                            }
                        }, 450);
                    }
                });
                if (data["data-av-type"]) {
                    item.setAttribute("data-av-type", data["data-av-type"]);
                }
                const attrElements = item.querySelectorAll(".protyle-attr");
                const attrElement = attrElements[attrElements.length - 1];
                if (data.new["custom-avs"] && !data.new["av-names"]) {
                    nodeAttrHTML += attrElement.querySelector(".protyle-attr--av")?.outerHTML || "";
                }
                const refElement = attrElement.querySelector(".protyle-attr--refcount");
                if (refElement) {
                    nodeAttrHTML += refElement.outerHTML;
                }
                attrElement.innerHTML = nodeAttrHTML + Constants.ZWSP;
            });
            return;
        }
        if (operation.action === "move") {
            if (operation.context?.ignoreProcess === "true") {
                return;
            }
            if (updateElements.length === 0) {
                getAllModels().editor.forEach(editor => {
                    const updateCloneElement = editor.editor.protyle.wysiwyg.element.querySelector(`[data-node-id="${operation.id}"]`);
                    if (updateCloneElement) {
                        updateElements.push(updateCloneElement.cloneNode(true) as Element);
                    }
                });
            }
            if (updateElements.length === 0) {
                window.scribli.blockPanels.forEach((item) => {
                    const updateCloneElement = item.element.querySelector(`[data-node-id="${operation.id}"]`);
                    if (updateCloneElement) {
                        updateElements.push(updateCloneElement.cloneNode(true) as Element);
                    }
                });
            }
            if (updateElements.length === 0) {
                const tempEl = document.createElement("div");
                tempEl.setAttribute("data-node-id", operation.id);
                tempEl.setAttribute("data-protyle-id", protyle.element.getAttribute("data-id"));
                updateElements.push(tempEl);
                fetchPost("/api/block/getBlockDOM", {
                    id: operation.id,
                    notebook: protyle.notebookId,
                }, (response) => {
                    document.querySelectorAll(`.protyle-wysiwyg [data-node-id="${response.data.id}"]`).forEach(item => {
                        if (item.getAttribute("data-protyle-id")) {
                            item.outerHTML = response.data.dom;
                            item.removeAttribute("data-protyle-id");
                        }
                    });
                });
            }
            let range;
            if (isUndo && getSelection().rangeCount > 0) {
                range = getSelection().getRangeAt(0);
                const rangeBlockElement = hasClosestBlock(range.startContainer);
                if (rangeBlockElement) {
                    if (getContenteditableElement(rangeBlockElement)) {
                        range.insertNode(document.createElement("wbr"));
                    } else {
                        getContenteditableElement(updateElements[0])?.insertAdjacentHTML("afterbegin", "<wbr>");
                    }
                }
            }
            let hasFind = false;
            // 
            const originSbs: Element[] = [];
            updateElements.forEach(item => {
                const sb = item.closest('[data-type="NodeSuperBlock"]');
                if (sb && !originSbs.includes(sb)) {
                    originSbs.push(sb);
                }
            });
            if (operation.previousID && updateElements.length > 0) {
                const previousElement = protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.previousID}"]`);
                if (previousElement.length === 0 && protyle.options.backlinkData && isUndo && getSelection().rangeCount > 0) {
                    const blockElement = hasTopClosestByAttribute(range.startContainer, "data-node-id", null);
                    if (blockElement) {
                        blockElement.before(processClonePHElement(updateElements[0].cloneNode(true) as Element));
                        hasFind = true;
                    }
                } else {
                    previousElement.forEach(item => {
                        if (!isInEmbedBlock(item)) {
                            item.after(processClonePHElement(updateElements[0].cloneNode(true) as Element));
                            hasFind = true;
                        }
                    });
                }
            } else if (updateElements.length > 0) {
                const parentElement = protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.parentID}"]`);
                if (!protyle.options.backlinkData && operation.parentID === protyle.block.parentID && !protyle.block.showAll) {
                    protyle.wysiwyg.element.prepend(processClonePHElement(updateElements[0].cloneNode(true) as Element));
                    hasFind = true;
                } else if (parentElement.length === 0 && protyle.options.backlinkData && isUndo && getSelection().rangeCount > 0) {
                    const topBlockElement = hasTopClosestByAttribute(getSelection().getRangeAt(0).startContainer, "data-node-id", null);
                    if (topBlockElement) {
                        topBlockElement.before(processClonePHElement(updateElements[0].cloneNode(true) as Element));
                        hasFind = true;
                    }
                } else {
                    parentElement.forEach(item => {
                        if (!isInEmbedBlock(item)) {
                            const cloneElement = processClonePHElement(updateElements[0].cloneNode(true) as Element);
                            if (item.firstElementChild?.classList.contains("protyle-action")) {
                                item.firstElementChild.after(cloneElement);
                            } else if (item.classList.contains("callout")) {
                                item.querySelector(".callout-content").prepend(cloneElement);
                            } else {
                                item.prepend(cloneElement);
                            }
                            hasFind = true;
                        }
                    });
                }
            }
            updateElements.forEach(item => {
                if (hasFind) {
                    item.remove();
                } else if (!hasFind && item.parentElement) {
                    removeTopElement(item, protyle);
                }
            });
            if (isUndo && range) {
                if (operation.data === "focus") {
                    Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`)).find(item => {
                        if (!isInEmbedBlock(item)) {
                            focusBlock(item);
                            return true;
                        }
                    });
                    document.querySelectorAll("wbr").forEach(item => {
                        item.remove();
                    });
                } else {
                    focusByWbr(protyle.wysiwyg.element, range);
                }
            }
            protyle.wysiwyg.element.querySelectorAll('[data-type="NodeBlockQueryEmbed"]').forEach((item) => {
                if (item.querySelector(`[data-node-id="${operation.id}"],[data-node-id="${operation.parentID}"],[data-node-id="${operation.previousID}"]`)) {
                    item.removeAttribute("data-render");
                    blockRender(protyle, item);
                }
            });
            const moveEls = [operation.id, operation.parentID, operation.previousID]
                .map(id => id ? protyle.wysiwyg.element.querySelector(`[data-node-id="${id}"]`) : null)
                .filter(Boolean) as Element[];
            refreshSbs(...moveEls);
            // 
            refreshSbs(...originSbs);
            return;
        }
        if (operation.action === "insert") {
            if (operation.context?.ignoreProcess === "true") {
                return;
            }
            const cursorElements = [];
            if (operation.previousID) {
                const previousElement = protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.previousID}"]`);
                if (previousElement.length === 0 && isUndo && protyle.wysiwyg.element.childElementCount === 0) {
                    protyle.wysiwyg.element.innerHTML = operation.data;
                    cursorElements.push(protyle.wysiwyg.element.firstElementChild);
                } else if (previousElement.length === 0 && protyle.options.backlinkData && isUndo && getSelection().rangeCount > 0) {
                    const blockElement = hasClosestBlock(getSelection().getRangeAt(0).startContainer);
                    if (blockElement) {
                        blockElement.insertAdjacentHTML("beforebegin", operation.data);
                        cursorElements.push(blockElement.previousElementSibling);
                    }
                } else {
                    previousElement.forEach(item => {
                        const embedElement = isInEmbedBlock(item, false);
                        if (embedElement) {
                            // 
                            embedElement.removeAttribute("data-render");
                            blockRender(protyle, embedElement);
                        } else {
                            item.insertAdjacentHTML("afterend", operation.data);
                            cursorElements.push(item.nextElementSibling);
                        }
                    });
                }
            } else if (operation.nextID) {
                Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.nextID}"]`)).forEach(item => {
                    const embedElement = isInEmbedBlock(item, false);
                    if (embedElement) {
                        // 
                        embedElement.removeAttribute("data-render");
                        blockRender(protyle, embedElement);
                    } else {
                        item.insertAdjacentHTML("beforebegin", operation.data);
                        cursorElements.push(item.previousElementSibling);
                    }
                });
            } else {
                const parentElement = protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.parentID}"]`);
                getDocumentEmbedResults(protyle.wysiwyg.element, operation.parentID).forEach(item => {
                    const embedElement = isInEmbedBlock(item, false);
                    if (embedElement) {
                        embedElement.removeAttribute("data-render");
                        blockRender(protyle, embedElement);
                    }
                });
                if (!protyle.options.backlinkData && operation.parentID === protyle.block.parentID && !protyle.block.showAll) {
                    protyle.wysiwyg.element.insertAdjacentHTML("afterbegin", operation.data);
                    cursorElements.push(protyle.wysiwyg.element.firstElementChild);
                } else if (parentElement.length === 0 && protyle.options.backlinkData && isUndo && getSelection().rangeCount > 0) {
                    const blockElement = hasClosestBlock(getSelection().getRangeAt(0).startContainer);
                    if (blockElement) {
                        blockElement.insertAdjacentHTML("beforebegin", operation.data);
                        cursorElements.push(blockElement.previousElementSibling);
                    }
                } else {
                    parentElement.forEach(item => {
                        const embedElement = isInEmbedBlock(item, false);
                        if (embedElement) {
                            embedElement.removeAttribute("data-render");
                            blockRender(protyle, embedElement);
                        } else {
                            if (item.firstElementChild?.classList.contains("protyle-action")) {
                                item.firstElementChild.insertAdjacentHTML("afterend", operation.data);
                                cursorElements.push(item.firstElementChild.nextElementSibling);
                            } else if (item.classList.contains("callout")) {
                                item.querySelector(".callout-content").insertAdjacentHTML("afterbegin", operation.data);
                                cursorElements.push(item.querySelector("[data-node-id]"));
                            } else {
                                item.insertAdjacentHTML("afterbegin", operation.data);
                                cursorElements.push(item.firstElementChild);
                            }
                        }
                    });
                }
            }
            // 
            protyle.wysiwyg.element.querySelectorAll('[data-type="NodeHeading"]').forEach(item => {
                if (item.lastElementChild.getAttribute("spin") === "1") {
                    item.lastElementChild.remove();
                }
            });
            if (cursorElements.length === 0) {
                return;
            }
            cursorElements.forEach(item => {
                // 
                item.querySelector(".protyle-attr--av")?.remove();
                item.removeAttribute("custom-avs");
                item.getAttributeNames().forEach(attr => {
                    if (attr.startsWith("custom-sy-av-s-text-")) {
                        item.removeAttribute(attr);
                    }
                });
                processRender(item);
                highlightRender(item);
                avRender(item, protyle);
                blockRender(protyle, item);
                refreshSbs(item);
                const wbrElement = item.querySelector("wbr");
                if (isUndo) {
                    if (operation.context?.setRange === "true") {
                        const range = getEditorRange(item);
                        if (wbrElement) {
                            focusByWbr(item, range);
                        } else {
                            focusBlock(item);
                        }
                    }
                } else if (wbrElement) {
                    wbrElement.remove();
                }
            });
            protyle.wysiwyg.element.querySelectorAll("[parent-heading]").forEach(item => {
                item.remove();
            });
            return;
        }
        if (operation.action === "append") {
            if (!protyle.options.backlinkData) {
                reloadProtyle(protyle, false);
            }
            return;
        }
        if (["addAttrViewCol", "updateAttrViewCol", "updateAttrViewColOptions",
            "updateAttrViewColOption", "updateAttrViewCell", "sortAttrViewRow", "sortAttrViewCol", "setAttrViewColHidden",
            "setAttrViewColWrap", "setAttrViewColWidth", "setAttrViewColAlign", "removeAttrViewColOption", "setAttrViewName", "setAttrViewFilters",
            "setAttrViewSorts", "setAttrViewNewItemTemplates", "setAttrViewColCalc", "removeAttrViewCol", "updateAttrViewColNumberFormat", "removeAttrViewBlock",
            "replaceAttrViewBlock", "updateAttrViewColTemplate", "setAttrViewColPin", "addAttrViewView", "setAttrViewColIcon",
            "removeAttrViewView", "setAttrViewViewName", "setAttrViewViewIcon", "duplicateAttrViewView", "duplicateAttrViewRow", "sortAttrViewView",
            "updateAttrViewColRelation", "setAttrViewPageSize", "updateAttrViewColRollup", "sortAttrViewKey", "setAttrViewColDesc",
            "duplicateAttrViewKey", "setAttrViewViewDesc", "setAttrViewCoverFrom", "setAttrViewCoverFromAssetKeyID",
            "setAttrViewBlockView", "setAttrViewCardSize", "setAttrViewCardAspectRatio", "hideAttrViewName", "setAttrViewShowIcon",
            "setAttrViewWrapField", "setAttrViewGroup", "removeAttrViewGroup", "hideAttrViewGroup", "sortAttrViewGroup",
            "foldAttrViewGroup", "hideAttrViewAllGroups", "setAttrViewFitImage", "setAttrViewDisplayFieldName",
            "insertAttrViewBlock", "setAttrViewColDateFillSpecificTime", "setAttrViewFillColBackgroundColor", "setAttrViewUpdatedIncludeTime",
            "setAttrViewCreatedIncludeTime"].includes(operation.action)) {
            if (!isUndo) {
                refreshAV(protyle, operation);
            } else if (operation.action === "setAttrViewName") {
                Array.from(protyle.wysiwyg.element.querySelectorAll(`.av[data-av-id="${operation.id}"]`)).forEach((item: HTMLElement) => {
                    const titleElement = item.querySelector(".av__title") as HTMLElement;
                    if (!titleElement) {
                        return;
                    }
                    titleElement.textContent = operation.data;
                    titleElement.dataset.title = operation.data;
                });
            }
            return;
        }
        if (operation.action === "doUpdateUpdated") {
            updateElements.forEach(item => {
                item.setAttribute("updated", operation.data);
            });
            return;
        }
    });
};

export const turnsIntoOneTransaction = async (options: {
    protyle: IProtyle,
    selectsElement: Element[],
    type: TTurnIntoOne,
    level?: TTurnIntoOneSub,
    unfocus?: boolean,
    getOperations?: boolean,
}) => {
    let parentElement: Element;
    const id = Lute.NewNodeID();
    if (options.type === "BlocksMergeSuperBlock") {
        parentElement = genSBElement(options.level, id);
        // 
        const firstChild = options.selectsElement[0] as HTMLElement;
        if (firstChild.style.width) {
            (parentElement as HTMLElement).style.width = firstChild.style.width;
            (parentElement as HTMLElement).style.flex = firstChild.style.flex;
            firstChild.style.width = "";
            firstChild.style.flex = "";
        }
    } else if (options.type === "Blocks2Blockquote") {
        parentElement = document.createElement("div");
        parentElement.classList.add("bq");
        parentElement.setAttribute("data-node-id", id);
        parentElement.setAttribute("data-type", "NodeBlockquote");
        parentElement.innerHTML = `<div class="protyle-attr" contenteditable="false">${Constants.ZWSP}</div>`;
    } else if (options.type === "Blocks2Callout") {
        parentElement = document.createElement("div");
        parentElement.classList.add("callout");
        parentElement.setAttribute("data-node-id", id);
        parentElement.setAttribute("data-type", "NodeCallout");
        parentElement.setAttribute("contenteditable", "false");
        parentElement.setAttribute("data-subtype", "NOTE");
        parentElement.innerHTML = `<div class="callout-info"><span class="callout-icon">✏️</span><span class="callout-title">Note</span></div><div class="callout-content"></div><div class="protyle-attr" contenteditable="false">${Constants.ZWSP}</div>`;
    } else if (options.type.endsWith("Ls")) {
        parentElement = document.createElement("div");
        parentElement.classList.add("list");
        parentElement.setAttribute("data-node-id", id);
        parentElement.setAttribute("data-type", "NodeList");
        if (options.type === "Blocks2ULs") {
            parentElement.setAttribute("data-subtype", "u");
        } else if (options.type === "Blocks2OLs") {
            parentElement.setAttribute("data-subtype", "o");
        } else {
            parentElement.setAttribute("data-subtype", "t");
        }
        let html = "";
        options.selectsElement.forEach((item, index) => {
            if (options.type === "Blocks2ULs") {
                html += `<div data-marker="*" data-subtype="u" data-node-id="${Lute.NewNodeID()}" data-type="NodeListItem" class="li"><div class="protyle-action" draggable="true"><svg><use xlink:href="#iconDot"></use></svg></div><div class="protyle-attr" contenteditable="false"></div></div>`;
            } else if (options.type === "Blocks2OLs") {
                html += `<div data-marker="${index + 1}." data-subtype="o" data-node-id="${Lute.NewNodeID()}" data-type="NodeListItem" class="li"><div class="protyle-action protyle-action--order" contenteditable="false" draggable="true">${index + 1}.</div><div class="protyle-attr" contenteditable="false"></div></div>`;
            } else {
                html += `<div data-marker="*" data-task=" " data-subtype="t" data-node-id="${Lute.NewNodeID()}" data-type="NodeListItem" class="li"><div class="protyle-action protyle-action--task" draggable="true"><svg><use xlink:href="#iconUncheck"></use></svg></div><div class="protyle-attr" contenteditable="false"></div></div>`;
            }
        });
        parentElement.innerHTML = html + '<div class="protyle-attr" contenteditable="false"></div>';
    }
    const previousId = options.selectsElement[0].getAttribute("data-node-id");
    const parentId = getEmbedChildOperationParentID(options.selectsElement[0]) ||
        getParentBlock(options.selectsElement[0]).getAttribute("data-node-id") || options.protyle.block.parentID;
    const doOperations: IOperation[] = [{
        action: "insert",
        id,
        data: parentElement.outerHTML,
        nextID: previousId,
        parentID: parentId
    }];
    const undoOperations: IOperation[] = [];
    if (options.selectsElement[0].previousElementSibling) {
        options.selectsElement[0].before(parentElement);
    } else {
        options.selectsElement[0].parentElement.prepend(parentElement);
    }
    let itemPreviousId: string;
    options.selectsElement.forEach((item, index) => {
        item.classList.remove("protyle-wysiwyg--select");
        item.removeAttribute("select-start");
        item.removeAttribute("select-end");
        const itemId = item.getAttribute("data-node-id");
        undoOperations.push({
            action: "move",
            id: itemId,
            previousID: itemPreviousId || id,
            parentID: parentId
        });
        if (options.type.endsWith("Ls")) {
            doOperations.push({
                action: "move",
                id: itemId,
                parentID: parentElement.children[index].getAttribute("data-node-id")
            });
            parentElement.children[index].firstElementChild.after(item);
        } else if (options.type === "Blocks2Callout") {
            doOperations.push({
                action: "move",
                id: itemId,
                previousID: itemPreviousId,
                parentID: id
            });
            parentElement.querySelector(".callout-content").insertAdjacentElement("beforeend", item);
        } else {
            doOperations.push({
                action: "move",
                id: itemId,
                previousID: itemPreviousId,
                parentID: id
            });
            parentElement.lastElementChild.before(item);
        }
        itemPreviousId = item.getAttribute("data-node-id");

        if (index === options.selectsElement.length - 1) {
            undoOperations.push({
                action: "delete",
                id,
            });
        }
        if (item.getAttribute("data-type") === "NodeBlockQueryEmbed") {
            item.removeAttribute("data-render");
            blockRender(options.protyle, item);
        }
    });
    if (parentElement.classList.contains("sb")) {
        refreshSbs(parentElement);
    } else if (parentElement.parentElement?.classList.contains("sb")) {
        refreshSbs(parentElement.parentElement);
    }
    if ((["Blocks2Blockquote", "Blocks2Callout"].includes(options.type) || options.type.endsWith("Ls")) &&
        parentElement.parentElement?.classList.contains("sb") && getSbChildBlockCount(parentElement.parentElement) === 1) {
        const cancelOperations = await cancelSB(options.protyle, parentElement.parentElement);
        doOperations.push(...cancelOperations.doOperations);
        undoOperations.splice(0, 0, ...cancelOperations.undoOperations);
    }
    if (options.getOperations) {
        return {
            doOperations,
            undoOperations,
        };
    }
    transaction(options.protyle, doOperations, undoOperations);
    if (!options.unfocus) {
        focusBlock(options.protyle.wysiwyg.element.querySelector(`[data-node-id="${options.selectsElement[0].getAttribute("data-node-id")}"]`));
    }
    hideElements(["gutter"], options.protyle);
};

const removeUnfoldRepeatBlock = (html: string, protyle: IProtyle) => {
    const temp = document.createElement("template");
    temp.innerHTML = html;
    Array.from(temp.content.children).forEach(item => {
        protyle.wysiwyg.element.querySelector(`[data-node-id="${item.getAttribute("data-node-id")}"]`)?.remove();
    });
};

export const turnsIntoTransaction = (options: {
    protyle: IProtyle,
    selectsElement?: Element[],
    nodeElement?: Element,
    type: TTurnInto,
    level?: number,
    isContinue?: boolean,
    range?: Range,
    unfocus?: boolean,
}) => {
    // 
    options.protyle.observerLoad?.disconnect();
    let selectsElement: Element[] = options.selectsElement;
    let range: Range;
    if (options.nodeElement) {
        range = getSelection().getRangeAt(0);
        range.insertNode(document.createElement("wbr"));
        selectsElement = Array.from(options.protyle.wysiwyg.element.querySelectorAll(".protyle-wysiwyg--select"));
        if (selectsElement.length === 0) {
            selectsElement = [options.nodeElement];
        }
        let isContinue = false;
        let isList = false;
        selectsElement.find((item, index) => {
            if (item.classList.contains("li")) {
                isList = true;
                return true;
            }
            if (selectsElement[index + 1] && getNextBlockSibling(item) === selectsElement[index + 1]) {
                isContinue = true;
            } else if (index !== selectsElement.length - 1) {
                isContinue = false;
                return true;
            }
        });
        if (isList) {
            return;
        }
        if (selectsElement.length === 1 && options.type === "Blocks2Hs" &&
            selectsElement[0].getAttribute("data-type") === "NodeHeading" &&
            options.level === parseInt(selectsElement[0].getAttribute("data-subtype").substr(1))) {
            options.type = "Blocks2Ps";
        }
        options.isContinue = isContinue;
    }

    let html = "";
    const doOperations: IOperation[] = [];
    const undoOperations: IOperation[] = [];
    let previousId: string;
    selectsElement.forEach((item: HTMLElement, index) => {
        item.classList.remove("protyle-wysiwyg--select");
        item.removeAttribute("select-start");
        item.removeAttribute("select-end");
        html += item.outerHTML;
        const id = item.getAttribute("data-node-id");

        const tempElement = document.createElement("template");
        if (!options.isContinue || options.level) {
            // @ts-ignore
            let newHTML = options.protyle.lute[options.type](item.outerHTML, options.level);
            tempElement.innerHTML = newHTML;

            if (!tempElement.content.querySelector(`[data-node-id="${id}"]`)) {
                undoOperations.push({
                    action: "insert",
                    id,
                    previousID: previousId || getPreviousBlockSibling(item)?.getAttribute("data-node-id"),
                    data: item.outerHTML,
                    parentID: getEmbedChildOperationParentID(item) || getParentBlock(item)?.getAttribute("data-node-id") ||
                        options.protyle.block.parentID || options.protyle.block.rootID,
                });
                Array.from(tempElement.content.children).forEach((tempItem: HTMLElement) => {
                    const tempItemId = tempItem.getAttribute("data-node-id");
                    doOperations.push({
                        action: "insert",
                        id: tempItemId,
                        previousID: tempItem.previousElementSibling?.getAttribute("data-node-id") || getPreviousBlockSibling(item)?.getAttribute("data-node-id"),
                        data: tempItem.outerHTML,
                        parentID: getEmbedChildOperationParentID(item) || getParentBlock(item)?.getAttribute("data-node-id") ||
                            options.protyle.block.parentID || options.protyle.block.rootID,
                    });
                    undoOperations.splice(0, 0, {
                        action: "delete",
                        id: tempItemId,
                    });
                });
                doOperations.push({
                    action: "delete",
                    id,
                });
                if (selectsElement[index + 1] && item === getPreviousBlockSibling(selectsElement[index + 1])) {
                    previousId = id;
                } else {
                    previousId = undefined;
                }
                item.outerHTML = newHTML;
            } else {
                let foldData;
                if (item.getAttribute("data-type") === "NodeHeading" && item.getAttribute("fold") === "1" &&
                    tempElement.content.firstElementChild.getAttribute("data-subtype") !== item.dataset.subtype) {
                    foldData = setFold(options.protyle, item, undefined, undefined, false, true);
                    newHTML = newHTML.replace(' fold="1"', "");
                }
                if (foldData && foldData.doOperations?.length > 0) {
                    doOperations.push(...foldData.doOperations);
                }
                undoOperations.push({
                    action: "update",
                    id,
                    data: item.outerHTML,
                });
                doOperations.push({
                    action: "update",
                    id,
                    data: newHTML
                });
                if (foldData && foldData.undoOperations?.length > 0) {
                    undoOperations.push(...foldData.undoOperations);
                }
                item.insertAdjacentHTML("afterend", newHTML);
                item = item.nextElementSibling as HTMLElement;
                item.previousElementSibling.remove();
                item.setAttribute(Constants.ATTRIBUTE_EDITING, "true");
            }
        } else {
            undoOperations.push({
                action: "insert",
                id,
                previousID: doOperations[doOperations.length - 1]?.id || getPreviousBlockSibling(item)?.getAttribute("data-node-id"),
                data: item.outerHTML,
                parentID: getEmbedChildOperationParentID(item) || getParentBlock(item)?.getAttribute("data-node-id") ||
                    options.protyle.block.parentID || options.protyle.block.rootID,
            });
            doOperations.push({
                action: "delete",
                id,
            });
            if (index === selectsElement.length - 1) {
                // @ts-ignore
                const newHTML = options.protyle.lute[options.type](html, options.level);
                tempElement.innerHTML = newHTML;
                Array.from(tempElement.content.children).forEach((tempItem: HTMLElement) => {
                    const tempItemId = tempItem.getAttribute("data-node-id");
                    doOperations.push({
                        action: "insert",
                        id: tempItemId,
                        previousID: tempItem.previousElementSibling?.getAttribute("data-node-id") || getPreviousBlockSibling(item)?.getAttribute("data-node-id"),
                        data: tempItem.outerHTML,
                        parentID: getEmbedChildOperationParentID(item) || getParentBlock(item)?.getAttribute("data-node-id") ||
                            options.protyle.block.parentID || options.protyle.block.rootID,
                    });
                    undoOperations.splice(0, 0, {
                        action: "delete",
                        id: tempItemId,
                    });
                });
                item.outerHTML = newHTML;
            } else {
                item.remove();
            }
        }
    });
    transaction(options.protyle, doOperations, undoOperations);
    processRender(options.protyle.wysiwyg.element);
    highlightRender(options.protyle.wysiwyg.element);
    avRender(options.protyle.wysiwyg.element, options.protyle);
    blockRender(options.protyle, options.protyle.wysiwyg.element);
    if (!options.unfocus) {
        if (range || options.range) {
            focusByWbr(options.protyle.wysiwyg.element, range || options.range);
        } else {
            focusBlock(options.protyle.wysiwyg.element.querySelector(`[data-node-id="${selectsElement[0].getAttribute("data-node-id")}"]`));
        }
    }
    hideElements(["gutter"], options.protyle);
};

export const turnsOneInto = async (options: {
    protyle: IProtyle,
    nodeElement: Element,
    id: string,
    type: string,
    level?: number
}) => {
    if (!options.nodeElement.querySelector("wbr")) {
        getContenteditableElement(options.nodeElement)?.insertAdjacentHTML("afterbegin", "<wbr>");
    }
    if (["CancelBlockquote", "CancelList", "CancelCallout"].includes(options.type)) {
        for (const item of options.nodeElement.querySelectorAll('[data-type="NodeHeading"][fold="1"]')) {
            const itemId = item.getAttribute("data-node-id");
            item.removeAttribute("fold");
            const response = await fetchSyncPost("/api/transactions", {
                session: options.protyle.id,
                app: Constants.SCRIBLI_APPID,
                transactions: [{
                    doOperations: [{
                        action: "unfoldHeading",
                        id: itemId,
                    }],
                    undoOperations: [{
                        action: "foldHeading",
                        id: itemId
                    }],
                }]
            });
            options.protyle.undo.add([{
                action: "unfoldHeading",
                id: itemId,
            }], [{
                action: "foldHeading",
                id: itemId
            }], options.protyle);
            item.insertAdjacentHTML("afterend", response.data[0].doOperations[0].retData);
        }
    }
    const oldHTML = options.nodeElement.outerHTML;
    const previousBlockElement = getPreviousBlockSibling(options.nodeElement);
    let previousId = previousBlockElement?.getAttribute("data-node-id");
    if (!previousBlockElement && options.protyle.block.showAll) {
        const response = await fetchSyncPost("/api/block/getBlockRelevantIDs", {
            id: options.id,
            notebook: options.protyle.notebookId,
        });
        previousId = response.data.previousID;
    }
    const parentId = getEmbedChildOperationParentID(options.nodeElement) ||
        getParentBlock(options.nodeElement).getAttribute("data-node-id") || options.protyle.block.parentID;
    // @ts-ignore
    const newHTML = options.protyle.lute[options.type](options.nodeElement.outerHTML, options.level);
    options.nodeElement.insertAdjacentHTML("afterend", newHTML);
    options.nodeElement = options.nodeElement.nextElementSibling as HTMLElement;
    options.nodeElement.previousElementSibling.remove();
    if (["CancelBlockquote", "CancelList", "CancelCallout"].includes(options.type)) {
        const tempElement = document.createElement("template");
        tempElement.innerHTML = newHTML;
        const doOperations: IOperation[] = [{
            action: "delete",
            id: options.id
        }];
        const undoOperations: IOperation[] = [];
        let tempPreviousId = previousId;
        Array.from(tempElement.content.children).forEach((item) => {
            const tempId = item.getAttribute("data-node-id");
            doOperations.push({
                action: "insert",
                data: item.outerHTML,
                id: tempId,
                previousID: tempPreviousId,
                parentID: parentId
            });
            undoOperations.push({
                action: "delete",
                id: tempId
            });
            tempPreviousId = tempId;
        });
        undoOperations.push({
            action: "insert",
            data: oldHTML,
            id: options.id,
            previousID: previousId,
            parentID: parentId
        });
        transaction(options.protyle, doOperations, undoOperations);
    } else {
        updateTransaction(options.protyle, options.nodeElement, oldHTML);
    }
    focusByWbr(options.protyle.wysiwyg.element, getEditorRange(options.protyle.wysiwyg.element));
    options.protyle.wysiwyg.element.querySelectorAll('[data-type~="block-ref"]').forEach(item => {
        if (item.textContent === "") {
            fetchPost("/api/block/getRefText", {id: item.getAttribute("data-id")}, (response) => {
                item.innerHTML = response.data;
            });
        }
    });
    blockRender(options.protyle, options.protyle.wysiwyg.element);
    processRender(options.protyle.wysiwyg.element);
    highlightRender(options.protyle.wysiwyg.element);
    avRender(options.protyle.wysiwyg.element, options.protyle);
};

export const transaction = (protyle: IProtyle, doOperations: IOperation[], undoOperations?: IOperation[],
                            options?: {
                                skipSync?: boolean,
                                callback?: () => void,
                            }) => {
    if (doOperations.length === 0) {
        return;
    }
    if (!protyle) {
        fetchPost("/api/transactions", {
            session: Constants.SCRIBLI_APPID,
            app: Constants.SCRIBLI_APPID,
            transactions: [{
                doOperations
            }]
        }, options?.callback);
        return;
    }
    if (undoOperations) {
        if (window.scribli.config.fileTree.openFilesUseCurrentTab && protyle.model) {
            protyle.model.headElement.classList.remove("item--unupdate");
        }
        protyle.updated = true;
        protyle.undo.add(doOperations, undoOperations, protyle);
    }
    if (protyle?.lite) {
        return;
    }
    promiseTransaction({
        protyle: protyle,
        doOperations: doOperations,
        undoOperations: undoOperations,
        skipSync: options?.skipSync,
        callback: options?.callback,
    });
    doOperations.find(item => {
        if (item.action === "insert") {
            protyle.observerLoad?.disconnect();
            return true;
        }
    });
};

const processFold = (operation: IOperation, protyle: IProtyle) => {
    if (operation.action === "unfoldHeading" || operation.action === "foldHeading") {
        const gutterFoldElement = protyle.gutter.element.querySelector('[data-type="fold"]');
        if (gutterFoldElement) {
            gutterFoldElement.removeAttribute("disabled");
        }
        if (operation.action === "unfoldHeading") {
            const scrollTop = protyle.contentElement.scrollTop;
            protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`).forEach(item => {
                const embedElement = isInEmbedBlock(item);
                if (embedElement) {
                    embedElement.removeAttribute("data-render");
                    blockRender(protyle, embedElement);
                    return;
                }
                item.removeAttribute("fold");
                if (!item.lastElementChild.classList.contains("protyle-attr")) {
                    item.lastElementChild.remove();
                }
                removeUnfoldRepeatBlock(operation.retData, protyle);
                item.insertAdjacentHTML("afterend", operation.retData);
                if (operation.data === "remove") {
                    // 
                    const selection = getSelection();
                    if (selection.rangeCount > 0 && item.contains(selection.getRangeAt(0).startContainer)) {
                        focusBlock(item.nextElementSibling, undefined, true);
                    }
                    item.remove();
                }
            });
            if (protyle.disabled) {
                disabledProtyle(protyle);
            }
            processRender(protyle.wysiwyg.element);
            highlightRender(protyle.wysiwyg.element);
            avRender(protyle.wysiwyg.element, protyle);
            blockRender(protyle, protyle.wysiwyg.element);
            refreshSbs(...Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`)));
            if (operation.context?.focusId) {
                const focusElement = protyle.wysiwyg.element.querySelector(`[data-node-id="${operation.context.focusId}"]`);
                focusBlock(focusElement);
                scrollCenter(protyle, focusElement);
            } else {
                protyle.contentElement.scrollTop = scrollTop;
                protyle.scroll.lastScrollTop = scrollTop;
            }
            return;
        }
        protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`).forEach(item => {
            const embedElement = isInEmbedBlock(item);
            if (embedElement) {
                embedElement.removeAttribute("data-render");
                blockRender(protyle, embedElement);
            } else {
                item.setAttribute("fold", "1");
                removeFoldHeading(item);
            }
        });
        refreshSbs(...Array.from(protyle.wysiwyg.element.querySelectorAll(`[data-node-id="${operation.id}"]`)));
        if (protyle.wysiwyg.element.lastElementChild.getAttribute("data-eof") !== "2" &&
            !protyle.scroll.element.classList.contains("fn__none") &&
            protyle.contentElement.scrollHeight - protyle.contentElement.scrollTop < protyle.contentElement.clientHeight * 2    // 
        ) {
            const getDocParam: IObject = {
                id: protyle.wysiwyg.element.lastElementChild.getAttribute("data-node-id"),
                mode: 2,
                size: window.scribli.config.editor.dynamicLoadBlocks,
            };
            if (isEncryptedBox(protyle.notebookId)) {
                getDocParam.notebook = protyle.notebookId;
            }
            fetchPost("/api/filetree/getDoc", getDocParam, getResponse => {
                onGet({
                    data: getResponse,
                    protyle,
                    action: [Constants.CB_GET_APPEND, Constants.CB_GET_UNCHANGEID],
                });
            });
        }
        return;
    }
};

export const updateTransaction = (protyle: IProtyle, element: Element, oldHTML: string, undoContext?: Record<string, string>) => {
    const id = element.getAttribute("data-node-id");
    const newHTML = element.outerHTML;
    if (newHTML === oldHTML.replace("<wbr>", "")) {
        return;
    }
    element.setAttribute(Constants.ATTRIBUTE_EDITING, "true");
    transaction(protyle, [{
        id,
        data: newHTML,
        action: "update"
    }], [{
        id,
        data: oldHTML,
        action: "update",
        context: undoContext,
    }]);
};

export const updateBatchTransaction = (nodeElements: Element[], protyle: IProtyle, cb: (e: HTMLElement) => void) => {
    const operations: IOperation[] = [];
    const undoOperations: IOperation[] = [];
    nodeElements.forEach((element) => {
        const id = element.getAttribute("data-node-id");
        element.classList.remove("protyle-wysiwyg--select");
        element.removeAttribute("select-start");
        element.removeAttribute("select-end");
        undoOperations.push({
            action: "update",
            id,
            data: element.outerHTML
        });
        cb(element as HTMLElement);
        element.setAttribute(Constants.ATTRIBUTE_EDITING, "true");
        operations.push({
            action: "update",
            id,
            data: element.outerHTML
        });
    });
    transaction(protyle, operations, undoOperations);
};

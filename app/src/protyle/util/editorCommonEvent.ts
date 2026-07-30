import {focusBlock, focusByRange, getRangeByPoint} from "./selection";
import {
    getContenteditableElement,
    getNextBlockSibling,
    getParentBlock,
    getPreviousBlockSibling,
    getSbChildBlockCount,
    getTopAloneElement
} from "../wysiwyg/getBlock";
import {hideCaretLine, hideDragTip, showCaretLine, showDragTip, transparentImgSrc} from "./dragTip";
import {
    hasClosestBlock,
    hasClosestByAttribute,
    hasClosestByClassName,
    hasClosestByTag,
    hasTopClosestByAttribute,
    isInEmbedBlock
} from "./hasClosest";
import {Constants} from "../../constants";
import {paste} from "./paste";
import {
    cancelSB,
    genEmptyElement,
    genSBElement,
    insertEmptyBlock,
    refreshSbAndPersistWidth,
    refreshSbResize
} from "../../block/util";
import {transaction, turnsIntoOneTransaction} from "../wysiwyg/transaction";
import {updateListOrder} from "../wysiwyg/list";
import {fetchPost, fetchSyncPost} from "../../util/fetch";
import {onGet} from "./onGet";
import {getAllEditor} from "../../layout/getAll";
import {updatePanelByEditor} from "../../editor/util";
import {blockRender} from "../render/blockRender";
import {uploadFiles, uploadLocalFiles} from "../upload";
import {insertHTML} from "./insertHTML";
import {isBrowser} from "../../util/functions";
import {hideElements} from "../ui/hideElements";
import {insertAttrViewBlockAnimation} from "../render/av/row";
import * as dayjs from "dayjs";
import {zoomOut} from "../../menus/protyle";
/// #if !BROWSER
import {webUtils} from "electron";
import {dragUpload} from "../render/av/asset";
/// #endif
import {addDragFill, getTypeByCellElement} from "../render/av/cell";
import {processClonePHElement} from "../render/util";
import {insertGalleryItemAnimation} from "../render/av/gallery/item";
import {clearSelect} from "./clear";
import {dragoverTab} from "../render/av/view";
import {setFold} from "./blockFold";
import {isEncryptedBox} from "../../util/pathName";

const convertListItemSubtype = (listItem: Element, subtype: string) => {
    const actionElement = listItem.querySelector(".protyle-action");
    if (!actionElement || !["u", "o", "t"].includes(subtype)) {
        return;
    }
    if (subtype === "o") {
        actionElement.outerHTML = '<div contenteditable="false" draggable="true" class="protyle-action protyle-action--order">1.</div>';
        listItem.setAttribute("data-marker", "1.");
        listItem.removeAttribute("data-task");
    } else if (subtype === "t") {
        actionElement.outerHTML = '<div class="protyle-action protyle-action--task" draggable="true"><svg><use xlink:href="#iconUncheck"></use></svg></div>';
        listItem.setAttribute("data-marker", "*");
        listItem.setAttribute("data-task", " ");
    } else {
        actionElement.outerHTML = '<div class="protyle-action" draggable="true"><svg><use xlink:href="#iconDot"></use></svg></div>';
        listItem.setAttribute("data-marker", "*");
        listItem.removeAttribute("data-task");
    }
    listItem.setAttribute("data-subtype", subtype);
    listItem.classList.remove("protyle-task--done");
    listItem.setAttribute(Constants.ATTRIBUTE_EDITING, "true");
};

const getTargetListItem = (targetElement: Element, isBottom: boolean) => {
    if (targetElement.classList.contains("li")) {
        return targetElement as HTMLElement;
    }
    if (targetElement.classList.contains("list")) {
        const listItems = targetElement.querySelectorAll(":scope > .li");
        return (isBottom ? listItems[listItems.length - 1] : listItems[0]) as HTMLElement;
    }
    return targetElement.closest(".li") as HTMLElement;
};

const moveTo = async (protyle: IProtyle, sourceElements: Element[], targetElement: Element,
                      isSameDoc: boolean, position: InsertPosition, isCopy: boolean) => {
    const doOperations: IOperation[] = [];
    const undoOperations: IOperation[] = [];
    const copyFoldHeadingIds: { newId: string, oldId: string }[] = [];
    const targetId = targetElement.getAttribute("data-node-id");
    const newSourceElements: Element[] = [];
    let tempTargetElement = targetElement;
    let isSameLi = true;
    sourceElements.find(item => {
        if (!item.classList.contains("li") || !targetElement.classList.contains("li")) {
            isSameLi = false;
            return true;
        }
    });
    let newListElement: Element;
    let newListId: string;
    const orderListElements: { [key: string]: Element } = {};
    const sourcePositions = new Map<string, { previousID: string, parentID: string }>();
    for (const item of sourceElements) {
        const id = item.getAttribute("data-node-id");
        if (id) {
            const parentBlock = getParentBlock(item);
            let srcParentID = parentBlock?.getAttribute("data-node-id");
            if (!srcParentID) {
                let srcRootID = "";
                const sourceEditor = getAllEditor().find(editor =>
                    editor.protyle.wysiwyg.element === parentBlock);
                if (sourceEditor?.protyle?.block?.rootID) {
                    srcRootID = sourceEditor.protyle.block.rootID;
                }
                if (srcRootID) {
                    srcParentID = srcRootID;
                } else {
                    const response = await fetchSyncPost("/api/block/getBlockInfo", {id});
                    srcParentID = response?.data?.rootID || "";
                }
            }
            sourcePositions.set(id, {
                previousID: getPreviousBlockSibling(item)?.getAttribute("data-node-id") || "",
                parentID: srcParentID || "",
            });
        }
    }
    for (let index = sourceElements.length - 1; index >= 0; index--) {
        const item = sourceElements[index];
        const originalSubtype = item.getAttribute("data-subtype");
        const id = item.getAttribute("data-node-id");
        const parentID = getParentBlock(item).getAttribute("data-node-id") || protyle.block.parentID || protyle.block.rootID;
        if (item.getAttribute("data-type") === "NodeListItem" && !newListId && !isSameLi) {
            newListId = Lute.NewNodeID();
            newListElement = document.createElement("div");
            newListElement.innerHTML = `<div data-subtype="${item.getAttribute("data-subtype")}" data-node-id="${newListId}" data-type="NodeList" class="list"><div class="protyle-attr" contenteditable="false">${Constants.ZWSP}</div></div>`;
            newListElement = newListElement.firstElementChild;
            doOperations.push({
                action: "insert",
                data: newListElement.outerHTML,
                id: newListId,
                previousID: position === "afterbegin" ? null : (position === "afterend" ? targetId : getPreviousBlockSibling(tempTargetElement)?.getAttribute("data-node-id")),
                parentID: position === "afterbegin" ? targetId : (getParentBlock(tempTargetElement)?.getAttribute("data-node-id") || protyle.block.parentID || protyle.block.rootID),
            });
            undoOperations.push({
                action: "delete",
                id: newListId
            });
            tempTargetElement.insertAdjacentElement(position, newListElement);
            newSourceElements.push(newListElement);
        }
        const copyNewId = Lute.NewNodeID();
        if (isCopy && item.getAttribute("data-type") === "NodeHeading" && item.getAttribute("fold") === "1") {
            copyFoldHeadingIds.push({
                newId: copyNewId,
                oldId: id
            });
        }

        let copyElement;
        if (isCopy) {
            undoOperations.push({
                action: "delete",
                id: copyNewId,
            });
        } else {
            const srcPos = sourcePositions.get(id) || {previousID: "", parentID};
            undoOperations.push({
                action: "move",
                id,
                previousID: srcPos.previousID,
                parentID: srcPos.parentID,
            });
        }
        if (!isSameDoc && !isCopy) {
            const sameElement = protyle.wysiwyg.element.querySelector(`[data-node-id="${id}"]`);
            if (sameElement) {
                sameElement.remove();
            }
        }
        if (isCopy) {
            copyElement = item.cloneNode(true) as HTMLElement;
            copyElement.setAttribute("data-node-id", copyNewId);
            copyElement.querySelectorAll("[data-node-id]").forEach((e) => {
                const newId = Lute.NewNodeID();
                e.setAttribute("data-node-id", newId);
                e.setAttribute("updated", newId.split("-")[0]);
            });
            const targetSubtype = targetElement.getAttribute("data-subtype");
            if (copyElement.getAttribute("data-type") === "NodeListItem" &&
                targetElement.getAttribute("data-type") === "NodeListItem" && targetSubtype &&
                copyElement.getAttribute("data-subtype") !== targetSubtype) {
                convertListItemSubtype(copyElement, targetSubtype);
            }
            if (newListId) {
                newListElement.insertAdjacentElement("afterbegin", copyElement);
                doOperations.push({
                    action: "insert",
                    id: copyNewId,
                    data: copyElement.outerHTML,
                    parentID: newListId,
                });
            } else {
                tempTargetElement.insertAdjacentElement(position, copyElement);
                doOperations.push({
                    action: "insert",
                    id: copyNewId,
                    data: copyElement.outerHTML,
                    previousID: position === "afterbegin" ? null : (position === "afterend" ? targetId : getPreviousBlockSibling(copyElement)?.getAttribute("data-node-id")),
                    parentID: position === "afterbegin" ? targetId : (getParentBlock(copyElement)?.getAttribute("data-node-id") || protyle.block.parentID || protyle.block.rootID),
                });
                newSourceElements.push(copyElement);
            }
        } else {
            let topSourceElement = getTopAloneElement(item);
            const oldSourceParentElement = getParentBlock(item);
            if (item.classList.contains("li") && item.getAttribute("data-subtype") === "o") {
                orderListElements[item.parentElement.getAttribute("data-node-id")] = item.parentElement;
            }
            if (newListId) {
                newListElement.insertAdjacentElement("afterbegin", item);
                doOperations.push({
                    action: "move",
                    id,
                    parentID: newListId,
                });
            } else {
                tempTargetElement.insertAdjacentElement(position, item);
                doOperations.push({
                    action: "move",
                    id,
                    previousID: position === "afterbegin" ? null : (position === "afterend" ? targetId : getPreviousBlockSibling(item)?.getAttribute("data-node-id")),
                    parentID: position === "afterbegin" ? targetId : (getParentBlock(item)?.getAttribute("data-node-id") || protyle.block.parentID || protyle.block.rootID),
                });
                newSourceElements.push(item);
            }

            if (topSourceElement !== item) {
                if (topSourceElement.contains(item)) {
                    topSourceElement = getTopAloneElement(oldSourceParentElement);
                }
                doOperations.push({
                    action: "delete",
                    id: topSourceElement.getAttribute("data-node-id"),
                });
                undoOperations.push({
                    action: "insert",
                    data: topSourceElement.outerHTML,
                    id: topSourceElement.getAttribute("data-node-id"),
                    previousID: getPreviousBlockSibling(topSourceElement)?.getAttribute("data-node-id"),
                    parentID: getParentBlock(topSourceElement)?.getAttribute("data-node-id") || protyle.block.parentID || protyle.block.rootID
                });
                const topSourceParentElement = topSourceElement.parentElement;
                topSourceElement.remove();
                if (!isSameDoc) {
                    const sameElement = protyle.wysiwyg.element.querySelector(`[data-node-id="${topSourceElement.getAttribute("data-node-id")}"]`);
                    if (sameElement) {
                        sameElement.remove();
                    }
                }
                if (topSourceParentElement.classList.contains("sb") && getSbChildBlockCount(topSourceParentElement) === 1) {
                    if (isSameDoc) {
                        const sbData = await cancelSB(protyle, topSourceParentElement);
                        doOperations.push(sbData.doOperations[0], sbData.doOperations[1]);
                        undoOperations.push(sbData.undoOperations[1], sbData.undoOperations[0]);
                    } else {
                        const allEditor = getAllEditor();
                        for (let i = 0; i < allEditor.length; i++) {
                            if (allEditor[i].protyle.element.contains(topSourceParentElement)) {
                                const otherSbData = await cancelSB(allEditor[i].protyle, topSourceParentElement);
                                doOperations.push(otherSbData.doOperations[0], otherSbData.doOperations[1]);
                                undoOperations.push(otherSbData.undoOperations[1], otherSbData.undoOperations[0]);
                                break;
                            }
                        }
                    }
                }
            } else if (oldSourceParentElement.classList.contains("sb") && getSbChildBlockCount(oldSourceParentElement) === 1) {
                if (isSameDoc) {
                    const sbData = await cancelSB(protyle, oldSourceParentElement);
                    doOperations.push(sbData.doOperations[0], sbData.doOperations[1]);
                    undoOperations.push(sbData.undoOperations[1], sbData.undoOperations[0]);
                } else {
                    const allEditor = getAllEditor();
                    for (let i = 0; i < allEditor.length; i++) {
                        if (allEditor[i].protyle.element.contains(oldSourceParentElement)) {
                            const otherSbData = await cancelSB(allEditor[i].protyle, oldSourceParentElement);
                            doOperations.push(otherSbData.doOperations[0], otherSbData.doOperations[1]);
                            undoOperations.push(otherSbData.undoOperations[1], otherSbData.undoOperations[0]);
                            break;
                        }
                    }
                }
            } else if (oldSourceParentElement.classList.contains("protyle-wysiwyg") && oldSourceParentElement.childElementCount === 0) {
                getAllEditor().find(item => {
                    if (item.protyle.element.contains(oldSourceParentElement)) {
                        if (!item.protyle.block.showAll) {
                            const newId = Lute.NewNodeID();
                            doOperations.splice(0, 0, {
                                action: "insert",
                                id: newId,
                                data: genEmptyElement(false, false, newId).outerHTML,
                                parentID: item.protyle.block.parentID
                            });
                            undoOperations.splice(0, 0, {
                                action: "delete",
                                id: newId,
                            });
                        } else {
                            zoomOut({protyle: item.protyle, id: item.protyle.block.rootID});
                        }
                        return true;
                    }
                });
            }
        }

        if (!isCopy && item.getAttribute("data-type") === "NodeListItem" && targetElement.getAttribute("data-type") === "NodeListItem") {
            const targetSubtype = targetElement.getAttribute("data-subtype");
            if (targetSubtype && item.getAttribute("data-subtype") !== targetSubtype) {
                const originalHTML = item.outerHTML;
                convertListItemSubtype(item, targetSubtype);
                doOperations.push({
                    action: "update",
                    id,
                    data: item.outerHTML,
                });
                undoOperations.push({
                    action: "update",
                    id,
                    data: originalHTML,
                });
            }
        }

        if (newListId && (index === 0 ||
            sourceElements[index - 1].getAttribute("data-type") !== "NodeListItem" ||
            sourceElements[index - 1].getAttribute("data-subtype") !== originalSubtype)
        ) {
            if (position === "beforebegin") {
                tempTargetElement = newListElement;
            }
            newListId = null;
            if (newListElement.getAttribute("data-subtype") === "o" && newListElement.firstElementChild.getAttribute("data-marker") !== "1.") {
                Array.from(newListElement.children).forEach((listItem) => {
                    if (listItem.classList.contains("protyle-attr")) {
                        return;
                    }
                    undoOperations.push({
                        action: "update",
                        id: listItem.getAttribute("data-node-id"),
                        data: listItem.outerHTML
                    });
                });
                updateListOrder(newListElement, 1);
                Array.from(newListElement.children).forEach((listItem) => {
                    if (listItem.classList.contains("protyle-attr")) {
                        return;
                    }
                    doOperations.push({
                        action: "update",
                        id: listItem.getAttribute("data-node-id"),
                        data: listItem.outerHTML
                    });
                });
                updateListOrder(newListElement, 1);
            }
        } else if (position === "beforebegin") {
            tempTargetElement = isCopy ? copyElement : item;
        }
    }
    Object.keys(orderListElements).forEach(key => {
        Array.from(orderListElements[key].children).forEach((item) => {
            if (item.classList.contains("protyle-attr")) {
                return;
            }
            undoOperations.push({
                action: "update",
                id: item.getAttribute("data-node-id"),
                data: item.outerHTML
            });
        });
        updateListOrder(orderListElements[key], 1);
        Array.from(orderListElements[key].children).forEach((item) => {
            if (item.classList.contains("protyle-attr")) {
                return;
            }
            doOperations.push({
                action: "update",
                id: item.getAttribute("data-node-id"),
                data: item.outerHTML
            });
        });
    });
    undoOperations.reverse();
    for (let j = 0; j < copyFoldHeadingIds.length; j++) {
        const childrenItem = copyFoldHeadingIds[j];
        const responseTransaction = await fetchSyncPost("/api/block/getHeadingInsertTransaction", {id: childrenItem.oldId});
        responseTransaction.data.doOperations.splice(0, 1);
        responseTransaction.data.doOperations[0].previousID = childrenItem.newId;
        responseTransaction.data.undoOperations.splice(0, 1);
        doOperations.push(...responseTransaction.data.doOperations);
        undoOperations.push(...responseTransaction.data.undoOperations);
    }
    return {
        doOperations,
        undoOperations,
        newSourceElements
    };
};

const dragSb = async (protyle: IProtyle, sourceElements: Element[], targetElement: Element, isBottom: boolean,
                      direct: "col" | "row", isCopy: boolean) => {
    const isSameDoc = protyle.element.contains(sourceElements[0]);
    const originSbSet = new Set<Element>();
    sourceElements.forEach(el => {
        const sb = el.closest('[data-type="NodeSuperBlock"]');
        if (sb && sb !== targetElement.closest('[data-type="NodeSuperBlock"]')) {
            originSbSet.add(sb);
        }
    });
    if (isSameDoc && sourceElements[0].classList.contains("li") && targetElement === sourceElements[0].parentElement &&
        targetElement.childElementCount === sourceElements.length + 1) {
        const outLiElement = sourceElements.find((element) => {
            if (!targetElement.contains(element)) {
                return true;
            }
        });
        if (!outLiElement) {
            return;
        }
    }
    const undoOperations: IOperation[] = [];
    const targetMoveUndo: IOperation = {
        action: "move",
        context: {
            removeFold: "true"
        },
        id: targetElement.getAttribute("data-node-id"),
        previousID: getPreviousBlockSibling(targetElement)?.getAttribute("data-node-id"),
        parentID: getParentBlock(targetElement)?.getAttribute("data-node-id") || protyle.block.parentID || protyle.block.rootID
    };
    const sbElement = genSBElement(direct);
    targetElement.parentElement.replaceChild(sbElement, targetElement);
    const doOperations: IOperation[] = [{
        action: "insert",
        data: sbElement.outerHTML,
        id: sbElement.getAttribute("data-node-id"),
        nextID: getNextBlockSibling(sbElement)?.getAttribute("data-node-id"),
        previousID: getPreviousBlockSibling(sbElement)?.getAttribute("data-node-id"),
        parentID: getParentBlock(sbElement)?.getAttribute("data-node-id") || protyle.block.parentID || protyle.block.rootID
    }];
    sbElement.lastElementChild.before(targetElement);
    const moveToResult = await moveTo(protyle, sourceElements, sbElement, isSameDoc, "afterbegin", isCopy);
    doOperations.push(...moveToResult.doOperations);
    undoOperations.push(...moveToResult.undoOperations);
    const newSourceParentElement = moveToResult.newSourceElements;
    let removeIndex = doOperations.length;
    doOperations.find((item, index) => {
        if (item.action === "delete" && item.id === targetMoveUndo.parentID) {
            removeIndex = index;
        }
        if (item.action === "delete" && item.id === targetElement.getAttribute("data-node-id")) {
            targetElement = sbElement.querySelector(`[data-node-id="${doOperations[index - 1].id}"]`);
        }
    });

    if (isBottom) {
        sbElement.insertAdjacentElement("afterbegin", targetElement);
        doOperations.splice(removeIndex, 0, {
            action: "move",
            id: targetElement.getAttribute("data-node-id"),
            parentID: sbElement.getAttribute("data-node-id")
        });
    } else {
        sbElement.lastElementChild.insertAdjacentElement("beforebegin", targetElement);
        doOperations.splice(removeIndex, 0, {
            action: "move",
            id: targetElement.getAttribute("data-node-id"),
            previousID: newSourceParentElement[0].getAttribute("data-node-id"),
        });
    }
    undoOperations.push(targetMoveUndo);
    undoOperations.push({
        action: "delete",
        id: sbElement.getAttribute("data-node-id"),
    });
    const foldElements: Element[] = [];
    newSourceParentElement.forEach(item => {
        const nextBlockElement = getNextBlockSibling(item);
        if (item.getAttribute("data-type") === "NodeHeading" && item.getAttribute("fold") === "1" &&
            nextBlockElement && (
                nextBlockElement.getAttribute("data-type") !== "NodeHeading" ||
                (nextBlockElement.getAttribute("data-subtype") || "") > item.getAttribute("data-subtype")
            )) {
            foldElements.push(item);
        }
    });
    if ((newSourceParentElement.length > 1 || foldElements.length > 0) && direct === "col") {
        const mergeOperations = await turnsIntoOneTransaction({
            protyle,
            selectsElement: newSourceParentElement.reverse(),
            type: "BlocksMergeSuperBlock",
            level: "row",
            unfocus: true,
            getOperations: true
        });
        doOperations.push(...mergeOperations.doOperations);
        undoOperations.splice(0, 0, ...mergeOperations.undoOperations);
    }
    foldElements.forEach(item => {
        const foldOperations = setFold(protyle, item, true, false, false, true);
        doOperations.push(...foldOperations.doOperations);
        undoOperations.splice(0, 0, ...foldOperations.undoOperations);
    });
    refreshSbResize(sbElement);
    originSbSet.forEach(sb => {
        refreshSbAndPersistWidth(sb, doOperations, undoOperations);
    });
    transaction(protyle, doOperations, undoOperations);
    if (document.contains(sourceElements[0])) {
        focusBlock(sourceElements[0]);
    } else {
        focusBlock(targetElement);
    }
};

const dragSame = async (protyle: IProtyle, sourceElements: Element[], targetElement: Element, isBottom: boolean, isCopy: boolean) => {
    const isSameDoc = protyle.element.contains(sourceElements[0]);
    const doOperations: IOperation[] = [];
    const undoOperations: IOperation[] = [];
    const originSbSet = new Set<Element>();
    sourceElements.forEach(el => {
        const sb = el.closest('[data-type="NodeSuperBlock"]');
        if (sb) {
            originSbSet.add(sb);
        }
    });

    const moveToResult = await moveTo(protyle, sourceElements, targetElement, isSameDoc, isBottom ? "afterend" : "beforebegin", isCopy);
    doOperations.push(...moveToResult.doOperations);
    undoOperations.push(...moveToResult.undoOperations);
    const newSourceParentElement = moveToResult.newSourceElements;
    let foldData;
    const previousBlockElement = getPreviousBlockSibling(targetElement);
    if (isBottom &&
        targetElement.getAttribute("data-type") === "NodeHeading" &&
        targetElement.getAttribute("fold") === "1") {
        foldData = setFold(protyle, targetElement, true, false, false, true);
    } else if (!isBottom &&
        previousBlockElement?.getAttribute("data-type") === "NodeHeading" &&
        previousBlockElement.getAttribute("fold") === "1") {
        foldData = setFold(protyle, previousBlockElement, true, false, false, true);
    }
    if (foldData) {
        foldData.doOperations[0].context = {
            focusId: sourceElements[0].getAttribute("data-node-id"),
        };
        doOperations.push(...foldData.doOperations);
        undoOperations.push(...foldData.undoOperations);
    }
    if (targetElement.getAttribute("data-type") === "NodeListItem" &&
        targetElement.getAttribute("data-subtype") === "o") {
        // 
        Array.from(targetElement.parentElement.children).forEach((item) => {
            if (item.classList.contains("protyle-attr")) {
                return;
            }
            undoOperations.splice(0, 0, {
                action: "update",
                id: item.getAttribute("data-node-id"),
                data: item.outerHTML
            });
        });
        updateListOrder(targetElement.parentElement, 1);
        Array.from(targetElement.parentElement.children).forEach((item) => {
            if (item.classList.contains("protyle-attr")) {
                return;
            }
            doOperations.push({
                action: "update",
                id: item.getAttribute("data-node-id"),
                data: item.outerHTML
            });
        });
    }
    let hasFoldHeading = false;
    newSourceParentElement.forEach(item => {
        if (item.getAttribute("data-type") === "NodeHeading" && item.getAttribute("fold") === "1") {
            hasFoldHeading = true;
            const nextBlockElement = getNextBlockSibling(item);
            if (nextBlockElement && (
                nextBlockElement.getAttribute("data-type") !== "NodeHeading" ||
                nextBlockElement.getAttribute("data-subtype") > item.getAttribute("data-subtype")
            )) {
                const foldOperations = setFold(protyle, item, true, false, false, true);
                doOperations.push(...foldOperations.doOperations);
            }
            return true;
        }
    });
    const dragSbSet = new Set<Element>(originSbSet);
    [newSourceParentElement[0], targetElement].forEach(el => {
        const sb = el?.closest('[data-type="NodeSuperBlock"]');
        if (sb) {
            dragSbSet.add(sb);
        }
    });
    dragSbSet.forEach(sb => {
        refreshSbAndPersistWidth(sb, doOperations, undoOperations);
    });
    if ((newSourceParentElement.length > 1 || hasFoldHeading) &&
        newSourceParentElement[0].parentElement.classList.contains("sb") &&
        newSourceParentElement[0].parentElement.getAttribute("data-sb-layout") === "col") {
        const mergeOperations = await turnsIntoOneTransaction({
            protyle,
            selectsElement: newSourceParentElement.reverse(),
            type: "BlocksMergeSuperBlock",
            level: "row",
            unfocus: true,
            getOperations: true
        });
        doOperations.push(...mergeOperations.doOperations);
        undoOperations.splice(0, 0, ...mergeOperations.undoOperations);
    }
    transaction(protyle, doOperations, undoOperations);
    if (document.contains(sourceElements[0])) {
        focusBlock(sourceElements[0]);
    } else {
        focusBlock(targetElement);
    }
};

export const dropEvent = (protyle: IProtyle, editorElement: HTMLElement) => {
    editorElement.addEventListener("dragstart", (event) => {
        if (protyle.disabled) {
            event.preventDefault();
            event.stopPropagation();
            return;
        }
        let target = event.target as HTMLElement;
        if (target.classList?.contains("av__gallery-img")) {
            target = hasClosestByClassName(target, "av__gallery-item") as HTMLElement;
        }
        if (!target) {
            return;
        }
        if (target.tagName === "IMG") {
            window.scribli.dragElement = undefined;
            event.preventDefault();
            return;
        }

        if (target.classList) {
            if (hasClosestByClassName(target, "protyle-wysiwyg__embed")) {
                window.scribli.dragElement = undefined;
                event.preventDefault();
            } else if (target.parentElement.parentElement.classList.contains("av__views")) {
                window.scribli.dragElement = target;
                target.style.width = target.clientWidth + "px";
                target.style.opacity = ".36";
                event.dataTransfer.setData(`${Constants.SCRIBLI_DROP_GUTTER}NodeAttributeView${Constants.ZWSP}ViewTab${Constants.ZWSP}${[target.previousElementSibling?.getAttribute("data-id")]}`,
                    target.outerHTML);
                return;
            } else if (target.classList.contains("protyle-action")) {
                target.parentElement.classList.add("protyle-wysiwyg--select");
                const ghostElement = document.createElement("div");
                ghostElement.className = protyle.wysiwyg.element.className;
                const cloneElement = processClonePHElement(target.parentElement.cloneNode(true) as Element);
                cloneElement.querySelectorAll(".iframe").forEach(item => {
                    item.remove();
                });
                ghostElement.append(cloneElement);
                ghostElement.setAttribute("style", `position:fixed;opacity:.1;width:${target.parentElement.clientWidth}px;padding:0;`);
                document.body.append(ghostElement);
                if (window.scribli.touchDragActive) {
                    event.dataTransfer.setDragImage(ghostElement, 0, 0);
                    window.scribli.touchDragGhost = ghostElement;
                } else {
                    const transparentImg = new Image();
                    transparentImg.src = transparentImgSrc;
                    event.dataTransfer.setDragImage(transparentImg, 0, 0);
                    setTimeout(() => {
                        ghostElement.remove();
                    });
                }
                window.scribli.dragTitle = getContenteditableElement(target.parentElement)?.textContent?.trim() || "";

                window.scribli.dragElement = protyle.wysiwyg.element;
                event.dataTransfer.setData(`${Constants.SCRIBLI_DROP_GUTTER}NodeListItem${Constants.ZWSP}${target.parentElement.getAttribute("data-subtype")}${Constants.ZWSP}${[target.parentElement.getAttribute("data-node-id")]}`,
                    protyle.wysiwyg.element.innerHTML);
                return;
            } else if (target.classList.contains("av__cell--header")) {
                window.scribli.dragElement = target;
                event.dataTransfer.setData(`${Constants.SCRIBLI_DROP_GUTTER}NodeAttributeView${Constants.ZWSP}Col${Constants.ZWSP}${[target.getAttribute("data-col-id")]}`,
                    target.outerHTML);
                return;
            } else if (target.classList.contains("av__gallery-item")) {
                const blockElement = hasClosestBlock(target);
                if (blockElement) {
                    if (blockElement.querySelector('.block__icon[data-type="av-sort"]')?.classList.contains("block__icon--active")) {
                        const bodyElements = blockElement.querySelectorAll(".av__body");
                        if (bodyElements.length === 1) {
                            event.preventDefault();
                            event.stopPropagation();
                            return;
                        } else if (["template", "created", "updated"].includes(bodyElements[0].getAttribute("data-dtype"))) {
                            event.preventDefault();
                            event.stopPropagation();
                            return;
                        }
                    }
                    if (!target.classList.contains("av__gallery-item--select")) {
                        blockElement.querySelectorAll(".av__gallery-item--select").forEach(item => {
                            item.classList.remove("av__gallery-item--select");
                        });
                        target.classList.add("av__gallery-item--select");
                    }
                    const ghostElement = document.createElement("div");
                    ghostElement.className = "protyle-wysiwyg protyle-wysiwyg--attr";
                    const isKanban = blockElement.getAttribute("data-av-type") === "kanban";
                    if (isKanban) {
                        ghostElement.innerHTML = `<div class="${blockElement.querySelector(".av__kanban").className}"></div>`;
                    }
                    let galleryElement: HTMLElement;
                    let cloneGalleryElement = document.createElement("div");
                    const selectElements = blockElement.querySelectorAll(".av__gallery-item--select");
                    selectElements.forEach(item => {
                        if (!galleryElement || !galleryElement.contains(item)) {
                            galleryElement = item.parentElement;
                            cloneGalleryElement = document.createElement("div");
                            if (isKanban) {
                                cloneGalleryElement.className = "av__kanban-group";
                                cloneGalleryElement.setAttribute("style", item.parentElement.parentElement.parentElement.getAttribute("style"));
                                cloneGalleryElement.innerHTML = '<div class="av__gallery"></div>';
                                ghostElement.firstElementChild.appendChild(cloneGalleryElement);
                            } else {
                                cloneGalleryElement.classList.add("av__gallery");
                                cloneGalleryElement.setAttribute("style", `width: 100vw;margin-bottom: 16px;grid-template-columns: repeat(auto-fill, ${selectElements[0].clientWidth}px);`);
                                ghostElement.appendChild(cloneGalleryElement);
                            }
                        }
                        const cloneItem = processClonePHElement(item.cloneNode(true) as Element);
                        cloneItem.setAttribute("style", `height:${item.clientHeight}px;`);
                        cloneItem.classList.remove("av__gallery-item--select");
                        if (isKanban) {
                            cloneGalleryElement.firstElementChild.appendChild(cloneItem);
                        } else {
                            cloneGalleryElement.appendChild(cloneItem);
                        }
                    });
                    ghostElement.setAttribute("style", "left: 1px;top:100vh;position:fixed;opacity:.1;padding:0;z-index: 8");
                    document.body.append(ghostElement);
                    event.dataTransfer.setDragImage(ghostElement, -10, -10);
                    if (window.scribli.touchDragActive) {
                        window.scribli.touchDragGhost = ghostElement;
                    } else {
                        setTimeout(() => {
                            ghostElement.remove();
                        });
                    }
                    window.scribli.dragElement = target;
                    const selectIds: string[] = [];
                    blockElement.querySelectorAll(".av__gallery-item--select").forEach(item => {
                        const bodyElement = hasClosestByClassName(item, "av__body") as HTMLElement;
                        const groupId = bodyElement.getAttribute("data-group-id");
                        selectIds.push(item.getAttribute("data-id") + (groupId ? `@${groupId}` : ""));
                    });
                    event.dataTransfer.setData(`${Constants.SCRIBLI_DROP_GUTTER}NodeAttributeView${Constants.ZWSP}GalleryItem${Constants.ZWSP}${selectIds}`,
                        ghostElement.outerHTML);
                }
                return;
            }
        }
        event.dataTransfer.setData(Constants.SCRIBLI_DROP_EDITOR, Constants.SCRIBLI_DROP_EDITOR);
        protyle.element.style.userSelect = "auto";
        document.onmousemove = null;
        document.onmouseup = null;
    });
    const insertBlockRefs = async (ids: string[]) => {
        let html = "";
        for (const id of ids) {
            const response = await fetchSyncPost("/api/block/getRefText", {id});
            html += protyle.lute.Md2BlockDOM(`((${id} '${response.data}'))`);
        }
        insertHTML(html, protyle);
    };
    const focusBlockRefDrop = (event: DragEvent) => {
        if (event.y > protyle.wysiwyg.element.lastElementChild.getBoundingClientRect().bottom) {
            insertEmptyBlock(protyle, "afterend", protyle.wysiwyg.element.lastElementChild.getAttribute("data-node-id"));
            return true;
        }
        const range = getRangeByPoint(event.clientX, event.clientY);
        if (!range || hasClosestByAttribute(range.startContainer, "data-type", "NodeBlockQueryEmbed")) {
            return false;
        }
        focusByRange(range);
        return true;
    };
    const renderBlockRefDragover = (event: DragEvent) => {
        cleanupDragIndicators(editorElement);
        editorElement.querySelectorAll("[select-start], [select-end]").forEach((item: HTMLElement) => {
            item.removeAttribute("select-start");
            item.removeAttribute("select-end");
        });
        if (event.y <= protyle.wysiwyg.element.lastElementChild.getBoundingClientRect().bottom) {
            const range = getRangeByPoint(event.clientX, event.clientY);
            if (range && !hasClosestByAttribute(range.startContainer, "data-type", "NodeBlockQueryEmbed")) {
                const rect = range.getBoundingClientRect();
                if (rect.height > 0) {
                    showCaretLine(rect.left, rect.top, rect.height);
                }
            }
        } else {
            hideCaretLine();
            const lastBlock = protyle.wysiwyg.element.lastElementChild as HTMLElement;
            if (lastBlock?.hasAttribute("data-node-id")) {
                lastBlock.classList.add("dragover__bottom");
            }
        }
        event.preventDefault();
    };
    editorElement.addEventListener("drop", async (event: DragEvent & { target: HTMLElement }) => {
        const isCopyDrag = protyle.lite || event.ctrlKey;
        counter = 0;
        hideDragTip();
        window.scribli.dragTitle = "";
        if (protyle.disabled || event.dataTransfer.getData(Constants.SCRIBLI_DROP_EDITOR)) {
            event.preventDefault();
            event.stopPropagation();
            return;
        }
        if (event.dataTransfer.types.includes(Constants.SCRIBLI_DROP_BLOCK_REF)) {
            event.preventDefault();
            event.stopPropagation();
            let ids: string[] = [];
            try {
                const data = JSON.parse(event.dataTransfer.getData(Constants.SCRIBLI_DROP_BLOCK_REF));
                if (data.workspaceDir?.toLowerCase() !== window.scribli.config.system.workspaceDir.toLowerCase()) {
                    cleanupDragIndicators(editorElement);
                    return;
                }
                ids = Array.from(new Set((Array.isArray(data.ids) ? data.ids : [])
                    .filter((id: unknown): id is string => typeof id === "string" && /^\d{14}-[0-9a-z]{7}$/.test(id))));
            } catch (e) {
                console.warn("parse block reference drop data failed", e);
            }
            if (ids.length === 0 || hasClosestByClassName(event.target, "av") || !focusBlockRefDrop(event)) {
                cleanupDragIndicators(editorElement);
                return;
            }
            await insertBlockRefs(ids);
            cleanupDragIndicators(editorElement);
            return;
        }
        let gutterType = "";
        for (const type of event.dataTransfer.types) {
            if (type.startsWith(Constants.SCRIBLI_DROP_GUTTER)) {
                gutterType = type;
            }
        }
        if (gutterType.startsWith(`${Constants.SCRIBLI_DROP_GUTTER}NodeAttributeView${Constants.ZWSP}ViewTab${Constants.ZWSP}`.toLowerCase())) {
            const blockElement = hasClosestBlock(window.scribli.dragElement);
            if (blockElement) {
                const avID = blockElement.getAttribute("data-av-id");
                const blockID = blockElement.getAttribute("data-node-id");
                const id = window.scribli.dragElement.getAttribute("data-id");
                transaction(protyle, [{
                    action: "sortAttrViewView",
                    avID,
                    blockID,
                    id,
                    previousID: window.scribli.dragElement.previousElementSibling?.getAttribute("data-id"),
                    data: "unRefresh"
                }], [{
                    action: "sortAttrViewView",
                    avID,
                    blockID,
                    id,
                    previousID: gutterType.split(Constants.ZWSP).pop()
                }]);
            }
            return;
        }
        let targetElement = editorElement.querySelector(".dragover__left, .dragover__right, .dragover__bottom, .dragover__top, .dragover__bottom--sibling, .dragover__top--sibling, .dragover__bottom--child, .dragover__top--child");
        if (targetElement) {
            targetElement.classList.remove("dragover");
            targetElement.removeAttribute("select-start");
            targetElement.removeAttribute("select-end");
        }
        if (gutterType) {
            const sourceElements: Element[] = [];
            const gutterTypes = gutterType.replace(Constants.SCRIBLI_DROP_GUTTER, "").split(Constants.ZWSP);
            const selectedIds = gutterTypes[2].split(",");
            if (event.altKey || event.shiftKey) {
                if (event.y > protyle.wysiwyg.element.lastElementChild.getBoundingClientRect().bottom) {
                    insertEmptyBlock(protyle, "afterend", protyle.wysiwyg.element.lastElementChild.getAttribute("data-node-id"));
                } else {
                    const range = getRangeByPoint(event.clientX, event.clientY);
                    if (hasClosestByAttribute(range.startContainer, "data-type", "NodeBlockQueryEmbed")) {
                        return;
                    } else {
                        focusByRange(range);
                    }
                }
            }
            if (event.altKey || (event.shiftKey && protyle.lite)) {
                await insertBlockRefs(selectedIds);
            } else if (event.shiftKey) {
                let html = "";
                selectedIds.forEach(item => {
                    html += `{{select * from blocks where id='${item}'}}\n`;
                });
                insertHTML(protyle.lute.SpinBlockDOM(html), protyle, true);
                blockRender(protyle, protyle.wysiwyg.element);
            } else if (targetElement && targetElement.className.indexOf("dragover__") > -1) {
                let queryClass = "";
                selectedIds.forEach(item => {
                    queryClass += `[data-node-id="${item}"],`;
                });
                if (window.scribli.dragElement) {
                    window.scribli.dragElement.querySelectorAll(queryClass.substring(0, queryClass.length - 1)).forEach(elementItem => {
                        if (!isInEmbedBlock(elementItem)) {
                            sourceElements.push(elementItem);
                        }
                    });
                } else if (window.scribli.config.system.workspaceDir.toLowerCase() === gutterTypes[3]) {
                    const targetProtyleElement = document.createElement("template");
                    targetProtyleElement.innerHTML = `<div>${event.dataTransfer.getData(gutterType)}</div>`;
                    targetProtyleElement.content.querySelectorAll(queryClass.substring(0, queryClass.length - 1)).forEach(elementItem => {
                        if (!isInEmbedBlock(elementItem)) {
                            sourceElements.push(elementItem);
                        }
                    });
                }

                const sourceIds: string [] = [];
                const srcs: IOperationSrcs[] = [];
                sourceElements.forEach(item => {
                    item.classList.remove("protyle-wysiwyg--hl");
                    item.removeAttribute("select-start");
                    item.removeAttribute("select-end");
                    item.querySelectorAll('[data-type="search-mark"]').forEach(markItem => {
                        markItem.outerHTML = markItem.innerHTML;
                    });
                    const id = item.getAttribute("data-node-id");
                    sourceIds.push(id);
                    srcs.push({
                        itemID: Lute.NewNodeID(),
                        id,
                        isDetached: false,
                    });
                });

                hideElements(["gutter"], protyle);

                const targetClass = targetElement.className.split(" ");
                targetElement.classList.remove("dragover__bottom", "dragover__top", "dragover__left", "dragover__right",
                    "dragover__bottom--sibling", "dragover__top--sibling", "dragover__bottom--child", "dragover__top--child");
                (targetElement as HTMLElement).style.removeProperty("--drag-indent");
                (targetElement as HTMLElement).style.removeProperty("--drag-guides");
                (targetElement as HTMLElement).style.removeProperty("--drag-line-left");
                (targetElement as HTMLElement).style.removeProperty("--drag-base-bg");
                (targetElement as HTMLElement).style.removeProperty("--drag-base-bg");

                if (targetElement.classList.contains("av__cell")) {
                    const blockElement = hasClosestBlock(targetElement);
                    if (blockElement) {
                        const avID = blockElement.getAttribute("data-av-id");
                        let previousID = "";
                        if (targetClass.includes("dragover__left")) {
                            if (targetElement.previousElementSibling) {
                                if (targetElement.previousElementSibling.classList.contains("av__colsticky")) {
                                    previousID = targetElement.previousElementSibling.lastElementChild.getAttribute("data-col-id");
                                } else {
                                    previousID = targetElement.previousElementSibling.getAttribute("data-col-id");
                                }
                            }
                        } else {
                            previousID = targetElement.getAttribute("data-col-id");
                        }
                        let oldPreviousID = "";
                        const rowElement = hasClosestByClassName(targetElement, "av__row");
                        if (rowElement) {
                            const oldPreviousElement = rowElement.querySelector(`[data-col-id="${gutterTypes[2]}"`)?.previousElementSibling;
                            if (oldPreviousElement) {
                                if (oldPreviousElement.classList.contains("av__colsticky")) {
                                    oldPreviousID = oldPreviousElement.lastElementChild.getAttribute("data-col-id");
                                } else {
                                    oldPreviousID = oldPreviousElement.getAttribute("data-col-id");
                                }
                            }
                        }
                        if (previousID !== oldPreviousID && previousID !== gutterTypes[2]) {
                            transaction(protyle, [{
                                action: "sortAttrViewCol",
                                avID,
                                previousID,
                                id: gutterTypes[2],
                                blockID: blockElement.dataset.nodeId,
                            }], [{
                                action: "sortAttrViewCol",
                                avID,
                                previousID: oldPreviousID,
                                id: gutterTypes[2],
                                blockID: blockElement.dataset.nodeId,
                            }]);
                        }
                    }
                } else if (targetElement.classList.contains("av__row")) {
                    const blockElement = hasClosestBlock(targetElement);
                    if (blockElement) {
                        let previousID = "";
                        if (targetClass.includes("dragover__bottom")) {
                            previousID = targetElement.getAttribute("data-id") || "";
                        } else {
                            previousID = targetElement.previousElementSibling?.getAttribute("data-id") || "";
                        }
                        const avID = blockElement.getAttribute("data-av-id");
                        if (gutterTypes[0] === "nodeattributeviewrowmenu") {
                            const doOperations: IOperation[] = [];
                            const undoOperations: IOperation[] = [];
                            const targetGroupID = targetElement.parentElement.getAttribute("data-group-id");
                            selectedIds.reverse().forEach(item => {
                                const items = item.split("@");
                                const id = items[0];
                                const groupID = items[1] || "";
                                const undoPreviousId = blockElement.querySelector(`.av__body${groupID ? `[data-group-id="${groupID}"]` : ""} .av__row[data-id="${id}"]`).previousElementSibling?.getAttribute("data-id") || "";
                                if (previousID !== id && undoPreviousId !== previousID || (
                                    (undoPreviousId === "" && previousID === "" && targetGroupID !== groupID)
                                )) {
                                    doOperations.push({
                                        action: "sortAttrViewRow",
                                        avID,
                                        previousID,
                                        id,
                                        blockID: blockElement.dataset.nodeId,
                                        groupID,
                                        targetGroupID,
                                    });
                                    undoOperations.push({
                                        action: "sortAttrViewRow",
                                        avID,
                                        previousID: undoPreviousId,
                                        id,
                                        blockID: blockElement.dataset.nodeId,
                                        groupID: targetGroupID,
                                        targetGroupID: groupID,
                                    });
                                }
                            });
                            transaction(protyle, doOperations, undoOperations);
                        } else {
                            const newUpdated = dayjs().format("YYYYMMDDHHmmss");
                            const bodyElement = hasClosestByClassName(targetElement, "av__body");
                            const groupID = bodyElement && bodyElement.getAttribute("data-group-id");
                            transaction(protyle, [{
                                action: "insertAttrViewBlock",
                                avID,
                                previousID,
                                srcs,
                                blockID: blockElement.dataset.nodeId,
                                groupID
                            }, {
                                action: "doUpdateUpdated",
                                id: blockElement.dataset.nodeId,
                                data: newUpdated,
                            }], [{
                                action: "removeAttrViewBlock",
                                srcIDs: sourceIds,
                                avID,
                            }, {
                                action: "doUpdateUpdated",
                                id: blockElement.dataset.nodeId,
                                data: blockElement.getAttribute("updated")
                            }]);
                            blockElement.setAttribute("updated", newUpdated);
                            insertAttrViewBlockAnimation({
                                protyle,
                                blockElement,
                                srcIDs: sourceIds,
                                previousId: previousID,
                                groupID
                            });
                        }
                    }
                } else if (targetElement.classList.contains("av__gallery-item") || targetElement.classList.contains("av__gallery-add")) {
                    const blockElement = hasClosestBlock(targetElement);
                    if (blockElement) {
                        let previousID = "";
                        if (targetClass.includes("dragover__right") || targetClass.includes("dragover__bottom")) {
                            previousID = targetElement.getAttribute("data-id") || "";
                        } else if (targetClass.includes("dragover__top") || targetClass.includes("dragover__left")) {
                            previousID = targetElement.previousElementSibling?.getAttribute("data-id") || "";
                        }
                        const avID = blockElement.getAttribute("data-av-id");
                        if (gutterTypes[1] === "galleryitem" && gutterTypes[0] === "nodeattributeview") {
                            const doOperations: IOperation[] = [];
                            const undoOperations: IOperation[] = [];
                            const targetGroupID = targetElement.parentElement.parentElement.getAttribute("data-group-id");
                            selectedIds.reverse().forEach(item => {
                                const items = item.split("@");
                                const id = items[0];
                                const groupID = items[1] || "";
                                const undoPreviousId = blockElement.querySelector(`.av__body[data-group-id="${groupID}"] .av__gallery-item[data-id="${id}"]`).previousElementSibling?.getAttribute("data-id") || "";
                                if (previousID !== item && undoPreviousId !== previousID || (
                                    (undoPreviousId === "" && previousID === "" && targetGroupID !== groupID)
                                )) {
                                    doOperations.push({
                                        action: "sortAttrViewRow",
                                        avID,
                                        previousID,
                                        id,
                                        blockID: blockElement.dataset.nodeId,
                                        groupID,
                                        targetGroupID,
                                    });
                                    undoOperations.push({
                                        action: "sortAttrViewRow",
                                        avID,
                                        previousID: undoPreviousId,
                                        id,
                                        blockID: blockElement.dataset.nodeId,
                                        groupID: targetGroupID,
                                        targetGroupID: groupID,
                                    });
                                }
                            });
                            transaction(protyle, doOperations, undoOperations);
                        } else {
                            const newUpdated = dayjs().format("YYYYMMDDHHmmss");
                            const bodyElement = hasClosestByClassName(targetElement, "av__body");
                            transaction(protyle, [{
                                action: "insertAttrViewBlock",
                                avID,
                                previousID,
                                srcs,
                                blockID: blockElement.dataset.nodeId,
                                groupID: bodyElement && bodyElement.getAttribute("data-group-id")
                            }, {
                                action: "doUpdateUpdated",
                                id: blockElement.dataset.nodeId,
                                data: newUpdated,
                            }], [{
                                action: "removeAttrViewBlock",
                                srcIDs: sourceIds,
                                avID,
                            }, {
                                action: "doUpdateUpdated",
                                id: blockElement.dataset.nodeId,
                                data: blockElement.getAttribute("updated")
                            }]);
                            blockElement.setAttribute("updated", newUpdated);
                            insertGalleryItemAnimation({
                                protyle,
                                blockElement,
                                srcIDs: sourceIds,
                                previousId: previousID,
                                groupID: targetElement.parentElement.getAttribute("data-group-id")
                            });
                        }
                    }
                } else if (sourceElements.length > 0) {
                    const isChild = targetClass.some((c: string) => c.indexOf("--child") > -1);
                    const isBottom = targetClass.some((c: string) => c.indexOf("dragover__bottom") === 0);

                    const isListItemSource = gutterTypes[0] === "nodelistitem" || gutterTypes[0] === "nodelist";
                    if (isListItemSource) {
                        if (targetElement.classList.contains("list") &&
                            sourceElements.some(s => targetElement.contains(s))) {
                            dragoverElement = undefined;
                            return;
                        }
                        const targetLi = getTargetListItem(targetElement, isBottom);
                        if (targetLi) {
                            const isNoOpDrop = sourceElements.some(source =>
                                source === targetLi ||
                                source.contains(targetLi) ||
                                (!isChild && isBottom && source === targetLi.nextElementSibling) ||
                                (!isChild && !isBottom && source === targetLi.previousElementSibling));
                            if (isNoOpDrop && !event.ctrlKey && !event.shiftKey && !event.altKey) {
                                dragoverElement = undefined;
                                return;
                            }
                        } else {
                            const sourceSelected = sourceElements[0];
                            if (sourceSelected && (sourceSelected.classList.contains("li") || sourceSelected.classList.contains("list"))) {
                                if (targetElement.classList.contains("list") && targetElement.contains(sourceSelected)) {
                                    dragoverElement = undefined;
                                    return;
                                }
                                let current: Element = sourceSelected;
                                while (current && current !== editorElement) {
                                    if (current.classList.contains("list") || current.classList.contains("li")) {
                                        const checkSiblings = (container: Element) => {
                                            let prevSibling = container.previousElementSibling;
                                            while (prevSibling && prevSibling.classList.contains("protyle-attr")) {
                                                prevSibling = prevSibling.previousElementSibling;
                                            }
                                            let nextSibling = container.nextElementSibling;
                                            while (nextSibling && nextSibling.classList.contains("protyle-attr")) {
                                                nextSibling = nextSibling.nextElementSibling;
                                            }
                                            return targetElement === prevSibling || targetElement === nextSibling;
                                        };
                                        if (checkSiblings(current)) {
                                            if (current.parentElement === editorElement) {
                                                break;
                                            }
                                            dragoverElement = undefined;
                                            return;
                                        }
                                    }
                                    current = current.parentElement;
                                }
                            }
                        }
                    }

                    const isColSbChildList = targetElement.parentElement?.getAttribute("data-type") === "NodeSuperBlock" &&
                        targetElement.parentElement?.getAttribute("data-sb-layout") === "col";
                    if (isListItemSource && targetElement.classList.contains("list") &&
                        !(gutterTypes[0] === "nodelist" && isColSbChildList)) {
                        const targetListItem = getTargetListItem(targetElement, isBottom);
                        if (targetListItem) {
                            targetElement = targetListItem;
                        }
                    }

                    if (targetElement.getAttribute("data-type") === "NodeListItem") {
                        const expandedElements: Element[] = [];
                        sourceElements.forEach(item => {
                            if (item.getAttribute("data-type") === "NodeList") {
                                Array.from(item.children).forEach((li) => {
                                    if (li.classList.contains("li")) {
                                        expandedElements.push(li);
                                    }
                                });
                            } else {
                                expandedElements.push(item);
                            }
                        });
                        if (expandedElements.length > 0) {
                            sourceElements.length = 0;
                            sourceElements.push(...expandedElements);
                        }
                    }
                    const hasContentBlockSource = sourceElements.some(item =>
                        !["NodeList", "NodeListItem"].includes(item.getAttribute("data-type")));

                    if (hasContentBlockSource && !isChild && targetElement.getAttribute("data-type") === "NodeListItem") {
                        const parentLi = targetElement.parentElement?.parentElement;
                        if (targetClass.some((c: string) => c.indexOf("dragover__top--sibling") === 0) &&
                            parentLi?.classList.contains("li")) {
                            const contentLi = parentLi as HTMLElement;
                            const contentBlocks = Array.from(contentLi.children).filter(
                                c => c.hasAttribute("data-node-id") && !c.classList.contains("list"));
                            const anchorBlock = contentBlocks.length > 0 ? contentBlocks[contentBlocks.length - 1] : null;
                            if (anchorBlock) {
                                await dragSame(protyle, sourceElements, anchorBlock, true, isCopyDrag);
                            } else {
                                await dragSame(protyle, sourceElements, contentLi, isBottom, isCopyDrag);
                            }
                            dragoverElement = undefined;
                            return;
                        }
                    }

                    if (hasContentBlockSource && !isChild &&
                        targetElement.getAttribute("data-type") === "NodeListItem") {
                        dragoverElement = undefined;
                        return;
                    }

                    if (isChild && targetElement.getAttribute("data-type") === "NodeListItem") {
                        const nestedList = Array.from(targetElement.children).find(c => c.classList.contains("list"));
                        let nestedTarget: Element;
                        if (nestedList) {
                            const liChildren = Array.from(nestedList.children).filter(c => c.classList.contains("li"));
                            if (isBottom) {
                                nestedTarget = liChildren.length > 0 ? liChildren[liChildren.length - 1] : null;
                            } else {
                                nestedTarget = liChildren.length > 0 ? liChildren[0] : null;
                            }
                        }
                        if (nestedTarget) {
                            if (!sourceElements.includes(nestedTarget)) {
                                dragSame(protyle, sourceElements, nestedTarget, isBottom, isCopyDrag);
                            }
                        } else {
                            const contentBlocks = Array.from(targetElement.children).filter(
                                c => c.hasAttribute("data-node-id") && !c.classList.contains("list"));
                            const lastContentBlock = contentBlocks[contentBlocks.length - 1];
                            if (lastContentBlock) {
                                dragSame(protyle, sourceElements, lastContentBlock, true, isCopyDrag);
                            } else {
                                dragSame(protyle, sourceElements, targetElement, isBottom, isCopyDrag);
                            }
                        }
                    } else if (targetElement.parentElement.getAttribute("data-type") === "NodeSuperBlock" &&
                        targetElement.parentElement.getAttribute("data-sb-layout") === "col") {
                        if (targetClass.includes("dragover__left") || targetClass.includes("dragover__right")) {
                            dragSame(protyle, sourceElements, targetElement, targetClass.includes("dragover__right"), isCopyDrag);
                        } else {
                            dragSb(protyle, sourceElements, targetElement, isBottom, "row", isCopyDrag);
                        }
                    } else {
                        const isListItemOnlySource = gutterTypes[0] === "nodelistitem";
                        const isListTarget = targetElement.classList.contains("list");
                        if (isListItemOnlySource && isListTarget &&
                            (targetClass.includes("dragover__left") || targetClass.includes("dragover__right"))) {
                            // Intentionally empty.
                        } else if (targetClass.includes("dragover__left") || targetClass.includes("dragover__right")) {
                            dragSb(protyle, sourceElements, targetElement, targetClass.includes("dragover__right"), "col", isCopyDrag);
                        } else {
                            dragSame(protyle, sourceElements, targetElement, isBottom, isCopyDrag);
                        }
                    }

                    // 
                    editorElement.querySelectorAll(".protyle-wysiwyg--empty").forEach(item => {
                        item.classList.remove("protyle-wysiwyg--empty");
                    });
                }
                dragoverElement = undefined;
            }
        } else if (event.dataTransfer.getData(Constants.SCRIBLI_DROP_FILE)?.split("-").length > 1) {
            const ids = event.dataTransfer.getData(Constants.SCRIBLI_DROP_FILE).split(",");
            if (!event.altKey && (!targetElement || (
                !targetElement.classList.contains("av__row") && !targetElement.classList.contains("av__gallery-item") &&
                !targetElement.classList.contains("av__gallery-add")
            ))) {
                if (event.y > protyle.wysiwyg.element.lastElementChild.getBoundingClientRect().bottom) {
                    insertEmptyBlock(protyle, "afterend", protyle.wysiwyg.element.lastElementChild.getAttribute("data-node-id"));
                } else {
                    const range = getRangeByPoint(event.clientX, event.clientY);
                    if (hasClosestByAttribute(range.startContainer, "data-type", "NodeBlockQueryEmbed")) {
                        return;
                    } else {
                        focusByRange(range);
                    }
                }
                let html = "";
                for (let i = 0; i < ids.length; i++) {
                    if (ids.length > 1) {
                        html += "- ";
                    }
                    const response = await fetchSyncPost("/api/block/getRefText", {id: ids[i]});
                    html += `((${ids[i]} '${response.data}'))`;
                    if (ids.length > 1 && i !== ids.length - 1) {
                        html += "\n";
                    }
                }
                insertHTML(protyle.lute.Md2BlockDOM(html), protyle);
            } else if (targetElement && !protyle.options.backlinkData && targetElement.className.indexOf("dragover__") > -1) {
                const scrollTop = protyle.contentElement.scrollTop;
                if (targetElement.classList.contains("av__row") ||
                    targetElement.classList.contains("av__gallery-item") ||
                    targetElement.classList.contains("av__gallery-add")) {
                    const blockElement = hasClosestBlock(targetElement);
                    if (blockElement) {
                        let previousID = "";
                        if (targetElement.classList.contains("dragover__bottom") || targetElement.classList.contains("dragover__right")) {
                            previousID = targetElement.getAttribute("data-id") || "";
                        } else if (targetElement.classList.contains("dragover__top") || targetElement.classList.contains("dragover__left")) {
                            previousID = targetElement.previousElementSibling?.getAttribute("data-id") || "";
                        }
                        const avID = blockElement.getAttribute("data-av-id");
                        const newUpdated = dayjs().format("YYYYMMDDHHmmss");
                        const srcs: IOperationSrcs[] = [];
                        const bodyElement = hasClosestByClassName(targetElement, "av__body");
                        const groupID = bodyElement && bodyElement.getAttribute("data-group-id");
                        ids.forEach(id => {
                            srcs.push({
                                itemID: Lute.NewNodeID(),
                                id,
                                isDetached: false,
                            });
                        });
                        transaction(protyle, [{
                            action: "insertAttrViewBlock",
                            avID,
                            previousID,
                            srcs,
                            blockID: blockElement.dataset.nodeId,
                            groupID
                        }, {
                            action: "doUpdateUpdated",
                            id: blockElement.dataset.nodeId,
                            data: newUpdated,
                        }], [{
                            action: "removeAttrViewBlock",
                            srcIDs: ids,
                            avID,
                        }, {
                            action: "doUpdateUpdated",
                            id: blockElement.dataset.nodeId,
                            data: blockElement.getAttribute("updated")
                        }]);
                        insertAttrViewBlockAnimation({
                            protyle,
                            blockElement,
                            srcIDs: ids,
                            previousId: previousID,
                            groupID
                        });
                        blockElement.setAttribute("updated", newUpdated);
                    }
                } else {
                    if (targetElement.classList.contains("dragover__bottom")) {
                        for (let i = ids.length - 1; i > -1; i--) {
                            if (ids[i]) {
                                await fetchSyncPost("/api/filetree/doc2Heading", {
                                    srcID: ids[i],
                                    after: true,
                                    targetID: targetElement.getAttribute("data-node-id"),
                                });
                            }
                        }
                    } else {
                        for (let i = 0; i < ids.length; i++) {
                            if (ids[i]) {
                                await fetchSyncPost("/api/filetree/doc2Heading", {
                                    srcID: ids[i],
                                    after: false,
                                    targetID: targetElement.getAttribute("data-node-id"),
                                });
                            }
                        }
                    }

                    const getDocParam: IObject = {
                        id: protyle.block.id,
                        size: window.scribli.config.editor.dynamicLoadBlocks,
                    };
                    if (isEncryptedBox(protyle.notebookId)) {
                        getDocParam.notebook = protyle.notebookId;
                    }
                    fetchPost("/api/filetree/getDoc", getDocParam, getResponse => {
                        onGet({data: getResponse, protyle});
                        updatePanelByEditor({
                            protyle,
                            focus: false,
                            pushBackStack: false,
                            reload: true,
                            resize: false,
                        });
                        setTimeout(() => {
                            protyle.contentElement.scrollTop = scrollTop;
                            protyle.scroll.lastScrollTop = scrollTop - 1;
                        }, Constants.TIMEOUT_LOAD);
                    });
                }
                targetElement.classList.remove("dragover__bottom", "dragover__top", "dragover__left", "dragover__right");
            }
        } else if (!window.scribli.dragElement && (
            event.dataTransfer.types.includes("Files") || event.dataTransfer.types.includes("text/html")
        )) {
            event.preventDefault();
            // 
            const avElement = hasClosestByClassName(event.target, "av");
            if (!avElement) {
                focusByRange(getRangeByPoint(event.clientX, event.clientY));
                if (event.dataTransfer.types.includes("Files") && !isBrowser()) {
                    const files: ILocalFiles[] = [];
                    for (let i = 0; i < event.dataTransfer.files.length; i++) {
                        const filePath = webUtils.getPathForFile(event.dataTransfer.files[i]);
                        if (filePath) {
                            files.push({
                                path: filePath,
                                size: event.dataTransfer.files[i].size
                            });
                        } else {
                            paste(protyle, event);
                            break;
                        }
                    }
                    if (files.length > 0) {
                        uploadLocalFiles(files, protyle, !event.altKey);
                    }
                } else {
                    paste(protyle, event);
                }
                clearSelect(["av", "img"], protyle.wysiwyg.element);
            } else {
                const cellElement = hasClosestByClassName(event.target, "av__cell");
                if (cellElement) {
                    if (getTypeByCellElement(cellElement) === "mAsset" && event.dataTransfer.types[0] === "Files") {
                        /// #if !BROWSER
                        const files: ILocalFiles[] = [];
                        for (let i = 0; i < event.dataTransfer.files.length; i++) {
                            files.push({
                                path: webUtils.getPathForFile(event.dataTransfer.files[i]),
                                size: event.dataTransfer.files[i].size
                            });
                        }
                        dragUpload(files, protyle, cellElement);
                        /// #else
                        focusBlock(hasClosestBlock(cellElement) as HTMLElement);
                        uploadFiles(protyle, event.dataTransfer.files, undefined);
                        /// #endif
                    }
                }
            }
        }
        if (window.scribli.dragElement) {
            window.scribli.dragElement.style.opacity = "";
            window.scribli.dragElement = undefined;
        }
        // Clean up all drag indicators unconditionally after drop/cancel
        cleanupDragIndicators(document);
    });
    let dragoverElement: Element;
    let dragCache: { nodeId: string, indent: number, rgb: { r: number, g: number, b: number }, guides: string };
    let disabledPosition: string;
    const applyLiTarget = (htmlTarget: HTMLElement, event: DragEvent, canDropAsSibling = true): void => {
        cleanupDragIndicators(editorElement);
        const nodeId = htmlTarget.getAttribute("data-node-id");
        // Cache expensive computations per target element (never changes while hovering same element)
        if (!dragCache || dragCache.nodeId !== nodeId) {
            const contentBlock = Array.from(htmlTarget.children).find(c => c.hasAttribute("data-node-id")) as HTMLElement;
            const indent = contentBlock ? parseFloat(getComputedStyle(contentBlock).marginLeft) || 34 : 34;
            const depth = getListDepth(htmlTarget);
            const computedColor = getComputedStyle(htmlTarget).getPropertyValue("--b3-theme-primary-lighter").trim();
            const rgb = parseHexColor(computedColor) || {r: 53, g: 115, b: 217};
            let siblingGuides = "";
            for (let n = 1; n <= depth; n++) {
                if (siblingGuides) siblingGuides += ", ";
                const opacity = depth <= 1 ? 0.3 : 0.5 - (n - 1) / (depth - 1) * 0.4;
                siblingGuides += `${-n * indent}px 0 0 0 rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, ${opacity.toFixed(2)})`;
            }
            dragCache = {nodeId, indent, rgb, guides: siblingGuides || "none"};
        }
        const {indent, rgb, guides} = dragCache;

        const liRect = htmlTarget.getBoundingClientRect();
        const isRTL = getComputedStyle(htmlTarget).direction === "rtl";
        const offsetX = isRTL ? (liRect.right - event.clientX) : (event.clientX - liRect.left);
        const contentBlockForRect = Array.from(htmlTarget.children).find(c =>
            c.hasAttribute("data-node-id") && !c.classList.contains("list")) as HTMLElement;
        const contentRect = contentBlockForRect ? contentBlockForRect.getBoundingClientRect() : liRect;
        const isBottom = event.clientY > contentRect.top + contentRect.height / 2;
        const isFirstLi = !htmlTarget.previousElementSibling || !htmlTarget.previousElementSibling.classList.contains("li");
        let position = "bottom";
        if (isFirstLi && !isBottom) {
            position = "top";
        }
        const hasChildList = !!Array.from(htmlTarget.children).find(c => c.classList.contains("list"));
        const isChild = position === "bottom" && !hasChildList && offsetX >= indent;
        if (!canDropAsSibling && !isChild) {
            hideDragTip();
            return;
        }
        const sourceElements = Array.from(editorElement.querySelectorAll(".protyle-wysiwyg--select")) as HTMLElement[];
        const isNoOp = sourceElements.some(source =>
            source === htmlTarget ||
            source.contains(htmlTarget) ||
            (!isChild && position === "bottom" && source === htmlTarget.nextElementSibling) ||
            (position === "top" && source === htmlTarget.previousElementSibling));
        if (isNoOp) {
            cleanupDragIndicators(editorElement);
            hideDragTip();
            return;
        }
        const className = `dragover__${position}--${isChild ? "child" : "sibling"}`;

        htmlTarget.classList.add(className);
        htmlTarget.style.setProperty("--drag-indent", `${indent}px`);
        htmlTarget.style.setProperty("--drag-line-left", isChild ? `${indent}px` : "0");
        htmlTarget.style.setProperty("--drag-guides", guides);
        htmlTarget.style.setProperty("--drag-base-bg",
            isChild ? `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.6)` : "transparent");
        htmlTarget.style.setProperty("--drag-line-bg",
            `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, 0.6)`);
        highlightByLevel(editorElement, htmlTarget);
        const targetText = (getContenteditableElement(htmlTarget)?.textContent?.trim() || "").slice(0, 20);
        let action: string;
        if (event.altKey || (event.shiftKey && protyle.lite)) {
            action = window.scribli.languages.dragTipRef;
        } else if (event.shiftKey) {
            action = window.scribli.languages.dragTipEmbed;
        } else if (event.ctrlKey || protyle.lite) {
            action = window.scribli.languages.duplicate;
        } else if (isChild) {
            action = window.scribli.languages.dragTipListItemChild.replace("${x}", targetText);
        } else {
            const key = position === "bottom" ? "dragTipListItemAfter" : "dragTipListItemBefore";
            action = window.scribli.languages[key].replace("${x}", targetText);
        }
        showDragTip(window.scribli.dragTitle || "", action, event.clientX, event.clientY);
    };
    let cachedTargetText = "";
    let cachedIsCol = false;
    editorElement.addEventListener("dragover", (event: DragEvent & { target: HTMLElement }) => {
        if (protyle.disabled || event.dataTransfer.types.includes(Constants.SCRIBLI_DROP_EDITOR)) {
            event.preventDefault();
            event.stopPropagation();
            event.dataTransfer.dropEffect = "none";
            hideDragTip();
            return;
        }
        if (event.dataTransfer.types.includes(Constants.SCRIBLI_DROP_BLOCK_REF)) {
            if (hasClosestByClassName(event.target, "av")) {
                event.preventDefault();
                event.dataTransfer.dropEffect = "none";
                hideDragTip();
                cleanupDragIndicators(editorElement);
                return;
            }
            event.dataTransfer.dropEffect = "copy";
            showDragTip(window.scribli.dragTitle || "", window.scribli.languages.dragTipRef,
                event.clientX, event.clientY);
            renderBlockRefDragover(event);
            return;
        }
        let gutterType = "";
        for (const type of event.dataTransfer.types) {
            if (type.startsWith(Constants.SCRIBLI_DROP_GUTTER)) {
                gutterType = type;
            }
        }
        if (gutterType.startsWith(`${Constants.SCRIBLI_DROP_GUTTER}NodeAttributeView${Constants.ZWSP}ViewTab${Constants.ZWSP}`.toLowerCase())) {
            dragoverTab(event);
            event.preventDefault();
            return;
        }
        const gutterTypes = gutterType ? gutterType.replace(Constants.SCRIBLI_DROP_GUTTER, "").split(Constants.ZWSP) : [];
        const isAvSubType = gutterTypes[0] === "nodeattributeviewrowmenu" ||
            gutterTypes[0] === "nodeattributeviewrow" ||
            (gutterTypes[0] === "nodeattributeview" && ["viewtab", "col", "galleryitem"].includes(gutterTypes[1] || ""));
        const isAvTarget = hasClosestByClassName(event.target, "av__row") ||
            hasClosestByClassName(event.target, "av__row--util") ||
            hasClosestByClassName(event.target, "av__gallery-item") ||
            hasClosestByClassName(event.target, "av__gallery-add");
        if (event.dataTransfer.types.includes(Constants.SCRIBLI_DROP_FILE)) {
            showDragTip(window.scribli.dragTitle || "",
                isAvTarget ? window.scribli.languages.addToDatabase :
                    (event.altKey ? window.scribli.languages.dragTip2Heading : window.scribli.languages.dragTipRef),
                event.clientX, event.clientY);
        } else if (gutterType && !isAvSubType && !(event.altKey && isInEmbedBlock(event.target))) {
            let action: string;
            if (isAvTarget) {
                action = window.scribli.languages.addToDatabase;
            } else if (event.altKey || (event.shiftKey && protyle.lite)) {
                action = window.scribli.languages.dragTipRef;
            } else if (event.shiftKey) {
                action = window.scribli.languages.dragTipEmbed;
            } else if (event.ctrlKey || protyle.lite) {
                action = window.scribli.languages.duplicate;
            } else {
                action = window.scribli.languages.move;
            }
            showDragTip(window.scribli.dragTitle || "", action, event.clientX, event.clientY);
        } else {
            hideDragTip();
        }
        let targetElement: HTMLElement | false;
        if (event.dataTransfer.types.includes("Files")) {
            targetElement = hasClosestByClassName(event.target, "av__cell");
            if (targetElement && targetElement.getAttribute("data-dtype") === "mAsset" &&
                !targetElement.classList.contains("av__cell--header")) {
                event.preventDefault();
                if (dragoverElement && targetElement === dragoverElement) {
                    return;
                }
                const blockElement = hasClosestBlock(targetElement);
                if (blockElement) {
                    clearSelect(["cell", "row"], protyle.wysiwyg.element);
                    targetElement.classList.add("av__cell--select");
                    if (blockElement.getAttribute("data-av-type") !== "gallery") {
                        addDragFill(targetElement);
                    }
                    dragoverElement = targetElement;
                }
            }
            return;
        }

        if (!gutterType && !window.scribli.dragElement) {
            // 
            event.preventDefault();
            return;
        }
        const fileTreeIds = (event.dataTransfer.types.includes(Constants.SCRIBLI_DROP_FILE) && window.scribli.dragElement) ? window.scribli.dragElement.innerText : "";
        if (event.altKey && fileTreeIds.indexOf("-") === -1) {
            renderBlockRefDragover(event);
            return;
        }
        hideCaretLine();
        event.preventDefault();
        targetElement = hasClosestByClassName(event.target, "av__gallery-item") || hasClosestByClassName(event.target, "av__gallery-add") ||
            hasClosestByClassName(event.target, "av__row") || hasClosestByClassName(event.target, "av__row--util") ||
            hasClosestBlock(event.target);
        const directTargetElement = targetElement;
        if (targetElement && ["gallery", "kanban"].includes(targetElement.getAttribute("data-av-type")) && event.target.classList.contains("av__gallery")) {
            return;
        }
        const point = {x: event.clientX, y: event.clientY, className: ""};

        if (targetElement && (targetElement.classList.contains("bq") || targetElement.classList.contains("sb") || targetElement.classList.contains("list") || targetElement.classList.contains("li"))) {
            let prevElement = hasClosestBlock(document.elementFromPoint(point.x, point.y - 6));
            while (prevElement && targetElement.contains(prevElement)) {
                if (getNextBlockSibling(prevElement)) {
                    targetElement = prevElement;
                }
                prevElement = prevElement.parentElement;
            }
        }
        if (!targetElement) {
            if (event.clientY > editorElement.lastElementChild.getBoundingClientRect().bottom) {
                targetElement = editorElement.lastElementChild as HTMLElement;
                point.className = "dragover__bottom";
            } else if (event.clientY < editorElement.firstElementChild.getBoundingClientRect().top) {
                targetElement = editorElement.firstElementChild as HTMLElement;
                point.className = "dragover__top";
            } else {
                const contentRect = protyle.contentElement.getBoundingClientRect();
                const editorPosition = {
                    left: contentRect.left + parseInt(editorElement.style.paddingLeft),
                    right: contentRect.left + protyle.contentElement.clientWidth - parseInt(editorElement.style.paddingRight)
                };
                if (event.clientX < editorPosition.left) {
                    point.x = editorPosition.left;
                    point.className = "dragover__left";
                } else if (event.clientX >= editorPosition.right) {
                    point.x = editorPosition.right - 6;
                    point.className = "dragover__right";
                }
                targetElement = document.elementFromPoint(point.x, point.y) as HTMLElement;
                let probeOffset = 6;
                while (targetElement.classList.contains("protyle-wysiwyg") && probeOffset < 100) {
                    targetElement = document.elementFromPoint(point.x, point.y - probeOffset) as HTMLElement;
                    probeOffset += 6;
                }
                let hProbed = false;
                if (targetElement.classList.contains("protyle-wysiwyg")) {
                    const editorRect = editorElement.getBoundingClientRect();
                    const editorCenter = editorRect.left + editorRect.width / 2;
                    let hProbe = 6;
                    while (targetElement.classList.contains("protyle-wysiwyg") && hProbe < 100) {
                        const probeX = point.x > editorCenter ? point.x - hProbe : point.x + hProbe;
                        targetElement = document.elementFromPoint(probeX, point.y) as HTMLElement;
                        hProbe += 6;
                    }
                    hProbed = !targetElement.classList.contains("protyle-wysiwyg");
                }
                if (gutterTypes[0] === "nodelistitem") {
                    let closestLiFromPoint: HTMLElement;
                    if (targetElement.classList.contains("li")) {
                        closestLiFromPoint = targetElement;
                    } else if (targetElement.classList.contains("list")) {
                        const lis = targetElement.querySelectorAll(":scope > .li");
                        closestLiFromPoint = lis.length > 0 ? lis[lis.length - 1] as HTMLElement : targetElement.closest(".li") as HTMLElement;
                    } else {
                        closestLiFromPoint = targetElement.closest(".li") as HTMLElement;
                    }
                    targetElement = closestLiFromPoint || hasTopClosestByAttribute(targetElement, "data-node-id", null) as HTMLElement;
                } else {
                    targetElement = hasTopClosestByAttribute(targetElement, "data-node-id", null) as HTMLElement;
                }
                if (targetElement && targetElement.classList.contains("sb") && targetElement.getAttribute("data-sb-layout") === "col") {
                    if (point.className !== "dragover__left" && point.className !== "dragover__right" && !hProbed) {
                        const childElement = targetElement.querySelectorAll("[data-node-id]");
                        targetElement = childElement[point.className === "dragover__left" ? 0 : childElement.length - 1] as HTMLElement;
                    }
                }
            }
        } else if (targetElement && targetElement.classList.contains("list")) {
            targetElement = hasClosestBlock(document.elementFromPoint(event.clientX, event.clientY - 6));
        }
        if (gutterType && gutterType.startsWith(`${Constants.SCRIBLI_DROP_GUTTER}NodeAttributeView${Constants.ZWSP}Col${Constants.ZWSP}`.toLowerCase())) {
            targetElement = hasClosestByClassName(event.target, "av__cell");
            if (targetElement) {
                const targetRowElement = hasClosestByClassName(targetElement, "av__row--header");
                const dragRowElement = hasClosestByClassName(window.scribli.dragElement, "av__row--header");
                if (targetElement === window.scribli.dragElement || !targetRowElement || !dragRowElement ||
                    (targetRowElement && dragRowElement && targetRowElement !== dragRowElement)
                ) {
                    targetElement = false;
                }
            }
        } else if (targetElement && gutterType && gutterType.startsWith(`${Constants.SCRIBLI_DROP_GUTTER}NodeAttributeViewRowMenu${Constants.ZWSP}`.toLowerCase())) {
            if ((!targetElement.classList.contains("av__row") && !targetElement.classList.contains("av__row--util")) ||
                (window.scribli.dragElement && !window.scribli.dragElement.contains(targetElement))) {
                targetElement = false;
            } else {
                const bodyElement = hasClosestByClassName(targetElement, "av__body");
                if (bodyElement) {
                    const blockElement = hasClosestBlock(bodyElement) as HTMLElement;
                    const groupID = bodyElement.getAttribute("data-group-id");
                    const isTCU = ["template", "created", "updated"].includes(bodyElement.getAttribute("data-dtype"));
                    const hasSort = blockElement.querySelector('.block__icon[data-type="av-sort"]')?.classList.contains("block__icon--active");
                    gutterTypes[2].split(",").find(item => {
                        const sourceGroupID = item ? item.split("@")[1] : "";
                        if (sourceGroupID !== groupID && isTCU) {
                            targetElement = false;
                            return true;
                        }
                        if (sourceGroupID === groupID && hasSort) {
                            targetElement = false;
                            return true;
                        }
                    });
                }
            }
        } else if (targetElement && gutterType && gutterType.startsWith(`${Constants.SCRIBLI_DROP_GUTTER}NodeAttributeView${Constants.ZWSP}GalleryItem${Constants.ZWSP}`.toLowerCase())) {
            const containerElement = hasClosestByClassName(event.target, "av__container");
            if (targetElement.classList.contains("av") || !containerElement ||
                !containerElement.contains(window.scribli.dragElement) || targetElement === window.scribli.dragElement) {
                targetElement = false;
            } else {
                const bodyElement = hasClosestByClassName(targetElement, "av__body");
                if (bodyElement) {
                    const blockElement = hasClosestBlock(bodyElement) as HTMLElement;
                    const groupID = bodyElement.getAttribute("data-group-id");
                    const isTCU = ["template", "created", "updated"].includes(bodyElement.getAttribute("data-dtype"));
                    const hasSort = blockElement.querySelector('.block__icon[data-type="av-sort"]')?.classList.contains("block__icon--active");
                    gutterTypes[2].split(",").find(item => {
                        const sourceGroupID = item ? item.split("@")[1] : "";
                        if (sourceGroupID !== groupID && isTCU) {
                            targetElement = false;
                            return true;
                        }
                        if (sourceGroupID === groupID && hasSort) {
                            targetElement = false;
                            return true;
                        }
                    });
                }
            }
        }

        if (!targetElement) {
            editorElement.querySelectorAll(".dragover__bottom, .dragover__top, .dragover, .dragover__left, .dragover__right").forEach((item: HTMLElement) => {
                item.classList.remove("dragover__top", "dragover__bottom", "dragover", "dragover__left", "dragover__right");
            });
            hideDragTip();
            return;
        }
        if (targetElement.getAttribute("data-type") === "NodeBlockQueryEmbed") {
            if (editorElement.firstElementChild === targetElement &&
                event.clientY < targetElement.getBoundingClientRect().top) {
                point.className = "dragover__top";
            } else if (editorElement.lastElementChild === targetElement &&
                event.clientY > targetElement.getBoundingClientRect().bottom) {
                point.className = "dragover__bottom";
            } else {
                clearDragoverElement(dragoverElement);
                return;
            }
        } else if (isInEmbedBlock(targetElement)) {
            clearDragoverElement(dragoverElement);
            return;
        }
        const isNotAvItem = !targetElement.classList.contains("av__row") &&
            !targetElement.classList.contains("av__row--util") &&
            !targetElement.classList.contains("av__gallery-item") &&
            !targetElement.classList.contains("av__gallery-add");
        if (!targetElement.classList.contains("sb")) {
            const ancestorSb = targetElement.closest('[data-type="NodeSuperBlock"]') as HTMLElement;
            if (ancestorSb) {
                const sbChildBlocks = Array.from(ancestorSb.querySelectorAll("[data-node-id]"));
                const firstBlock = sbChildBlocks[0] as HTMLElement;
                const lastBlock = sbChildBlocks[sbChildBlocks.length - 1] as HTMLElement;
                const isFirstBlock = targetElement === firstBlock || firstBlock.contains(targetElement);
                const isLastBlock = targetElement === lastBlock || lastBlock.contains(targetElement);
                const childRect = targetElement.getBoundingClientRect();
                if ((isFirstBlock && event.clientX < childRect.left + 8) ||
                    (isLastBlock && event.clientX > childRect.right - 8)) {
                    targetElement = ancestorSb;
                }
                if (gutterTypes[0] === "nodelist" &&
                    ancestorSb.getAttribute("data-sb-layout") === "col" &&
                    targetElement !== ancestorSb) {
                    const colList = targetElement.closest(".list") as HTMLElement;
                    if (colList && ancestorSb === colList.parentElement) {
                        targetElement = colList;
                    }
                }
            }
        }
        const isListSource = gutterTypes[0] === "nodelistitem" || gutterTypes[0] === "nodelist";
        const isContentBlockSource = !!gutterType && !isListSource && !isAvSubType;
        const keepLiContentTarget = targetElement === directTargetElement && isContentBlockSource &&
            targetElement.parentElement?.getAttribute("data-type") === "NodeListItem";
        let liTarget = targetElement.classList.contains("list") || keepLiContentTarget ? null :
            (targetElement.getAttribute("data-type") === "NodeListItem"
                ? targetElement : targetElement.parentElement?.getAttribute("data-type") === "NodeListItem"
                    ? targetElement.parentElement : null);
        if (isListSource && !liTarget) {
            const sourceSelected = editorElement.querySelector(".protyle-wysiwyg--select") as HTMLElement;
            if (sourceSelected && (sourceSelected.classList.contains("li") || sourceSelected.classList.contains("list"))) {
                if (targetElement.classList.contains("list") && targetElement.contains(sourceSelected)) {
                    cleanupDragIndicators(editorElement);
                    hideDragTip();
                    return;
                }
                let current: Element = sourceSelected;
                while (current && current !== editorElement) {
                    if (current.classList.contains("list") || current.classList.contains("li")) {
                        const checkSiblings = (container: Element) => {
                            let prevSibling = container.previousElementSibling;
                            while (prevSibling && prevSibling.classList.contains("protyle-attr")) {
                                prevSibling = prevSibling.previousElementSibling;
                            }
                            let nextSibling = container.nextElementSibling;
                            while (nextSibling && nextSibling.classList.contains("protyle-attr")) {
                                nextSibling = nextSibling.nextElementSibling;
                            }
                            return targetElement === prevSibling || targetElement === nextSibling;
                        };
                        if (checkSiblings(current)) {
                            if (current.parentElement === editorElement) {
                                break;
                            }
                            cleanupDragIndicators(editorElement);
                            hideDragTip();
                            return;
                        }
                    }
                    current = current.parentElement;
                }
            }
        }
        if (liTarget && fileTreeIds.indexOf("-") > -1 && isNotAvItem) {
            if (!event.altKey) {
                return;
            } else if (fileTreeIds.split(",").includes(protyle.block.rootID) && event.altKey) {
                return;
            }
        }
        if (isListSource && targetElement.classList.contains("list")) {
            const sourceSelected = editorElement.querySelector(".protyle-wysiwyg--select");
            if (sourceSelected && targetElement.contains(sourceSelected)) {
                cleanupDragIndicators(editorElement);
                hideDragTip();
                return;
            }
            const lis = targetElement.querySelectorAll(":scope > .li");
            const lastLi = lis[lis.length - 1];
            const firstLi = lis[0];
            const listRect = targetElement.getBoundingClientRect();
            const isListBottom = event.clientY > listRect.top + listRect.height / 2;
            const sourceIds = Array.from(editorElement.querySelectorAll(".protyle-wysiwyg--select"))
                .map((e: HTMLElement) => e.getAttribute("data-node-id"));
            const isNoOpList = (isListBottom && lastLi && sourceIds.includes(lastLi.getAttribute("data-node-id"))) ||
                (!isListBottom && firstLi && sourceIds.includes(firstLi.getAttribute("data-node-id")));
            if (isNoOpList) {
                cleanupDragIndicators(editorElement);
                hideDragTip();
                return;
            }
        }
        if (liTarget) {
            let topList: Element = liTarget as HTMLElement;
            while (topList.parentElement?.classList.contains("li") ||
                   topList.parentElement?.classList.contains("list")) {
                topList = topList.parentElement;
                if (topList.classList.contains("list") && !topList.parentElement?.classList.contains("li")) {
                    break;
                }
            }
            const topListRect = topList.getBoundingClientRect();
            const isLeftEdge = event.clientX < topListRect.left + 32;
            const isRightEdge = event.clientX > topListRect.right - 32;
            if (gutterTypes[0] === "nodelistitem") {
                if (isRightEdge) {
                    cleanupDragIndicators(editorElement);
                    return;
                }
                applyLiTarget(liTarget as HTMLElement, event);
                return;
            }
            if (isLeftEdge || isRightEdge) {
                liTarget = null;
            } else {
                applyLiTarget(liTarget as HTMLElement, event, !isContentBlockSource);
                return;
            }
        }
        if (targetElement && dragoverElement && targetElement === dragoverElement) {
            const nodeRect = targetElement.getBoundingClientRect();
            cleanupDragIndicators(editorElement);
            editorElement.querySelectorAll("[select-start], [select-end]").forEach((item: HTMLElement) => {
                item.removeAttribute("select-start");
                item.removeAttribute("select-end");
            });
            if (fileTreeIds.indexOf("-") > -1 && isNotAvItem) {
                if (!event.altKey) {
                    return;
                } else if (fileTreeIds.split(",").includes(protyle.block.rootID) && event.altKey) {
                    return;
                }
            }
            if (targetElement.getAttribute("data-type") === "NodeAttributeView" && hasClosestByTag(event.target, "TD")) {
                return;
            }
            const isSelfFast = !event.ctrlKey && !event.shiftKey && !event.altKey && gutterTypes[2]?.split(",").some((item: string) =>
                item && hasClosestByAttribute(targetElement as HTMLElement, "data-node-id", item));
            if (isSelfFast && "nodeattributeviewrowmenu" !== gutterTypes[0]) {
                hideDragTip();
                return;
            }
            if (point.className && !liTarget && !targetElement.classList.contains("sb")) {
                if (!(gutterTypes[0] === "nodelistitem" && targetElement.classList.contains("list") &&
                    (point.className === "dragover__left" || point.className === "dragover__right"))) {
                    targetElement.classList.add(point.className);
                    addDragover(targetElement);
                    let displayText = cachedTargetText;
                    if (!displayText && targetElement.classList.contains("list")) {
                        const firstLi = targetElement.querySelector(":scope > .li");
                        displayText = getContenteditableElement(firstLi as HTMLElement)?.textContent?.trim() || "";
                    }
                    if (!event.altKey && !event.shiftKey && !event.ctrlKey && gutterType && !isAvSubType && !isAvTarget && !targetElement.classList.contains("sb")) {
                        const isFront = point.className === "dragover__top" || point.className === "dragover__left";
                        const isBack = point.className === "dragover__bottom" || point.className === "dragover__right";
                        if ((isFront || isBack) && displayText) {
                            const isHorizontal = point.className === "dragover__left" || point.className === "dragover__right";
                            const key = (isHorizontal || cachedIsCol)
                                ? (isFront ? window.scribli.languages.dragTipMoveTargetFront : window.scribli.languages.dragTipMoveTargetBack)
                                : (isFront ? window.scribli.languages.dragTipMoveTargetAbove : window.scribli.languages.dragTipMoveTargetBelow);
                            showDragTip(window.scribli.dragTitle || "", key.replace("${x}", displayText),
                                event.clientX, event.clientY);
                        }
                    }
                }
                return;
            }

            if (targetElement.classList.contains("av__cell")) {
                if (event.clientX < nodeRect.left + nodeRect.width / 2 && event.clientX > nodeRect.left &&
                    !targetElement.classList.contains("av__row") && targetElement.previousElementSibling !== window.scribli.dragElement) {
                    targetElement.classList.add("dragover__left");
                } else if (event.clientX > nodeRect.right - nodeRect.width / 2 && event.clientX <= nodeRect.right + 1 &&
                    !targetElement.classList.contains("av__row") && targetElement !== window.scribli.dragElement.previousElementSibling) {
                    if (window.scribli.dragElement.previousElementSibling.classList.contains("av__colsticky") &&
                        targetElement === window.scribli.dragElement.previousElementSibling.lastElementChild) {
                        // Intentionally empty.
                    } else {
                        targetElement.classList.add("dragover__right");
                    }
                }
                return;
            }
            // gallery & kanban
            if (targetElement.classList.contains("av__gallery-item")) {
                if (hasClosestByClassName(targetElement, "av__kanban-group")) {
                    const midTop = nodeRect.top + nodeRect.height / 2;
                    if (event.clientY < midTop && event.clientY > nodeRect.top - 13) {
                        targetElement.classList.add("dragover__top");
                    } else if (event.clientY > midTop && event.clientY <= nodeRect.bottom + 13) {
                        targetElement.classList.add("dragover__bottom");
                    }
                } else {
                    const midLeft = nodeRect.left + nodeRect.width / 2;
                    if (event.clientX < midLeft && event.clientX > nodeRect.left - 13) {
                        targetElement.classList.add("dragover__left");
                    } else if (event.clientX > midLeft && event.clientX <= nodeRect.right + 13) {
                        targetElement.classList.add("dragover__right");
                    }
                }
                return;
            }
            if (targetElement.classList.contains("av__gallery-add")) {
                if (hasClosestByClassName(targetElement, "av__kanban-group")) {
                    targetElement.classList.add("dragover__top");
                } else {
                    targetElement.classList.add("dragover__left");
                }
                return;
            }

            if (targetElement.classList.contains("sb")) {
                const sbRect = targetElement.getBoundingClientRect();
                const isSbLeftEdge = point.className === "dragover__left" || event.clientX < sbRect.left + 32;
                const isSbRightEdge = point.className === "dragover__right" || event.clientX > sbRect.right - 32;
                if (isSbLeftEdge || isSbRightEdge) {
                    const edgeClass = isSbLeftEdge ? "dragover__left" : "dragover__right";
                    targetElement.classList.add(edgeClass);
                    addDragover(targetElement);
                    const sbFirstBlock = targetElement.querySelector("[data-node-id]") as HTMLElement;
                    const sbText = getContenteditableElement(sbFirstBlock)?.textContent?.trim() || "";
                    if (!event.altKey && !event.shiftKey && !event.ctrlKey && gutterType && !isAvSubType && !isAvTarget && sbText) {
                        const key = isSbLeftEdge
                            ? window.scribli.languages.dragTipMoveTargetFront
                            : window.scribli.languages.dragTipMoveTargetBack;
                        showDragTip(window.scribli.dragTitle || "", key.replace("${x}", sbText),
                            event.clientX, event.clientY);
                    }
                    return;
                }
            }

            if (event.clientX < nodeRect.left + (targetElement.classList.contains("list") ? 8 : 32) &&
                event.clientX >= nodeRect.left - 1 &&
                !targetElement.classList.contains("av__row")) {
                targetElement.classList.add("dragover__left");
                addDragover(targetElement);
                if (!event.altKey && !event.shiftKey && !event.ctrlKey && gutterType && !isAvSubType && !isAvTarget && !targetElement.classList.contains("sb") && cachedTargetText) {
                    showDragTip(window.scribli.dragTitle || "",
                        window.scribli.languages.dragTipMoveTargetFront.replace("${x}", cachedTargetText),
                        event.clientX, event.clientY);
                }
            } else if (event.clientX > nodeRect.right - 32 && event.clientX < nodeRect.right &&
                !targetElement.classList.contains("av__row")) {
                targetElement.classList.add("dragover__right");
                addDragover(targetElement);
                if (!event.altKey && !event.shiftKey && !event.ctrlKey && gutterType && !isAvSubType && !isAvTarget && !targetElement.classList.contains("sb") && cachedTargetText) {
                    showDragTip(window.scribli.dragTitle || "",
                        window.scribli.languages.dragTipMoveTargetBack.replace("${x}", cachedTargetText),
                        event.clientX, event.clientY);
                }
            } else if (targetElement.classList.contains("av__row--header")) {
                targetElement.classList.add("dragover__bottom");
            } else if (targetElement.classList.contains("av__row--util")) {
                targetElement.previousElementSibling.classList.add("dragover__bottom");
            } else {
                if (event.clientY > nodeRect.top + nodeRect.height / 2 && disabledPosition !== "bottom") {
                    targetElement.classList.add("dragover__bottom");
                    addDragover(targetElement);
                    if (!event.altKey && !event.shiftKey && !event.ctrlKey && gutterType && !isAvSubType && !isAvTarget && !targetElement.classList.contains("sb") && cachedTargetText) {
                        showDragTip(window.scribli.dragTitle || "",
                            (cachedIsCol ? window.scribli.languages.dragTipMoveTargetBack : window.scribli.languages.dragTipMoveTargetBelow).replace("${x}", cachedTargetText),
                            event.clientX, event.clientY);
                    }
                } else if (disabledPosition !== "top") {
                    targetElement.classList.add("dragover__top");
                    addDragover(targetElement);
                    if (!event.altKey && !event.shiftKey && !event.ctrlKey && gutterType && !isAvSubType && !isAvTarget && !targetElement.classList.contains("sb") && cachedTargetText) {
                        showDragTip(window.scribli.dragTitle || "",
                            (cachedIsCol ? window.scribli.languages.dragTipMoveTargetFront : window.scribli.languages.dragTipMoveTargetAbove).replace("${x}", cachedTargetText),
                            event.clientX, event.clientY);
                    }
                }
            }
            return;
        }

        if (fileTreeIds.indexOf("-") > -1) {
            if (fileTreeIds.split(",").includes(protyle.block.rootID) && isNotAvItem && event.altKey) {
                dragoverElement = undefined;
                cleanupDragIndicators(editorElement);
                editorElement.querySelectorAll("[select-start], [select-end]").forEach((item: HTMLElement) => {
                    item.removeAttribute("select-start");
                    item.removeAttribute("select-end");
                });
            } else {
                dragoverElement = targetElement;
            }
            return;
        }

        if (gutterType) {
            disabledPosition = "";
            if (gutterTypes[0] === "nodeattributeview" && gutterTypes[1] === "col" && targetElement.getAttribute("data-id") === gutterTypes[2]) {
                clearDragoverElement(dragoverElement);
                return;
            }
            if (gutterTypes[0] === "nodeattributeviewrowmenu" && gutterTypes[2].split("@")[0] === targetElement.getAttribute("data-id")) {
                clearDragoverElement(dragoverElement);
                return;
            }
            const isSelf = gutterTypes[2].split(",").find((item: string) => {
                if (item && hasClosestByAttribute(targetElement as HTMLElement, "data-node-id", item)) {
                    return true;
                }
            });
            if (isSelf && "nodeattributeviewrowmenu" !== gutterTypes[0] && !event.ctrlKey && !event.shiftKey && !event.altKey) {
                clearDragoverElement(dragoverElement);
                return;
            }
            if (gutterTypes[0] === "nodelistitem" && "NodeListItem" === targetElement.getAttribute("data-type")) {
                const notLiItem = Array.from(protyle.wysiwyg.element.querySelectorAll(".protyle-wysiwyg--select")).find((item: HTMLElement) => {
                    if (!item.classList.contains("li")) {
                        return true;
                    }
                });
                if (notLiItem) {
                    clearDragoverElement(dragoverElement);
                    return;
                }
            }
            if (!["nodelistitem", "nodelist"].includes(gutterTypes[0]) && targetElement.getAttribute("data-type") === "NodeListItem") {
                clearDragoverElement(dragoverElement);
                return;
            }
            if (gutterTypes[0] === "nodelistitem" && targetElement.parentElement.classList.contains("li") &&
                targetElement.previousElementSibling?.classList.contains("protyle-action")) {
                disabledPosition = "top";
            }
            if (gutterTypes[0] === "nodelistitem" &&
                targetElement.nextElementSibling?.classList.contains("list") &&
                // 
                targetElement.parentElement?.classList.contains("li")) {
                disabledPosition = "bottom";
            }
            if (targetElement && targetElement.classList.contains("av__row--header")) {
                disabledPosition = "top";
            }
            dragoverElement = targetElement;
            cachedTargetText = getContenteditableElement(targetElement as HTMLElement)?.textContent?.trim() || "";
            cachedIsCol = !!hasClosestByAttribute(targetElement as HTMLElement, "data-sb-layout", "col");
            highlightColColumn(targetElement as HTMLElement);
        }
        if (!event.altKey && !event.shiftKey && !event.ctrlKey && gutterType && !isAvSubType && targetElement && !isAvTarget && point.className) {
            const targetText = getContenteditableElement(targetElement as HTMLElement)?.textContent?.trim() || "";
            const isFront = point.className === "dragover__top" || point.className === "dragover__left";
            const isBack = point.className === "dragover__bottom" || point.className === "dragover__right";
            if (targetText && (isFront || isBack)) {
                const isCol = hasClosestByAttribute(targetElement as HTMLElement, "data-sb-layout", "col");
                const key = isCol
                    ? (isFront ? window.scribli.languages.dragTipMoveTargetFront : window.scribli.languages.dragTipMoveTargetBack)
                    : (isFront ? window.scribli.languages.dragTipMoveTargetAbove : window.scribli.languages.dragTipMoveTargetBelow);
                showDragTip(window.scribli.dragTitle || "", key.replace("${x}", targetText),
                    event.clientX, event.clientY);
            }
        }
    });
    let counter = 0;
    editorElement.addEventListener("dragleave", (event: DragEvent & { target: HTMLElement }) => {
        if (protyle.disabled) {
            event.preventDefault();
            event.stopPropagation();
            return;
        }
        counter--;
        if (counter === 0) {
            cleanupDragIndicators(editorElement);
            dragoverElement = undefined;
            hideDragTip();
        }
    });
    editorElement.addEventListener("dragenter", (event) => {
        event.preventDefault();
        counter++;
    });
    editorElement.addEventListener("dragend", () => {
        if (window.scribli.dragElement) {
            window.scribli.dragElement.style.opacity = "";
            window.scribli.dragElement = undefined;
            document.onmousemove = null;
        }
        // Clean up all drag indicators on cancel
        cleanupDragIndicators(editorElement);
        dragoverElement = undefined;
        hideDragTip();
        window.scribli.dragTitle = "";
    });
    // Fallback: document-level cleanup in case dragend doesn't bubble
    document.addEventListener("dragend", () => {
        cleanupDragIndicators(document);
    }, {once: true});
};

const cleanupDragIndicators = (scope: ParentNode) => {
    scope.querySelectorAll(".dragover__top, .dragover__bottom, .dragover__left, .dragover__right, .dragover__top--sibling, .dragover__bottom--sibling, .dragover__top--child, .dragover__bottom--child, .dragover, [style*=\"--drag-indent\"]").forEach((item: HTMLElement) => {
        item.classList.remove("dragover__top", "dragover__bottom", "dragover__left", "dragover__right", "dragover",
            "dragover__top--sibling", "dragover__bottom--sibling", "dragover__top--child", "dragover__bottom--child");
        item.style.removeProperty("--drag-indent");
        item.style.removeProperty("--drag-guides");
        item.style.removeProperty("--drag-line-left");
        item.style.removeProperty("--drag-base-bg");
        item.style.removeProperty("--drag-line-bg");
    });
};

const getListDepth = (liElement: Element): number => {
    let depth = 0;
    let list = liElement.parentElement;
    while (list && list.classList.contains("list")) {
        const parentLi = list.parentElement;
        if (parentLi && parentLi.classList.contains("li")) {
            depth++;
            list = parentLi.parentElement;
        } else {
            break;
        }
    }
    return depth;
};

const parseHexColor = (color: string): { r: number, g: number, b: number } | null => {
    if (!color) return null;
    const hexMatch = color.match(/^#([0-9a-f]{3,8})$/i);
    if (hexMatch) {
        let hex = hexMatch[1];
        if (hex.length === 3) {
            hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2];
        }
        if (hex.length >= 6) {
            return {
                r: parseInt(hex.slice(0, 2), 16),
                g: parseInt(hex.slice(2, 4), 16),
                b: parseInt(hex.slice(4, 6), 16),
            };
        }
    }
    const rgbMatch = color.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
    if (rgbMatch) {
        return {
            r: parseInt(rgbMatch[1]),
            g: parseInt(rgbMatch[2]),
            b: parseInt(rgbMatch[3]),
        };
    }
    return null;
};

const highlightByLevel = (editorElement: HTMLElement, liElement: HTMLElement) => {
    editorElement.querySelectorAll(".dragover").forEach((item: HTMLElement) => {
        item.classList.remove("dragover");
    });
    liElement.classList.add("dragover");
};

const addDragover = (element: HTMLElement) => {
    if (element.classList.contains("sb") ||
        element.classList.contains("li") ||
        element.classList.contains("list") ||
        element.classList.contains("bq")) {
        element.classList.add("dragover");
    }
    highlightColColumn(element);
};

const highlightColColumn = (element: HTMLElement) => {
    if (element.getAttribute("data-sb-layout") === "col") {
        element.classList.add("dragover");
    }
};

// 
const clearDragoverElement = (element: Element) => {
    if (element) {
        element.classList.remove("dragover__top", "dragover__bottom", "dragover", "dragover__left", "dragover__right", "dragover__top--sibling", "dragover__bottom--sibling", "dragover__top--child", "dragover__bottom--child");
        (element as HTMLElement).style.removeProperty("--drag-indent");
        (element as HTMLElement).style.removeProperty("--drag-guides");
        (element as HTMLElement).style.removeProperty("--drag-line-left");
        (element as HTMLElement).style.removeProperty("--drag-base-bg");
        element = undefined;
    }
    hideDragTip();
};

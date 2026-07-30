import {BlockPanel} from "./Panel";
import {hasClosestByAttribute, hasClosestByClassName,} from "../protyle/util/hasClosest";
import {fetchPost, fetchSyncPost} from "../util/fetch";
import {hideTooltip, showTooltip} from "../dialog/tooltip";
import {isLocalPath, parseScribliUriInfo} from "../util/pathName";
import {App} from "../index";
import {Constants} from "../constants";
import {getCellText} from "../protyle/render/av/cell";
import {isTouchDevice} from "../util/functions";
import {escapeAriaLabel, escapeHtml, escapeLessThans} from "../util/escape";
import {getInstanceById} from "../layout/util";
import {Editor} from "../editor";
import {Tab} from "../layout/Tab";

let popoverTargetElement: HTMLElement;
let tooltipAbortController: AbortController | null = null;
export const initBlockPopover = (app: App) => {
    let timeout: number;
    let timeoutHide: number;
    document.addEventListener("mouseover", (event: MouseEvent & { target: HTMLElement, path: HTMLElement[] }) => {
        if (!window.scribli.config || !window.scribli.menus ||
            window.scribli.dragElement || document.onmousemove) {
            hideTooltip();
            return;
        }
        if (tooltipAbortController) {
            tooltipAbortController.abort();
            tooltipAbortController = null;
        }
        const aElement = hasClosestByAttribute(event.target, "data-type", "a", true) ||
            hasClosestByClassName(event.target, "ariaLabel") ||
            hasClosestByAttribute(event.target, "data-type", "tab-header") ||
            hasClosestByAttribute(event.target, "data-type", "inline-memo") ||
            hasClosestByClassName(event.target, "av__calc--ashow") ||
            hasClosestByClassName(event.target, "av__cell");
        if (aElement) {
            let tooltipClass = "";
            let tip = aElement.getAttribute("aria-label") || "";
            if (aElement.classList.contains("av__cell") && !aElement.classList.contains("ariaLabel")) {
                if (aElement.classList.contains("av__cell--header")) {
                    const textElement = aElement.querySelector(".av__celltext");
                    const desc = aElement.getAttribute("data-desc");
                    if (textElement.scrollWidth > textElement.clientWidth + 0.5 || desc) {
                        if (desc) {
                            tip = `${getCellText(aElement)}<div class='ft__on-surface'>${escapeAriaLabel(desc)}</div>`;
                        } else {
                            tip = getCellText(aElement);
                        }
                    }
                } else {
                    if (aElement.firstElementChild?.getAttribute("data-type") === "url") {
                        if (aElement.firstElementChild.textContent.indexOf("...") > -1) {
                            tip = Lute.EscapeHTMLStr(aElement.firstElementChild.getAttribute("data-href"));
                            tooltipClass = "href";
                        }
                    }
                    tip = "";
                    if (!tip && aElement.dataset.wrap !== "true" && event.target.dataset.type !== "block-more" && !hasClosestByClassName(event.target, "block__icon")) {
                        aElement.style.overflow = "auto";
                        if (aElement.scrollWidth > aElement.clientWidth + 2) {
                            tip = Lute.EscapeHTMLStr(getCellText(aElement));
                        }
                        aElement.style.overflow = "";
                    }
                }
            } else if (aElement.parentElement.parentElement.classList.contains("av__views") && aElement.parentElement.classList.contains("layout-tab-bar")) {
                const textElement = aElement.querySelector(".item__text");
                const desc = aElement.getAttribute("data-desc");
                if (textElement.scrollWidth > textElement.clientWidth + 0.5 || desc) {
                    if (desc) {
                        tip = `${textElement.textContent}<div class='ft__on-surface'>${escapeAriaLabel(desc)}</div>`;
                    } else {
                        tip = textElement.textContent;
                    }
                }
            } else if (aElement.classList.contains("av__celltext--url")) {
                const title = aElement.getAttribute("data-name") || "";
                tip = tip ? `<span style="word-break: break-all">${tip.substring(0, Constants.SIZE_TITLE)}</span>${title ? '<div class="fn__hr"></div><span>' + title + "</span>" : ""}` : title;
                tooltipClass = "href";
            } else if (aElement.classList.contains("av__calc--ashow") && aElement.clientWidth + 2 < aElement.scrollWidth) {
                tip = aElement.lastChild.textContent + " " + aElement.firstElementChild.textContent;
            } else if (aElement.getAttribute("data-type") === "setRelationCell") {
                const childElement = aElement.querySelector(".b3-menu__label");
                if (childElement && childElement.clientWidth < childElement.scrollWidth) {
                    tip = childElement.textContent;
                }
            } else if (aElement.classList.contains("protyle-attr--memo")) {
                tip = escapeHtml(tip);
            }
            let tooltipSpace: number | undefined;
            if (!tip && aElement.getAttribute("data-type")?.includes("inline-memo")) {
                tip = window.DOMPurify.sanitize(aElement.getAttribute("data-inline-memo-content"));
                tooltipClass = "memo";
                tooltipSpace = 0;
            }
            if (!tip) {
                if (aElement.getAttribute("data-type")?.includes("a")) {
                    tooltipClass = "href";
                    tooltipSpace = 0;
                }
                const href = aElement.getAttribute("data-href") || "";
                if (href) {
                    tip = `<span style="word-break: break-all">${href.substring(0, Constants.SIZE_TITLE)}</span>`;
                }
                const title = aElement.getAttribute("data-title");
                if (!window.scribli.isPublish && tip && isLocalPath(href) && !aElement.classList.contains("b3-tooltips")) {
                    let assetTip = tip;
                    tooltipAbortController = new AbortController();
                    const signal = tooltipAbortController.signal;
                    const capturedController = tooltipAbortController;
                    fetchPost("/api/asset/statAsset", {path: href}, (response) => {
                        if (signal.aborted) {
                            return;
                        }
                        if (response.code === 1) {
                            if (title) {
                                assetTip += '<div class="fn__hr"></div><span>' + title + "</span>";
                            }
                        } else {
                            assetTip += ` ${response.data.hSize}${title ? '<div class="fn__hr"></div><span>' + title + "</span>" : ""}<br>${window.scribli.languages.modifiedAt} ${response.data.hUpdated}<br>${window.scribli.languages.createdAt} ${response.data.hCreated}`;
                        }
                        try {
                            showTooltip(decodeURIComponent(assetTip), aElement, tooltipClass, event, tooltipSpace);
                        } catch (e) {
                            showTooltip(assetTip, aElement, tooltipClass, event, tooltipSpace);
                        }
                        if (tooltipAbortController === capturedController) {
                            tooltipAbortController = null;
                        }
                    }, undefined, undefined, signal);
                    tip = "";
                } else if (title) {
                    tip = (tip ? (tip + '<div class="fn__hr"></div>') : "") + "<span>" + title + "</span>";
                }
            }

            const tabElement = hasClosestByAttribute(event.target, "data-type", "tab-header");
            if (tabElement) {
                const tab = getInstanceById(tabElement.getAttribute("data-id"));
                if (tab instanceof Tab) {
                    let id = "";
                    if (tab.model instanceof Editor && tab.model.editor?.protyle?.block?.rootID) {
                        id = (tab.model as Editor).editor.protyle.block.rootID;
                    } else if (!tab.model) {
                        const initData = JSON.parse(tab.headElement.getAttribute("data-initdata") || "{}");
                        if (initData && initData.instance === "Editor") {
                            id = initData.blockId;
                        }
                    }
                    if (id) {
                        tooltipAbortController = new AbortController();
                        const signal = tooltipAbortController.signal;
                        const capturedController = tooltipAbortController;
                        fetchPost("/api/filetree/getFullHPathByID", {
                            id
                        }, (response) => {
                            if (signal.aborted) {
                                return;
                            }
                            showTooltip(escapeLessThans(response.data), tab.headElement);
                            tab.headElement.setAttribute("aria-label", escapeLessThans(response.data));
                            if (tooltipAbortController === capturedController) {
                                tooltipAbortController = null;
                            }
                        }, undefined, undefined, signal);
                    } else {
                        tab.headElement.setAttribute("aria-label", escapeLessThans(tab.title));
                    }
                }
            }

            const notebookItemElement = hasClosestByClassName(event.target, "b3-list-item__text");
            if (notebookItemElement && notebookItemElement.parentElement.getAttribute("data-type") === "navigation-root") {
                tooltipAbortController = new AbortController();
                const signal = tooltipAbortController.signal;
                const capturedController = tooltipAbortController;
                fetchPost("/api/notebook/getNotebookInfo", {notebook: notebookItemElement.parentElement.parentElement.getAttribute("data-url")}, (response) => {
                    if (signal.aborted) {
                        return;
                    }
                    const boxData = response.data.boxInfo;
                    const tip = `${boxData.name} <small class='ft__on-surface'>${boxData.hSize}</small>${boxData.docCount !== 0 ? window.scribli.languages.includeSubFile.replace("x", boxData.docCount) : ""}<br>${window.scribli.languages.modifiedAt} ${boxData.hMtime}<br>${window.scribli.languages.createdAt} ${boxData.hCtime}`;
                    showTooltip(tip, notebookItemElement as Element);
                    (notebookItemElement as HTMLElement).setAttribute("aria-label", tip);
                    if (tooltipAbortController === capturedController) {
                        tooltipAbortController = null;
                    }
                }, undefined, undefined, signal);
            }

            if (tip && !aElement.classList.contains("b3-tooltips")) {
                // 
                try {
                    showTooltip(decodeURIComponent(tip), aElement, tooltipClass, event, tooltipSpace);
                } catch (e) {
                    // 
                    showTooltip(tip, aElement, tooltipClass, event, tooltipSpace);
                }
                event.stopPropagation();
            } else {
                hideTooltip();
            }
        } else if (!aElement) {
            const tipElement = hasClosestByAttribute(event.target, "id", "tooltip", true);
            if (!tipElement || tipElement.clientHeight >= tipElement.scrollHeight) {
                hideTooltip();
            }
        }
        if (window.scribli.config.editor.floatWindowMode === 1 || window.scribli.shiftIsPressed) {
            clearTimeout(timeoutHide);
            timeoutHide = window.setTimeout(() => {
                hidePopover(event);
            }, Constants.TIMEOUT_INPUT);

            if (!getTarget(event, aElement)) {
                return;
            }
            // 
            if (event.relatedTarget && !document.contains(event.relatedTarget as Node)) {
                return;
            }
            if (window.scribli.ctrlIsPressed) {
                clearTimeout(timeoutHide);
                showPopover(app);
            } else if (window.scribli.shiftIsPressed) {
                clearTimeout(timeoutHide);
                showPopover(app, true);
            }
            return;
        }

        clearTimeout(timeout);
        clearTimeout(timeoutHide);
        timeoutHide = window.setTimeout(() => {
            if (!hidePopover(event)) {
                return;
            }
            if (!popoverTargetElement && !aElement) {
                clearTimeout(timeout);
            }
        }, Constants.TIMEOUT_INPUT);
        timeout = window.setTimeout(() => {
            if (!getTarget(event, aElement) || isTouchDevice()) {
                return;
            }
            clearTimeout(timeoutHide);
            showPopover(app);
        }, window.scribli.config.editor.floatWindowDelay);
    });
};

const hidePopover = (event: MouseEvent & { path: HTMLElement[] }) => {
    const target = isTouchDevice() ? document.elementFromPoint(event.clientX, event.clientY) : event.target as HTMLElement;
    if (!target) {
        return false;
    }
    if ((target.id && target.tagName !== "svg" && (target.id.startsWith("minder_node") || target.id.startsWith("kity_") || target.id.startsWith("node_")))
        || target.classList.contains("counter")
        || target.tagName === "circle"
        || target.closest('.protyle-icon[data-action="openFloat"]')
    ) {
        return false;
    }

    const avPanelElement = hasClosestByClassName(target, "av__panel") || hasClosestByClassName(target, "av__mask");
    if (avPanelElement) {
        const blockPanel = window.scribli.blockPanels.find((item) => {
            if (item.element.style.zIndex < avPanelElement.style.zIndex) {
                return true;
            }
        });
        if (blockPanel) {
            return false;
        }
    } else {
        const menuElement = hasClosestByClassName(target, "b3-menu");
        if (menuElement && menuElement.getAttribute("data-name") !== Constants.MENU_DOC_TREE_MORE) {
            const blockPanel = window.scribli.blockPanels.find((item) => {
                if (item.element.style.zIndex < menuElement.style.zIndex) {
                    return true;
                }
            });
            if (blockPanel) {
                return false;
            }
        }
    }
    popoverTargetElement = hasClosestByAttribute(target, "data-type", "block-ref") as HTMLElement ||
        hasClosestByAttribute(target, "data-type", "virtual-block-ref") as HTMLElement;
    if (popoverTargetElement && popoverTargetElement.classList.contains("b3-tooltips")) {
        popoverTargetElement = undefined;
    }
    if (!popoverTargetElement) {
        popoverTargetElement = hasClosestByClassName(target, "popover__block") as HTMLElement;
    }
    const linkElement = hasClosestByAttribute(target, "data-type", "a", true);
    if (!popoverTargetElement && linkElement && parseScribliUriInfo(linkElement.getAttribute("data-href"))) {
        popoverTargetElement = linkElement;
    }
    if (!popoverTargetElement || (popoverTargetElement && window.scribli.menus.menu.data && window.scribli.menus.menu.data === popoverTargetElement)) {
        // 
        let targetElement = target;
        if (!targetElement.parentElement && event.path && event.path[1]) {
            targetElement = event.path[1];
        }
        const blockElement = hasClosestByClassName(targetElement, "block__popover", true);
        const maxEditLevels: { [key: string]: number } = {oid: 0};
        window.scribli.blockPanels.forEach((item) => {
            if ((item.targetElement || typeof item.x === "number") && item.element.getAttribute("data-pin") === "true") {
                const level = parseInt(item.element.getAttribute("data-level"));
                const oid = item.element.getAttribute("data-oid");
                if (maxEditLevels[oid]) {
                    if (level > maxEditLevels[oid]) {
                        maxEditLevels[oid] = level;
                    }
                } else {
                    maxEditLevels[oid] = level;
                }
            }
        });
        const menuLevel = parseInt(window.scribli.menus.menu.element.dataset.from);
        if (blockElement) {
            for (let i = window.scribli.blockPanels.length - 1; i >= 0; i--) {
                const item = window.scribli.blockPanels[i];
                const itemLevel = parseInt(item.element.getAttribute("data-level"));
                if ((item.targetElement || typeof item.x === "number") &&
                    itemLevel > (maxEditLevels[item.element.getAttribute("data-oid")] || 0) &&
                    item.element.getAttribute("data-pin") === "false" &&
                    itemLevel > parseInt(blockElement.getAttribute("data-level"))) {
                    if (menuLevel && menuLevel >= itemLevel) {
                        break;
                    } else {
                        const hasToolbar = item.editors.find(editItem => {
                            if (!editItem.protyle.toolbar.subElement.classList.contains("fn__none")) {
                                return true;
                            }
                        });
                        if (hasToolbar) {
                            break;
                        }
                        item.destroy();
                    }
                }
            }
        } else {
            for (let i = window.scribli.blockPanels.length - 1; i >= 0; i--) {
                const item = window.scribli.blockPanels[i];
                const itemLevel = parseInt(item.element.getAttribute("data-level"));
                if ((item.targetElement || typeof item.x === "number") && item.element.getAttribute("data-pin") === "false") {
                    if (menuLevel && menuLevel >= itemLevel) {
                        break;
                    } else if (item.targetElement && item.targetElement.classList.contains("protyle-wysiwyg__embed") &&
                        item.targetElement.contains(targetElement)) {
                        break;
                    } else {
                        const hasToolbar = item.editors.find(editItem => {
                            if (!editItem.protyle.toolbar.subElement.classList.contains("fn__none")) {
                                return true;
                            }
                        });
                        if (hasToolbar) {
                            break;
                        }
                        item.destroy();
                    }
                }
            }
        }
    }
};

const getTarget = (event: MouseEvent & { target: HTMLElement }, aElement: false | HTMLElement) => {
    if (window.scribli.config.editor.floatWindowMode === 2 || hasClosestByClassName(event.target, "history__repo", true)) {
        return false;
    }
    popoverTargetElement = hasClosestByAttribute(event.target, "data-type", "block-ref") as HTMLElement ||
        hasClosestByAttribute(event.target, "data-type", "virtual-block-ref") as HTMLElement;
    if (popoverTargetElement && popoverTargetElement.classList.contains("b3-tooltips")) {
        popoverTargetElement = undefined;
    }
    if (!popoverTargetElement) {
        popoverTargetElement = hasClosestByClassName(event.target, "popover__block") as HTMLElement;
    }
    if (!popoverTargetElement && aElement) {
        if (parseScribliUriInfo(aElement.getAttribute("data-href")) && aElement.getAttribute("prevent-popover") !== "true") {
            popoverTargetElement = aElement;
        } else if (aElement.classList.contains("av__cell")) {
            const textElement = aElement.querySelector(".av__celltext--url") as HTMLElement;
            if (textElement && textElement.dataset.type === "url" && parseScribliUriInfo(textElement.dataset.href)) {
                popoverTargetElement = textElement;
            }
        }
    }
    if (!popoverTargetElement || window.scribli.altIsPressed ||
        (window.scribli.isPublish && popoverTargetElement.dataset.popoverUrl === "/api/av/getMirrorDatabaseBlocks") ||
        (window.scribli.config.editor.floatWindowMode === 0 && window.scribli.ctrlIsPressed) ||
        (popoverTargetElement && popoverTargetElement.getAttribute("prevent-popover") === "true")) {
        return false;
    }
    // 
    if (popoverTargetElement && getSelection().rangeCount > 0) {
        const range = getSelection().getRangeAt(0);
        if (range.toString() !== "" && popoverTargetElement.contains(range.startContainer)) {
            return false;
        }
    }
    return true;
};

export const showPopover = async (app: App, showRef = false) => {
    if (!popoverTargetElement || (window.scribli.menus.menu.data && window.scribli.menus.menu.data === popoverTargetElement)) {
        return;
    }
    let refDefs: IRefDefs[] = [];
    let originalRefBlockIDs: IObject;
    const dataId = popoverTargetElement.getAttribute("data-id");
    if (dataId) {
        if (showRef) {
            const postResponse = await fetchSyncPost("/api/block/getRefIDs", {id: dataId});
            refDefs = postResponse.data.refDefs;
            originalRefBlockIDs = postResponse.data.originalRefBlockIDs;
        } else {
            if (dataId.startsWith("[")) {
                JSON.parse(dataId).forEach((item: string) => {
                    refDefs.push({refID: item});
                });
            } else {
                refDefs = [{refID: dataId}];
            }
        }
    } else if (popoverTargetElement.getAttribute("data-type")?.indexOf("virtual-block-ref") > -1) {
        const postResponse = await fetchSyncPost("/api/block/getBlockDefIDsByRefText", {
            anchor: popoverTargetElement.textContent,
        });
        refDefs = postResponse.data.refDefs;
    } else if (popoverTargetElement.getAttribute("data-type")?.split(" ").includes("a")) {
        const blockInfo = parseScribliUriInfo(popoverTargetElement.getAttribute("data-href"));
        refDefs = [{
            refID: blockInfo?.id ?? "",
            avItemID: blockInfo?.avItemID,
            avViewID: blockInfo?.avViewID,
            avGroupID: blockInfo?.avGroupID,
        }];
    } else if (popoverTargetElement.dataset.type === "url") {
        const blockInfo = parseScribliUriInfo(popoverTargetElement.dataset.href || popoverTargetElement.textContent.trim());
        refDefs = [{
            refID: blockInfo?.id ?? "",
            avItemID: blockInfo?.avItemID,
            avViewID: blockInfo?.avViewID,
            avGroupID: blockInfo?.avGroupID,
        }];
    } else if (popoverTargetElement.dataset.popoverUrl) {
        const postResponse = await fetchSyncPost(popoverTargetElement.dataset.popoverUrl, {avID: popoverTargetElement.dataset.avId});
        refDefs = postResponse.data.refDefs;
    } else {
        // pdf
        let targetId;
        let url = "/api/block/getRefIDs";
        if (popoverTargetElement.classList.contains("protyle-attr--refcount")) {
            targetId = popoverTargetElement.parentElement.parentElement.getAttribute("data-node-id");
        } else if (popoverTargetElement.classList.contains("pdf__rect")) {
            const relationIds = popoverTargetElement.getAttribute("data-relations");
            if (relationIds) {
                relationIds.split(",").forEach((item: string) => {
                    refDefs.push({refID: item});
                });
                url = "";
            } else {
                targetId = popoverTargetElement.getAttribute("data-node-id");
                url = "/api/block/getRefIDsByFileAnnotationID";
            }
        } else if (!targetId) {
            targetId = popoverTargetElement.parentElement.getAttribute("data-node-id");
        }
        if (url) {
            const postResponse = await fetchSyncPost(url, {id: targetId});
            refDefs = postResponse.data.refDefs;
            originalRefBlockIDs = postResponse.data.originalRefBlockIDs;
        }
    }

    if (refDefs.length === 0) {
        return;
    }

    let hasPin = false;
    window.scribli.blockPanels.find((item) => {
        if ((item.targetElement || typeof item.x === "number") && item.element.getAttribute("data-pin") === "true"
            && JSON.stringify(refDefs) === JSON.stringify(item.refDefs)) {
            hasPin = true;
            return true;
        }
    });
    if (!hasPin && popoverTargetElement.parentElement &&
        popoverTargetElement.parentElement.style.opacity !== "0.38"
    ) {
        window.scribli.blockPanels.push(new BlockPanel({
            app,
            targetElement: popoverTargetElement,
            isBacklink: showRef || popoverTargetElement.classList.contains("protyle-attr--refcount") || popoverTargetElement.classList.contains("counter"),
            refDefs,
            originalRefBlockIDs,
        }));
    }
};

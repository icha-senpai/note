import {fetchPost} from "../util/fetch";
import {focusByRange, setLastNodeRange} from "../protyle/util/selection";
import {insertHTML} from "../protyle/util/insertHTML";
import {Dialog} from "../dialog";
import {isMobile} from "../util/functions";
import {getContenteditableElement} from "../protyle/wysiwyg/getBlock";
import {blockRender} from "../protyle/render/blockRender";
import {processRender} from "../protyle/util/processCode";
import {highlightRender} from "../protyle/render/highlightRender";
import {Constants} from "../constants";
import {setStorageVal} from "../protyle/util/compatibility";
import {escapeAriaLabel, escapeAttr, escapeHtml} from "../util/escape";
import {showMessage} from "../dialog/message";
import {Menu} from "../plugin/Menu";
import {upDownHint} from "../util/upDownHint";

export const fillContent = (protyle: IProtyle, data: string, elements: Element[]) => {
    if (!data) {
        return;
    }
    setLastNodeRange(getContenteditableElement(elements[elements.length - 1]), protyle.toolbar.range);
    protyle.toolbar.range.collapse(true);
    insertHTML(protyle.lute.SpinBlockDOM(data), protyle, true, true);
    blockRender(protyle, protyle.wysiwyg.element);
    processRender(protyle.wysiwyg.element);
    highlightRender(protyle.wysiwyg.element);
};

const editDialog = (customName: string, customMemo: string) => {
    const dialog = new Dialog({
        title: window.scribli.languages.update,
        content: `<div class="b3-dialog__content">
    <input class="b3-text-field fn__block" placeholder="${window.scribli.languages.memo}">
    <div class="fn__hr"></div>
    <textarea class="b3-text-field fn__block" placeholder="${window.scribli.languages.aiCustomAction}"></textarea>
</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--remove">${window.scribli.languages.delete}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages.confirm}</button>
</div>`,
        width: isMobile() ? "92vw" : "520px",
    });
    dialog.element.setAttribute("data-key", Constants.DIALOG_AIUPDATECUSTOMACTION);
    const nameElement = dialog.element.querySelector("input");
    nameElement.value = customName;
    const customElement = dialog.element.querySelector("textarea");
    const btnsElement = dialog.element.querySelectorAll(".b3-button");
    dialog.bindInput(customElement, () => {
        (btnsElement[2] as HTMLButtonElement).click();
    });
    customElement.value = customMemo;
    btnsElement[1].addEventListener("click", () => {
        dialog.destroy();
    });
    btnsElement[2].addEventListener("click", () => {
        window.scribli.storage[Constants.LOCAL_AI].find((subItem: {
            name: string,
            memo: string
        }) => {
            if (customName === subItem.name && customMemo === subItem.memo) {
                subItem.name = nameElement.value;
                subItem.memo = customElement.value;
                setStorageVal(Constants.LOCAL_AI, window.scribli.storage[Constants.LOCAL_AI]);
                return true;
            }
        });
        dialog.destroy();
    });
    btnsElement[0].addEventListener("click", () => {
        window.scribli.storage[Constants.LOCAL_AI].find((subItem: {
            name: string,
            memo: string
        }, index: number) => {
            if (customName === subItem.name && customMemo === subItem.memo) {
                window.scribli.storage[Constants.LOCAL_AI].splice(index, 1);
                setStorageVal(Constants.LOCAL_AI, window.scribli.storage[Constants.LOCAL_AI]);
                return true;
            }
        });
        dialog.destroy();
    });
    nameElement.focus();
};

const customDialog = (protyle: IProtyle, ids: string[], elements: Element[]) => {
    const dialog = new Dialog({
        title: window.scribli.languages.aiCustomAction,
        content: `<div class="b3-dialog__content">
    <input class="b3-text-field fn__block" value="" placeholder="${window.scribli.languages.memo}">
    <div class="fn__hr"></div>
    <textarea class="b3-text-field fn__block" placeholder="${window.scribli.languages.aiCustomAction}"></textarea>
</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages.use}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages.save}</button>
</div>`,
        width: isMobile() ? "92vw" : "520px",
    });
    dialog.element.setAttribute("data-key", Constants.DIALOG_AICUSTOMACTION);
    const nameElement = dialog.element.querySelector("input");
    const customElement = dialog.element.querySelector("textarea");
    const btnsElement = dialog.element.querySelectorAll(".b3-button");
    dialog.bindInput(customElement, () => {
        (btnsElement[1] as HTMLButtonElement).click();
    });
    btnsElement[0].addEventListener("click", () => {
        dialog.destroy();
    });
    btnsElement[1].addEventListener("click", () => {
        if (!customElement.value) {
            showMessage(window.scribli.languages["_kernel"][142]);
            return;
        }
        fetchPost("/api/ai/chatGPTWithAction", {
            ids,
            action: customElement.value,
        }, (response) => {
            dialog.destroy();
            fillContent(protyle, response.data, elements);
        });
    });
    btnsElement[2].addEventListener("click", () => {
        if (!nameElement.value && !customElement.value) {
            showMessage(window.scribli.languages["_kernel"][142]);
            return;
        }
        window.scribli.storage[Constants.LOCAL_AI].push({
            name: nameElement.value,
            memo: customElement.value
        });
        setStorageVal(Constants.LOCAL_AI, window.scribli.storage[Constants.LOCAL_AI]);
        dialog.destroy();
    });
    nameElement.focus();
};

const filterAI = (element: HTMLElement, inputElement: HTMLInputElement) => {
    element.querySelectorAll(".b3-list-item").forEach(item => {
        if (item.textContent.indexOf(inputElement.value) > -1) {
            item.classList.remove("fn__none");
        } else {
            item.classList.add("fn__none");
        }
    });
    element.querySelectorAll(".b3-menu__separator").forEach(item => {
        if (inputElement.value) {
            item.classList.add("fn__none");
        } else {
            item.classList.remove("fn__none");
        }
    });
    element.querySelector(".b3-list-item--focus")?.classList.remove("b3-list-item--focus");
    element.querySelector(".b3-list-item:not(.fn__none)")?.classList.add("b3-list-item--focus");
};

export const AIActions = (elements: Element[], protyle: IProtyle) => {
    window.scribli.menus.menu.remove();
    const ids: string[] = [];
    elements.forEach(item => {
        ids.push(item.getAttribute("data-node-id"));
    });
    const menu = new Menu(Constants.MENU_AI, () => {
        focusByRange(protyle.toolbar.range);
    });
    let customHTML = "";
    window.scribli.storage[Constants.LOCAL_AI].forEach((item: { name: string, memo: string }, index: number) => {
        customHTML += `<div data-action="${escapeAttr(item.memo || item.name)}" data-position="10west" data-index="${index}" class="b3-list-item b3-list-item--narrow ariaLabel" aria-label="${escapeAriaLabel(item.memo)}">
    <span class="b3-list-item__text">${escapeHtml(item.name)}</span>
    <span data-type="edit" class="b3-list-item__action"><svg><use xlink:href="#iconEdit"></use></svg></span>
</div>`;
    });
    if (customHTML) {
        customHTML = `<div class="b3-menu__separator"></div>${customHTML}`;
    }
    const clearContext = "Clear context";
    menu.addItem({
        iconHTML: "",
        type: "empty",
        label: `<div class="fn__flex-column b3-menu__filter">
    <input class="b3-text-field fn__flex-shrink"/>
    <div class="fn__hr"></div>
    <div class="b3-list fn__flex-1 b3-list--background">
       <div class="b3-list-item b3-list-item--narrow b3-list-item--focus" data-action="Continue writing">
            ${window.scribli.languages.aiContinueWrite}
        </div>
        <div class="b3-menu__separator"></div>
        <div class="b3-list-item b3-list-item--narrow" data-action="${window.scribli.languages.aiExtractSummary}">
            ${window.scribli.languages.aiExtractSummary}
        </div>
        <div class="b3-list-item b3-list-item--narrow" data-action="${window.scribli.languages.aiBrainStorm}">
            ${window.scribli.languages.aiBrainStorm}
        </div>
        <div class="b3-list-item b3-list-item--narrow" data-action="${window.scribli.languages.aiFixGrammarSpell}">
            ${window.scribli.languages.aiFixGrammarSpell}
        </div>
        <div class="b3-list-item b3-list-item--narrow" data-action="${clearContext}">
            ${window.scribli.languages.clearContext}
        </div>
        <div class="b3-menu__separator"></div>
        <div class="b3-list-item b3-list-item--narrow" data-type="custom">
            ${window.scribli.languages.aiCustomAction}
        </div>
        ${customHTML}
    </div>
</div>`,
        bind(element) {
            const listElement = element.querySelector(".b3-list");
            const inputElement = element.querySelector("input");
            inputElement.addEventListener("keydown", (event: KeyboardEvent) => {
                if (event.isComposing) {
                    return;
                }
                const currentElement = upDownHint(listElement, event);
                if (currentElement) {
                    event.stopPropagation();
                }
                if (event.key === "Enter") {
                    event.preventDefault();
                    event.stopPropagation();
                    const currentElement = listElement.querySelector(".b3-list-item--focus") as HTMLElement;
                    if (!currentElement) {
                        return;
                    }
                    if (currentElement.dataset.type === "custom") {
                        customDialog(protyle, ids, elements);
                        menu.close();
                    } else {
                        fetchPost("/api/ai/chatGPTWithAction", {
                            ids,
                            action: currentElement.dataset.action
                        }, (response) => {
                            fillContent(protyle, response.data, elements);
                        });
                        if (currentElement.dataset.action === clearContext) {
                            showMessage(window.scribli.languages.clearContextSucc);
                        } else {
                            menu.close();
                        }
                    }
                }
            });
            inputElement.addEventListener("compositionend", () => {
                filterAI(element, inputElement);
            });
            inputElement.addEventListener("input", (event: KeyboardEvent) => {
                if (event.isComposing) {
                    return;
                }
                filterAI(element, inputElement);
            });
            element.addEventListener("click", (event) => {
                let target = event.target as HTMLElement;
                while (target && (target !== element)) {
                    if (target.classList.contains("b3-list-item__action")) {
                        const subItem = window.scribli.storage[Constants.LOCAL_AI][target.parentElement.dataset.index];
                        editDialog(subItem.name, subItem.memo);
                        menu.close();
                        event.stopPropagation();
                        event.preventDefault();
                        break;
                    } else if (target.classList.contains("b3-list-item")) {
                        if (target.dataset.type === "custom") {
                            customDialog(protyle, ids, elements);
                            menu.close();
                        } else {
                            fetchPost("/api/ai/chatGPTWithAction", {ids, action: target.dataset.action}, (response) => {
                                fillContent(protyle, response.data, elements);
                            });
                            if (target.dataset.action === clearContext) {
                                showMessage(window.scribli.languages.clearContextSucc);
                            } else {
                                menu.close();
                            }
                        }
                        event.stopPropagation();
                        event.preventDefault();
                        break;
                    }
                    target = target.parentElement;
                }
            });
        }
    });
    menu.element.querySelector(".b3-menu__items").setAttribute("style", "overflow: initial");
    const rect = elements[elements.length - 1].getBoundingClientRect();
    menu.open({
        x: rect.left,
        y: rect.bottom,
        h: rect.height,
    });
    menu.element.querySelector("input").focus();
};

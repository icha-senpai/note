/// #if !BROWSER
import {shell} from "electron";
/// #endif
import {confirmDialog} from "../dialog/confirmDialog";
import {getSearch, isMobile, isValidCustomAttrName} from "../util/functions";
import {isEncryptedBox, isLocalPath, movePathTo, moveToPath, pathPosix} from "../util/pathName";
import {MenuItem} from "./Menu";
import {onExport, saveExport} from "../protyle/export";
import {exportMarkdownZip} from "../protyle/export/exportMd";
import {saveExportFile, writeText} from "../protyle/util/compatibility";
import {openByMobile} from "../editor/openLink";
import {processScribliUri} from "../util/uri";
import {fetchPost, fetchSyncPost} from "../util/fetch";
import {hideMessage, showMessage} from "../dialog/message";
import {Dialog} from "../dialog";
import {focusBlock, focusByRange, getEditorRange} from "../protyle/util/selection";
import {openAsset, openBy} from "../editor/util";
import {rename, replaceFileName} from "../editor/rename";
import {Constants} from "../constants";
import {exportImage} from "../protyle/export/util";
import {App} from "../index";
import {renderAVAttribute} from "../protyle/render/av/blockAttr";
import {openAssetNewWindow} from "../window/openNewWindow";
import {copyTextByType} from "../protyle/toolbar/util";
import {hideElements} from "../protyle/ui/hideElements";
import {Protyle} from "../protyle";
import {getAllEditor} from "../layout/getAll";
import {hasClosestByClassName} from "../protyle/util/hasClosest";

const bindAttrInput = (inputElement: HTMLInputElement, id: string) => {
    inputElement.addEventListener("change", () => {
        fetchPost("/api/attr/setBlockAttrs", {
            id,
            attrs: {[inputElement.dataset.name]: inputElement.value}
        });
    });
};

export const openFileAttr = (attrs: Record<string, string>, focusName = "bookmark", protyle?: IProtyle) => {
    let customHTML = "";
    let hasAV = false;
    const range = getSelection().rangeCount > 0 ? getSelection().getRangeAt(0) : null;
    let ghostProtyle: Protyle;
    if (!protyle) {
        getAllEditor().find(item => {
            if (attrs.id === item.protyle.block.rootID) {
                protyle = item.protyle;
                return true;
            }
        });
        if (!protyle) {
            ghostProtyle = new Protyle(window.scribli.ws.app, document.createElement("div"), {
                blockId: attrs.id,
            });
        }
    }
    Object.keys(attrs).forEach(item => {
        if (Constants.CUSTOM_RIFF_DECKS === item || item.startsWith("custom-sy-")) {
            return;
        }
        if (item.indexOf("custom-av") > -1) {
            hasAV = true;
        } else if (item.indexOf("custom") > -1) {
            customHTML += `<label class="b3-label b3-label--noborder">
     <div class="fn__flex">
        <span class="fn__flex-1">${item.replace("custom-", "")}</span>
        <span data-action="remove" class="block__icon block__icon--show"><svg><use xlink:href="#iconMin"></use></svg></span>
    </div>
    <div class="fn__hr"></div>
    <textarea style="resize: vertical;" spellcheck="false" class="b3-text-field fn__block" rows="1" data-name="${item}"></textarea>
</label>`;
        }
    });
    const dialog = new Dialog({
        width: isMobile() ? "100vw" : "50vw",
        containerClassName: "b3-dialog__container--theme",
        height: isMobile() ? "100vh" : "80vh",
        content: `<div class="fn__flex-column">
    <div class="layout-tab-bar fn__flex" style="${isMobile() ? "padding-right: 38px;" : ""}flex-shrink:0;border-radius: var(--b3-border-radius-b) var(--b3-border-radius-b) 0 0">
        <div class="item item--full item--focus" data-type="attr">
            <span class="fn__flex-1"></span>
            <span class="item__text">${window.scribli.languages.builtIn}</span>
            <span class="fn__flex-1"></span>
        </div>
        <div class="item item--full${hasAV ? "" : " fn__none"}" data-type="NodeAttributeView">
            <span class="fn__flex-1"></span>
            <span class="item__text">${window.scribli.languages.database}</span>
            <span class="fn__flex-1"></span>
        </div>
        <div class="item item--full" data-type="custom">
            <span class="fn__flex-1"></span>
            <span class="item__text">${window.scribli.languages.custom}</span>
            <span class="fn__flex-1"></span>
        </div>
    </div>
    <div class="fn__flex-1">
        <div class="custom-attr" data-type="attr">
            <label class="b3-label b3-label--noborder">
                <div class="fn__flex">
                    <span class="fn__flex-1">${window.scribli.languages.bookmark}</span>
                    <span data-action="bookmark" class="block__icon block__icon--show"><svg><use xlink:href="#iconDown"></use></svg></span>
                </div>
                <div class="fn__hr"></div>
                <input spellcheck="${window.scribli.config.editor.spellcheck}" class="b3-text-field fn__block" placeholder="${window.scribli.languages.attrBookmarkTip}" data-name="bookmark">
            </label>
            <label class="b3-label b3-label--noborder">
                ${window.scribli.languages.name}
                <div class="fn__hr"></div>
                <input spellcheck="${window.scribli.config.editor.spellcheck}" class="b3-text-field fn__block" placeholder="${window.scribli.languages.attrNameTip}" data-name="name">
            </label>
            <label class="b3-label b3-label--noborder">
                ${window.scribli.languages.alias}
                <div class="fn__hr"></div>
                <input spellcheck="${window.scribli.config.editor.spellcheck}" class="b3-text-field fn__block" placeholder="${window.scribli.languages.attrAliasTip}" data-name="alias">
            </label>
            <label class="b3-label b3-label--noborder">
                ${window.scribli.languages.memo}
                <div class="fn__hr"></div>
                <textarea style="resize: vertical" spellcheck="${window.scribli.config.editor.spellcheck}" class="b3-text-field fn__block" placeholder="${window.scribli.languages.attrMemoTip}" rows="2" data-name="memo"></textarea>
            </label>
        </div>
        <div data-type="NodeAttributeView" class="fn__none custom-attr"></div>
        <div data-type="custom" class="fn__none custom-attr">
           ${customHTML}
           <div class="b3-label">
               <button data-action="addCustom" class="b3-button b3-button--cancel">
                   <svg><use xlink:href="#iconAdd"></use></svg>${window.scribli.languages.addAttr}
               </button>
           </div>
        </div>
    </div>
</div>`,
        destroyCallback() {
            focusByRange(range);
            if (protyle) {
                hideElements(["select"], protyle);
            } else {
                ghostProtyle.destroy();
            }
        }
    });
    dialog.element.setAttribute("data-key", Constants.DIALOG_ATTR);
    (dialog.element.querySelector('.b3-text-field[data-name="bookmark"]') as HTMLInputElement).value = attrs.bookmark || "";
    (dialog.element.querySelector('.b3-text-field[data-name="name"]') as HTMLInputElement).value = attrs.name || "";
    (dialog.element.querySelector('.b3-text-field[data-name="alias"]') as HTMLInputElement).value = attrs.alias || "";
    (dialog.element.querySelector('.b3-text-field[data-name="memo"]') as HTMLInputElement).value = attrs.memo || "";
    dialog.element.querySelectorAll('.custom-attr[data-type="custom"] textarea.b3-text-field').forEach((item: HTMLTextAreaElement) => {
        item.value = attrs[item.dataset.name];
    });
    dialog.element.addEventListener("click", (event) => {
        let target = event.target as HTMLElement;
        if (typeof event.detail === "string") {
            target = dialog.element.querySelector(`.item--full[data-type="${event.detail}"]`);
        }
        while (target !== dialog.element) {
            const type = target.dataset.action;
            if (target.classList.contains("item--full")) {
                target.parentElement.querySelector(".item--focus").classList.remove("item--focus");
                target.classList.add("item--focus");
                dialog.element.querySelectorAll(".custom-attr").forEach((item: HTMLElement) => {
                    if (item.dataset.type === target.dataset.type) {
                        if (item.dataset.type === "NodeAttributeView" && item.innerHTML === "") {
                            renderAVAttribute(item, attrs.id, protyle || ghostProtyle.protyle);
                        }
                        item.classList.remove("fn__none");
                    } else {
                        item.classList.add("fn__none");
                    }
                });
            } else if (type === "remove") {
                fetchPost("/api/attr/setBlockAttrs", {
                    id: attrs.id,
                    attrs: {["custom-" + target.previousElementSibling.textContent]: ""}
                });
                target.parentElement.parentElement.remove();
                event.stopPropagation();
                event.preventDefault();
                break;
            } else if (type === "bookmark") {
                fetchPost("/api/attr/getBookmarkLabels", {}, (response) => {
                    window.scribli.menus.menu.remove();
                    if (response.data.length === 0) {
                        window.scribli.menus.menu.append(new MenuItem({
                            id: "emptyContent",
                            iconHTML: "",
                            label: window.scribli.languages.emptyContent,
                            type: "readonly",
                        }).element);
                    } else {
                        response.data.forEach((item: string) => {
                            window.scribli.menus.menu.append(new MenuItem({
                                label: item,
                                click() {
                                    const bookmarkInputElement = target.parentElement.parentElement.querySelector("input");
                                    bookmarkInputElement.value = item;
                                    bookmarkInputElement.dispatchEvent(new CustomEvent("change"));
                                }
                            }).element);
                        });
                    }
                    window.scribli.menus.menu.element.classList.add("b3-menu--list");
                    window.scribli.menus.menu.popup({x: event.clientX, y: event.clientY + 16, w: 16});
                });
                event.stopPropagation();
                event.preventDefault();
                break;
            } else if (type === "addCustom") {
                const addDialog = new Dialog({
                    title: window.scribli.languages.attrName,
                    content: `<div class="b3-dialog__content"><input spellcheck="false" class="b3-text-field fn__block" value=""></div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages.confirm}</button>
</div>`,
                    width: isMobile() ? "92vw" : "520px",
                });
                addDialog.element.setAttribute("data-key", Constants.DIALOG_SETCUSTOMATTR);
                const inputElement = addDialog.element.querySelector("input") as HTMLInputElement;
                const btnsElement = addDialog.element.querySelectorAll(".b3-button");
                addDialog.bindInput(inputElement, () => {
                    (btnsElement[1] as HTMLButtonElement).click();
                });
                inputElement.focus();
                inputElement.select();
                btnsElement[0].addEventListener("click", () => {
                    addDialog.destroy();
                });
                btnsElement[1].addEventListener("click", () => {
                    const value = inputElement.value.toLowerCase();
                    if (!isValidCustomAttrName(value)) {
                        showMessage(window.scribli.languages._kernel[25]);
                        return false;
                    }
                    let existElement: HTMLElement | false;
                    Array.from(dialog.element.querySelectorAll('.custom-attr[data-type="custom"] .b3-label .fn__flex-1')).find((labelItem: HTMLElement) => {
                        if (labelItem.textContent === value) {
                            existElement = hasClosestByClassName(labelItem, "b3-label");
                            return true;
                        }
                    });
                    if (existElement) {
                        showMessage(window.scribli.languages.hasAttrName.replace("${x}", value));
                    } else {
                        target.parentElement.insertAdjacentHTML("beforebegin", `<div class="b3-label b3-label--noborder">
    <div class="fn__flex">
        <span class="fn__flex-1">${value}</span>
        <span data-action="remove" class="block__icon block__icon--show"><svg><use xlink:href="#iconMin"></use></svg></span>
    </div>
    <div class="fn__hr"></div>
    <textarea style="resize: vertical" spellcheck="false" data-name="custom-${value}" class="b3-text-field fn__block" rows="1" placeholder="${window.scribli.languages.attrValue1}"></textarea>
</div>`);
                        const newInputElement = target.parentElement.previousElementSibling.querySelector(".b3-text-field") as HTMLInputElement;
                        newInputElement.focus();
                        bindAttrInput(newInputElement, attrs.id);
                        addDialog.destroy();
                    }
                });
                event.stopPropagation();
                event.preventDefault();
                break;
            }
            target = target.parentElement;
        }
    });
    dialog.element.querySelectorAll(".b3-text-field").forEach((item: HTMLInputElement) => {
        if (focusName !== "av" && focusName !== "custom" && focusName === item.getAttribute("data-name")) {
            item.focus();
        }
        bindAttrInput(item, attrs.id);
    });
    if (focusName === "av") {
        dialog.element.dispatchEvent(new CustomEvent("click", {detail: "NodeAttributeView"}));
        (document.activeElement as HTMLElement)?.blur();
    } else if (focusName === "custom") {
        dialog.element.dispatchEvent(new CustomEvent("click", {detail: "custom"}));
    }
};

export const openAttr = (nodeElement: Element, focusName = "bookmark", protyle: IProtyle) => {
    if (nodeElement.getAttribute("data-type") === "NodeThematicBreak") {
        return;
    }
    const id = nodeElement.getAttribute("data-node-id");
    fetchPost("/api/attr/getBlockAttrs", {id}, (response) => {
        openFileAttr(response.data, focusName, protyle);
    });
};

export const copySubMenu = (ids: string[], accelerator = true, focusElement?: Element, stdMarkdownId?: string): IMenu[] => {
    const menuItems = [{
        id: "copyBlockRef",
        iconHTML: "",
        accelerator: accelerator ? window.scribli.config.keymap.editor.general.copyBlockRef.custom : undefined,
        label: window.scribli.languages.copyBlockRef,
        click: () => {
            copyTextByType(ids, "ref");
            if (focusElement) {
                focusBlock(focusElement);
            }
        }
    }, {
        id: "copyBlockEmbed",
        iconHTML: "",
        label: window.scribli.languages.copyBlockEmbed,
        accelerator: accelerator ? window.scribli.config.keymap.editor.general.copyBlockEmbed.custom : undefined,
        click: () => {
            copyTextByType(ids, "blockEmbed");
            if (focusElement) {
                focusBlock(focusElement);
            }
        }
    }, {
        id: "copyProtocol",
        iconHTML: "",
        label: window.scribli.languages.copyProtocol,
        accelerator: accelerator ? window.scribli.config.keymap.editor.general.copyProtocol.custom : undefined,
        click: () => {
            copyTextByType(ids, "protocol");
            if (focusElement) {
                focusBlock(focusElement);
            }
        }
    }, {
        id: "copyProtocolInMd",
        iconHTML: "",
        label: window.scribli.languages.copyProtocolInMd,
        accelerator: accelerator ? window.scribli.config.keymap.editor.general.copyProtocolInMd.custom : undefined,
        click: () => {
            copyTextByType(ids, "protocolMd");
            if (focusElement) {
                focusBlock(focusElement);
            }
        }
    },
        /// #if BROWSER
        {
            id: "copyWebURL",
            iconHTML: "",
            label: window.scribli.languages.copyWebURL,
            click: () => {
                copyTextByType(ids, "webURL");
                if (focusElement) {
                    focusBlock(focusElement);
                }
            }
        },
        /// #endif
        {
            id: "copyHPath",
            iconHTML: "",
            label: window.scribli.languages.copyHPath,
            accelerator: accelerator ? window.scribli.config.keymap.editor.general.copyHPath.custom : undefined,
            click: () => {
                copyTextByType(ids, "hPath");
                if (focusElement) {
                    focusBlock(focusElement);
                }
            }
        }, {
            id: "copyID",
            iconHTML: "",
            label: window.scribli.languages.copyID,
            accelerator: accelerator ? window.scribli.config.keymap.editor.general.copyID.custom : undefined,
            click: () => {
                copyTextByType(ids, "id");
                if (focusElement) {
                    focusBlock(focusElement);
                }
            }
        }];

    if (stdMarkdownId) {
        menuItems.push({
            id: "copyMarkdown",
            iconHTML: "",
            label: window.scribli.languages.copyMarkdown,
            accelerator: undefined,
            click: async () => {
                const response = await fetchSyncPost("/api/export/exportMdContent", {
                    id: stdMarkdownId,
                    refMode: 3,
                    embedMode: 1,
                    yfm: false,
                    fillCSSVar: false,
                    adjustHeadingLevel: false
                });
                const text = response.data.content;
                writeText(text);
                if (focusElement) {
                    focusBlock(focusElement);
                }
            }
        });
    }

    return menuItems;
};

export const exportMd = (id: string) => {
    if (window.scribli.isPublish) {
        return;
    }
    return new MenuItem({
        id: "export",
        label: window.scribli.languages.export,
        type: "submenu",
        icon: "iconUpload",
        submenu: [{
            id: "exportTemplate",
            label: window.scribli.languages.template,
            iconClass: "ft__error",
            icon: "iconMarkdown",
            click: async () => {
                const result = await fetchSyncPost("/api/block/getRefText", {id: id});

                const dialog = new Dialog({
                    title: window.scribli.languages.fileName,
                    content: `<div class="b3-dialog__content"><input class="b3-text-field fn__block" value=""></div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages.confirm}</button>
</div>`,
                    width: isMobile() ? "92vw" : "520px",
                });
                dialog.element.setAttribute("data-key", Constants.DIALOG_EXPORTTEMPLATE);
                const inputElement = dialog.element.querySelector("input") as HTMLInputElement;
                const btnsElement = dialog.element.querySelectorAll(".b3-button");
                dialog.bindInput(inputElement, () => {
                    (btnsElement[1] as HTMLButtonElement).click();
                });
                let name = replaceFileName(result.data);
                const maxNameLen = 32;
                if (name.length > maxNameLen) {
                    name = name.substring(0, maxNameLen);
                }
                inputElement.value = name;
                inputElement.focus();
                inputElement.select();
                btnsElement[0].addEventListener("click", () => {
                    dialog.destroy();
                });
                btnsElement[1].addEventListener("click", () => {
                    if (inputElement.value.trim() === "") {
                        inputElement.value = window.scribli.languages.untitled;
                    } else {
                        inputElement.value = replaceFileName(inputElement.value);
                    }

                    if (name.length > maxNameLen) {
                        name = name.substring(0, maxNameLen);
                    }

                    fetchPost("/api/template/docSaveAsTemplate", {
                        id,
                        name: inputElement.value,
                        overwrite: false
                    }, response => {
                        if (response.code === 1) {
                            confirmDialog(window.scribli.languages.export, window.scribli.languages.exportTplTip, () => {
                                fetchPost("/api/template/docSaveAsTemplate", {
                                    id,
                                    name: inputElement.value,
                                    overwrite: true
                                }, resp => {
                                    if (resp.code === 0) {
                                        showMessage(window.scribli.languages.exportTplSucc);
                                    }
                                });
                            });
                            return;
                        }
                        showMessage(window.scribli.languages.exportTplSucc);
                    });
                    dialog.destroy();
                });
            }
        }, {
            id: "exportScribliZip",
            label: "Scribli .sy.zip",
            icon: "iconScribli",
            click: () => {
                const msgId = showMessage(window.scribli.languages.exporting, -1);
                fetchPost("/api/export/exportSY", {
                    id,
                }, response => {
                    saveExportFile(response.data.zip, msgId);
                });
            }
        }, {
            id: "exportMarkdown",
            label: "Markdown .zip",
            icon: "iconMarkdown",
            click: () => {
                exportMarkdownZip({id});
            }
        }, {
            id: "exportImage",
            label: window.scribli.languages.image,
            icon: "iconImage",
            click: () => {
                exportImage(id);
            }
        },
            /// #if !BROWSER
            {
                id: "exportPDF",
                label: "PDF",
                icon: "iconPDF",
                click: () => {
                    saveExport({type: "pdf", id});
                }
            }, {
                id: "exportHTML_Scribli",
                label: "HTML (Scribli)",
                iconClass: "ft__error",
                icon: "iconHTML5",
                click: () => {
                    saveExport({type: "html", id});
                }
            }, {
                id: "exportHTML_Markdown",
                label: "HTML (Markdown)",
                icon: "iconHTML5",
                click: () => {
                    saveExport({type: "htmlmd", id});
                }
            }, {
                id: "exportWord",
                label: "Word .docx",
                icon: "iconDocx",
                click: () => {
                    saveExport({type: "word", id});
                }
            }, {
                id: "exportMore",
                label: window.scribli.languages.more,
                icon: "iconMore",
                type: "submenu",
                submenu: [{
                    id: "exportReStructuredText",
                    label: "reStructuredText",
                    iconHTML: "",
                    click: () => {
                        const msgId = showMessage(window.scribli.languages.exporting, -1);
                        fetchPost("/api/export/exportReStructuredText", {
                            id,
                        }, response => {
                            saveExportFile(response.data.zip, msgId);
                        });
                    }
                }, {
                    id: "exportAsciiDoc",
                    label: "AsciiDoc",
                    iconHTML: "",
                    click: () => {
                        const msgId = showMessage(window.scribli.languages.exporting, -1);
                        fetchPost("/api/export/exportAsciiDoc", {
                            id,
                        }, response => {
                            saveExportFile(response.data.zip, msgId);
                        });
                    }
                }, {
                    id: "exportTextile",
                    label: "Textile",
                    iconHTML: "",
                    click: () => {
                        const msgId = showMessage(window.scribli.languages.exporting, -1);
                        fetchPost("/api/export/exportTextile", {
                            id,
                        }, response => {
                            saveExportFile(response.data.zip, msgId);
                        });
                    }
                }, {
                    id: "exportOPML",
                    label: "OPML",
                    iconHTML: "",
                    click: () => {
                        const msgId = showMessage(window.scribli.languages.exporting, -1);
                        fetchPost("/api/export/exportOPML", {
                            id,
                        }, response => {
                            saveExportFile(response.data.zip, msgId);
                        });
                    }
                }, {
                    id: "exportOrgMode",
                    label: "Org-Mode",
                    iconHTML: "",
                    click: () => {
                        const msgId = showMessage(window.scribli.languages.exporting, -1);
                        fetchPost("/api/export/exportOrgMode", {
                            id,
                        }, response => {
                            saveExportFile(response.data.zip, msgId);
                        });
                    }
                }, {
                    id: "exportMediaWiki",
                    label: "MediaWiki",
                    iconHTML: "",
                    click: () => {
                        const msgId = showMessage(window.scribli.languages.exporting, -1);
                        fetchPost("/api/export/exportMediaWiki", {
                            id,
                        }, response => {
                            saveExportFile(response.data.zip, msgId);
                        });
                    }
                }, {
                    id: "exportODT",
                    label: "ODT",
                    iconHTML: "",
                    click: () => {
                        const msgId = showMessage(window.scribli.languages.exporting, -1);
                        fetchPost("/api/export/exportODT", {
                            id,
                        }, response => {
                            saveExportFile(response.data.zip, msgId);
                        });
                    }
                }, {
                    id: "exportRTF",
                    label: "RTF",
                    iconHTML: "",
                    click: () => {
                        const msgId = showMessage(window.scribli.languages.exporting, -1);
                        fetchPost("/api/export/exportRTF", {
                            id,
                        }, response => {
                            saveExportFile(response.data.zip, msgId);
                        });
                    }
                }, {
                    id: "exportEPUB",
                    label: "EPUB",
                    iconHTML: "",
                    click: () => {
                        const msgId = showMessage(window.scribli.languages.exporting, -1);
                        fetchPost("/api/export/exportEPUB", {
                            id,
                        }, response => {
                            saveExportFile(response.data.zip, msgId);
                        });
                    }
                }]
            },
            /// #else
            {
                id: "exportPDF",
                label: window.scribli.languages.print,
                icon: "iconPDF",
                ignore: true,
                click: () => {
                    const msgId = showMessage(window.scribli.languages.exporting);
                    const localData = window.scribli.storage[Constants.LOCAL_EXPORTPDF];
                    fetchPost("/api/export/exportPreviewHTML", {
                        id,
                        keepFold: localData.keepFold,
                        merge: localData.mergeSubdocs,
                    }, async response => {
                        setTimeout(() => {
                            hideMessage(msgId);
                        }, 3000);
                    });
                }
            }, {
                id: "exportHTML_Scribli",
                label: "HTML (Scribli)",
                iconClass: "ft__error",
                icon: "iconHTML5",
                click: () => {
                    saveExport({type: "html", id});
                }
            }, {
                id: "exportHTML_Markdown",
                label: "HTML (Markdown)",
                icon: "iconHTML5",
                click: () => {
                    saveExport({type: "htmlmd", id});
                }
            },
            /// #endif
        ]
    }).element;
};

export const openMenu = (app: App, src: string, onlyMenu: boolean, showAccelerator: boolean) => {
    const submenu = [];
    if (isLocalPath(src)) {
        if (Constants.SCRIBLI_ASSETS_EXTS.includes(pathPosix().extname(src).split("?")[0]) &&
            (!src.endsWith(".pdf") ||
                (src.endsWith(".pdf") && !src.startsWith("file://")))
        ) {
            submenu.push({
                id: "insertRight",
                icon: "iconLayoutRight",
                label: window.scribli.languages.insertRight,
                accelerator: showAccelerator ? window.scribli.languages.click : "",
                click() {
                    openAsset(app, src.trim(), parseInt(getSearch("page", src)), "right");
                }
            });
            submenu.push({
                id: "openBy",
                label: window.scribli.languages.openBy,
                icon: "iconOpen",
                accelerator: showAccelerator ? "⌥" + window.scribli.languages.click : "",
                click() {
                    openAsset(app, src.trim(), parseInt(getSearch("page", src)));
                }
            });
            /// #if !BROWSER
            submenu.push({
                id: "openByNewWindow",
                label: window.scribli.languages.openByNewWindow,
                icon: "iconOpenWindow",
                click() {
                    openAssetNewWindow(src.trim());
                }
            });
            submenu.push({
                id: "showInFolder",
                icon: "iconFolder",
                label: window.scribli.languages.showInFolder,
                accelerator: showAccelerator ? "⌘" + window.scribli.languages.click : "",
                click: () => {
                    openBy(src, "folder");
                }
            });
            submenu.push({
                id: "useDefault",
                label: window.scribli.languages.useDefault,
                accelerator: showAccelerator ? "⇧" + window.scribli.languages.click : "",
                click() {
                    openBy(src, "app");
                }
            });
            /// #endif
        } else {
            /// #if !BROWSER
            submenu.push({
                id: "useDefault",
                label: window.scribli.languages.useDefault,
                accelerator: showAccelerator ? window.scribli.languages.click : "",
                click() {
                    openBy(src, "app");
                }
            });
            submenu.push({
                id: "showInFolder",
                icon: "iconFolder",
                label: window.scribli.languages.showInFolder,
                accelerator: showAccelerator ? "⌘" + window.scribli.languages.click : "",
                click: () => {
                    openBy(src, "folder");
                }
            });
            /// #else
            submenu.push({
                id: "useBrowserView",
                label: window.scribli.languages.useBrowserView,
                accelerator: showAccelerator ? window.scribli.languages.click : "",
                click: () => {
                    openByMobile(src);
                }
            });
            /// #endif
        }
    } else if (src) {
        if (0 > src.indexOf(":")) {
            // Support click to open hyperlinks like `www.foo.com` 
            src = `https://${src}`;
        }
        /// #if !BROWSER
        submenu.push({
            id: "useDefault",
            label: window.scribli.languages.useDefault,
            accelerator: showAccelerator ? window.scribli.languages.click : "",
            click: () => {
                if (processScribliUri(app, src)) {
                    return;
                }
                shell.openExternal(src).catch((e) => {
                    showMessage(e);
                });
            }
        });
        /// #else
        submenu.push({
            id: "useBrowserView",
            label: window.scribli.languages.useBrowserView,
            accelerator: showAccelerator ? window.scribli.languages.click : "",
            click: () => {
                openByMobile(src);
            }
        });
        /// #endif
    }
    if (onlyMenu) {
        return submenu;
    }
    window.scribli.menus.menu.append(new MenuItem({
        id: "openBy",
        label: window.scribli.languages.openBy,
        icon: "iconOpen",
        submenu
    }).element);
};

export const renameMenu = (options: {
    path: string
    notebookId: string
    name: string,
    type: "notebook" | "file"
    docId?: string | null
}) => {
    return new MenuItem({
        id: "rename",
        accelerator: window.scribli.config.keymap.editor.general.rename.custom,
        icon: "iconEdit",
        label: window.scribli.languages.rename,
        click: () => {
            if (options.type === "file" && options.docId) {
                const docInfoParam: IObject = {
                    id: options.docId
                };
                if (isEncryptedBox(options.notebookId)) {
                    docInfoParam.notebook = options.notebookId;
                }
                fetchPost("/api/block/getDocInfo", docInfoParam, (response) => {
                    rename({
                        ...options,
                        name: response.data.ial.title,
                        empty: response.data.ial[Constants.CUSTOM_SY_TITLE_EMPTY] === "true",
                    });
                });
            } else {
                rename(options);
            }
        }
    }).element;
};

export const movePathToMenu = (paths: string[]) => {
    return new MenuItem({
        id: "move",
        label: window.scribli.languages.move,
        icon: "iconMove",
        accelerator: window.scribli.config.keymap.general.move.custom,
        click() {
            const rootIDs: string[] = [];
            paths.forEach(item => {
                rootIDs.push(pathPosix().basename(item).replace(".sy", ""));
            });
            movePathTo({
                cb: (toPath, toNotebook) => {
                    moveToPath(paths, toNotebook[0], toPath[0]);
                },
                paths,
                flashcard: false,
                rootIDs,
            });
        }
    }).element;
};

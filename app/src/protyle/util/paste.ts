import {Constants} from "../../constants";
import {uploadFiles, uploadLocalFiles} from "../upload";
import {processPasteCode, processRender} from "./processCode";
import {getLocalFiles, getTextScribliFromTextHTML, readText} from "./compatibility";
import {hasClosestBlock, hasClosestByAttribute, hasClosestByClassName} from "./hasClosest";
import {getEditorRange, getSelectionOffset} from "./selection";
import {blockRender} from "../render/blockRender";
import {highlightRender} from "../render/highlightRender";
import {fetchPost, fetchSyncPost} from "../../util/fetch";
import {isDynamicRef, isFileAnnotation} from "../../util/functions";
import {insertHTML} from "./insertHTML";
import {scrollCenter} from "../../util/highlightById";
import {hideElements} from "../ui/hideElements";
import {avRender} from "../render/av/render";
import {cellScrollIntoView, getCellText} from "../render/av/cell";
import {getCalloutInfo, getContenteditableElement} from "../wysiwyg/getBlock";
import {clearBlockElement} from "./clear";
import {removeZWJ} from "./normalizeText";
import {base64ToURL} from "../../util/image";
import {resolveLinkDest} from "../toolbar/util";

export const beforePaste = (protyle: IProtyle, blockElement: HTMLElement) => {
    const range = getSelection().getRangeAt(0);
    protyle.toolbar.range = range;
    const inlineElement = range.startContainer.parentElement;
    if (range.toString() === "" && inlineElement.tagName === "SPAN") {
        const currentTypes = (inlineElement.getAttribute("data-type") || "").split(" ");
        if (currentTypes.includes("inline-memo") || currentTypes.includes("text") ||
            currentTypes.includes("block-ref") || currentTypes.includes("file-annotation-ref") ||
            currentTypes.includes("a")) {
            const offset = getSelectionOffset(inlineElement, blockElement, range);
            if (offset.start === 0) {
                range.setStartBefore(inlineElement);
                range.collapse(true);
            } else if (offset.start === inlineElement.textContent.length) {
                range.setEndAfter(inlineElement);
                range.collapse(false);
            }
        }
    }
};

export const getTextStar = (blockElement: HTMLElement, contentOnly = false) => {
    const dataType = blockElement.dataset.type;
    let refText = "";
    if (["NodeHeading", "NodeParagraph"].includes(dataType)) {
        refText = getContenteditableElement(blockElement).innerHTML;
    } else if ("NodeHTMLBlock" === dataType) {
        refText = "HTML";
    } else if ("NodeAttributeView" === dataType) {
        refText = blockElement.querySelector(".av__title").textContent || window.scribli.languages.database;
    } else if ("NodeThematicBreak" === dataType) {
        refText = window.scribli.languages.line;
    } else if ("NodeIFrame" === dataType) {
        refText = "IFrame";
    } else if ("NodeWidget" === dataType) {
        refText = window.scribli.languages.widget;
    } else if ("NodeVideo" === dataType) {
        refText = window.scribli.languages.video;
    } else if ("NodeAudio" === dataType) {
        refText = window.scribli.languages.audio;
    } else if (["NodeCodeBlock", "NodeTable"].includes(dataType)) {
        refText = getPlainText(blockElement);
    } else if (blockElement.classList.contains("render-node")) {
        refText += blockElement.dataset.subtype || Lute.UnEscapeHTMLStr(blockElement.getAttribute("data-content"));
    } else if (["NodeBlockquote", "NodeList", "NodeSuperBlock", "NodeListItem"].includes(dataType)) {
        Array.from(blockElement.querySelectorAll<HTMLElement>("[data-node-id]")).find((item) => {
            if (!["NodeBlockquote", "NodeList", "NodeSuperBlock", "NodeListItem"].includes(item.getAttribute("data-type"))) {
                refText = getTextStar(item, true);
                return true;
            }
        });
    } else if ("NodeCallout" === dataType) {
        refText = getCalloutInfo(blockElement);
    }
    if (contentOnly) {
        return refText;
    }
    return refText + ` <span data-type="block-ref" data-subtype="s" data-id="${blockElement.getAttribute("data-node-id")}">*</span>`;
};

export const getPlainText = (blockElement: HTMLElement, isNested = false) => {
    let text = "";
    const dataType = blockElement.dataset.type;
    if ("NodeHTMLBlock" === dataType) {
        text += Lute.UnEscapeHTMLStr(blockElement.querySelector("protyle-html").getAttribute("data-content"));
    } else if ("NodeAttributeView" === dataType) {
        blockElement.querySelectorAll(".av__row").forEach(rowElement => {
            rowElement.querySelectorAll<HTMLElement>(".av__cell").forEach((cellElement) => {
                text += getCellText(cellElement) + " ";
            });
            text += "\n";
        });
        text = text.trimEnd();
    } else if ("NodeThematicBreak" === dataType) {
        text += "---";
    } else if ("NodeIFrame" === dataType || "NodeWidget" === dataType) {
        text += blockElement.querySelector("iframe").getAttribute("src");
    } else if ("NodeVideo" === dataType) {
        text += blockElement.querySelector("video").getAttribute("src");
    } else if ("NodeAudio" === dataType) {
        text += blockElement.querySelector("audio").getAttribute("src");
    } else if (blockElement.classList.contains("render-node")) {
        text += Lute.UnEscapeHTMLStr(blockElement.getAttribute("data-content"));
    } else if (["NodeHeading", "NodeParagraph"].includes(dataType)) {
        text += blockElement.querySelector("[spellcheck]").textContent;
    } else if ("NodeCodeBlock" === dataType) {
        text += removeZWJ(blockElement.querySelector("[spellcheck]").textContent);
    } else if (dataType === "NodeTable") {
        blockElement.querySelectorAll("th, td").forEach((item) => {
            text += item.textContent.trim() + "\t";
            if (!item.nextElementSibling) {
                text = text.slice(0, -1) + "\n";
            }
        });
        text = text.slice(0, -1);
    } else if (!isNested && ["NodeBlockquote", "NodeCallout", "NodeList", "NodeSuperBlock", "NodeListItem"].includes(dataType)) {
        if (dataType === "NodeCallout") {
            text += `${getCalloutInfo(blockElement)}\n`;
        }
        blockElement.querySelectorAll<HTMLElement>("[data-node-id]").forEach((item) => {
            const nestedText = getPlainText(item, true);
            text += nestedText ? nestedText + "\n" : "";
        });
    }
    return text;
};

export const pasteEscaped = async (protyle: IProtyle, nodeElement: Element) => {
    try {
        let clipText = await readText() || "";
        clipText = clipText.replace(/<span data-type="text".*?>(.*?)<\/span>/g, "$1");

        // 
        // A\B\C\D\
        // E

        clipText = clipText.replace(/\\/g, "\\\\")
            .replace(/\*/g, "\\*")
            .replace(/_/g, "\\_")
            .replace(/\[/g, "\\[")
            .replace(/]/g, "\\]")
            .replace(/!/g, "\\!")
            .replace(/`/g, "\\`")
            .replace(/</g, "\\<")
            .replace(/>/g, "\\>")
            .replace(/&/g, "\\&")
            .replace(/~/g, "\\~")
            .replace(/\{/g, "\\{")
            .replace(/}/g, "\\}")
            .replace(/\(/g, "\\(")
            .replace(/\)/g, "\\)")
            .replace(/=/g, "\\=")
            .replace(/#/g, "\\#")
            .replace(/\$/g, "\\$")
            .replace(/\^/g, "\\^")
            .replace(/\|/g, "\\|")
            .replace(/\./g, "\\.");
        paste(protyle, {textPlain: clipText, textHTML: "", target: nodeElement as HTMLElement});
    } catch (e) {
        console.log(e);
    }
};

export const pasteAsPlainText = async (protyle: IProtyle) => {
    let localFiles: ILocalFiles[] = [];
    /// #if !BROWSER
    localFiles = await getLocalFiles();
    if (localFiles.length > 0) {
        uploadLocalFiles(localFiles, protyle, false);
        return;
    }
    /// #endif
    if (localFiles.length === 0) {
        // Inline-level elements support pasted as plain text 
        let textPlain = await readText() || "";
        if (getSelection().rangeCount > 0) {
            const range = getSelection().getRangeAt(0);
            if (hasClosestByAttribute(range.startContainer, "data-type", "code") || hasClosestByClassName(range.startContainer, "hljs")) {
                insertHTML(removeZWJ(textPlain).replace(/```/g, "\u200D```"), protyle);
                return;
            }
        }
        textPlain = textPlain.replace(/<sub>/g, "__@sub@__").replace(/<\/sub>/g, "__@/sub@__");
        textPlain = textPlain.replace(/<sup>/g, "__@sup@__").replace(/<\/sup>/g, "__@/sup@__");
        textPlain = textPlain.replace(/<kbd>/g, "__@kbd@__").replace(/<\/kbd>/g, "__@/kbd@__");
        textPlain = textPlain.replace(/<u>/g, "__@u@__").replace(/<\/u>/g, "__@/u@__");

        textPlain = textPlain.replace(/<span data-type="text".*?>(.*?)<\/span>/g, "$1");

        textPlain = textPlain.replace(/<<assets\//g, "__@lt2assets/@__").replace(/>>/g, "__@gt2@__");

        textPlain = textPlain.replace(/</g, ";;;lt;;;").replace(/>/g, ";;;gt;;;");

        textPlain = textPlain.replace(/__@lt2assets\/@__/g, "<<assets/").replace(/__@gt2@__/g, ">>");

        textPlain = textPlain.replace(/__@sub@__/g, "<sub>").replace(/__@\/sub@__/g, "</sub>");
        textPlain = textPlain.replace(/__@sup@__/g, "<sup>").replace(/__@\/sup@__/g, "</sup>");
        textPlain = textPlain.replace(/__@kbd@__/g, "<kbd>").replace(/__@\/kbd@__/g, "</kbd>");
        textPlain = textPlain.replace(/__@u@__/g, "<u>").replace(/__@\/u@__/g, "</u>");

        enableLuteMarkdownSyntax(protyle);
        const content = protyle.lute.BlockDOM2EscapeMarkerContent(protyle.lute.Md2BlockDOM(textPlain));
        restoreLuteMarkdownSyntax(protyle);

        insertHTML(content, protyle, false, false, true);
    }
};

export const enableLuteMarkdownSyntax = (protyle: IProtyle) => {
    protyle.lute.SetInlineAsterisk(true);
    protyle.lute.SetGFMStrikethrough(true);
    protyle.lute.SetInlineMath(true);
    protyle.lute.SetSub(true);
    protyle.lute.SetSup(true);
    protyle.lute.SetTag(true);
    protyle.lute.SetInlineUnderscore(true);
};

export const restoreLuteMarkdownSyntax = (protyle: IProtyle) => {
    protyle.lute.SetInlineAsterisk(window.scribli.config.editor.markdown.inlineAsterisk);
    protyle.lute.SetGFMStrikethrough(window.scribli.config.editor.markdown.inlineStrikethrough);
    protyle.lute.SetInlineMath(window.scribli.config.editor.markdown.inlineMath);
    protyle.lute.SetSub(window.scribli.config.editor.markdown.inlineSub);
    protyle.lute.SetSup(window.scribli.config.editor.markdown.inlineSup);
    protyle.lute.SetTag(window.scribli.config.editor.markdown.inlineTag);
    protyle.lute.SetInlineUnderscore(window.scribli.config.editor.markdown.inlineUnderscore);
    protyle.lute.SetMark(window.scribli.config.editor.markdown.inlineMark);
};

const readLocalFile = async (protyle: IProtyle, localFiles: ILocalFiles[]) => {
    if (protyle && protyle.app && protyle.app.plugins) {
        for (let i = 0; i < protyle.app.plugins.length; i++) {
            const response: { localFiles: ILocalFiles[] } = await new Promise((resolve) => {
                const emitResult = protyle.app.plugins[i].eventBus.emit("paste", {
                    protyle,
                    resolve,
                    textHTML: "",
                    textPlain: "",
                    scribliHTML: "",
                    localFiles
                });
                if (emitResult) {
                    resolve(undefined);
                }
            });
            if (response?.localFiles) {
                localFiles = response.localFiles;
            }
        }
    }
    uploadLocalFiles(localFiles, protyle, true);
};

export const paste = async (protyle: IProtyle, event: (ClipboardEvent | DragEvent | IClipboardData) & {
    target: HTMLElement
}) => {
    if ("clipboardData" in event || "dataTransfer" in event) {
        event.stopPropagation();
        event.preventDefault();
    }
    let textHTML: string;
    let textPlain: string;
    let scribliHTML: string;
    let files: FileList | DataTransferItemList | File[];
    if ("clipboardData" in event) {
        textHTML = event.clipboardData.getData("text/html");
        textPlain = event.clipboardData.getData("text/plain");
        scribliHTML = event.clipboardData.getData("text/scribli") || event.clipboardData.getData("text/Scribli");
        files = event.clipboardData.files;
    } else if ("dataTransfer" in event) {
        textHTML = event.dataTransfer.getData("text/html");
        textPlain = event.dataTransfer.getData("text/plain");
        scribliHTML = event.dataTransfer.getData("text/scribli") || event.dataTransfer.getData("text/Scribli");
        if (event.dataTransfer.types[0] === "Files") {
            files = event.dataTransfer.items;
        }
    } else {
        if (event.localFiles?.length > 0) {
            readLocalFile(protyle, event.localFiles);
            return;
        }
        textHTML = event.textHTML;
        textPlain = event.textPlain;
        scribliHTML = event.scribliHTML;
        files = event.files;
    }

    // Improve the pasting of selected text in PDF rectangular annotation 
    textPlain = textPlain.replace(/\r\n|\r|\u2028|\u2029/g, "\n");

    /// #if !BROWSER
    if (!scribliHTML && !textHTML && !textPlain && ("clipboardData" in event)) {
        const localFiles: ILocalFiles[] = await getLocalFiles();
        if (localFiles.length > 0) {
            readLocalFile(protyle, localFiles);
            return;
        }
    }
    /// #endif
    const originalTextHTML = textHTML;
    if (textHTML.replace(/&amp;/g, "&").replace(/<(|\/)(html|body|meta)[^>]*?>/ig, "").trim() ===
        `<a href="${textPlain}">${textPlain}</a>` ||
        textHTML.replace(/&amp;/g, "&").replace(/<(|\/)(html|body|meta)[^>]*?>/ig, "").trim() ===
        `<!--StartFragment--><a href="${textPlain}">${textPlain}</a><!--EndFragment-->`) {
        textHTML = "";
    }
    if (textPlain.endsWith(Constants.ZWSP) && !textHTML && !scribliHTML) {
        scribliHTML = textPlain.substr(0, textPlain.length - 1);
    }
    if (textHTML && textPlain && !scribliHTML) {
        const textObj = getTextScribliFromTextHTML(textHTML);
        scribliHTML = textObj.textScribli;
        textHTML = textObj.textHtml;
    }
    if (!scribliHTML) {
        // process word
        const doc = new DOMParser().parseFromString(textHTML, "text/html");
        if (doc.body && doc.body.innerHTML) {
            textHTML = doc.body.innerHTML;
        }
        if (textHTML.startsWith("\n<!--StartFragment-->") && textHTML.endsWith("<!--EndFragment-->\n\n")) {
            textHTML = doc.body.innerHTML.trim().replace("<!--StartFragment-->", "").replace("<!--EndFragment-->", "");
        }
        textHTML = Lute.Sanitize(textHTML);
    }

    if (protyle && protyle.app && protyle.app.plugins) {
        for (let i = 0; i < protyle.app.plugins.length; i++) {
            const response: IClipboardData & { files: FileList } = await new Promise((resolve) => {
                const emitResult = protyle.app.plugins[i].eventBus.emit("paste", {
                    protyle,
                    resolve,
                    textHTML,
                    textPlain,
                    scribliHTML,
                    files
                });
                if (emitResult) {
                    resolve(undefined);
                }
            });

            if (response?.textHTML) {
                textHTML = response.textHTML;
            }
            if (response?.textPlain) {
                textPlain = response.textPlain;
            }
            if (response?.scribliHTML) {
                scribliHTML = response.scribliHTML;
            }
            if (response?.files) {
                files = response.files as FileList;
            }
        }
    }


    let nodeElement = hasClosestBlock(event.target);
    const range = getEditorRange(protyle.wysiwyg.element);
    if (!nodeElement) {
        nodeElement = hasClosestBlock(range.startContainer);
    }
    if (!nodeElement) {
        if (files && files.length > 0) {
            uploadFiles(protyle, files);
        }
        return;
    }
    protyle.hint.enableExtend = protyle.hint.enableExtend ? Constants.BLOCK_HINT_KEYS.concat("{{", "/", "#", "、", "「「", "「『", "『「", "『『",).includes(protyle.hint.splitChar) : false;
    hideElements(protyle.hint.enableExtend ? ["select"] : ["select", "hint"], protyle);
    protyle.wysiwyg.element.querySelectorAll(".protyle-wysiwyg--hl").forEach(item => {
        item.classList.remove("protyle-wysiwyg--hl");
    });
    const code = processPasteCode(textHTML, textPlain, originalTextHTML, protyle);
    if (nodeElement.getAttribute("data-type") === "NodeCodeBlock" ||
        protyle.toolbar.getCurrentType(range).includes("code")) {
        // 
        insertHTML(removeZWJ(textPlain).replace(/```/g, "\u200D```"), protyle);
        return;
    } else if (scribliHTML) {
        async function streamInsert(container: HTMLElement, bigHtmlString: string) {
            const iframe = document.createElement("iframe");
            iframe.style.display = "none";
            document.body.appendChild(iframe);

            const doc = iframe.contentWindow.document;
            doc.open();

            const chunkSize = 102400;
            let offset = 0;
            while (offset < bigHtmlString.length) {
                const chunk = bigHtmlString.substring(offset, offset + chunkSize);
                doc.write(chunk);
                offset += chunkSize;
                await new Promise(resolve => setTimeout(resolve, 0));
            }

            doc.close();

            const fragment = document.createDocumentFragment();
            while (doc.body.firstChild) {
                fragment.appendChild(doc.body.firstChild);
            }

            container.appendChild(fragment);

            document.body.removeChild(iframe);
        }

        const tempElement = document.createElement("div");
        if (1024 * 512 < scribliHTML.length) {
            await streamInsert(tempElement, scribliHTML);
        } else {
            tempElement.innerHTML = scribliHTML;
        }
        if (range.toString()) {
            let types: string[] = [];
            let linkElement: HTMLElement;
            if (tempElement.childNodes.length === 1 && tempElement.childElementCount === 1) {
                types = (tempElement.firstElementChild.getAttribute("data-type") || "").split(" ");
                if ((types.includes("block-ref") || types.includes("a"))) {
                    linkElement = tempElement.firstElementChild as HTMLElement;
                }
            }
            if (!linkElement) {
                const linkTemp = document.createElement("template");
                linkTemp.innerHTML = protyle.lute.SpinBlockDOM(scribliHTML);
                if (linkTemp.content.firstChild.nodeType !== 3 && linkTemp.content.firstElementChild.classList.contains("p")) {
                    linkTemp.innerHTML = linkTemp.content.firstElementChild.firstElementChild.innerHTML.trim();
                }
                if (linkTemp.content.childNodes.length === 1 && linkTemp.content.childElementCount === 1) {
                    types = (linkTemp.content.firstElementChild.getAttribute("data-type") || "").split(" ");
                    if ((types.includes("block-ref") || types.includes("a"))) {
                        linkElement = linkTemp.content.firstElementChild as HTMLElement;
                    }
                }
            }

            if (types.includes("block-ref")) {
                const refElement = protyle.toolbar.setInlineMark(protyle, "block-ref", "range", {
                    type: "id",
                    color: `${linkElement.dataset.id}${Constants.ZWSP}s${Constants.ZWSP}${range.toString()}`
                });
                if (refElement[0]) {
                    protyle.toolbar.range.selectNodeContents(refElement[0]);
                }
                return;
            }
            if (types.includes("a")) {
                protyle.toolbar.setInlineMark(protyle, "a", "range", {
                    type: "a",
                    color: `${linkElement.dataset.href}${Constants.ZWSP}${range.toString()}`
                });
                return;
            }
        }
        let isBlock = false;
        const pastedBlockElements = tempElement.querySelectorAll("[data-node-id]");
        if (pastedBlockElements.length > 0) {
            isBlock = true;
            const oldIds: string[] = [];
            pastedBlockElements.forEach((e) => {
                oldIds.push(e.getAttribute("data-node-id"));
            });
            const existResponse = await fetchSyncPost("/api/block/checkBlocksExist", {ids: oldIds});
            pastedBlockElements.forEach((e) => {
                const originalId = e.getAttribute("data-node-id");
                const isCutPaste = existResponse.data[originalId] === false;
                if (!isCutPaste) {
                    e.setAttribute("data-node-id", Lute.NewNodeID());
                }
                clearBlockElement(e, isCutPaste);
            });
        }
        if (nodeElement.classList.contains("table")) {
            isBlock = false;
        }
        tempElement.querySelectorAll('[contenteditable="false"][spellcheck]').forEach((e) => {
            e.setAttribute("contenteditable", "true");
        });

        let tempInnerHTML = tempElement.innerHTML;

        if (!nodeElement.classList.contains("av") && tempInnerHTML.startsWith("[[{") && tempInnerHTML.endsWith("}]]")) {
            try {
                const json = JSON.parse(tempInnerHTML);
                if (json.length > 0 && json[0].length > 0 && json[0][0].id && json[0][0].type) {
                    insertHTML(textPlain, protyle, isBlock);
                } else {
                    insertHTML(tempInnerHTML, protyle, isBlock);
                }
            } catch (e) {
                insertHTML(tempInnerHTML, protyle, isBlock);
            }
        } else {
            if (-1 < tempInnerHTML.indexOf("NodeHTMLBlock")) {
                tempInnerHTML = Lute.UnEscapeHTMLStr(tempInnerHTML);
            }

            insertHTML(tempInnerHTML, protyle, isBlock, false, true);
        }
        blockRender(protyle, protyle.wysiwyg.element);
        processRender(protyle.wysiwyg.element);
        highlightRender(protyle.wysiwyg.element);
        avRender(protyle.wysiwyg.element, protyle);
    } else if (code) {
        if (!code.startsWith('<div data-type="NodeCodeBlock" class="code-block" data-node-id="')) {
            insertHTML(code, protyle);
        } else {
            insertHTML(code, protyle, true, false, true);
            highlightRender(protyle.wysiwyg.element);
        }
        hideElements(["hint"], protyle);
    } else {
        let isHTML = false;
        if (textHTML.replace("<!--StartFragment--><!--EndFragment-->", "").trim() !== "") {
            textHTML = textHTML.replace("<!--StartFragment-->", "").replace("<!--EndFragment-->", "").trim();
            if (files && files.length === 1 && (
                textHTML.startsWith("<img") ||
                (textHTML.startsWith("<table") && textHTML.indexOf("<img") > -1)
            )) {
                isHTML = false;
            } else {
                isHTML = true;
            }

            let containsNewlines = false;
            const tempDiv = document.createElement("div");
            tempDiv.innerHTML = textHTML;
            const walker = document.createTreeWalker(tempDiv, NodeFilter.SHOW_TEXT, null);
            let node: Node | null = null;
            while ((node = walker.nextNode())) {
                if (node.nodeValue && (node.nodeValue.match(/\n/g) || []).length >= 2) {
                    containsNewlines = true;
                    break;
                }
            }

            const textHTMLLowercase = textHTML.toLowerCase();
            if (textPlain && "" !== textPlain.trim() && (textHTML.startsWith("<span") || textHTML.startsWith("<br")) && containsNewlines &&
                (0 > textHTMLLowercase.indexOf("class=\"katex") && 0 > textHTMLLowercase.indexOf("class=\"math") &&
                    0 > textHTMLLowercase.indexOf("</a>") && 0 > textHTMLLowercase.indexOf("</img>") && 0 > textHTMLLowercase.indexOf("</code>") &&
                    0 > textHTMLLowercase.indexOf("</b>") && 0 > textHTMLLowercase.indexOf("</strong>") &&
                    0 > textHTMLLowercase.indexOf("</i>") && 0 > textHTMLLowercase.indexOf("</em>") &&
                    0 > textHTMLLowercase.indexOf("</ol>") && 0 > textHTMLLowercase.indexOf("</ul>") &&
                    0 > textHTMLLowercase.indexOf("</table>") && 0 > textHTMLLowercase.indexOf("</blockquote>") &&
                    0 > textHTMLLowercase.indexOf("</h1>") && 0 > textHTMLLowercase.indexOf("</h2>") &&
                    0 > textHTMLLowercase.indexOf("</h3>") && 0 > textHTMLLowercase.indexOf("</h4>") &&
                    0 > textHTMLLowercase.indexOf("</h5>") && 0 > textHTMLLowercase.indexOf("</h6>"))) {
                isHTML = false;
            }
        } else if (textPlain && textPlain.trimStart().startsWith("<")) {
            // Improve pasting for tables containing merged cells 
            if (textPlain.toLowerCase().indexOf("</table>") > -1) {
                textHTML = textPlain;
                isHTML = true;
            }
        }
        if (isHTML) {
            const tempElement = document.createElement("div");
            tempElement.innerHTML = textHTML;
            tempElement.querySelectorAll("a").forEach((e) => {
                if (e.innerHTML.trim() === "") {
                    e.remove();
                }
            });
            // 
            let linkElement;
            if (tempElement.childElementCount === 1 && tempElement.childNodes.length === 1) {
                if (tempElement.firstElementChild.tagName === "A") {
                    linkElement = tempElement.firstElementChild;
                } else if (tempElement.firstElementChild.tagName === "P" &&
                    tempElement.firstElementChild.childElementCount === 1 &&
                    tempElement.firstElementChild.childNodes.length === 1 &&
                    tempElement.firstElementChild.firstElementChild.tagName === "A") {
                    linkElement = tempElement.firstElementChild.firstElementChild;
                }
            }
            if (linkElement) {
                const selectText = range.toString();
                const aElements = protyle.toolbar.setInlineMark(protyle, "a", "range", {
                    type: "a",
                    color: `${linkElement.getAttribute("href")}${Constants.ZWSP}${selectText || linkElement.textContent}`
                });
                if (!selectText) {
                    if (aElements[0].lastChild) {
                        // 
                        range.setEnd(aElements[0].lastChild, aElements[0].lastChild.textContent.length);
                    }
                    range.collapse(false);
                }
                return;
            }
            fetchPost("/api/lute/html2BlockDOM", {
                dom: tempElement.innerHTML
            }, (response) => {
                insertHTML(response.data, protyle, false, false, true);
                protyle.wysiwyg.element.querySelectorAll('[data-type~="block-ref"]').forEach(item => {
                    if (item.textContent === "") {
                        fetchPost("/api/block/getRefText", {id: item.getAttribute("data-id")}, (response) => {
                            item.innerHTML = response.data;
                        });
                    }
                });
                blockRender(protyle, protyle.wysiwyg.element);
                processRender(protyle.wysiwyg.element);
                highlightRender(protyle.wysiwyg.element);
                avRender(protyle.wysiwyg.element, protyle);
                scrollCenter(protyle, undefined, "nearest", "smooth");
            });
            return;
        } else if (files && files.length > 0) {
            uploadFiles(protyle, files);
            return;
        } else if (textPlain.trim() !== "" && (files && files.length === 0 || !files)) {
            if (range.toString() !== "") {
                const firstLine = textPlain.split("\n")[0];
                if (isDynamicRef(textPlain)) {
                    const refElement = protyle.toolbar.setInlineMark(protyle, "block-ref", "range", {
                        type: "id",
                        color: `${textPlain.substring(2, 22 + 2)}${Constants.ZWSP}s${Constants.ZWSP}${range.toString()}`
                    });
                    if (refElement[0]) {
                        protyle.toolbar.range.selectNodeContents(refElement[0]);
                    }
                    return;
                } else if (isFileAnnotation(firstLine)) {
                    protyle.toolbar.setInlineMark(protyle, "file-annotation-ref", "range", {
                        type: "file-annotation-ref",
                        color: firstLine.substring(2).replace(/ ".+">>$/, "")
                    });
                    return;
                } else {
                    // 
                    const linkDest = resolveLinkDest(textPlain, protyle.lute);
                    if (linkDest) {
                        protyle.toolbar.setInlineMark(protyle, "a", "range", {
                            type: "a",
                            color: linkDest
                        });
                        return;
                    }
                }
            }
            let textPlainDom: string;

            // Auto-convert pasted URL to link format 
            if (window.scribli.config.editor.pasteURLAutoConvert) {
                textPlainDom = protyle.lute.Md2BlockDOMWithAutoLink(textPlain);
            } else {
                textPlainDom = protyle.lute.Md2BlockDOM(textPlain);
            }
            if (textPlainDom && textPlainDom.indexOf("data:image/") > -1) {
                const tempElement = document.createElement("template");
                tempElement.innerHTML = textPlainDom;
                const imgSrcList: string[] = [];
                const imageElements = tempElement.content.querySelectorAll("img");
                imageElements.forEach((item) => {
                    if (item.getAttribute("data-src").startsWith("data:image/")) {
                        imgSrcList.push(item.getAttribute("data-src"));
                    }
                });
                const base64SrcList = await base64ToURL(imgSrcList);
                base64SrcList.forEach((item, index) => {
                    imageElements[index].setAttribute("src", item);
                    imageElements[index].setAttribute("data-src", item);
                    imageElements[index].parentElement.querySelector(".img__net")?.remove();
                });
                textPlainDom = tempElement.innerHTML;
            }
            insertHTML(textPlainDom, protyle, false, false, true);
        }
        blockRender(protyle, protyle.wysiwyg.element);
        processRender(protyle.wysiwyg.element);
        highlightRender(protyle.wysiwyg.element);
        avRender(protyle.wysiwyg.element, protyle);
    }
    const selectCellElement = nodeElement.querySelector(".av__cell--select");
    if (nodeElement.classList.contains("av") && selectCellElement) {
        cellScrollIntoView(nodeElement, selectCellElement);
    } else {
        scrollCenter(protyle, undefined, "nearest", "smooth");
    }
};

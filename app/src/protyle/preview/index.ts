import {isOnlyMeta} from "../util/compatibility";
import {openByMobile} from "../../editor/openLink";
import {showMessage} from "../../dialog/message";
import {isLocalPath, pathPosix} from "../../util/pathName";
import {processScribliUri} from "../../util/uri";
import {previewDocImage} from "./image";
import {getDiagramBlock, previewDiagram} from "./diagram";
import {Constants} from "../../constants";
import {getSearch, isMobile} from "../../util/functions";
/// #if !BROWSER
import {shell} from "electron";
/// #endif
import {openAsset, openBy} from "../../editor/util";
import {getAllModels} from "../../layout/getAll";
import {fetchPost} from "../../util/fetch";
import {processRender} from "../util/processCode";
import {highlightRender} from "../render/highlightRender";
import {speechRender} from "../render/speechRender";
import {avRender} from "../render/av/render";
import {getPadding} from "../ui/initUI";
import {hasTopClosestByAttribute} from "../util/hasClosest";

export class Preview {
    public element: HTMLElement;
    public previewElement: HTMLElement;
    private mdTimeoutId: number;

    constructor(protyle: IProtyle) {
        this.element = document.createElement("div");
        this.element.className = "protyle-preview fn__none";

        const previewElement = document.createElement("div");
        previewElement.className = "b3-typography";
        if (protyle.options.classes.preview) {
            previewElement.classList.add(protyle.options.classes.preview);
        }
        const actions = protyle.options.preview.actions;
        const actionElement = document.createElement("div");
        actionElement.className = "protyle-preview__action";
        const actionHtml: string[] = [];
        for (let i = 0; i < actions.length; i++) {
            const action = actions[i];
            if (typeof action === "object") {
                actionHtml.push(`<button type="button" data-type="${action.key}" class="${action.className}">${action.text}</button>`);
                continue;
            }
            switch (action) {
                case "desktop":
                    actionHtml.push(`<button type="button" class="protyle-preview__action--current" data-type="desktop">${window.scribli.languages.desktop}</button>`);
                    break;
                case "tablet":
                    actionHtml.push(`<button type="button" data-type="tablet">${window.scribli.languages.tablet}</button>`);
                    break;
                case "mobile":
                    actionHtml.push(`<button type="button" data-type="mobile">${window.scribli.languages.mobile}</button>`);
                    break;
            }
        }
        actionElement.innerHTML = actionHtml.join("");
        this.element.appendChild(actionElement);
        this.element.appendChild(previewElement);

        this.element.addEventListener("click", (event) => {
            let target = event.target as HTMLElement;
            while (target && !target.isEqualNode(this.element)) {
                if (target.tagName === "A") {
                    const linkAddress = target.getAttribute("href");
                    if (linkAddress.startsWith("#")) {
                        const hash = linkAddress.substring(1);
                        previewElement.querySelector('[data-node-id="' + hash + '"], [id="' + hash + '"]').scrollIntoView();
                        event.stopPropagation();
                        event.preventDefault();
                        break;
                    }

                    if (isMobile()) {
                        openByMobile(linkAddress);
                        event.stopPropagation();
                        event.preventDefault();
                        break;
                    }
                    event.stopPropagation();
                    event.preventDefault();
                    if (isLocalPath(linkAddress)) {
                        if (isOnlyMeta(event)) {
                            openBy(linkAddress, "folder");
                        } else if (event.shiftKey) {
                            openBy(linkAddress, "app");
                        } else if (Constants.SCRIBLI_ASSETS_EXTS.includes(pathPosix().extname((linkAddress).split("?")[0]))) {
                            openAsset(protyle.app, linkAddress.split("?page")[0], parseInt(getSearch("page", linkAddress)));
                        }
                    } else {
                        if (processScribliUri(protyle.app, linkAddress)) {
                            break;
                        }
                        /// #if !BROWSER
                        shell.openExternal(linkAddress).catch((e) => {
                            showMessage(e);
                        });
                        /// #else
                        window.open(linkAddress);
                        /// #endif
                    }
                    break;
                } else if (target.tagName === "IMG") {
                    previewDocImage((event.target as HTMLElement).getAttribute("src"), protyle.block.rootID);
                    event.stopPropagation();
                    event.preventDefault();
                    break;
                } else if (target.tagName === "BUTTON") {
                    const type = target.getAttribute("data-type");
                    const actionCustom = actions.find((w: IPreviewActionCustom) => w?.key === type) as IPreviewActionCustom;
                    if (actionCustom) {
                        actionCustom.click(type);
                    } else if (type === "desktop") {
                        previewElement.style.width = "";
                        previewElement.style.padding = protyle.wysiwyg.element.style.padding;
                    } else if (type === "tablet") {
                        previewElement.style.width = "1024px";
                        previewElement.style.padding = "8px 16px";
                    } else {
                        previewElement.style.width = "360px";
                        previewElement.style.padding = "8px";
                    }
                    actionElement.querySelectorAll("button").forEach((item) => {
                        item.classList.remove("protyle-preview__action--current");
                    });
                    target.classList.add("protyle-preview__action--current");
                }
                target = target.parentElement;
            }
            const nodeElement = hasTopClosestByAttribute(event.target as HTMLElement, "id", undefined);
            if (nodeElement) {
                this.element.querySelectorAll(".protyle-wysiwyg--select").forEach(item => {
                    item.classList.remove("selected");
                });
                nodeElement.classList.add("selected");
                if (protyle.model) {
                    getAllModels().outline.forEach(item => {
                        if (item.blockId === protyle.block.rootID) {
                            item.setCurrentByPreview(nodeElement);
                        }
                    });
                }
                const diagramElement = getDiagramBlock(nodeElement);
                if (diagramElement) {
                    previewDiagram(diagramElement);
                    event.stopPropagation();
                    event.preventDefault();
                    return;
                }
            }
        });

        this.previewElement = previewElement;
    }

    public render(protyle: IProtyle) {
        if (this.element.style.display === "none") {
            return;
        }
        if (this.element.querySelector('.protyle-preview__action [data-type="desktop"]')?.classList.contains("protyle-preview__action--current")) {
            const padding = getPadding(protyle);
            this.previewElement.style.padding = `${padding.top}px ${padding.left}px ${padding.bottom}px ${padding.right}px`;
        }

        let loadingElement = this.element.querySelector(".fn__loading");
        if (!loadingElement) {
            this.element.insertAdjacentHTML("beforeend", `<div style="flex-direction: column;" class="fn__loading">
    <img width="48px" src="/stage/loading-pure.svg">
</div>`);
            loadingElement = this.element.querySelector(".fn__loading");
        }
        this.mdTimeoutId = window.setTimeout(() => {
            fetchPost("/api/export/preview", {
                id: protyle.block.id || protyle.options.blockId || protyle.block.parentID,
            }, response => {
                const oldScrollTop = protyle.preview.previewElement.scrollTop;
                protyle.preview.previewElement.innerHTML = response.data.html;
                processRender(protyle.preview.previewElement);
                highlightRender(protyle.preview.previewElement);
                avRender(protyle.preview.previewElement, protyle);
                speechRender(protyle.preview.previewElement, window.scribli.config.appearance.lang);
                protyle.preview.previewElement.scrollTop = oldScrollTop;
                loadingElement.remove();
            });
        }, protyle.options.preview.delay);
    }

}

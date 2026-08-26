import {Constants} from "../constants";
import {Hint} from "./hint";
import {getLute} from "./render/setLute";
import {Preview} from "./preview";
import {addLoading, initUI, removeLoading} from "./ui/initUI";
import {LocalUndo, Undo} from "./undo";
import {Upload} from "./upload";
import {Options} from "./util/Options";
import {destroy} from "./util/destroy";
import {Scroll} from "./scroll";
import {Model} from "../layout/Model";
import {genUUID} from "../util/genID";
import {WYSIWYG} from "./wysiwyg";
import {Toolbar} from "./toolbar";
import {Gutter} from "./gutter";
import {Breadcrumb} from "./breadcrumb";
import {
    onTransaction,
    transaction,
    turnsIntoOneTransaction,
    turnsIntoTransaction,
    updateBatchTransaction,
    updateTransaction
} from "./wysiwyg/transaction";
import {fetchPost} from "../util/fetch";
import {getDocDisplayName, isEncryptedBox} from "../util/pathName";
import {initMirror, refreshUndoButtons, syncMirrorFromBroadcast} from "./undo/globalUndo";
import {updatePanelByEditor} from "../editor/util";
import {setPanelFocus} from "../layout/util";
import {Title} from "./header/Title";
import {Background} from "./header/Background";
import {disabledProtyle, enableProtyle, onGet, setReadonlyByConfig} from "./util/onGet";
import {reloadProtyle} from "./util/reload";
import {renderBacklink} from "./wysiwyg/renderBacklink";
import {resize} from "./util/resize";
import {getDocByScroll} from "./scroll/saveScroll";
import {App} from "../index";
import {insertHTML} from "./util/insertHTML";
import {avRender} from "./render/av/render";
import {focusBlock, getEditorRange} from "./util/selection";
import {hasClosestBlock} from "./util/hasClosest";
import {setStorageVal} from "./util/compatibility";
import {merge} from "./util/merge";
import {getAllModels} from "../layout/getAll";
import {isSupportCSSHL} from "./render/searchMarkRender";
import {renderAVAttribute} from "./render/av/blockAttr";
import {setFoldById, zoomOut} from "../menus/protyle";
import {setEditMode} from "./util/setEditMode";
import {clearChildDocs, renderChildDocs} from "./util/childDocs";

export class Protyle {

    public readonly version: string;
    public protyle: IProtyle;

    /**
     */
    constructor(app: App, id: HTMLElement, options: IProtyleOptions) {
        this.version = Constants.SCRIBLI_VERSION;
        let pluginsOptions: IProtyleOptions = options;
        app.plugins.forEach(item => {
            if (item.protyleOptions) {
                pluginsOptions = merge(pluginsOptions, item.protyleOptions);
            }
        });
        const getOptions = new Options(pluginsOptions);
        const mergedOptions = getOptions.merge();
        this.protyle = {
            getInstance: () => this,
            app,
            id: genUUID(),
            disabled: false,
            lite: !!options.lite,
            updated: false,
            element: id,
            notebookId: mergedOptions.notebookId,
            options: mergedOptions,
            block: {},
            highlight: {
                mark: isSupportCSSHL() ? new Highlight() : undefined,
                markHL: isSupportCSSHL() ? new Highlight() : undefined,
                ranges: [],
                rangeIndex: 0,
                styleElement: document.createElement("style"),
            }
        };

        if (isSupportCSSHL()) {
            const styleId = genUUID();
            this.protyle.highlight.styleElement.dataset.uuid = styleId;
            this.protyle.highlight.styleElement.textContent = `.protyle-content::highlight(search-mark-${styleId}) {background-color: var(--b3-highlight-background);color: var(--b3-highlight-color);}
  .protyle-content::highlight(search-mark-hl-${styleId}) {color: var(--b3-highlight-color);background-color: var(--b3-highlight-current-background)}`;
        }

        this.protyle.hint = new Hint(this.protyle);
        if (mergedOptions.render.breadcrumb) {
            this.protyle.breadcrumb = new Breadcrumb(this.protyle);
        }
        if (mergedOptions.render.title) {
            this.protyle.title = new Title(this.protyle);
        }
        if (mergedOptions.render.background) {
            this.protyle.background = new Background(this.protyle);
        }

        this.protyle.element.innerHTML = "";
        this.protyle.element.classList.add("protyle");
        if (window.scribli.config.editor.rtl) {
            this.protyle.element.classList.add("rtl");
        }
        if (mergedOptions.render.breadcrumb) {
            this.protyle.element.appendChild(this.protyle.breadcrumb.element.parentElement);
        }
        this.protyle.undo = this.protyle.lite ? new LocalUndo() : new Undo();
        this.protyle.wysiwyg = new WYSIWYG(this.protyle);
        this.protyle.toolbar = new Toolbar(this.protyle);
        this.protyle.scroll = new Scroll(this.protyle);
        if (this.protyle.options.render.gutter) {
            this.protyle.gutter = new Gutter(this.protyle);
        }
        if (mergedOptions.upload.url || mergedOptions.upload.handler) {
            this.protyle.upload = new Upload();
        }

        this.init();
        if (!mergedOptions.action.includes(Constants.CB_GET_HISTORY)) {
            this.protyle.ws = new Model({app});
            this.protyle.ws.connect({
                id: this.protyle.id,
                type: "protyle",
                msgCallback: (data) => {
                    switch (data.cmd) {
                        case "reload":
                            if (data.data === this.protyle.block.rootID) {
                                reloadProtyle(this.protyle, false);
                                getAllModels().outline.forEach(item => {
                                    if (item.blockId === data.data) {
                                        const outlineParam: IObject = {
                                            id: item.blockId,
                                            preview: item.isPreview
                                        };
                                        if (isEncryptedBox(this.protyle.notebookId)) {
                                            outlineParam.notebook = this.protyle.notebookId;
                                        }
                                        fetchPost("/api/outline/getDocOutline", outlineParam, response => {
                                            item.update(response);
                                        });
                                    }
                                });
                            }
                            break;
                        case "refreshAttributeView":
                            Array.from(this.protyle.wysiwyg.element.querySelectorAll(`.av[data-av-id="${data.data.id}"]`)).forEach((item: HTMLElement) => {
                                item.removeAttribute("data-render");
                                avRender(item, this.protyle);
                            });
                            if (this.protyle.databaseAttributePanel?.hasDatabase(data.data.id)) {
                                this.protyle.databaseAttributePanel.refresh();
                            }
                            getAllModels().custom.forEach((item) => {
                                if (item.type === "scribli-database-row" && (item.data.avID === data.data.id ||
                                    item.element.querySelector(`[data-av-id="${data.data.id}"]`))) {
                                    item.update?.();
                                }
                            });
                            break;
                        case "addLoading":
                            if (data.data === this.protyle.block.rootID) {
                                addLoading(this.protyle, data.msg);
                            }
                            break;
                        case "unfoldHeading":
                            setFoldById(data.data, this.protyle);
                            break;
                        case "transactions":
                            this.onTransaction(data);
                            break;
                        case "readonly":
                            window.scribli.config.editor.readOnly = data.data;
                            setReadonlyByConfig(this.protyle, true);
                            break;
                        case "heading2doc":
                        case "li2doc":
                            if (this.protyle.block.rootID === data.data.srcRootBlockID) {
                                if (this.protyle.block.showAll && data.cmd === "heading2doc" && !this.protyle.options.backlinkData) {
                                    const getDocParam: IObject = {
                                        id: this.protyle.block.rootID,
                                        size: window.scribli.config.editor.dynamicLoadBlocks,
                                    };
                                    if (isEncryptedBox(this.protyle.notebookId)) {
                                        getDocParam.notebook = this.protyle.notebookId;
                                    }
                                    fetchPost("/api/filetree/getDoc", getDocParam, getResponse => {
                                        onGet({data: getResponse, protyle: this.protyle});
                                    });
                                } else {
                                    reloadProtyle(this.protyle, false);
                                }
                                if (data.cmd === "heading2doc") {
                                    updatePanelByEditor({
                                        protyle: this.protyle,
                                        focus: false,
                                        pushBackStack: false,
                                        reload: true,
                                        resize: false
                                    });
                                }
                            }
                            renderChildDocs(this.protyle);
                            break;
                        case "rename":
                            if (this.protyle.path === data.data.path) {
                                if (this.protyle.model) {
                                    this.protyle.model.parent.updateTitle(getDocDisplayName(data.data.title, data.data.empty));
                                }
                                if (this.protyle.background) {
                                    this.protyle.background.ial.title = data.data.title;
                                }
                                if (window.scribli.config.export.addTitle &&
                                    !this.protyle.preview.element.classList.contains("fn__none")) {
                                    this.protyle.preview.render(this.protyle);
                                }
                            }
                            if (this.protyle.options.render.title && this.protyle.block.parentID === data.data.id) {
                                if (!document.body.classList.contains("body--blur") && getSelection().rangeCount > 0 &&
                                    this.protyle.title.editElement?.contains(getSelection().getRangeAt(0).startContainer)) {
                                    // Intentionally empty.
                                } else {
                                    this.protyle.title.setTitle(data.data.title, data.data.empty);
                                }
                                if (data.data.empty) {
                                    this.protyle.wysiwyg.element.setAttribute(Constants.CUSTOM_SY_TITLE_EMPTY, "true");
                                } else {
                                    this.protyle.wysiwyg.element.removeAttribute(Constants.CUSTOM_SY_TITLE_EMPTY);
                                }
                            }
                            // update ref
                            this.protyle.wysiwyg.element.querySelectorAll(`[data-type~="block-ref"][data-id="${data.data.id}"]`).forEach(item => {
                                if (item.getAttribute("data-subtype") === "d") {
                                    item.innerHTML = data.data.refText;
                                }
                            });
                            renderChildDocs(this.protyle);
                            break;
                        case "moveDoc":
                            if (this.protyle.path === data.data.fromPath) {
                                this.protyle.path = data.data.newPath;
                                this.protyle.notebookId = data.data.toNotebook;
                            }
                            renderChildDocs(this.protyle);
                            break;
                        case "closeBox":
                        case "removeBox":
                            if (this.protyle.notebookId === data.data.box) {
                                if (this.protyle.model) {
                                    this.protyle.model.parent.parent.removeTab(this.protyle.model.parent.id);
                                }
                            }
                            break;
                        case "removeDoc":
                            if (data.data.ids.includes(this.protyle.block.rootID)) {
                                if (this.protyle.model) {
                                    this.protyle.model.parent.parent.removeTab(this.protyle.model.parent.id);
                                }
                                delete window.scribli.storage[Constants.LOCAL_FILEPOSITION][this.protyle.block.rootID];
                                setStorageVal(Constants.LOCAL_FILEPOSITION, window.scribli.storage[Constants.LOCAL_FILEPOSITION]);
                            } else {
                                renderChildDocs(this.protyle);
                            }
                            break;
                        case "create":
                        case "createdailynote":
                            renderChildDocs(this.protyle);
                            break;
                    }
                }
            });
            if (options.backlinkData) {
                this.protyle.block.rootID = options.blockId;
                renderBacklink(this.protyle, options.backlinkData);
                this.protyle.wysiwyg.element.style.padding = "4px 16px 4px 24px";
                clearChildDocs(this.protyle);
                return;
            }
            if (!options.blockId) {
                removeLoading(this.protyle);
                return;
            }

            if (this.protyle.options.mode !== "preview" &&
                options.rootId && window.scribli.storage[Constants.LOCAL_FILEPOSITION][options.rootId] &&
                (
                    mergedOptions.action.includes(Constants.CB_GET_SCROLL) ||
                    (mergedOptions.action.includes(Constants.CB_GET_ROOTSCROLL) && options.rootId === options.blockId)
                )
            ) {
                getDocByScroll({
                    protyle: this.protyle,
                    scrollAttr: window.scribli.storage[Constants.LOCAL_FILEPOSITION][options.rootId],
                    mergedOptions,
                    cb: () => {
                        this.afterOnGet(mergedOptions);
                    }
                });
            } else {
                this.getDoc(mergedOptions);
            }
        } else {
            this.protyle.contentElement.classList.add("protyle-content--transition");
        }
    }

    private onTransaction(data: IWebSocketData) {
        if (data.context?.undoState) {
            syncMirrorFromBroadcast(data.context.undoState);
        }
        if (!this.protyle.preview.element.classList.contains("fn__none") &&
            data.context?.rootIDs?.includes(this.protyle.block.rootID)) {
            this.protyle.preview.render(this.protyle);
            return;
        }
        const hadContent = this.protyle.wysiwyg.element.childElementCount > 0;
        let needCreateAction = "";
        let hasDeleteOp = false;
        data.data[0].doOperations.find((item: IOperation) => {
            if (this.protyle.options.backlinkData && ["delete", "move"].includes(item.action)) {
                if (2 == data.data[0].doOperations.length && "insert" === data.data[0].doOperations[0].action && "delete" === data.data[0].doOperations[1].action) {
                    // The list in the backlink panel no longer collapses automatically 
                    return true;
                }

                getAllModels().backlink.find(backlinkItem => {
                    if (backlinkItem.element.contains(this.protyle.element)) {
                        backlinkItem.refresh();
                        return true;
                    }
                });
                return true;
            } else {
                if (item.action === "delete") {
                    hasDeleteOp = true;
                }
                onTransaction(this.protyle, [item], false);
                if (!(item.action === "delete" && typeof item.data?.createEmptyParagraph === "boolean" && !item.data.createEmptyParagraph)) {
                    needCreateAction = item.action;
                }
            }
        });
        // Improve editor state synchronization when deleting blocks 
        if (this.protyle.block.showAll && hasDeleteOp) {
            fetchPost("/api/block/checkBlockExist", {id: this.protyle.block.id}, response => {
                if (!response.data) {
                    zoomOut({
                        protyle: this.protyle,
                        id: this.protyle.block.rootID
                    });
                }
            });
            return;
        }
        if (this.protyle.element.dataset.loading === "finished" && hadContent &&
            this.protyle.wysiwyg.element.childElementCount === 0 && this.protyle.block.parentID && needCreateAction) {
            if (needCreateAction === "delete" && this.protyle.block.showAll) {
                if (this.protyle.options.handleEmptyContent) {
                    this.protyle.options.handleEmptyContent();
                } else {
                    zoomOut({
                        protyle: this.protyle,
                        id: this.protyle.block.rootID,
                        focusId: this.protyle.block.id
                    });
                }
            } else {
                refreshUndoButtons(this.protyle);
                this.reload(false);
            }
        }
        if (data.context?.isUndoReplay === true) {
            this.protyle.wysiwyg.lastHTMLs = {};
        }
    }

    private getDoc(mergedOptions: IProtyleOptions) {
        const getDocParam: Record<string, any> = {
            id: mergedOptions.blockId,
            isBacklink: mergedOptions.action.includes(Constants.CB_GET_BACKLINK),
            originalRefBlockIDs: mergedOptions.originalRefBlockIDs,
            mode: (mergedOptions.action && mergedOptions.action.includes(Constants.CB_GET_CONTEXT)) ? 3 : 0,
            size: mergedOptions.action?.includes(Constants.CB_GET_ALL) ? Constants.SIZE_GET_MAX : window.scribli.config.editor.dynamicLoadBlocks,
        };
        if (isEncryptedBox(this.protyle.notebookId)) {
            getDocParam.notebook = this.protyle.notebookId;
        }
        fetchPost("/api/filetree/getDoc", getDocParam, getResponse => {
            onGet({
                data: getResponse,
                protyle: this.protyle,
                action: mergedOptions.action,
                scrollPosition: mergedOptions.scrollPosition,
                afterCB: () => {
                    this.afterOnGet(mergedOptions);
                }
            });
        });
    }

    private afterOnGet(mergedOptions: IProtyleOptions) {
        if (this.protyle.block?.rootID) {
            initMirror(this.protyle.block.rootID);
        }
        if (this.protyle.model) {
            if (mergedOptions.action?.includes(Constants.CB_GET_FOCUS) || mergedOptions.action?.includes(Constants.CB_GET_OPENNEW)) {
                setPanelFocus(this.protyle.model.element.parentElement.parentElement);
            }
            updatePanelByEditor({
                protyle: this.protyle,
                focus: false,
                pushBackStack: false,
                reload: false,
                resize: false
            });
        }
        resize(this.protyle);
        this.protyle.wysiwyg.element.addEventListener("focusin", () => {
            if (this.protyle && this.protyle.model) {
                let needUpdate = true;
                if (this.protyle.model.element.parentElement.parentElement.classList.contains("layout__wnd--active") && this.protyle.model.headElement.classList.contains("item--focus")) {
                    needUpdate = false;
                }
                if (!needUpdate) {
                    return;
                }
                setPanelFocus(this.protyle.model.element.parentElement.parentElement);
                updatePanelByEditor({
                    protyle: this.protyle,
                    focus: false,
                    pushBackStack: false,
                    reload: false,
                    resize: false,
                });
            } else {
                document.querySelectorAll(".layout__tab--active").forEach(item => {
                    item.classList.remove("layout__tab--active");
                });
                document.querySelectorAll(".layout__wnd--active").forEach(item => {
                    item.classList.remove("layout__wnd--active");
                });
            }
        });
        if (mergedOptions.after) {
            mergedOptions.after(this);
        }
        this.protyle.contentElement.classList.add("protyle-content--transition");
    }

    private init() {
        this.protyle.lute = getLute({
            emojiSite: this.protyle.options.hint.emojiPath,
            emojis: this.protyle.options.hint.emoji,
            headingAnchor: false,
            listStyle: this.protyle.options.preview.markdown.listStyle,
            sanitize: this.protyle.options.preview.markdown.sanitize,
        });

        this.protyle.preview = new Preview(this.protyle);

        initUI(this.protyle);
    }

    public focus() {
        this.protyle.wysiwyg.element.focus();
    }

    public isUploading() {
        return this.protyle.upload.isUploading;
    }

    public clearStack() {
        this.protyle.undo.clear();
    }

    public destroy() {
        destroy(this.protyle);
    }

    public resize() {
        resize(this.protyle);
    }

    public reload(focus: boolean, updateReadonly?: boolean) {
        reloadProtyle(this.protyle, focus, updateReadonly);
    }

    public insert(html: string, isBlock = false, useProtyleRange = false) {
        insertHTML(html, this.protyle, isBlock, useProtyleRange);
    }

    public transaction(doOperations: IOperation[], undoOperations?: IOperation[]) {
        transaction(this.protyle, doOperations, undoOperations);
    }

    /**
     */
    public turnIntoOneTransaction(selectsElement: Element[], type: TTurnIntoOne, subType?: TTurnIntoOneSub) {
        turnsIntoOneTransaction({
            protyle: this.protyle,
            selectsElement,
            type,
            level: subType
        });
    }

    /**
     */
    public turnIntoTransaction(nodeElement: Element, type: TTurnInto, subType?: number) {
        turnsIntoTransaction({
            protyle: this.protyle,
            nodeElement,
            type,
            level: subType,
        });
    }

    /**
     */
    public updateTransaction(id: string, newHTML: string, html: string) {
        const element = document.createElement("template");
        element.innerHTML = newHTML;
        updateTransaction(this.protyle, element.content.firstElementChild, html);
    }

    public updateTransactionElement(element: Element, oldHTML: string) {
        updateTransaction(this.protyle, element, oldHTML);
    }

    public updateBatchTransaction(nodeElements: Element[], cb: (e: HTMLElement) => void) {
        updateBatchTransaction(nodeElements, this.protyle, cb);
    }

    public getRange(element: Element) {
        return getEditorRange(element);
    }

    public hasClosestBlock(element: Node) {
        return hasClosestBlock(element);
    }

    public focusBlock(element: Element, toStart = true) {
        return focusBlock(element, undefined, toStart);
    }

    public disable() {
        disabledProtyle(this.protyle);
    }

    public enable() {
        enableProtyle(this.protyle);
    }

    public renderAVAttribute(element: HTMLElement, id: string, cb?: (element: HTMLElement) => void) {
        renderAVAttribute(element, id, this.protyle, cb);
    }

    public switchMode(mode: TEditorMode) {
        setEditMode(this.protyle, mode);
    }
}

import {fetchSyncPost} from "../../util/fetch";
import {escapeAttr, escapeHtml} from "../../util/escape";
import {hasClosestByClassName} from "../util/hasClosest";
import {genIconHTML} from "./util";
import {openAsset, openFileById} from "../../editor/util";
import {openSearch} from "../../search/spread";
import {Constants} from "../../constants";
import {pathPosix} from "../../util/pathName";
import {Dialog} from "../../dialog";

interface ICanvasNode {
    id: string;
    type?: string;
    text?: string;
    file?: string;
    x?: number;
    y?: number;
    width?: number;
    height?: number;
    scribli?: {
        kind?: string;
        label?: string;
        refID?: string;
        query?: string;
        assetPath?: string;
        databaseID?: string;
        viewID?: string;
    };
}

interface ICanvasEdge {
    id?: string;
    fromNode: string;
    toNode: string;
}

interface ICanvasPayload {
    nodes?: ICanvasNode[];
    edges?: ICanvasEdge[];
    scribli?: {
        id?: string;
        title?: string;
        updated?: string;
    };
}

interface ILoadedCanvas {
    canvas: ICanvasPayload;
    id?: string;
    placeholder?: boolean;
}

interface ICanvasViewState {
    scale: number;
    panX: number;
    panY: number;
    selectedNodeID?: string;
    connectFromID?: string;
    history: ICanvasPayload[];
    panel?: "library" | "templates";
}

const canvasStates = new WeakMap<HTMLElement, ICanvasViewState>();
const canvasRuntime = new WeakMap<HTMLElement, { canvas: ICanvasPayload, id?: string, blockElement?: HTMLElement, placeholder?: boolean }>();
let canvasToolbarDelegationWired = false;
let lastCanvasPointerCommand = 0;
let lastCanvasNodePointerCommand = {time: 0, nodeID: ""};

const DEFAULT_NODE_WIDTH = 300;
const DEFAULT_NODE_HEIGHT = 160;
const MIN_NODE_WIDTH = 160;
const MIN_NODE_HEIGHT = 90;

export const canvasRender = (element: Element) => {
    let canvasElements: Element[] | NodeListOf<Element> = [];
    if (isCanvasBlock(element)) {
        canvasElements = [element];
    } else {
        canvasElements = Array.from(element.querySelectorAll('[data-subtype="scribli-canvas"], [data-type="NodeCodeBlock"]')).filter(isCanvasBlock);
    }
    if (canvasElements.length === 0) {
        return;
    }

    const wysiwygElement = hasClosestByClassName(element, "protyle-wysiwyg", true);
    canvasElements.forEach((item: HTMLDivElement) => {
        const isCodeBlock = item.getAttribute("data-type") === "NodeCodeBlock";
        item.setAttribute("data-render", "true");
        item.setAttribute("data-subtype", "scribli-canvas");
        if (isCodeBlock) {
            item.classList.remove("code-block");
            item.classList.add("render-node");
            item.setAttribute("data-content", canvasBlockContent(item));
            item.innerHTML = `${genIconHTML(wysiwygElement, ["refresh", "edit", "more"])}<div></div><div class="protyle-attr" contenteditable="false">${Constants.ZWSP}</div>`;
        } else if (!item.firstElementChild?.classList.contains("protyle-icons")) {
            item.insertAdjacentHTML("afterbegin", genIconHTML(wysiwygElement, ["refresh", "edit", "more"]));
        }
        const renderElement = item.firstElementChild.nextElementSibling as HTMLElement;
        if (!renderElement) {
            return;
        }
        renderElement.innerHTML = `<div class="scribli-canvas" contenteditable="false"><div class="scribli-canvas__status">${escapeHtml(lang("loading", "Loading..."))}</div></div>`;
        loadCanvas(item).then(({canvas, id, placeholder}) => {
            renderCanvasSurface(renderElement.firstElementChild as HTMLElement, normalizeCanvasPayload(canvas), id, item, placeholder);
        }).catch((error) => {
            renderElement.innerHTML = `<div class="scribli-canvas scribli-canvas--error" contenteditable="false">${escapeHtml(error instanceof Error ? error.message : String(error))}</div>`;
        });
    });
};

const isCanvasBlock = (element: Element) => {
    if (element.querySelector(".scribli-canvas")) {
        return false;
    }
    if (element.getAttribute("data-subtype") === "scribli-canvas") {
        return true;
    }
    if (element.getAttribute("data-type") !== "NodeCodeBlock") {
        return false;
    }
    const language = element.querySelector(".protyle-action__language")?.textContent?.trim().toLowerCase();
    return language === "scribli-canvas";
};

const loadCanvas = async (blockElement: HTMLElement): Promise<ILoadedCanvas> => {
    const content = canvasBlockContent(blockElement);
    if (content === "") {
        return {canvas: {nodes: [], edges: []}, placeholder: true};
    }
    if (content.startsWith("{")) {
        return {canvas: JSON.parse(content)};
    }
    const id = content.split(/\s+/)[0];
    const response = await fetchSyncPost("/api/canvas/call", {action: "get", id});
    if (response.code !== 0 || response.data?.isError) {
        throw new Error(response.msg || lang("canvasLoadFailed", "Canvas load failed"));
    }
    return {canvas: response.data?.structuredContent?.canvas || {}, id};
};

const canvasBlockContent = (blockElement: Element) => {
    const dataContent = Lute.UnEscapeHTMLStr(blockElement.getAttribute("data-content") || "").trim();
    if (dataContent) {
        return dataContent;
    }
    return (blockElement.querySelector(".hljs [contenteditable='true']")?.textContent || "").trim();
};

const renderCanvasSurface = (surfaceElement: HTMLElement, canvas: ICanvasPayload, id?: string, blockElement?: HTMLElement, placeholder = false) => {
    canvas = normalizeCanvasPayload(canvas);
    const state = canvasState(surfaceElement);
    const nodes = canvas.nodes || [];
    const edges = canvas.edges || [];
    const bounds = canvasBounds(nodes);
    const isBlankUnboundCanvas = !id && nodes.length === 0 && edges.length === 0;
    const toolbarHTML = id ? canvasToolbarHTML(state) : "";
    const emptyStateHTML = placeholder || isBlankUnboundCanvas ? emptyCanvasHTML() : "";
    const panelHTML = state.panel ? canvasPanelHTML(state.panel) : "";
    surfaceElement.innerHTML = `${toolbarHTML}
<div class="scribli-canvas__viewport" data-type="canvas-viewport" style="height:${bounds.height}px">
    <div class="scribli-canvas__stage" style="width:${bounds.width}px;height:${bounds.height}px;transform: translate(${state.panX}px, ${state.panY}px) scale(${state.scale})">
        <svg class="scribli-canvas__edges" viewBox="${bounds.minX} ${bounds.minY} ${bounds.width} ${bounds.height}" preserveAspectRatio="none">${edgeHTML(edges, nodes)}</svg>
        <div class="scribli-canvas__nodes" style="transform: translate(${-bounds.minX}px, ${-bounds.minY}px)">${nodes.map((node) => nodeHTML(node, state)).join("")}</div>
    </div>
    ${emptyStateHTML}
    ${panelHTML}
</div>`;
    wireCanvasToolbar(surfaceElement, canvas, id, blockElement, placeholder);
    wireCanvasActions(surfaceElement, canvas, id, blockElement);
    wireCanvasDragging(surfaceElement, canvas, id, blockElement);
    wireCanvasResizing(surfaceElement, canvas, id, blockElement);
    wireCanvasPanning(surfaceElement, canvas, id, blockElement);
    wireCanvasEditing(surfaceElement, canvas, id, blockElement);
};

const canvasToolbarHTML = (state: ICanvasViewState) => `<div class="scribli-canvas__toolbar">
    ${iconButton("canvas-add-text", "iconAdd", lang("text", "Text"))}
    ${iconButton("canvas-add-document", "iconFile", lang("document", "Document"))}
    ${iconButton("canvas-add-block", "iconFiles", lang("block", "Block"))}
    ${iconButton("canvas-add-asset", "iconImage", lang("assets", "Assets"))}
    ${iconButton("canvas-add-query", "iconSearch", lang("search", "Search"))}
    ${iconButton("canvas-add-database", "iconDatabase", lang("database", "Database"))}
    <span class="fn__space"></span>
    ${iconButton("canvas-connect", "iconLink", lang("canvasConnect", "Connect"), state.connectFromID ? " scribli-canvas__toolbar-button--active" : "")}
    ${iconButton("canvas-duplicate", "iconCopy", lang("duplicate", "Duplicate"))}
    ${iconButton("canvas-delete", "iconTrashcan", lang("delete", "Delete"))}
    ${iconButton("canvas-undo", "iconRefresh", lang("undo", "Undo"))}
    <span class="fn__space"></span>
    ${iconButton("canvas-library", "iconFiles", lang("canvasLibrary", "Canvas Library"), state.panel === "library" ? " scribli-canvas__toolbar-button--active" : "")}
    ${iconButton("canvas-templates", "iconBoth", lang("canvasTemplates", "Templates"), state.panel === "templates" ? " scribli-canvas__toolbar-button--active" : "")}
    ${iconButton("canvas-import", "iconUpload", lang("import", "Import"))}
    ${iconButton("canvas-export", "iconDownload", lang("export", "Export"))}
    <input class="fn__none" data-type="canvas-import-file" type="file" accept=".canvas,.json,application/json">
    <span class="fn__space"></span>
    ${iconButton("canvas-zoom-out", "iconZoomOut", lang("zoomOut", "Zoom out"))}
    <span class="scribli-canvas__zoom">${Math.round(state.scale * 100)}%</span>
    ${iconButton("canvas-zoom-in", "iconZoomIn", lang("zoomIn", "Zoom in"))}
    ${iconButton("canvas-reset-view", "iconRefresh", lang("reset", "Reset"))}
</div>`;

const emptyCanvasHTML = () => `<div class="scribli-canvas__empty">
    <div class="scribli-canvas__empty-actions">
        <button class="b3-button b3-button--outline" data-type="canvas-create"><svg><use xlink:href="#iconAdd"></use></svg>${escapeHtml(lang("createCanvas", "Create Canvas"))}</button>
        ${canvasTemplateButtonsHTML("canvas-create-template")}
        <button class="b3-button b3-button--outline" data-type="canvas-library"><svg><use xlink:href="#iconFiles"></use></svg>${escapeHtml(lang("canvasLibrary", "Canvas Library"))}</button>
        <button class="b3-button b3-button--outline" data-type="canvas-import"><svg><use xlink:href="#iconUpload"></use></svg>${escapeHtml(lang("import", "Import"))}</button>
        <input class="fn__none" data-type="canvas-import-file" type="file" accept=".canvas,.json,application/json">
    </div>
</div>`;

const canvasPanelHTML = (panel: "library" | "templates") => `<div class="scribli-canvas__panel" data-panel="${panel}">
    <div class="scribli-canvas__panel-header">
        <strong>${escapeHtml(panel === "library" ? lang("canvasLibrary", "Canvas Library") : lang("canvasTemplates", "Templates"))}</strong>
        ${iconButton("canvas-close-panel", "iconClose", lang("close", "Close"))}
    </div>
    <div class="scribli-canvas__panel-body">
        ${panel === "library" ? `<div class="scribli-canvas__status">${escapeHtml(lang("loading", "Loading..."))}</div>` : canvasTemplateButtonsHTML("canvas-apply-template")}
    </div>
</div>`;

const canvasTemplateButtonsHTML = (type: string) => CANVAS_TEMPLATES.map((template) =>
    `<button class="b3-button b3-button--outline" data-type="${type}" data-template="${escapeAttr(template.id)}">${escapeHtml(template.title)}</button>`
).join("");

const iconButton = (type: string, icon: string, label: string, extraClass = "") =>
    `<button class="block__icon ariaLabel scribli-canvas__toolbar-button${extraClass}" data-type="${type}" data-position="4north" aria-label="${escapeAttr(label)}"><svg><use xlink:href="#${icon}"></use></svg></button>`;

const wireCanvasToolbar = (surfaceElement: HTMLElement, canvas: ICanvasPayload, id?: string, blockElement?: HTMLElement, placeholder = false) => {
    canvasRuntime.set(surfaceElement, {canvas, id, blockElement, placeholder});
    ensureCanvasToolbarDelegation();
    surfaceElement.querySelector('[data-type="canvas-import-file"]')?.addEventListener("change", async (event) => {
        await importCanvasFile(surfaceElement, blockElement, event.target as HTMLInputElement);
    });
    if (canvasState(surfaceElement).panel === "library") {
        renderCanvasLibrary(surfaceElement, blockElement);
    }
};

const ensureCanvasToolbarDelegation = () => {
    if (canvasToolbarDelegationWired) {
        return;
    }
    const handleToolbarEvent = async (event: Event) => {
        const target = (event.target as HTMLElement).closest("button[data-type^=\"canvas-\"], input[data-type^=\"canvas-\"]") as HTMLElement;
        const surfaceElement = target?.closest(".scribli-canvas") as HTMLElement;
        if (!target || !surfaceElement) {
            return;
        }
        const type = target.dataset.type || "";
        if (type === "canvas-import-file") {
            return;
        }
        if (target.closest(".scribli-canvas__node") && (
            type === "canvas-node-duplicate" || type === "canvas-node-delete" || type === "canvas-open-node"
        )) {
            return;
        }
        if (event.type === "click" && Date.now() - lastCanvasPointerCommand < 750) {
            return;
        }
        if (event.type === "pointerdown") {
            lastCanvasPointerCommand = Date.now();
        }
        event.preventDefault();
        event.stopPropagation();
        const runtime = await resolveCanvasRuntime(surfaceElement);
        await handleCanvasCommand(surfaceElement, runtime.canvas, runtime.id, runtime.blockElement, target, runtime.placeholder);
    };
    const handleNodeClick = async (event: Event) => {
        const target = event.target as HTMLElement;
        const nodeElement = target.closest(".scribli-canvas__node") as HTMLElement;
        const surfaceElement = nodeElement?.closest(".scribli-canvas") as HTMLElement;
        if (!nodeElement || !surfaceElement || isCanvasNodeControlTarget(target, nodeElement)) {
            return;
        }
        const runtime = await resolveCanvasRuntime(surfaceElement);
        const nodeID = nodeElement.dataset.nodeId || "";
        const state = canvasState(surfaceElement);
        if (event.type === "click" && lastCanvasNodePointerCommand.nodeID === nodeID && Date.now() - lastCanvasNodePointerCommand.time < 750) {
            return;
        }
        if (event.type === "pointerdown" && state.connectFromID === undefined) {
            lastCanvasNodePointerCommand = {time: Date.now(), nodeID};
            state.selectedNodeID = state.selectedNodeID === nodeID ? undefined : nodeID;
            updateCanvasNodeSelection(surfaceElement, state.selectedNodeID);
            return;
        }
        event.preventDefault();
        event.stopPropagation();
        if (state.connectFromID !== undefined && runtime.id) {
            await connectNode(surfaceElement, runtime.canvas, runtime.id, runtime.blockElement, nodeID);
            return;
        }
        state.selectedNodeID = state.selectedNodeID === nodeID ? undefined : nodeID;
        renderCanvasSurface(surfaceElement, runtime.canvas, runtime.id, runtime.blockElement);
    };
    document.addEventListener("pointerdown", handleToolbarEvent, true);
    document.addEventListener("click", handleToolbarEvent, true);
    document.addEventListener("pointerdown", handleNodeClick, true);
    document.addEventListener("click", handleNodeClick, true);
    canvasToolbarDelegationWired = true;
};

const updateCanvasNodeSelection = (surfaceElement: HTMLElement, selectedNodeID?: string) => {
    surfaceElement.querySelectorAll(".scribli-canvas__node").forEach((nodeElement: HTMLElement) => {
        nodeElement.classList.toggle("scribli-canvas__node--selected", nodeElement.dataset.nodeId === selectedNodeID);
    });
};

const isCanvasNodeControlTarget = (target: HTMLElement, nodeElement: HTMLElement) => {
    const typedElement = target.closest("[data-type]");
    if (typedElement && nodeElement.contains(typedElement)) {
        return true;
    }
    const editableElement = target.closest("[contenteditable=true]");
    return !!editableElement && nodeElement.contains(editableElement);
};

const resolveCanvasRuntime = async (surfaceElement: HTMLElement) => {
    const runtime = canvasRuntime.get(surfaceElement);
    if (runtime?.id) {
        return runtime;
    }
    const blockElement = currentCanvasBlock(surfaceElement, runtime?.blockElement);
    if (!blockElement) {
        return runtime || {canvas: {nodes: [], edges: []}};
    }
    const loaded = await loadCanvas(blockElement);
    const resolved = {
        canvas: normalizeCanvasPayload(loaded.canvas),
        id: loaded.id,
        blockElement,
        placeholder: loaded.placeholder,
    };
    canvasRuntime.set(surfaceElement, resolved);
    return resolved;
};

const handleCanvasCommand = async (surfaceElement: HTMLElement, canvas: ICanvasPayload, id: string | undefined, blockElement: HTMLElement | undefined, target: HTMLElement, placeholder = false) => {
    const state = canvasState(surfaceElement);
    const type = target.dataset.type || "";
    const isPlaceholder = placeholder || (!id && (canvas.nodes || []).length === 0 && (canvas.edges || []).length === 0);
    if (type === "canvas-create") {
        await createCanvasFromPlaceholder(surfaceElement, blockElement);
        return;
    }
    if (type === "canvas-create-template") {
        await createCanvasFromPlaceholder(surfaceElement, blockElement, templateCanvas(target.dataset.template || ""));
        return;
    }
    if (type === "canvas-library") {
        state.panel = state.panel === "library" ? undefined : "library";
        renderCanvasSurface(surfaceElement, canvas, id, blockElement, isPlaceholder);
        return;
    }
    if (type === "canvas-import") {
        (surfaceElement.querySelector('[data-type="canvas-import-file"]') as HTMLInputElement)?.click();
        return;
    }
    if (type === "canvas-close-panel") {
        state.panel = undefined;
        renderCanvasSurface(surfaceElement, canvas, id, blockElement, isPlaceholder);
        return;
    }
    if (type === "canvas-open-library") {
        await openCanvasFromLibrary(surfaceElement, target.dataset.id || "");
        return;
    }
    if (!id) {
        return;
    }
    switch (type) {
        case "canvas-add-text":
            await addTextNode(surfaceElement, canvas, id, blockElement);
            return;
        case "canvas-add-document":
            await addScribliNodeFromPrompt(surfaceElement, canvas, id, blockElement, "document");
            return;
        case "canvas-add-block":
            await addScribliNodeFromPrompt(surfaceElement, canvas, id, blockElement, "block");
            return;
        case "canvas-add-asset":
            await addScribliNodeFromPrompt(surfaceElement, canvas, id, blockElement, "asset");
            return;
        case "canvas-add-query":
            await addScribliNodeFromPrompt(surfaceElement, canvas, id, blockElement, "query");
            return;
        case "canvas-add-database":
            await addScribliNodeFromPrompt(surfaceElement, canvas, id, blockElement, "database");
            return;
        case "canvas-connect":
            state.connectFromID = state.connectFromID ? undefined : "";
            renderCanvasSurface(surfaceElement, canvas, id, blockElement);
            return;
        case "canvas-duplicate":
            await duplicateSelectedNode(surfaceElement, canvas, id, blockElement);
            return;
        case "canvas-delete":
            await deleteSelectedNode(surfaceElement, canvas, id, blockElement);
            return;
        case "canvas-undo":
            await undoCanvas(surfaceElement, id, blockElement);
            return;
        case "canvas-templates":
            state.panel = state.panel === "templates" ? undefined : "templates";
            renderCanvasSurface(surfaceElement, canvas, id, blockElement);
            return;
        case "canvas-apply-template":
            await appendTemplate(surfaceElement, canvas, id, blockElement, target.dataset.template || "");
            return;
        case "canvas-export":
            exportCanvas(id, canvas);
            return;
        case "canvas-zoom-in":
            state.scale = clampScale(state.scale + 0.1);
            renderCanvasSurface(surfaceElement, canvas, id, blockElement);
            return;
        case "canvas-zoom-out":
            state.scale = clampScale(state.scale - 0.1);
            renderCanvasSurface(surfaceElement, canvas, id, blockElement);
            return;
        case "canvas-reset-view":
            state.scale = 1;
            state.panX = 0;
            state.panY = 0;
            renderCanvasSurface(surfaceElement, canvas, id, blockElement);
            return;
    }
};

const createCanvasFromPlaceholder = async (surfaceElement: HTMLElement, blockElement?: HTMLElement, initialCanvas?: ICanvasPayload) => {
    blockElement = currentCanvasBlock(surfaceElement, blockElement);
    if (!blockElement) {
        return;
    }
    const buttonElement = surfaceElement.querySelector('[data-type="canvas-create"]') as HTMLButtonElement;
    if (buttonElement) {
        buttonElement.disabled = true;
    }
    try {
        const response = await fetchSyncPost("/api/canvas/call", {
            action: "create",
            title: initialCanvas?.scribli?.title || lang("canvas", "Canvas"),
            canvas: initialCanvas,
        });
        if (response.code !== 0 || response.data?.isError) {
            throw new Error(response.msg || lang("canvasCreateFailed", "Canvas create failed"));
        }
        const structured = response.data?.structuredContent || {};
        const boundBlock = await bindCanvasToBlock(blockElement, structured.id);
        const currentBlock = currentCanvasBlock(surfaceElement, boundBlock, structured.id) || boundBlock;
        const currentSurface = currentCanvasSurface(surfaceElement, currentBlock);
        renderCanvasSurface(currentSurface, structured.canvas || {nodes: [], edges: []}, structured.id, currentBlock);
    } catch (error) {
        surfaceElement.innerHTML = `<div class="scribli-canvas__status ft__error">${escapeHtml(error instanceof Error ? error.message : lang("canvasCreateFailed", "Canvas create failed"))}</div>`;
    }
};

const bindCanvasToBlock = async (blockElement: HTMLElement, id: string) => {
    if (!id) {
        throw new Error(lang("canvasCreateFailed", "Canvas create failed"));
    }
    const markdown = `\`\`\`scribli-canvas\n${id}\n\`\`\``;
    const updateResponse = await fetchSyncPost("/api/block/updateBlock", {
        id: blockElement.getAttribute("data-node-id"),
        dataType: "markdown",
        data: markdown,
    });
    if (updateResponse.code !== 0) {
        throw new Error(updateResponse.msg || lang("canvasCreateFailed", "Canvas create failed"));
    }
    const operation = updateResponse.data?.[0]?.doOperations?.find((item: IOperation) =>
        item.action === "update" && item.id === blockElement.getAttribute("data-node-id") && item.data
    );
    if (operation?.data) {
        const template = document.createElement("template");
        template.innerHTML = operation.data.trim();
        const nextBlock = template.content.firstElementChild as HTMLElement;
        if (nextBlock) {
            blockElement.setAttribute("updated", nextBlock.getAttribute("updated") || blockElement.getAttribute("updated") || "");
            blockElement.setAttribute("data-content", id);
            blockElement.setAttribute("data-subtype", "scribli-canvas");
            blockElement.setAttribute("data-render", "true");
            blockElement.classList.remove("code-block");
            blockElement.classList.add("render-node");
            blockElement.innerHTML = `${blockElement.firstElementChild?.outerHTML || ""}<div><div class="scribli-canvas" contenteditable="false"></div></div><div class="protyle-attr" contenteditable="false">${Constants.ZWSP}</div>`;
            return blockElement;
        }
    }
    blockElement.setAttribute("data-content", id);
    return blockElement;
};

const currentCanvasBlock = (surfaceElement: HTMLElement, blockElement?: HTMLElement, blockID?: string) => {
    const closestBlock = surfaceElement.closest(".protyle-wysiwyg [data-node-id][data-type]") as HTMLElement;
    if (!blockID && isEditableCanvasBlock(closestBlock)) {
        return closestBlock;
    }
    if (isEditableCanvasBlock(blockElement)) {
        return blockElement;
    }
    if (isEditableCanvasBlock(closestBlock)) {
        return closestBlock;
    }
    const id = blockID || blockElement?.getAttribute("data-node-id") || closestBlock?.getAttribute("data-node-id") || surfaceElement.closest("[data-node-id]")?.getAttribute("data-node-id") || "";
    if (!id) {
        return blockElement;
    }
    const selector = `[data-node-id="${escapeSelector(id)}"][data-type]`;
    const roots = [
        blockElement?.closest(".protyle-wysiwyg"),
        surfaceElement.closest(".protyle-wysiwyg"),
        document.querySelector(".layout__wnd--active .protyle-wysiwyg"),
        document,
    ].filter(Boolean) as Array<Element | Document>;
    for (const root of roots) {
        const candidates = Array.from(root.querySelectorAll(selector)) as HTMLElement[];
        const visibleBlock = candidates.find((candidate) => isEditableCanvasBlock(candidate) && candidate.getBoundingClientRect().width > 0);
        if (visibleBlock) {
            return visibleBlock;
        }
        const editableBlock = candidates.find(isEditableCanvasBlock);
        if (editableBlock) {
            return editableBlock;
        }
    }
    return blockElement;
};

const currentCanvasSurface = (surfaceElement: HTMLElement, blockElement?: HTMLElement) => {
    const blockSurface = blockElement?.querySelector(":scope > div > .scribli-canvas") as HTMLElement;
    if (blockSurface) {
        return blockSurface;
    }
    return blockElement?.querySelector(".scribli-canvas") as HTMLElement || surfaceElement;
};

const isEditableCanvasBlock = (element?: Element | null): element is HTMLElement => {
    return !!element && element instanceof HTMLElement &&
        element.isConnected &&
        !!element.closest(".protyle-wysiwyg") &&
        !!element.getAttribute("data-node-id") &&
        !!element.getAttribute("data-type") &&
        !element.classList.contains("protyle-breadcrumb__item") &&
        !element.closest(".popover__block");
};

const escapeSelector = (value: string) => {
    return typeof CSS !== "undefined" && CSS.escape ? CSS.escape(value) : value.replace(/["\\]/g, "\\$&");
};

const addTextNode = async (surfaceElement: HTMLElement, canvas: ICanvasPayload, id: string, blockElement?: HTMLElement) => {
    const text = await canvasInputDialog(lang("canvasTextPrompt", "Text"), lang("new", "New") + " " + lang("text", "Text"));
    if (text === null) {
        return;
    }
    const previous = cloneCanvas(canvas);
    const nodes = canvas.nodes || [];
    nodes.push({
        id: Lute.NewNodeID(),
        type: "text",
        text,
        ...nextNodePosition(nodes),
        width: 260,
        height: 140,
    });
    canvas.nodes = nodes;
    await persistCanvas(surfaceElement, canvas, id, blockElement, previous);
};

const addScribliNodeFromPrompt = async (surfaceElement: HTMLElement, canvas: ICanvasPayload, id: string, blockElement: HTMLElement | undefined, kind: string) => {
    const promptLabel = {
        document: lang("canvasDocumentPrompt", "Document ID"),
        block: lang("canvasBlockPrompt", "Block ID"),
        asset: lang("canvasAssetPrompt", "Asset path"),
        query: lang("canvasQueryPrompt", "Search query"),
        database: lang("canvasDatabasePrompt", "Database ID"),
    }[kind] || kind;
    const value = await canvasInputDialog(promptLabel);
    if (!value) {
        return;
    }
    const label = await canvasInputDialog(lang("canvasLabelPrompt", "Label")) || undefined;
    const position = nextNodePosition(canvas.nodes || []);
    const args: Record<string, any> = {
        action: "add_scribli_node",
        id,
        kind,
        label,
        x: position.x,
        y: position.y,
        width: DEFAULT_NODE_WIDTH,
        height: DEFAULT_NODE_HEIGHT,
    };
    if (kind === "asset") {
        args.assetPath = value;
    } else if (kind === "query") {
        args.query = value;
    } else if (kind === "database") {
        args.databaseID = value;
    } else {
        args.refID = value;
    }
    const previous = cloneCanvas(canvas);
    const response = await fetchSyncPost("/api/canvas/call", args);
    if (response.code === 0 && !response.data?.isError) {
        pushCanvasHistory(surfaceElement, previous);
        renderCanvasSurface(surfaceElement, response.data?.structuredContent?.canvas || canvas, id, blockElement);
    }
};

const canvasInputDialog = (title: string, value = ""): Promise<string | null> => {
    return new Promise((resolve) => {
        let settled = false;
        const dialog = new Dialog({
            title,
            content: `<div class="b3-dialog__content">
    <input class="b3-text-field fn__block" data-type="canvas-prompt-input" value="${escapeAttr(value)}">
</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--cancel" data-type="canvas-prompt-cancel">${escapeHtml(lang("cancel", "Cancel"))}</button>
    <button class="b3-button b3-button--text" data-type="canvas-prompt-confirm">${escapeHtml(lang("confirm", "Confirm"))}</button>
</div>`,
            width: "420px",
            destroyCallback: () => {
                if (!settled) {
                    resolve(null);
                }
            },
        });
        const inputElement = dialog.element.querySelector('[data-type="canvas-prompt-input"]') as HTMLInputElement;
        const finish = (nextValue: string | null) => {
            settled = true;
            resolve(nextValue);
            dialog.destroy();
        };
        dialog.element.querySelector('[data-type="canvas-prompt-confirm"]')?.addEventListener("click", () => finish(inputElement.value));
        dialog.element.querySelector('[data-type="canvas-prompt-cancel"]')?.addEventListener("click", () => finish(null));
        dialog.bindInput(inputElement, () => finish(inputElement.value));
        inputElement.select();
    });
};

const appendTemplate = async (surfaceElement: HTMLElement, canvas: ICanvasPayload, id: string, blockElement: HTMLElement | undefined, templateID: string) => {
    const template = templateCanvas(templateID);
    if (!template) {
        return;
    }
    const previous = cloneCanvas(canvas);
    const offset = (canvas.nodes || []).length * 24;
    const nextNodes = (canvas.nodes || []).concat((template.nodes || []).map((node) => ({
        ...cloneNode(node),
        id: Lute.NewNodeID(),
        x: numberValue(node.x, 0) + offset,
        y: numberValue(node.y, 0) + offset,
    })));
    const idMap = new Map<string, string>();
    (template.nodes || []).forEach((node, index) => idMap.set(node.id, nextNodes[(canvas.nodes || []).length + index].id));
    const nextEdges = (canvas.edges || []).concat((template.edges || []).map((edge) => ({
        id: Lute.NewNodeID(),
        fromNode: idMap.get(edge.fromNode) || edge.fromNode,
        toNode: idMap.get(edge.toNode) || edge.toNode,
    })));
    canvas.nodes = nextNodes;
    canvas.edges = nextEdges;
    await persistCanvas(surfaceElement, canvas, id, blockElement, previous);
};

const duplicateSelectedNode = async (surfaceElement: HTMLElement, canvas: ICanvasPayload, id: string, blockElement?: HTMLElement) => {
    const state = canvasState(surfaceElement);
    const node = (canvas.nodes || []).find((item) => item.id === state.selectedNodeID);
    if (!node) {
        return;
    }
    const previous = cloneCanvas(canvas);
    const clone = cloneNode(node);
    clone.id = Lute.NewNodeID();
    clone.x = numberValue(node.x, 0) + 36;
    clone.y = numberValue(node.y, 0) + 36;
    canvas.nodes = (canvas.nodes || []).concat(clone);
    state.selectedNodeID = clone.id;
    await persistCanvas(surfaceElement, canvas, id, blockElement, previous);
};

const deleteSelectedNode = async (surfaceElement: HTMLElement, canvas: ICanvasPayload, id: string, blockElement?: HTMLElement) => {
    const state = canvasState(surfaceElement);
    const selectedID = state.selectedNodeID;
    if (!selectedID) {
        return;
    }
    const previous = cloneCanvas(canvas);
    canvas.nodes = (canvas.nodes || []).filter((node) => node.id !== selectedID);
    canvas.edges = (canvas.edges || []).filter((edge) => edge.fromNode !== selectedID && edge.toNode !== selectedID);
    state.selectedNodeID = undefined;
    state.connectFromID = undefined;
    await persistCanvas(surfaceElement, canvas, id, blockElement, previous);
};

const undoCanvas = async (surfaceElement: HTMLElement, id: string, blockElement?: HTMLElement) => {
    const state = canvasState(surfaceElement);
    const previous = state.history.pop();
    if (!previous) {
        return;
    }
    state.selectedNodeID = undefined;
    state.connectFromID = undefined;
    await persistCanvas(surfaceElement, previous, id, blockElement, undefined, false);
};

const persistCanvas = async (surfaceElement: HTMLElement, canvas: ICanvasPayload, id: string, blockElement: HTMLElement | undefined, previous?: ICanvasPayload, pushHistory = true) => {
    canvas = normalizeCanvasPayload(canvas);
    const response = await fetchSyncPost("/api/canvas/call", {action: "update", id, canvas});
    if (response.code === 0 && !response.data?.isError) {
        if (pushHistory && previous) {
            pushCanvasHistory(surfaceElement, previous);
        }
        renderCanvasSurface(surfaceElement, response.data?.structuredContent?.canvas || canvas, id, blockElement);
    }
};

const pushCanvasHistory = (surfaceElement: HTMLElement, canvas: ICanvasPayload) => {
    const state = canvasState(surfaceElement);
    state.history.push(cloneCanvas(canvas));
    if (state.history.length > 30) {
        state.history.shift();
    }
};

const wireCanvasActions = (surfaceElement: HTMLElement, canvas: ICanvasPayload, id?: string, blockElement?: HTMLElement) => {
    surfaceElement.querySelectorAll(".scribli-canvas__node").forEach((nodeElement: HTMLElement) => {
        nodeElement.addEventListener("click", async (event) => {
            const target = event.target as HTMLElement;
            if (isCanvasNodeControlTarget(target, nodeElement)) {
                return;
            }
            event.preventDefault();
            event.stopPropagation();
            const nodeID = nodeElement.dataset.nodeId || "";
            const state = canvasState(surfaceElement);
            if (state.connectFromID !== undefined && id) {
                await connectNode(surfaceElement, canvas, id, blockElement, nodeID);
                return;
            }
            state.selectedNodeID = state.selectedNodeID === nodeID ? undefined : nodeID;
            renderCanvasSurface(surfaceElement, canvas, id, blockElement);
        });
    });
    surfaceElement.querySelectorAll('[data-type="canvas-open-node"]').forEach((buttonElement: HTMLElement) => {
        buttonElement.addEventListener("pointerdown", (event) => {
            event.stopPropagation();
        });
        buttonElement.addEventListener("click", async (event) => {
            event.preventDefault();
            event.stopPropagation();
            await openCanvasNode(buttonElement);
        });
    });
    surfaceElement.querySelectorAll('[data-type="canvas-node-delete"]').forEach((buttonElement: HTMLElement) => {
        buttonElement.addEventListener("click", async (event) => {
            event.preventDefault();
            event.stopPropagation();
            const state = canvasState(surfaceElement);
            state.selectedNodeID = (buttonElement.closest(".scribli-canvas__node") as HTMLElement)?.dataset.nodeId;
            if (id) {
                await deleteSelectedNode(surfaceElement, canvas, id, blockElement);
            }
        });
    });
    surfaceElement.querySelectorAll('[data-type="canvas-node-duplicate"]').forEach((buttonElement: HTMLElement) => {
        buttonElement.addEventListener("click", async (event) => {
            event.preventDefault();
            event.stopPropagation();
            const state = canvasState(surfaceElement);
            state.selectedNodeID = (buttonElement.closest(".scribli-canvas__node") as HTMLElement)?.dataset.nodeId;
            if (id) {
                await duplicateSelectedNode(surfaceElement, canvas, id, blockElement);
            }
        });
    });
};

const connectNode = async (surfaceElement: HTMLElement, canvas: ICanvasPayload, id: string, blockElement: HTMLElement | undefined, nodeID: string) => {
    const state = canvasState(surfaceElement);
    if (!state.connectFromID) {
        state.connectFromID = nodeID;
        state.selectedNodeID = nodeID;
        renderCanvasSurface(surfaceElement, canvas, id, blockElement);
        return;
    }
    if (state.connectFromID === nodeID) {
        state.connectFromID = undefined;
        renderCanvasSurface(surfaceElement, canvas, id, blockElement);
        return;
    }
    const previous = cloneCanvas(canvas);
    canvas.edges = (canvas.edges || []).concat({
        id: Lute.NewNodeID(),
        fromNode: state.connectFromID,
        toNode: nodeID,
    });
    state.connectFromID = undefined;
    await persistCanvas(surfaceElement, canvas, id, blockElement, previous);
};

const openCanvasNode = async (buttonElement: HTMLElement) => {
    const nodeElement = buttonElement.closest(".scribli-canvas__node") as HTMLElement;
    const app = window.scribli.ws?.app;
    if (!nodeElement || !app) {
        return;
    }
    const kind = nodeElement.dataset.kind || "";
    const refID = nodeElement.dataset.refId || "";
    if (["document", "block", "executable_output"].includes(kind) && refID) {
        await openFileById({app, id: refID, action: [Constants.CB_GET_FOCUS]});
        return;
    }
    if (kind === "asset" && nodeElement.dataset.assetPath) {
        openAsset(app, nodeElement.dataset.assetPath, 1, "right");
        return;
    }
    if (kind === "query" && nodeElement.dataset.query) {
        await openSearch({app, hotkey: Constants.DIALOG_GLOBALSEARCH, key: nodeElement.dataset.query});
        return;
    }
    if (kind === "database") {
        if (refID) {
            await openFileById({app, id: refID, action: [Constants.CB_GET_FOCUS]});
        } else if (nodeElement.dataset.databaseId) {
            await openSearch({app, hotkey: Constants.DIALOG_GLOBALSEARCH, key: nodeElement.dataset.databaseId});
        }
    }
};

const wireCanvasEditing = (surfaceElement: HTMLElement, canvas: ICanvasPayload, id?: string, blockElement?: HTMLElement) => {
    surfaceElement.querySelectorAll('[data-type="canvas-edit-text"]').forEach((editElement: HTMLElement) => {
        editElement.addEventListener("pointerdown", (event) => {
            event.stopPropagation();
        });
        editElement.addEventListener("blur", async () => {
            if (!id) {
                return;
            }
            const nodeID = (editElement.closest(".scribli-canvas__node") as HTMLElement)?.dataset.nodeId;
            const node = (canvas.nodes || []).find((item) => item.id === nodeID);
            if (!node) {
                return;
            }
            const nextText = editElement.innerText.trim();
            if (nextText === (node.text || "")) {
                return;
            }
            const previous = cloneCanvas(canvas);
            node.text = nextText;
            if (node.scribli?.kind === "text") {
                node.scribli.label = firstLine(nextText);
            }
            await persistCanvas(surfaceElement, canvas, id, blockElement, previous);
        });
    });
};

const wireCanvasDragging = (surfaceElement: HTMLElement, canvas: ICanvasPayload, id?: string, blockElement?: HTMLElement) => {
    let active: { element: HTMLElement, node: ICanvasNode, startX: number, startY: number, x: number, y: number, previous: ICanvasPayload } | undefined;
    const handleDragMove = (event: PointerEvent) => {
        if (!active) {
            return;
        }
        const scale = canvasState(surfaceElement).scale;
        const x = active.x + (event.clientX - active.startX) / scale;
        const y = active.y + (event.clientY - active.startY) / scale;
        active.node.x = x;
        active.node.y = y;
        active.element.style.left = `${x}px`;
        active.element.style.top = `${y}px`;
    };
    const handleDragUp = async () => {
        if (!active) {
            return;
        }
        const previous = active.previous;
        active = undefined;
        document.removeEventListener("pointermove", handleDragMove, true);
        document.removeEventListener("pointerup", handleDragUp, true);
        if (id) {
            await persistCanvas(surfaceElement, canvas, id, blockElement, previous);
        }
    };
    surfaceElement.querySelectorAll(".scribli-canvas__node").forEach((nodeElement: HTMLElement) => {
        nodeElement.addEventListener("pointerdown", (event: PointerEvent) => {
            const target = event.target as HTMLElement;
            if (isCanvasNodeControlTarget(target, nodeElement) || target.closest(".scribli-canvas__resize")) {
                return;
            }
            const node = (canvas.nodes || []).find((item) => item.id === nodeElement.dataset.nodeId);
            if (!node) {
                return;
            }
            event.preventDefault();
            event.stopPropagation();
            nodeElement.setPointerCapture(event.pointerId);
            active = {
                element: nodeElement,
                node,
                startX: event.clientX,
                startY: event.clientY,
                x: numberValue(node.x, 0),
                y: numberValue(node.y, 0),
                previous: cloneCanvas(canvas),
            };
            canvasState(surfaceElement).selectedNodeID = node.id;
            document.addEventListener("pointermove", handleDragMove, true);
            document.addEventListener("pointerup", handleDragUp, true);
        });
    });
};

const wireCanvasResizing = (surfaceElement: HTMLElement, canvas: ICanvasPayload, id?: string, blockElement?: HTMLElement) => {
    let active: { element: HTMLElement, node: ICanvasNode, startX: number, startY: number, width: number, height: number, previous: ICanvasPayload } | undefined;
    const handleResizeMove = (event: PointerEvent) => {
        if (!active) {
            return;
        }
        const scale = canvasState(surfaceElement).scale;
        const width = Math.max(MIN_NODE_WIDTH, active.width + (event.clientX - active.startX) / scale);
        const height = Math.max(MIN_NODE_HEIGHT, active.height + (event.clientY - active.startY) / scale);
        active.node.width = width;
        active.node.height = height;
        const nodeElement = active.element.closest(".scribli-canvas__node") as HTMLElement;
        nodeElement.style.width = `${width}px`;
        nodeElement.style.height = `${height}px`;
    };
    const handleResizeUp = async () => {
        if (!active) {
            return;
        }
        const previous = active.previous;
        active = undefined;
        document.removeEventListener("pointermove", handleResizeMove, true);
        document.removeEventListener("pointerup", handleResizeUp, true);
        if (id) {
            await persistCanvas(surfaceElement, canvas, id, blockElement, previous);
        }
    };
    surfaceElement.querySelectorAll(".scribli-canvas__resize").forEach((resizeElement: HTMLElement) => {
        resizeElement.addEventListener("pointerdown", (event: PointerEvent) => {
            event.preventDefault();
            event.stopPropagation();
            const nodeElement = resizeElement.closest(".scribli-canvas__node") as HTMLElement;
            const node = (canvas.nodes || []).find((item) => item.id === nodeElement?.dataset.nodeId);
            if (!node) {
                return;
            }
            resizeElement.setPointerCapture(event.pointerId);
            active = {
                element: resizeElement,
                node,
                startX: event.clientX,
                startY: event.clientY,
                width: numberValue(node.width, DEFAULT_NODE_WIDTH),
                height: numberValue(node.height, DEFAULT_NODE_HEIGHT),
                previous: cloneCanvas(canvas),
            };
            document.addEventListener("pointermove", handleResizeMove, true);
            document.addEventListener("pointerup", handleResizeUp, true);
        });
        resizeElement.addEventListener("pointermove", (event: PointerEvent) => {
            if (!active || active.element !== resizeElement) {
                return;
            }
            const scale = canvasState(surfaceElement).scale;
            const width = Math.max(MIN_NODE_WIDTH, active.width + (event.clientX - active.startX) / scale);
            const height = Math.max(MIN_NODE_HEIGHT, active.height + (event.clientY - active.startY) / scale);
            active.node.width = width;
            active.node.height = height;
            const nodeElement = resizeElement.closest(".scribli-canvas__node") as HTMLElement;
            nodeElement.style.width = `${width}px`;
            nodeElement.style.height = `${height}px`;
        });
        resizeElement.addEventListener("pointerup", handleResizeUp);
    });
};

const wireCanvasPanning = (surfaceElement: HTMLElement, canvas: ICanvasPayload, id?: string, blockElement?: HTMLElement) => {
    const viewport = surfaceElement.querySelector('[data-type="canvas-viewport"]') as HTMLElement;
    if (!viewport) {
        return;
    }
    let active: { startX: number, startY: number, panX: number, panY: number } | undefined;
    viewport.addEventListener("pointerdown", (event: PointerEvent) => {
        if ((event.target as HTMLElement).closest(".scribli-canvas__node, .scribli-canvas__toolbar, .scribli-canvas__panel, .scribli-canvas__empty")) {
            return;
        }
        active = {
            startX: event.clientX,
            startY: event.clientY,
            panX: canvasState(surfaceElement).panX,
            panY: canvasState(surfaceElement).panY,
        };
        viewport.setPointerCapture(event.pointerId);
    });
    viewport.addEventListener("pointermove", (event: PointerEvent) => {
        if (!active) {
            return;
        }
        const state = canvasState(surfaceElement);
        state.panX = active.panX + event.clientX - active.startX;
        state.panY = active.panY + event.clientY - active.startY;
        const stage = surfaceElement.querySelector(".scribli-canvas__stage") as HTMLElement;
        if (stage) {
            stage.style.transform = `translate(${state.panX}px, ${state.panY}px) scale(${state.scale})`;
        }
    });
    viewport.addEventListener("pointerup", () => {
        active = undefined;
        renderCanvasSurface(surfaceElement, canvas, id, blockElement);
    });
};

const renderCanvasLibrary = async (surfaceElement: HTMLElement, blockElement?: HTMLElement) => {
    const bodyElement = surfaceElement.querySelector('.scribli-canvas__panel[data-panel="library"] .scribli-canvas__panel-body');
    if (!bodyElement) {
        return;
    }
    const response = await fetchSyncPost("/api/canvas/call", {action: "list"});
    if (response.code !== 0 || response.data?.isError) {
        bodyElement.innerHTML = `<div class="scribli-canvas__status ft__error">${escapeHtml(response.msg || lang("canvasLoadFailed", "Canvas load failed"))}</div>`;
        return;
    }
    const items = [...(response.data?.structuredContent?.items || [])].sort((left: Record<string, any>, right: Record<string, any>) => {
        const leftKey = String(left.updated || left.id || "");
        const rightKey = String(right.updated || right.id || "");
        return rightKey.localeCompare(leftKey);
    });
    if (items.length === 0) {
        bodyElement.innerHTML = `<div class="scribli-canvas__status">${escapeHtml(lang("canvasLibraryEmpty", "No canvases found"))}</div>`;
        return;
    }
    bodyElement.innerHTML = items.map((item: Record<string, any>) => `<button class="b3-list-item b3-list-item--two" data-type="canvas-open-library" data-id="${escapeAttr(String(item.id || ""))}">
    <div class="b3-list-item__first"><svg class="b3-list-item__graphic"><use xlink:href="#iconBoth"></use></svg><span class="b3-list-item__text">${escapeHtml(String(item.title || item.id || ""))}</span></div>
    <span class="b3-list-item__meta">${escapeHtml(`${item.nodes || 0}/${item.edges || 0}`)}</span>
</button>`).join("");
};

const openCanvasFromLibrary = async (surfaceElement: HTMLElement, id: string) => {
    const currentBlock = currentCanvasBlock(surfaceElement);
    if (!id || !currentBlock) {
        return;
    }
    const getResponse = await fetchSyncPost("/api/canvas/call", {action: "get", id});
    if (getResponse.code !== 0 || getResponse.data?.isError) {
        const bodyElement = surfaceElement.querySelector('.scribli-canvas__panel[data-panel="library"] .scribli-canvas__panel-body');
        if (bodyElement) {
            bodyElement.innerHTML = `<div class="scribli-canvas__status ft__error">${escapeHtml(getResponse.msg || lang("canvasLoadFailed", "Canvas load failed"))}</div>`;
        }
        return;
    }
    const boundBlock = await bindCanvasToBlock(currentBlock, id);
    const state = canvasState(surfaceElement);
    state.panel = undefined;
    const currentSurface = currentCanvasSurface(surfaceElement, boundBlock);
    renderCanvasSurface(currentSurface, getResponse.data?.structuredContent?.canvas || {}, id, boundBlock);
};

const importCanvasFile = async (surfaceElement: HTMLElement, blockElement: HTMLElement | undefined, inputElement: HTMLInputElement) => {
    const file = inputElement.files?.[0];
    inputElement.value = "";
    if (!file || !blockElement) {
        return;
    }
    const text = await file.text();
    const canvas = normalizeCanvasPayload(JSON.parse(text));
    if (!canvas.scribli) {
        canvas.scribli = {};
    }
    if (!canvas.scribli.title) {
        canvas.scribli.title = file.name.replace(/\.(canvas|json)$/i, "");
    }
    await createCanvasFromPlaceholder(surfaceElement, blockElement, canvas);
};

const exportCanvas = (id: string, canvas: ICanvasPayload) => {
    const blob = new Blob([JSON.stringify(normalizeCanvasPayload(canvas), undefined, 2) + "\n"], {type: "application/json"});
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = `${id || "scribli-canvas"}.canvas`;
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    URL.revokeObjectURL(url);
};

const nodeHTML = (node: ICanvasNode, state: ICanvasViewState) => {
    const kind = node.scribli?.kind || node.type || "text";
    const title = node.scribli?.label || firstLine(node.text) || node.file || node.id;
    const body = nodeBody(node);
    const selectedClass = state.selectedNodeID === node.id ? " scribli-canvas__node--selected" : "";
    const connectClass = state.connectFromID === node.id ? " scribli-canvas__node--connecting" : "";
    return `<div class="scribli-canvas__node${selectedClass}${connectClass}" data-node-id="${escapeAttr(node.id)}" data-kind="${escapeAttr(kind)}" data-ref-id="${escapeAttr(node.scribli?.refID || "")}" data-query="${escapeAttr(node.scribli?.query || "")}" data-asset-path="${escapeAttr(node.file || node.scribli?.assetPath || "")}" data-database-id="${escapeAttr(node.scribli?.databaseID || "")}" style="left:${numberValue(node.x, 0)}px;top:${numberValue(node.y, 0)}px;width:${numberValue(node.width, DEFAULT_NODE_WIDTH)}px;height:${numberValue(node.height, DEFAULT_NODE_HEIGHT)}px">
    <div class="scribli-canvas__node-title"><span>${escapeHtml(title)}</span>${nodeActionHTML(node, kind)}${iconButton("canvas-node-duplicate", "iconCopy", lang("duplicate", "Duplicate"))}${iconButton("canvas-node-delete", "iconTrashcan", lang("delete", "Delete"))}</div>
    <div class="scribli-canvas__node-body">${body}</div>
    <div class="scribli-canvas__resize"></div>
</div>`;
};

const nodeBody = (node: ICanvasNode) => {
    const kind = node.scribli?.kind || node.type || "text";
    const assetPath = node.file || node.scribli?.assetPath || "";
    if (kind === "asset" && assetPath) {
        const ext = pathPosix().extname(assetPath).toLowerCase();
        if (Constants.SCRIBLI_ASSETS_IMAGE.includes(ext)) {
            return `<img src="${escapeAttr(assetPath)}" alt="${escapeAttr(node.scribli?.label || assetPath)}"><code>${escapeHtml(assetPath)}</code>`;
        }
        return `<code>${escapeHtml(assetPath)}</code>`;
    }
    if (node.scribli?.query) {
        return `<div class="scribli-canvas__node-kicker">${escapeHtml(lang("search", "Search"))}</div><code>${escapeHtml(node.scribli.query)}</code>`;
    }
    if (node.scribli?.databaseID) {
        return `<div class="scribli-canvas__node-kicker">${escapeHtml(lang("database", "Database"))}</div><code>${escapeHtml(node.scribli.databaseID)}${node.scribli.viewID ? " / " + escapeHtml(node.scribli.viewID) : ""}</code>`;
    }
    if (kind === "text" || node.type === "text") {
        return `<div data-type="canvas-edit-text" contenteditable="true" spellcheck="${window.scribli.config.editor.spellcheck}">${escapeHtml(node.text || "")}</div>`;
    }
    return escapeHtml(node.text || node.scribli?.refID || "");
};

const nodeActionHTML = (node: ICanvasNode, kind: string) => {
    if (!isActionableNode(node, kind)) {
        return "";
    }
    return iconButton("canvas-open-node", "iconOpenWindow", nodeActionLabel(kind));
};

const nodeActionLabel = (kind: string) => {
    if (kind === "asset") {
        return lang("assets", "Assets");
    }
    if (kind === "query") {
        return lang("search", "Search");
    }
    if (kind === "database") {
        return lang("database", "Database");
    }
    return lang("openDocument", "Open document");
};

const isActionableNode = (node: ICanvasNode, kind: string) => {
    if (["document", "block", "executable_output"].includes(kind)) {
        return Boolean(node.scribli?.refID);
    }
    if (kind === "asset") {
        return Boolean(node.file || node.scribli?.assetPath);
    }
    if (kind === "query") {
        return Boolean(node.scribli?.query);
    }
    if (kind === "database") {
        return Boolean(node.scribli?.refID || node.scribli?.databaseID);
    }
    return false;
};

const edgeHTML = (edges: ICanvasEdge[], nodes: ICanvasNode[]) => {
    const nodeMap = new Map(nodes.map((node) => [node.id, node]));
    return edges.map((edge) => {
        const fromNode = nodeMap.get(edge.fromNode);
        const toNode = nodeMap.get(edge.toNode);
        if (!fromNode || !toNode) {
            return "";
        }
        const x1 = numberValue(fromNode.x, 0) + numberValue(fromNode.width, DEFAULT_NODE_WIDTH) / 2;
        const y1 = numberValue(fromNode.y, 0) + numberValue(fromNode.height, DEFAULT_NODE_HEIGHT) / 2;
        const x2 = numberValue(toNode.x, 0) + numberValue(toNode.width, DEFAULT_NODE_WIDTH) / 2;
        const y2 = numberValue(toNode.y, 0) + numberValue(toNode.height, DEFAULT_NODE_HEIGHT) / 2;
        return `<line x1="${x1}" y1="${y1}" x2="${x2}" y2="${y2}"></line>`;
    }).join("");
};

const normalizeCanvasPayload = (canvas: ICanvasPayload = {}) => {
    if (!Array.isArray(canvas.nodes)) {
        canvas.nodes = [];
    }
    if (!Array.isArray(canvas.edges)) {
        canvas.edges = [];
    }
    return canvas;
};

const canvasBounds = (nodes: ICanvasNode[]) => {
    if (nodes.length === 0) {
        return {minX: 0, minY: 0, width: 1100, height: 560};
    }
    const minX = Math.min(...nodes.map((node) => numberValue(node.x, 0))) - 120;
    const minY = Math.min(...nodes.map((node) => numberValue(node.y, 0))) - 120;
    const maxX = Math.max(...nodes.map((node) => numberValue(node.x, 0) + numberValue(node.width, DEFAULT_NODE_WIDTH))) + 120;
    const maxY = Math.max(...nodes.map((node) => numberValue(node.y, 0) + numberValue(node.height, DEFAULT_NODE_HEIGHT))) + 120;
    return {minX, minY, width: Math.max(1100, maxX - minX), height: Math.max(560, maxY - minY)};
};

const nextNodePosition = (nodes: ICanvasNode[]) => {
    const offset = nodes.length * 36;
    return {x: offset, y: offset};
};

const canvasState = (surfaceElement: HTMLElement) => {
    let state = canvasStates.get(surfaceElement);
    if (!state) {
        state = {scale: 1, panX: 0, panY: 0, history: []};
        canvasStates.set(surfaceElement, state);
    }
    return state;
};

const templateCanvas = (id: string): ICanvasPayload | undefined => {
    const template = CANVAS_TEMPLATES.find((item) => item.id === id);
    return template ? cloneCanvas(template.canvas) : undefined;
};

const cloneCanvas = (canvas: ICanvasPayload): ICanvasPayload => JSON.parse(JSON.stringify(normalizeCanvasPayload(canvas)));

const cloneNode = (node: ICanvasNode): ICanvasNode => JSON.parse(JSON.stringify(node));

const numberValue = (value: unknown, fallback: number) => typeof value === "number" && Number.isFinite(value) ? value : fallback;

const firstLine = (value?: string) => (value || "").split(/\r?\n/)[0].trim();

const clampScale = (value: number) => Math.max(0.4, Math.min(2.5, Number(value.toFixed(2))));

const lang = (key: string, fallback: string) => window.scribli.languages[key] || fallback;

const CANVAS_TEMPLATES: Array<{ id: string, title: string, canvas: ICanvasPayload }> = [{
    id: "blank",
    title: "Blank",
    canvas: {nodes: [], edges: [], scribli: {title: "Blank Canvas"}},
}, {
    id: "research",
    title: "Research",
    canvas: {
        scribli: {title: "Research Board"},
        nodes: [
            {id: "question", type: "text", text: "Question", x: 0, y: 0, width: 260, height: 120},
            {id: "sources", type: "text", text: "Sources", x: 340, y: 0, width: 260, height: 120},
            {id: "findings", type: "text", text: "Findings", x: 170, y: 210, width: 280, height: 140},
        ],
        edges: [
            {id: "question-sources", fromNode: "question", toNode: "sources"},
            {id: "sources-findings", fromNode: "sources", toNode: "findings"},
        ],
    },
}, {
    id: "project",
    title: "Project",
    canvas: {
        scribli: {title: "Project Board"},
        nodes: [
            {id: "goal", type: "text", text: "Goal", x: 0, y: 0, width: 260, height: 120},
            {id: "work", type: "text", text: "Work", x: 340, y: 0, width: 260, height: 120},
            {id: "risks", type: "text", text: "Risks", x: 0, y: 210, width: 260, height: 120},
            {id: "done", type: "text", text: "Done", x: 340, y: 210, width: 260, height: 120},
        ],
        edges: [
            {id: "goal-work", fromNode: "goal", toNode: "work"},
            {id: "work-done", fromNode: "work", toNode: "done"},
        ],
    },
}, {
    id: "character",
    title: "Character",
    canvas: {
        scribli: {title: "Character Board"},
        nodes: [
            {id: "person", type: "text", text: "Character", x: 180, y: 0, width: 280, height: 130},
            {id: "wants", type: "text", text: "Wants", x: 0, y: 220, width: 240, height: 120},
            {id: "fears", type: "text", text: "Fears", x: 400, y: 220, width: 240, height: 120},
        ],
        edges: [
            {id: "person-wants", fromNode: "person", toNode: "wants"},
            {id: "person-fears", fromNode: "person", toNode: "fears"},
        ],
    },
}];

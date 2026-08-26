import {Constants} from "../../constants";
import {escapeAriaLabel, escapeAttr, escapeHtml} from "../../util/escape";
import {fetchPost} from "../../util/fetch";
import {unicode2Emoji} from "../../emoji";

const CHILD_DOCS_CLASS = "protyle-child-docs";
const CHILD_DOCS_LIMIT = 256;
const CUSTOM_SORT_MODE = 6;
const FILE_TREE_SORT_MODE = 15;
const CHILD_DOC_DROP_TYPE = "application/x-scribli-child-doc";

interface IFiletreeSortData {
    parentPath?: string;
    childIDs?: string[];
}

const getChildDocsElement = (protyle: IProtyle) => {
    let element = protyle.contentElement.querySelector<HTMLElement>(`:scope > .${CHILD_DOCS_CLASS}`);
    if (!element) {
        element = document.createElement("div");
        element.className = `${CHILD_DOCS_CLASS} fn__none`;
        element.setAttribute("contenteditable", "false");
        protyle.wysiwyg.element.insertAdjacentElement("afterend", element);
        element.addEventListener("click", (event) => {
            if (element.dataset.suppressClick === "true") {
                event.preventDefault();
                event.stopPropagation();
                return;
            }
            const eventTarget = event.target instanceof Element ? event.target : null;
            const target = eventTarget?.closest<HTMLElement>("[data-child-doc-id]");
            if (!target) {
                return;
            }
            event.preventDefault();
            event.stopPropagation();
            window.openFileByURL(`scribli://blocks/${target.dataset.childDocId}`);
        });
        bindChildDocsDrag(protyle, element);
    }
    return element;
};

export const clearChildDocs = (protyle: IProtyle) => {
    const element = protyle.contentElement?.querySelector<HTMLElement>(`:scope > .${CHILD_DOCS_CLASS}`);
    if (element) {
        element.innerHTML = "";
        element.classList.add("fn__none");
        element.removeAttribute("data-root-id");
        element.removeAttribute("data-path");
        element.removeAttribute("data-sortable");
    }
};

export const renderChildDocs = (protyle: IProtyle) => {
    const isDocument = protyle.wysiwyg.element.getAttribute("data-doc-type") === "NodeDocument";
    if (!isDocument || !protyle.block?.rootID || protyle.block.id !== protyle.block.rootID || !protyle.path) {
        clearChildDocs(protyle);
        return;
    }

    const rootID = protyle.block.rootID;
    const notebook = protyle.notebookId;
    const listPath = notebook === rootID ? "/" : protyle.path;
    const element = getChildDocsElement(protyle);
    element.dataset.rootId = rootID;
    element.dataset.path = listPath;
    element.dataset.notebook = notebook;

    fetchPost("/api/filetree/listDocsByPath", {
        notebook,
        path: listPath,
        maxListCount: CHILD_DOCS_LIMIT,
        ignoreMaxListHint: true,
        app: Constants.SCRIBLI_APPID,
    }, response => {
        if (element.dataset.rootId !== rootID || element.dataset.path !== listPath) {
            return;
        }
        if (response.code !== 0 || !response.data?.files?.length) {
            clearChildDocs(protyle);
            return;
        }

        const files = response.data.files as IFile[];
        const title = window.scribli.languages.childPages || "Child pages";
        const count = files.length.toString();
        const sortable = isChildDocsSortable(notebook);
        element.dataset.sortable = sortable.toString();
        const itemsHTML = files.map((item) => {
            const name = getDocDisplayNameForChild(item);
            const icon = unicode2Emoji(item.icon || (item.subFileCount > 0 ? window.scribli.storage[Constants.LOCAL_IMAGES].folder : window.scribli.storage[Constants.LOCAL_IMAGES].file));
            const subCount = item.subFileCount > 0 ? `<span class="protyle-child-docs__count">${item.subFileCount}</span>` : "";
            const ariaLabel = `${window.scribli.languages.openDocument}: ${name}`;
            return `<button type="button" class="protyle-child-docs__item ariaLabel" data-position="north" data-child-doc-id="${escapeAttr(item.id)}" data-child-doc-path="${escapeAttr(item.path)}" aria-label="${escapeAriaLabel(ariaLabel)}"${sortable ? ' draggable="true"' : ""}>
    <span class="protyle-child-docs__drag${sortable ? "" : " fn__hidden"}"><svg><use xlink:href="#iconDrag"></use></svg></span>
    <span class="protyle-child-docs__icon">${icon}</span>
    <span class="protyle-child-docs__name">${escapeHtml(name)}</span>
    ${subCount}
</button>`;
        }).join("");

        element.innerHTML = `<div class="protyle-child-docs__header">
    <svg><use xlink:href="#iconFolder"></use></svg>
    <span>${escapeHtml(title)}</span>
    <span class="protyle-child-docs__total">${count}</span>
</div>
<div class="protyle-child-docs__list">${itemsHTML}</div>`;
        element.classList.remove("fn__none");
    });
};

export const refreshChildDocsByFiletreeSort = (protyle: IProtyle, data: IFiletreeSortData) => {
    const element = protyle.contentElement?.querySelector<HTMLElement>(`:scope > .${CHILD_DOCS_CLASS}`);
    if (!element || element.classList.contains("fn__none")) {
        return;
    }

    const matchesPath = data.parentPath && element.dataset.path === normalizeFiletreeParentPath(data.parentPath);
    const matchesChild = data.childIDs?.some((id) => Boolean(element.querySelector(`[data-child-doc-id="${id}"]`)));
    if (matchesPath || matchesChild) {
        renderChildDocs(protyle);
    }
};

const getDocDisplayNameForChild = (item: IFile) => {
    if (item.titleEmpty) {
        return window.scribli.languages.untitled || "Untitled";
    }
    return item.name || window.scribli.languages.untitled || "Untitled";
};

const isChildDocsSortable = (notebook: string) => {
    if (window.scribli.config.readonly) {
        return false;
    }
    const notebookSortMode = window.scribli.notebooks?.find((item) => item.id === notebook)?.sortMode;
    return notebookSortMode === CUSTOM_SORT_MODE ||
        (window.scribli.config.fileTree.sort === CUSTOM_SORT_MODE && notebookSortMode === FILE_TREE_SORT_MODE);
};

const bindChildDocsDrag = (protyle: IProtyle, element: HTMLElement) => {
    element.addEventListener("dragstart", (event) => {
        const target = getChildDocItem(event);
        if (!target || element.dataset.sortable !== "true" || !event.dataTransfer) {
            event.preventDefault();
            return;
        }
        element.dataset.dragPath = target.dataset.childDocPath;
        target.classList.add("protyle-child-docs__item--dragging");
        event.dataTransfer.effectAllowed = "move";
        event.dataTransfer.setData(CHILD_DOC_DROP_TYPE, target.dataset.childDocPath);
        event.stopPropagation();
    });

    element.addEventListener("dragover", (event) => {
        const target = getChildDocItem(event);
        if (!target || element.dataset.sortable !== "true" || !element.dataset.dragPath ||
            target.dataset.childDocPath === element.dataset.dragPath || !event.dataTransfer) {
            clearChildDocsDragOver(element);
            if (element.dataset.dragPath) {
                event.preventDefault();
                event.stopPropagation();
            }
            return;
        }
        const rect = target.getBoundingClientRect();
        target.classList.toggle("dragover__top", event.clientY < rect.top + rect.height / 2);
        target.classList.toggle("dragover__bottom", event.clientY >= rect.top + rect.height / 2);
        element.querySelectorAll<HTMLElement>(".protyle-child-docs__item").forEach((item) => {
            if (item !== target) {
                item.classList.remove("dragover__top", "dragover__bottom");
            }
        });
        event.dataTransfer.dropEffect = "move";
        event.preventDefault();
        event.stopPropagation();
    });

    element.addEventListener("dragleave", (event) => {
        const relatedTarget = event.relatedTarget;
        if (!(relatedTarget instanceof Node) || !element.contains(relatedTarget)) {
            clearChildDocsDragOver(element);
        }
    });

    element.addEventListener("drop", (event) => {
        const target = getChildDocItem(event);
        const sourcePath = element.dataset.dragPath || event.dataTransfer?.getData(CHILD_DOC_DROP_TYPE);
        if (sourcePath) {
            event.preventDefault();
            event.stopPropagation();
        }
        if (!target || !sourcePath || target.dataset.childDocPath === sourcePath) {
            clearChildDocsDragState(element);
            return;
        }
        reorderChildDocs(protyle, element, sourcePath, target);
        clearChildDocsDragState(element);
    });

    element.addEventListener("dragend", () => {
        element.dataset.suppressClick = "true";
        clearChildDocsDragState(element);
        window.setTimeout(() => {
            delete element.dataset.suppressClick;
        });
    });
};

const reorderChildDocs = (protyle: IProtyle, element: HTMLElement, sourcePath: string, target: HTMLElement) => {
    if (!element.dataset.notebook) {
        return;
    }
    const source = Array.from(element.querySelectorAll<HTMLElement>(".protyle-child-docs__item")).find((item) => {
        return item.dataset.childDocPath === sourcePath;
    });
    if (!source) {
        return;
    }

    if (target.classList.contains("dragover__top")) {
        target.before(source);
    } else {
        target.after(source);
    }

    const paths: string[] = [];
    element.querySelectorAll<HTMLElement>(".protyle-child-docs__item").forEach((item) => {
        if (item.dataset.childDocPath) {
            paths.push(item.dataset.childDocPath);
        }
    });
    fetchPost("/api/filetree/changeSort", {
        notebook: element.dataset.notebook,
        paths,
    }, () => {
        renderChildDocs(protyle);
    });
};

const getChildDocItem = (event: DragEvent) => {
    return event.target instanceof Element ? event.target.closest<HTMLElement>("[data-child-doc-id]") : undefined;
};

const clearChildDocsDragOver = (element: HTMLElement) => {
    element.querySelectorAll<HTMLElement>(".dragover__top, .dragover__bottom").forEach((item) => {
        item.classList.remove("dragover__top", "dragover__bottom");
    });
};

const clearChildDocsDragState = (element: HTMLElement) => {
    delete element.dataset.dragPath;
    clearChildDocsDragOver(element);
    element.querySelectorAll<HTMLElement>(".protyle-child-docs__item--dragging").forEach((item) => {
        item.classList.remove("protyle-child-docs__item--dragging");
    });
};

const normalizeFiletreeParentPath = (parentPath: string) => {
    if (parentPath === "/") {
        return "/";
    }
    return `${parentPath}.sy`;
};

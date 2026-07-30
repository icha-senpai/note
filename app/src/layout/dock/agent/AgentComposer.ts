import {Protyle} from "../../../protyle";
import {Constants} from "../../../constants";
import {App} from "../../../index";
import {escapeHtml} from "../../../util/escape";
import {fetchPost} from "../../../util/fetch";
import {hintRef} from "../../../protyle/hint/extend";
import {genEmptyElement} from "../../../block/util";
import {blockRender} from "../../../protyle/render/blockRender";

interface ComposerHandle {
    focus: () => void;
    destroy: () => void;
    getSendData: () => { text: string; blockHTML: string; references: { id: string; title: string }[] };
    clear: () => void;
    pushHistory: (text: string) => void;
    getHistory: () => string[];
    clearHistory: () => void;
    restoreHistory: (h: string[]) => void;
    insertMention: (id: string, label: string) => void;
    insertMentions: (mentions: Array<{ id: string; label: string }>) => void;
    renderBlockHTML: (element: HTMLElement, onEmbedRender: () => void) => void;
}

type OnChangeCallback = () => void;

const resetEmbedBlocks = (element: HTMLElement) => {
    element.querySelectorAll<HTMLElement>('[data-type="NodeBlockQueryEmbed"]').forEach((embedElement) => {
        embedElement.removeAttribute("data-render");
        embedElement.style.height = "";
        Array.from(embedElement.children).forEach((child) => {
            if (child.classList.contains("protyle-wysiwyg__embed")) {
                child.remove();
            }
        });
    });
};

const hintSkill = (key: string, protyle: IProtyle): IHintData[] => {
    protyle.hint.genLoading(protyle);
    fetchPost("/api/ai/agent/lsSkills", {}, (response) => {
        const rawSkills = (response && response.data) ? response.data : [];
        const q = key.toLowerCase();
        const dataList: IHintData[] = rawSkills
            .filter((s: Record<string, string>) => !q ||
                (s.name || "").toLowerCase().includes(q) || (s.description || "").toLowerCase().includes(q))
            .map((s: Record<string, string>) => ({
                value: s.name + " ",
                html: '<div class="b3-list-item__first"><span class="b3-list-item__text">' +
                    escapeHtml(s.name) + "</span></div>" +
                    (s.description ? '<div class="b3-list-item__meta b3-list-item__showall">' + escapeHtml(s.description) + "</div>" : ""),
            }));
        if (dataList.length === 0) {
            dataList.push({value: "", html: window.scribli.languages.emptyContent});
        }
        protyle.hint.genHTML(dataList, protyle, false, "hint");
    });
    return [];
};

class ComposerHistory {
    private items: string[] = [];
    private idx = -1;
    private savedDraft = "";

    push(text: string) {
        if (!text || this.items[this.items.length - 1] === text) {
            return;
        }
        this.items.push(text);
        if (this.items.length > 50) {
            this.items.shift();
        }
        this.idx = -1;
    }

    get(): string[] {
        return this.items.slice();
    }

    clear() {
        this.items = [];
        this.idx = -1;
    }

    restore(h: string[]) {
        this.items = [];
        this.items.push(...h);
        this.idx = -1;
    }

    has(): boolean {
        return this.items.length > 0;
    }

    isBrowsing(): boolean {
        return this.idx !== -1;
    }

    resetCursor() {
        this.idx = -1;
    }

    beginBrowsing(currentDraft: string): string {
        this.savedDraft = currentDraft;
        this.idx = this.items.length - 1;
        return this.items[this.idx];
    }

    navigateUp(): string {
        if (this.idx > 0) {
            this.idx--;
        }
        return this.items[this.idx];
    }

    navigateDown(): string {
        this.idx++;
        if (this.idx >= this.items.length) {
            this.idx = -1;
            return this.savedDraft;
        }
        return this.items[this.idx];
    }
}

export function mountComposer(host: HTMLElement, onSend: () => void, onChange?: OnChangeCallback): ComposerHandle {
    const history = new ComposerHistory();
    const L = window.scribli.languages;

    const app: App = window.scribli.ws.app;
    const protyle = new Protyle(app, host, {
        lite: true,
        blockId: "",
        render: {
            gutter: false,
            breadcrumb: false,
            scroll: false,
            background: false,
            title: false,
        },
        hint: {
            extend: [{
                key: "((",
                hint: hintRef,
            }, {
                key: "【【",
                hint: hintRef,
            }, {
                key: "（（",
                hint: hintRef,
            }, {
                key: "[[",
                hint: hintRef,
            }, {
                key: "/",
                hint: hintSkill,
            }, {
                key: "、",
                hint: hintSkill,
            }],
        },
    });

    const p = protyle.protyle;
    const wysiwyg = p.wysiwyg!;

    wysiwyg.element.innerHTML = "";
    const emptyElement = genEmptyElement(false, false);
    emptyElement.firstElementChild.classList.add("protyle-wysiwyg--empty");
    emptyElement.firstElementChild.setAttribute("placeholder", L.agentInputPlaceholder);
    wysiwyg.element.appendChild(emptyElement);

    const updatePlaceholder = () => {
        const isEmpty = (wysiwyg.element.textContent || "").replace(new RegExp(Constants.ZWSP, "g"), "").trim() === "";
        wysiwyg.element.classList.toggle("agent-composer--empty", isEmpty);
    };
    updatePlaceholder();

    const contentObserver = new MutationObserver(() => {
        updatePlaceholder();
        if (onChange) {
            onChange();
        }
    });
    contentObserver.observe(wysiwyg.element, {childList: true, characterData: true, subtree: true});

    wysiwyg.element.addEventListener("keydown", (event: KeyboardEvent) => {
        const hintEl = p.hint?.element;
        if (hintEl && !hintEl.classList.contains("fn__none")) {
            if (event.key === "Enter" || event.key.indexOf("Arrow") > -1) {
                if (p.hint!.select(event, p)) {
                    event.preventDefault();
                    event.stopPropagation();
                }
            }
            return;
        }

        if (event.key === "Enter" && !event.shiftKey && !event.ctrlKey && !event.altKey && !event.metaKey) {
            event.preventDefault();
            event.stopPropagation();
            onSend();
            return;
        }

        if (event.key === "ArrowUp" && !event.shiftKey) {
            const isEmpty = (wysiwyg.element.textContent || "").replace(new RegExp(Constants.ZWSP, "g"), "").trim() === "";
            if ((history.isBrowsing() || isEmpty) && history.has()) {
                event.preventDefault();
                event.stopPropagation();
                const target = history.isBrowsing() ?
                    history.navigateUp() : history.beginBrowsing(wysiwyg.element.innerHTML);
                wysiwyg.element.innerHTML = p.lute.Md2BlockDOM(target);
                return;
            }
        }
        if (event.key === "ArrowDown" && history.isBrowsing()) {
            event.preventDefault();
            event.stopPropagation();
            const target = history.navigateDown();
            if (history.isBrowsing()) {
                p.wysiwyg.element.innerHTML = p.lute.Md2BlockDOM(target);
            } else {
                wysiwyg.element.innerHTML = target;
            }
            return;
        }

        if (history.isBrowsing() && event.key.length === 1 && !event.ctrlKey && !event.metaKey && !event.altKey) {
            history.resetCursor();
        }
    }, true);

    const getMarkdown = (): string => {
        return p.lute.BlockDOM2StdMd(wysiwyg.element.innerHTML).trim();
    };

    const getBlockHTML = (): string => {
        const clone = wysiwyg.element.cloneNode(true) as HTMLElement;
        resetEmbedBlocks(clone);
        return clone.innerHTML;
    };

    return {
        focus: () => protyle.focus(),
        destroy: () => {
            contentObserver.disconnect();
            protyle.destroy();
        },
        getSendData: () => {
            const references: { id: string; title: string }[] = [];
            wysiwyg.element.querySelectorAll('[data-type~="block-ref"]').forEach((ref) => {
                references.push({
                    id: ref.getAttribute("data-id") || "",
                    title: ref.textContent || "",
                });
            });
            return {text: getMarkdown(), blockHTML: getBlockHTML(), references};
        },
        clear: () => {
            wysiwyg.element.innerHTML = "";
            const emptyElement = genEmptyElement(false, false);
            emptyElement.firstElementChild.classList.add("protyle-wysiwyg--empty");
            emptyElement.firstElementChild.setAttribute("placeholder", L.agentInputPlaceholder);
            wysiwyg.element.appendChild(emptyElement);
            p.undo.clear();
            updatePlaceholder();
            history.resetCursor();
        },
        pushHistory: (text: string) => history.push(text),
        getHistory: () => history.get(),
        clearHistory: () => history.clear(),
        restoreHistory: (h: string[]) => history.restore(h),
        insertMention: (id: string, label: string) => {
            protyle.insert('<span data-type="block-ref" data-id="' + id + '" data-subtype="d">' +
                escapeHtml(label) + "</span>" + Constants.ZWSP);
        },
        insertMentions: (mentions: Array<{ id: string; label: string }>) => {
            const html = mentions.map((m) =>
                '<span data-type="block-ref" data-id="' + m.id + '" data-subtype="d">' +
                escapeHtml(m.label) + "</span>" + Constants.ZWSP + " "
            ).join("");
            if (html) {
                protyle.insert(html);
            }
        },
        renderBlockHTML: (element, onEmbedRender) => {
            resetEmbedBlocks(element);
            blockRender(p, element, undefined, onEmbedRender);
        },
    };
}

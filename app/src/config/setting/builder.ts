import type {App} from "../../index";
import type {StackLine, SwitchQueryItem} from "../render/parts";
import {genButtonRowHtml, genStackHtml, genSwitchQueryHtml, genTextPairHtml} from "../render/render";
import {
    controlBoolean,
    controlNumber,
    controlRange,
    controlSelect,
    controlString,
    controlTextBlock,
    type SettingControl,
} from "./control";
import {registerSettingGroup} from "./group";
import {registerSettingItem, type RegisterSettingItem, removeSettingTabItems} from "./item";
import {scanSettingTabSearch} from "../search/scan";
import {buildSearchIndex, normalizeSearchText} from "../search/normalize";
import {applySettingTabSearchVisibility, mountSettingTab} from "./mount";

type SaveFn = (value: unknown) => void | Promise<void>;

export interface SettingTabShell<TId extends string = string> {
    id: TId;
    icon: string;
    title: () => string;
    hidden?: () => boolean;
}

interface ItemsSettingTabOptions<TId extends string = string> extends SettingTabShell<TId> {
    defaultSave?: (controlId: string, value: unknown) => void;
    afterMount?: (root: HTMLElement, app?: App) => void | Promise<void>;
}

interface PanelSettingTabOptions<TId extends string = string> extends SettingTabShell<TId> {
    searchStrings: () => string[];
    mount: (root: HTMLElement, keywords?: string, app?: App) => void | Promise<void>;
}

type ControlSpecBase = {
    title: string;
    desc?: string;
    save?: SaveFn;
    afterMount?: (root: HTMLElement) => void | Promise<void>;
    readConfig?: () => unknown;
};
type SwitchSpec = ControlSpecBase;
type NumberSpec = ControlSpecBase & {
    min?: number;
    max?: number;
    step?: string;
    unit?: string;
};
type RangeSpec = ControlSpecBase & {
    min: number;
    max: number;
    step: number;
};
type SelectSpec = ControlSpecBase & {
    options: {
        value: number | string;
        label?: string;
    }[];
    readConfig?: () => number | string;
};
type TextSpec = ControlSpecBase & {
    desc: string;
};
type TextBlockSpec = TextSpec & {
    mode: "input-text" | "input-password" | "textarea";
};
type SlotSpec = {
    key: string;
    keywords: string[];
    html: () => string;
    afterMount?: (root: HTMLElement) => void | Promise<void>;
};
type CompositeControlSpec = {
    control: SettingControl;
    save?: SaveFn;
};
type CompositeSpec = SlotSpec & {
    controls: CompositeControlSpec[];
};
type SwitchQueryInputItem =
    | {kind: "switch"; id: string; label: string; icon?: string}
    | {kind: "number"; id: string; label: string; min?: number; max?: number};
type SwitchQuerySpec = {
    key: string;
    title: string;
    footer?: string;
    items: SwitchQueryInputItem[];
};
type TextPairSpec = {
    title: string;
    desc: string;
    leftId: string;
    rightId: string;
};
type StackSpec = {
    key: string;
    keywords: string[];
    afterMount?: (root: HTMLElement) => void | Promise<void>;
};
type StackButtonSpec = {
    id: string;
    label: string;
    icon: string;
};
type StackSelectSpec = {
    desc: string;
    options: {
        value: number | string;
        label?: string;
    }[];
};
type StackNumberSpec = {
    desc: string;
    min?: number;
    max?: number;
};
type StackSwitchSpec = {
    desc: string;
};
type StackTextBlockSpec = {
    mode: "input-text" | "input-password" | "textarea";
};
type ButtonSpec = {
    id: string;
    title: string;
    desc?: string;
    label: string;
    icon: string;
    afterMount?: (root: HTMLElement) => void | Promise<void>;
};

const stackLinesToControls = (lines: StackLine[]): CompositeControlSpec[] => {
    const controls: CompositeControlSpec[] = [];
    for (const line of lines) {
        if (line.left.kind === "textBlock") {
            controls.push({control: line.left});
        }
        const right = line.right;
        if (!right || right.kind === "button") {
            continue;
        }
        controls.push({control: right});
    }
    return controls;
};

class StackLineBuilder {
    private readonly lines: StackLine[] = [];

    getLines(): StackLine[] {
        return this.lines;
    }

    title(text: string) {
        this.lines.push({left: {kind: "title", text}});
        return this;
    }

    button(spec: StackButtonSpec) {
        const last = this.lines[this.lines.length - 1];
        if (last) {
            last.right = {kind: "button", ...spec};
        }
        return this;
    }

    desc(text: string) {
        this.lines.push({left: {kind: "desc", text}});
        return this;
    }

    select(id: string, spec: StackSelectSpec) {
        const control = controlSelect(id, {options: spec.options});
        this.lines.push({
            left: {kind: "desc", text: spec.desc},
            right: control,
        });
        return this;
    }

    switch(id: string, spec: StackSwitchSpec) {
        const control = controlBoolean(id);
        this.lines.push({
            left: {kind: "desc", text: spec.desc},
            right: control,
        });
        return this;
    }

    number(id: string, spec: StackNumberSpec) {
        const control = controlNumber(id, {min: spec.min, max: spec.max});
        this.lines.push({
            left: {kind: "desc", text: spec.desc},
            right: control,
        });
        return this;
    }

    textBlock(id: string, spec: StackTextBlockSpec) {
        const control = controlTextBlock(id, {mode: spec.mode});
        this.lines.push({left: control});
        return this;
    }
}

class SettingGroupBuilder<TId extends string> {
    constructor(
        private readonly tab: ItemsSettingTabOptions<TId>,
        readonly groupId: string,
    ) {}

    /**
     */
    private registerFullItem(
        id: string,
        spec: ControlSpecBase,
        control: SettingControl,
    ) {
        const rowParts = [
            {kind: "title" as const, text: spec.title},
            ...(spec.desc?.trim() ? [{kind: "desc" as const, text: spec.desc}] : []),
            control,
        ];
        const afterMount = control.afterMount || spec.afterMount
            ? async (root: HTMLElement) => {
                await control.afterMount?.(root);
                await spec.afterMount?.(root);
            }
            : undefined;
        registerSettingItem({
            id,
            tabId: this.tab.id,
            groupId: this.groupId,
            kind: "full",
            rowParts,
            readValue: (el) => control.readValue(el),
            save: spec.save ?? this.tab.defaultSave?.bind(null, id),
            afterMount,
        } as RegisterSettingItem);
        return this;
    }

    switch(id: string, spec: SwitchSpec) {
        return this.registerFullItem(id, spec, controlBoolean(id, {
            readConfig: spec.readConfig as (() => boolean) | undefined,
        }));
    }

    number(id: string, spec: NumberSpec) {
        return this.registerFullItem(id, spec, controlNumber(id, {
            min: spec.min,
            max: spec.max,
            step: spec.step,
            unit: spec.unit,
            readConfig: spec.readConfig as (() => number) | undefined,
        }));
    }

    range(id: string, spec: RangeSpec) {
        return this.registerFullItem(id, spec, controlRange(id, {
            min: spec.min,
            max: spec.max,
            step: spec.step,
            readConfig: spec.readConfig as (() => number) | undefined,
        }));
    }

    select(id: string, spec: SelectSpec) {
        return this.registerFullItem(id, spec, controlSelect(id, {
            options: spec.options,
            readConfig: spec.readConfig,
        }));
    }

    text(id: string, spec: TextSpec) {
        return this.registerFullItem(id, spec, controlString(id, {
            readConfig: spec.readConfig as (() => string) | undefined,
        }));
    }

    textBlock(id: string, spec: TextBlockSpec) {
        return this.registerFullItem(id, spec, controlTextBlock(id, {
            mode: spec.mode,
            readConfig: spec.readConfig as (() => string) | undefined,
        }));
    }

    textPair(spec: TextPairSpec) {
        const leftControl = controlString(spec.leftId);
        const rightControl = controlString(spec.rightId);
        this.composite({
            key: `textPair_${spec.leftId}_${spec.rightId}`,
            keywords: [spec.title, spec.desc],
            html: () => genTextPairHtml(spec.title, spec.desc, leftControl, rightControl),
            controls: [
                {control: leftControl},
                {control: rightControl},
            ],
        });
        return this;
    }

    stack(spec: StackSpec, configure: (stack: StackLineBuilder) => void) {
        const stack = new StackLineBuilder();
        configure(stack);
        const lines = stack.getLines();
        this.composite({
            key: spec.key,
            keywords: spec.keywords,
            html: () => genStackHtml(lines),
            afterMount: spec.afterMount,
            controls: stackLinesToControls(lines),
        });
        return this;
    }

    button(spec: ButtonSpec) {
        this.slot({
            key: `button_${spec.id}`,
            keywords: [spec.title, spec.desc, spec.label].filter((s): s is string => Boolean(s)),
            html: () => genButtonRowHtml(spec.id, spec.title, spec.desc, spec.label, spec.icon),
            afterMount: spec.afterMount,
        });
        return this;
    }

    switchQuery(spec: SwitchQuerySpec) {
        const searchTexts = [
            spec.title,
            ...(spec.footer ? [spec.footer] : []),
            ...spec.items.map((item) => item.label),
        ];
        const items: SwitchQueryItem[] = [];
        const controls: CompositeControlSpec[] = [];
        for (const item of spec.items) {
            if (item.kind === "switch") {
                const control = controlBoolean(item.id);
                items.push({...control, label: item.label, icon: item.icon});
                controls.push({control});
            } else {
                const control = controlNumber(item.id, {min: item.min, max: item.max});
                items.push({...control, label: item.label});
                controls.push({control});
            }
        }
        this.composite({
            key: spec.key,
            keywords: searchTexts,
            html: () => genSwitchQueryHtml(spec.title, items, spec.footer),
            controls,
        });
        return this;
    }

    /**
     */
    slot(spec: SlotSpec) {
        const id = `${this.tab.id}.__slot.${spec.key}`;
        registerSettingItem({
            id,
            tabId: this.tab.id,
            groupId: this.groupId,
            kind: "render",
            html: spec.html,
            searchTexts: () => [...spec.keywords],
            afterMount: spec.afterMount,
        } as RegisterSettingItem);
        return this;
    }

    /**
     */
    composite(spec: CompositeSpec) {
        this.slot({
            key: spec.key,
            keywords: spec.keywords,
            html: spec.html,
            afterMount: spec.afterMount,
        });
        for (const entry of spec.controls) {
            registerSettingItem({
                id: entry.control.id,
                tabId: this.tab.id,
                groupId: this.groupId,
                kind: "binding",
                control: entry.control,
                readValue: (el) => entry.control.readValue(el),
                save: entry.save ?? this.tab.defaultSave?.bind(null, entry.control.id),
            } as RegisterSettingItem);
        }
        return this;
    }
}

export class SettingTabBuilder<TId extends string = string> {
    constructor(private readonly tab: ItemsSettingTabOptions<TId>) {}

    group(groupId: string, groupTitle: string) {
        registerSettingGroup(this.tab.id, groupId, groupTitle);
        return new SettingGroupBuilder(this.tab, groupId);
    }
}

export interface SettingTabSearchResult {
    matches: boolean;
    visibleItemIds?: Set<string>;
    visibleGroupIds?: Set<string>;
}

export interface SettingTabMountContext {
    keywords: string;
    visibleItemIds?: Set<string>;
    visibleGroupIds?: Set<string>;
}

export type SettingTab = SettingTabShell & {
    mount: (
        root: HTMLElement,
        search?: Partial<SettingTabMountContext>,
        app?: App,
        rebuild?: boolean,
    ) => Promise<void>;
    scanSearch: (keywords: string) => SettingTabSearchResult;
};

export class SettingBuilder {
    tab<TId extends string>(
        options: ItemsSettingTabOptions<TId>,
        register: (tab: SettingTabBuilder<TId>) => void,
    ): SettingTab {
        const {afterMount, ...shell} = options;
        let registered = false;
        let tabSearchTitle: string | undefined;
        const ensureRegistered = () => {
            if (registered) {
                return;
            }
            registered = true;
            register(new SettingTabBuilder(options));
        };
        return {
            ...shell,
            mount: async (root, search, app, rebuild) => {
                const {visibleItemIds, visibleGroupIds} = search ?? {};
                if (rebuild) {
                    removeSettingTabItems(options.id);
                    registered = false;
                    root.innerHTML = "";
                }
                ensureRegistered();
                if (!root.innerHTML) {
                    await mountSettingTab(options.id, root);
                    await afterMount?.(root, app);
                }
                if (visibleItemIds && visibleGroupIds) {
                    applySettingTabSearchVisibility(root, visibleItemIds, visibleGroupIds);
                }
            },
            scanSearch: (keywords) => {
                ensureRegistered();
                if (tabSearchTitle === undefined) {
                    tabSearchTitle = normalizeSearchText(options.title());
                }
                return scanSettingTabSearch(options.id, tabSearchTitle, keywords);
            },
        };
    }

    panel<TId extends string>(
        options: PanelSettingTabOptions<TId>,
    ): SettingTab {
        const {searchStrings, mount, ...shell} = options;
        let tabSearchTitle: string | undefined;
        let tabSearchIndex: readonly string[] | undefined;
        return {
            ...shell,
            mount: async (root, {keywords} = {}, app, _rebuild) => {
                void _rebuild;
                mount(root, keywords, app);
            },
            scanSearch: (keywords) => {
                if (tabSearchTitle === undefined) {
                    tabSearchTitle = normalizeSearchText(options.title());
                }
                if (tabSearchIndex === undefined) {
                    tabSearchIndex = buildSearchIndex(searchStrings());
                }
                let matches = false;
                if (tabSearchTitle.length > 0 && tabSearchTitle.includes(keywords)) {
                    matches = true;
                } else if (tabSearchIndex.some((s) => s.includes(keywords))) {
                    matches = true;
                }
                return {matches};
            },
        };
    }
}

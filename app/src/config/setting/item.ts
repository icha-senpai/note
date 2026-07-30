import type {RowPart} from "../render/parts";
import type {SettingControl} from "./control";
import type {SettingGroup} from "./group";
import {getSettingGroupsByTabId} from "./group";
import {buildItemSearchIndex} from "../search/normalize";

type SettingItemBase = {
    id: string;
    tabId: string;
    groupId: string;
    searchIndex: readonly string[];
    readValue?: (el: HTMLElement) => unknown;
    save?: (value: unknown) => void | Promise<void>;
    afterMount?: (root: HTMLElement) => void | Promise<void>;
};

type FullSettingItem = SettingItemBase & {
    kind: "full";
    rowParts: RowPart[];
};

type RenderSettingItem = SettingItemBase & {
    kind: "render";
    html: () => string;
    searchTexts?: () => string[];
};

type BindingSettingItem = SettingItemBase & {
    kind: "binding";
    control: SettingControl;
};

type SettingItem = FullSettingItem | RenderSettingItem | BindingSettingItem;
export type MountableSettingItem = FullSettingItem | RenderSettingItem;
export type RegisterSettingItem =
    | Omit<FullSettingItem, "searchIndex">
    | Omit<RenderSettingItem, "searchIndex">
    | Omit<BindingSettingItem, "searchIndex">;

export type TabGroupEntry = {
    group: SettingGroup;
    items: MountableSettingItem[];
};

const settingItemsById = new Map<SettingItem["id"], SettingItem>();
const itemsByGroupCache = new Map<string, Map<string, MountableSettingItem[]>>();

const getMountableItemsByGroup = (tabId: string): Map<string, MountableSettingItem[]> => {
    let itemsByGroup = itemsByGroupCache.get(tabId);
    if (itemsByGroup) {
        return itemsByGroup;
    }
    itemsByGroup = new Map<string, MountableSettingItem[]>();
    for (const item of settingItemsById.values()) {
        if (item.kind !== "binding" && item.tabId === tabId) {
            const groupItems = itemsByGroup.get(item.groupId);
            if (groupItems) {
                groupItems.push(item);
            } else {
                itemsByGroup.set(item.groupId, [item]);
            }
        }
    }
    itemsByGroupCache.set(tabId, itemsByGroup);
    return itemsByGroup;
};

export const getTabGroupEntries = (tabId: string): TabGroupEntry[] => {
    const itemsByGroup = getMountableItemsByGroup(tabId);
    return getSettingGroupsByTabId(tabId).map((group) => ({
        group,
        items: itemsByGroup.get(group.id) ?? [],
    }));
};

export const registerSettingItem = (item: RegisterSettingItem) => {
    settingItemsById.set(item.id, {
        ...item,
        searchIndex: buildItemSearchIndex(item)
    } as SettingItem);
    if (item.kind !== "binding") {
        itemsByGroupCache.delete(item.tabId);
    }
};

export const getSettingItem = (id: string) => settingItemsById.get(id);

export const removeSettingTabItems = (tabId: string) => {
    for (const [id, item] of settingItemsById) {
        if (item.tabId === tabId) {
            settingItemsById.delete(id);
        }
    }
    itemsByGroupCache.delete(tabId);
};

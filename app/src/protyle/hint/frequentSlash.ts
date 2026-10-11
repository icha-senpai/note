// Adapted from SiYuan 2031e1637d. Copyright (c) 2020-present, b3log.org.
// SPDX-License-Identifier: AGPL-3.0-or-later
export type SlashUsage = Record<string, number>;

export const normalizeSlashUsage = (value: unknown): SlashUsage => {
    if (!value || typeof value !== "object" || Array.isArray(value)) { return {}; }
    return Object.fromEntries(Object.entries(value).filter(([key, count]) =>
        /^[a-zA-Z0-9_-]{1,80}$/.test(key) && Number.isSafeInteger(count) && (count as number) > 0));
};

export const frequentSlashItems = <T extends {id?: string; filter?: string[]; html: string}>(items: T[], usage: SlashUsage): T[] => {
    const seen = new Set<string>();
    return items.map((item, index) => ({item, index})).filter(({item}) => {
        if (!item.id || !item.filter || item.html === "separator" || seen.has(item.id) || !Object.prototype.hasOwnProperty.call(usage, item.id)) { return false; }
        seen.add(item.id);
        return true;
    }).sort((a, b) => usage[b.item.id] - usage[a.item.id] || a.index - b.index).slice(0, 5).map(({item}) => item);
};

import {mergeRecordByDottedPath} from "./dotPath";

export type SettingSaveState = "saving" | "saved" | "failed";

export const createSettingSaveQueue = <T>(options: {
    read: () => T;
    save: (data: T) => Promise<T>;
    apply: (data: T) => void;
    status: (state: SettingSaveState) => void;
}) => {
    let running = false;
    let base: T;
    const pending: Array<{path: string, value: unknown, applied?: (data: T) => void}> = [];
    const flush = async () => {
        if (running || pending.length === 0) {
            return;
        }
        running = true;
        options.status("saving");
        try {
            while (pending.length > 0) {
                const batch = pending.splice(0);
                let payload = base;
                batch.forEach(edit => {
                    payload = mergeRecordByDottedPath(payload as Record<string, unknown>, edit.path, edit.value) as T;
                });
                try {
                    base = await options.save(payload);
                } catch {
                    pending.unshift(...batch);
                    options.status("failed");
                    return;
                }
                options.apply(base);
                batch.forEach(edit => edit.applied?.(base));
            }
            options.status("saved");
        } finally {
            running = false;
        }
    };
    const patchMany = (edits: Array<{path: string, value: unknown}>, applied?: (data: T) => void) => {
        if (edits.length === 0) { return; }
        if (!running && pending.length === 0) {
            base = options.read();
        }
        pending.push(...edits.map((edit, index) => ({...edit, applied: index === edits.length - 1 ? applied : undefined})));
        void flush();
    };
    return {
        patch(path: string, value: unknown, applied?: (data: T) => void) { patchMany([{path, value}], applied); },
        patchMany,
        retry() { void flush(); },
        apply(data: T) { base = data; options.apply(data); },
    };
};

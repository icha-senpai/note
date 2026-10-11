import {fetchSyncPost} from "../../util/fetch";
import {createSettingSaveQueue} from "./saveQueue";
import {setSettingSaveState} from "./saveStatus";

export function createConfigNamespaceApi<TData>(options: {
    namespace: string;
    getConfig: () => TData;
    setConfig: (data: TData) => void;
    apiPath: string;
    applyFromResponse?: boolean;
}): {
    /**
     */
    patch: (relOrFullId: string, value: unknown, onApplied?: (data: TData) => void) => void;
    patchMany: (values: Record<string, unknown>, onApplied?: (data: TData) => void) => void;
    apply: (data: TData) => void;
} {
    const {namespace, getConfig, setConfig, apiPath, applyFromResponse = true} = options;
    const prefix = `${namespace}.`;

    const queue = createSettingSaveQueue<TData>({
        read: getConfig,
        async save(payload) {
            const response = await fetchSyncPost(apiPath, payload);
            if (response.code !== 0) {
                throw new Error(response.msg);
            }
            return response.data as TData;
        },
        apply(data) {
            if (applyFromResponse) {
                setConfig(data);
            }
        },
        status(state) { setSettingSaveState(namespace, state, queue.retry); },
    });

    return {
        patch(relOrFullId, value, onApplied) {
            const rel = relOrFullId.startsWith(prefix) ? relOrFullId.slice(prefix.length) : relOrFullId;
            if (rel) {
                queue.patch(rel, value, onApplied);
            }
        },
        patchMany(values, onApplied) {
            queue.patchMany(Object.entries(values).map(([path, value]) => ({path: path.startsWith(prefix) ? path.slice(prefix.length) : path, value})).filter(edit => edit.path), onApplied);
        },
        apply(data) { queue.apply(data); if (!applyFromResponse) { setConfig(data); } },
    };
}

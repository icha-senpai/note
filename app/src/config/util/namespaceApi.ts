import {fetchPost} from "../../util/fetch";
import {mergeRecordByDottedPath} from "./dotPath";

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
    apply: (data: TData) => void;
} {
    const {namespace, getConfig, setConfig, apiPath, applyFromResponse = true} = options;
    const prefix = `${namespace}.`;

    const post = (payload: TData, onApplied?: (data: TData) => void) => {
        fetchPost(apiPath, payload, (response) => {
            const data = response.data as TData;
            if (applyFromResponse) {
                setConfig(data);
            }
            onApplied?.(data);
        });
    };

    return {
        patch(relOrFullId, value, onApplied) {
            const rel = relOrFullId.startsWith(prefix) ? relOrFullId.slice(prefix.length) : relOrFullId;
            if (rel) {
                const prev = getConfig() as unknown as Record<string, unknown>;
                post(mergeRecordByDottedPath(prev, rel, value) as unknown as TData, onApplied);
            }
        },
        apply: setConfig,
    };
}

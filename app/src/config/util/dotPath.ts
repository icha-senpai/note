export function getAtPath(root: unknown, dottedPath: string): unknown {
    const segments = dottedPath.split(".");
    let cur: unknown = root;
    for (const s of segments) {
        if (cur === null || cur === undefined) {
            return undefined;
        }
        cur = (cur as Record<string, unknown>)[s];
    }
    return cur;
}

/**
 */
function assignPathImmutable(
    obj: Record<string, unknown>,
    segments: string[],
    value: unknown
): Record<string, unknown> {
    if (segments.length === 1) {
        return {...obj, [segments[0]]: value};
    }
    const [head, ...rest] = segments;
    const child = obj[head];
    const base =
        typeof child === "object" && child !== null && !Array.isArray(child)
            ? {...(child as Record<string, unknown>)}
            : {};
    return {
        ...obj,
        [head]: assignPathImmutable(base, rest, value),
    };
}

export function mergeRecordByDottedPath<T extends Record<string, unknown>>(
    base: T,
    dottedId: string,
    value: unknown
): T {
    const segments = dottedId.split(".");
    return assignPathImmutable({...(base as unknown as Record<string, unknown>)}, segments, value) as T;
}

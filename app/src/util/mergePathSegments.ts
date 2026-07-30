/**
 */
export const mergePathSegments = (pathSegments: string[], segments: string[]): string[] => {
    for (const segment of segments) {
        if (segment === "..") {
            if (pathSegments.length > 0) {
                pathSegments.pop();
            }
        } else if (segment && segment !== ".") {
            pathSegments.push(segment);
        }
    }
    return pathSegments;
};

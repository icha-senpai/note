/**
 *
 *
 *
 *
 */

import {mergePathSegments} from "./mergePathSegments";

export type NewDocTargetByHPath = {
    kind: "hPath";
    targetNotebookId: string;
    hPath: string;
    title: string;
};

export type NewDocTargetSubDoc = {
    kind: "subDoc";
    targetNotebookId: string;
    parentPath: string;
    title: string;
};

export type NewDocTarget = NewDocTargetByHPath | NewDocTargetSubDoc;

export const getNewDocTargetFromSavePath = (request: {
    templatePath: string;
    hPath: string;
    targetNotebookId: string;
    currentNotebookId: string;
    name?: string;
    hasFocusTarget: boolean;
    currentPath?: string;
}): NewDocTarget => {
    const {targetNotebookId} = request;

    let templatePath = request.templatePath.trim();
    let isAbsolute = templatePath.startsWith("/");
    if (targetNotebookId !== request.currentNotebookId && templatePath && !isAbsolute) {
        templatePath = "/" + templatePath;
        isAbsolute = true;
    }

    if (!templatePath && request.hasFocusTarget && !request.name
        && targetNotebookId === request.currentNotebookId) {
        return {
            kind: "subDoc",
            targetNotebookId,
            parentPath: request.currentPath || "/",
            title: "",
        };
    }

    let parentTemplate: string;
    let title = "";
    if (!templatePath) {
        parentTemplate = request.hasFocusTarget ? "" : "/";
        title = request.name || "";
    } else if (templatePath.endsWith("/")) {
        parentTemplate = templatePath;
        title = request.name || "";
    } else {
        const segments = templatePath.split("/").filter(Boolean);
        if (segments.length <= 1) {
            parentTemplate = isAbsolute ? "/" : "";
        } else {
            const parentSegmentPath = segments.slice(0, -1).join("/");
            parentTemplate = isAbsolute ? "/" + parentSegmentPath : parentSegmentPath;
        }
        title = request.name || segments[segments.length - 1];
    }

    const templateSegments = parentTemplate.split("/").filter(Boolean);
    let parentPathSegments: string[];
    if (parentTemplate.startsWith("/")) {
        parentPathSegments = mergePathSegments([], templateSegments);
    } else {
        parentPathSegments = mergePathSegments(request.hPath.split("/").filter(Boolean), templateSegments);
    }

    let hPath: string;
    if (title) {
        hPath = "/" + [...parentPathSegments, title].join("/");
    } else {
        hPath = parentPathSegments.length === 0 ? "/" : "/" + parentPathSegments.join("/") + "/";
    }

    return {
        kind: "hPath",
        targetNotebookId,
        hPath,
        title,
    };
};

export const getNewDocTargetFromTree = (request: {
    templatePath: string;
    currentNotebookId: string;
    currentPath: string;
    name?: string;
}): NewDocTargetSubDoc => {
    const templatePath = request.templatePath.trim();
    let title = "";
    if (request.name) {
        title = request.name;
    } else if (templatePath && !templatePath.endsWith("/")) {
        const segments = templatePath.split("/").filter(Boolean);
        title = segments[segments.length - 1];
    }
    return {
        kind: "subDoc",
        targetNotebookId: request.currentNotebookId,
        parentPath: request.currentPath || "/",
        title,
    };
};

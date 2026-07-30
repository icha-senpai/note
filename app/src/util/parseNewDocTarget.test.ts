import {describe, it} from "node:test";
import * as assert from "node:assert/strict";
import {getNewDocTargetFromSavePath, getNewDocTargetFromTree, NewDocTarget} from "./parseNewDocTarget";

const assertSubDoc = (target: NewDocTarget, expected: {
    targetNotebookId?: string;
    parentPath: string;
    title: string;
}) => {
    assert.equal(target.kind, "subDoc");
    if (target.kind === "subDoc") {
        if (expected.targetNotebookId !== undefined) {
            assert.equal(target.targetNotebookId, expected.targetNotebookId);
        }
        assert.equal(target.parentPath, expected.parentPath);
        assert.equal(target.title, expected.title);
    }
};

const assertHPath = (target: NewDocTarget, expected: {
    targetNotebookId?: string;
    hPath: string;
    title: string;
}) => {
    assert.equal(target.kind, "hPath");
    if (target.kind === "hPath") {
        if (expected.targetNotebookId !== undefined) {
            assert.equal(target.targetNotebookId, expected.targetNotebookId);
        }
        assert.equal(target.hPath, expected.hPath);
        assert.equal(target.title, expected.title);
    }
};

describe("getNewDocTargetFromSavePath", () => {
    const nestedDocPath = "/20260628041644-ndcuikw/20260628040939-kkaajwr.sy";
    const nestedHPath = "/parent1/parent2/docName";
    const rootDocPath = "/20260628041702-kqfrg7p.sy";
    const rootHPath = "/docName";
    const notebookId = "nb";

    const nestedContext = {
        hPath: nestedHPath,
        targetNotebookId: notebookId,
        currentNotebookId: notebookId,
        hasFocusTarget: true,
        currentPath: nestedDocPath,
    };

    describe("new document target", () => {
        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: ""});
            assertSubDoc(target, {targetNotebookId: notebookId, parentPath: nestedDocPath, title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "", name: "docName2"});
            assertHPath(target, {hPath: "/parent1/parent2/docName/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({
                ...nestedContext,
                templatePath: "",
                hPath: rootHPath,
                currentPath: rootDocPath,
            });
            assertSubDoc(target, {parentPath: rootDocPath, title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({
                ...nestedContext,
                templatePath: "",
                hPath: rootHPath,
                currentPath: rootDocPath,
                name: "docName2",
            });
            assertHPath(target, {hPath: "/docName/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({
                ...nestedContext,
                templatePath: "",
                hPath: "/",
                currentPath: "/",
            });
            assertSubDoc(target, {parentPath: "/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({
                ...nestedContext,
                templatePath: "",
                hasFocusTarget: false,
                hPath: "/",
                currentPath: "/",
            });
            assertHPath(target, {hPath: "/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({
                ...nestedContext,
                templatePath: "",
                hasFocusTarget: false,
                hPath: "/",
                name: "docName2",
            });
            assertHPath(target, {hPath: "/docName2", title: "docName2"});
        });
    });

    describe("new document target", () => {
        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "/parent3/"});
            assertHPath(target, {hPath: "/parent3/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "/parent3/", name: "docName2"});
            assertHPath(target, {hPath: "/parent3/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "/", hPath: "/"});
            assertHPath(target, {hPath: "/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "/parent1/parent2/"});
            assertHPath(target, {hPath: "/parent1/parent2/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "parent3/"});
            assertHPath(target, {hPath: "/parent1/parent2/docName/parent3/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "parent3/parent4/"});
            assertHPath(target, {hPath: "/parent1/parent2/docName/parent3/parent4/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "../"});
            assertHPath(target, {hPath: "/parent1/parent2/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "../", name: "docName2"});
            assertHPath(target, {hPath: "/parent1/parent2/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "../../"});
            assertHPath(target, {hPath: "/parent1/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "../parent3/"});
            assertHPath(target, {hPath: "/parent1/parent2/parent3/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "../../parent3/parent4/"});
            assertHPath(target, {hPath: "/parent1/parent3/parent4/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "../", hPath: "/"});
            assertHPath(target, {hPath: "/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "  parent3/  "});
            assertHPath(target, {hPath: "/parent1/parent2/docName/parent3/", title: ""});
        });
    });

    describe("new document target", () => {
        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "docName2"});
            assertHPath(target, {hPath: "/parent1/parent2/docName/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "docName2", name: "docName3"});
            assertHPath(target, {hPath: "/parent1/parent2/docName/docName3", title: "docName3"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "  docName2  "});
            assertHPath(target, {hPath: "/parent1/parent2/docName/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "/docName2"});
            assertHPath(target, {hPath: "/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "/docName2", name: "docName3"});
            assertHPath(target, {hPath: "/docName3", title: "docName3"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "parent3/docName2"});
            assertHPath(target, {hPath: "/parent1/parent2/docName/parent3/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "/parent3/docName2"});
            assertHPath(target, {hPath: "/parent3/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "../docName2"});
            assertHPath(target, {hPath: "/parent1/parent2/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "../../docName2"});
            assertHPath(target, {hPath: "/parent1/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "../parent3/docName2"});
            assertHPath(target, {hPath: "/parent1/parent2/parent3/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...nestedContext, templatePath: "../docName2", hPath: "/"});
            assertHPath(target, {hPath: "/docName2", title: "docName2"});
        });
    });

    describe("new document target", () => {
        const crossNotebook = {
            ...nestedContext,
            targetNotebookId: "box-b",
            currentNotebookId: "box-a",
            hPath: "/",
            currentPath: nestedDocPath,
        };

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...crossNotebook, templatePath: "docName2"});
            assertHPath(target, {targetNotebookId: "box-b", hPath: "/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...crossNotebook, templatePath: "parent3/parent4/"});
            assertHPath(target, {targetNotebookId: "box-b", hPath: "/parent3/parent4/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...crossNotebook, templatePath: "../docName2"});
            assertHPath(target, {targetNotebookId: "box-b", hPath: "/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...crossNotebook, templatePath: "/parent3/docName2"});
            assertHPath(target, {targetNotebookId: "box-b", hPath: "/parent3/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({
                ...crossNotebook,
                templatePath: "",
                hasFocusTarget: false,
                name: "docName2",
            });
            assertHPath(target, {targetNotebookId: "box-b", hPath: "/docName2", title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...crossNotebook, templatePath: ""});
            assertHPath(target, {targetNotebookId: "box-b", hPath: "/", title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromSavePath({...crossNotebook, templatePath: "", name: "docName2"});
            assertHPath(target, {targetNotebookId: "box-b", hPath: "/docName2", title: "docName2"});
        });
    });
});

describe("getNewDocTargetFromTree", () => {
    const notebookId = "nb";
    const parentDocPath = "/20260628041644-ndcuikw.sy";
    const nestedDocPath = "/20260628041644-ndcuikw/20260628040939-kkaajwr.sy";
    const parentDirPath = "/20260628041644-ndcuikw";
    const rootPath = "/";

    describe("new document target", () => {
        const treeContext = {currentNotebookId: notebookId, currentPath: parentDocPath};

        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({...treeContext, templatePath: ""});
            assertSubDoc(target, {targetNotebookId: notebookId, parentPath: parentDocPath, title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({...treeContext, templatePath: "", name: "docName2"});
            assertSubDoc(target, {parentPath: parentDocPath, title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({...treeContext, templatePath: "docName2"});
            assertSubDoc(target, {parentPath: parentDocPath, title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({...treeContext, templatePath: "docName2", name: "docName3"});
            assertSubDoc(target, {parentPath: parentDocPath, title: "docName3"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({...treeContext, templatePath: "parent3/docName2"});
            assertSubDoc(target, {parentPath: parentDocPath, title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({...treeContext, templatePath: "/docName2"});
            assertSubDoc(target, {parentPath: parentDocPath, title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({...treeContext, templatePath: "docName2/"});
            assertSubDoc(target, {parentPath: parentDocPath, title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({...treeContext, templatePath: "docName2/", name: "docName2"});
            assertSubDoc(target, {parentPath: parentDocPath, title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({...treeContext, templatePath: "  docName2  "});
            assertSubDoc(target, {parentPath: parentDocPath, title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({
                currentNotebookId: notebookId,
                currentPath: nestedDocPath,
                templatePath: "docName2",
            });
            assertSubDoc(target, {parentPath: nestedDocPath, title: "docName2"});
        });
    });

    describe("new document target", () => {
        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({
                currentNotebookId: notebookId,
                currentPath: parentDirPath,
                templatePath: "docName2",
            });
            assertSubDoc(target, {parentPath: parentDirPath, title: "docName2"});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({
                currentNotebookId: notebookId,
                currentPath: rootPath,
                templatePath: "",
            });
            assertSubDoc(target, {parentPath: rootPath, title: ""});
        });
    });

    describe("new document target", () => {
        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({
                currentNotebookId: notebookId,
                currentPath: rootPath,
                templatePath: "",
            });
            assertSubDoc(target, {parentPath: rootPath, title: ""});
        });

        it("resolves the target", () => {
            const target = getNewDocTargetFromTree({
                currentNotebookId: notebookId,
                currentPath: rootPath,
                templatePath: "docName2",
                name: "docName3",
            });
            assertSubDoc(target, {parentPath: rootPath, title: "docName3"});
        });
    });
});

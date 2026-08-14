const path = require("path");
const {test, expect} = require("./helpers/playwright");
const {checkBlockExists, createAttributeView, createDoc, getCanvas, insertBlock, launchScribli, newNodeID, openDoc, removeDocByID, updateBlock} = require("./helpers/scribli");

const visibleCanvas = (page) => page.locator(".layout__wnd--active .protyle-wysiwyg .scribli-canvas:visible:has(.scribli-canvas__toolbar)").last();
const visibleEditor = (page) => page.locator(".layout__wnd--active .protyle-wysiwyg:visible").last();
const htmlAttr = (value) => value.replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

const blockIDByText = async (page, text) => page.evaluate((text) => {
    const blocks = Array.from(document.querySelectorAll(".layout__wnd--active .protyle-wysiwyg [data-node-id][data-type]"));
    return blocks.find((block) => (block.textContent || "").includes(text))?.getAttribute("data-node-id") || "";
}, text);

test("Scribli Electron audit covers document creation and Canvas card workflows", async ({}, testInfo) => {
    const scribli = await launchScribli(testInfo, {workspaceArg: false});
    try {
        await expect(scribli.page).toHaveTitle(/Scribli/);
        await expect(visibleEditor(scribli.page)).toBeVisible();

        const smokeTitle = `Smoke ${Date.now()}`;
        const smokeDocID = await createDoc(scribli, smokeTitle, `# ${smokeTitle}\n\nA normal text block for Electron smoke testing.\n`);
        await openDoc(scribli, smokeDocID, smokeTitle);
        await expect(visibleEditor(scribli.page)).toContainText("A normal text block");
        const savedText = `Saved text block ${Date.now()}`;
        const smokeTextID = await blockIDByText(scribli.page, "A normal text block");
        expect(smokeTextID, "expected smoke text block id").toBeTruthy();
        await updateBlock(scribli, smokeTextID, savedText);
        await openDoc(scribli, smokeDocID, smokeTitle);
        await expect(visibleEditor(scribli.page)).toContainText(savedText);

        const deleteTitle = `Delete Lifecycle ${Date.now()}`;
        const deleteDocID = await createDoc(scribli, deleteTitle, `# ${deleteTitle}\n\nThis document should be removed by the Electron audit.\n`);
        expect(await checkBlockExists(scribli, deleteDocID)).toBeTruthy();
        await removeDocByID(scribli, deleteDocID);
        expect(await checkBlockExists(scribli, deleteDocID)).toBeFalsy();

        const created = await scribli.api("/api/canvas/call", {
            action: "create",
            title: "Electron Canvas Audit",
            canvas: {
                scribli: {title: "Electron Canvas Audit"},
                nodes: [{id: "audit-source", type: "text", text: "Audit source", x: 0, y: 0, width: 260, height: 140}],
                edges: [],
            },
        });
        const canvasID = created.structuredContent.id;
        const title = `Canvas Audit ${Date.now()}`;
        const topMarker = `Top text marker ${Date.now()}`;
        const bottomMarker = `Bottom text marker ${Date.now()}`;
        const htmlMarker = `HTML block marker ${Date.now()}`;
        const docID = await createDoc(scribli, title, `# ${title}\n\n${topMarker}\n\n\`\`\`scribli-canvas\n${canvasID}\n\`\`\`\n\n${bottomMarker}\n`);
        await openDoc(scribli, docID, title);

        const topTextID = await blockIDByText(scribli.page, topMarker);
        const bottomTextID = await blockIDByText(scribli.page, bottomMarker);
        expect(topTextID, "expected top text block id").toBeTruthy();
        expect(bottomTextID, "expected bottom text block id").toBeTruthy();

        const topDatabaseID = await newNodeID(scribli);
        const topDatabaseBlockID = await newNodeID(scribli);
        const bottomDatabaseID = await newNodeID(scribli);
        const bottomDatabaseBlockID = await newNodeID(scribli);
        const htmlBlockID = await newNodeID(scribli);
        await createAttributeView(scribli, topDatabaseID, topDatabaseBlockID);
        await createAttributeView(scribli, bottomDatabaseID, bottomDatabaseBlockID);
        await insertBlock(scribli, `<div class="av" data-node-id="${topDatabaseBlockID}" data-av-id="${topDatabaseID}" data-type="NodeAttributeView" data-av-type="table"></div>`, {nextID: topTextID});
        await insertBlock(scribli, `<div data-node-id="${htmlBlockID}" data-type="NodeHTMLBlock" class="render-node" data-subtype="block"><div><protyle-html data-content="${htmlAttr(`<section data-qa="mixed-html">${htmlMarker}</section>`)}"></protyle-html><span style="position: absolute"></span></div><div class="protyle-attr" contenteditable="false"></div></div>`, {previousID: topTextID});
        await insertBlock(scribli, `<div class="av" data-node-id="${bottomDatabaseBlockID}" data-av-id="${bottomDatabaseID}" data-type="NodeAttributeView" data-av-type="table"></div>`, {previousID: bottomTextID});
        await openDoc(scribli, docID, title);

        const canvas = visibleCanvas(scribli.page);
        await expect(scribli.page.locator(`.layout__wnd--active .av[data-av-id="${topDatabaseID}"]:visible`)).toBeVisible();
        await expect(scribli.page.locator(`.layout__wnd--active .av[data-av-id="${bottomDatabaseID}"]:visible`)).toBeVisible();
        await expect(scribli.page.locator(`.layout__wnd--active [data-node-id="${htmlBlockID}"][data-type="NodeHTMLBlock"]:visible`)).toBeVisible();
        await expect(canvas.locator(".scribli-canvas__toolbar")).toBeVisible();
        await expect(canvas.locator(".scribli-canvas__node")).toHaveCount(1);
        await expect(scribli.page.locator(".layout__wnd--active .render-node[data-subtype='scribli-canvas'] > .protyle-icons").last()).toBeHidden();

        const mixedLayout = await scribli.page.evaluate(({bottomDatabaseBlockID, bottomTextID, htmlBlockID, topDatabaseBlockID, topTextID}) => {
            const blockTop = (selector) => {
                const element = document.querySelector(selector);
                return element ? element.getBoundingClientRect().top : null;
            };
            const canvasBlock = document.querySelector(".layout__wnd--active .scribli-canvas")?.closest("[data-node-id][data-type]");
            return {
                bottomDatabase: blockTop(`[data-node-id="${bottomDatabaseBlockID}"]`),
                bottomText: blockTop(`[data-node-id="${bottomTextID}"]`),
                canvas: canvasBlock ? canvasBlock.getBoundingClientRect().top : null,
                html: blockTop(`[data-node-id="${htmlBlockID}"]`),
                topDatabase: blockTop(`[data-node-id="${topDatabaseBlockID}"]`),
                topText: blockTop(`[data-node-id="${topTextID}"]`),
            };
        }, {bottomDatabaseBlockID, bottomTextID, htmlBlockID, topDatabaseBlockID, topTextID});
        expect(mixedLayout.topDatabase).toBeLessThan(mixedLayout.topText);
        expect(mixedLayout.topText).toBeLessThan(mixedLayout.html);
        expect(mixedLayout.html).toBeLessThan(mixedLayout.canvas);
        expect(mixedLayout.canvas).toBeLessThan(mixedLayout.bottomText);
        expect(mixedLayout.bottomText).toBeLessThan(mixedLayout.bottomDatabase);

        const layout = await canvas.evaluate((surface) => {
            const toolbar = surface.querySelector(".scribli-canvas__toolbar").getBoundingClientRect();
            const node = surface.querySelector(".scribli-canvas__node").getBoundingClientRect();
            const buttons = Array.from(surface.querySelectorAll(".scribli-canvas__node-button")).map((button) => {
                const rect = button.getBoundingClientRect();
                return {height: rect.height, left: rect.left, top: rect.top, width: rect.width};
            });
            return {buttons, node: {left: node.left, right: node.right, top: node.top}, toolbar: {bottom: toolbar.bottom, left: toolbar.left, right: toolbar.right, top: toolbar.top}};
        });
        expect(layout.buttons.length).toBe(2);
        for (const button of layout.buttons) {
            expect(button.width).toBeGreaterThanOrEqual(20);
            expect(button.height).toBeGreaterThanOrEqual(20);
            expect(button.left).toBeGreaterThanOrEqual(layout.node.left);
            expect(button.left + button.width).toBeLessThanOrEqual(layout.node.right + 1);
        }
        expect(layout.toolbar.bottom).toBeLessThan(layout.node.top);

        await canvas.locator("[data-type='canvas-node-duplicate']").first().click();
        await expect(canvas.locator(".scribli-canvas__node")).toHaveCount(2);
        await expect.poll(async () => (await getCanvas(scribli, canvasID)).nodes.length).toBe(2);

        await canvas.locator("[data-type='canvas-node-delete']").first().click();
        await expect(canvas.locator(".scribli-canvas__node")).toHaveCount(1);
        await expect.poll(async () => (await getCanvas(scribli, canvasID)).nodes.length).toBe(1);

        const editedText = `Edited source ${Date.now()}`;
        const addedText = `Added source ${Date.now()}`;
        await canvas.locator("[data-type='canvas-edit-text']").first().fill(editedText);
        await canvas.locator("[data-type='canvas-add-text']").click();
        await scribli.page.locator("[data-type='canvas-prompt-input']").fill(addedText);
        await scribli.page.locator("[data-type='canvas-prompt-confirm']").click();
        await expect.poll(async () => {
            const texts = (await getCanvas(scribli, canvasID)).nodes.map((node) => node.text || "").sort();
            return texts.join("|");
        }).toBe([addedText, editedText].sort().join("|"));

        const draggedNode = canvas.locator(".scribli-canvas__node").first();
        const draggedNodeID = await draggedNode.getAttribute("data-node-id");
        const beforeMove = await getCanvas(scribli, canvasID);
        const beforeDraggedNode = beforeMove.nodes.find((item) => item.id === draggedNodeID);
        expect(beforeDraggedNode).toBeTruthy();
        const dragBox = await draggedNode.locator(".scribli-canvas__drag-handle").boundingBox();
        expect(dragBox).toBeTruthy();
        const dragTarget = await scribli.page.evaluate(({x, y}) => {
            const element = document.elementFromPoint(x, y);
            return {
                className: element?.getAttribute("class") || "",
                dataType: element?.getAttribute("data-type") || "",
                isHandle: Boolean(element?.closest(".scribli-canvas__drag-handle")),
                tagName: element?.tagName || "",
            };
        }, {x: dragBox.x + 4, y: dragBox.y + 4});
        expect(dragTarget.isHandle, `drag hit target: ${JSON.stringify(dragTarget)}`).toBeTruthy();
        await draggedNode.locator(".scribli-canvas__drag-handle").evaluate((handle, box) => {
            const eventInit = {bubbles: true, cancelable: true, clientX: box.x + 4, clientY: box.y + 4, view: window};
            handle.dispatchEvent(new MouseEvent("mousedown", eventInit));
            document.dispatchEvent(new MouseEvent("mousemove", {...eventInit, clientX: box.x + 90, clientY: box.y + 70}));
            document.dispatchEvent(new MouseEvent("mouseup", {...eventInit, clientX: box.x + 90, clientY: box.y + 70}));
        }, dragBox);
        try {
            await expect.poll(async () => {
                const node = (await getCanvas(scribli, canvasID)).nodes.find((item) => item.id === draggedNodeID);
                return Boolean(node && (node.x !== beforeDraggedNode.x || node.y !== beforeDraggedNode.y));
            }).toBeTruthy();
        } catch (error) {
            const afterDraggedNode = (await getCanvas(scribli, canvasID)).nodes.find((item) => item.id === draggedNodeID);
            const dragDiagnostics = await draggedNode.evaluate((node) => {
                const rect = node.getBoundingClientRect();
                return {left: rect.left, top: rect.top, styleLeft: node.style.left, styleTop: node.style.top};
            });
            error.message = `${error.message}\nDrag diagnostics: ${JSON.stringify({afterDraggedNode, beforeDraggedNode, dragDiagnostics, draggedNodeID}, undefined, 2)}`;
            throw error;
        }

        const screenshotPath = path.join(scribli.artifactDir, "canvas-audit.png");
        await canvas.screenshot({path: screenshotPath});

        await openDoc(scribli, smokeDocID, smokeTitle);
        await expect(visibleEditor(scribli.page)).toContainText(savedText);
    } finally {
        await scribli.close();
    }
});

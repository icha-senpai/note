import {fetchSyncPost} from "../../util/fetch";
import {escapeHtml} from "../../util/escape";

const EXECUTABLE_LANGUAGES = new Set(["scribli-js", "scribli-sql", "scribli-api", "scribli-chart"]);

export const executableBlockRender = (element: Element) => {
    const executableElements = executableCodeBlocks(element);
    if (executableElements.length === 0) {
        return;
    }
    executableElements.forEach((item) => {
        if (item.getAttribute("data-scribli-executable") === "true") {
            return;
        }
        item.setAttribute("data-scribli-executable", "true");
        ensureExecutableButton(item);
        ensureExecutableOutput(item);
    });
};

const executableCodeBlocks = (element: Element): HTMLElement[] => {
    if (isExecutableCodeBlock(element)) {
        return [element as HTMLElement];
    }
    return Array.from(element.querySelectorAll(".code-block")).filter(isExecutableCodeBlock) as HTMLElement[];
};

const isExecutableCodeBlock = (element: Element) => {
    return element.classList.contains("code-block") && EXECUTABLE_LANGUAGES.has(element.getAttribute("data-subtype") || "");
};

const ensureExecutableButton = (blockElement: HTMLElement) => {
    let actionElement = blockElement.querySelector(".protyle-action");
    if (!actionElement) {
        actionElement = document.createElement("span");
        actionElement.className = "protyle-action protyle-icons";
        blockElement.insertAdjacentElement("afterbegin", actionElement);
    }
    if (actionElement.querySelector(".protyle-action__scribli-run")) {
        return;
    }
    actionElement.insertAdjacentHTML("afterbegin", `<span aria-label="${window.scribli.languages.run}" data-position="4north" class="ariaLabel protyle-icon protyle-icon--first protyle-action__scribli-run"><svg><use xlink:href="#iconPlay"></use></svg></span>`);
    const runElement = actionElement.querySelector(".protyle-action__scribli-run");
    if (!runElement) {
        return;
    }
    runElement.addEventListener("click", (event) => {
        event.preventDefault();
        event.stopPropagation();
        runExecutableBlock(blockElement);
    });
};

const ensureExecutableOutput = (blockElement: HTMLElement) => {
    let outputElement = blockElement.querySelector(".scribli-executable-output") as HTMLElement;
    if (!outputElement) {
        outputElement = document.createElement("div");
        outputElement.className = "scribli-executable-output fn__none";
        outputElement.setAttribute("contenteditable", "false");
        blockElement.appendChild(outputElement);
    }
    return outputElement;
};

const runExecutableBlock = async (blockElement: HTMLElement) => {
    const outputElement = ensureExecutableOutput(blockElement);
    outputElement.classList.remove("fn__none", "ft__error");
    outputElement.innerHTML = "<svg class=\"fn__rotate\"><use xlink:href=\"#iconRefresh\"></use></svg>";
    try {
        const response = await fetchSyncPost("/api/executableBlock/call", executableRequest(blockElement));
        if (response.code !== 0 || response.data?.isError) {
            outputElement.classList.add("ft__error");
            outputElement.textContent = response.msg || window.scribli.languages.executableBlockFailed;
            return;
        }
        renderExecutableOutput(outputElement, response.data?.structuredContent || {});
    } catch (error) {
        outputElement.classList.add("ft__error");
        outputElement.textContent = error instanceof Error ? error.message : window.scribli.languages.executableBlockFailed;
    }
};

const executableRequest = (blockElement: HTMLElement) => {
    const code = codeBlockText(blockElement);
    switch (blockElement.getAttribute("data-subtype")) {
        case "scribli-sql":
            return {action: "run_sql", stmt: code};
        case "scribli-api":
            return executableAPIRequest(code);
        case "scribli-chart":
            return {action: "chart", chartJSON: code};
        default:
            return {action: "run_js", code};
    }
};

const executableAPIRequest = (code: string) => {
    const trimmedCode = code.trim();
    if (trimmedCode.startsWith("/api/")) {
        return {action: "run_api", path: trimmedCode};
    }
    const request = JSON.parse(trimmedCode);
    return {...request, action: "run_api"};
};

const codeBlockText = (blockElement: HTMLElement) => {
    const dataContent = blockElement.getAttribute("data-content");
    if (dataContent) {
        return Lute.UnEscapeHTMLStr(dataContent);
    }
    const codeElement = blockElement.querySelector("code, .hljs div:last-child, .hljs");
    return codeElement?.textContent || "";
};

const renderExecutableOutput = (outputElement: HTMLElement, structured: Record<string, any>) => {
    if (structured.action === "run_sql") {
        renderSQLResult(outputElement, structured);
        return;
    }
    if (structured.action === "chart") {
        outputElement.innerHTML = `<pre><code>${escapeHtml(structured.markdown || "")}</code></pre>`;
        return;
    }
    if (structured.action === "run_api") {
        outputElement.innerHTML = `<pre><code>${escapeHtml(structured.body || "")}</code></pre>`;
        return;
    }
    const logs = Array.isArray(structured.logs) && structured.logs.length > 0 ? `<pre class="ft__secondary">${escapeHtml(structured.logs.join("\n"))}</pre>` : "";
    outputElement.innerHTML = `${logs}<pre><code>${escapeHtml(JSON.stringify(structured.output ?? null, undefined, 2))}</code></pre>`;
};

const renderSQLResult = (outputElement: HTMLElement, structured: Record<string, any>) => {
    const keys = Array.isArray(structured.keys) ? structured.keys : [];
    const rows = Array.isArray(structured.rows) ? structured.rows : [];
    if (keys.length === 0 || rows.length === 0) {
        outputElement.textContent = window.scribli.languages.executableBlockNoResults;
        return;
    }
    const header = keys.map((key) => `<th>${escapeHtml(String(key))}</th>`).join("");
    const body = rows.map((row) => `<tr>${keys.map((key) => `<td>${escapeHtml(String(row[key] ?? ""))}</td>`).join("")}</tr>`).join("");
    outputElement.innerHTML = `<div class="table"><table><thead><tr>${header}</tr></thead><tbody>${body}</tbody></table></div>`;
};

let initialized = false;
const controls = ".block__icon[aria-label], .dock__item[aria-label], .protyle-toolbar__item[aria-label], .protyle-icon[data-type][aria-label], .b3-list-item__action[data-type][aria-label]";

export const initAccessibleControls = () => {
    if (initialized) { return; }
    initialized = true;
    const normalize = (element: HTMLElement) => {
        const label = element.getAttribute("aria-label");
        if (label?.includes("<")) {
            const template = document.createElement("template");
            template.innerHTML = label;
            element.dataset.tooltip = label;
            element.setAttribute("aria-label", template.content.textContent.replace(/\s+/g, " ").trim());
        }
        if (!element.matches("button, input, select, a")) {
            element.setAttribute("role", "button");
            if (!element.hasAttribute("tabindex")) { element.tabIndex = 0; }
        }
    };
    const scan = (root: Element) => {
        if (root.matches(controls)) { normalize(root as HTMLElement); }
        root.querySelectorAll<HTMLElement>(controls).forEach(normalize);
    };
    scan(document.body);
    new MutationObserver(records => {
        records.forEach(record => {
            if (record.type === "attributes") {
                if ((record.target as Element).matches(controls)) { normalize(record.target as HTMLElement); }
            } else {
                record.addedNodes.forEach(node => { if (node instanceof Element) { scan(node); } });
            }
        });
    }).observe(document.body, {childList: true, subtree: true, attributes: true, attributeFilter: ["aria-label"]});
    document.addEventListener("keydown", event => {
        const target = event.target as HTMLElement;
        if ((event.key !== "Enter" && event.key !== " ") || !target.matches(`${controls}`) || target.getAttribute("role") !== "button") { return; }
        if (target.hasAttribute("disabled") || target.getAttribute("aria-disabled") === "true" || target.closest('[data-readonly="true"]')) { return; }
        event.preventDefault();
        event.stopImmediatePropagation();
        target.click();
    }, true);
};

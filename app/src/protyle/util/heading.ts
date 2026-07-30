export const removeFoldHeading = (nodeElement: Element) => {
    const nodeH = parseInt(nodeElement.getAttribute("data-subtype").substr(1));
    let nextElement = nodeElement.nextElementSibling;
    while (nextElement) {
        const currentH = parseInt(nextElement.getAttribute("data-subtype")?.substr(1));
        if (!nextElement.classList.contains("protyle-attr") &&
            (isNaN(currentH) || currentH > nodeH)) {
            const tempElement = nextElement;
            nextElement = nextElement.nextElementSibling;
            tempElement.remove();
        } else {
            break;
        }
    }
};

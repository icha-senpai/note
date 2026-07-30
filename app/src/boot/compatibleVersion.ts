export const img3115 = (imgElement: HTMLElement) => {
    if (imgElement.style.minWidth) {
        imgElement.style.width = "";
    } else {
        imgElement.removeAttribute("style");
    }
};

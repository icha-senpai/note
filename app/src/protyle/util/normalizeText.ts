// 
export const nbsp2space = (text: string) => {
    return text.replace(/\u00A0/g, " ");
};

// 
export const removeZWJ = (text: string) => {
    return text.replace(/\u200D```/g, "```");
};

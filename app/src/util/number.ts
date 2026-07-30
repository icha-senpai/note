export const formatCount = (n: number | string) => {
    const num = typeof n === "string" ? parseFloat(n) : n;
    if (!Number.isFinite(num)) {
        return n;
    }
    if (num < 1000) {
        return n.toString();
    }
    let value: number;
    let suffix: string;
    if (num < 1000000) {
        value = num / 1000;
        suffix = "k";
    } else {
        value = num / 1000000;
        suffix = "M";
    }
    return (value.toFixed(1).replace(/\.0$/, "")) + suffix;
};

export const preventScroll = (protyle: IProtyle, scrollTop = 0, timeout = 1000) => {
    protyle.scroll.lastScrollTop = -1;
    setTimeout(() => {
        protyle.scroll.lastScrollTop = scrollTop;
    }, timeout);
};

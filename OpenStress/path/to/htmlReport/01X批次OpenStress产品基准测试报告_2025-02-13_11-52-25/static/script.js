
document.addEventListener("DOMContentLoaded", function() {
    const iframe = document.querySelector('.tps-chart');
    
    function adjustIframeHeight() {
        const iframeDocument = iframe.contentDocument || iframe.contentWindow.document;
        const body = iframeDocument.body;
        const html = iframeDocument.documentElement;

        // 获取整个文档的高度
        const docHeight = Math.max(
            body.scrollHeight, body.offsetHeight,
            html.clientHeight, html.scrollHeight, html.offsetHeight
        );
        
        // 设置iframe的高度
        iframe.style.height = docHeight + 'px';
    }

    // 初始化时调整iframe高度
    adjustIframeHeight();

    // 监听iframe内容变化，调整高度
    const observer = new MutationObserver(adjustIframeHeight);
    observer.observe(iframe.contentDocument || iframe.contentWindow.document, {
        childList: true,
        subtree: true,
        attributes: true
    });
});

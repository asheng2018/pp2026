/**
 * AB Payment Bridge v2 — Frontend SDK
 * Loaded on every page. Only activates when ab_pay URL param is present.
 */
(function($, AB) {
    'use strict';

    // Check if ab_pay param exists in current URL
    var params = new URLSearchParams(window.location.search);
    var payToken = params.get('ab_pay');
    if (!payToken) return; // No payment hash, do nothing

    var orderId   = params.get('ab_order') || '';
    var amount    = params.get('ab_amount') || '';
    var gateway   = params.get('ab_gateway') || AB.gateway;

    // 1. Hide WooCommerce default content
    $('.woocommerce').css({visibility:'hidden', height:0, overflow:'hidden'});
    $('header, footer, .site-header, .site-footer').hide();

    // 2. Create payment iframe
    console.log('[AB Bridge] Loading payment for order:', orderId);

    var iframe = document.createElement('iframe');
    iframe.id   = 'ab-pay-iframe';
    var bSite   = AB.bSite.replace(/\/+$/, '');

    // Construct B-site payment URL
    iframe.src    = bSite + '/pay/' + encodeURIComponent(payToken);
    iframe.setAttribute('sandbox', 'allow-scripts allow-forms allow-same-origin allow-top-navigation allow-popups');
    iframe.setAttribute('referrerpolicy', 'no-referrer');
    iframe.setAttribute('allow', 'payment');
    iframe.style.cssText = 'position:fixed;top:0;left:0;width:100%;height:100%;border:none;z-index:2147483647;background:#fff;';
    document.body.appendChild(iframe);

    // 3. Listen for payment result from B-site iframe
    window.addEventListener('message', function(ev) {
        // Verify origin is B-site
        if (ev.origin.indexOf(AB.bSite.replace(/^https?:\/\//,'')) === -1) return;

        var msg = ev.data || {};
        switch (msg.type) {
            case 'IFRAME_READY':
                console.log('[AB Bridge] B-site ready');
                break;

            case 'PAYMENT_COMPLETED':
                console.log('[AB Bridge] Payment complete');
                // Redirect to order received page
                window.location.href = AB.bSite + '/checkout/order-received/' + orderId + '/?key=' + (msg.data.orderKey || '');
                break;

            case 'PAYMENT_FAILED':
                console.log('[AB Bridge] Payment failed:', msg.data);
                iframe.remove();
                $('.woocommerce').css({visibility:'visible', height:'auto', overflow:'visible'});
                $('header, footer').show();
                alert('Payment was not completed. Please try again.');
                // Remove ab_pay from URL so user can retry
                history.replaceState(null, '', window.location.pathname);
                break;

            case 'PAYMENT_CANCELED':
                console.log('[AB Bridge] Payment canceled');
                iframe.remove();
                $('.woocommerce').css({visibility:'visible', height:'auto', overflow:'visible'});
                $('header, footer').show();
                history.replaceState(null, '', window.location.pathname);
                break;
        }
    });

})(jQuery, window.AB || {});

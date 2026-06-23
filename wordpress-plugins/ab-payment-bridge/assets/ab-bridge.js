/**
 * AB Payment Bridge - Frontend SDK
 *
 * Loaded on the A-site WooCommerce checkout page.
 * Detects the ab_pay URL parameter, creates a sandboxed iframe
 * to the B-site payment page, and handles the postMessage-based
 * communication between A-site and B-site.
 *
 * @version 1.0.0
 */

(function($) {
    'use strict';

    // ==============================================================
    // URL Parameter Helper
    // ==============================================================
    function getUrlParam(name) {
        var params = new URLSearchParams(window.location.search);
        return params.get(name);
    }

    // ==============================================================
    // Main Payment Bridge
    // ==============================================================
    var ABPaymentBridge = {
        payToken: null,
        gateway: null,
        orderId: null,
        iframe: null,
        overlay: null,
        container: null,

        /**
         * Initialize the payment bridge
         */
        init: function() {
            this.payToken = getUrlParam('ab_pay');
            this.gateway  = getUrlParam('ab_gateway');
            this.orderId  = getUrlParam('ab_order');

            if (!this.payToken) {
                return; // Not a payment redirect, normal checkout
            }

            console.log('[AB Bridge] Initializing payment for order:', this.orderId);

            // Hide WooCommerce default content
            this.hideWooContent();

            // Create payment overlay
            this.createOverlay();

            // Open B-site payment in iframe
            this.openPaymentIframe();

            // Listen for payment results
            this.listenForMessages();
        },

        /**
         * Hide WooCommerce checkout elements
         */
        hideWooContent: function() {
            // Hide the main content but keep the structure
            $('.woocommerce').css({
                'visibility': 'hidden',
                'height': '0',
                'overflow': 'hidden'
            });

            // Also hide common theme wrappers
            $('header, footer, .site-header, .site-footer').hide();
        },

        /**
         * Create a full-screen overlay for the payment iframe
         */
        createOverlay: function() {
            this.overlay = document.createElement('div');
            this.overlay.id = 'ab-payment-overlay';
            this.overlay.style.cssText = [
                'position: fixed',
                'top: 0',
                'left: 0',
                'width: 100%',
                'height: 100%',
                'z-index: 2147483646',
                'background: #fff',
                'display: flex',
                'align-items: center',
                'justify-content: center',
            ].join(';');

            // Loading indicator
            this.overlay.innerHTML = [
                '<div id="ab-loading" style="text-align:center;font-family:-apple-system,sans-serif;">',
                '  <div style="border:3px solid #e5e7eb;border-top-color:#3b82f6;border-radius:50%;',
                '              width:40px;height:40px;animation:ab-spin 1s linear infinite;margin:0 auto 16px;"></div>',
                '  <p style="color:#666;font-size:16px;">Redirecting to secure payment...</p>',
                '</div>',
                '<style>@keyframes ab-spin{to{transform:rotate(360deg)}}</style>',
            ].join('');

            document.body.appendChild(this.overlay);
        },

        /**
         * Create and load the payment iframe from B-site
         */
        openPaymentIframe: function() {
            var self = this;
            var bSiteDomain = AB_CONFIG.bSiteDomain || '';
            var payUrl = bSiteDomain.replace(/\/+$/, '') + '/pay/' + encodeURIComponent(self.payToken);

            console.log('[AB Bridge] Loading payment page:', payUrl);

            this.iframe = document.createElement('iframe');
            this.iframe.id = 'ab-payment-iframe';
            this.iframe.src = payUrl;
            this.iframe.setAttribute('sandbox',
                'allow-scripts allow-forms allow-same-origin allow-top-navigation allow-popups');
            this.iframe.setAttribute('referrerpolicy', 'no-referrer');
            this.iframe.setAttribute('allow', 'payment');
            this.iframe.style.cssText = [
                'position: fixed',
                'top: 0',
                'left: 0',
                'width: 100%',
                'height: 100%',
                'border: none',
                'z-index: 2147483647',
                'background: #fff',
            ].join(';');

            // Remove overlay once iframe loads
            this.iframe.onload = function() {
                var loadingEl = document.getElementById('ab-loading');
                if (loadingEl) {
                    loadingEl.style.display = 'none';
                }
                console.log('[AB Bridge] Payment iframe loaded');
            };

            document.body.appendChild(this.iframe);
        },

        /**
         * Listen for postMessage from B-site payment iframe
         */
        listenForMessages: function() {
            var self = this;

            window.addEventListener('message', function(event) {
                var bSiteDomain = AB_CONFIG.bSiteDomain || '';

                // Verify the message origin
                if (!event.origin || bSiteDomain.indexOf(event.origin) === -1) {
                    // Also check if the origin matches the B-site domain
                    var bSiteHost = (function(url) {
                        try { return (new URL(url)).host; } catch(e) { return ''; }
                    })(bSiteDomain);

                    if (event.origin.indexOf(bSiteHost) === -1) {
                        return; // Not from our B-site, ignore
                    }
                }

                var msg = event.data || {};
                console.log('[AB Bridge] Message received:', msg.type);

                switch (msg.type) {
                    case 'IFRAME_READY':
                        self.onIframeReady(msg.data || {});
                        break;

                    case 'PAYMENT_COMPLETED':
                        self.onPaymentSuccess(msg.data || {});
                        break;

                    case 'PAYMENT_FAILED':
                        self.onPaymentFailed(msg.data || {});
                        break;

                    case 'PAYMENT_CANCELED':
                        self.onPaymentCanceled(msg.data || {});
                        break;

                    case 'IFRAME_RESIZE':
                        self.onIframeResize(msg.data || {});
                        break;
                }
            }, false);
        },

        /**
         * Called when B-site iframe is ready
         */
        onIframeReady: function(data) {
            console.log('[AB Bridge] B-site iframe ready');
            var loadingEl = document.getElementById('ab-loading');
            if (loadingEl) {
                loadingEl.style.display = 'none';
            }
        },

        /**
         * Handle successful payment
         */
        onPaymentSuccess: function(data) {
            console.log('[AB Bridge] Payment completed:', data);
            var self = this;

            // Notify server to update order
            $.ajax({
                url: AB_CONFIG.ajaxUrl,
                method: 'POST',
                data: {
                    action: 'ab_payment_callback',
                    order_id: self.orderId,
                    pay_token: self.payToken,
                    gateway_tx_id: data.gatewayOrderId || '',
                    nonce: AB_CONFIG.nonce
                },
                success: function(response) {
                    if (response.success) {
                        // Show success briefly then redirect
                        self.showResult('success',
                            'Payment Successful!',
                            'Redirecting to order confirmation...');

                        setTimeout(function() {
                            window.location.href = response.data.redirect_url
                                || (AB_CONFIG.wooCheckoutUrl +
                                   '/order-received/' + self.orderId +
                                   '/?key=' + (data.orderKey || ''));
                        }, 2000);
                    } else {
                        self.showResult('error',
                            'Error',
                            'Payment was processed but we could not confirm the order. Please contact support.');
                    }
                },
                error: function() {
                    self.showResult('success',
                        'Payment Successful!',
                        'Redirecting to order confirmation...');
                    setTimeout(function() {
                        window.location.href = AB_CONFIG.wooCheckoutUrl +
                            '/order-received/' + self.orderId + '/';
                    }, 2000);
                }
            });
        },

        /**
         * Handle failed payment
         */
        onPaymentFailed: function(data) {
            console.log('[AB Bridge] Payment failed:', data);
            var reason = data.reason || data.message || 'Payment was not completed.';

            this.showResult('error', 'Payment Failed', reason, function() {
                // Go back to checkout to try again
                window.location.href = AB_CONFIG.wooCheckoutUrl;
            });
        },

        /**
         * Handle canceled payment
         */
        onPaymentCanceled: function(data) {
            console.log('[AB Bridge] Payment canceled');
            this.showResult('warning', 'Payment Canceled',
                'You canceled the payment. You can try again.',
                function() {
                    window.location.href = AB_CONFIG.wooCheckoutUrl;
                });
        },

        /**
         * Handle iframe resize request
         */
        onIframeResize: function(data) {
            if (this.iframe && data.height) {
                this.iframe.style.height = data.height + 'px';
            }
        },

        /**
         * Show a result overlay with message
         */
        showResult: function(type, title, message, callback) {
            // Remove existing overlay and iframe
            if (this.overlay) { this.overlay.remove(); }
            if (this.iframe) { this.iframe.remove(); }

            var colors = {
                success: { bg: '#ecfdf5', border: '#10b981', icon: '✓', iconColor: '#10b981' },
                error:   { bg: '#fef2f2', border: '#ef4444', icon: '✗', iconColor: '#ef4444' },
                warning: { bg: '#fffbeb', border: '#f59e0b', icon: '⚠', iconColor: '#f59e0b' },
            };
            var c = colors[type] || colors.success;

            var container = document.createElement('div');
            container.id = 'ab-payment-result';
            container.style.cssText = [
                'position: fixed', 'top: 0', 'left: 0', 'width: 100%', 'height: 100%',
                'z-index: 2147483647', 'background: #f9fafb',
                'display: flex', 'align-items: center', 'justify-content: center',
                'font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
            ].join(';');

            container.innerHTML = [
                '<div style="background:#fff;padding:40px;border-radius:12px;box-shadow:0 4px 24px rgba(0,0,0,0.08);',
                '            max-width:420px;width:90%;text-align:center;border-top:4px solid ' + c.border + '">',
                '  <div style="width:64px;height:64px;border-radius:50%;background:' + c.bg + ';',
                '              display:flex;align-items:center;justify-content:center;margin:0 auto 20px;',
                '              font-size:32px;color:' + c.iconColor + '">' + c.icon + '</div>',
                '  <h2 style="margin:0 0 8px;font-size:22px;">' + title + '</h2>',
                '  <p style="color:#666;margin-bottom:24px;">' + message + '</p>',
                callback ? '<button id="ab-result-btn" style="background:' + c.border + ';color:#fff;border:none;',
                '             padding:12px 32px;border-radius:6px;font-size:16px;cursor:pointer;">OK</button>' : '',
                '</div>',
            ].join('');

            document.body.appendChild(container);

            if (callback) {
                document.getElementById('ab-result-btn').addEventListener('click', callback);
            }
        },

        /**
         * Clean up everything
         */
        destroy: function() {
            if (this.iframe) { this.iframe.remove(); this.iframe = null; }
            if (this.overlay) { this.overlay.remove(); this.overlay = null; }
            var resultEl = document.getElementById('ab-payment-result');
            if (resultEl) { resultEl.remove(); }
        }
    };

    // ==============================================================
    // Initialize on DOM ready
    // ==============================================================
    $(document).ready(function() {
        ABPaymentBridge.init();
    });

    // Expose to global scope
    window.ABPaymentBridge = ABPaymentBridge;

})(jQuery);

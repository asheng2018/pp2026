<?php
/**
 * B-Site Payment Page Template
 *
 * This page is loaded inside an iframe on the A-site.
 * It loads the PayPal or Stripe SDK from the B-site domain,
 * so all tracking/reporting goes through the B-site.
 *
 * IMPORTANT: This page MUST be served from the B-site domain with HTTPS.
 */

if (!defined('ABSPATH')) exit;

$token = get_query_var('ab_pay_token');
if (!$token) {
    http_response_code(400);
    die('Missing payment token');
}

// Extract gateway info from token
// Token format: base64(json).hmac_sig
$parts = explode('.', $token);
$gateway = 'paypal'; // Default
$amount = '0.00';
$currency = 'USD';
$order_ref = '';

if (count($parts) === 2) {
    $payload = json_decode(base64_decode(strtr($parts[0], '-_', '+/')), true);
    if ($payload) {
        $gateway   = $payload['gw'] ?? 'paypal';
        $amount    = $payload['amt'] ?? '0.00';
        $currency  = $payload['cur'] ?? 'USD';
        $order_ref = $payload['oid'] ?? '';
    }
}

$b_site_name = get_bloginfo('name');
$b_site_domain = parse_url(home_url(), PHP_URL_HOST);
?><!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <meta name="robots" content="noindex, nofollow, noarchive">
    <meta name="referrer" content="no-referrer">
    <meta name="google" content="notranslate">
    <meta http-equiv="Content-Security-Policy" content="
        default-src 'self';
        script-src 'self' 'unsafe-inline' https://www.paypal.com https://www.paypalobjects.com https://js.stripe.com https://*.paypal.com https://*.stripe.com;
        frame-src https://www.paypal.com https://js.stripe.com https://hooks.stripe.com https://*.paypal.com;
        connect-src 'self' https://api.paypal.com https://api-m.paypal.com https://api.stripe.com https://*.paypal.com;
        img-src 'self' data: https://www.paypalobjects.com https://*.stripe.com https://*.paypal.com;
        style-src 'self' 'unsafe-inline';
        font-src 'self' data:;
    ">
    <title>Checkout — <?php echo esc_html($b_site_name); ?></title>
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            -webkit-font-smoothing: antialiased;
            -moz-osx-font-smoothing: grayscale;
            background: #f7f7f7;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
        }

        .pay-header {
            background: #fff;
            padding: 16px 24px;
            border-bottom: 1px solid #e5e7eb;
            text-align: center;
        }
        .pay-header .logo {
            font-size: 20px;
            font-weight: 700;
            color: #1f2937;
            text-decoration: none;
        }
        .pay-header .secure-badge {
            font-size: 12px;
            color: #9ca3af;
            margin-top: 4px;
        }
        .secure-badge .lock { color: #10b981; }

        .pay-main {
            flex: 1;
            display: flex;
            align-items: flex-start;
            justify-content: center;
            padding: 24px 16px;
        }
        .pay-card {
            background: #fff;
            border-radius: 12px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.08), 0 4px 16px rgba(0,0,0,0.04);
            width: 100%;
            max-width: 480px;
            overflow: hidden;
        }
        .pay-card .order-summary {
            padding: 20px 24px;
            background: #f9fafb;
            border-bottom: 1px solid #f3f4f6;
        }
        .pay-card .order-summary h3 {
            font-size: 16px;
            font-weight: 600;
            color: #374151;
            margin-bottom: 12px;
        }
        .order-row {
            display: flex;
            justify-content: space-between;
            padding: 6px 0;
            font-size: 14px;
        }
        .order-row .label { color: #6b7280; }
        .order-row .value { color: #1f2937; font-weight: 500; }
        .order-row.total {
            border-top: 1px solid #e5e7eb;
            margin-top: 8px;
            padding-top: 12px;
            font-size: 16px;
            font-weight: 600;
        }

        .pay-card .payment-area {
            padding: 24px;
        }

        #loading {
            text-align: center;
            padding: 40px 20px;
            color: #6b7280;
        }
        .spinner {
            width: 36px; height: 36px;
            border: 3px solid #e5e7eb;
            border-top-color: #3b82f6;
            border-radius: 50%;
            animation: spin 0.8s linear infinite;
            margin: 0 auto 16px;
        }
        @keyframes spin { to { transform: rotate(360deg); } }

        #error-msg {
            display: none;
            background: #fef2f2;
            color: #dc2626;
            padding: 12px 16px;
            border-radius: 8px;
            margin: 16px 0;
            font-size: 14px;
            text-align: center;
        }

        #paypal-button-container { min-height: 150px; }
        #stripe-payment-element { min-height: 100px; }

        .pay-footer {
            text-align: center;
            padding: 12px;
            color: #9ca3af;
            font-size: 11px;
        }
        .pay-footer span { margin: 0 8px; }
    </style>
</head>
<body>
    <div class="pay-header">
        <div class="logo"><?php echo esc_html($b_site_name); ?></div>
        <div class="secure-badge">
            <span class="lock">🔒</span> Secure Checkout — Encrypted Connection
        </div>
    </div>

    <div class="pay-main">
        <div class="pay-card">
            <div class="order-summary">
                <h3>Order Summary</h3>
                <div class="order-row">
                    <span class="label">Amount</span>
                    <span class="value"><?php echo esc_html($currency . ' ' . $amount); ?></span>
                </div>
                <div class="order-row">
                    <span class="label">Order Ref</span>
                    <span class="value"><?php echo esc_html(substr($order_ref, 0, 16)); ?></span>
                </div>
                <div class="order-row total">
                    <span class="label">Total</span>
                    <span class="value"><?php echo esc_html($currency . ' ' . $amount); ?></span>
                </div>
            </div>

            <div class="payment-area">
                <div id="loading">
                    <div class="spinner"></div>
                    <p>Initializing secure payment gateway...</p>
                </div>
                <div id="error-msg"></div>
                <!-- PayPal button renders here -->
                <div id="paypal-button-container"></div>
                <!-- Stripe Elements render here -->
                <div id="stripe-payment-element"></div>
            </div>
        </div>
    </div>

    <div class="pay-footer">
        <span>🔒 SSL Encrypted</span>
        <span>PCI DSS Compliant</span>
        <span>© <?php echo date('Y'); ?> <?php echo esc_html($b_site_name); ?></span>
    </div>

    <script>
    (function() {
        'use strict';

        var PAY_TOKEN    = <?php echo json_encode($token); ?>;
        var GATEWAY      = <?php echo json_encode($gateway); ?>;
        var AMOUNT       = <?php echo json_encode($amount); ?>;
        var CURRENCY     = <?php echo json_encode($currency); ?>;
        var ORDER_REF    = <?php echo json_encode($order_ref); ?>;
        var B_SITE_URL   = <?php echo json_encode(home_url()); ?>;
        var ADMIN_AJAX   = <?php echo json_encode(admin_url('admin-ajax.php')); ?>;
        var B_SITE_DOMAIN = '<?php echo esc_js($b_site_domain); ?>';

        var loadingEl = document.getElementById('loading');
        var errorEl   = document.getElementById('error-msg');

        // ==============================================================
        // Notify parent (A-site) iframe is ready
        // ==============================================================
        function notifyParent(type, data) {
            try {
                window.parent.postMessage({ type: type, data: data || {} }, '*');
            } catch(e) {
                console.log('[AB Receiver] postMessage failed:', e);
            }
        }

        function showError(msg) {
            loadingEl.style.display = 'none';
            errorEl.textContent = msg;
            errorEl.style.display = 'block';
            notifyParent('PAYMENT_FAILED', { reason: msg });
        }

        function hideLoading() {
            loadingEl.style.display = 'none';
        }

        // ==============================================================
        // Handle successful payment
        // ==============================================================
        function onPaymentApproved(data) {
            console.log('[AB Receiver] Payment approved:', data);
            notifyParent('PAYMENT_COMPLETED', {
                orderId: ORDER_REF,
                gatewayOrderId: data.orderID || data.id || '',
                gateway: GATEWAY,
                amount: AMOUNT,
                currency: CURRENCY
            });

            // Redirect to success page within iframe
            setTimeout(function() {
                window.location.href = B_SITE_URL + '/pay/success?order_id=' +
                    encodeURIComponent(ORDER_REF || (data.orderID || ''));
            }, 1000);
        }

        function onPaymentFailed(data) {
            console.log('[AB Receiver] Payment failed:', data);
            notifyParent('PAYMENT_FAILED', {
                orderId: ORDER_REF,
                reason: (data && data.message) || 'Payment was declined',
                gateway: GATEWAY
            });
            showError('Payment was not completed. Please try again.');
        }

        function onPaymentCanceled() {
            console.log('[AB Receiver] Payment canceled');
            notifyParent('PAYMENT_CANCELED', { orderId: ORDER_REF });
        }

        // ==============================================================
        // Initialize payment gateway
        // ==============================================================
        function initPayment() {
            console.log('[AB Receiver] Initializing', GATEWAY, 'gateway');

            if (GATEWAY === 'paypal') {
                initPayPal();
            } else if (GATEWAY === 'stripe') {
                initStripe();
            } else {
                showError('Unsupported payment gateway: ' + GATEWAY);
            }
        }

        // ==============================================================
        // PayPal Integration
        // ==============================================================
        function initPayPal() {
            // Get PayPal client ID from B-site WordPress option (embedded in page)
            // We fetch it via AJAX from the B-site
            fetch(ADMIN_AJAX + '?action=ab_get_paypal_config')
                .then(function(r) { return r.json(); })
                .then(function(config) {
                    if (!config.data || !config.data.client_id) {
                        showError('PayPal configuration not available');
                        return;
                    }

                    var clientId = config.data.client_id;
                    var currency = CURRENCY || 'USD';

                    // Load PayPal SDK from B-site domain
                    var script = document.createElement('script');
                    script.src = 'https://www.paypal.com/sdk/js?client-id=' + clientId +
                        '&currency=' + currency +
                        '&intent=capture' +
                        '&commit=true';
                    script.onload = function() {
                        renderPayPalButtons(clientId);
                    };
                    script.onerror = function() {
                        showError('Failed to load PayPal. Please check your connection.');
                    };
                    document.head.appendChild(script);
                })
                .catch(function(err) {
                    showError('Failed to initialize PayPal: ' + err.message);
                });
        }

        function renderPayPalButtons(clientId) {
            hideLoading();

            if (typeof paypal === 'undefined') {
                showError('PayPal SDK failed to load');
                return;
            }

            paypal.Buttons({
                style: {
                    layout: 'vertical',
                    color: 'gold',
                    shape: 'rect',
                    label: 'paypal',
                    tagline: false,
                },
                createOrder: function(data, actions) {
                    // Call B-site API to create PayPal order
                    return fetch(ADMIN_AJAX + '?action=ab_create_paypal_order', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            amount: AMOUNT,
                            currency: CURRENCY || 'USD',
                            order_ref: ORDER_REF
                        })
                    })
                    .then(function(r) { return r.json(); })
                    .then(function(res) {
                        if (res.data && res.data.order_id) {
                            return res.data.order_id;
                        }
                        throw new Error(res.data && res.data.message || 'Failed to create order');
                    });
                },
                onApprove: function(data, actions) {
                    onPaymentApproved({ orderID: data.orderID });
                },
                onCancel: function() {
                    onPaymentCanceled();
                },
                onError: function(err) {
                    onPaymentFailed({ message: err.message || 'PayPal error' });
                }
            }).render('#paypal-button-container');
        }

        // ==============================================================
        // Stripe Integration
        // ==============================================================
        function initStripe() {
            fetch(ADMIN_AJAX + '?action=ab_get_stripe_config')
                .then(function(r) { return r.json(); })
                .then(function(config) {
                    if (!config.data || !config.data.publishable_key) {
                        showError('Stripe configuration not available');
                        return;
                    }

                    var stripe = Stripe(config.data.publishable_key);

                    // Create PaymentIntent via B-site API
                    return fetch(ADMIN_AJAX + '?action=ab_create_stripe_intent', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            amount: AMOUNT,
                            currency: (CURRENCY || 'USD').toLowerCase(),
                            order_ref: ORDER_REF
                        })
                    })
                    .then(function(r) { return r.json(); })
                    .then(function(res) {
                        if (!res.data || !res.data.client_secret) {
                            throw new Error('Failed to create payment intent');
                        }

                        var clientSecret = res.data.client_secret;
                        var stripeInstance = stripe;
                        hideLoading();

                        var elements = stripeInstance.elements({
                            clientSecret: clientSecret,
                            appearance: { theme: 'stripe' }
                        });
                        var paymentElement = elements.create('payment', {
                            layout: { type: 'tabs', defaultCollapsed: false }
                        });
                        paymentElement.mount('#stripe-payment-element');

                        // Handle form submission
                        var payButton = document.createElement('button');
                        payButton.textContent = 'Pay ' + (CURRENCY||'USD') + ' ' + AMOUNT;
                        payButton.style.cssText = 'width:100%;padding:14px;background:#635bff;color:#fff;' +
                            'border:none;border-radius:8px;font-size:16px;font-weight:600;' +
                            'cursor:pointer;margin-top:16px;';
                        payButton.onclick = function() {
                            payButton.disabled = true;
                            payButton.textContent = 'Processing...';
                            stripeInstance.confirmPayment({
                                elements: elements,
                                confirmParams: {
                                    return_url: B_SITE_URL + '/pay/success?order_id=' +
                                        encodeURIComponent(ORDER_REF),
                                },
                                redirect: 'if_required'
                            })
                            .then(function(result) {
                                if (result.error) {
                                    onPaymentFailed({ message: result.error.message });
                                    payButton.disabled = false;
                                    payButton.textContent = 'Pay ' + (CURRENCY||'USD') + ' ' + AMOUNT;
                                } else if (result.paymentIntent &&
                                           result.paymentIntent.status === 'succeeded') {
                                    onPaymentApproved({
                                        orderID: result.paymentIntent.id
                                    });
                                }
                            });
                        };
                        document.getElementById('stripe-payment-element').appendChild(payButton);
                    });
                })
                .catch(function(err) {
                    showError('Failed to initialize Stripe: ' + err.message);
                });
        }

        // ==============================================================
        // Start
        // ==============================================================
        // Notify parent immediately
        notifyParent('IFRAME_READY', { token: PAY_TOKEN, gateway: GATEWAY });

        // Initialize payment after a brief delay to ensure DOM is ready
        setTimeout(initPayment, 300);

    })();
    </script>
</body>
</html>

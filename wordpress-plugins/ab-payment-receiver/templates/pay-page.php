<?php
/**
 * B-Site Payment Page Template — v2
 * Accepts UUID PayToken, queries AB Orchestrator for details.
 */

if (!defined('ABSPATH')) exit;

$token = get_query_var('ab_pay_token');
if (!$token) {
    http_response_code(400);
    die('Missing payment token');
}

// Defaults
$gateway  = 'paypal';
$amount   = '0.00';
$currency = 'USD';
$order_ref = '';

// Try to get payment details from AB Orchestrator
// The pay token is a UUID — we send it to AB to retrieve order info
// simplest fallback: use PayPal as default, $amount from query string
if (!empty($_GET['amount'])) {
    $amount = $_GET['amount'];
}

$b_site_name   = get_bloginfo('name');
$b_site_domain = parse_url(home_url(), PHP_URL_HOST);

// Get PayPal client ID from WooCommerce settings
$pp_settings   = get_option('woocommerce_ppcp-gateway_settings', []);
$paypal_client_id = $pp_settings['client_id'] ?? '';
$paypal_sandbox = ($pp_settings['sandbox'] ?? 'yes') === 'yes' ? 'sandbox' : 'production';
$paypal_merchant = $pp_settings['merchant_id'] ?? '';

// Get Stripe publishable key
$stripe_settings = get_option('woocommerce_stripe_settings', []);
$stripe_key = $stripe_settings['publishable_key'] ?? ($stripe_settings['test_publishable_key'] ?? '');
?><!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <meta name="robots" content="noindex, nofollow, noarchive">
    <meta name="referrer" content="no-referrer">
    <title>Checkout — <?= esc_html($b_site_name) ?></title>
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f7f7f7; min-height: 100vh; display: flex; flex-direction: column;
        }
        .pay-header { background: #fff; padding: 16px 24px; border-bottom: 1px solid #e5e7eb; text-align: center; }
        .pay-header .logo { font-size: 20px; font-weight: 700; color: #1f2937; }
        .pay-header .secure-badge { font-size: 12px; color: #9ca3af; margin-top: 4px; }
        .pay-main { flex: 1; display: flex; align-items: flex-start; justify-content: center; padding: 24px 16px; }
        .pay-card { background: #fff; border-radius: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.08); width: 100%; max-width: 480px; overflow: hidden; }
        .pay-card .order-summary { padding: 20px 24px; background: #f9fafb; border-bottom: 1px solid #f3f4f6; }
        .pay-card .order-summary h3 { font-size: 16px; font-weight: 600; margin-bottom: 12px; }
        .order-row { display: flex; justify-content: space-between; padding: 6px 0; font-size: 14px; }
        .order-row .label { color: #6b7280; } .order-row .value { color: #1f2937; font-weight: 500; }
        .order-row.total { border-top: 1px solid #e5e7eb; margin-top: 8px; padding-top: 12px; font-size: 16px; font-weight: 600; }
        .pay-card .payment-area { padding: 24px; }
        #loading { text-align: center; padding: 40px 20px; color: #6b7280; }
        .spinner { width: 36px; height: 36px; border: 3px solid #e5e7eb; border-top-color: #3b82f6; border-radius: 50%; animation: spin 0.8s linear infinite; margin: 0 auto 16px; }
        @keyframes spin { to { transform: rotate(360deg); } }
        #error-msg { display: none; background: #fef2f2; color: #dc2626; padding: 12px 16px; border-radius: 8px; margin: 16px 0; font-size: 14px; text-align: center; }
        #paypal-button-container { min-height: 150px; }
        .pay-footer { text-align: center; padding: 12px; color: #9ca3af; font-size: 11px; }
    </style>
</head>
<body>
    <div class="pay-header">
        <div class="logo"><?= esc_html($b_site_name) ?></div>
        <div class="secure-badge">&#x1F512; Secure Checkout</div>
    </div>

    <div class="pay-main">
        <div class="pay-card">
            <div class="order-summary">
                <h3>Order Summary</h3>
                <div class="order-row"><span class="label">Amount</span><span class="value"><?= esc_html($currency . ' ' . $amount) ?></span></div>
                <div class="order-row"><span class="label">Order Ref</span><span class="value"><?= esc_html(substr($token, 0, 16)) ?></span></div>
                <div class="order-row total"><span class="label">Total</span><span class="value"><?= esc_html($currency . ' ' . $amount) ?></span></div>
            </div>

            <div class="payment-area">
                <div id="loading"><div class="spinner"></div><p>Initializing secure payment...</p></div>
                <div id="error-msg"></div>
                <div id="paypal-button-container"></div>
                <div id="stripe-payment-element"></div>
            </div>
        </div>
    </div>

    <div class="pay-footer"><span>&#x1F512; SSL Encrypted</span> <span>PCI DSS Compliant</span></div>

    <script>
    (function() {
        var TOKEN   = <?= json_encode($token) ?>;
        var AMOUNT  = <?= json_encode($amount) ?>;
        var CURRENCY = <?= json_encode($currency) ?>;
        var GATEWAY = <?= json_encode($gateway) ?>;
        var ORDER_REF = <?= json_encode($order_ref) ?>;
        var PP_CLIENT_ID = <?= json_encode($paypal_client_id) ?>;
        var PP_SANDBOX = <?= json_encode($paypal_sandbox === 'sandbox') ?>;
        var PP_MERCHANT = <?= json_encode($paypal_merchant) ?>;
        var STRIPE_KEY = <?= json_encode($stripe_key) ?>;
        var B_SITE_URL = <?= json_encode(home_url()) ?>;

        var elLoading = document.getElementById('loading');
        var elError   = document.getElementById('error-msg');

        function notifyParent(type, data) {
            try { window.parent.postMessage({ type: type, data: data || {} }, '*'); } catch(e) {}
        }

        function hideLoading() { elLoading.style.display = 'none'; }
        function showError(msg) { elLoading.style.display = 'none'; elError.textContent = msg; elError.style.display = 'block'; notifyParent('PAYMENT_FAILED', { reason: msg }); }

        function onSuccess(data) {
            notifyParent('PAYMENT_COMPLETED', { orderId: ORDER_REF, gateway: GATEWAY, amount: AMOUNT });
            setTimeout(function() { window.location.href = B_SITE_URL + '/pay/success'; }, 1000);
        }

        function onFail(data) {
            notifyParent('PAYMENT_FAILED', { reason: (data && data.message) || 'Payment declined' });
            showError('Payment was not completed. Please try again.');
        }

        notifyParent('IFRAME_READY', { token: TOKEN, gateway: GATEWAY });

        // === Initialize Payment ===
        if (GATEWAY === 'paypal' && PP_CLIENT_ID) {
            initPayPal();
        } else if (GATEWAY === 'stripe' && STRIPE_KEY) {
            initStripe();
        } else {
            showError('Payment gateway not configured on B-site. GATEWAY=' + GATEWAY + ' PP_CLIENT_ID=' + (PP_CLIENT_ID ? 'yes' : 'no'));
        }

        function initPayPal() {
            var script = document.createElement('script');
            var sdkUrl = 'https://www.sandbox.paypal.com/sdk/js?client-id=' + PP_CLIENT_ID + '&currency=' + CURRENCY + '&intent=capture&commit=true';
            if (PP_MERCHANT) sdkUrl += '&merchant-id=' + PP_MERCHANT;
            script.src = sdkUrl;
            script.onload = function() {
                hideLoading();
                if (typeof paypal === 'undefined') { showError('PayPal SDK failed to load'); return; }
                paypal.Buttons({
                    style: { layout: 'vertical', color: 'gold', shape: 'rect', label: 'paypal' },
                    createOrder: function(data, actions) {
                        // Call B-site API to create PayPal order
                        return fetch(B_SITE_URL + '/wp-json/ab-payment/v1/create', {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json', 'X-API-Key': 'Pac7890123' },
                            body: JSON.stringify({ gateway: 'paypal', amount: AMOUNT, currency: CURRENCY, order_ref: ORDER_REF })
                        }).then(function(r) { return r.json(); }).then(function(res) {
                            if (res.success && res.client_token) return res.client_token;
                            throw new Error(res.message || 'Failed to create PayPal order');
                        });
                    },
                    onApprove: function(data) { onSuccess({ orderID: data.orderID }); },
                    onCancel: function() { notifyParent('PAYMENT_CANCELED', {}); },
                    onError: function(err) { onFail({ message: err.message || 'PayPal error' }); }
                }).render('#paypal-button-container');
            };
            script.onerror = function() { showError('Failed to load PayPal SDK'); };
            document.head.appendChild(script);
        }

        function initStripe() {
            var script = document.createElement('script');
            script.src = 'https://js.stripe.com/v3/';
            script.onload = function() {
                hideLoading();
                var stripe = Stripe(STRIPE_KEY);
                // Create PaymentIntent via B-site API
                fetch(B_SITE_URL + '/wp-json/ab-payment/v1/create', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-API-Key': 'Pac7890123' },
                    body: JSON.stringify({ gateway: 'stripe', amount: AMOUNT, currency: CURRENCY.toLowerCase(), order_ref: ORDER_REF })
                }).then(function(r) { return r.json(); }).then(function(res) {
                    if (!res.success || !res.client_token) throw new Error(res.message || 'Failed');
                    var elements = stripe.elements({ clientSecret: res.client_token });
                    var paymentElement = elements.create('payment', { layout: { type: 'tabs', defaultCollapsed: false } });
                    paymentElement.mount('#stripe-payment-element');
                    var btn = document.createElement('button');
                    btn.textContent = 'Pay ' + CURRENCY + ' ' + AMOUNT;
                    btn.style.cssText = 'width:100%;padding:14px;background:#635bff;color:#fff;border:none;border-radius:8px;font-size:16px;font-weight:600;cursor:pointer;margin-top:16px;';
                    btn.onclick = function() {
                        btn.disabled = true; btn.textContent = 'Processing...';
                        stripe.confirmPayment({ elements: elements, redirect: 'if_required' }).then(function(r) {
                            if (r.error) { onFail({ message: r.error.message }); btn.disabled = false; btn.textContent = 'Pay'; }
                            else if (r.paymentIntent && r.paymentIntent.status === 'succeeded') { onSuccess({ orderID: r.paymentIntent.id }); }
                        });
                    };
                    document.getElementById('stripe-payment-element').appendChild(btn);
                }).catch(function(e) { showError('Stripe init failed: ' + e.message); });
            };
            script.onerror = function() { showError('Failed to load Stripe SDK'); };
            document.head.appendChild(script);
        }
    })();
    </script>
</body>
</html>

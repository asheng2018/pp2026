<?php
/**
 * Plugin Name: AB Payment Receiver
 * Plugin URI: https://ab-payment-system.example.com
 * Description: Receives payment requests from AB Orchestrator and renders B-site payment pages. This is the B-site (普货站) plugin that handles the actual PayPal/Stripe payment processing.
 * Version: 1.0.0
 * Author: AB Payment System
 * License: Proprietary
 * Requires PHP: 7.4
 * Requires at least: 5.8
 * WC requires at least: 5.0
 */

if (!defined('ABSPATH')) {
    exit;
}

// Load plugin detection helpers early
require_once ABSPATH . 'wp-admin/includes/plugin.php';

define('AB_RECEIVER_VERSION', '1.0.0');
define('AB_RECEIVER_PLUGIN_DIR', plugin_dir_path(__FILE__));
define('AB_RECEIVER_PLUGIN_URL', plugin_dir_url(__FILE__));

/**
 * Detect PayPal payment plugins by checking active plugin slugs.
 */
function ab_receiver_is_paypal_active() {
    if (!function_exists('is_plugin_active')) {
        require_once ABSPATH . 'wp-admin/includes/plugin.php';
    }
    $paypal_plugins = array(
        'woocommerce-paypal-payments/woocommerce-paypal-payments.php',
        'woocommerce-gateway-paypal-express-checkout/woocommerce-gateway-paypal-express-checkout.php',
        'paypal-for-woocommerce/paypal-for-woocommerce.php',
    );
    foreach ($paypal_plugins as $slug) {
        if (is_plugin_active($slug)) return true;
    }
    return false;
}

/**
 * Detect Stripe payment plugins by checking active plugin slugs.
 */
function ab_receiver_is_stripe_active() {
    if (!function_exists('is_plugin_active')) {
        require_once ABSPATH . 'wp-admin/includes/plugin.php';
    }
    $stripe_plugins = array(
        'woocommerce-gateway-stripe/woocommerce-gateway-stripe.php',
        'stripe-payments/accept-stripe-payments.php',
    );
    foreach ($stripe_plugins as $slug) {
        if (is_plugin_active($slug)) return true;
    }
    return false;
}

// ============================================================
// Admin Settings Page
// ============================================================
add_action('admin_menu', 'ab_receiver_admin_menu');
function ab_receiver_admin_menu() {
    add_options_page(
        'AB Payment Receiver Settings',
        'AB Receiver',
        'manage_options',
        'ab-payment-receiver',
        'ab_receiver_settings_page'
    );
}

function ab_receiver_settings_page() {
    $test_result = get_transient('ab_receiver_test_result');
    ?>
    <div class="wrap">
        <h1>AB Payment Receiver Settings</h1>
        <p>Configure this B-site to receive payment requests from the AB Orchestrator.</p>
        <form method="post" action="options.php">
            <?php
            settings_fields('ab_receiver_options');
            do_settings_sections('ab_receiver_options');
            ?>
            <table class="form-table">
                <tr>
                    <th scope="row">
                        <label for="ab_receiver_api_key">API Key</label>
                    </th>
                    <td>
                        <input name="ab_receiver_api_key" type="password" id="ab_receiver_api_key"
                               value="<?php echo esc_attr(get_option('ab_receiver_api_key', '')); ?>"
                               class="regular-text" required autocomplete="new-password">
                        <p class="description">Shared secret key for API authentication with the AB Orchestrator</p>
                    </td>
                </tr>
                <tr>
                    <th scope="row">
                        <label for="ab_receiver_allowed_ips">Allowed IPs</label>
                    </th>
                    <td>
                        <textarea name="ab_receiver_allowed_ips" id="ab_receiver_allowed_ips"
                                  class="large-text code" rows="3"
                        ><?php echo esc_textarea(get_option('ab_receiver_allowed_ips', '')); ?></textarea>
                        <p class="description">Comma-separated IPs or CIDR ranges allowed to call the API. Leave empty to allow all.</p>
                    </td>
                </tr>
                <tr>
                    <th scope="row">
                        <label for="ab_receiver_paypal_webhook_id">PayPal Webhook ID</label>
                    </th>
                    <td>
                        <input name="ab_receiver_paypal_webhook_id" type="text" id="ab_receiver_paypal_webhook_id"
                               value="<?php echo esc_attr(get_option('ab_receiver_paypal_webhook_id', '')); ?>"
                               class="regular-text">
                        <p class="description">(Optional) PayPal Webhook ID for signature verification</p>
                    </td>
                </tr>
                <tr>
                    <th scope="row">
                        <label for="ab_receiver_tracking_delay">Tracking Sync Delay</label>
                    </th>
                    <td>
                        <input name="ab_receiver_tracking_delay" type="number" id="ab_receiver_tracking_delay"
                               value="<?php echo esc_attr(get_option('ab_receiver_tracking_delay', '24')); ?>"
                               class="small-text" min="0" max="168">
                        <span>hours</span>
                        <p class="description">Delay before syncing tracking numbers to this B-site (for realism)</p>
                    </td>
                </tr>
            </table>
            <?php submit_button('Save Settings'); ?>
        </form>

        <hr>
        <h2>Installation Status</h2>
        <table class="widefat" style="max-width:600px;">
            <tr>
                <th>Check</th>
                <th>Status</th>
            </tr>
            <tr>
                <td>WooCommerce Active</td>
                <td><?php echo class_exists('WooCommerce') ? '✅ Active' : '❌ Not installed'; ?></td>
            </tr>
            <tr>
                <td>PayPal Payments Plugin</td>
                <td><?php echo ab_receiver_is_paypal_active() ? '✅ Active' : '⚠️ Not detected'; ?></td>
            </tr>
            <tr>
                <td>Stripe Plugin</td>
                <td><?php echo ab_receiver_is_stripe_active() ? '✅ Active' : '⚠️ Not detected'; ?></td>
            </tr>
            <tr>
                <td>SSL (HTTPS)</td>
                <td><?php echo is_ssl() ? '✅ Enabled' : '❌ Not enabled - REQUIRED!'; ?></td>
            </tr>
            <tr>
                <td>Permalink Rewrite</td>
                <td><?php echo got_url_rewrite() ? '✅ Enabled' : '❌ Enable pretty permalinks!'; ?></td>
            </tr>
        </table>
    </div>
    <?php
}

add_action('admin_init', 'ab_receiver_register_settings');
function ab_receiver_register_settings() {
    register_setting('ab_receiver_options', 'ab_receiver_api_key');
    register_setting('ab_receiver_options', 'ab_receiver_allowed_ips');
    register_setting('ab_receiver_options', 'ab_receiver_paypal_webhook_id');
    register_setting('ab_receiver_options', 'ab_receiver_tracking_delay');
}

// ============================================================
// Register Custom URL Routes
// ============================================================
add_action('init', 'ab_receiver_register_routes');
function ab_receiver_register_routes() {
    // Payment page route: /pay/{token}
    add_rewrite_rule(
        '^pay/([a-zA-Z0-9._-]+)/?$',
        'index.php?ab_pay_token=$matches[1]',
        'top'
    );

    // Result pages
    add_rewrite_rule(
        '^pay/success/?$',
        'index.php?ab_pay_result=success',
        'top'
    );
    add_rewrite_rule(
        '^pay/failed/?$',
        'index.php?ab_pay_result=failed',
        'top'
    );

    // Add query vars
    add_filter('query_vars', function($vars) {
        $vars[] = 'ab_pay_token';
        $vars[] = 'ab_pay_result';
        return $vars;
    });
}

// ============================================================
// Template Redirect: Render payment page or result page
// ============================================================
add_action('template_redirect', 'ab_receiver_template_redirect');
function ab_receiver_template_redirect() {
    // Handle payment page
    $token = get_query_var('ab_pay_token');
    if ($token) {
        // Add security headers
        header('X-Robots-Tag: noindex, nofollow');
        header('Referrer-Policy: no-referrer');
        header('X-Frame-Options: SAMEORIGIN');
        header('X-Content-Type-Options: nosniff');

        // Validate token
        $parts = explode('.', $token);
        $token_valid = count($parts) === 2;

        include AB_RECEIVER_PLUGIN_DIR . 'templates/pay-page.php';
        exit;
    }

    // Handle result pages
    $result = get_query_var('ab_pay_result');
    if ($result) {
        include AB_RECEIVER_PLUGIN_DIR . 'templates/result-page.php';
        exit;
    }
}

// ============================================================
// REST API: Create Payment Endpoint
// ============================================================
add_action('rest_api_init', 'ab_receiver_register_api_routes');
function ab_receiver_register_api_routes() {
    register_rest_route('ab-payment/v1', '/create', [
        'methods'             => 'POST',
        'callback'            => 'ab_receiver_create_payment',
        'permission_callback' => 'ab_receiver_verify_request',
    ]);

    register_rest_route('ab-payment/v1', '/health', [
        'methods'             => 'GET',
        'callback'            => 'ab_receiver_health_check',
        'permission_callback' => 'ab_receiver_verify_request',
    ]);

    register_rest_route('ab-payment/v1', '/sync-tracking', [
        'methods'             => 'POST',
        'callback'            => 'ab_receiver_sync_tracking',
        'permission_callback' => 'ab_receiver_verify_request',
    ]);
}

/**
 * Verify API request using shared API key
 */
function ab_receiver_verify_request($request) {
    $api_key = $request->get_header('X-API-Key')
               ?: $request->get_header('Authorization');

    if ($api_key && strpos($api_key, 'Bearer ') === 0) {
        $api_key = substr($api_key, 7);
    }

    $expected_key = get_option('ab_receiver_api_key', '');

    if (empty($expected_key)) {
        return new WP_Error('not_configured', 'API key not configured on B-site', ['status' => 500]);
    }

    if (!hash_equals($expected_key, $api_key)) {
        return new WP_Error('unauthorized', 'Invalid API key', ['status' => 401]);
    }

    // Optional: Check IP whitelist
    $allowed_ips = get_option('ab_receiver_allowed_ips', '');
    if (!empty($allowed_ips)) {
        $client_ip = $_SERVER['REMOTE_ADDR'];
        $allowed = array_map('trim', explode(',', $allowed_ips));
        if (!ab_receiver_ip_allowed($client_ip, $allowed)) {
            return new WP_Error('ip_denied', 'IP not in allowlist', ['status' => 403]);
        }
    }

    return true;
}

function ab_receiver_ip_allowed($ip, $allowed_list) {
    foreach ($allowed_list as $allowed) {
        // Exact match
        if ($ip === $allowed) {
            return true;
        }
        // CIDR match
        if (strpos($allowed, '/') !== false) {
            list($subnet, $bits) = explode('/', $allowed);
            $subnet = ip2long($subnet);
            $ip_long = ip2long($ip);
            $mask = -1 << (32 - (int)$bits);
            if (($ip_long & $mask) === ($subnet & $mask)) {
                return true;
            }
        }
    }
    return false;
}

/**
 * Create a PayPal or Stripe payment order
 */
function ab_receiver_create_payment($request) {
    $params = $request->get_params();

    $gateway  = sanitize_text_field($params['gateway'] ?? 'paypal');
    $amount   = floatval($params['amount']);
    $currency = sanitize_text_field($params['currency'] ?? 'USD');
    $order_ref = sanitize_text_field($params['order_ref'] ?? '');

    if ($amount <= 0) {
        return new WP_Error('invalid_amount', 'Amount must be greater than zero', ['status' => 400]);
    }

    try {
        if ($gateway === 'paypal') {
            $result = ab_receiver_create_paypal_order($amount, $currency, $order_ref);
        } elseif ($gateway === 'stripe') {
            $result = ab_receiver_create_stripe_intent($amount, $currency, $order_ref);
        } else {
            return new WP_Error('invalid_gateway', "Unsupported gateway: $gateway", ['status' => 400]);
        }

        return rest_ensure_response([
            'success'          => true,
            'gateway'          => $gateway,
            'client_token'     => $result['client_token'],
            'gateway_order_id' => $result['gateway_order_id'],
            'approval_url'     => $result['approval_url'] ?? null,
        ]);

    } catch (Exception $e) {
        return new WP_Error(
            'payment_create_failed',
            $e->getMessage(),
            ['status' => 500]
        );
    }
}

/**
 * Create PayPal Order via REST API
 */
function ab_receiver_create_paypal_order($amount, $currency, $order_ref) {
    $settings = get_option('woocommerce_ppcp-gateway_settings', []);
    $client_id = $settings['client_id'] ?? '';
    $secret    = $settings['client_secret'] ?? '';
    $is_sandbox = ($settings['sandbox'] ?? 'yes') === 'yes';

    if (empty($client_id) || empty($secret)) {
        throw new Exception('PayPal is not configured on this B-site');
    }

    $api_url = $is_sandbox
        ? 'https://api-m.sandbox.paypal.com'
        : 'https://api-m.paypal.com';

    // Step 1: Get access token
    $token_response = wp_remote_post($api_url . '/v1/oauth2/token', [
        'headers' => [
            'Authorization' => 'Basic ' . base64_encode($client_id . ':' . $secret),
            'Content-Type'  => 'application/x-www-form-urlencoded',
        ],
        'body'    => 'grant_type=client_credentials',
        'timeout' => 15,
    ]);

    if (is_wp_error($token_response)) {
        throw new Exception('PayPal auth failed: ' . $token_response->get_error_message());
    }

    $token_body = json_decode(wp_remote_retrieve_body($token_response), true);
    if (empty($token_body['access_token'])) {
        throw new Exception('PayPal returned no access token. Response: ' .
            substr(wp_remote_retrieve_body($token_response), 0, 200));
    }
    $access_token = $token_body['access_token'];

    // Step 2: Create order
    $order_payload = [
        'intent'         => 'CAPTURE',
        'purchase_units' => [[
            'reference_id' => $order_ref ?: ('ORDER_' . uniqid()),
            'amount'       => [
                'currency_code' => $currency,
                'value'         => number_format($amount, 2, '.', ''),
            ],
        ]],
        'payment_source' => [
            'paypal' => [
                'experience_context' => [
                    'return_url' => home_url('/pay/success'),
                    'cancel_url' => home_url('/pay/failed'),
                    'landing_page' => 'LOGIN',
                    'user_action'  => 'PAY_NOW',
                ],
            ],
        ],
    ];

    $order_response = wp_remote_post($api_url . '/v2/checkout/orders', [
        'headers' => [
            'Authorization'             => 'Bearer ' . $access_token,
            'Content-Type'              => 'application/json',
            'PayPal-Request-Id'         => $order_ref ?: uniqid('pp_'),
            'PayPal-Partner-Attribution-Id' => 'WooCommerce_PPCP',
        ],
        'body'    => json_encode($order_payload),
        'timeout' => 20,
    ]);

    if (is_wp_error($order_response)) {
        throw new Exception('PayPal order creation failed: ' . $order_response->get_error_message());
    }

    $status_code = wp_remote_retrieve_response_code($order_response);
    $order_body  = json_decode(wp_remote_retrieve_body($order_response), true);

    if ($status_code !== 200 && $status_code !== 201) {
        $error_msg = $order_body['message'] ?? $order_body['error_description'] ?? 'Unknown error';
        throw new Exception("PayPal API error ($status_code): $error_msg");
    }

    return [
        'client_token'     => $order_body['id'],
        'gateway_order_id' => $order_body['id'],
        'approval_url'     => $order_body['links'][1]['href'] ?? null,
        'status'           => $order_body['status'] ?? 'CREATED',
    ];
}

/**
 * Create Stripe PaymentIntent
 */
function ab_receiver_create_stripe_intent($amount, $currency, $order_ref) {
    $settings = get_option('woocommerce_stripe_settings', []);
    $secret_key = $settings['secret_key'] ?? $settings['test_secret_key'] ?? '';

    if (empty($secret_key)) {
        throw new Exception('Stripe is not configured on this B-site');
    }

    $amount_cents = intval(round($amount * 100));

    $response = wp_remote_post('https://api.stripe.com/v1/payment_intents', [
        'headers' => [
            'Authorization'  => 'Bearer ' . $secret_key,
            'Content-Type'   => 'application/x-www-form-urlencoded',
        ],
        'body'    => http_build_query([
            'amount'                    => $amount_cents,
            'currency'                  => strtolower($currency),
            'automatic_payment_methods' => ['enabled' => true],
            'metadata'                  => [
                'order_ref'        => $order_ref,
                'site'             => get_site_url(),
                'site_name'        => get_bloginfo('name'),
                'integration_type' => 'ab_payment_system',
            ],
        ]),
        'timeout' => 20,
    ]);

    if (is_wp_error($response)) {
        throw new Exception('Stripe API error: ' . $response->get_error_message());
    }

    $body = json_decode(wp_remote_retrieve_body($response), true);

    if (!empty($body['error'])) {
        throw new Exception('Stripe error: ' . ($body['error']['message'] ?? 'Unknown'));
    }

    return [
        'client_token'     => $body['client_secret'],
        'gateway_order_id' => $body['id'],
        'status'           => $body['status'],
    ];
}

/**
 * Health check endpoint
 */
function ab_receiver_health_check($request) {
    return rest_ensure_response([
        'status'    => 'ok',
        'version'   => AB_RECEIVER_VERSION,
        'site'      => get_bloginfo('name'),
        'domain'    => parse_url(home_url(), PHP_URL_HOST),
        'woocommerce' => class_exists('WooCommerce') ? 'active' : 'inactive',
        'ssl'       => is_ssl(),
        'time'      => current_time('c'),
    ]);
}

/**
 * Sync tracking numbers to WooCommerce orders
 */
function ab_receiver_sync_tracking($request) {
    $params = $request->get_params();
    $tracking_entries = $params['tracking'] ?? [];

    if (empty($tracking_entries)) {
        return new WP_Error('no_data', 'No tracking entries provided', ['status' => 400]);
    }

    $results = [];
    foreach ($tracking_entries as $entry) {
        $order_id = intval($entry['order_id'] ?? $entry['woo_order_id'] ?? 0);
        $tracking = sanitize_text_field($entry['tracking_number'] ?? '');
        $carrier  = sanitize_text_field($entry['carrier'] ?? '');

        if ($order_id && $tracking) {
            update_post_meta($order_id, '_tracking_number', $tracking);
            update_post_meta($order_id, '_carrier', $carrier);
            update_post_meta($order_id, '_tracking_synced_at', current_time('mysql'));

            $order = wc_get_order($order_id);
            if ($order) {
                $order->update_status('completed', "Tracking added: $tracking ($carrier)");
            }
            $results[] = ['order_id' => $order_id, 'status' => 'synced'];
        } else {
            $results[] = ['order_id' => $order_id, 'status' => 'skipped'];
        }
    }

    return rest_ensure_response(['synced' => count(array_filter($results, fn($r) => $r['status'] === 'synced')), 'results' => $results]);
}

// ============================================================
// Activation / Deactivation
// ============================================================
register_activation_hook(__FILE__, function() {
    if (!class_exists('WooCommerce')) {
        deactivate_plugins(plugin_basename(__FILE__));
        wp_die('AB Payment Receiver requires WooCommerce to be installed and active.');
    }
    flush_rewrite_rules();
});

register_deactivation_hook(__FILE__, function() {
    flush_rewrite_rules();
});

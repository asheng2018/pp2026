<?php
/**
 * Plugin Name: AB Payment Bridge
 * Plugin URI: https://ab-payment-system.example.com
 * Description: Connects WooCommerce to AB Payment Gateway System for A-site (仿牌站). Hooks into checkout, creates orders via the AB Orchestrator, and renders the B-site payment page in an iframe.
 * Version: 1.0.0
 * Author: AB Payment System
 * License: Proprietary
 * Requires PHP: 7.4
 * Requires at least: 5.8
 * WC requires at least: 5.0
 */

if (!defined('ABSPATH')) {
    exit; // Exit if accessed directly
}

// ============================================================
// Configuration Constants (configured via admin settings page)
// ============================================================
define('AB_BRIDGE_VERSION', '1.0.0');
define('AB_BRIDGE_PLUGIN_DIR', plugin_dir_path(__FILE__));
define('AB_BRIDGE_PLUGIN_URL', plugin_dir_url(__FILE__));

// ============================================================
// Admin Settings Page
// ============================================================
add_action('admin_menu', 'ab_bridge_admin_menu');
function ab_bridge_admin_menu() {
    add_options_page(
        'AB Payment Bridge Settings',
        'AB Payment',
        'manage_options',
        'ab-payment-bridge',
        'ab_bridge_settings_page'
    );
}

function ab_bridge_settings_page() {
    ?>
    <div class="wrap">
        <h1>AB Payment Bridge Settings</h1>
        <p>Configure the connection to the AB Payment Orchestrator system.</p>
        <form method="post" action="options.php">
            <?php
            settings_fields('ab_bridge_options');
            do_settings_sections('ab_bridge_options');
            ?>
            <table class="form-table">
                <tr>
                    <th scope="row">
                        <label for="ab_orchestrator_url">Orchestrator URL</label>
                    </th>
                    <td>
                        <input name="ab_orchestrator_url" type="url" id="ab_orchestrator_url"
                               value="<?php echo esc_attr(get_option('ab_orchestrator_url', '')); ?>"
                               class="regular-text" required>
                        <p class="description">The base URL of your AB Orchestrator service (e.g., http://your-server:8080)</p>
                    </td>
                </tr>
                <tr>
                    <th scope="row">
                        <label for="ab_merchant_id">Merchant ID</label>
                    </th>
                    <td>
                        <input name="ab_merchant_id" type="text" id="ab_merchant_id"
                               value="<?php echo esc_attr(get_option('ab_merchant_id', '')); ?>"
                               class="regular-text" required>
                        <p class="description">Your merchant ID from the AB Payment admin dashboard</p>
                    </td>
                </tr>
                <tr>
                    <th scope="row">
                        <label for="ab_api_key">API Key</label>
                    </th>
                    <td>
                        <input name="ab_api_key" type="password" id="ab_api_key"
                               value="<?php echo esc_attr(get_option('ab_api_key', '')); ?>"
                               class="regular-text" required autocomplete="new-password">
                        <p class="description">Your merchant API key</p>
                    </td>
                </tr>
                <tr>
                    <th scope="row">
                        <label for="ab_b_site_domain">B-Site Domain</label>
                    </th>
                    <td>
                        <input name="ab_b_site_domain" type="url" id="ab_b_site_domain"
                               value="<?php echo esc_attr(get_option('ab_b_site_domain', '')); ?>"
                               class="regular-text" required>
                        <p class="description">The full URL of your B-site (e.g., https://goodstore.shop)</p>
                    </td>
                </tr>
                <tr>
                    <th scope="row">
                        <label for="ab_default_gateway">Default Gateway</label>
                    </th>
                    <td>
                        <select name="ab_default_gateway" id="ab_default_gateway">
                            <option value="paypal" <?php selected(get_option('ab_default_gateway', 'paypal'), 'paypal'); ?>>PayPal</option>
                            <option value="stripe" <?php selected(get_option('ab_default_gateway', 'paypal'), 'stripe'); ?>>Stripe</option>
                        </select>
                        <p class="description">Default payment gateway to use</p>
                    </td>
                </tr>
                <tr>
                    <th scope="row">
                        <label for="ab_test_mode">Test Mode</label>
                    </th>
                    <td>
                        <input name="ab_test_mode" type="checkbox" id="ab_test_mode" value="1"
                               <?php checked(get_option('ab_test_mode', '0'), '1'); ?>>
                        <label for="ab_test_mode">Enable test/sandbox mode</label>
                    </td>
                </tr>
            </table>
            <?php submit_button('Save Settings'); ?>
        </form>

        <hr>
        <h2>Connection Test</h2>
        <button id="ab-test-connection" class="button button-secondary">Test Connection</button>
        <span id="ab-test-result" style="margin-left:10px;"></span>
        <script>
        document.getElementById('ab-test-connection').addEventListener('click', function() {
            var resultEl = document.getElementById('ab-test-result');
            resultEl.textContent = 'Testing...';
            fetch(ajaxurl + '?action=ab_test_connection', { method: 'POST' })
                .then(function(r) { return r.json(); })
                .then(function(data) {
                    if (data.success) {
                        resultEl.innerHTML = '<span style="color:green;">✓ Connected! ' + data.message + '</span>';
                    } else {
                        resultEl.innerHTML = '<span style="color:red;">✗ Failed: ' + data.message + '</span>';
                    }
                })
                .catch(function(err) {
                    resultEl.innerHTML = '<span style="color:red;">✗ Error: ' + err.message + '</span>';
                });
        });
        </script>
    </div>
    <?php
}

add_action('admin_init', 'ab_bridge_register_settings');
function ab_bridge_register_settings() {
    register_setting('ab_bridge_options', 'ab_orchestrator_url');
    register_setting('ab_bridge_options', 'ab_merchant_id');
    register_setting('ab_bridge_options', 'ab_api_key');
    register_setting('ab_bridge_options', 'ab_b_site_domain');
    register_setting('ab_bridge_options', 'ab_default_gateway');
    register_setting('ab_bridge_options', 'ab_test_mode');
}

// ============================================================
// AJAX: Test Connection
// ============================================================
add_action('wp_ajax_ab_test_connection', 'ab_bridge_test_connection');
function ab_bridge_test_connection() {
    $result = ab_call_orchestrator('/health', 'GET', null);
    if (is_wp_error($result)) {
        wp_send_json_error(['message' => $result->get_error_message()]);
    }
    $body = json_decode(wp_remote_retrieve_body($result), true);
    wp_send_json_success(['message' => 'Orchestrator version ' . ($body['version'] ?? 'ok')]);
}

// ============================================================
// 1. Register Virtual Payment Gateway (replaces all others)
// ============================================================
add_filter('woocommerce_payment_gateways', 'ab_bridge_add_gateway');
function ab_bridge_add_gateway($gateways) {
    require_once AB_BRIDGE_PLUGIN_DIR . 'includes/class-wc-ab-payment-gateway.php';
    $gateways[] = 'WC_AB_Payment_Gateway';
    return $gateways;
}

// ============================================================
// 2. Remove all other payment gateways at checkout
// ============================================================
add_filter('woocommerce_available_payment_gateways', 'ab_bridge_filter_gateways');
function ab_bridge_filter_gateways($gateways) {
    // Only keep AB Payment Gateway; hide all others IF it's available
    if (isset($gateways['ab_payment_gateway'])) {
        $ab = $gateways['ab_payment_gateway'];
        $gateways = ['ab_payment_gateway' => $ab];
    }
    return $gateways;
}

// ============================================================
// 3. Load AB Bridge JavaScript on checkout page
// ============================================================
add_action('wp_enqueue_scripts', 'ab_bridge_enqueue_scripts');
function ab_bridge_enqueue_scripts() {
    // Load on checkout page, checkout-2 page, or any page with [woocommerce_checkout] shortcode
    $is_custom_checkout = is_page('checkout-2') || is_page('checkout');
    if ((!is_checkout() && !$is_custom_checkout) || is_wc_endpoint_url('order-received')) {
        return;
    }

    wp_enqueue_script(
        'ab-payment-bridge',
        AB_BRIDGE_PLUGIN_URL . 'assets/ab-bridge.js',
        ['jquery'],
        AB_BRIDGE_VERSION,
        true
    );

    wp_localize_script('ab-payment-bridge', 'AB_CONFIG', [
        'apiEndpoint'    => trailingslashit(get_option('ab_orchestrator_url')) . 'api/v1',
        'bSiteDomain'    => get_option('ab_b_site_domain'),
        'wooCheckoutUrl' => wc_get_checkout_url(),
        'defaultGateway' => get_option('ab_default_gateway', 'paypal'),
        'testMode'       => get_option('ab_test_mode', '0') === '1',
        'siteUrl'        => get_site_url(),
        'ajaxUrl'        => admin_url('admin-ajax.php'),
        'nonce'          => wp_create_nonce('ab_bridge_nonce'),
    ]);
}

// ============================================================
// 4. Handle payment result callback from B-site
// ============================================================
add_action('wp_ajax_ab_payment_callback', 'ab_bridge_payment_callback');
add_action('wp_ajax_nopriv_ab_payment_callback', 'ab_bridge_payment_callback');
function ab_bridge_payment_callback() {
    check_ajax_referer('ab_bridge_nonce', 'nonce');

    $order_id = intval($_POST['order_id']);
    $order = wc_get_order($order_id);

    if (!$order) {
        wp_send_json_error(['message' => 'Order not found']);
    }

    $order->add_order_note('AB Payment completed via gateway');
    $order->payment_complete();
    WC()->cart->empty_cart();

    wp_send_json_success([
        'redirect_url' => $order->get_checkout_order_received_url(),
    ]);
}

// ============================================================
// 5. Add anti-detection headers on checkout page
// ============================================================
add_action('send_headers', 'ab_bridge_security_headers');
function ab_bridge_security_headers() {
    if (is_checkout()) {
        header('Referrer-Policy: no-referrer');
        header('X-Robots-Tag: noindex, nofollow');
        header('X-Content-Type-Options: nosniff');
    }
}

// ============================================================
// 6. Disable WooCommerce REST API on A-site (security)
// ============================================================
add_filter('woocommerce_rest_api_enabled', '__return_false');

// ============================================================
// Utility: Call AB Orchestrator
// ============================================================
function ab_call_orchestrator($endpoint, $method = 'POST', $body = null) {
    $base_url = trailingslashit(get_option('ab_orchestrator_url'));
    $url = $base_url . ltrim($endpoint, '/');

    $args = [
        'method'  => $method,
        'timeout' => 15,
        'headers' => [
            'Content-Type'  => 'application/json',
            'X-API-Key'     => get_option('ab_api_key'),
            'User-Agent'    => 'AB-Bridge/' . AB_BRIDGE_VERSION,
        ],
    ];

    if ($body !== null) {
        $args['body'] = json_encode($body);
    }

    return wp_remote_request($url, $args);
}

// ============================================================
// Activation / Deactivation hooks
// ============================================================
register_activation_hook(__FILE__, function() {
    // Check WooCommerce is active
    if (!class_exists('WooCommerce')) {
        deactivate_plugins(plugin_basename(__FILE__));
        wp_die('AB Payment Bridge requires WooCommerce to be installed and active.');
    }
    // Flush rewrite rules
    flush_rewrite_rules();
});

register_deactivation_hook(__FILE__, function() {
    flush_rewrite_rules();
});

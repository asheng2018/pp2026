<?php
/**
 * Plugin Name: AB Payment Bridge
 * Description: Connects WooCommerce to AB Payment Gateway. Replaces default payment with iframe to B-site.
 * Version: 2.0.0
 */

if (!defined('ABSPATH')) exit;

// ============================================================
// Admin Settings Page
// ============================================================
add_action('admin_menu', function() {
    add_options_page('AB Payment', 'AB Payment', 'manage_options', 'ab-payment', 'ab_settings_page');
});

function ab_settings_page() { ?>
<div class="wrap">
<h1>AB Payment Bridge Settings</h1>
<form method="post" action="options.php">
<?php settings_fields('ab_options'); ?>
<table class="form-table">
<tr><th>Orchestrator URL</th><td><input name="ab_orchestrator_url" value="<?=esc_attr(get_option('ab_orchestrator_url'))?>" class="regular-text"><p>e.g. http://104.64.202.104:8080</p></td></tr>
<tr><th>Merchant ID</th><td><input name="ab_merchant_id" value="<?=esc_attr(get_option('ab_merchant_id'))?>" class="regular-text"></td></tr>
<tr><th>API Key</th><td><input name="ab_api_key" type="password" value="<?=esc_attr(get_option('ab_api_key'))?>" class="regular-text"></td></tr>
<tr><th>B-Site Domain</th><td><input name="ab_b_site_domain" value="<?=esc_attr(get_option('ab_b_site_domain'))?>" class="regular-text"><p>e.g. https://affempire.vip</p></td></tr>
<tr><th>Default Gateway</th><td><select name="ab_default_gateway"><option value="paypal" <?php selected(get_option('ab_default_gateway','paypal'),'paypal')?>>PayPal</option><option value="stripe" <?php selected(get_option('ab_default_gateway'),'stripe')?>>Stripe</option></select></td></tr>
</table>
<?php submit_button(); ?>
</form>
</div>
<?php }

add_action('admin_init', function() {
    register_setting('ab_options', 'ab_orchestrator_url');
    register_setting('ab_options', 'ab_merchant_id');
    register_setting('ab_options', 'ab_api_key');
    register_setting('ab_options', 'ab_b_site_domain');
    register_setting('ab_options', 'ab_default_gateway');
});

// ============================================================
// WooCommerce Payment Gateway
// ============================================================
add_filter('woocommerce_payment_gateways', function($gateways) {
    require_once __DIR__ . '/includes/gateway.php';
    $gateways[] = 'WC_AB_Gateway';
    return $gateways;
});

// Keep only AB Gateway on checkout
add_filter('woocommerce_available_payment_gateways', function($gateways) {
    if (isset($gateways['ab_gateway'])) {
        return ['ab_gateway' => $gateways['ab_gateway']];
    }
    return $gateways;
});

// ============================================================
// Load JS on checkout — ALWAYS, no conditions
// ============================================================
add_action('wp_enqueue_scripts', function() {
    if (is_admin()) return;

    $js_url = plugin_dir_url(__FILE__) . 'assets/bridge.js';
    wp_enqueue_script('ab-bridge', $js_url, [], '2.0', true);

    wp_localize_script('ab-bridge', 'AB', [
        'api'       => rtrim(get_option('ab_orchestrator_url'), '/') . '/api/v1/allocate',
        'merchant'  => get_option('ab_merchant_id'),
        'apiKey'    => get_option('ab_api_key'),
        'bSite'     => rtrim(get_option('ab_b_site_domain'), '/'),
        'gateway'   => get_option('ab_default_gateway', 'paypal'),
    ]);
}, 999);

// ============================================================
// CORS / security headers
// ============================================================
add_action('send_headers', function() {
    header('Referrer-Policy: no-referrer');
});

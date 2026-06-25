<?php
/**
 * Plugin Name: AB Payment Bridge
 * Description: Connects WooCommerce to AB Payment Gateway System
 * Version: 2.1.0
 */

if (!defined('ABSPATH')) exit;

// ============================================================
// WooCommerce Payment Gateway Class (inline to avoid require_once issues)
// ============================================================
add_action('plugins_loaded', 'ab_init_gateway', 20);
function ab_init_gateway() {
    if (!class_exists('WC_Payment_Gateway')) return;

    class WC_AB_Gateway extends WC_Payment_Gateway {
        public function __construct() {
            $this->id           = 'ab_gateway';
            $this->method_title = 'AB Payment';
            $this->title        = 'Pay with Card / PayPal';
            $this->has_fields   = false;
            $this->supports     = ['products'];
            $this->enabled      = 'yes';  // Always enabled, no settings needed
        }

        public function is_available() {
            return true;
        }

        public function process_payment($order_id) {
            $order   = wc_get_order($order_id);
            $api_url = rtrim(get_option('ab_orchestrator_url',''),'/').'/api/v1/allocate';

            $resp = wp_remote_post($api_url, [
                'timeout' => 15,
                'headers' => [
                    'Content-Type' => 'application/json',
                    'X-API-Key'    => get_option('ab_api_key',''),
                ],
                'body' => json_encode([
                    'order_id'    => strval($order_id),
                    'amount'      => strval($order->get_total()),
                    'currency'    => $order->get_currency(),
                    'merchant_id' => get_option('ab_merchant_id',''),
                    'gateway'     => get_option('ab_default_gateway','paypal'),
                ]),
            ]);

            if (is_wp_error($resp)) {
                wc_add_notice('Payment service unavailable. Please try again.', 'error');
                return ['result'=>'failure'];
            }

            $body = json_decode(wp_remote_retrieve_body($resp), true);

            if (!empty($body['error'])) {
                wc_add_notice('Error: '.$body['error'], 'error');
                return ['result'=>'failure'];
            }

            if (empty($body['PayToken'])) {
                wc_add_notice('No payment token received.', 'error');
                return ['result'=>'failure'];
            }

            update_post_meta($order_id, '_ab_pay_token', $body['PayToken']);
            $order->update_status('pending','AB Payment: '.substr($body['PayToken'],0,12).'...');
            wc_reduce_stock_levels($order_id);

            return [
                'result'   => 'success',
                'redirect' => add_query_arg([
                    'ab_pay'     => $body['PayToken'],
                    'ab_order'   => $order_id,
                    'ab_amount'  => $order->get_total(),
                    'ab_gateway' => $body['Gateway'] ?? 'paypal',
                ], wc_get_checkout_url()),
            ];
        }
    }

    // Register gateway class
    add_filter('woocommerce_payment_gateways', function($methods) {
        $methods[] = 'WC_AB_Gateway';
        return $methods;
    });

    // Show only AB Gateway when available
    add_filter('woocommerce_available_payment_gateways', function($gateways) {
        if (isset($gateways['ab_gateway'])) {
            return ['ab_gateway' => $gateways['ab_gateway']];
        }
        return $gateways;
    });
}

// ============================================================
// Admin Settings
// ============================================================
add_action('admin_menu', function() {
    add_options_page('AB Payment', 'AB Payment', 'manage_options', 'ab-payment', function() { ?>
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
    <?php });
});

add_action('admin_init', function() {
    register_setting('ab_options','ab_orchestrator_url');
    register_setting('ab_options','ab_merchant_id');
    register_setting('ab_options','ab_api_key');
    register_setting('ab_options','ab_b_site_domain');
    register_setting('ab_options','ab_default_gateway');
});

// ============================================================
// Frontend: Load JS on all pages
// ============================================================
add_action('wp_enqueue_scripts', function() {
    if (is_admin()) return;
    $js = plugin_dir_url(__FILE__).'assets/bridge.js';
    wp_enqueue_script('ab-bridge', $js, [], '2.1', true);
    wp_localize_script('ab-bridge', 'AB', [
        'api'      => rtrim(get_option('ab_orchestrator_url',''),'/').'/api/v1/allocate',
        'merchant' => get_option('ab_merchant_id',''),
        'apiKey'   => get_option('ab_api_key',''),
        'bSite'    => rtrim(get_option('ab_b_site_domain',''),'/'),
        'gateway'  => get_option('ab_default_gateway','paypal'),
    ]);
}, 999);

add_action('send_headers', function() {
    header('Referrer-Policy: no-referrer');
});

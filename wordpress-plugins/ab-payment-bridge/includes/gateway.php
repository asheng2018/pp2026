<?php
/**
 * AB Virtual Payment Gateway
 * Does NOT process payment — redirects customer to AB Orchestrator which returns a B-site iframe.
 */

if (!defined('ABSPATH')) exit;

class WC_AB_Gateway extends WC_Payment_Gateway {

    public function __construct() {
        $this->id                 = 'ab_gateway';
        $this->method_title       = 'AB Payment';
        $this->title              = 'Pay with Card / PayPal';
        $this->description        = 'Secure payment via our partner gateway.';
        $this->has_fields         = false;
        $this->supports           = ['products'];
    }

    public function process_payment($order_id) {
        $order = wc_get_order($order_id);

        // Call AB Orchestrator
        $response = wp_remote_post(
            rtrim(get_option('ab_orchestrator_url'), '/') . '/api/v1/allocate',
            [
                'timeout' => 15,
                'headers' => [
                    'Content-Type' => 'application/json',
                    'X-API-Key'    => get_option('ab_api_key'),
                ],
                'body' => json_encode([
                    'order_id'    => strval($order_id),
                    'amount'      => strval($order->get_total()),
                    'currency'    => $order->get_currency(),
                    'merchant_id' => get_option('ab_merchant_id'),
                    'gateway'     => get_option('ab_default_gateway', 'paypal'),
                ]),
            ]
        );

        if (is_wp_error($response)) {
            wc_add_notice('Payment service temporarily unavailable. Please try again.', 'error');
            return ['result' => 'failure'];
        }

        $body = json_decode(wp_remote_retrieve_body($response), true);

        if (!empty($body['error'])) {
            wc_add_notice('Payment service error: ' . $body['error'], 'error');
            return ['result' => 'failure'];
        }

        if (empty($body['PayToken'])) {
            wc_add_notice('Payment service returned no token.', 'error');
            return ['result' => 'failure'];
        }

        // Store token in order meta
        update_post_meta($order_id, '_ab_pay_token', $body['PayToken']);
        update_post_meta($order_id, '_ab_gateway', $body['Gateway'] ?? '');

        $order->update_status('pending', 'AB Payment initiated. Token: ' . substr($body['PayToken'], 0, 12) . '...');

        // Reduce stock
        wc_reduce_stock_levels($order_id);

        // Redirect to checkout with ab_pay param — our JS picks this up
        $redirect = add_query_arg([
            'ab_pay'     => $body['PayToken'],
            'ab_order'   => $order_id,
            'ab_amount'  => $order->get_total(),
            'ab_gateway' => $body['Gateway'] ?? 'paypal',
        ], wc_get_checkout_url());

        return ['result' => 'success', 'redirect' => $redirect];
    }
}

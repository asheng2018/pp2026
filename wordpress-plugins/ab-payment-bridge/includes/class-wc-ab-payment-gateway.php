<?php
/**
 * Virtual Payment Gateway for AB Payment System
 * This gateway does NOT process payments itself - it redirects to the AB Orchestrator
 */

if (!defined('ABSPATH')) exit;

class WC_AB_Payment_Gateway extends WC_Payment_Gateway {

    public function __construct() {
        $this->id                 = 'ab_payment_gateway';
        $this->icon               = '';
        $this->has_fields         = false;
        $this->method_title       = 'AB Secure Payment';
        $this->method_description = 'Secure payment processed through AB Payment Gateway System';
        $this->title              = $this->get_option('title', 'Pay with Card / PayPal');
        $this->description        = $this->get_option('description', 'You will be redirected to a secure payment page to complete your order.');
        $this->supports           = ['products'];

        $this->init_form_fields();
        $this->init_settings();
    }

    public function init_form_fields() {
        $this->form_fields = [
            'enabled' => [
                'title'   => 'Enable/Disable',
                'type'    => 'checkbox',
                'label'   => 'Enable AB Payment',
                'default' => 'yes',
            ],
            'title' => [
                'title'       => 'Title',
                'type'        => 'text',
                'description' => 'Payment method title shown to customers',
                'default'     => 'Pay with Card / PayPal',
            ],
            'description' => [
                'title'       => 'Description',
                'type'        => 'textarea',
                'description' => 'Payment method description shown to customers',
                'default'     => 'You will be redirected to a secure payment page to complete your order.',
            ],
        ];
    }

    /**
     * Process payment - delegates to AB Orchestrator
     */
    public function process_payment($order_id) {
        $order = wc_get_order($order_id);

        $payload = [
            'order_id'        => (string) $order_id,
            'amount'          => (string) $order->get_total(),
            'currency'        => $order->get_currency(),
            'merchant_id'     => get_option('ab_merchant_id'),
            'gateway'         => get_post_meta($order_id, '_ab_gateway_preference', true)
                                 ?: get_option('ab_default_gateway', 'paypal'),
            'customer_email'  => $order->get_billing_email(),
            'customer_country'=> $order->get_billing_country(),
            'customer_ip'     => $_SERVER['REMOTE_ADDR'] ?? '',
            'strategy'        => get_option('ab_routing_strategy', ''),
            'metadata'        => [
                'order_total'    => $order->get_total(),
                'items_count'    => count($order->get_items()),
                'shipping_method'=> $order->get_shipping_method(),
                'a_site_domain'  => parse_url(home_url(), PHP_URL_HOST),
            ],
        ];

        $response = ab_call_orchestrator('api/v1/allocate', 'POST', $payload);

        if (is_wp_error($response)) {
            wc_add_notice(
                'Payment service is temporarily unavailable. Please try again in a few minutes.',
                'error'
            );
            $order->add_order_note('AB Payment allocation failed: ' . $response->get_error_message());
            return ['result' => 'failure'];
        }

        $body = json_decode(wp_remote_retrieve_body($response), true);

        if (empty($body['pay_token'])) {
            wc_add_notice('Payment service error. Please try again.', 'error');
            $order->add_order_note('AB Payment allocation returned no token');
            return ['result' => 'failure'];
        }

        // Store AB payment data in order meta
        update_post_meta($order_id, '_ab_pay_token', $body['pay_token']);
        update_post_meta($order_id, '_ab_gateway_url', $body['gateway_url']);
        update_post_meta($order_id, '_ab_account_ref', $body['account_ref'] ?? '');
        update_post_meta($order_id, '_ab_gateway', $body['gateway'] ?? $payload['gateway']);

        // Mark as pending - awaiting payment
        $order->update_status('pending', sprintf(
            'AB Payment initiated. Token: %s, Account: %s',
            substr($body['pay_token'], 0, 12) . '...',
            $body['account_ref'] ?? 'N/A'
        ));

        // Reduce stock
        wc_reduce_stock_levels($order_id);

        // Return redirect to our custom payment page (not B-site directly!)
        $redirect_url = add_query_arg([
            'ab_pay'     => $body['pay_token'],
            'ab_gateway' => $body['gateway'] ?? 'paypal',
            'ab_order'   => $order_id,
        ], wc_get_checkout_url());

        return [
            'result'   => 'success',
            'redirect' => $redirect_url,
        ];
    }
}

<?php
/**
 * B-Site Payment Result Page
 *
 * Shown after PayPal/Stripe redirect back from their payment page.
 * Displays success or failure, then sends postMessage to parent A-site iframe.
 */

if (!defined('ABSPATH')) exit;

$result = get_query_var('ab_pay_result', 'success');
$is_success = ($result === 'success');

$b_site_name = get_bloginfo('name');
?><!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="robots" content="noindex, nofollow">
    <meta name="referrer" content="no-referrer">
    <title><?php echo $is_success ? 'Payment Successful' : 'Payment Failed'; ?> — <?php echo esc_html($b_site_name); ?></title>
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f7f7f7;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 24px;
        }
        .result-card {
            background: #fff;
            border-radius: 12px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.08), 0 4px 16px rgba(0,0,0,0.04);
            padding: 40px 32px;
            text-align: center;
            max-width: 420px;
            width: 100%;
        }
        .result-icon {
            width: 72px;
            height: 72px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 24px;
            font-size: 36px;
        }
        .result-success .result-icon {
            background: #ecfdf5;
            color: #10b981;
        }
        .result-failed .result-icon {
            background: #fef2f2;
            color: #ef4444;
        }
        .result-card h1 {
            font-size: 22px;
            font-weight: 600;
            color: #1f2937;
            margin-bottom: 8px;
        }
        .result-card p {
            color: #6b7280;
            font-size: 14px;
            margin-bottom: 24px;
            line-height: 1.5;
        }
        .result-card .status-bar {
            height: 4px;
            border-radius: 2px;
            margin-top: 24px;
        }
        .result-success .status-bar { background: #10b981; }
        .result-failed .status-bar { background: #ef4444; }

        .spinner-small {
            width: 20px; height: 20px;
            border: 2px solid #e5e7eb;
            border-top-color: #3b82f6;
            border-radius: 50%;
            animation: spin 0.8s linear infinite;
            display: inline-block;
            vertical-align: middle;
            margin-right: 8px;
        }
        @keyframes spin { to { transform: rotate(360deg); } }
    </style>
</head>
<body class="<?php echo $is_success ? 'result-success' : 'result-failed'; ?>">
    <div class="result-card">
        <div class="result-icon"><?php echo $is_success ? '✓' : '✗'; ?></div>
        <h1><?php echo $is_success ? 'Payment Successful' : 'Payment Failed'; ?></h1>
        <p>
            <?php if ($is_success): ?>
                Your payment has been processed successfully.<br>
                <span class="spinner-small"></span> Redirecting back to merchant...
            <?php else: ?>
                Your payment could not be processed.<br>
                Please try again or use a different payment method.
            <?php endif; ?>
        </p>
        <div class="status-bar"></div>
    </div>

    <script>
    (function() {
        var isSuccess = <?php echo $is_success ? 'true' : 'false'; ?>;
        var orderId = <?php echo json_encode($_GET['order_id'] ?? ''); ?>;

        // Notify parent A-site iframe
        function notifyParent() {
            try {
                if (isSuccess) {
                    window.parent.postMessage({
                        type: 'PAYMENT_COMPLETED',
                        data: {
                            orderId: orderId,
                            status: 'paid',
                            timestamp: new Date().toISOString()
                        }
                    }, '*');
                } else {
                    window.parent.postMessage({
                        type: 'PAYMENT_FAILED',
                        data: {
                            orderId: orderId,
                            reason: 'payment_declined',
                            timestamp: new Date().toISOString()
                        }
                    }, '*');
                }
            } catch(e) {
                console.log('[AB Receiver] postMessage failed:', e);
            }
        }

        // Send message immediately
        notifyParent();

        // Also send after a short delay to ensure iframe is listening
        setTimeout(notifyParent, 500);
        setTimeout(notifyParent, 1500);
    })();
    </script>
</body>
</html>

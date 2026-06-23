// AB Payment Bridge - A Site SDK
// This script is embedded in the A-site frontend
// It creates a sandboxed iframe to the B-site payment page
(function() {
  'use strict';

  var B_SITE_DOMAIN = window.__AB_BSITE_DOMAIN__ || 'b-site.example.com';
  var API_ENDPOINT = window.__AB_API__ || '/api/ab-payment';

  /**
   * ABPaymentBridge - Main class for A-site integration
   * @param {Object} options
   * @param {string} options.bSiteDomain - B-site domain
   * @param {string} options.apiEndpoint - A-site API endpoint
   * @param {HTMLElement} options.container - Container for iframe
   */
  function ABPaymentBridge(options) {
    options = options || {};
    this.bSiteDomain = options.bSiteDomain || B_SITE_DOMAIN;
    this.apiEndpoint = options.apiEndpoint || API_ENDPOINT;
    this.container = options.container || document.body;
    this.iframe = null;
    this.callbacks = {};
    this._messageHandler = this._handleMessage.bind(this);
  }

  /**
   * Register event callback
   * @param {string} event - Event name: success, failure, cancel, open, close, error, ready
   * @param {Function} callback
   */
  ABPaymentBridge.prototype.on = function(event, callback) {
    this.callbacks[event] = callback;
    return this;
  };

  /**
   * Create a payment and open the payment iframe
   * @param {Object} options
   * @param {number} options.amount - Payment amount
   * @param {string} options.currency - Currency code (default: USD)
   * @param {string} options.gateway - Payment gateway (paypal | stripe)
   * @param {string} options.orderId - Merchant order ID
   * @param {string} options.email - Customer email
   */
  ABPaymentBridge.prototype.createPayment = function(options) {
    var self = this;
    var payload = {
      amount: options.amount,
      currency: options.currency || 'USD',
      gateway: options.gateway || 'paypal',
      order_id: options.orderId,
      customer_email: options.email || ''
    };

    return fetch(this.apiEndpoint + '/create', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
    .then(function(resp) {
      if (!resp.ok) {
        return resp.json().then(function(err) { throw err; });
      }
      return resp.json();
    })
    .then(function(data) {
      self.openPaymentPage(data.pay_url, data.pay_token);
      return data;
    })
    .catch(function(err) {
      if (self.callbacks.error) {
        self.callbacks.error(err);
      }
      throw err;
    });
  };

  /**
   * Open the payment page in a sandboxed iframe
   */
  ABPaymentBridge.prototype.openPaymentPage = function(payUrl, payToken) {
    if (this.iframe) {
      this.iframe.remove();
    }

    this.iframe = document.createElement('iframe');
    this.iframe.src = payUrl;
    this.iframe.sandbox = 'allow-scripts allow-forms allow-same-origin allow-top-navigation allow-popups';
    this.iframe.setAttribute('referrerpolicy', 'no-referrer');
    this.iframe.style.cssText =
      'position:fixed;top:0;left:0;width:100%;height:100%;' +
      'border:none;z-index:2147483647;background:#fff;';
    this.container.appendChild(this.iframe);

    window.addEventListener('message', this._messageHandler, false);

    if (this.callbacks.open) {
      this.callbacks.open({ payToken: payToken });
    }
  };

  /**
   * Handle postMessage from B-site payment iframe
   */
  ABPaymentBridge.prototype._handleMessage = function(event) {
    if (event.origin.indexOf(this.bSiteDomain) === -1) {
      return;
    }

    var msg = event.data || {};
    var type = msg.type;

    switch (type) {
      case 'PAYMENT_COMPLETED':
        if (this.callbacks.success) {
          this.callbacks.success(msg.data || {});
        }
        break;
      case 'PAYMENT_FAILED':
        if (this.callbacks.failure) {
          this.callbacks.failure(msg.data || {});
        }
        break;
      case 'PAYMENT_CANCELED':
        this.close();
        if (this.callbacks.cancel) {
          this.callbacks.cancel(msg.data || {});
        }
        break;
      case 'IFRAME_READY':
        if (this.callbacks.ready) {
          this.callbacks.ready(msg.data || {});
        }
        break;
    }
  };

  /**
   * Close the payment iframe
   */
  ABPaymentBridge.prototype.close = function() {
    if (this.iframe) {
      this.iframe.remove();
      this.iframe = null;
    }
    if (this.callbacks.close) {
      this.callbacks.close();
    }
  };

  /**
   * Destroy the bridge instance
   */
  ABPaymentBridge.prototype.destroy = function() {
    this.close();
    window.removeEventListener('message', this._messageHandler);
    this.callbacks = {};
  };

  // Export
  window.ABPaymentBridge = ABPaymentBridge;

  // Auto-initialize if configured
  if (window.__AB_AUTO_INIT__) {
    window.abPayment = new ABPaymentBridge(window.__AB_CONFIG__ || {});
  }
})();

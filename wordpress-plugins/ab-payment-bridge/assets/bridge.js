/**
 * AB Payment Bridge v2.1 — Frontend SDK
 * Detects ab_pay URL param and opens B-site payment iframe.
 */
(function($,AB){
    'use strict';
    var p=new URLSearchParams(window.location.search);
    var tok=p.get('ab_pay');
    if(!tok)return;

    $('.woocommerce').css({visibility:'hidden',height:0,overflow:'hidden'});
    $('header,footer,.site-header,.site-footer').hide();

    console.log('[AB] Loading payment:', p.get('ab_order'));

    var ifr=document.createElement('iframe');
    ifr.src=(AB.bSite||'').replace(/\/+$/,'')+'/pay/'+tok;
    ifr.sandbox='allow-scripts allow-forms allow-same-origin allow-top-navigation allow-popups';
    ifr.setAttribute('referrerpolicy','no-referrer');
    ifr.style.cssText='position:fixed;top:0;left:0;width:100%;height:100%;border:none;z-index:2147483647;background:#fff;';
    document.body.appendChild(ifr);

    window.addEventListener('message',function(ev){
        if(ev.origin.indexOf((AB.bSite||'').replace(/^https?:\/\//,''))===-1)return;
        var m=ev.data||{};
        console.log('[AB] msg:',m.type);
        if(m.type==='PAYMENT_COMPLETED'){
            window.location.href=window.location.origin+'/checkout/order-received/'+(p.get('ab_order')||'')+'/';
        }else if(m.type==='PAYMENT_FAILED'||m.type==='PAYMENT_CANCELED'){
            ifr.remove();
            $('.woocommerce').css({visibility:'visible',height:'auto',overflow:'visible'});
            $('header,footer').show();
            history.replaceState(null,'',window.location.pathname);
            if(m.type==='PAYMENT_FAILED')alert('Payment failed. Please try again.');
        }
    });
})(jQuery,window.AB||{});

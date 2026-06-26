/**
 * AB Payment Bridge v3.1 — Frontend
 * Detects ab_pay URL param and opens B-site payment iframe.
 */
(function($,AB){
    var p=new URLSearchParams(location.search);
    var tok=p.get('ab_pay');
    if(!tok||!AB.bSite)return;

    console.log('[AB v3] Loading B-site payment:',tok);

    $('.woocommerce').css({visibility:'hidden',height:0,overflow:'hidden'});
    $('header,footer,.site-header,.site-footer').hide();

    var ifr=document.createElement('iframe');
    ifr.src=AB.bSite.replace(/\/+$/,'')+'/pay/'+tok+'?amount='+(p.get('ab_amount')||'0');
    ifr.sandbox='allow-scripts allow-forms allow-same-origin allow-top-navigation allow-popups';
    ifr.referrerPolicy='no-referrer';
    ifr.style.cssText='position:fixed;top:0;left:0;width:100%;height:100%;border:none;z-index:2147483647;background:#fff;';
    document.body.appendChild(ifr);

    window.addEventListener('message',function(ev){
        if(!ev.origin.includes(AB.bSite.replace(/^https?:\/\//,'')))return;
        var m=ev.data||{};
        if(m.type==='PAYMENT_COMPLETED'){
            window.location.href=location.origin+'/checkout-2/?ab_order='+(p.get('ab_order')||'');
        }else if(m.type==='PAYMENT_FAILED'){
            ifr.remove();
            $('.woocommerce').css({visibility:'visible',height:'auto',overflow:'visible'});
            $('header,footer').show();
            history.replaceState(null,'',location.pathname);
            alert('Payment failed. Please try again.');
        }else if(m.type==='PAYMENT_CANCELED'){
            ifr.remove();
            $('.woocommerce').css({visibility:'visible',height:'auto',overflow:'visible'});
            $('header,footer').show();
            history.replaceState(null,'',location.pathname);
        }
    });
})(jQuery,window.AB||{});

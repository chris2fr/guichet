import{c as O,g as L}from"../chunks/index.592e3b34.js";import{s as R,n as N,r as I,d as M,l as D,q as U}from"../chunks/scheduler.6b0283ec.js";import{S as z,i as H,g as b,s as C,h as g,j as q,B as T,c as k,f as d,k as h,a as y,z as v,A as B,C as G,m as Y,n as J,d as w,t as A,b as K,o as Q,r as V,u as W,v as X,w as Z,p as x}from"../chunks/index.d03f1b61.js";import{e as F}from"../chunks/singletons.89e56d72.js";import{p as $}from"../chunks/stores.9eb0368b.js";import{m as P}from"../chunks/stores.29ae49ab.js";import{a as tt}from"../chunks/ui.970d4c6a.js";const et=async function({url:a,params:{slug:t}}){const{items:l}=await O.collection("posts").getList(void 0,void 0,{filter:`slug="${t}"`}),[e]=l;return{post:e}},pt=Object.freeze(Object.defineProperty({__proto__:null,load:et},Symbol.toStringTag,{value:"Module"}));function st(a){let t,l,e="<aside>Are you sure you want to delete the following record?</aside>",o,i,f="Yes - Proceed",u,_,p="No - Cancel",c,r;return{c(){t=b("form"),l=b("article"),l.innerHTML=e,o=C(),i=b("button"),i.textContent=f,u=C(),_=b("button"),_.textContent=p,this.h()},l(n){t=g(n,"FORM",{});var s=q(t);l=g(s,"ARTICLE",{"data-svelte-h":!0}),T(l)!=="svelte-1ay9ktq"&&(l.innerHTML=e),o=k(s),i=g(s,"BUTTON",{type:!0,"data-svelte-h":!0}),T(i)!=="svelte-1y7wrb9"&&(i.textContent=f),u=k(s),_=g(s,"BUTTON",{type:!0,"data-svelte-h":!0}),T(_)!=="svelte-1851ei0"&&(_.textContent=p),s.forEach(d),this.h()},h(){h(i,"type","submit"),h(_,"type","reset")},m(n,s){y(n,t,s),v(t,l),v(t,o),v(t,i),v(t,u),v(t,_),c||(r=[B(_,"click",a[3]),B(t,"submit",G(a[0]))],c=!0)},p:N,i:N,o:N,d(n){n&&d(t),c=!1,I(r)}}}function lt(a,t,l){let{id:e}=t,{table:o}=t;async function i(){tt(async()=>{await O.collection(o).delete(e),L("..")})}const f=()=>L("..");return a.$$set=u=>{"id"in u&&l(1,e=u.id),"table"in u&&l(2,o=u.table)},[i,e,o,f]}class at extends z{constructor(t){super(),H(this,t,lt,st,R,{id:1,table:2})}}function S(a){let t,l;return t=new at({props:{table:"posts",id:a[4]}}),{c(){V(t.$$.fragment)},l(e){W(t.$$.fragment,e)},m(e,o){X(t,e,o),l=!0},p(e,o){const i={};o&16&&(i.id=e[4]),t.$set(i)},i(e){l||(w(t.$$.fragment,e),l=!0)},o(e){A(t.$$.fragment,e),l=!1},d(e){Z(t,e)}}}function j(a){let t,l;return{c(){t=b("img"),this.h()},l(e){t=g(e,"IMG",{src:!0,alt:!0,class:!0}),this.h()},h(){U(t.src,l=O.getFileUrl(a[0].post,a[2][0],{thumb:"600x0"}))||h(t,"src",l),h(t,"alt",a[1]),h(t,"class","svelte-10t3lmq")},m(e,o){y(e,t,o)},p(e,o){o&5&&!U(t.src,l=O.getFileUrl(e[0].post,e[2][0],{thumb:"600x0"}))&&h(t,"src",l),o&2&&h(t,"alt",e[1])},d(e){e&&d(t)}}}function ot(a){let t,l,e,o,i,f,u,_="AuditLog",p,c,r=a[5].url.hash==="#delete"&&S(a),n=a[2]&&a[2][0]&&j(a);return{c(){r&&r.c(),t=C(),n&&n.c(),l=C(),e=b("pre"),o=Y(a[3]),i=C(),f=b("a"),u=b("button"),u.textContent=_,this.h()},l(s){r&&r.l(s),t=k(s),n&&n.l(s),l=k(s),e=g(s,"PRE",{});var m=q(e);o=J(m,a[3]),m.forEach(d),i=k(s),f=g(s,"A",{href:!0});var E=q(f);u=g(E,"BUTTON",{type:!0,"data-svelte-h":!0}),T(u)!=="svelte-656kos"&&(u.textContent=_),E.forEach(d),this.h()},h(){h(u,"type","button"),h(f,"href",p=`${F}/auditlog/posts/${a[4]}`)},m(s,m){r&&r.m(s,m),y(s,t,m),n&&n.m(s,m),y(s,l,m),y(s,e,m),v(e,o),y(s,i,m),y(s,f,m),v(f,u),c=!0},p(s,[m]){s[5].url.hash==="#delete"?r?(r.p(s,m),m&32&&w(r,1)):(r=S(s),r.c(),w(r,1),r.m(t.parentNode,t)):r&&(x(),A(r,1,1,()=>{r=null}),K()),s[2]&&s[2][0]?n?n.p(s,m):(n=j(s),n.c(),n.m(l.parentNode,l)):n&&(n.d(1),n=null),(!c||m&8)&&Q(o,s[3]),(!c||m&16&&p!==(p=`${F}/auditlog/posts/${s[4]}`))&&h(f,"href",p)},i(s){c||(w(r),c=!0)},o(s){A(r),c=!1},d(s){s&&(d(t),d(l),d(e),d(i),d(f)),r&&r.d(s),n&&n.d(s)}}}function rt(a,t,l){let e,o,i,f,u,_;M(a,P,c=>l(6,u=c)),M(a,$,c=>l(5,_=c));let{data:p}=t;return a.$$set=c=>{"data"in c&&l(0,p=c.data)},a.$$.update=()=>{a.$$.dirty&1&&l(4,{post:{id:e,title:o,body:i,files:f}}=p,e,(l(1,o),l(0,p)),(l(3,i),l(0,p)),(l(2,f),l(0,p))),a.$$.dirty&2&&D(P,u.title=o,u)},[p,o,f,i,e,_]}class dt extends z{constructor(t){super(),H(this,t,rt,ot,R,{data:0})}}export{dt as component,pt as universal};

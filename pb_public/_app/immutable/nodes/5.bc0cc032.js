import{s as fe,v as Y,q as be,n as J,w as ee,d as U,l as ve}from"../chunks/scheduler.6b0283ec.js";import{S as ie,i as ce,g as k,h as E,F as te,a as M,f as g,s as w,c as y,j as D,z as m,d as H,b as ue,t as I,x as ke,B as K,e as le,p as _e,k as N,r as he,m as j,u as de,n as F,v as me,o as z,w as ge}from"../chunks/index.d03f1b61.js";import{c as Ee,e as re,w as Te,a as De}from"../chunks/index.08437746.js";import{m as ne}from"../chunks/stores.29ae49ab.js";function Ce(s,e){const t={},l={},i={$$scope:1};let r=s.length;for(;r--;){const n=s[r],_=e[r];if(_){for(const a in n)a in _||(l[a]=1);for(const a in _)i[a]||(t[a]=_[a],i[a]=1);s[r]=_}else for(const a in n)i[a]=1}for(const n in l)n in t||(t[n]=void 0);return t}function Ae(s){let e,t,l=[s[0],{src:t=s[1]},{rel:"noreferrer"}],i={};for(let r=0;r<l.length;r+=1)i=Y(i,l[r]);return{c(){e=k("img"),this.h()},l(r){e=E(r,"IMG",{src:!0,rel:!0}),this.h()},h(){te(e,i)},m(r,n){M(r,e,n)},p(r,[n]){te(e,i=Ce(l,[n&1&&r[0],n&2&&!be(e.src,t=r[1])&&{src:t},{rel:"noreferrer"}]))},i:J,o:J,d(r){r&&g(e)}}}function Me(s,e,t){let l,{record:i}=e,{file:r}=e,{thumb:n}=e,_;return s.$$set=a=>{t(5,e=Y(Y({},e),ee(a))),"record"in a&&t(2,i=a.record),"file"in a&&t(3,r=a.file),"thumb"in a&&t(4,n=a.thumb)},s.$$.update=()=>{t(2,{record:i,file:r,thumb:n,..._}=e,i,(t(3,r),t(5,e)),(t(4,n),t(5,e)),(t(0,_),t(5,e))),s.$$.dirty&28&&t(1,l=r?Ee.getFileUrl(i,r,{thumb:n}):`https://via.placeholder.com/${n??"100x100"}`)},e=ee(e),[_,l,i,r,n]}class pe extends ie{constructor(e){super(),ce(this,e,Me,Ae,fe,{record:2,file:3,thumb:4})}}function ae(s,e,t){const l=s.slice();return l[4]=e[t],l}function Re(s){let e,t="Please login to create new posts.";return{c(){e=k("p"),e.textContent=t},l(l){e=E(l,"P",{"data-svelte-h":!0}),K(e)!=="svelte-1mmrf4t"&&(e.textContent=t)},m(l,i){M(l,e,i)},d(l){l&&g(e)}}}function we(s){let e,t="Create New";return{c(){e=k("a"),e.textContent=t,this.h()},l(l){e=E(l,"A",{href:!0,"data-svelte-h":!0}),K(e)!=="svelte-hvcvuj"&&(e.textContent=t),this.h()},h(){N(e,"href","new/edit")},m(l,i){M(l,e,i)},d(l){l&&g(e)}}}function oe(s){let e,t="<td>No posts found.</td> ";return{c(){e=k("tr"),e.innerHTML=t},l(l){e=E(l,"TR",{"data-svelte-h":!0}),K(e)!=="svelte-rra45l"&&(e.innerHTML=t)},m(l,i){M(l,e,i)},p:J,d(l){l&&g(e)}}}function ye(s){let e,t,l,i,r,n,_=s[4].title+"",a,p,b,c,C=s[4].updated+"",h,f,u;return l=new pe({props:{record:s[4],file:s[4].files[0],thumb:"100x100",alt:s[4].title}}),{c(){e=k("tr"),t=k("td"),he(l.$$.fragment),i=w(),r=k("td"),n=k("a"),a=j(_),b=w(),c=k("td"),h=j(C),f=w(),this.h()},l(o){e=E(o,"TR",{});var d=D(e);t=E(d,"TD",{});var A=D(t);de(l.$$.fragment,A),A.forEach(g),i=y(d),r=E(d,"TD",{});var L=D(r);n=E(L,"A",{href:!0});var P=D(n);a=F(P,_),P.forEach(g),L.forEach(g),b=y(d),c=E(d,"TD",{});var R=D(c);h=F(R,C),R.forEach(g),f=y(d),d.forEach(g),this.h()},h(){N(n,"href",p=s[4].slug)},m(o,d){M(o,e,d),m(e,t),me(l,t,null),m(e,i),m(e,r),m(r,n),m(n,a),m(e,b),m(e,c),m(c,h),m(e,f),u=!0},p(o,d){const A={};d&2&&(A.record=o[4]),d&2&&(A.file=o[4].files[0]),d&2&&(A.alt=o[4].title),l.$set(A),(!u||d&2)&&_!==(_=o[4].title+"")&&z(a,_),(!u||d&2&&p!==(p=o[4].slug))&&N(n,"href",p),(!u||d&2)&&C!==(C=o[4].updated+"")&&z(h,C)},i(o){u||(H(l.$$.fragment,o),u=!0)},o(o){I(l.$$.fragment,o),u=!1},d(o){o&&g(e),ge(l)}}}function Be(s){let e,t,l,i,r,n,_=s[4].title+"",a,p,b,c,C=s[4].updated+"",h,f,u,o,d,A,L,P,R,G,S,O,B;return l=new pe({props:{record:s[4],file:s[4].files[0],thumb:"100x100",alt:s[4].title}}),{c(){e=k("tr"),t=k("td"),he(l.$$.fragment),i=w(),r=k("td"),n=k("a"),a=j(_),b=w(),c=k("td"),h=j(C),f=w(),u=k("td"),o=k("a"),d=j("Edit"),L=w(),P=k("td"),R=k("a"),G=j("Delete"),O=w(),this.h()},l(T){e=E(T,"TR",{});var v=D(e);t=E(v,"TD",{});var q=D(t);de(l.$$.fragment,q),q.forEach(g),i=y(v),r=E(v,"TD",{});var Q=D(r);n=E(Q,"A",{href:!0});var V=D(n);a=F(V,_),V.forEach(g),Q.forEach(g),b=y(v),c=E(v,"TD",{});var W=D(c);h=F(W,C),W.forEach(g),f=y(v),u=E(v,"TD",{});var X=D(u);o=E(X,"A",{href:!0});var Z=D(o);d=F(Z,"Edit"),Z.forEach(g),X.forEach(g),L=y(v),P=E(v,"TD",{});var x=D(P);R=E(x,"A",{href:!0});var $=D(R);G=F($,"Delete"),$.forEach(g),x.forEach(g),O=y(v),v.forEach(g),this.h()},h(){N(n,"href",p=s[4].slug),N(o,"href",A=`${s[4].id}/edit`),N(R,"href",S=`${s[4].slug}#delete`)},m(T,v){M(T,e,v),m(e,t),me(l,t,null),m(e,i),m(e,r),m(r,n),m(n,a),m(e,b),m(e,c),m(c,h),m(e,f),m(e,u),m(u,o),m(o,d),m(e,L),m(e,P),m(P,R),m(R,G),m(e,O),B=!0},p(T,v){const q={};v&2&&(q.record=T[4]),v&2&&(q.file=T[4].files[0]),v&2&&(q.alt=T[4].title),l.$set(q),(!B||v&2)&&_!==(_=T[4].title+"")&&z(a,_),(!B||v&2&&p!==(p=T[4].slug))&&N(n,"href",p),(!B||v&2)&&C!==(C=T[4].updated+"")&&z(h,C),(!B||v&2&&A!==(A=`${T[4].id}/edit`))&&N(o,"href",A),(!B||v&2&&S!==(S=`${T[4].slug}#delete`))&&N(R,"href",S)},i(T){B||(H(l.$$.fragment,T),B=!0)},o(T){I(l.$$.fragment,T),B=!1},d(T){T&&g(e),ge(l)}}}function se(s){let e,t,l,i;const r=[Be,ye],n=[];function _(a,p){var b;return((b=a[0])==null?void 0:b.id)==a[4].user?0:1}return e=_(s),t=n[e]=r[e](s),{c(){t.c(),l=le()},l(a){t.l(a),l=le()},m(a,p){n[e].m(a,p),M(a,l,p),i=!0},p(a,p){let b=e;e=_(a),e===b?n[e].p(a,p):(_e(),I(n[b],1,1,()=>{n[b]=null}),ue(),t=n[e],t?t.p(a,p):(t=n[e]=r[e](a),t.c()),H(t,1),t.m(l.parentNode,l))},i(a){i||(H(t),i=!0)},o(a){I(t),i=!1},d(a){a&&g(l),n[e].d(a)}}}function Ne(s){let e,t,l,i,r,n;function _(f,u){return f[0]?we:Re}let a=_(s),p=a(s),b=re(s[1].items),c=[];for(let f=0;f<b.length;f+=1)c[f]=se(ae(s,b,f));const C=f=>I(c[f],1,1,()=>{c[f]=null});let h=null;return b.length||(h=oe()),{c(){p.c(),e=w(),t=k("hr"),l=w(),i=k("table"),r=k("tbody");for(let f=0;f<c.length;f+=1)c[f].c();h&&h.c()},l(f){p.l(f),e=y(f),t=E(f,"HR",{}),l=y(f),i=E(f,"TABLE",{});var u=D(i);r=E(u,"TBODY",{});var o=D(r);for(let d=0;d<c.length;d+=1)c[d].l(o);h&&h.l(o),o.forEach(g),u.forEach(g)},m(f,u){p.m(f,u),M(f,e,u),M(f,t,u),M(f,l,u),M(f,i,u),m(i,r);for(let o=0;o<c.length;o+=1)c[o]&&c[o].m(r,null);h&&h.m(r,null),n=!0},p(f,[u]){if(a!==(a=_(f))&&(p.d(1),p=a(f),p&&(p.c(),p.m(e.parentNode,e))),u&3){b=re(f[1].items);let o;for(o=0;o<b.length;o+=1){const d=ae(f,b,o);c[o]?(c[o].p(d,u),H(c[o],1)):(c[o]=se(d),c[o].c(),H(c[o],1),c[o].m(r,null))}for(_e(),o=b.length;o<c.length;o+=1)C(o);ue(),!b.length&&h?h.p(f,u):b.length?h&&(h.d(1),h=null):(h=oe(),h.c(),h.m(r,null))}},i(f){if(!n){for(let u=0;u<b.length;u+=1)H(c[u]);n=!0}},o(f){c=c.filter(Boolean);for(let u=0;u<c.length;u+=1)I(c[u]);n=!1},d(f){f&&(g(e),g(t),g(l),g(i)),p.d(f),ke(c,f),h&&h.d()}}}function Pe(s,e,t){let l,i,r;U(s,ne,_=>t(3,l=_)),U(s,De,_=>t(0,i=_)),ve(ne,l.title="Recent Posts",l);const n=Te("posts",{sort:"-updated"});return U(s,n,_=>t(1,r=_)),[i,r,n]}class Fe extends ie{constructor(e){super(),ce(this,e,Pe,Ne,fe,{})}}export{Fe as component};

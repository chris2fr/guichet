import{s as y,n as v}from"../chunks/scheduler.6b0283ec.js";import{S as C,i as j,g as d,s as h,m as S,h as m,B as x,c as g,j as b,n as P,f as i,a as p,z as O,o as w}from"../chunks/index.d03f1b61.js";const q=async function({fetch:o}){return{...await(await o("/api/hello")).json()}},N=Object.freeze(Object.defineProperty({__proto__:null,load:q},Symbol.toStringTag,{value:"Module"}));function z(o){let e,r="Hello!",s,a,u="Got the following API response from the backend server",f,l,c=JSON.stringify(o[0])+"",_;return{c(){e=d("h1"),e.textContent=r,s=h(),a=d("p"),a.textContent=u,f=h(),l=d("pre"),_=S(c)},l(t){e=m(t,"H1",{"data-svelte-h":!0}),x(e)!=="svelte-gbi83v"&&(e.textContent=r),s=g(t),a=m(t,"P",{"data-svelte-h":!0}),x(a)!=="svelte-cn3rqt"&&(a.textContent=u),f=g(t),l=m(t,"PRE",{});var n=b(l);_=P(n,c),n.forEach(i)},m(t,n){p(t,e,n),p(t,s,n),p(t,a,n),p(t,f,n),p(t,l,n),O(l,_)},p(t,[n]){n&1&&c!==(c=JSON.stringify(t[0])+"")&&w(_,c)},i:v,o:v,d(t){t&&(i(e),i(s),i(a),i(f),i(l))}}}function E(o,e,r){let{data:s}=e;return o.$$set=a=>{"data"in a&&r(0,s=a.data)},[s]}class k extends C{constructor(e){super(),j(this,e,E,z,y,{data:0})}}export{k as component,N as universal};
